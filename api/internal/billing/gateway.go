package billing

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mercadopago/sdk-go/pkg/config"
	"github.com/mercadopago/sdk-go/pkg/payment"
	"github.com/mercadopago/sdk-go/pkg/refund"
)

// ErrGatewayDisabled indica que MP_ACCESS_TOKEN não foi configurado — as
// rotas de checkout respondem 503 em vez de quebrar o restante do site.
var ErrGatewayDisabled = errors.New("integração com Mercado Pago não configurada")

// Gateway encapsula as chamadas ao Mercado Pago (Checkout API / Pagamentos).
type Gateway struct {
	payments payment.Client
	refunds  refund.Client
}

// NewGateway cria o gateway a partir do access token. Token vazio retorna
// nil (gateway desabilitado), o que as rotas tratam com ErrGatewayDisabled.
func NewGateway(accessToken string) (*Gateway, error) {
	if accessToken == "" {
		return nil, nil
	}
	cfg, err := config.New(accessToken)
	if err != nil {
		return nil, fmt.Errorf("configurando SDK do Mercado Pago: %w", err)
	}
	return &Gateway{
		payments: payment.NewClient(cfg),
		refunds:  refund.NewClient(cfg),
	}, nil
}

// PaymentInput reúne os dados para criar um pagamento no Mercado Pago.
// Token/Installments/IssuerID só se aplicam a cartões; endereço só a boleto.
type PaymentInput struct {
	Amount            float64
	Description       string
	Method            string // pix | boleto | debit_card | credit_card
	PaymentMethodID   string // pix, bolbradesco, visa, debvisa, master, ...
	Token             string
	Installments      int
	IssuerID          string
	PayerEmail        string
	PayerFirstName    string
	PayerLastName     string
	DocType           string // CPF | CNPJ
	DocNumber         string
	ZipCode           string
	StreetName        string
	StreetNumber      string
	Neighborhood      string
	City              string
	FederalUnit       string
	ExternalReference string
	NotificationURL   string
}

// PaymentResult é o subconjunto da resposta do Mercado Pago que interessa ao
// checkout: status e, conforme o método, QR Code PIX ou link do boleto.
type PaymentResult struct {
	ID           int64
	Status       string
	StatusDetail string
	QRCode       string
	QRCodeBase64 string
	TicketURL    string
}

// CreatePayment cria o pagamento na Checkout API do Mercado Pago.
func (g *Gateway) CreatePayment(ctx context.Context, in PaymentInput) (*PaymentResult, error) {
	if g == nil {
		return nil, ErrGatewayDisabled
	}

	req := payment.Request{
		TransactionAmount: in.Amount,
		Description:       in.Description,
		ExternalReference: in.ExternalReference,
		NotificationURL:   in.NotificationURL,
		StatementDescriptor: "RECONTA",
		Payer: &payment.PayerRequest{
			Email:     in.PayerEmail,
			FirstName: in.PayerFirstName,
			LastName:  in.PayerLastName,
		},
	}
	if in.DocNumber != "" {
		req.Payer.Identification = &payment.IdentificationRequest{
			Type:   strings.ToUpper(in.DocType),
			Number: in.DocNumber,
		}
	}

	switch in.Method {
	case MethodPix:
		req.PaymentMethodID = "pix"
		// PIX de assinatura expira rápido: se não for pago, o usuário
		// simplesmente gera outro no próximo checkout.
		exp := time.Now().Add(30 * time.Minute)
		req.DateOfExpiration = &exp

	case MethodBoleto:
		req.PaymentMethodID = "bolbradesco"
		exp := time.Now().AddDate(0, 0, 3)
		req.DateOfExpiration = &exp
		if in.ZipCode != "" {
			req.Payer.Address = &payment.AddressRequest{
				ZipCode:      in.ZipCode,
				StreetName:   in.StreetName,
				StreetNumber: in.StreetNumber,
				Neighborhood: in.Neighborhood,
				City:         in.City,
				FederalUnit:  in.FederalUnit,
			}
		}

	case MethodDebit, MethodCredit:
		req.PaymentMethodID = in.PaymentMethodID
		req.Token = in.Token
		req.IssuerID = in.IssuerID
		req.Installments = in.Installments
		if req.Installments <= 0 {
			req.Installments = 1
		}
		// binary_mode: cartão aprova ou recusa na hora, sem ficar "em análise".
		req.BinaryMode = true

	default:
		return nil, fmt.Errorf("método de pagamento inválido: %s", in.Method)
	}

	resp, err := g.payments.Create(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("criando pagamento no Mercado Pago: %w", err)
	}
	return resultFromResponse(resp), nil
}

// GetPayment consulta um pagamento existente no Mercado Pago.
func (g *Gateway) GetPayment(ctx context.Context, id int64) (*PaymentResult, error) {
	if g == nil {
		return nil, ErrGatewayDisabled
	}
	resp, err := g.payments.Get(ctx, int(id))
	if err != nil {
		return nil, fmt.Errorf("consultando pagamento %d no Mercado Pago: %w", id, err)
	}
	return resultFromResponse(resp), nil
}

// PartialRefund devolve parte do valor de um pagamento aprovado (usado no
// cancelamento proporcional ao tempo não usado).
func (g *Gateway) PartialRefund(ctx context.Context, paymentID int64, amount float64) error {
	if g == nil {
		return ErrGatewayDisabled
	}
	if _, err := g.refunds.CreatePartialRefund(ctx, int(paymentID), amount); err != nil {
		return fmt.Errorf("criando reembolso parcial no Mercado Pago: %w", err)
	}
	return nil
}

func resultFromResponse(resp *payment.Response) *PaymentResult {
	data := resp.PointOfInteraction.TransactionData
	return &PaymentResult{
		ID:           int64(resp.ID),
		Status:       resp.Status,
		StatusDetail: resp.StatusDetail,
		QRCode:       data.QRCode,
		QRCodeBase64: data.QRCodeBase64,
		TicketURL:    firstNonEmpty(data.TicketURL, resp.TransactionDetails.ExternalResourceURL),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
