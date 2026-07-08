package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/re-conta/reconta/api/internal/email"
	"github.com/re-conta/reconta/api/internal/user"
)

const (
	cookieName    = "session_token"
	sessionTTL    = 7 * 24 * time.Hour
	resetTokenTTL = 1 * time.Hour
)

type Handler struct {
	sessions *Repository
	users    *user.Repository
	secure   bool
	mail     *email.Queue
	appURL   string
}

// NewHandler cria o handler de autenticação. secure define se o cookie de
// sessão deve ser marcado como Secure (deve ser true em produção, atrás de HTTPS).
func NewHandler(sessions *Repository, users *user.Repository, secure bool) *Handler {
	return &Handler{sessions: sessions, users: users, secure: secure}
}

// SetMail registra a fila de e-mail e a URL base do front-end, usadas para
// enviar o link de redefinição de senha. Deve ser chamado antes de
// RegisterRoutes. Se não for chamado, a rota de "esqueci minha senha" fica
// habilitada mas os e-mails apenas serão registrados em log (ver email.Mailer).
func (h *Handler) SetMail(mail *email.Queue, appURL string) {
	h.mail = mail
	h.appURL = appURL
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", h.login)
	mux.HandleFunc("POST /api/auth/logout", h.logout)
	mux.HandleFunc("GET /api/auth/me", h.me)
	mux.HandleFunc("POST /api/auth/forgot-password", h.forgotPassword)
	mux.HandleFunc("POST /api/auth/reset-password", h.resetPassword)
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
	password := strings.TrimSpace(req.Password)

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

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		writeError(w, http.StatusUnauthorized, "e-mail ou senha inválidos")
		return
	}

	if err := h.createSession(w, r, u.ID); err != nil {
		log.Printf("erro ao criar sessão: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	writeJSON(w, http.StatusOK, u)
}

// createSession gera um token de sessão para o usuário informado e o grava
// como cookie na resposta. Reutilizado pelo login por e-mail/senha e pelo
// callback do Google OAuth.
func (h *Handler) createSession(w http.ResponseWriter, r *http.Request, userID int64) error {
	token, err := generateToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(sessionTTL)
	if err := h.sessions.Create(r.Context(), token, userID, expiresAt); err != nil {
		return err
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
	return nil
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

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

// forgotPassword gera um token de redefinição de senha e envia um e-mail com
// o link para o usuário, caso o e-mail exista. Sempre responde 204, mesmo
// quando o e-mail não é encontrado, para não revelar quais e-mails estão
// cadastrados.
func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))

	u, err := h.users.GetByEmail(r.Context(), email)
	if err != nil {
		if !errors.Is(err, user.ErrNotFound) {
			log.Printf("erro ao buscar usuário para redefinição de senha: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	token, err := generateToken()
	if err != nil {
		log.Printf("erro ao gerar token de redefinição de senha: %v", err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.sessions.CreateResetToken(r.Context(), token, u.ID, time.Now().Add(resetTokenTTL)); err != nil {
		log.Printf("erro ao criar token de redefinição de senha: %v", err)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if h.mail != nil {
		link := fmt.Sprintf("%s/redefinir-senha?token=%s", strings.TrimRight(h.appURL, "/"), token)
		body := fmt.Sprintf(
			"Olá, %s!\n\nRecebemos uma solicitação para redefinir sua senha no Reconta.\n\nClique no link abaixo para criar uma nova senha (válido por 1 hora):\n%s\n\nSe você não solicitou isso, pode ignorar este e-mail.",
			u.Name, link,
		)
		h.mail.Enqueue(u.Email, "Redefinição de senha - Reconta", body)
	}

	w.WriteHeader(http.StatusNoContent)
}

type resetPasswordRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

// resetPassword define uma nova senha a partir de um token válido gerado por
// forgotPassword.
func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	password := strings.TrimSpace(req.Password)
	if len(password) < 8 {
		writeError(w, http.StatusUnprocessableEntity, "senha deve ter no mínimo 8 caracteres")
		return
	}

	userID, err := h.sessions.GetResetToken(r.Context(), req.Token)
	if err != nil {
		if errors.Is(err, ErrResetTokenNotFound) {
			writeError(w, http.StatusUnprocessableEntity, "link de redefinição inválido ou expirado")
			return
		}
		log.Printf("erro ao buscar token de redefinição de senha: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("erro ao gerar hash de senha: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if err := h.users.UpdatePassword(r.Context(), userID, string(hash)); err != nil {
		log.Printf("erro ao atualizar senha do usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	_ = h.sessions.DeleteResetToken(r.Context(), req.Token)

	w.WriteHeader(http.StatusNoContent)
}

// me informa o usuário autenticado atual. Retorna sempre 200: o front-end
// verifica ausência de sessão em toda carga de página, então tratar isso como
// 401 gera ruído constante no console/network do navegador para um caso
// esperado (usuário ainda não logado).
func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	u, err := h.CurrentUser(r)
	if err != nil {
		writeJSON(w, http.StatusOK, nil)
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

// RequireUser envolve um handler que precisa do usuário autenticado, resolvendo-o
// a partir da sessão e respondendo 401 automaticamente quando ausente/inválida.
func (h *Handler) RequireUser(next func(w http.ResponseWriter, r *http.Request, userID int64)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.CurrentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		next(w, r, u.ID)
	}
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
