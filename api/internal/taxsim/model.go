// Package taxsim simula, de forma simplificada, o Imposto de Renda Pessoa
// Física (IRPF) devido sobre as receitas tributáveis do usuário em um ano,
// usando a tabela progressiva oficial da Receita Federal. É uma estimativa:
// não substitui a declaração de ajuste anual, pois não considera deduções
// por dependentes, saúde, educação, previdência ou o desconto simplificado.
package taxsim

// Bracket é uma faixa da tabela progressiva anual do IRPF.
type Bracket struct {
	// UpTo é o teto da faixa (em reais). Zero indica "sem teto" (última faixa).
	UpTo float64 `json:"upTo"`
	Rate float64 `json:"rate"`
	// Deduction é a parcela a deduzir do imposto calculado com a alíquota da
	// faixa, de forma que o cálculo permaneça contínuo entre faixas.
	Deduction float64 `json:"deduction"`
}

// BracketResult descreve quanto da renda caiu em uma faixa e o imposto correspondente.
type BracketResult struct {
	Rate            float64 `json:"rate"`
	UpTo            float64 `json:"upTo"`
	Deduction       float64 `json:"deduction"`
	TaxableInBand   float64 `json:"taxableInBand"`
	IsIncomeBracket bool    `json:"isIncomeBracket"`
}

// Result é a simulação completa de IR para um ano.
type Result struct {
	Enabled       bool            `json:"enabled"`
	Year          int             `json:"year"`
	TotalIncome   float64         `json:"totalIncome"`
	EstimatedTax  float64         `json:"estimatedTax"`
	EffectiveRate float64         `json:"effectiveRate"`
	MarginalRate  float64         `json:"marginalRate"`
	Brackets      []BracketResult `json:"brackets"`
	Source        string          `json:"source"`
}
