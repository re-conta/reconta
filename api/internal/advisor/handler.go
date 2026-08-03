package advisor

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/health"
)

type Handler struct {
	repo       *Repository
	healthRepo *health.Repository
	auth       *auth.Handler
	queue      *Queue
}

// queue é nil quando GROQ_API_KEY não está configurada — nesse caso o
// endpoint sempre responde "disabled", sem tocar no banco de dados.
func NewHandler(repo *Repository, healthRepo *health.Repository, authHandler *auth.Handler, queue *Queue) *Handler {
	return &Handler{repo: repo, healthRepo: healthRepo, auth: authHandler, queue: queue}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/financial-health/recommendations", h.auth.RequireUser(h.get))
}

type recommendationsResponse struct {
	Status          string           `json:"status"` // disabled | no_data | pending | ready
	Stars           int              `json:"stars,omitempty"`
	SavingsRate     float64          `json:"savingsRate,omitempty"`
	GeneratedAt     *time.Time       `json:"generatedAt,omitempty"`
	Recommendations []Recommendation `json:"recommendations,omitempty"`
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request, userID int64) {
	if h.queue == nil {
		writeJSON(w, http.StatusOK, recommendationsResponse{Status: "disabled"})
		return
	}

	ctx := r.Context()

	settings, err := h.healthRepo.GetSettings(ctx)
	if err != nil {
		log.Printf("erro ao ler configuração de saúde financeira: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !settings.Enabled {
		writeJSON(w, http.StatusOK, recommendationsResponse{Status: "disabled"})
		return
	}

	now := time.Now()
	month := queryInt(r, "month", int(now.Month()))
	year := queryInt(r, "year", now.Year())
	if month < 1 || month > 12 || year < 1900 || year > 3000 {
		writeError(w, http.StatusBadRequest, "mês ou ano inválido")
		return
	}

	income, expense, err := h.healthRepo.MonthTotals(ctx, userID, month, year)
	if err != nil {
		log.Printf("erro ao calcular totais do mês: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if income == 0 && expense == 0 {
		writeJSON(w, http.StatusOK, recommendationsResponse{Status: "no_data"})
		return
	}

	rate, _, stars := health.Classify(income, expense, settings)

	existing, err := h.repo.Get(ctx, userID, month, year)
	if err != nil {
		log.Printf("erro ao ler recomendações salvas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if needsGeneration(existing, stars, now) {
		h.queue.Enqueue(userID, month, year)
	}

	resp := recommendationsResponse{Stars: stars, SavingsRate: rate}
	switch {
	case existing == nil || existing.Status == "error":
		resp.Status = "pending"
	default:
		resp.Status = "ready"
		resp.Recommendations = existing.Items
		generatedAt := existing.GeneratedAt
		resp.GeneratedAt = &generatedAt
	}
	writeJSON(w, http.StatusOK, resp)
}

// needsGeneration decide se uma nova análise deve ser enfileirada: nunca foi
// gerada, a saúde financeira mudou de nível desde a última análise (novos
// lançamentos podem ter mudado o cenário), a tentativa anterior falhou (tenta
// de novo depois de uma hora) ou já passou mais de 24h desde a última.
func needsGeneration(existing *Record, stars int, now time.Time) bool {
	if existing == nil {
		return true
	}
	if existing.Stars != stars {
		return true
	}
	if existing.Status == "error" {
		return now.Sub(existing.GeneratedAt) > time.Hour
	}
	return now.Sub(existing.GeneratedAt) > 24*time.Hour
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
