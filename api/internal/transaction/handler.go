package transaction

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/re-conta/reconta/api/internal/account"
	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/category"
	"github.com/re-conta/reconta/api/internal/tag"
)

type Handler struct {
	repo       *Repository
	tags       *tag.Repository
	categories *category.Repository
	accounts   *account.Repository
	auth       *auth.Handler
}

func NewHandler(repo *Repository, tags *tag.Repository, categories *category.Repository, accounts *account.Repository, authHandler *auth.Handler) *Handler {
	return &Handler{repo: repo, tags: tags, categories: categories, accounts: accounts, auth: authHandler}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/transactions", h.auth.RequireUser(h.list))
	mux.HandleFunc("GET /api/transactions/periods", h.auth.RequireUser(h.periods))
	mux.HandleFunc("POST /api/transactions", h.auth.RequireUser(h.create))
	mux.HandleFunc("PATCH /api/transactions", h.auth.RequireUser(h.bulkUpdate))
	mux.HandleFunc("DELETE /api/transactions", h.auth.RequireUser(h.bulkDelete))
	mux.HandleFunc("POST /api/transactions/auto-categorize", h.auth.RequireUser(h.autoCategorize))
	mux.HandleFunc("GET /api/transactions/opening-balance", h.auth.RequireUser(h.getOpeningBalance))
	mux.HandleFunc("POST /api/transactions/opening-balance", h.auth.RequireUser(h.postOpeningBalance))
	mux.HandleFunc("GET /api/transactions/{id}", h.auth.RequireUser(h.get))
	mux.HandleFunc("PUT /api/transactions/{id}", h.auth.RequireUser(h.update))
	mux.HandleFunc("DELETE /api/transactions/{id}", h.auth.RequireUser(h.delete))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, userID int64) {
	q := r.URL.Query()

	filters := ListFilters{
		Type:   q.Get("type"),
		Search: q.Get("search"),
	}
	filters.Month, _ = strconv.Atoi(q.Get("month"))
	filters.Year, _ = strconv.Atoi(q.Get("year"))
	filters.CategoryID, _ = strconv.ParseInt(q.Get("categoryId"), 10, 64)
	filters.TagID, _ = strconv.ParseInt(q.Get("tagId"), 10, 64)
	filters.Page, _ = strconv.Atoi(q.Get("page"))
	filters.Limit, _ = strconv.Atoi(q.Get("limit"))

	result, err := h.repo.List(r.Context(), userID, filters)
	if err != nil {
		log.Printf("erro ao listar transações: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if err := h.attachTags(r, result.Data); err != nil {
		log.Printf("erro ao carregar tags das transações: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *Handler) periods(w http.ResponseWriter, r *http.Request, userID int64) {
	periods, err := h.repo.ListPeriods(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao listar períodos: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, periods)
}

func (h *Handler) attachTags(r *http.Request, txs []Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	ids := make([]int64, len(txs))
	for i, t := range txs {
		ids[i] = t.ID
	}
	tagsByTx, err := h.tags.ListByTransactionIDs(r.Context(), ids)
	if err != nil {
		return err
	}
	for i := range txs {
		if tags, ok := tagsByTx[txs[i].ID]; ok {
			txs[i].Tags = tags
		}
	}
	return nil
}

type transactionRequest struct {
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	CategoryID  *int64  `json:"categoryId"`
	AccountID   *int64  `json:"accountId"`
	Notes       *string `json:"notes"`
	TagIDs      []int64 `json:"tagIds"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, userID int64) {
	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if req.Date == "" || req.Description == "" || req.Amount == 0 || req.Type == "" {
		writeError(w, http.StatusBadRequest, "campos obrigatórios faltando")
		return
	}

	tx, err := h.repo.Create(r.Context(), userID, Input{
		Date:        req.Date,
		Description: req.Description,
		Amount:      absFloat(req.Amount),
		Type:        req.Type,
		CategoryID:  req.CategoryID,
		AccountID:   req.AccountID,
		Notes:       req.Notes,
	})
	if err != nil {
		log.Printf("erro ao criar transação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if len(req.TagIDs) > 0 {
		owned, err := h.tags.FilterOwnedIDs(r.Context(), userID, req.TagIDs)
		if err != nil {
			log.Printf("erro ao validar tags: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		if err := h.tags.SetTransactionTags(r.Context(), tx.ID, owned); err != nil {
			log.Printf("erro ao associar tags: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		tx.Tags, _ = h.tags.ListByTransactionID(r.Context(), tx.ID)
	}

	writeJSON(w, http.StatusCreated, tx)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	tx, err := h.repo.GetByID(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrado")
			return
		}
		log.Printf("erro ao buscar transação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	tx.Tags, err = h.tags.ListByTransactionID(r.Context(), tx.ID)
	if err != nil {
		log.Printf("erro ao carregar tags da transação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, tx)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	tx, err := h.repo.Update(r.Context(), userID, id, Input{
		Date:        req.Date,
		Description: req.Description,
		Amount:      absFloat(req.Amount),
		Type:        req.Type,
		CategoryID:  req.CategoryID,
		AccountID:   req.AccountID,
		Notes:       req.Notes,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrado")
			return
		}
		log.Printf("erro ao atualizar transação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if req.TagIDs != nil {
		owned, err := h.tags.FilterOwnedIDs(r.Context(), userID, req.TagIDs)
		if err != nil {
			log.Printf("erro ao validar tags: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		if err := h.tags.SetTransactionTags(r.Context(), tx.ID, owned); err != nil {
			log.Printf("erro ao associar tags: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
	}
	tx.Tags, _ = h.tags.ListByTransactionID(r.Context(), tx.ID)

	writeJSON(w, http.StatusOK, tx)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	if err := h.repo.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrado")
			return
		}
		log.Printf("erro ao remover transação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) bulkUpdate(w http.ResponseWriter, r *http.Request, userID int64) {
	var req struct {
		IDs    []int64                    `json:"ids"`
		Fields map[string]json.RawMessage `json:"fields"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if len(req.IDs) == 0 {
		writeError(w, http.StatusBadRequest, "ids obrigatórios")
		return
	}

	var fields BulkUpdateFields

	if raw, ok := req.Fields["type"]; ok {
		var v string
		if err := json.Unmarshal(raw, &v); err == nil {
			fields.Type = &v
		}
	}
	if raw, ok := req.Fields["categoryId"]; ok {
		id, err := parseNullableID(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "categoryId inválido")
			return
		}
		fields.CategoryID = &id
	}
	if raw, ok := req.Fields["accountId"]; ok {
		id, err := parseNullableID(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "accountId inválido")
			return
		}
		fields.AccountID = &id
	}
	if raw, ok := req.Fields["date"]; ok {
		var v string
		if err := json.Unmarshal(raw, &v); err == nil {
			fields.Date = &v
		}
	}

	n, err := h.repo.BulkUpdate(r.Context(), userID, req.IDs, fields)
	if err != nil {
		log.Printf("erro ao atualizar transações em lote: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"updated": n})
}

// parseNullableID interpreta um campo que pode ser um número, a string "_none"
// (limpa o vínculo) ou null/ausente (também limpa o vínculo).
func parseNullableID(raw json.RawMessage) (*int64, error) {
	trimmed := strings.TrimSpace(string(raw))
	if trimmed == "null" || trimmed == `"_none"` || trimmed == "" {
		return nil, nil
	}
	var id int64
	if err := json.Unmarshal(raw, &id); err != nil {
		return nil, err
	}
	return &id, nil
}

func (h *Handler) bulkDelete(w http.ResponseWriter, r *http.Request, userID int64) {
	var req struct {
		Scope string `json:"scope"`
		Month int    `json:"month"`
		Year  int    `json:"year"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	scope := BulkDeleteScope(req.Scope)
	if scope != ScopeMonth && scope != ScopeYear && scope != ScopeAll {
		writeError(w, http.StatusBadRequest, "escopo inválido")
		return
	}

	n, err := h.repo.BulkDelete(r.Context(), userID, scope, req.Month, req.Year)
	if err != nil {
		log.Printf("erro ao remover transações em lote: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"deleted": n})
}

func (h *Handler) autoCategorize(w http.ResponseWriter, r *http.Request, userID int64) {
	cats, err := h.categories.ListWithPatterns(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao listar categorias com padrões: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	type rule struct {
		categoryID int64
		patterns   []*regexp.Regexp
	}

	rules := make([]rule, 0, len(cats))
	for _, c := range cats {
		var patterns []*regexp.Regexp
		for line := range strings.SplitSeq(c.Patterns, "\n") {
			p := strings.TrimSpace(line)
			if p == "" {
				continue
			}
			re, err := regexp.Compile("(?i)" + p)
			if err != nil {
				re = regexp.MustCompile("(?i)" + regexp.QuoteMeta(p))
			}
			patterns = append(patterns, re)
		}
		if len(patterns) > 0 {
			rules = append(rules, rule{categoryID: c.ID, patterns: patterns})
		}
	}

	if len(rules) == 0 {
		writeJSON(w, http.StatusOK, map[string]int{"updated": 0, "checked": 0})
		return
	}

	uncategorized, err := h.repo.ListUncategorized(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao listar transações não categorizadas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	updates := map[int64]int64{}
	for _, tx := range uncategorized {
		pix := ""
		if tx.PixBeneficiary != nil {
			pix = *tx.PixBeneficiary
		}
		haystack := tx.Description + " " + pix

		for _, ru := range rules {
			matched := false
			for _, re := range ru.patterns {
				if re.MatchString(haystack) {
					matched = true
					break
				}
			}
			if matched {
				updates[tx.ID] = ru.categoryID
				break
			}
		}
	}

	if err := h.repo.BulkSetCategory(r.Context(), userID, updates); err != nil {
		log.Printf("erro ao aplicar auto-categorização: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, map[string]int{"updated": len(updates), "checked": len(uncategorized)})
}

func (h *Handler) getOpeningBalance(w http.ResponseWriter, r *http.Request, userID int64) {
	month, errM := strconv.Atoi(r.URL.Query().Get("month"))
	year, errY := strconv.Atoi(r.URL.Query().Get("year"))
	if errM != nil || errY != nil || month == 0 || year == 0 {
		writeError(w, http.StatusBadRequest, "month e year são obrigatórios")
		return
	}

	amount, err := h.repo.GetOpeningBalance(r.Context(), userID, month, year)
	if err != nil {
		log.Printf("erro ao buscar saldo de abertura: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if amount != nil {
		writeJSON(w, http.StatusOK, map[string]float64{"amount": *amount})
		return
	}

	total, err := h.accounts.SumBalance(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao somar saldo das contas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"amount": total})
}

func (h *Handler) postOpeningBalance(w http.ResponseWriter, r *http.Request, userID int64) {
	var req struct {
		Month  int     `json:"month"`
		Year   int     `json:"year"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if req.Month == 0 || req.Year == 0 {
		writeError(w, http.StatusBadRequest, "month, year e amount são obrigatórios")
		return
	}

	if err := h.repo.UpsertOpeningBalance(r.Context(), userID, req.Month, req.Year, req.Amount); err != nil {
		log.Printf("erro ao salvar saldo de abertura: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]float64{"amount": req.Amount})
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("erro ao codificar resposta JSON: %v", err)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
