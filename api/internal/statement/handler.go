package statement

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/category"
	"github.com/re-conta/reconta/api/internal/transaction"
)

// maxUploadSize limita o tamanho do PDF de extrato aceito no upload.
const maxUploadSize = 20 << 20 // 20MB

type Handler struct {
	transactions *transaction.Repository
	categories   *category.Repository
	auth         *auth.Handler
}

func NewHandler(transactions *transaction.Repository, categories *category.Repository, authHandler *auth.Handler) *Handler {
	return &Handler{transactions: transactions, categories: categories, auth: authHandler}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/transactions/import/banks", h.auth.RequireUser(h.listBanks))
	mux.HandleFunc("POST /api/transactions/import/preview", h.auth.RequireUser(h.preview))
	mux.HandleFunc("POST /api/transactions/import/confirm", h.auth.RequireUser(h.confirm))
}

func (h *Handler) listBanks(w http.ResponseWriter, r *http.Request, userID int64) {
	writeJSON(w, http.StatusOK, SupportedBanks)
}

func (h *Handler) preview(w http.ResponseWriter, r *http.Request, userID int64) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "arquivo muito grande ou requisição inválida")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "arquivo não enviado")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".pdf") {
		writeError(w, http.StatusUnprocessableEntity, "envie um arquivo PDF")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "falha ao ler o arquivo enviado")
		return
	}

	text, err := ExtractText(data)
	if err != nil {
		log.Printf("erro ao extrair texto do pdf: %v", err)
		writeError(w, http.StatusUnprocessableEntity, "não foi possível ler o PDF (verifique se não está protegido por senha)")
		return
	}

	bank := DetectBank(text)
	if override := r.FormValue("bank"); override != "" {
		bank = BankByKey(override)
	}

	parsed := ParseStatement(bank.Key, text)
	if len(parsed) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "nenhum lançamento foi reconhecido neste extrato")
		return
	}

	cats, err := h.categories.ListWithPatterns(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao carregar categorias com padrões: %v", err)
	}
	rules := compileCategoryRules(cats)

	for i := range parsed {
		dup, err := h.transactions.FindDuplicate(r.Context(), userID, parsed[i].Date, parsed[i].Amount, parsed[i].Description)
		if err != nil {
			log.Printf("erro ao verificar duplicidade: %v", err)
		}
		parsed[i].Duplicate = dup

		haystack := parsed[i].Description
		if parsed[i].PixBeneficiary != nil {
			haystack += " " + *parsed[i].PixBeneficiary
		}
		for _, ru := range rules {
			if ru.matches(haystack) {
				id, name := ru.categoryID, ru.categoryName
				parsed[i].CategoryID = &id
				parsed[i].CategoryName = &name
				break
			}
		}
	}

	writeJSON(w, http.StatusOK, PreviewResult{Bank: bank.Key, BankLabel: bank.Label, Transactions: parsed})
}

type importRow struct {
	Date           string  `json:"date"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Type           string  `json:"type"`
	CategoryID     *int64  `json:"categoryId"`
	PixBeneficiary *string `json:"pixBeneficiary"`
}

func (h *Handler) confirm(w http.ResponseWriter, r *http.Request, userID int64) {
	var req struct {
		Bank         string      `json:"bank"`
		AccountID    *int64      `json:"accountId"`
		Transactions []importRow `json:"transactions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if len(req.Transactions) == 0 {
		writeError(w, http.StatusBadRequest, "nenhuma transação para importar")
		return
	}

	bankLabel := BankByKey(req.Bank).Label
	imported := 0
	for _, row := range req.Transactions {
		if row.Date == "" || row.Description == "" || row.Amount == 0 || (row.Type != "income" && row.Type != "expense") {
			continue
		}
		_, err := h.transactions.Create(r.Context(), userID, transaction.Input{
			Date:           row.Date,
			Description:    row.Description,
			Amount:         absFloat(row.Amount),
			Type:           row.Type,
			CategoryID:     row.CategoryID,
			AccountID:      req.AccountID,
			ImportedFrom:   new("pdf"),
			Bank:           new(bankLabel),
			PixBeneficiary: row.PixBeneficiary,
		})
		if err != nil {
			log.Printf("erro ao importar transação: %v", err)
			continue
		}
		imported++
	}

	writeJSON(w, http.StatusOK, map[string]int{"imported": imported, "total": len(req.Transactions)})
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

type categoryRule struct {
	categoryID   int64
	categoryName string
	patterns     []*regexp.Regexp
}

func (c categoryRule) matches(haystack string) bool {
	for _, re := range c.patterns {
		if re.MatchString(haystack) {
			return true
		}
	}
	return false
}

func compileCategoryRules(cats []category.Category) []categoryRule {
	rules := make([]categoryRule, 0, len(cats))
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
			rules = append(rules, categoryRule{categoryID: c.ID, categoryName: c.Name, patterns: patterns})
		}
	}
	return rules
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
