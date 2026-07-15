// Package share implementa o compartilhamento de transações/contas bancárias
// entre usuários cadastrados: convite, aceite/rejeição, cancelamento e a
// resolução de acesso (quais contas, qual período, se permite edição) usada
// para servir os dados compartilhados ao convidado.
package share

import "time"

const (
	StatusPending   = "pending"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusCancelled = "cancelled"
)

// Share representa um convite de compartilhamento, aceito ou não.
type Share struct {
	ID            int64      `json:"id"`
	OwnerID       int64      `json:"ownerId"`
	OwnerName     string     `json:"ownerName"`
	RecipientID   int64      `json:"recipientId"`
	RecipientName string     `json:"recipientName"`
	AccountIDs    []int64    `json:"accountIds"`
	AccountNames  []string   `json:"accountNames"`
	CanEdit       bool       `json:"canEdit"`
	IncludeFuture bool       `json:"includeFuture"`
	PeriodStart   *string    `json:"periodStart"`
	PeriodEnd     *string    `json:"periodEnd"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"createdAt"`
	RespondedAt   *time.Time `json:"respondedAt"`
}

// AccessGrant resume o acesso efetivo de um convidado a um compartilhamento
// já aceito: quais contas, qual janela de datas (nil = sem limite) e se pode
// editar os lançamentos.
type AccessGrant struct {
	ShareID    int64
	OwnerID    int64
	AccountIDs []int64
	CanEdit    bool
	DateFrom   *string
	DateTo     *string
}
