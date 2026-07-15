package billing

import "time"

// Ciclos de cobrança de uma assinatura.
const (
	CycleMonthly = "monthly"
	CycleYearly  = "yearly"
)

// Status de uma assinatura.
const (
	StatusPending  = "pending"  // aguardando confirmação do primeiro pagamento
	StatusActive   = "active"   // paga e dentro do período vigente
	StatusCanceled = "canceled" // cancelada pelo usuário (com ou sem reembolso)
	StatusExpired  = "expired"  // período vigente terminou sem renovação
)

// Métodos de pagamento aceitos no checkout.
const (
	MethodPix    = "pix"
	MethodBoleto = "boleto"
	MethodDebit  = "debit_card"
	MethodCredit = "credit_card"
)

// PlanFree é o código do plano gratuito — usuários sem assinatura ativa
// pertencem a ele e não passam pelo checkout.
const PlanFree = "gratuito"

// Plan é um plano do site. Preços e benefícios dos planos pagos são
// configuráveis no painel de admin.
type Plan struct {
	ID           int64    `json:"id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	PriceMonthly float64  `json:"priceMonthly"`
	PriceYearly  float64  `json:"priceYearly"`
	Benefits     []string `json:"benefits"`
	Highlight    bool     `json:"highlight"`
	SortOrder    int      `json:"sortOrder"`
}

// Subscription é a assinatura de um usuário em um plano pago.
type Subscription struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"userId"`
	PlanID            int64      `json:"planId"`
	PlanCode          string     `json:"planCode"`
	PlanName          string     `json:"planName"`
	Cycle             string     `json:"cycle"`
	Status            string     `json:"status"`
	PaymentMethod     string     `json:"paymentMethod"`
	StartedAt         *time.Time `json:"startedAt"`
	CurrentPeriodEnd  *time.Time `json:"currentPeriodEnd"`
	CancelAtPeriodEnd bool       `json:"cancelAtPeriodEnd"`
	CanceledAt        *time.Time `json:"canceledAt"`
	RefundAmount      *float64   `json:"refundAmount"`
	LastReminderDays  *int       `json:"-"`
	CreatedAt         time.Time  `json:"createdAt"`
}

// Payment é uma cobrança individual de assinatura (primeira compra ou
// renovação) espelhando o pagamento correspondente no Mercado Pago.
type Payment struct {
	ID             int64     `json:"id"`
	SubscriptionID int64     `json:"subscriptionId"`
	UserID         int64     `json:"userId"`
	MPPaymentID    *int64    `json:"mpPaymentId"`
	Amount         float64   `json:"amount"`
	Method         string    `json:"method"`
	Status         string    `json:"status"`
	StatusDetail   string    `json:"statusDetail"`
	PixQR          string    `json:"pixQr,omitempty"`
	PixQRBase64    string    `json:"pixQrBase64,omitempty"`
	TicketURL      string    `json:"ticketUrl,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
}

// PriceFor retorna o preço do plano para o ciclo informado.
func (p *Plan) PriceFor(cycle string) float64 {
	if cycle == CycleYearly {
		return p.PriceYearly
	}
	return p.PriceMonthly
}

// IsFree indica se este é o plano gratuito.
func (p *Plan) IsFree() bool {
	return p.Code == PlanFree
}

// ValidCycle valida o ciclo de cobrança recebido do cliente.
func ValidCycle(cycle string) bool {
	return cycle == CycleMonthly || cycle == CycleYearly
}

// ValidMethod valida o método de pagamento recebido do cliente.
func ValidMethod(method string) bool {
	switch method {
	case MethodPix, MethodBoleto, MethodDebit, MethodCredit:
		return true
	}
	return false
}
