package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/lucasbrum/reconta/api/internal/user"
)

const (
	cookieName = "session_token"
	sessionTTL = 7 * 24 * time.Hour
)

type Handler struct {
	sessions *Repository
	users    *user.Repository
	secure   bool
}

// NewHandler cria o handler de autenticação. secure define se o cookie de
// sessão deve ser marcado como Secure (deve ser true em produção, atrás de HTTPS).
func NewHandler(sessions *Repository, users *user.Repository, secure bool) *Handler {
	return &Handler{sessions: sessions, users: users, secure: secure}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/logout", h.logout)
	mux.HandleFunc("GET /api/auth/me", h.me)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	u, passwordHash, err := h.users.GetByEmailWithPasswordHash(r.Context(), email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			writeError(w, http.StatusUnauthorized, "e-mail ou senha inválidos")
			return
		}
		log.Printf("erro ao buscar usuário para login: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "e-mail ou senha inválidos")
		return
	}

	token, err := generateToken()
	if err != nil {
		log.Printf("erro ao gerar token de sessão: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	expiresAt := time.Now().Add(sessionTTL)
	if err := h.sessions.Create(r.Context(), token, u.ID, expiresAt); err != nil {
		log.Printf("erro ao criar sessão: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})

	writeJSON(w, http.StatusOK, u)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(cookieName); err == nil {
		_ = h.sessions.Delete(r.Context(), cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	u, err := h.CurrentUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "não autenticado")
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// CurrentUser resolve o usuário autenticado a partir do cookie de sessão da
// requisição. Pode ser reutilizado por outros handlers que exijam autenticação.
func (h *Handler) CurrentUser(r *http.Request) (*user.User, error) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return nil, ErrSessionNotFound
	}

	session, err := h.sessions.GetByToken(r.Context(), cookie.Value)
	if err != nil {
		return nil, err
	}

	return h.users.GetByID(r.Context(), session.UserID)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
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
