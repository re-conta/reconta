package share

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound é retornado quando o compartilhamento não existe ou não pertence
// ao usuário que está tentando acessá-lo.
var ErrNotFound = errors.New("compartilhamento não encontrado")

// ErrInvalidState é retornado ao tentar aceitar/rejeitar um convite que não
// está mais pendente, ou cancelar um já cancelado.
var ErrInvalidState = errors.New("compartilhamento não está no estado esperado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Create insere o convite de compartilhamento e as contas incluídas numa
// única transação SQL.
func (r *Repository) Create(ctx context.Context, ownerID, recipientID int64, accountIDs []int64, canEdit, includeFuture bool, periodStart, periodEnd *string) (*Share, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("iniciando transação sql: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO shares (owner_id, recipient_id, can_edit, include_future, period_start, period_end)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		ownerID, recipientID, canEdit, includeFuture, periodStart, periodEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo compartilhamento: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id do compartilhamento: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO share_accounts (share_id, account_id) VALUES (?, ?)`)
	if err != nil {
		return nil, fmt.Errorf("preparando inserção de contas do compartilhamento: %w", err)
	}
	defer stmt.Close()
	for _, accountID := range accountIDs {
		if _, err := stmt.ExecContext(ctx, id, accountID); err != nil {
			return nil, fmt.Errorf("associando conta %d ao compartilhamento: %w", accountID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("confirmando criação do compartilhamento: %w", err)
	}

	return r.GetByID(ctx, id)
}

const selectShare = `
	SELECT s.id, s.owner_id, o.name, s.recipient_id, rc.name,
	       s.can_edit, s.include_future, s.period_start, s.period_end, s.status, s.created_at, s.responded_at
	FROM shares s
	JOIN users o ON o.id = s.owner_id
	JOIN users rc ON rc.id = s.recipient_id`

func (r *Repository) GetByID(ctx context.Context, id int64) (*Share, error) {
	row := r.db.QueryRowContext(ctx, selectShare+` WHERE s.id = ?`, id)
	s, err := scanShare(row)
	if err != nil {
		return nil, err
	}
	if err := r.attachAccounts(ctx, s); err != nil {
		return nil, err
	}
	return s, nil
}

// ListSent retorna os compartilhamentos criados pelo usuário, mais recentes primeiro.
func (r *Repository) ListSent(ctx context.Context, ownerID int64) ([]Share, error) {
	return r.list(ctx, selectShare+` WHERE s.owner_id = ? ORDER BY s.created_at DESC, s.id DESC`, ownerID)
}

// ListReceived retorna os compartilhamentos em que o usuário é o convidado, mais recentes primeiro.
func (r *Repository) ListReceived(ctx context.Context, recipientID int64) ([]Share, error) {
	return r.list(ctx, selectShare+` WHERE s.recipient_id = ? ORDER BY s.created_at DESC, s.id DESC`, recipientID)
}

func (r *Repository) list(ctx context.Context, query string, arg int64) ([]Share, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("listando compartilhamentos: %w", err)
	}
	defer rows.Close()

	items := []Share{}
	for rows.Next() {
		s, err := scanShare(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range items {
		if err := r.attachAccounts(ctx, &items[i]); err != nil {
			return nil, err
		}
	}
	return items, nil
}

// attachAccounts preenche AccountIDs/AccountNames do compartilhamento.
func (r *Repository) attachAccounts(ctx context.Context, s *Share) error {
	rows, err := r.db.QueryContext(ctx,
		`SELECT a.id, a.name FROM share_accounts sa JOIN accounts a ON a.id = sa.account_id WHERE sa.share_id = ? ORDER BY a.name`,
		s.ID,
	)
	if err != nil {
		return fmt.Errorf("listando contas do compartilhamento: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		s.AccountIDs = append(s.AccountIDs, id)
		s.AccountNames = append(s.AccountNames, name)
	}
	return rows.Err()
}

// Accept marca o convite como aceito. Só tem efeito se o usuário for o
// convidado e o convite ainda estiver pendente.
func (r *Repository) Accept(ctx context.Context, recipientID, id int64) (*Share, error) {
	return r.respond(ctx, recipientID, id, StatusAccepted)
}

// Reject marca o convite como rejeitado. Só tem efeito se o usuário for o
// convidado e o convite ainda estiver pendente.
func (r *Repository) Reject(ctx context.Context, recipientID, id int64) (*Share, error) {
	return r.respond(ctx, recipientID, id, StatusRejected)
}

func (r *Repository) respond(ctx context.Context, recipientID, id int64, status string) (*Share, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE shares SET status = ?, responded_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		 WHERE id = ? AND recipient_id = ? AND status = ?`,
		status, id, recipientID, StatusPending,
	)
	if err != nil {
		return nil, fmt.Errorf("respondendo compartilhamento: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		var exists int
		_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM shares WHERE id = ? AND recipient_id = ?`, id, recipientID).Scan(&exists)
		if exists == 0 {
			return nil, ErrNotFound
		}
		return nil, ErrInvalidState
	}
	return r.GetByID(ctx, id)
}

// Cancel marca o compartilhamento como cancelado. Permitido a qualquer
// momento pelo dono, enquanto ainda não estiver cancelado.
func (r *Repository) Cancel(ctx context.Context, ownerID, id int64) (*Share, error) {
	existing, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing.OwnerID != ownerID {
		return nil, ErrNotFound
	}
	if existing.Status == StatusCancelled {
		return nil, ErrInvalidState
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE shares SET status = ?, responded_at = COALESCE(responded_at, strftime('%Y-%m-%dT%H:%M:%fZ', 'now')) WHERE id = ?`,
		StatusCancelled, id,
	); err != nil {
		return nil, fmt.Errorf("cancelando compartilhamento: %w", err)
	}
	return r.GetByID(ctx, id)
}

// GetActiveGrant resolve o acesso efetivo do convidado a um compartilhamento
// aceito: contas incluídas, permissão de edição e janela de datas
// (DateTo == nil quando include_future estiver ativo, ou seja, sem limite superior).
func (r *Repository) GetActiveGrant(ctx context.Context, recipientID, shareID int64) (*AccessGrant, error) {
	s, err := r.GetByID(ctx, shareID)
	if err != nil {
		return nil, err
	}
	if s.RecipientID != recipientID || s.Status != StatusAccepted {
		return nil, ErrNotFound
	}

	grant := &AccessGrant{
		ShareID:    s.ID,
		OwnerID:    s.OwnerID,
		AccountIDs: s.AccountIDs,
		CanEdit:    s.CanEdit,
		DateFrom:   s.PeriodStart,
	}
	if !s.IncludeFuture {
		grant.DateTo = s.PeriodEnd
	}
	return grant, nil
}

// ListActiveGrants resolve o acesso efetivo de todos os compartilhamentos
// aceitos do convidado — usado para a listagem "compartilhado comigo".
func (r *Repository) ListActiveGrants(ctx context.Context, recipientID int64) ([]AccessGrant, error) {
	shares, err := r.ListReceived(ctx, recipientID)
	if err != nil {
		return nil, err
	}

	grants := make([]AccessGrant, 0, len(shares))
	for _, s := range shares {
		if s.Status != StatusAccepted {
			continue
		}
		grant := AccessGrant{
			ShareID:    s.ID,
			OwnerID:    s.OwnerID,
			AccountIDs: s.AccountIDs,
			CanEdit:    s.CanEdit,
			DateFrom:   s.PeriodStart,
		}
		if !s.IncludeFuture {
			grant.DateTo = s.PeriodEnd
		}
		grants = append(grants, grant)
	}
	return grants, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanShare(s scanner) (*Share, error) {
	var sh Share
	var periodStart, periodEnd sql.NullString
	var respondedAt sql.NullString
	var createdAt string

	if err := s.Scan(
		&sh.ID, &sh.OwnerID, &sh.OwnerName, &sh.RecipientID, &sh.RecipientName,
		&sh.CanEdit, &sh.IncludeFuture, &periodStart, &periodEnd, &sh.Status, &createdAt, &respondedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo compartilhamento: %w", err)
	}

	if periodStart.Valid {
		sh.PeriodStart = &periodStart.String
	}
	if periodEnd.Valid {
		sh.PeriodEnd = &periodEnd.String
	}
	if respondedAt.Valid {
		t := parseTimestamp(respondedAt.String)
		sh.RespondedAt = &t
	}
	sh.CreatedAt = parseTimestamp(createdAt)
	sh.AccountIDs = []int64{}
	sh.AccountNames = []string{}

	return &sh, nil
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
