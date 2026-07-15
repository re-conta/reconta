package health

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/user"
)

type Handler struct {
	repo *Repository
	auth *auth.Handler
}

func NewHandler(repo *Repository, authHandler *auth.Handler) *Handler {
	return &Handler{repo: repo, auth: authHandler}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/financial-health", h.auth.RequireUser(h.score))
	mux.HandleFunc("GET /api/admin/financial-health", h.requirePermission(h.getSettings, user.PermAdminPanel))
	mux.HandleFunc("PUT /api/admin/financial-health", h.requirePermission(h.updateSettings, user.PermAdminPanel))
}

// requirePermission exige que o usuário autenticado possua a permissão
// informada (o Super Admin sempre passa), no mesmo espírito do pacote user.
func (h *Handler) requirePermission(next http.HandlerFunc, perm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.auth.CurrentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		if !u.HasPermission(perm) {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
		next(w, r)
	}
}

// scoreResponse é a "saúde" das finanças do usuário no mês: nível calculado a
// partir da taxa de poupança (saldo/receitas) comparada aos limites globais.
type scoreResponse struct {
	Enabled     bool    `json:"enabled"`
	HasData     bool    `json:"hasData"`
	Level       string  `json:"level"`
	Stars       int     `json:"stars"`
	Income      float64 `json:"income"`
	Expense     float64 `json:"expense"`
	Balance     float64 `json:"balance"`
	SavingsRate float64 `json:"savingsRate"`
}

func (h *Handler) score(w http.ResponseWriter, r *http.Request, userID int64) {
	settings, err := h.repo.GetSettings(r.Context())
	if err != nil {
		log.Printf("erro ao ler configuração de saúde financeira: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !settings.Enabled {
		writeJSON(w, http.StatusOK, scoreResponse{Enabled: false})
		return
	}

	now := time.Now()
	month := queryInt(r, "month", int(now.Month()))
	year := queryInt(r, "year", now.Year())
	if month < 1 || month > 12 || year < 1900 || year > 3000 {
		writeError(w, http.StatusBadRequest, "mês ou ano inválido")
		return
	}

	income, expense, err := h.repo.MonthTotals(r.Context(), userID, month, year)
	if err != nil {
		log.Printf("erro ao calcular totais do mês: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	resp := scoreResponse{
		Enabled: true,
		HasData: income > 0 || expense > 0,
		Income:  income,
		Expense: expense,
		Balance: income - expense,
	}
	if resp.HasData {
		resp.SavingsRate, resp.Level, resp.Stars = classify(income, expense, settings)
	}
	writeJSON(w, http.StatusOK, resp)
}

// classify converte receitas vs despesas em taxa de poupança (%) e a compara
// aos limites configurados, do melhor para o pior nível.
func classify(income, expense float64, s Settings) (rate float64, level string, stars int) {
	if income > 0 {
		rate = (income - expense) / income * 100
	} else {
		// Só despesas no mês: pior cenário possível.
		rate = -100
	}

	switch {
	case rate >= s.ThresholdOtima:
		return rate, "otima", 5
	case rate >= s.ThresholdBoa:
		return rate, "boa", 4
	case rate >= s.ThresholdEstavel:
		return rate, "estavel", 3
	case rate >= s.ThresholdRuim:
		return rate, "ruim", 2
	default:
		return rate, "pessima", 1
	}
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.repo.GetSettings(r.Context())
	if err != nil {
		log.Printf("erro ao ler configuração de saúde financeira: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request) {
	var s Settings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	// Os limites precisam ser estritamente decrescentes para que cada taxa de
	// poupança caia em exatamente um nível.
	if !(s.ThresholdOtima > s.ThresholdBoa && s.ThresholdBoa > s.ThresholdEstavel && s.ThresholdEstavel > s.ThresholdRuim) {
		writeError(w, http.StatusUnprocessableEntity, "os limites devem ser decrescentes: Ótima > Boa > Estável > Ruim")
		return
	}

	if err := h.repo.SaveSettings(r.Context(), s); err != nil {
		log.Printf("erro ao salvar configuração de saúde financeira: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
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

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
