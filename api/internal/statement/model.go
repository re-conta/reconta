// Package statement lê extratos bancários em PDF (Banco do Brasil, Sicredi,
// Nubank, Mercado Pago, Itaú, entre outros) e extrai os lançamentos para
// pré-visualização e importação como transações.
package statement

// ParsedTransaction é um lançamento extraído de um extrato em PDF.
type ParsedTransaction struct {
	Date           string  `json:"date"`
	Description    string  `json:"description"`
	Amount         float64 `json:"amount"`
	Type           string  `json:"type"`
	PixBeneficiary *string `json:"pixBeneficiary,omitempty"`
	CategoryID     *int64  `json:"categoryId,omitempty"`
	CategoryName   *string `json:"categoryName,omitempty"`
	Duplicate      bool    `json:"duplicate"`
	RawLine        string  `json:"-"`
}

// PreviewResult é o retorno da pré-visualização de um extrato importado.
type PreviewResult struct {
	Bank         string              `json:"bank"`
	BankLabel    string              `json:"bankLabel"`
	Transactions []ParsedTransaction `json:"transactions"`
}
