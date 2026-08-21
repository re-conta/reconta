package taxsim

// annualBrackets é a tabela progressiva anual do IRPF derivada da tabela
// mensal oficial vigente desde 01/02/2024 (Lei 14.848/2024, Receita Federal
// do Brasil), com os limites e parcelas a deduzir multiplicados por 12 —
// mesmo critério usado pela Receita para a "declaração de ajuste anual"
// quando a renda mensal é aproximadamente constante ao longo do ano.
//
// Tabela mensal de referência (rendimentos.economia.gov.br / gov.br/receitafederal):
//
//	Até R$ 2.259,20:                     isento
//	De R$ 2.259,21 até R$ 2.826,65:       7,5%  — deduzir R$   169,44
//	De R$ 2.826,66 até R$ 3.751,05:      15,0%  — deduzir R$   381,44
//	De R$ 3.751,06 até R$ 4.664,68:      22,5%  — deduzir R$   662,77
//	Acima de R$ 4.664,68:                27,5%  — deduzir R$   896,00
//
// Como a lei não mudou desde então, a tabela abaixo vale tanto para o ano
// de referência 2024 quanto para 2025 e 2026. Sempre que a Receita Federal
// publicar uma nova tabela, atualize apenas os valores aqui.
var annualBrackets = []Bracket{
	{UpTo: 27110.40, Rate: 0, Deduction: 0},
	{UpTo: 33919.80, Rate: 0.075, Deduction: 2033.28},
	{UpTo: 45012.60, Rate: 0.15, Deduction: 4577.28},
	{UpTo: 55976.16, Rate: 0.225, Deduction: 7953.24},
	{UpTo: 0, Rate: 0.275, Deduction: 10760.64},
}

const tableSource = "Tabela progressiva mensal do IRPF vigente desde 01/02/2024 (Lei 14.848/2024), " +
	"com limites e deduções anualizados (×12). Estimativa simplificada, sem deduções por " +
	"dependentes, saúde, educação ou previdência."
