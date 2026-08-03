package advisor

import (
	"fmt"
	"strings"
)

// systemPrompt instrui o Groq a atuar como consultor financeiro brasileiro e
// a responder em JSON estrito, no formato que sanitizeRecommendations espera.
const systemPrompt = `Você é um consultor financeiro pessoal brasileiro, direto, prático e criativo.
Responda SEMPRE em português do Brasil, em um único objeto JSON válido, sem nenhum texto fora do JSON.
Formato exato: {"recommendations": [{"kind": "cut", "title": "...", "description": "...", "impact": "..."}]}.
"kind" deve ser "cut" (corte/redução de gastos) ou "invest" (investir o que sobrou).
Gere entre 3 e 6 recomendações, específicas e acionáveis, baseadas nos dados informados pelo usuário.
Nunca gere recomendações genéricas como apenas "gaste menos" ou "economize mais" sem uma ação concreta.`

// buildPrompt monta o prompt do usuário com os sinais extraídos das
// transações do mês, para o Groq gerar recomendações de corte de gastos e,
// quando a saúde financeira permitir, de investimento do valor excedente.
func buildPrompt(s Signals, income, expense, rate float64, level string, stars int) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Receita do mês: R$ %.2f\n", income)
	fmt.Fprintf(&b, "Despesa do mês: R$ %.2f\n", expense)
	fmt.Fprintf(&b, "Saldo do mês: R$ %.2f (taxa de poupança: %.1f%%)\n", income-expense, rate)
	fmt.Fprintf(&b, "Nível de saúde financeira: %s (%d de 5 estrelas)\n\n", level, stars)

	if len(s.TopExpenseCategories) > 0 {
		b.WriteString("Maiores categorias de despesa no mês:\n")
		for _, c := range s.TopExpenseCategories {
			fmt.Fprintf(&b, "- %s: R$ %.2f\n", c.Name, c.Total)
		}
		b.WriteString("\n")
	}

	if len(s.StreamingServices) > 1 {
		fmt.Fprintf(&b, "Foram detectadas %d assinaturas de streaming/vídeo/música distintas neste mês (%s), somando R$ %.2f. "+
			"Gere uma recomendação do tipo \"cut\" sugerindo cancelar as excedentes e manter apenas uma.\n\n",
			len(s.StreamingServices), strings.Join(s.StreamingServices, "; "), s.StreamingTotal)
	}

	if s.FuelCount > 0 {
		fmt.Fprintf(&b, "Foram encontrados %d lançamentos de combustível neste mês, somando R$ %.2f. "+
			"Gere uma recomendação do tipo \"cut\" com dicas práticas de economia (comparar preços entre postos da região do usuário, "+
			"apps de comparação de combustível, direção econômica, etc.).\n\n", s.FuelCount, s.FuelTotal)
	}

	if stars >= 3 && income-expense > 0 {
		fmt.Fprintf(&b, "A saúde financeira está boa e sobraram R$ %.2f neste mês. Inclua também 1 a 3 recomendações do tipo \"invest\" "+
			"com ideias concretas e adequadas para quem está começando a investir (ex.: Tesouro Direto, CDB com liquidez diária, fundos DI, "+
			"reserva de emergência) para esse valor excedente.\n\n", income-expense)
	} else {
		b.WriteString("Não inclua nenhuma recomendação do tipo \"invest\" neste caso — foque exclusivamente em cortes de gastos (\"cut\").\n\n")
	}

	b.WriteString("Além dos pontos citados acima, analise os demais dados e seja criativo: procure outros padrões de gasto que valham uma recomendação.")

	return b.String()
}
