package auditlog

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/re-conta/reconta/api/internal/user"
)

// CurrentUserFunc resolve o usuário autenticado a partir da requisição.
// Recebido como função (em vez de *auth.Handler) para evitar dependência
// cíclica: o próprio pacote auth registra ações neste pacote.
type CurrentUserFunc func(r *http.Request) (*user.User, error)

type Handler struct {
	repo        *Repository
	currentUser CurrentUserFunc
}

func NewHandler(repo *Repository, currentUser CurrentUserFunc) *Handler {
	return &Handler{repo: repo, currentUser: currentUser}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admin/logs", h.list)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	u, err := h.currentUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}
	if !u.HasPermission(user.PermAdminPanel) {
		writeError(w, http.StatusForbidden, "acesso negado")
		return
	}

	var userID int64
	if raw := r.URL.Query().Get("userId"); raw != "" {
		userID, _ = strconv.ParseInt(raw, 10, 64)
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if v, err := strconv.Atoi(raw); err == nil && v > 0 {
			limit = v
		}
	}

	entries, err := h.repo.List(r.Context(), userID, limit)
	if err != nil {
		log.Printf("erro ao carregar log de ações: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if entries == nil {
		entries = []LogEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
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
