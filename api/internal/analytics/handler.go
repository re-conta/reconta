package analytics

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/user"
)

const (
	visitorCookie = "rc_vid"
	sessionCookie = "rc_sid"
	visitorTTL    = 365 * 24 * time.Hour
	maxPathLen    = 512
)

type Handler struct {
	repo   *Repository
	auth   *auth.Handler
	geo    *GeoIP
	secure bool
}

// NewHandler cria o handler de analytics. secure define se os cookies de
// rastreamento devem ser marcados como Secure (true em produção, atrás de HTTPS).
func NewHandler(repo *Repository, authHandler *auth.Handler, geo *GeoIP, secure bool) *Handler {
	return &Handler{repo: repo, auth: authHandler, geo: geo, secure: secure}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/track", h.track)

	mux.HandleFunc("GET /api/admin/analytics/overview", h.requirePermission(h.overview))
	mux.HandleFunc("GET /api/admin/analytics/pages", h.requirePermission(h.pages))
	mux.HandleFunc("GET /api/admin/analytics/referrers", h.requirePermission(h.referrers))
	mux.HandleFunc("GET /api/admin/analytics/locations", h.requirePermission(h.locations))
	mux.HandleFunc("GET /api/admin/analytics/devices", h.requirePermission(h.devices))
	mux.HandleFunc("GET /api/admin/analytics/visitors", h.requirePermission(h.visitors))
	mux.HandleFunc("GET /api/admin/analytics/active", h.requirePermission(h.active))
}

func (h *Handler) requirePermission(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.auth.CurrentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		if !u.HasPermission(user.PermAdminPanel) {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
		next(w, r)
	}
}

type trackRequest struct {
	Path     string `json:"path"`
	Referrer string `json:"referrer"`
}

// track registra uma navegação de página do front-end (SPA — não há esse
// sinal em nenhum log de servidor). Sempre responde 204: nunca deve
// atrapalhar a navegação do usuário, mesmo em erro.
func (h *Handler) track(w http.ResponseWriter, r *http.Request) {
	var req trackRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	req.Path = strings.TrimSpace(req.Path)
	if req.Path == "" || len(req.Path) > maxPathLen {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	visitorID := h.ensureCookie(w, r, visitorCookie, visitorTTL)
	sessionID := h.ensureCookie(w, r, sessionCookie, 0)

	ip := clientIP(r)
	ua := parseUA(r.UserAgent())
	geo := h.geo.Lookup(ip)

	var userID *int64
	if u, err := h.auth.CurrentUser(r); err == nil {
		userID = &u.ID
	}

	var lat, lon *float64
	if geo.Latitude != 0 || geo.Longitude != 0 {
		lat, lon = &geo.Latitude, &geo.Longitude
	}

	visit := Visit{
		VisitorID:      visitorID,
		SessionID:      sessionID,
		UserID:         userID,
		Path:           req.Path,
		Referrer:       strings.TrimSpace(req.Referrer),
		IP:             ip,
		Country:        geo.Country,
		Region:         geo.Region,
		City:           geo.City,
		Latitude:       lat,
		Longitude:      lon,
		UserAgent:      r.UserAgent(),
		Browser:        ua.Browser,
		BrowserVersion: ua.BrowserVersion,
		OS:             ua.OS,
		DeviceType:     ua.DeviceType,
		IsBot:          ua.IsBot,
	}

	if err := h.repo.InsertVisit(r.Context(), visit); err != nil {
		log.Printf("erro ao registrar visita: %v", err)
	}

	w.WriteHeader(http.StatusNoContent)
}

// ensureCookie lê o cookie informado ou gera e grava um novo valor aleatório.
// ttl == 0 grava um cookie de sessão (sem Max-Age/Expires).
func (h *Handler) ensureCookie(w http.ResponseWriter, r *http.Request, name string, ttl time.Duration) string {
	if cookie, err := r.Cookie(name); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	value := generateID()
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
	}
	if ttl > 0 {
		cookie.Expires = time.Now().Add(ttl)
	}
	http.SetCookie(w, cookie)
	return value
}

func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
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

func dateRange(r *http.Request) (from, to time.Time) {
	now := time.Now().UTC()
	to = now.AddDate(0, 0, 1).Truncate(24 * time.Hour)
	from = to.AddDate(0, 0, -30)

	if raw := r.URL.Query().Get("from"); raw != "" {
		if parsed, err := time.Parse("2006-01-02", raw); err == nil {
			from = parsed
		}
	}
	if raw := r.URL.Query().Get("to"); raw != "" {
		if parsed, err := time.Parse("2006-01-02", raw); err == nil {
			to = parsed.AddDate(0, 0, 1)
		}
	}
	return from, to
}

func queryIntDefault(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}

func (h *Handler) overview(w http.ResponseWriter, r *http.Request) {
	from, to := dateRange(r)
	data, err := h.repo.Overview(r.Context(), from, to)
	if err != nil {
		log.Printf("erro ao calcular visão geral de analytics: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) pages(w http.ResponseWriter, r *http.Request) {
	from, to := dateRange(r)
	data, err := h.repo.TopPages(r.Context(), from, to, queryIntDefault(r, "limit", 20))
	if err != nil {
		log.Printf("erro ao calcular páginas mais visitadas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) referrers(w http.ResponseWriter, r *http.Request) {
	from, to := dateRange(r)
	data, err := h.repo.TopReferrers(r.Context(), from, to, queryIntDefault(r, "limit", 20))
	if err != nil {
		log.Printf("erro ao calcular referrers: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) locations(w http.ResponseWriter, r *http.Request) {
	from, to := dateRange(r)
	data, err := h.repo.TopLocations(r.Context(), from, to, queryIntDefault(r, "limit", 20))
	if err != nil {
		log.Printf("erro ao calcular localizações: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) devices(w http.ResponseWriter, r *http.Request) {
	from, to := dateRange(r)
	data, err := h.repo.DeviceBreakdown(r.Context(), from, to)
	if err != nil {
		log.Printf("erro ao calcular dispositivos: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) visitors(w http.ResponseWriter, r *http.Request) {
	from, to := dateRange(r)
	data, err := h.repo.RecentVisits(r.Context(), from, to, queryIntDefault(r, "limit", 50))
	if err != nil {
		log.Printf("erro ao carregar visitas recentes: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) active(w http.ResponseWriter, r *http.Request) {
	count, err := h.repo.ActiveNow(r.Context())
	if err != nil {
		log.Printf("erro ao calcular visitantes ativos: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"active": count})
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
