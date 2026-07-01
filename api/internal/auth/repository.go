package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrSessionNotFound é retornado quando o token de sessão é inválido ou expirou.
var ErrSessionNotFound = errors.New("sessão não encontrada")

type Session struct {
	Token     string
	UserID    int64
	ExpiresAt time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("criando sessão: %w", err)
	}
	return nil
}

func (r *Repository) GetByToken(ctx context.Context, token string) (*Session, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT token, user_id, expires_at FROM sessions WHERE token = ?`, token,
	)

	var s Session
	var expiresAt string
	if err := row.Scan(&s.Token, &s.UserID, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("lendo sessão: %w", err)
	}

	t, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return nil, fmt.Errorf("lendo validade da sessão: %w", err)
	}
	s.ExpiresAt = t

	if time.Now().After(s.ExpiresAt) {
		_ = r.Delete(ctx, token)
		return nil, ErrSessionNotFound
	}

	return &s, nil
}

func (r *Repository) Delete(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("removendo sessão: %w", err)
	}
	return nil
}
