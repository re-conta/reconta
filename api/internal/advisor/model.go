package advisor

// Recommendation é uma sugestão de corte de gastos ou investimento gerada
// pelo Groq a partir da análise das transações do usuário no mês.
type Recommendation struct {
	Kind        string `json:"kind"` // "cut" (corte de gastos) ou "invest" (investimento)
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact,omitempty"`
}

// sanitize normaliza o que o modelo retornou, sem confiar cegamente no valor
// de Kind — qualquer coisa fora do esperado vira "cut".
func sanitizeRecommendations(items []Recommendation) []Recommendation {
	out := make([]Recommendation, 0, len(items))
	for _, it := range items {
		if it.Title == "" || it.Description == "" {
			continue
		}
		if it.Kind != "invest" {
			it.Kind = "cut"
		}
		out = append(out, it)
	}
	return out
}
