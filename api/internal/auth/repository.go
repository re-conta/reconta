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

// ErrResetTokenNotFound é retornado quando o token de redefinição de senha é
// inválido, já foi usado ou expirou.
var ErrResetTokenNotFound = errors.New("token de redefinição não encontrado")

// CreateResetToken registra um token de redefinição de senha para o usuário
// informado, substituindo qualquer token anterior ainda pendente.
func (r *Repository) CreateResetToken(ctx context.Context, token string, userID int64, expiresAt time.Time) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE user_id = ?`, userID); err != nil {
		return fmt.Errorf("limpando tokens de redefinição anteriores: %w", err)
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO password_reset_tokens (token, user_id, expires_at) VALUES (?, ?, ?)`,
		token, userID, expiresAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("criando token de redefinição: %w", err)
	}
	return nil
}

// GetResetToken busca o token de redefinição de senha, retornando
// ErrResetTokenNotFound se estiver ausente ou expirado.
func (r *Repository) GetResetToken(ctx context.Context, token string) (int64, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM password_reset_tokens WHERE token = ?`, token,
	)

	var userID int64
	var expiresAt string
	if err := row.Scan(&userID, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrResetTokenNotFound
		}
		return 0, fmt.Errorf("lendo token de redefinição: %w", err)
	}

	t, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return 0, fmt.Errorf("lendo validade do token de redefinição: %w", err)
	}
	if time.Now().After(t) {
		_ = r.DeleteResetToken(ctx, token)
		return 0, ErrResetTokenNotFound
	}

	return userID, nil
}

// DeleteResetToken remove um token de redefinição de senha após o uso (ou expiração).
func (r *Repository) DeleteResetToken(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM password_reset_tokens WHERE token = ?`, token)
	if err != nil {
		return fmt.Errorf("removendo token de redefinição: %w", err)
	}
	return nil
}
