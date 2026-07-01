package transaction

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lucasbrum/reconta/api/internal/tag"
)

// ErrNotFound é retornado quando a transação não existe (ou não pertence ao usuário).
var ErrNotFound = errors.New("transação não encontrada")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Input struct {
	Date           string
	Description    string
	Amount         float64
	Type           string
	CategoryID     *int64
	AccountID      *int64
	Notes          *string
	ImportedFrom   *string
	Bank           *string
	PixBeneficiary *string
}

func (r *Repository) Create(ctx context.Context, userID int64, in Input) (*Transaction, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions (user_id, date, description, amount, type, category_id, account_id, notes, imported_from, bank, pix_beneficiary)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, in.Date, in.Description, in.Amount, in.Type, in.CategoryID, in.AccountID, in.Notes,
		in.ImportedFrom, in.Bank, in.PixBeneficiary,
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo transação: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da transação: %w", err)
	}

	return r.GetByID(ctx, userID, id)
}

func (r *Repository) GetByID(ctx context.Context, userID, id int64) (*Transaction, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT t.id, t.date, t.description, t.amount, t.type, t.category_id, c.name, c.color,
		       t.account_id, t.notes, t.imported_from, t.bank, t.pix_beneficiary, t.created_at
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE t.id = ? AND t.user_id = ?`, id, userID,
	)
	return scanTransaction(row)
}

func (r *Repository) Update(ctx context.Context, userID, id int64, in Input) (*Transaction, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE transactions SET date = ?, description = ?, amount = ?, type = ?, category_id = ?, account_id = ?, notes = ?
		 WHERE id = ? AND user_id = ?`,
		in.Date, in.Description, in.Amount, in.Type, in.CategoryID, in.AccountID, in.Notes, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("atualizando transação: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM transactions WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("removendo transação: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// List retorna as transações que casam com os filtros, com totais e paginação.
func (r *Repository) List(ctx context.Context, userID int64, f ListFilters) (*ListResult, error) {
	where := []string{"t.user_id = ?"}
	args := []any{userID}

	if f.Month > 0 && f.Year > 0 {
		start, end := monthRange(f.Month, f.Year)
		where = append(where, "t.date >= ?", "t.date <= ?")
		args = append(args, start, end)
	}
	if f.Type == "income" || f.Type == "expense" {
		where = append(where, "t.type = ?")
		args = append(args, f.Type)
	}
	if f.CategoryID > 0 {
		where = append(where, "t.category_id = ?")
		args = append(args, f.CategoryID)
	}
	if f.TagID > 0 {
		where = append(where, "t.id IN (SELECT transaction_id FROM transaction_tags WHERE tag_id = ?)")
		args = append(args, f.TagID)
	}
	if f.Search != "" {
		where = append(where, "t.description LIKE ?")
		args = append(args, "%"+f.Search+"%")
	}

	whereClause := strings.Join(where, " AND ")

	page := max(f.Page, 1)
	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := (page - 1) * limit

	var totals Totals
	err := r.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT
			COALESCE(SUM(CASE WHEN t.type = 'income' THEN t.amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN t.type = 'expense' THEN t.amount ELSE 0 END), 0),
			COUNT(*)
		FROM transactions t
		WHERE %s`, whereClause), args...,
	).Scan(&totals.Income, &totals.Expense, &totals.Count)
	if err != nil {
		return nil, fmt.Errorf("calculando totais das transações: %w", err)
	}
	totals.Balance = totals.Income - totals.Expense

	listArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT t.id, t.date, t.description, t.amount, t.type, t.category_id, c.name, c.color,
		       t.account_id, t.notes, t.imported_from, t.bank, t.pix_beneficiary, t.created_at
		FROM transactions t
		LEFT JOIN categories c ON c.id = t.category_id
		WHERE %s
		ORDER BY t.date DESC, t.id DESC
		LIMIT ? OFFSET ?`, whereClause), listArgs...,
	)
	if err != nil {
		return nil, fmt.Errorf("listando transações: %w", err)
	}
	defer rows.Close()

	data := []Transaction{}
	for rows.Next() {
		tx, err := scanTransaction(rows)
		if err != nil {
			return nil, err
		}
		data = append(data, *tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &ListResult{
		Data:       data,
		Totals:     totals,
		Pagination: Pagination{Page: page, Limit: limit, Total: totals.Count},
	}, nil
}

type BulkUpdateFields struct {
	Type       *string
	CategoryID **int64
	AccountID  **int64
	Date       *string
}

// BulkUpdate atualiza campos em lote para uma lista de transações do usuário.
func (r *Repository) BulkUpdate(ctx context.Context, userID int64, ids []int64, fields BulkUpdateFields) (int, error) {
	if len(ids) == 0 {
		return 0, nil
	}

	sets := []string{}
	args := []any{}

	if fields.Type != nil {
		sets = append(sets, "type = ?")
		args = append(args, *fields.Type)
	}
	if fields.CategoryID != nil {
		sets = append(sets, "category_id = ?")
		args = append(args, *fields.CategoryID)
	}
	if fields.AccountID != nil {
		sets = append(sets, "account_id = ?")
		args = append(args, *fields.AccountID)
	}
	if fields.Date != nil {
		sets = append(sets, "date = ?")
		args = append(args, *fields.Date)
	}
	if len(sets) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, userID)

	query := fmt.Sprintf(`UPDATE transactions SET %s WHERE id IN (%s) AND user_id = ?`,
		strings.Join(sets, ", "), strings.Join(placeholders, ","))

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("atualizando transações em lote: %w", err)
	}
	n, err := res.RowsAffected()
	return int(n), err
}

