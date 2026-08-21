package taxsim

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

// TaxableIncome soma as receitas do usuário no ano informado cujas
// categorias estão marcadas como tributáveis, ignorando transferências.
func (r *Repository) TaxableIncome(ctx context.Context, userID int64, year int) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(t.amount), 0)
		FROM transactions t
		INNER JOIN categories c ON c.id = t.category_id
		WHERE t.user_id = ?
		  AND t.type = 'income'
		  AND t.is_transfer = 0
		  AND c.is_taxable = 1
		  AND strftime('%Y', t.date) = ?`,
		userID, fmt.Sprintf("%04d", year),
	).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("somando receitas tributáveis: %w", err)
	}
	return total, nil
}
