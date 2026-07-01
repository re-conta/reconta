package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrEmailTaken é retornado ao tentar cadastrar um e-mail já existente.
var ErrEmailTaken = errors.New("e-mail já cadastrado")

// ErrNotFound é retornado quando o usuário não existe.
var ErrNotFound = errors.New("usuário não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, name, email, passwordHash string) (*User, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (name, email, password_hash) VALUES (?, ?, ?)`,
		name, email, passwordHash,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("inserindo usuário: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id do usuário: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at FROM users WHERE id = ?`, id,
	)
	return scanUser(row)
}

// GetByEmailWithPasswordHash retorna o usuário e o hash de senha correspondente,
// usado exclusivamente pelo fluxo de autenticação.
func (r *Repository) GetByEmailWithPasswordHash(ctx context.Context, email string) (*User, string, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at, password_hash FROM users WHERE email = ?`, email,
	)

	var u User
	var createdAt, passwordHash string
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &createdAt, &passwordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", ErrNotFound
		}
		return nil, "", fmt.Errorf("lendo usuário: %w", err)
	}
	u.CreatedAt = parseTimestamp(createdAt)

	return &u, passwordHash, nil
}

// GetByGoogleID busca um usuário previamente vinculado a uma conta Google.
func (r *Repository) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at FROM users WHERE google_id = ?`, googleID,
	)
	return scanUser(row)
}

// GetByEmail busca um usuário pelo e-mail, usado para vincular uma conta Google
// a um cadastro já existente por e-mail/senha.
func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at FROM users WHERE email = ?`, email,
	)
	return scanUser(row)
}

// CreateWithGoogle cria um usuário autenticado apenas via Google, sem senha.
func (r *Repository) CreateWithGoogle(ctx context.Context, name, email, googleID string) (*User, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO users (name, email, password_hash, google_id) VALUES (?, ?, '', ?)`,
		name, email, googleID,
	)
	if err != nil {
		if isUniqueConstraintErr(err) {
			return nil, ErrEmailTaken
		}
		return nil, fmt.Errorf("inserindo usuário via Google: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id do usuário: %w", err)
	}

	return r.GetByID(ctx, id)
}

// LinkGoogleID vincula um google_id a um usuário existente (cadastrado antes via e-mail/senha).
func (r *Repository) LinkGoogleID(ctx context.Context, id int64, googleID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET google_id = ? WHERE id = ?`, googleID, id)
	if err != nil {
		return fmt.Errorf("vinculando conta Google: %w", err)
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, email, created_at FROM users ORDER BY id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listando usuários: %w", err)
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		u, err := scanUserRows(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(s scanner) (*User, error) {
	var u User
	var createdAt string
	if err := s.Scan(&u.ID, &u.Name, &u.Email, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo usuário: %w", err)
	}
	u.CreatedAt = parseTimestamp(createdAt)
	return &u, nil
}

func scanUserRows(rows *sql.Rows) (*User, error) {
	return scanUser(rows)
}

func isUniqueConstraintErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "UNIQUE constraint failed") || strings.Contains(msg, "constraint failed: UNIQUE")
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