// BulkDeleteScope define o escopo de uma exclusão em lote.
type BulkDeleteScope string

const (
	ScopeMonth BulkDeleteScope = "month"
	ScopeYear  BulkDeleteScope = "year"
	ScopeAll   BulkDeleteScope = "all"
)

// BulkDelete remove transações do usuário dentro de um escopo (mês, ano ou tudo).
func (r *Repository) BulkDelete(ctx context.Context, userID int64, scope BulkDeleteScope, month, year int) (int, error) {
	where := []string{"user_id = ?"}
	args := []any{userID}

	switch scope {
	case ScopeMonth:
		start, end := monthRange(month, year)
		where = append(where, "date >= ?", "date <= ?")
		args = append(args, start, end)
	case ScopeYear:
		where = append(where, "date >= ?", "date <= ?")
		args = append(args, fmt.Sprintf("%04d-01-01", year), fmt.Sprintf("%04d-12-31", year))
	case ScopeAll:
		// sem filtro adicional
	default:
		return 0, fmt.Errorf("escopo inválido: %s", scope)
	}

	res, err := r.db.ExecContext(ctx,
		fmt.Sprintf(`DELETE FROM transactions WHERE %s`, strings.Join(where, " AND ")), args...,
	)
	if err != nil {
		return 0, fmt.Errorf("removendo transações em lote: %w", err)
	}
	n, err := res.RowsAffected()
	return int(n), err
}

// UncategorizedTransaction é o subconjunto de campos usado pela auto-categorização.
type UncategorizedTransaction struct {
	ID             int64
	Description    string
	PixBeneficiary *string
}

