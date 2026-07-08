package category

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotFound é retornado quando a categoria não existe (ou não pertence ao usuário).
var ErrNotFound = errors.New("categoria não encontrada")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID int64, name, color, icon, catType, patterns string) (*Category, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO categories (user_id, name, color, icon, type, patterns) VALUES (?, ?, ?, ?, ?, ?)`,
		userID, name, color, icon, catType, nullableString(patterns),
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo categoria: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da categoria: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

func (r *Repository) GetByID(ctx context.Context, userID, id int64) (*Category, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, color, icon, type, patterns FROM categories WHERE id = ? AND user_id = ?`, id, userID,
	)
	return scanCategory(row)
}

func (r *Repository) List(ctx context.Context, userID int64) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color, icon, type, patterns FROM categories WHERE user_id = ? ORDER BY name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando categorias: %w", err)
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *c)
	}
	return categories, rows.Err()
}

// ListWithPatterns retorna apenas categorias que possuem padrões de auto-categorização definidos.
func (r *Repository) ListWithPatterns(ctx context.Context, userID int64) ([]Category, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color, icon, type, patterns FROM categories
		 WHERE user_id = ? AND patterns IS NOT NULL AND patterns != ''`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando categorias com padrões: %w", err)
	}
	defer rows.Close()

	categories := []Category{}
	for rows.Next() {
		c, err := scanCategory(rows)
		if err != nil {
			return nil, err
		}
		categories = append(categories, *c)
	}
	return categories, rows.Err()
}

// FindOrCreateByName retorna a categoria do usuário com o nome informado (case-insensitive),
// criando uma nova (sem padrões de auto-categorização) se não existir — usado ao restaurar backups.
func (r *Repository) FindOrCreateByName(ctx context.Context, userID int64, name, color, icon, catType string) (*Category, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, color, icon, type, patterns FROM categories WHERE user_id = ? AND LOWER(name) = LOWER(?)`,
		userID, name,
	)
	c, err := scanCategory(row)
	if err == nil {
		return c, nil
	}
	if !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	return r.Create(ctx, userID, name, color, icon, catType, "")
}

func (r *Repository) Update(ctx context.Context, userID, id int64, name, color, icon, catType, patterns string) (*Category, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE categories SET name = ?, color = ?, icon = ?, type = ?, patterns = ? WHERE id = ? AND user_id = ?`,
		name, color, icon, catType, nullableString(patterns), id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("atualizando categoria: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("removendo categoria: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCategory(s scanner) (*Category, error) {
	var c Category
	var patterns sql.NullString
	if err := s.Scan(&c.ID, &c.Name, &c.Color, &c.Icon, &c.Type, &patterns); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo categoria: %w", err)
	}
	c.Patterns = patterns.String
	return &c, nil
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
