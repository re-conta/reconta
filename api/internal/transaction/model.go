package transaction

import (
	"time"

	"github.com/lucasbrum/reconta/api/internal/tag"
)

// Transaction representa um lançamento financeiro (receita ou despesa).
type Transaction struct {
	ID             int64     `json:"id"`
	Date           string    `json:"date"`
	Description    string    `json:"description"`
	Amount         float64   `json:"amount"`
	Type           string    `json:"type"`
	CategoryID     *int64    `json:"categoryId"`
	CategoryName   *string   `json:"categoryName,omitempty"`
	CategoryColor  *string   `json:"categoryColor,omitempty"`
	AccountID      *int64    `json:"accountId"`
	Notes          *string   `json:"notes"`
	ImportedFrom   *string   `json:"importedFrom"`
	Bank           *string   `json:"bank"`
	PixBeneficiary *string   `json:"pixBeneficiary"`
	CreatedAt      time.Time `json:"createdAt"`
	Tags           []tag.Tag `json:"tags"`
}

// Totals resume receitas, despesas e saldo de um conjunto de transações.
type Totals struct {
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
	Count   int     `json:"count"`
}

// Pagination descreve a página atual de uma listagem.
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// ListResult é o retorno paginado de uma listagem de transações.
type ListResult struct {
	Data       []Transaction `json:"data"`
	Totals     Totals        `json:"totals"`
	Pagination Pagination    `json:"pagination"`
}

// ListFilters agrupa os filtros aceitos pela listagem de transações.
type ListFilters struct {
	Month      int
	Year       int
	Type       string
	CategoryID int64
	TagID      int64
	Search     string
	Page       int
	Limit      int
}
