package advisor

import (
	"regexp"
	"sort"
	"strings"

	"github.com/re-conta/reconta/api/internal/transaction"
)

// streamingPattern e fuelPattern detectam, por palavras-chave na descrição do
// lançamento, assinaturas de streaming e gastos com combustível — os dois
// exemplos concretos citados no pedido do produto. O restante da análise é
// deixado a cargo do Groq, que recebe as maiores categorias de despesa e é
// instruído a ser criativo além desses dois casos.
var streamingPattern = regexp.MustCompile(`(?i)netflix|spotify|disney\+?|hbo\s*max|amazon\s*prime|prime\s*video|deezer|youtube\s*premium|paramount\+?|globoplay|apple\s*tv|crunchyroll|star\+|claro\s*video|now\s*(nbo|hbo)`)

var fuelPattern = regexp.MustCompile(`(?i)\bposto\b|combust[íi]vel|gasolina|etanol|\bgnv\b|\bshell\b|ipiranga|petrobras|br\s*mania|alesat|raizen`)

// CategoryTotal resume o total gasto em uma categoria no período analisado.
type CategoryTotal struct {
	Name  string
	Total float64
}

// Signals são os padrões extraídos deterministicamente das transações,
// enviados como contexto para o Groq além dos totais de receita/despesa.
type Signals struct {
	StreamingServices    []string
	StreamingTotal       float64
	FuelTotal            float64
	FuelCount            int
	TopExpenseCategories []CategoryTotal
}

func analyzeSignals(txs []transaction.Transaction) Signals {
	var s Signals
	seenStreaming := map[string]bool{}
	categoryTotals := map[string]float64{}

	for _, t := range txs {
		if t.Type != "expense" || t.IsTransfer {
			continue
		}
		if t.CategoryName != nil && *t.CategoryName != "" {
			categoryTotals[*t.CategoryName] += t.Amount
		}

		if match := streamingPattern.FindString(t.Description); match != "" {
			key := strings.ToLower(match)
			if !seenStreaming[key] {
				seenStreaming[key] = true
				s.StreamingServices = append(s.StreamingServices, t.Description)
			}
			s.StreamingTotal += t.Amount
		}

		if fuelPattern.MatchString(t.Description) {
			s.FuelCount++
			s.FuelTotal += t.Amount
		}
	}

	for name, total := range categoryTotals {
		s.TopExpenseCategories = append(s.TopExpenseCategories, CategoryTotal{Name: name, Total: total})
	}
	sort.Slice(s.TopExpenseCategories, func(i, j int) bool {
		return s.TopExpenseCategories[i].Total > s.TopExpenseCategories[j].Total
	})
	if len(s.TopExpenseCategories) > 5 {
		s.TopExpenseCategories = s.TopExpenseCategories[:5]
	}

	return s
}
