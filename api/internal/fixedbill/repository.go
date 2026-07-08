package fixedbill

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// ErrNotFound é retornado quando a conta fixa não existe (ou não pertence ao usuário).
var ErrNotFound = errors.New("conta fixa não encontrada")

// ErrNotActive é retornado ao tentar pagar uma conta congelada ou encerrada.
var ErrNotActive = errors.New("conta fixa não está ativa")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

type Input struct {
	Name        string
	Amount      float64
	CategoryID  *int64
	AccountID   *int64
	Periodicity string
	DueDate     string
	Notes       *string
}

func (r *Repository) Create(ctx context.Context, userID int64, in Input) (*FixedBill, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO fixed_bills (user_id, name, amount, category_id, account_id, periodicity, due_date, notes)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, in.Name, in.Amount, in.CategoryID, in.AccountID, in.Periodicity, in.DueDate, in.Notes,
	)
	if err != nil {
		return nil, fmt.Errorf("inserindo conta fixa: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da conta fixa: %w", err)
	}
	return r.GetByID(ctx, userID, id)
}

const selectFixedBill = `
	SELECT fb.id, fb.user_id, fb.name, fb.amount, fb.category_id, c.name, c.color, fb.account_id, a.name,
	       fb.periodicity, fb.due_date, fb.status, fb.notes, fb.created_at, fb.updated_at
	FROM fixed_bills fb
	LEFT JOIN categories c ON c.id = fb.category_id
	LEFT JOIN accounts a ON a.id = fb.account_id`

func (r *Repository) GetByID(ctx context.Context, userID, id int64) (*FixedBill, error) {
	row := r.db.QueryRowContext(ctx, selectFixedBill+` WHERE fb.id = ? AND fb.user_id = ?`, id, userID)
	return scanFixedBill(row)
}

func (r *Repository) List(ctx context.Context, userID int64) ([]FixedBill, error) {
	rows, err := r.db.QueryContext(ctx,
		selectFixedBill+` WHERE fb.user_id = ? ORDER BY fb.status = 'closed', fb.due_date`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando contas fixas: %w", err)
	}
	defer rows.Close()

	bills := []FixedBill{}
	for rows.Next() {
		b, err := scanFixedBill(rows)
		if err != nil {
			return nil, err
		}
		bills = append(bills, *b)
	}
	return bills, rows.Err()
}

// ListActiveWithReminders retorna as contas fixas ativas de todos os usuários,
// usada pela varredura periódica de notificações.
func (r *Repository) ListActive(ctx context.Context) ([]FixedBill, error) {
	rows, err := r.db.QueryContext(ctx, selectFixedBill+` WHERE fb.status = ?`, StatusActive)
	if err != nil {
		return nil, fmt.Errorf("listando contas fixas ativas: %w", err)
	}
	defer rows.Close()

	bills := []FixedBill{}
	for rows.Next() {
		b, err := scanFixedBill(rows)
		if err != nil {
			return nil, err
		}
		bills = append(bills, *b)
	}
	return bills, rows.Err()
}

