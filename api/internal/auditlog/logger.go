package auditlog

import (
	"context"
	"log"
	"net"
	"net/http"
)

// Logger grava ações de usuários (criar/editar/excluir/etc) para auditoria.
// Métodos toleram receiver nil para simplificar handlers que ainda não
// tiverem o logger configurado (ex.: em testes).
type Logger struct {
	repo *Repository
}

func NewLogger(repo *Repository) *Logger {
	return &Logger{repo: repo}
}

// Log registra uma ação associada à requisição atual. entityID pode ser nil
// quando a ação não se refere a uma entidade específica (ex.: login).
// Nunca deve interromper o fluxo do handler: erros só são logados no servidor.
func (l *Logger) Log(r *http.Request, userID int64, action, entity string, entityID *int64, details string) {
	if l == nil || l.repo == nil {
		return
	}
	var uid *int64
	if userID != 0 {
		uid = &userID
	}
	entry := Entry{
		UserID:    uid,
		Action:    action,
		Entity:    entity,
		EntityID:  entityID,
		Details:   details,
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
	}
	if err := l.repo.Insert(context.Background(), entry); err != nil {
		log.Printf("erro ao registrar ação de auditoria: %v", err)
	}
}

// clientIP resolve o IP real do requisitante, na mesma ordem de prioridade
// usada pelo pacote analytics: Cloudflare, Nginx, e por fim a conexão TCP.
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
