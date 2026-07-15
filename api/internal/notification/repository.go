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

// ErrRequiresAction é retornado ao tentar marcar como lida uma notificação de
// convite de compartilhamento ainda pendente — ela só pode ser resolvida
// aceitando ou rejeitando o convite, nunca marcada como lida diretamente.
var ErrRequiresAction = errors.New("notificação requer uma ação antes de ser marcada como lida")

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

// CreateGeneral insere uma notificação sem conta fixa associada (ex.: avisos
// de assinatura expirando). A deduplicação fica a cargo de quem chama, já que
// a chave UNIQUE da tabela não se aplica quando fixed_bill_id é NULL.
func (r *Repository) CreateGeneral(ctx context.Context, userID int64, kind, title, message, dueDate string, offsetMinutes int) (*Notification, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (user_id, fixed_bill_id, kind, title, message, due_date, offset_minutes)
		 VALUES (?, NULL, ?, ?, ?, ?, ?)`,
		userID, kind, title, message, dueDate, offsetMinutes,
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo notificação geral: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da notificação: %w", err)
	}
	return r.GetByID(ctx, userID, id)
}

// CreateForShare insere uma notificação geral associando-a a um compartilhamento
// (convite, aceite, rejeição ou cancelamento) — usada pelo pacote share.
func (r *Repository) CreateForShare(ctx context.Context, userID, shareID int64, kind, title, message string) (*Notification, error) {
	today := time.Now().UTC().Format("2006-01-02")
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO notifications (user_id, fixed_bill_id, share_id, kind, title, message, due_date, offset_minutes)
		 VALUES (?, NULL, ?, ?, ?, ?, ?, 0)`,
		userID, shareID, kind, title, message, today,
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo notificação de compartilhamento: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da notificação: %w", err)
	}
	return r.GetByID(ctx, userID, id)
}

const selectNotification = `
	SELECT n.id, n.fixed_bill_id, fb.name, n.share_id, n.kind, n.title, n.message, n.due_date, n.read_at, n.email_sent_at, n.created_at
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
	notif, err := r.GetByID(ctx, userID, id)
	if err != nil {
		return err
	}
	if notif.Kind == KindShareInvited && notif.ReadAt == nil {
		return ErrRequiresAction
	}

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

// MarkReadByShare marca como lida a notificação de um determinado kind
// associada a um compartilhamento, sem passar pelo guard de MarkRead —
// chamado pelo pacote share depois que o convite foi aceito, rejeitado ou
// cancelado, quando a ação que o desbloqueia já aconteceu.
func (r *Repository) MarkReadByShare(ctx context.Context, shareID int64, kind string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE share_id = ? AND kind = ? AND read_at IS NULL`,
		shareID, kind,
	)
	if err != nil {
		return fmt.Errorf("marcando notificação de compartilhamento como lida: %w", err)
	}
	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE notifications SET read_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE user_id = ? AND read_at IS NULL AND kind != ?`,
		userID, KindShareInvited,
	)
	if err != nil {
		return fmt.Errorf("marcando todas as notificações como lidas: %w", err)
	}
	return nil
}

// GetOrCreateSettings retorna as preferências de notificação do usuário,
// criando um registro com valores padrão na primeira vez que é acessado.
// A leitura vem primeiro para não pagar uma escrita (fsync) a cada chamada —
// esse método roda em todo GET de notificações.
func (r *Repository) GetOrCreateSettings(ctx context.Context, userID int64) (*Settings, error) {
	settings, err := r.getSettings(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		defaultOffsets, _ := json.Marshal(DefaultOffsets)
		if _, err := r.db.ExecContext(ctx,
			`INSERT OR IGNORE INTO notification_settings (user_id, offsets) VALUES (?, ?)`,
			userID, string(defaultOffsets),
		); err != nil {
			return nil, fmt.Errorf("criando preferências de notificação padrão: %w", err)
		}
		settings, err = r.getSettings(ctx, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("lendo preferências de notificação: %w", err)
	}
	return settings, nil
}

func (r *Repository) getSettings(ctx context.Context, userID int64) (*Settings, error) {
	var siteEnabled, emailEnabled, overdueEnabled bool
	var offsetsJSON string
	err := r.db.QueryRowContext(ctx,
		`SELECT site_enabled, email_enabled, offsets, overdue_enabled FROM notification_settings WHERE user_id = ?`, userID,
	).Scan(&siteEnabled, &emailEnabled, &offsetsJSON, &overdueEnabled)
	if err != nil {
		return nil, err
	}

	var offsets []int
	if err := json.Unmarshal([]byte(offsetsJSON), &offsets); err != nil {
		offsets = DefaultOffsets
	}
	return &Settings{SiteEnabled: siteEnabled, EmailEnabled: emailEnabled, Offsets: offsets, OverdueEnabled: overdueEnabled}, nil
}

func (r *Repository) UpdateSettings(ctx context.Context, userID int64, s Settings) (*Settings, error) {
	offsetsJSON, err := json.Marshal(s.Offsets)
	if err != nil {
		return nil, fmt.Errorf("codificando antecedências: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO notification_settings (user_id, site_enabled, email_enabled, offsets, overdue_enabled, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (user_id) DO UPDATE SET
			site_enabled = excluded.site_enabled,
			email_enabled = excluded.email_enabled,
			offsets = excluded.offsets,
			overdue_enabled = excluded.overdue_enabled,
			updated_at = excluded.updated_at`,
		userID, s.SiteEnabled, s.EmailEnabled, string(offsetsJSON), s.OverdueEnabled, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
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
	var shareID sql.NullInt64
	var readAt sql.NullString
	var emailSentAt sql.NullString
	var createdAt string

	if err := s.Scan(
		&n.ID, &fixedBillID, &fixedBillName, &shareID, &n.Kind, &n.Title, &n.Message, &n.DueDate, &readAt, &emailSentAt, &createdAt,
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
	if shareID.Valid {
		n.ShareID = &shareID.Int64
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
