// Package fixedbill implementa as contas fixas (despesas recorrentes) do
// usuário: cadastro, ciclo de vida (ativa/congelada/encerrada) e pagamentos,
// que geram transações reais em /transacoes.
package fixedbill

import "time"

const (
	StatusActive = "active"
	StatusFrozen = "frozen"
	StatusClosed = "closed"

	PeriodicityWeekly     = "weekly"
	PeriodicityBiweekly   = "biweekly"
	PeriodicityMonthly    = "monthly"
	PeriodicityBimonthly  = "bimonthly"
	PeriodicityQuarterly  = "quarterly"
	PeriodicitySemiannual = "semiannual"
	PeriodicityAnnual     = "annual"
	PeriodicityBiennial   = "biennial"
)

var validPeriodicities = map[string]bool{
	PeriodicityWeekly:     true,
	PeriodicityBiweekly:   true,
	PeriodicityMonthly:    true,
	PeriodicityBimonthly:  true,
	PeriodicityQuarterly:  true,
	PeriodicitySemiannual: true,
	PeriodicityAnnual:     true,
	PeriodicityBiennial:   true,
}

// IsValidPeriodicity indica se o valor informado é uma periodicidade suportada.
func IsValidPeriodicity(p string) bool {
	return validPeriodicities[p]
}

// FixedBill representa uma conta fixa (despesa recorrente) do usuário.
type FixedBill struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"-"`
	Name          string    `json:"name"`
	Amount        float64   `json:"amount"`
	CategoryID    *int64    `json:"categoryId"`
	CategoryName  *string   `json:"categoryName,omitempty"`
	CategoryColor *string   `json:"categoryColor,omitempty"`
	AccountID     *int64    `json:"accountId"`
	AccountName   *string   `json:"accountName,omitempty"`
	Periodicity   string    `json:"periodicity"`
	DueDate       string    `json:"dueDate"`
	Status        string    `json:"status"`
	Notes         *string   `json:"notes"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// Payment representa um pagamento efetuado para um ciclo de uma conta fixa.
type Payment struct {
	ID            int64     `json:"id"`
	FixedBillID   int64     `json:"fixedBillId"`
	DueDate       string    `json:"dueDate"`
	PaidAt        string    `json:"paidAt"`
	AmountPaid    float64   `json:"amountPaid"`
	Bank          *string   `json:"bank"`
	PaymentMethod *string   `json:"paymentMethod"`
	Notes         *string   `json:"notes"`
	TransactionID *int64    `json:"transactionId"`
	CreatedAt     time.Time `json:"createdAt"`
}

// NextDueDate calcula a próxima data de vencimento a partir da data atual do
// ciclo e da periodicidade da conta, usando AddDate para lidar corretamente
// com meses de tamanhos diferentes.
func NextDueDate(current time.Time, periodicity string) time.Time {
	switch periodicity {
	case PeriodicityWeekly:
		return current.AddDate(0, 0, 7)
	case PeriodicityBiweekly:
		return current.AddDate(0, 0, 14)
	case PeriodicityBimonthly:
		return current.AddDate(0, 2, 0)
	case PeriodicityQuarterly:
		return current.AddDate(0, 3, 0)
	case PeriodicitySemiannual:
		return current.AddDate(0, 6, 0)
	case PeriodicityAnnual:
		return current.AddDate(1, 0, 0)
	case PeriodicityBiennial:
		return current.AddDate(2, 0, 0)
	case PeriodicityMonthly:
		fallthrough
	default:
		return current.AddDate(0, 1, 0)
	}
}
