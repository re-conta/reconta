package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/email"
	"github.com/re-conta/reconta/api/internal/fixedbill"
	"github.com/re-conta/reconta/api/internal/user"
)

type Handler struct {
	repo          *Repository
	auth          *auth.Handler
	hub           *Hub
	bills         *fixedbill.Repository
	users         *user.Repository
	mailQueue     *email.Queue
	internalToken string
}

func NewHandler(repo *Repository, authHandler *auth.Handler, hub *Hub, bills *fixedbill.Repository, users *user.Repository, mailQueue *email.Queue, internalToken string) *Handler {
	return &Handler{repo: repo, auth: authHandler, hub: hub, bills: bills, users: users, mailQueue: mailQueue, internalToken: internalToken}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/notifications", h.auth.RequireUser(h.list))
	mux.HandleFunc("GET /api/notifications/unread-count", h.auth.RequireUser(h.unreadCount))
	mux.HandleFunc("GET /api/notifications/stream", h.auth.RequireUser(h.stream))
	mux.HandleFunc("POST /api/notifications/{id}/read", h.auth.RequireUser(h.markRead))
	mux.HandleFunc("POST /api/notifications/read-all", h.auth.RequireUser(h.markAllRead))
	mux.HandleFunc("GET /api/notification-settings", h.auth.RequireUser(h.getSettings))
	mux.HandleFunc("PUT /api/notification-settings", h.auth.RequireUser(h.updateSettings))

	mux.HandleFunc("POST /api/internal/notifications/scan", h.requireInternalToken(h.scan))
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, userID int64) {
	settings, err := h.repo.GetOrCreateSettings(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao ler preferências de notificação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !settings.SiteEnabled {
		writeJSON(w, http.StatusOK, []Notification{})
		return
	}

	items, err := h.repo.List(r.Context(), userID, 100)
	if err != nil {
		log.Printf("erro ao listar notificações: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) unreadCount(w http.ResponseWriter, r *http.Request, userID int64) {
	settings, err := h.repo.GetOrCreateSettings(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao ler preferências de notificação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	if !settings.SiteEnabled {
		writeJSON(w, http.StatusOK, map[string]int{"count": 0})
		return
	}

	count, err := h.repo.UnreadCount(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao contar notificações não lidas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"count": count})
}

func (h *Handler) markRead(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.repo.MarkRead(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrada")
			return
		}
		if errors.Is(err, ErrRequiresAction) {
			writeError(w, http.StatusConflict, "responda ao convite de compartilhamento para marcar como lida")
			return
		}
		log.Printf("erro ao marcar notificação como lida: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) markAllRead(w http.ResponseWriter, r *http.Request, userID int64) {
	if err := h.repo.MarkAllRead(r.Context(), userID); err != nil {
		log.Printf("erro ao marcar notificações como lidas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

// stream mantém uma conexão SSE aberta, empurrando notificações novas do
// usuário assim que a varredura periódica as cria.
func (h *Handler) stream(w http.ResponseWriter, r *http.Request, userID int64) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming não suportado")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Instrui o Nginx a não bufferizar esta resposta (SSE em tempo real).
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch, unsubscribe := h.hub.Subscribe(userID)
	defer unsubscribe()

	keepAlive := time.NewTicker(25 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case payload, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", payload)
			flusher.Flush()
		case <-keepAlive.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func (h *Handler) getSettings(w http.ResponseWriter, r *http.Request, userID int64) {
	settings, err := h.repo.GetOrCreateSettings(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao ler preferências de notificação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

func (h *Handler) updateSettings(w http.ResponseWriter, r *http.Request, userID int64) {
	var req Settings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if len(req.Offsets) == 0 {
		req.Offsets = DefaultOffsets
	}

	settings, err := h.repo.UpdateSettings(r.Context(), userID, req)
	if err != nil {
		log.Printf("erro ao salvar preferências de notificação: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, settings)
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

// scan varre as contas fixas ativas de todos os usuários e gera notificações
// (e e-mails) para os lembretes configurados que já chegaram na janela de
// antecedência, e para contas vencidas e ainda não pagas.
func (h *Handler) scan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	bills, err := h.bills.ListActive(ctx)
	if err != nil {
		log.Printf("erro ao listar contas fixas ativas para varredura: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	now := time.Now().UTC()
	created := 0
	for _, bill := range bills {
		dueDate, err := time.Parse("2006-01-02", bill.DueDate)
		if err != nil {
			continue
		}
		dueEnd := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 23, 59, 59, 0, time.UTC)
		minutesUntilDue := int(math.Round(dueEnd.Sub(now).Minutes()))

		settings, err := h.repo.GetOrCreateSettings(ctx, bill.UserID)
		if err != nil {
			log.Printf("erro ao ler preferências de notificação do usuário %d: %v", bill.UserID, err)
			continue
		}
		if !settings.SiteEnabled && !settings.EmailEnabled {
			continue
		}

		if minutesUntilDue >= 0 {
			for _, offset := range settings.Offsets {
				if minutesUntilDue > offset {
					continue
				}
				title := fmt.Sprintf("Conta vencendo: %s", bill.Name)
				message := fmt.Sprintf("%s vence em %s (%s).", bill.Name, formatDuration(minutesUntilDue), bill.DueDate)
				if h.notify(ctx, bill, settings, KindDueSoon, title, message, offset) {
					created++
				}
			}
		} else {
			daysOverdue := int(math.Ceil(float64(-minutesUntilDue) / 1440))
			title := fmt.Sprintf("Conta vencida: %s", bill.Name)
			message := fmt.Sprintf("%s venceu em %s e ainda não foi paga.", bill.Name, bill.DueDate)
			if h.notify(ctx, bill, settings, KindOverdue, title, message, -daysOverdue) {
				created++
			}
		}
	}

	writeJSON(w, http.StatusOK, map[string]int{"created": created})
}

func (h *Handler) notify(ctx context.Context, bill fixedbill.FixedBill, settings *Settings, kind, title, message string, offsetMinutes int) bool {
	notif, wasCreated, err := h.repo.Create(ctx, bill.UserID, bill.ID, kind, title, message, bill.DueDate, offsetMinutes)
	if err != nil {
		log.Printf("erro ao criar notificação para conta fixa %d: %v", bill.ID, err)
		return false
	}

	if wasCreated && settings.SiteEnabled {
		if payload, err := json.Marshal(notif); err == nil {
			h.hub.Publish(bill.UserID, payload)
		}
	}

	// O e-mail é tentado sempre que ainda não foi enviado para esta
	// notificação, mesmo em varreduras seguintes à criação — isso cobre o
	// caso de o usuário ativar "receber por e-mail" depois que a notificação
	// já apareceu no site.
	if settings.EmailEnabled && notif.EmailSentAt == nil {
		h.sendEmail(ctx, bill.UserID, title, message)
		if err := h.repo.MarkEmailSent(ctx, notif.ID); err != nil {
			log.Printf("erro ao marcar e-mail como enviado para notificação %d: %v", notif.ID, err)
		}
	}

	return wasCreated
}

func (h *Handler) sendEmail(ctx context.Context, userID int64, subject, body string) {
	u, err := h.users.GetByID(ctx, userID)
	if err != nil {
		log.Printf("erro ao buscar usuário %d para envio de e-mail: %v", userID, err)
		return
	}
	h.mailQueue.Enqueue(u.Email, subject, body)
}

func formatDuration(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%d minuto(s)", minutes)
	}
	if minutes < 1440 {
		return fmt.Sprintf("%d hora(s)", minutes/60)
	}
	return fmt.Sprintf("%d dia(s)", minutes/1440)
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
