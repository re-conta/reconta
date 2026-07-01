package tag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrNotFound é retornado quando a tag não existe (ou não pertence ao usuário).
var ErrNotFound = errors.New("tag não encontrada")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, userID int64, name, color string) (*Tag, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?)`,
		userID, name, color,
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo tag: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da tag: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

func (r *Repository) GetByID(ctx context.Context, userID, id int64) (*Tag, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, name, color FROM tags WHERE id = ? AND user_id = ?`, id, userID,
	)
	return scanTag(row)
}

func (r *Repository) List(ctx context.Context, userID int64) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, name, color FROM tags WHERE user_id = ? ORDER BY name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando tags: %w", err)
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		t, err := scanTag(rows)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *t)
	}
	return tags, rows.Err()
}

func (r *Repository) Update(ctx context.Context, userID, id int64, name, color string) (*Tag, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE tags SET name = ?, color = ? WHERE id = ? AND user_id = ?`,
		name, color, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("atualizando tag: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("removendo tag: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// FilterOwnedIDs recebe uma lista de ids de tag e retorna apenas os que pertencem ao usuário.
func (r *Repository) FilterOwnedIDs(ctx context.Context, userID int64, ids []int64) ([]int64, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]any, 0, len(ids)+1)
	args = append(args, userID)
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(`SELECT id FROM tags WHERE user_id = ? AND id IN (%s)`, strings.Join(placeholders, ","))
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("filtrando tags do usuário: %w", err)
	}
	defer rows.Close()

	owned := []int64{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		owned = append(owned, id)
	}
	return owned, rows.Err()
}

// ListByTransactionIDs retorna, para cada transação informada, suas tags associadas.
func (r *Repository) ListByTransactionIDs(ctx context.Context, txIDs []int64) (map[int64][]Tag, error) {
	result := map[int64][]Tag{}
	if len(txIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(txIDs))
	args := make([]any, len(txIDs))
	for i, id := range txIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT tt.transaction_id, t.id, t.name, t.color
		FROM transaction_tags tt
		INNER JOIN tags t ON t.id = tt.tag_id
		WHERE tt.transaction_id IN (%s)`, strings.Join(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listando tags das transações: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var txID int64
		var t Tag
		if err := rows.Scan(&txID, &t.ID, &t.Name, &t.Color); err != nil {
			return nil, err
		}
		result[txID] = append(result[txID], t)
	}
	return result, rows.Err()
}

// ListByTransactionID retorna as tags de uma única transação.
func (r *Repository) ListByTransactionID(ctx context.Context, txID int64) ([]Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.color
		FROM transaction_tags tt
		INNER JOIN tags t ON t.id = tt.tag_id
		WHERE tt.transaction_id = ?`, txID)
	if err != nil {
		return nil, fmt.Errorf("listando tags da transação: %w", err)
	}
	defer rows.Close()

	tags := []Tag{}
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// SetTransactionTags substitui completamente as tags associadas a uma transação.
func (r *Repository) SetTransactionTags(ctx context.Context, txID int64, tagIDs []int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciando transação sql: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM transaction_tags WHERE transaction_id = ?`, txID); err != nil {
		return fmt.Errorf("removendo tags da transação: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO transaction_tags (transaction_id, tag_id) VALUES (?, ?)`, txID, tagID,
		); err != nil {
			return fmt.Errorf("associando tag à transação: %w", err)
		}
	}

	return tx.Commit()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTag(s scanner) (*Tag, error) {
	var t Tag
	if err := s.Scan(&t.ID, &t.Name, &t.Color); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo tag: %w", err)
	}
	return &t, nil
}