func (r *Repository) Update(ctx context.Context, userID, id int64, in Input) (*FixedBill, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE fixed_bills SET name = ?, amount = ?, category_id = ?, account_id = ?, periodicity = ?, due_date = ?, notes = ?,
		 updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		 WHERE id = ? AND user_id = ?`,
		in.Name, in.Amount, in.CategoryID, in.AccountID, in.Periodicity, in.DueDate, in.Notes, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("atualizando conta fixa: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *Repository) UpdateStatus(ctx context.Context, userID, id int64, status string) (*FixedBill, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE fixed_bills SET status = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ? AND user_id = ?`,
		status, id, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("atualizando status da conta fixa: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, userID, id)
}

func (r *Repository) Delete(ctx context.Context, userID, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM fixed_bills WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("removendo conta fixa: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

// PayInput agrupa os dados opcionais de um pagamento detalhado. Campos nulos
// usam valores padrão (data de hoje, valor estimado da conta, conta bancária padrão).
type PayInput struct {
	Bank          *string
	PaymentMethod *string
	PaidAt        *string
	AmountPaid    *float64
	AccountID     *int64
	Notes         *string
}

// Pay registra o pagamento do ciclo atual de uma conta fixa: cria a transação
// correspondente, registra o pagamento e avança o vencimento para o próximo
// ciclo, tudo em uma única transação SQL.
func (r *Repository) Pay(ctx context.Context, userID, billID int64, in PayInput) (*Payment, *FixedBill, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("iniciando transação sql: %w", err)
	}
	defer tx.Rollback()

	var (
		name        string
		amount      float64
		categoryID  sql.NullInt64
		accountID   sql.NullInt64
		periodicity string
		dueDate     string
		status      string
	)
	err = tx.QueryRowContext(ctx,
		`SELECT name, amount, category_id, account_id, periodicity, due_date, status
		 FROM fixed_bills WHERE id = ? AND user_id = ?`, billID, userID,
	).Scan(&name, &amount, &categoryID, &accountID, &periodicity, &dueDate, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, fmt.Errorf("lendo conta fixa: %w", err)
	}
	if status != StatusActive {
		return nil, nil, ErrNotActive
	}

	paidAt := time.Now().UTC().Format("2006-01-02")
	if in.PaidAt != nil && *in.PaidAt != "" {
		paidAt = *in.PaidAt
	}
	amountPaid := amount
	if in.AmountPaid != nil {
		amountPaid = *in.AmountPaid
	}
	var payAccountID *int64
	if in.AccountID != nil {
		payAccountID = in.AccountID
	} else if accountID.Valid {
		payAccountID = &accountID.Int64
	}
	var payCategoryID *int64
	if categoryID.Valid {
		payCategoryID = &categoryID.Int64
	}

	txRes, err := tx.ExecContext(ctx,
		`INSERT INTO transactions (user_id, date, description, amount, type, category_id, account_id, notes, bank)
		 VALUES (?, ?, ?, ?, 'expense', ?, ?, ?, ?)`,
		userID, paidAt, name, amountPaid, payCategoryID, payAccountID, in.Notes, in.Bank,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("inserindo transação do pagamento: %w", err)
	}
	transactionID, err := txRes.LastInsertId()
	if err != nil {
		return nil, nil, fmt.Errorf("obtendo id da transação: %w", err)
	}

	paymentRes, err := tx.ExecContext(ctx,
		`INSERT INTO fixed_bill_payments (fixed_bill_id, user_id, due_date, paid_at, amount_paid, bank, payment_method, notes, transaction_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		billID, userID, dueDate, paidAt, amountPaid, in.Bank, in.PaymentMethod, in.Notes, transactionID,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("inserindo pagamento da conta fixa: %w", err)
	}
	paymentID, err := paymentRes.LastInsertId()
	if err != nil {
		return nil, nil, fmt.Errorf("obtendo id do pagamento: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE transactions SET fixed_bill_payment_id = ? WHERE id = ?`, paymentID, transactionID,
	); err != nil {
		return nil, nil, fmt.Errorf("vinculando transação ao pagamento: %w", err)
	}

	currentDue, err := time.Parse("2006-01-02", dueDate)
	if err != nil {
		currentDue = time.Now().UTC()
	}
	nextDue := NextDueDate(currentDue, periodicity).Format("2006-01-02")

	if _, err := tx.ExecContext(ctx,
		`UPDATE fixed_bills SET due_date = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ?`,
		nextDue, billID,
	); err != nil {
		return nil, nil, fmt.Errorf("avançando vencimento da conta fixa: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("confirmando pagamento da conta fixa: %w", err)
	}

	payment, err := r.GetPayment(ctx, userID, paymentID)
	if err != nil {
		return nil, nil, err
	}
	bill, err := r.GetByID(ctx, userID, billID)
	if err != nil {
		return nil, nil, err
	}
	return payment, bill, nil
}

func (r *Repository) GetPayment(ctx context.Context, userID, id int64) (*Payment, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, fixed_bill_id, due_date, paid_at, amount_paid, bank, payment_method, notes, transaction_id, created_at
		 FROM fixed_bill_payments WHERE id = ? AND user_id = ?`, id, userID,
	)
	return scanPayment(row)
}

func (r *Repository) ListPayments(ctx context.Context, userID, billID int64) ([]Payment, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, fixed_bill_id, due_date, paid_at, amount_paid, bank, payment_method, notes, transaction_id, created_at
		 FROM fixed_bill_payments WHERE fixed_bill_id = ? AND user_id = ? ORDER BY paid_at DESC, id DESC`,
		billID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listando pagamentos da conta fixa: %w", err)
	}
	defer rows.Close()

	payments := []Payment{}
	for rows.Next() {
		p, err := scanPayment(rows)
		if err != nil {
			return nil, err
		}
		payments = append(payments, *p)
	}
	return payments, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanFixedBill(s scanner) (*FixedBill, error) {
	var b FixedBill
	var categoryID, accountID sql.NullInt64
	var categoryName, categoryColor, accountName, notes sql.NullString
	var createdAt, updatedAt string

	if err := s.Scan(
		&b.ID, &b.UserID, &b.Name, &b.Amount, &categoryID, &categoryName, &categoryColor,
		&accountID, &accountName, &b.Periodicity, &b.DueDate, &b.Status, &notes, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo conta fixa: %w", err)
	}

	if categoryID.Valid {
		b.CategoryID = &categoryID.Int64
	}
	if categoryName.Valid {
		b.CategoryName = &categoryName.String
	}
	if categoryColor.Valid {
		b.CategoryColor = &categoryColor.String
	}
	if accountID.Valid {
		b.AccountID = &accountID.Int64
	}
	if accountName.Valid {
		b.AccountName = &accountName.String
	}
	if notes.Valid {
		b.Notes = &notes.String
	}
	b.CreatedAt = parseTimestamp(createdAt)
	b.UpdatedAt = parseTimestamp(updatedAt)
	return &b, nil
}

func scanPayment(s scanner) (*Payment, error) {
	var p Payment
	var bank, paymentMethod, notes sql.NullString
	var transactionID sql.NullInt64
	var createdAt string

	if err := s.Scan(
		&p.ID, &p.FixedBillID, &p.DueDate, &p.PaidAt, &p.AmountPaid,
		&bank, &paymentMethod, &notes, &transactionID, &createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo pagamento: %w", err)
	}

	if bank.Valid {
		p.Bank = &bank.String
	}
	if paymentMethod.Valid {
		p.PaymentMethod = &paymentMethod.String
	}
	if notes.Valid {
		p.Notes = &notes.String
	}
	if transactionID.Valid {
		p.TransactionID = &transactionID.Int64
	}
	p.CreatedAt = parseTimestamp(createdAt)
	return &p, nil
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
