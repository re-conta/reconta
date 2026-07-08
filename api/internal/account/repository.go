package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound é retornado quando a conta não existe (ou não pertence ao usuário).
var ErrNotFound = errors.New("conta não encontrada")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID int64, name, accType string, balance float64) (*Account, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (user_id, name, type, balance) VALUES (?, ?, ?, ?)`,
		userID, name, accType, balance,
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo conta: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da conta: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

func (r *Repository) GetByID(ctx context.Context, userID, id int64) (*Account, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, type, balance, created_at FROM accounts WHERE id = ? AND user_id = ?`, id, userID,
	)
	return scanAccount(row)
}

func (r *Repository) List(ctx context.Context, userID int64) ([]Account, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, type, balance, created_at FROM accounts WHERE user_id = ? ORDER BY name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando contas: %w", err)
	}
	defer rows.Close()

	accounts := []Account{}
	for rows.Next() {
		a, err := scanAccount(rows)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, *a)
	}
	return accounts, rows.Err()
}

// FindOrCreateByName retorna a conta do usuário com o nome informado (case-insensitive),
// criando uma nova com saldo zero se não existir — usado ao restaurar backups.
func (r *Repository) FindOrCreateByName(ctx context.Context, userID int64, name, accType string) (*Account, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, type, balance, created_at FROM accounts WHERE user_id = ? AND LOWER(name) = LOWER(?)`,
		userID, name,
	)
	a, err := scanAccount(row)
	if err == nil {
		return a, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	return r.Create(ctx, userID, name, accType, 0)
}

func (r *Repository) Update(ctx context.Context, userID, id int64, name, accType string, balance float64) (*Account, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET name = ?, type = ?, balance = ? WHERE id = ? AND user_id = ?`,
		name, accType, balance, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("atualizando conta: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM accounts WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("removendo conta: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// SumBalance retorna a soma dos saldos de todas as contas do usuário.
func (r *Repository) SumBalance(ctx context.Context, userID int64) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(balance), 0) FROM accounts WHERE user_id = ?`, userID,
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("somando saldo das contas: %w", err)
	}
	return total, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanAccount(s scanner) (*Account, error) {
	var a Account
	var createdAt string
	if err := s.Scan(&a.ID, &a.Name, &a.Type, &a.Balance, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo conta: %w", err)
	}
	a.CreatedAt = parseTimestamp(createdAt)
	return &a, nil
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
