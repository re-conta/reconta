package tag

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/re-conta/reconta/api/internal/auth"
)

type Handler struct {
	repo *Repository
	auth *auth.Handler
}

func NewHandler(repo *Repository, authHandler *auth.Handler) *Handler {
	return &Handler{repo: repo, auth: authHandler}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tags", h.auth.RequireUser(h.list))
	mux.HandleFunc("POST /api/tags", h.auth.RequireUser(h.create))
	mux.HandleFunc("PUT /api/tags/{id}", h.auth.RequireUser(h.update))
	mux.HandleFunc("DELETE /api/tags/{id}", h.auth.RequireUser(h.delete))
}

type tagRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, userID int64) {
	tags, err := h.repo.List(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao listar tags: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, tags)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, userID int64) {
	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "nome é obrigatório")
		return
	}
	if req.Color == "" {
		req.Color = "#6366f1"
	}

	t, err := h.repo.Create(r.Context(), userID, req.Name, req.Color)
	if err != nil {
		log.Printf("erro ao criar tag: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req tagRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	t, err := h.repo.Update(r.Context(), userID, id, req.Name, req.Color)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrado")
			return
		}
		log.Printf("erro ao atualizar tag: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, t)
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
		log.Printf("erro ao remover tag: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
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
