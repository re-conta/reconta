package taxsim

// Compute aplica a tabela progressiva anual sobre a renda tributável
// informada e devolve o imposto estimado, a alíquota efetiva e o
// detalhamento por faixa (quanto da renda caiu em cada uma).
func Compute(year int, totalIncome float64) Result {
	result := Result{
		Enabled:     true,
		Year:        year,
		TotalIncome: totalIncome,
		Source:      tableSource,
		Brackets:    make([]BracketResult, len(annualBrackets)),
	}

	if totalIncome < 0 {
		totalIncome = 0
	}

	var tax float64
	var marginalRate float64
	lowerBound := 0.0
	for i, b := range annualBrackets {
		upper := b.UpTo
		if upper == 0 || upper > totalIncome {
			upper = totalIncome
		}

		taxableInBand := upper - lowerBound
		if taxableInBand < 0 {
			taxableInBand = 0
		}

		isIncomeBracket := totalIncome > lowerBound
		result.Brackets[i] = BracketResult{
			Rate:            b.Rate,
			UpTo:            b.UpTo,
			Deduction:       b.Deduction,
			TaxableInBand:   taxableInBand,
			IsIncomeBracket: isIncomeBracket,
		}

		if isIncomeBracket {
			marginalRate = b.Rate
		}

		if b.UpTo == 0 {
			lowerBound = totalIncome
		} else {
			lowerBound = b.UpTo
		}
	}

	// O imposto é calculado diretamente pela fórmula "renda × alíquota da
	// faixa − parcela a deduzir" da própria faixa em que a renda se encontra,
	// já que as deduções da tabela tornam esse cálculo equivalente à soma
	// progressiva por faixa.
	bracket := bracketFor(totalIncome)
	tax = totalIncome*bracket.Rate - bracket.Deduction
	if tax < 0 {
		tax = 0
	}

	result.EstimatedTax = tax
	result.MarginalRate = marginalRate
	if totalIncome > 0 {
		result.EffectiveRate = tax / totalIncome * 100
	}
	return result
}

func bracketFor(income float64) Bracket {
	for _, b := range annualBrackets {
		if b.UpTo == 0 || income <= b.UpTo {
			return b
		}
	}
	return annualBrackets[len(annualBrackets)-1]
}
