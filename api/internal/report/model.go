package report

import (
	"time"

	"github.com/re-conta/reconta/api/internal/transaction"
)

// Scope representa o período resolvido de um relatório.
type Scope struct {
	DateFrom string // formato YYYY-MM-DD, vazio quando o escopo é "tudo"
	DateTo   string
	Label    string
}

// ChartImage é uma imagem PNG (base64, sem o prefixo data:) de um gráfico
// renderizado no frontend, enviada junto do pedido de exportação.
type ChartImage struct {
	Title     string `json:"title"`
	PNGBase64 string `json:"pngBase64"`
}

// backupVersion identifica o formato do arquivo JSON de backup/restauração.
const backupVersion = 1

// CategoryRef, AccountRef e TagRef identificam categoria/conta/tag pelo nome
// (não pelo id, que pode não existir no banco de destino de uma restauração).
type CategoryRef struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
	Type  string `json:"type"`
}

type AccountRef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type TagRef struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TxRecord é uma transação achatada por nome, usada no backup JSON.
type TxRecord struct {
	Date           string   `json:"date"`
	Description    string   `json:"description"`
	Amount         float64  `json:"amount"`
	Type           string   `json:"type"`
	CategoryName   *string  `json:"categoryName,omitempty"`
	AccountName    *string  `json:"accountName,omitempty"`
	Notes          *string  `json:"notes,omitempty"`
	PixBeneficiary *string  `json:"pixBeneficiary,omitempty"`
	Bank           *string  `json:"bank,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

// BackupPayload é o formato do arquivo JSON exportado/importado como backup.
type BackupPayload struct {
	Version      int                `json:"version"`
	ExportedAt   time.Time          `json:"exportedAt"`
	PeriodLabel  string             `json:"periodLabel"`
	Categories   []CategoryRef      `json:"categories"`
	Accounts     []AccountRef       `json:"accounts"`
	Tags         []TagRef           `json:"tags"`
	Transactions []TxRecord         `json:"transactions"`
	Totals       transaction.Totals `json:"totals"`
}
