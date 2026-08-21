package auditlog

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Entry é um registro de ação executada por um usuário no sistema (criar,
// editar, excluir, banir, etc), usado para auditoria no painel de admin.
type Entry struct {
	UserID    *int64
	Action    string
	Entity    string
	EntityID  *int64
	Details   string
	IP        string
	UserAgent string
}

func (r *Repository) Insert(ctx context.Context, e Entry) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO action_logs (user_id, action, entity, entity_id, details, ip, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		e.UserID, e.Action, e.Entity, e.EntityID, e.Details, e.IP, e.UserAgent,
	)
	if err != nil {
		return fmt.Errorf("registrando ação: %w", err)
	}
	return nil
}

type LogEntry struct {
	ID        int64  `json:"id"`
	UserID    *int64 `json:"userId"`
	UserName  string `json:"userName"`
	UserEmail string `json:"userEmail"`
	Action    string `json:"action"`
	Entity    string `json:"entity"`
	EntityID  *int64 `json:"entityId"`
	Details   string `json:"details"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
	CreatedAt string `json:"createdAt"`
}

// List retorna as ações mais recentes, com nome/e-mail do usuário quando
// disponível. userID == 0 não filtra por usuário.
func (r *Repository) List(ctx context.Context, userID int64, limit int) ([]LogEntry, error) {
	query := `
		SELECT a.id, a.user_id, COALESCE(u.name, ''), COALESCE(u.email, ''),
			a.action, a.entity, a.entity_id, a.details, a.ip, a.user_agent, a.created_at
		FROM action_logs a
		LEFT JOIN users u ON u.id = a.user_id`
	args := []any{}
	if userID != 0 {
		query += " WHERE a.user_id = ?"
		args = append(args, userID)
	}
	query += " ORDER BY a.created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("carregando log de ações: %w", err)
	}
	defer rows.Close()

	var out []LogEntry
	for rows.Next() {
		var e LogEntry
		if err := rows.Scan(
			&e.ID, &e.UserID, &e.UserName, &e.UserEmail,
			&e.Action, &e.Entity, &e.EntityID, &e.Details, &e.IP, &e.UserAgent, &e.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("lendo log de ações: %w", err)
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