func (r *Repository) ListUncategorized(ctx context.Context, userID int64) ([]UncategorizedTransaction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, description, pix_beneficiary FROM transactions WHERE user_id = ? AND category_id IS NULL`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando transações não categorizadas: %w", err)
	}
	defer rows.Close()

	result := []UncategorizedTransaction{}
	for rows.Next() {
		var t UncategorizedTransaction
		var pix sql.NullString
		if err := rows.Scan(&t.ID, &t.Description, &pix); err != nil {
			return nil, err
		}
		if pix.Valid {
			t.PixBeneficiary = &pix.String
		}
		result = append(result, t)
	}
	return result, rows.Err()
}

// BulkSetCategory aplica, em uma única transação SQL, a categoria escolhida para cada id informado.
func (r *Repository) BulkSetCategory(ctx context.Context, userID int64, categoryByTxID map[int64]int64) error {
	if len(categoryByTxID) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciando transação sql: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `UPDATE transactions SET category_id = ? WHERE id = ? AND user_id = ?`)
	if err != nil {
		return fmt.Errorf("preparando atualização de categoria: %w", err)
	}
	defer stmt.Close()

	for txID, categoryID := range categoryByTxID {
		if _, err := stmt.ExecContext(ctx, categoryID, txID, userID); err != nil {
			return fmt.Errorf("aplicando categoria %d à transação %d: %w", categoryID, txID, err)
		}
	}

	return tx.Commit()
}

// GetOpeningBalance retorna o valor manual salvo para o mês/ano, se houver.
func (r *Repository) GetOpeningBalance(ctx context.Context, userID int64, month, year int) (*float64, error) {
	var amount float64
	err := r.db.QueryRowContext(ctx,
		`SELECT amount FROM monthly_opening_balances WHERE user_id = ? AND month = ? AND year = ?`,
		userID, month, year,
	).Scan(&amount)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lendo saldo de abertura: %w", err)
	}
	return &amount, nil
}

// UpsertOpeningBalance cria ou atualiza o saldo de abertura manual de um mês/ano.
func (r *Repository) UpsertOpeningBalance(ctx context.Context, userID int64, month, year int, amount float64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO monthly_opening_balances (user_id, month, year, amount, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (user_id, month, year) DO UPDATE SET amount = excluded.amount, updated_at = excluded.updated_at`,
		userID, month, year, amount, time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
	)
	if err != nil {
		return fmt.Errorf("salvando saldo de abertura: %w", err)
	}
	return nil
}

// FindDuplicate verifica se já existe uma transação do usuário com a mesma
// data, valor e descrição — usado para sinalizar possíveis duplicatas ao
// importar um extrato já lançado anteriormente.
func (r *Repository) FindDuplicate(ctx context.Context, userID int64, date string, amount float64, description string) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM transactions WHERE user_id = ? AND date = ? AND amount = ? AND description = ?`,
		userID, date, amount, description,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("verificando duplicidade de transação: %w", err)
	}
	return count > 0, nil
}

// Period representa um mês/ano com pelo menos uma transação lançada.
type Period struct {
	Month int `json:"month"`
	Year  int `json:"year"`
}

// ListPeriods retorna os meses/anos distintos em que o usuário possui transações lançadas.
func (r *Repository) ListPeriods(ctx context.Context, userID int64) ([]Period, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT CAST(strftime('%m', date) AS INTEGER), CAST(strftime('%Y', date) AS INTEGER)
		FROM transactions
		WHERE user_id = ?
		ORDER BY 2 DESC, 1 DESC`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando períodos com transações: %w", err)
	}
	defer rows.Close()

	periods := []Period{}
	for rows.Next() {
		var p Period
		if err := rows.Scan(&p.Month, &p.Year); err != nil {
			return nil, err
		}
		periods = append(periods, p)
	}
	return periods, rows.Err()
}

func monthRange(month, year int) (start, end string) {
	first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := first.AddDate(0, 1, -1)
	return first.Format("2006-01-02"), lastDay.Format("2006-01-02")
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTransaction(s scanner) (*Transaction, error) {
	var t Transaction
	var categoryID sql.NullInt64
	var categoryName, categoryColor sql.NullString
	var accountID sql.NullInt64
	var notes, importedFrom, bank, pixBeneficiary sql.NullString
	var createdAt string

	if err := s.Scan(
		&t.ID, &t.Date, &t.Description, &t.Amount, &t.Type,
		&categoryID, &categoryName, &categoryColor,
		&accountID, &notes, &importedFrom, &bank, &pixBeneficiary, &createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo transação: %w", err)
	}

	if categoryID.Valid {
		t.CategoryID = &categoryID.Int64
	}
	if categoryName.Valid {
		t.CategoryName = &categoryName.String
	}
	if categoryColor.Valid {
		t.CategoryColor = &categoryColor.String
	}
	if accountID.Valid {
		t.AccountID = &accountID.Int64
	}
	if notes.Valid {
		t.Notes = &notes.String
	}
	if importedFrom.Valid {
		t.ImportedFrom = &importedFrom.String
	}
	if bank.Valid {
		t.Bank = &bank.String
	}
	if pixBeneficiary.Valid {
		t.PixBeneficiary = &pixBeneficiary.String
	}
	t.CreatedAt = parseTimestamp(createdAt)
	t.Tags = []tag.Tag{} // preenchido posteriormente pelo handler via tag.Repository

	return &t, nil
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
