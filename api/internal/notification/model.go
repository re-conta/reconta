// Package notification implementa as notificações de contas fixas vencendo
// ou vencidas: geração periódica (via rota interna acionada pelo timer
// systemd), entrega em tempo real no site (SSE) e por e-mail, além das
// preferências de lembrete configuráveis por usuário.
package notification

import "time"

const (
	KindDueSoon = "bill_due_soon"
	KindOverdue = "bill_overdue"
)

// Notification representa um lembrete gerado para o usuário.
type Notification struct {
	ID            int64      `json:"id"`
	FixedBillID   *int64     `json:"fixedBillId"`
	FixedBillName *string    `json:"fixedBillName,omitempty"`
	Kind          string     `json:"kind"`
	Title         string     `json:"title"`
	Message       string     `json:"message"`
	DueDate       string     `json:"dueDate"`
	ReadAt        *time.Time `json:"readAt"`
	EmailSentAt   *time.Time `json:"emailSentAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

// Settings são as preferências de notificação de um usuário.
type Settings struct {
	SiteEnabled  bool  `json:"siteEnabled"`
	EmailEnabled bool  `json:"emailEnabled"`
	Offsets      []int `json:"offsets"` // minutos de antecedência antes do vencimento
}

// DefaultOffsets usados quando o usuário ainda não configurou preferências.
var DefaultOffsets = []int{2880, 1440, 120, 60}
