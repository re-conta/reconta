package fixedbill

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/re-conta/reconta/api/internal/db"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	// Arquivo temporário por teste: "file::memory:" vira um arquivo literal
	// com esse nome no DSN e persistia dados entre execuções.
	conn, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("abrindo banco de teste: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func createTestUser(t *testing.T, conn *sql.DB) int64 {
	t.Helper()
	res, err := conn.Exec(`INSERT INTO users (name, email, password_hash) VALUES ('Teste', 'teste@example.com', 'hash')`)
	if err != nil {
		t.Fatalf("criando usuário de teste: %v", err)
	}
	id, _ := res.LastInsertId()
	return id
}

func TestRepositoryPayFlow(t *testing.T) {
	conn := newTestDB(t)
	ctx := context.Background()
	userID := createTestUser(t, conn)

	repo := NewRepository(conn)
	bill, err := repo.Create(ctx, userID, Input{
		Name:        "Energia elétrica",
		Amount:      150.50,
		Periodicity: PeriodicityMonthly,
		DueDate:     "2026-01-10",
	})
	if err != nil {
		t.Fatalf("criando conta fixa: %v", err)
	}

	payment, updatedBill, err := repo.Pay(ctx, userID, bill.ID, PayInput{})
	if err != nil {
		t.Fatalf("pagando conta fixa: %v", err)
	}

	if payment.AmountPaid != 150.50 {
		t.Errorf("valor pago = %v, esperava 150.50 (valor padrão da conta)", payment.AmountPaid)
	}
	if payment.TransactionID == nil {
		t.Fatal("esperava transactionId preenchido")
	}
	if updatedBill.DueDate != "2026-02-10" {
		t.Errorf("due_date após pagamento = %s, esperava 2026-02-10", updatedBill.DueDate)
	}

	var txAmount float64
	var txType string
	var fixedBillPaymentID sql.NullInt64
	err = conn.QueryRowContext(ctx,
		`SELECT amount, type, fixed_bill_payment_id FROM transactions WHERE id = ?`, *payment.TransactionID,
	).Scan(&txAmount, &txType, &fixedBillPaymentID)
	if err != nil {
		t.Fatalf("lendo transação gerada: %v", err)
	}
	if txAmount != 150.50 || txType != "expense" {
		t.Errorf("transação gerada inesperada: amount=%v type=%s", txAmount, txType)
	}
	if !fixedBillPaymentID.Valid || fixedBillPaymentID.Int64 != payment.ID {
		t.Errorf("fixed_bill_payment_id da transação não vinculado corretamente")
	}

	// Pagar uma conta congelada deve falhar.
	if _, err := repo.UpdateStatus(ctx, userID, bill.ID, StatusFrozen); err != nil {
		t.Fatalf("congelando conta fixa: %v", err)
	}
	if _, _, err := repo.Pay(ctx, userID, bill.ID, PayInput{}); err != ErrNotActive {
		t.Errorf("pagar conta congelada = %v, esperava ErrNotActive", err)
	}
}
