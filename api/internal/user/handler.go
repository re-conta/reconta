package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/re-conta/reconta/api/internal/email"
	"github.com/re-conta/reconta/api/internal/turnstile"
)

// currentUserFunc resolve o usuário autenticado a partir da requisição. É
// fornecida pelo pacote auth via SetAuth, evitando um ciclo de import (auth
// já depende de user para o tipo User).
type currentUserFunc func(r *http.Request) (*User, error)

// loginFunc autentica o usuário recém-confirmado, gravando a sessão como
// cookie na resposta. Fornecida pelo pacote auth via SetLogin, evitando um
// ciclo de import.
type loginFunc func(w http.ResponseWriter, r *http.Request, userID int64) error

type Handler struct {
	repo          *Repository
	currentUser   currentUserFunc
	afterCreate   func(ctx context.Context, userID int64)
	onBan         func(ctx context.Context, userID int64)
	turnstile     *turnstile.Verifier
	mail          *email.Queue
	appURL        string
	login         loginFunc
	internalToken string
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

// requirePermission envolve um handler exigindo que o usuário autenticado
// possua a permissão informada. O Super Admin sempre passa.
func (h *Handler) requirePermission(next func(w http.ResponseWriter, r *http.Request, u *User), perm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.currentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		if !u.HasPermission(perm) {
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
	mux.HandleFunc("POST /api/users/verify-otp", h.verifyOTP)
	mux.HandleFunc("POST /api/users/resend-otp", h.resendOTP)
	mux.HandleFunc("POST /api/internal/users/scan", h.requireInternalToken(h.scan))
	mux.HandleFunc("GET /api/users", h.requirePermission(h.list, PermAdminPanel))
	mux.HandleFunc("PATCH /api/users/{id}/role", h.requirePermission(h.updateRole, PermManageUsers))
	mux.HandleFunc("PATCH /api/users/me", h.requireAuth(h.updateProfile))
	mux.HandleFunc("PATCH /api/users/me/password", h.requireAuth(h.updatePassword))
	mux.HandleFunc("GET /api/admin/permissions", h.requirePermission(h.listRolePermissions, PermAdminPanel))
	mux.HandleFunc("PUT /api/admin/permissions/{role}", h.requirePermission(h.updateRolePermissions, PermManagePermissions))

	mux.HandleFunc("POST /api/admin/users", h.requirePermission(h.adminCreateUser, PermManageUsers))
	mux.HandleFunc("PATCH /api/admin/users/{id}", h.requirePermission(h.adminUpdateUser, PermManageUsers))
	mux.HandleFunc("DELETE /api/admin/users/{id}", h.requirePermission(h.adminDeleteUser, PermManageUsers))
	mux.HandleFunc("PATCH /api/admin/users/{id}/ban", h.requirePermission(h.adminBanUser, PermManageUsers))
}

// SetOnBan registra um callback executado após um usuário ser banido, usado
// para encerrar imediatamente as sessões ativas dele. Deve ser chamado antes
// de RegisterRoutes.
func (h *Handler) SetOnBan(fn func(ctx context.Context, userID int64)) {
	h.onBan = fn
}

// SetTurnstile registra o verificador do Cloudflare Turnstile usado para
// proteger o cadastro público contra bots. Deve ser chamado antes de
// RegisterRoutes.
func (h *Handler) SetTurnstile(v *turnstile.Verifier) {
	h.turnstile = v
}

// SetMail registra a fila de e-mail e a URL base do front-end, usadas para
// enviar o código de confirmação do cadastro (OTP). Deve ser chamada antes de
// RegisterRoutes. Se não for chamada, o código apenas será registrado em log
// (ver email.Mailer).
func (h *Handler) SetMail(mail *email.Queue, appURL string) {
	h.mail = mail
	h.appURL = appURL
}

// SetLogin registra a função que autentica o usuário logo após a confirmação
// do código OTP de cadastro. Deve ser chamada antes de RegisterRoutes.
func (h *Handler) SetLogin(fn loginFunc) {
	h.login = fn
}

// SetInternalToken registra o token usado para proteger a rota de varredura
// interna que encerra cadastros pendentes cujo prazo de confirmação expirou.
// Deve ser chamada antes de RegisterRoutes.
func (h *Handler) SetInternalToken(token string) {
	h.internalToken = token
}

// clientIP resolve o IP real do visitante: primeiro o cabeçalho que o
// Cloudflare injeta na borda (CF-Connecting-IP), depois X-Real-IP (setado
// pelo Nginx), e por fim o endereço da conexão TCP.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// canManageTarget impede que um admin comum (não Super Admin) altere,
// bane ou exclua outro admin ou o próprio Super Admin — reservado ao Super
// Admin, mesmo que o ator tenha a permissão de gerenciar usuários.
func canManageTarget(actor, target *User) bool {
	if actor.Role == RoleSuperAdmin {
		return true
	}
	return target.Role != RoleAdmin && target.Role != RoleSuperAdmin
}

type createUserRequest struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	Password       string `json:"password"`
	Role           string `json:"role"`
	CNPJ           string `json:"cnpj"`
	TurnstileToken string `json:"turnstileToken"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if h.turnstile != nil && h.turnstile.Enabled() {
		ok, err := h.turnstile.Verify(r.Context(), req.TurnstileToken, clientIP(r))
		if err != nil {
			log.Printf("erro ao verificar turnstile: %v", err)
			writeError(w, http.StatusServiceUnavailable, "não foi possível verificar o captcha, tente novamente")
			return
		}
		if !ok {
			writeError(w, http.StatusUnprocessableEntity, "falha na verificação anti-robô, tente novamente")
			return
		}
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

	if req.Role == "" {
		req.Role = RolePessoaFisica
	}
	if !slices.Contains(SignupRoles, req.Role) {
		writeError(w, http.StatusUnprocessableEntity, "tipo de conta inválido")
		return
	}

	req.CNPJ = NormalizeCNPJ(req.CNPJ)
	if req.Role == RolePessoaJuridica {
		if !IsValidCNPJ(req.CNPJ) {
			writeError(w, http.StatusUnprocessableEntity, "CNPJ inválido")
			return
		}
	} else {
		req.CNPJ = ""
	}

	if _, err := h.repo.GetByEmail(r.Context(), req.Email); err == nil {
		writeError(w, http.StatusConflict, "e-mail já cadastrado")
		return
	} else if !errors.Is(err, ErrNotFound) {
		log.Printf("erro ao verificar e-mail existente: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("erro ao gerar hash de senha: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	code, err := h.repo.CreatePendingSignup(r.Context(), req.Name, req.Email, string(hash), req.Role, req.CNPJ)
	if err != nil {
		log.Printf("erro ao criar cadastro pendente: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	h.sendOTPEmail(req.Email, req.Name, code)

	writeJSON(w, http.StatusAccepted, pendingSignupResponse{
		Email:            req.Email,
		ExpiresInMinutes: int(otpTTL.Minutes()),
	})
}

type pendingSignupResponse struct {
	Email            string `json:"email"`
	ExpiresInMinutes int    `json:"expiresInMinutes"`
}

// sendOTPEmail envia o código de confirmação de cadastro. Se a fila de e-mail
// não estiver configurada (SetMail nunca chamado), apenas registra em log.
func (h *Handler) sendOTPEmail(to, name, code string) {
	if h.mail == nil {
		log.Printf("código OTP para %s: %s (fila de e-mail não configurada)", to, code)
		return
	}
	msg := email.Message{
		Preheader: "Confirme seu e-mail para concluir o cadastro no ReConta.",
		Heading:   fmt.Sprintf("Olá, %s!", name),
		Paragraphs: []string{
			"Use o código abaixo para confirmar seu e-mail e concluir seu cadastro no ReConta.",
			fmt.Sprintf("Código de confirmação: %s", code),
			fmt.Sprintf("O código é válido por %d minutos. Se você não solicitou este cadastro, pode ignorar este e-mail.", int(otpTTL.Minutes())),
		},
		Footnote: "Perdeu o código? Você pode solicitar um novo na própria tela de cadastro.",
	}
	h.mail.Enqueue(to, "Confirme seu cadastro - ReConta", msg)
}

type verifyOTPRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

// verifyOTP confirma o código enviado por e-mail e, se válido, cria a conta
// efetivamente e autentica o usuário.
func (h *Handler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var req verifyOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	code := strings.TrimSpace(req.Code)

	u, err := h.repo.VerifyPendingSignup(r.Context(), email, code)
	if err != nil {
		switch {
		case errors.Is(err, ErrPendingNotFound):
			writeError(w, http.StatusNotFound, "cadastro pendente não encontrado, refaça o cadastro")
		case errors.Is(err, ErrOTPExpired):
			writeError(w, http.StatusGone, "código expirado, refaça o cadastro")
		case errors.Is(err, ErrOTPLocked):
			writeError(w, http.StatusTooManyRequests, "número de tentativas excedido, solicite um novo código")
		case errors.Is(err, ErrOTPInvalid):
			writeError(w, http.StatusUnprocessableEntity, "código inválido")
		case errors.Is(err, ErrEmailTaken):
			writeError(w, http.StatusConflict, "e-mail já cadastrado")
		default:
			log.Printf("erro ao confirmar cadastro pendente: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
		}
		return
	}

	if h.afterCreate != nil {
		h.afterCreate(r.Context(), u.ID)
	}
	if h.login != nil {
		if err := h.login(w, r, u.ID); err != nil {
			log.Printf("erro ao autenticar usuário após confirmação de cadastro: %v", err)
		}
	}

	writeJSON(w, http.StatusCreated, u)
}

type resendOTPRequest struct {
	Email string `json:"email"`
}

// resendOTP gera e envia um novo código de confirmação para um cadastro
// pendente existente, usado quando o visitante perde o código original.
// Sempre responde 204, mesmo quando o e-mail não tem cadastro pendente, para
// não revelar quais e-mails têm cadastro em andamento.
func (h *Handler) resendOTP(w http.ResponseWriter, r *http.Request) {
	var req resendOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	addr := strings.ToLower(strings.TrimSpace(req.Email))

	code, err := h.repo.ResendPendingOTP(r.Context(), addr)
	if err != nil {
		if errors.Is(err, ErrResendTooSoon) {
			writeError(w, http.StatusTooManyRequests, "aguarde um pouco antes de solicitar um novo código")
			return
		}
		if !errors.Is(err, ErrPendingNotFound) && !errors.Is(err, ErrOTPExpired) {
			log.Printf("erro ao reenviar código OTP: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	h.sendOTPEmail(addr, addr, code)
	w.WriteHeader(http.StatusNoContent)
}

// requireInternalToken protege a rota de varredura interna, chamada apenas
// pelo timer systemd (não usa sessão de usuário).
func (h *Handler) requireInternalToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.internalToken == "" || r.Header.Get("X-Internal-Token") != h.internalToken {
			writeError(w, http.StatusUnauthorized, "não autorizado")
			return
		}
		next(w, r)
	}
}

// scan remove cadastros pendentes cujo prazo de confirmação (2h) expirou sem
// que o código OTP tivesse sido validado.
func (h *Handler) scan(w http.ResponseWriter, r *http.Request) {
	n, err := h.repo.DeleteExpiredPendingSignups(r.Context())
	if err != nil {
		log.Printf("erro ao remover cadastros pendentes expirados: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if n > 0 {
		log.Printf("%d cadastro(s) pendente(s) expirado(s) removido(s)", n)
	}
	w.WriteHeader(http.StatusNoContent)
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

func (h *Handler) updateRole(w http.ResponseWriter, r *http.Request, actor *User) {
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

	// Promover alguém a admin (ou rebaixar um admin) é reservado ao Super
	// Admin, mesmo que a role do ator tenha a permissão de gerenciar usuários.
	if actor.Role != RoleSuperAdmin {
		target, err := h.repo.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "usuário não encontrado")
				return
			}
			log.Printf("erro ao buscar usuário: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
		if req.Role == RoleAdmin || target.Role == RoleAdmin {
			writeError(w, http.StatusForbidden, "apenas o Super Admin pode promover ou rebaixar administradores")
			return
		}
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

type adminCreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	CNPJ     string `json:"cnpj"`
}

// adminCreateUser cria uma conta com a role escolhida diretamente pelo
// administrador. Promover alguém direto para admin é reservado ao Super Admin.
func (h *Handler) adminCreateUser(w http.ResponseWriter, r *http.Request, actor *User) {
	var req adminCreateUserRequest
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
	if !slices.Contains(AssignableRoles, req.Role) {
		writeError(w, http.StatusUnprocessableEntity, "cargo inválido")
		return
	}
	if req.Role == RoleAdmin && actor.Role != RoleSuperAdmin {
		writeError(w, http.StatusForbidden, "apenas o Super Admin pode criar administradores")
		return
	}

	req.CNPJ = NormalizeCNPJ(req.CNPJ)
	if req.Role == RolePessoaJuridica {
		if !IsValidCNPJ(req.CNPJ) {
			writeError(w, http.StatusUnprocessableEntity, "CNPJ inválido")
			return
		}
	} else {
		req.CNPJ = ""
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("erro ao gerar hash de senha: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	u, err := h.repo.AdminCreate(r.Context(), req.Name, req.Email, string(hash), req.Role, req.CNPJ)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, http.StatusConflict, "e-mail já cadastrado")
			return
		}
		log.Printf("erro ao criar usuário via admin: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if h.afterCreate != nil {
		h.afterCreate(r.Context(), u.ID)
	}

	writeJSON(w, http.StatusCreated, u)
}

type adminUpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	CNPJ  string `json:"cnpj"`
}

func (h *Handler) adminUpdateUser(w http.ResponseWriter, r *http.Request, actor *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	target, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		log.Printf("erro ao buscar usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !canManageTarget(actor, target) {
		writeError(w, http.StatusForbidden, "apenas o Super Admin pode editar este usuário")
		return
	}

	var req adminUpdateUserRequest
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

	req.CNPJ = NormalizeCNPJ(req.CNPJ)
	if req.CNPJ != "" && !IsValidCNPJ(req.CNPJ) {
		writeError(w, http.StatusUnprocessableEntity, "CNPJ inválido")
		return
	}

	updated, err := h.repo.AdminUpdate(r.Context(), id, req.Name, req.Email, req.CNPJ)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, http.StatusConflict, "e-mail já cadastrado")
			return
		}
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		log.Printf("erro ao atualizar usuário via admin: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) adminDeleteUser(w http.ResponseWriter, r *http.Request, actor *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	if id == actor.ID {
		writeError(w, http.StatusForbidden, "você não pode excluir sua própria conta por aqui")
		return
	}

	target, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		log.Printf("erro ao buscar usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !canManageTarget(actor, target) {
		writeError(w, http.StatusForbidden, "apenas o Super Admin pode excluir este usuário")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if errors.Is(err, ErrProtectedUser) {
			writeError(w, http.StatusForbidden, "não é possível excluir esta conta")
			return
		}
		log.Printf("erro ao excluir usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type adminBanUserRequest struct {
	Banned bool `json:"banned"`
}

func (h *Handler) adminBanUser(w http.ResponseWriter, r *http.Request, actor *User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	if id == actor.ID {
		writeError(w, http.StatusForbidden, "você não pode banir sua própria conta")
		return
	}

	var req adminBanUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	target, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "usuário não encontrado")
			return
		}
		log.Printf("erro ao buscar usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !canManageTarget(actor, target) {
		writeError(w, http.StatusForbidden, "apenas o Super Admin pode banir este usuário")
		return
	}

	updated, err := h.repo.BanUser(r.Context(), id, req.Banned)
	if err != nil {
		if errors.Is(err, ErrProtectedUser) {
			writeError(w, http.StatusForbidden, "não é possível banir esta conta")
			return
		}
		log.Printf("erro ao atualizar banimento do usuário: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if req.Banned && h.onBan != nil {
		h.onBan(r.Context(), id)
	}

	writeJSON(w, http.StatusOK, updated)
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

// rolePermissionsResponse descreve as permissões editáveis do painel de admin:
// as roles atribuíveis, as permissões disponíveis e o que cada role possui.
type rolePermissionsResponse struct {
	Roles       []string            `json:"roles"`
	Available   []string            `json:"available"`
	Permissions map[string][]string `json:"permissions"`
}

func (h *Handler) listRolePermissions(w http.ResponseWriter, r *http.Request, _ *User) {
	perms, err := h.repo.PermissionsByRole(r.Context())
	if err != nil {
		log.Printf("erro ao listar permissões: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, rolePermissionsResponse{
		Roles:       AssignableRoles,
		Available:   AllPermissions,
		Permissions: perms,
	})
}

type updateRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

func (h *Handler) updateRolePermissions(w http.ResponseWriter, r *http.Request, _ *User) {
	role := r.PathValue("role")
	if !slices.Contains(AssignableRoles, role) {
		writeError(w, http.StatusUnprocessableEntity, "role inválida")
		return
	}

	var req updateRolePermissionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	perms := []string{}
	for _, p := range req.Permissions {
		if !slices.Contains(AllPermissions, p) {
			writeError(w, http.StatusUnprocessableEntity, "permissão inválida: "+p)
			return
		}
		if !slices.Contains(perms, p) {
			perms = append(perms, p)
		}
	}

	if err := h.repo.SetRolePermissions(r.Context(), role, perms); err != nil {
		log.Printf("erro ao atualizar permissões da role: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string][]string{"permissions": perms})
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
