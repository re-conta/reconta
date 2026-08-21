package taxsim

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
	repo     *Repository
	userRepo *user.Repository
	auth     *auth.Handler
}

func NewHandler(repo *Repository, userRepo *user.Repository, authHandler *auth.Handler) *Handler {
	return &Handler{repo: repo, userRepo: userRepo, auth: authHandler}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tax-simulation", h.auth.RequireUser(h.simulate))
}

func (h *Handler) simulate(w http.ResponseWriter, r *http.Request, userID int64) {
	u, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao ler usuário para simulação de imposto: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !u.TaxSimulationEnabled {
		writeJSON(w, http.StatusOK, Result{Enabled: false})
		return
	}

	year := time.Now().Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1900 || parsed > 3000 {
			writeError(w, http.StatusBadRequest, "ano inválido")
			return
		}
		year = parsed
	}

	income, err := h.repo.TaxableIncome(r.Context(), userID, year)
	if err != nil {
		log.Printf("erro ao calcular receitas tributáveis: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, Compute(year, income))
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
