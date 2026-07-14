package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound é retornado quando a notificação não existe (ou não pertence ao usuário).
var ErrNotFound = errors.New("notificação não encontrada")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create insere a notificação se ainda não existir uma igual para o mesmo
// (fixedBillID, dueDate, offsetMinutes) — essa combinação é a chave de
// deduplicação que impede reenviar o mesmo lembrete a cada varredura.
// Retorna created=false quando a notificação já existia.
func (r *Repository) Create(ctx context.Context, userID int64, fixedBillID int64, kind, title, message, dueDate string, offsetMinutes int) (*Notification, bool, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO notifications (user_id, fixed_bill_id, kind, title, message, due_date, offset_minutes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, fixedBillID, kind, title, message, dueDate, offsetMinutes,
	)
	if err != nil {
		return nil, false, fmt.Errorf("inserindo notificação: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, false, fmt.Errorf("verificando notificação inserida: %w", err)
	}
	if n == 0 {
		existing, err := r.getByBillDueOffset(ctx, fixedBillID, dueDate, offsetMinutes)
		if err != nil {
			return nil, false, err
		}
		return existing, false, nil
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, false, fmt.Errorf("obtendo id da notificação: %w", err)
	}
	created, err := r.GetByID(ctx, userID, id)
	if err != nil {
		return nil, false, err
	}
	return created, true, nil
}

const selectNotification = `
	SELECT n.id, n.fixed_bill_id, fb.name, n.kind, n.title, n.message, n.due_date, n.read_at, n.email_sent_at, n.created_at
	FROM notifications n
	LEFT JOIN fixed_bills fb ON fb.id = n.fixed_bill_id`

// getByBillDueOffset busca a notificação já existente para a chave de
// deduplicação (fixed_bill_id, due_date, offset_minutes), usada quando Create
// não insere por já haver uma igual.
func (r *Repository) getByBillDueOffset(ctx context.Context, fixedBillID int64, dueDate string, offsetMinutes int) (*Notification, error) {
	row := r.db.QueryRowContext(ctx,
		selectNotification+` WHERE n.fixed_bill_id = ? AND n.due_date = ? AND n.offset_minutes = ?`,
		fixedBillID, dueDate, offsetMinutes,
	)
	return scanNotification(row)
}

// MarkEmailSent registra que o e-mail de lembrete desta notificação foi
// enfileirado para envio, evitando reenvios a cada varredura.
func (r *Repository) MarkEmailSent(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET email_sent_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ?`, id,
	)
	if err != nil {
		return fmt.Errorf("marcando e-mail como enviado: %w", err)
	}
	return nil
}

func (r *Repository) GetByID(ctx context.Context, userID, id int64) (*Notification, error) {
	row := r.db.QueryRowContext(ctx, selectNotification+` WHERE n.id = ? AND n.user_id = ?`, id, userID)
	return scanNotification(row)
}

func (r *Repository) List(ctx context.Context, userID int64, limit int) ([]Notification, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx,
		selectNotification+` WHERE n.user_id = ? ORDER BY n.created_at DESC, n.id DESC LIMIT ?`, userID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("listando notificações: %w", err)
	}
	defer rows.Close()

	items := []Notification{}
	for rows.Next() {
		notif, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *notif)
	}
	return items, rows.Err()
}

func (r *Repository) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM notifications WHERE user_id = ? AND read_at IS NULL`, userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("contando notificações não lidas: %w", err)
	}
	return count, nil
}

func (r *Repository) MarkRead(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ? AND user_id = ? AND read_at IS NULL`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("marcando notificação como lida: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		var exists int
		_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE id = ? AND user_id = ?`, id, userID).Scan(&exists)
		if exists == 0 {
			return ErrNotFound
		}
	}
	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE user_id = ? AND read_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("marcando todas as notificações como lidas: %w", err)
	}
	return nil
}

// GetOrCreateSettings retorna as preferências de notificação do usuário,
// criando um registro com valores padrão na primeira vez que é acessado.
func (r *Repository) GetOrCreateSettings(ctx context.Context, userID int64) (*Settings, error) {
	defaultOffsets, _ := json.Marshal(DefaultOffsets)
	if _, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO notification_settings (user_id, offsets) VALUES (?, ?)`,
		userID, string(defaultOffsets),
	); err != nil {
		return nil, fmt.Errorf("criando preferências de notificação padrão: %w", err)
	}

	var siteEnabled, emailEnabled bool
	var offsetsJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT site_enabled, email_enabled, offsets FROM notification_settings WHERE user_id = ?`, userID,
	).Scan(&siteEnabled, &emailEnabled, &offsetsJSON)
	if err != nil {
		return nil, fmt.Errorf("lendo preferências de notificação: %w", err)
	}

	var offsets []int
	if err := json.Unmarshal([]byte(offsetsJSON), &offsets); err != nil {
		offsets = DefaultOffsets
	}
	return &Settings{SiteEnabled: siteEnabled, EmailEnabled: emailEnabled, Offsets: offsets}, nil
}

func (r *Repository) UpdateSettings(ctx context.Context, userID int64, s Settings) (*Settings, error) {
	offsetsJSON, err := json.Marshal(s.Offsets)
	if err != nil {
		return nil, fmt.Errorf("codificando antecedências: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO notification_settings (user_id, site_enabled, email_enabled, offsets, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			site_enabled = excluded.site_enabled,
			email_enabled = excluded.email_enabled,
			offsets = excluded.offsets,
			updated_at = excluded.updated_at`,
		userID, s.SiteEnabled, s.EmailEnabled, string(offsetsJSON), time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	)
	if err != nil {
		return nil, fmt.Errorf("salvando preferências de notificação: %w", err)
	}
	return r.GetOrCreateSettings(ctx, userID)
}

// ListUserIDsWithActiveBills retorna os ids de usuários que possuem ao menos
// uma conta fixa ativa, usado pela varredura periódica.
func (r *Repository) ListUserIDsWithActiveBills(ctx context.Context) ([]int64, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT DISTINCT user_id FROM fixed_bills WHERE status = 'active'`,
	)
	if err != nil {
		return nil, fmt.Errorf("listando usuários com contas fixas ativas: %w", err)
	}
	defer rows.Close()

	ids := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(s scanner) (*Notification, error) {
	var n Notification
	var fixedBillID sql.NullInt64
	var fixedBillName sql.NullString
	var readAt sql.NullString
	var emailSentAt sql.NullString
	var createdAt string

	if err := s.Scan(
		&n.ID, &fixedBillID, &fixedBillName, &n.Kind, &n.Title, &n.Message, &n.DueDate, &readAt, &emailSentAt, &createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo notificação: %w", err)
	}

	if fixedBillID.Valid {
		n.FixedBillID = &fixedBillID.Int64
	}
	if fixedBillName.Valid {
		n.FixedBillName = &fixedBillName.String
	}
	if readAt.Valid {
		t := parseTimestamp(readAt.String)
		n.ReadAt = &t
	}
	if emailSentAt.Valid {
		t := parseTimestamp(emailSentAt.String)
		n.EmailSentAt = &t
	}
	n.CreatedAt = parseTimestamp(createdAt)
	return &n, nil
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
