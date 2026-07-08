package user

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// currentUserFunc resolve o usuário autenticado a partir da requisição. É
// fornecida pelo pacote auth via SetAuth, evitando um ciclo de import (auth
// já depende de user para o tipo User).
type currentUserFunc func(r *http.Request) (*User, error)

type Handler struct {
	repo        *Repository
	currentUser currentUserFunc
	afterCreate func(ctx context.Context, userID int64)
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// SetAfterCreate registra um callback executado logo após a criação de um
// usuário (ex.: popular categorias/conta padrão). Deve ser chamado antes de
// RegisterRoutes.
func (h *Handler) SetAfterCreate(fn func(ctx context.Context, userID int64)) {
	h.afterCreate = fn
}

// SetAuth registra a função de resolução do usuário autenticado, usada para
// proteger as rotas que exigem role de admin/super_admin. Deve ser chamada
// antes de RegisterRoutes.
func (h *Handler) SetAuth(fn func(r *http.Request) (*User, error)) {
	h.currentUser = fn
}

// requireRole envolve um handler exigindo que o usuário autenticado tenha uma
// das roles informadas.
func (h *Handler) requireRole(next func(w http.ResponseWriter, r *http.Request, u *User), roles ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.currentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		if !slices.Contains(roles, u.Role) {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
		next(w, r, u)
	}
}

// requireAuth envolve um handler exigindo apenas que exista um usuário
// autenticado, sem restrição de role.
func (h *Handler) requireAuth(next func(w http.ResponseWriter, r *http.Request, u *User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.currentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		next(w, r, u)
	}
}

// RegisterRoutes registra as rotas de usuário no mux informado.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/users", h.create)
	mux.HandleFunc("GET /api/users", h.requireRole(h.list, RoleAdmin, RoleSuperAdmin))
	mux.HandleFunc("PATCH /api/users/{id}/role", h.requireRole(h.updateRole, RoleSuperAdmin))
	mux.HandleFunc("PATCH /api/users/me", h.requireAuth(h.updateProfile))
	mux.HandleFunc("PATCH /api/users/me/password", h.requireAuth(h.updatePassword))
}

type createUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Password = strings.TrimSpace(req.Password)

	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "nome é obrigatório")
		return
	}
	if !isValidEmail(req.Email) {
		writeError(w, http.StatusUnprocessableEntity, "e-mail inválido")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "senha deve ter no mínimo 8 caracteres")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("erro ao gerar hash de senha: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	u, err := h.repo.Create(r.Context(), req.Name, req.Email, string(hash))
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, http.StatusConflict, "e-mail já cadastrado")
			return
		}
		log.Printf("erro ao criar usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if h.afterCreate != nil {
		h.afterCreate(r.Context(), u.ID)
	}

	writeJSON(w, http.StatusCreated, u)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, _ *User) {
	users, err := h.repo.List(r.Context())
	if err != nil {
		log.Printf("erro ao listar usuários: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

type updateRoleRequest struct {
	Role string `json:"role"`
}

func (h *Handler) updateRole(w http.ResponseWriter, r *http.Request, _ *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req updateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	u, err := h.repo.UpdateRole(r.Context(), id, req.Role)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		if errors.Is(err, ErrCannotModifyRole) {
			writeError(w, http.StatusForbidden, "não é possível alterar a role deste usuário")
			return
		}
		log.Printf("erro ao atualizar role do usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

type updateProfileRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (h *Handler) updateProfile(w http.ResponseWriter, r *http.Request, u *User) {
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "nome é obrigatório")
		return
	}
	if !isValidEmail(req.Email) {
		writeError(w, http.StatusUnprocessableEntity, "e-mail inválido")
		return
	}

	updated, err := h.repo.UpdateProfile(r.Context(), u.ID, req.Name, req.Email)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, http.StatusConflict, "e-mail já cadastrado")
			return
		}
		log.Printf("erro ao atualizar perfil do usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

type updatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (h *Handler) updatePassword(w http.ResponseWriter, r *http.Request, u *User) {
	var req updatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "senha deve ter no mínimo 8 caracteres")
		return
	}

	currentHash, err := h.repo.GetPasswordHashByID(r.Context(), u.ID)
	if err != nil {
		log.Printf("erro ao buscar senha do usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	// Usuários cadastrados apenas via Google não têm senha ainda: a primeira
	// definição de senha não exige a senha atual.
	if currentHash != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.CurrentPassword)); err != nil {
			writeError(w, http.StatusUnprocessableEntity, "senha atual inválida")
			return
		}
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("erro ao gerar hash de senha: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if err := h.repo.UpdatePassword(r.Context(), u.ID, string(newHash)); err != nil {
		log.Printf("erro ao atualizar senha do usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func isValidEmail(email string) bool {
	at := strings.IndexByte(email, '@')
	return at > 0 && at < len(email)-1 && !strings.Contains(email[at+1:], "@") && strings.Contains(email[at+1:], ".")
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
