package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/re-conta/reconta/api/internal/user"
)

const (
	googleStateCookie = "google_oauth_state"
	googleStateTTL    = 10 * time.Minute
	googleUserInfoURL = "https://www.googleapis.com/oauth2/v3/userinfo"
)

// GoogleHandler adiciona o fluxo de login via Google OAuth2 ao Handler de auth.
type GoogleHandler struct {
	auth        *Handler
	users       *user.Repository
	oauthConfig *oauth2.Config
	appURL      string
	afterCreate func(ctx context.Context, userID int64)
}

// NewGoogleHandler cria o handler de OAuth do Google. clientID/clientSecret/redirectURL
// vêm de variáveis de ambiente; appURL é para onde o usuário é redirecionado após o login.
func NewGoogleHandler(authHandler *Handler, users *user.Repository, clientID, clientSecret, redirectURL, appURL string) *GoogleHandler {
	return &GoogleHandler{
		auth:  authHandler,
		users: users,
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
		appURL: appURL,
	}
}

// SetAfterCreate registra um callback executado quando um novo usuário é criado
// através do login via Google (ex.: popular categorias/conta padrão).
func (g *GoogleHandler) SetAfterCreate(fn func(ctx context.Context, userID int64)) {
	g.afterCreate = fn
}

func (g *GoogleHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/auth/google/login", g.login)
	mux.HandleFunc("GET /api/auth/google/callback", g.callback)
}

func (g *GoogleHandler) login(w http.ResponseWriter, r *http.Request) {
	state, err := generateToken()
	if err != nil {
		log.Printf("erro ao gerar state do oauth: %v", err)
		http.Redirect(w, r, g.appURL+"/login?error=oauth", http.StatusFound)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     googleStateCookie,
		Value:    state,
		Path:     "/",
		Expires:  time.Now().Add(googleStateTTL),
		HttpOnly: true,
		Secure:   g.auth.secure,
		SameSite: http.SameSiteLaxMode,
	})

	http.Redirect(w, r, g.oauthConfig.AuthCodeURL(state), http.StatusFound)
}

func (g *GoogleHandler) callback(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(googleStateCookie)
	if err != nil || r.URL.Query().Get("state") != cookie.Value {
		http.Redirect(w, r, g.appURL+"/login?error=oauth_state", http.StatusFound)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Redirect(w, r, g.appURL+"/login?error=oauth_code", http.StatusFound)
		return
	}

	token, err := g.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("erro ao trocar código oauth do Google: %v", err)
		http.Redirect(w, r, g.appURL+"/login?error=oauth_exchange", http.StatusFound)
		return
	}

	info, err := g.fetchUserInfo(r.Context(), token)
	if err != nil {
		log.Printf("erro ao buscar perfil do Google: %v", err)
		http.Redirect(w, r, g.appURL+"/login?error=oauth_profile", http.StatusFound)
		return
	}

	u, err := g.findOrCreateUser(r.Context(), info)
	if err != nil {
		log.Printf("erro ao resolver usuário do login Google: %v", err)
		http.Redirect(w, r, g.appURL+"/login?error=oauth_user", http.StatusFound)
		return
	}

	if err := g.auth.createSession(w, r, u.ID); err != nil {
		log.Printf("erro ao criar sessão para login Google: %v", err)
		http.Redirect(w, r, g.appURL+"/login?error=oauth_session", http.StatusFound)
		return
	}

	http.Redirect(w, r, g.appURL+"/", http.StatusFound)
}

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

func (g *GoogleHandler) fetchUserInfo(ctx context.Context, token *oauth2.Token) (*googleUserInfo, error) {
	client := g.oauthConfig.Client(ctx, token)
	resp, err := client.Get(googleUserInfoURL)
	if err != nil {
		return nil, fmt.Errorf("chamando userinfo do Google: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("lendo resposta do userinfo: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo do Google retornou status %d", resp.StatusCode)
	}

	var info googleUserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("decodificando userinfo do Google: %w", err)
	}
	if info.Sub == "" || info.Email == "" {
		return nil, errors.New("resposta do Google sem sub/email")
	}
	return &info, nil
}

func (g *GoogleHandler) findOrCreateUser(ctx context.Context, info *googleUserInfo) (*user.User, error) {
	if u, err := g.users.GetByGoogleID(ctx, info.Sub); err == nil {
		if err := g.users.UpdateAvatarURL(ctx, u.ID, info.Picture); err != nil {
			return nil, err
		}
		u.AvatarURL = info.Picture
		return u, nil
	} else if !errors.Is(err, user.ErrNotFound) {
		return nil, err
	}

	if u, err := g.users.GetByEmail(ctx, info.Email); err == nil {
		if err := g.users.LinkGoogleID(ctx, u.ID, info.Sub); err != nil {
			return nil, err
		}
		if err := g.users.UpdateAvatarURL(ctx, u.ID, info.Picture); err != nil {
			return nil, err
		}
		u.AvatarURL = info.Picture
		return u, nil
	} else if !errors.Is(err, user.ErrNotFound) {
		return nil, err
	}

	name := info.Name
	if name == "" {
		name = info.Email
	}
	u, err := g.users.CreateWithGoogle(ctx, name, info.Email, info.Sub, info.Picture)
	if err != nil {
		return nil, err
	}
	if g.afterCreate != nil {
		g.afterCreate(ctx, u.ID)
	}
	return u, nil
}
