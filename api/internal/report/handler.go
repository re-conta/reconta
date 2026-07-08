package report

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/re-conta/reconta/api/internal/account"
	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/category"
	"github.com/re-conta/reconta/api/internal/tag"
	"github.com/re-conta/reconta/api/internal/transaction"
)

// maxUploadSize limita o tamanho do arquivo de backup JSON aceito no upload.
const maxUploadSize = 10 << 20 // 10MB

type Handler struct {
	transactions *transaction.Repository
	categories   *category.Repository
	accounts     *account.Repository
	tags         *tag.Repository
	auth         *auth.Handler
}

func NewHandler(transactions *transaction.Repository, categories *category.Repository, accounts *account.Repository, tags *tag.Repository, authHandler *auth.Handler) *Handler {
	return &Handler{transactions: transactions, categories: categories, accounts: accounts, tags: tags, auth: authHandler}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/reports/export", h.auth.RequireUser(h.export))
	mux.HandleFunc("POST /api/reports/import", h.auth.RequireUser(h.importBackup))
}

type exportRequest struct {
	Format   string       `json:"format"`
	Scope    string       `json:"scope"`
	Month    int          `json:"month"`
	Year     int          `json:"year"`
	DateFrom string       `json:"dateFrom"`
	DateTo   string       `json:"dateTo"`
	Charts   []ChartImage `json:"charts"`
}

func (h *Handler) export(w http.ResponseWriter, r *http.Request, userID int64) {
	var req exportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	scope, err := ResolveScope(req.Scope, req.Month, req.Year, req.DateFrom, req.DateTo)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	filters := transaction.ListFilters{
		Type:     "", // relatório inclui receitas e despesas
		DateFrom: scope.DateFrom,
		DateTo:   scope.DateTo,
	}
	txs, totals, err := h.transactions.ListAll(r.Context(), userID, filters)
	if err != nil {
		log.Printf("erro ao listar transações para relatório: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if err := h.attachTags(r, txs); err != nil {
		log.Printf("erro ao carregar tags para relatório: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	accountNames, err := h.accountNameMap(r, userID)
	if err != nil {
		log.Printf("erro ao carregar contas para relatório: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	var (
		data        []byte
		contentType string
		ext         string
	)

	switch req.Format {
	case "xlsx":
		data, err = BuildXLSX(scope, txs, totals, req.Charts, accountNames)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		ext = "xlsx"
	case "ods":
		data, err = BuildODS(scope, txs, totals, req.Charts, accountNames)
		contentType = "application/vnd.oasis.opendocument.spreadsheet"
		ext = "ods"
	case "pdf":
		data, err = BuildPDF(scope, txs, totals, req.Charts, accountNames)
		contentType = "application/pdf"
		ext = "pdf"
	case "json":
		payload := ToBackupPayload(scope, txs, totals, accountNames)
		data, err = BuildJSON(payload)
		contentType = "application/json"
		ext = "json"
	default:
		writeError(w, http.StatusBadRequest, "formato inválido")
		return
	}
	if err != nil {
		log.Printf("erro ao gerar relatório em %s: %v", req.Format, err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	filename := fmt.Sprintf("relatorio-gastos.%s", ext)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (h *Handler) importBackup(w http.ResponseWriter, r *http.Request, userID int64) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "arquivo muito grande ou requisição inválida")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "arquivo não enviado")
		return
	}
	defer file.Close()

	raw, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "falha ao ler o arquivo enviado")
		return
	}

	var payload BackupPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "arquivo de backup inválido")
		return
	}

	categoryCache := map[string]int64{}
	accountCache := map[string]int64{}
	tagCache := map[string]int64{}

	imported := 0
	skipped := 0
	for _, rec := range payload.Transactions {
		if rec.Date == "" || rec.Description == "" || rec.Amount == 0 || (rec.Type != "income" && rec.Type != "expense") {
			continue
		}

		dup, err := h.transactions.FindDuplicate(r.Context(), userID, rec.Date, rec.Amount, rec.Description)
		if err != nil {
			log.Printf("erro ao verificar duplicidade ao restaurar backup: %v", err)
		}
		if dup {
			skipped++
			continue
		}

		var categoryID, accountID *int64
		if rec.CategoryName != nil && *rec.CategoryName != "" {
			id, err := h.resolveCategory(r, userID, categoryCache, *rec.CategoryName, rec.Type)
			if err != nil {
				log.Printf("erro ao resolver categoria do backup: %v", err)
			} else {
				categoryID = &id
			}
		}
		if rec.AccountName != nil && *rec.AccountName != "" {
			id, err := h.resolveAccount(r, userID, accountCache, *rec.AccountName)
			if err != nil {
				log.Printf("erro ao resolver conta do backup: %v", err)
			} else {
				accountID = &id
			}
		}

		tx, err := h.transactions.Create(r.Context(), userID, transaction.Input{
			Date:           rec.Date,
			Description:    rec.Description,
			Amount:         absFloat(rec.Amount),
			Type:           rec.Type,
			CategoryID:     categoryID,
			AccountID:      accountID,
			Notes:          rec.Notes,
			ImportedFrom:   strPtr("backup"),
			Bank:           rec.Bank,
			PixBeneficiary: rec.PixBeneficiary,
		})
		if err != nil {
			log.Printf("erro ao restaurar transação do backup: %v", err)
			continue
		}

		if len(rec.Tags) > 0 {
			tagIDs := make([]int64, 0, len(rec.Tags))
			for _, name := range rec.Tags {
				id, err := h.resolveTag(r, userID, tagCache, name)
				if err != nil {
					log.Printf("erro ao resolver tag do backup: %v", err)
					continue
				}
				tagIDs = append(tagIDs, id)
			}
			if len(tagIDs) > 0 {
				if err := h.tags.SetTransactionTags(r.Context(), tx.ID, tagIDs); err != nil {
					log.Printf("erro ao associar tags do backup: %v", err)
				}
			}
		}

		imported++
	}

	writeJSON(w, http.StatusOK, map[string]int{
		"imported": imported,
		"skipped":  skipped,
		"total":    len(payload.Transactions),
	})
}

func (h *Handler) resolveCategory(r *http.Request, userID int64, cache map[string]int64, name, txType string) (int64, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}
	c, err := h.categories.FindOrCreateByName(r.Context(), userID, name, "#94a3b8", "tag", txType)
	if err != nil {
		return 0, err
	}
	cache[name] = c.ID
	return c.ID, nil
}

func (h *Handler) resolveAccount(r *http.Request, userID int64, cache map[string]int64, name string) (int64, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}
	a, err := h.accounts.FindOrCreateByName(r.Context(), userID, name, "checking")
	if err != nil {
		return 0, err
	}
	cache[name] = a.ID
	return a.ID, nil
}

func (h *Handler) resolveTag(r *http.Request, userID int64, cache map[string]int64, name string) (int64, error) {
	if id, ok := cache[name]; ok {
		return id, nil
	}
	t, err := h.tags.FindOrCreateByName(r.Context(), userID, name, "#94a3b8")
	if err != nil {
		return 0, err
	}
	cache[name] = t.ID
	return t.ID, nil
}

func (h *Handler) attachTags(r *http.Request, txs []transaction.Transaction) error {
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

func (h *Handler) accountNameMap(r *http.Request, userID int64) (map[int64]string, error) {
	accounts, err := h.accounts.List(r.Context(), userID)
	if err != nil {
		return nil, err
	}
	names := make(map[int64]string, len(accounts))
	for _, a := range accounts {
		names[a.ID] = a.Name
	}
	return names, nil
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func strPtr(s string) *string { return &s }

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
