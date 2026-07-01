package statement

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseStatement extrai os lançamentos do texto de um extrato bancário.
//
// Os bancos suportados (Banco do Brasil, Sicredi, Nubank, Mercado Pago,
// Itaú) exportam PDFs com layouts diferentes, mas todos seguem o mesmo
// padrão de linha: data no início, valor monetário (com marcador opcional
// de crédito/débito) em algum ponto da linha, e a descrição entre os dois.
// Por isso um único motor genérico cobre todos eles; o bankKey só existe
// para permitir ajustes futuros específicos de um banco, se necessário.
func ParseStatement(bankKey string, text string) []ParsedTransaction {
	_ = bankKey
	year := fallbackYear(text)

	var results []ParsedTransaction
	for line := range strings.SplitSeq(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if tx, ok := parseLine(line, year); ok {
			results = append(results, tx)
		}
	}
	return results
}

var (
	dateNumericRe = regexp.MustCompile(`^(\d{2})/(\d{2})(?:/(\d{2,4}))?\s+(.*)$`)
	dateMonthRe   = regexp.MustCompile(`(?i)^(\d{1,2})\s+(jan|fev|mar|abr|mai|jun|jul|ago|set|out|nov|dez)[a-zç]*\.?\s+(\d{4})?\s*(.*)$`)
	amountRe      = regexp.MustCompile(`(?i)(-)?\s*(?:R\$\s*)?(\d{1,3}(?:\.\d{3})*,\d{2})(-)?\s*(DB|D|CR|C)?`)
	periodYearRe  = regexp.MustCompile(`\d{2}/\d{2}/(\d{4})`)
	pixNameRe     = regexp.MustCompile(`(?i)pix\s+(?:enviado|recebido|transferido)(?:\s+(?:para|de))?\s*[:\-]?\s*(.+)$`)

	monthAbbrev = map[string]string{
		"jan": "01", "fev": "02", "mar": "03", "abr": "04", "mai": "05", "jun": "06",
		"jul": "07", "ago": "08", "set": "09", "out": "10", "nov": "11", "dez": "12",
	}

	incomeHints = []string{
		"recebido", "recebimento", "credito", "crédito", "deposito", "depósito",
		"estorno", "reembolso", "salario", "salário", "rendimento",
		"transferencia recebida", "transferência recebida", "ted recebida", "doc recebido",
	}
	expenseHints = []string{
		"pago", "pagamento de", "compra", "debito", "débito", "enviado", "saque",
		"boleto", "fatura", "tarifa", "anuidade", "encargo", "juros",
		"transferencia enviada", "transferência enviada", "ted enviada", "doc enviado",
	}
)

func fallbackYear(text string) int {
	if m := periodYearRe.FindStringSubmatch(text); m != nil {
		if y, err := strconv.Atoi(m[1]); err == nil {
			return y
		}
	}
	return time.Now().Year()
}

func parseLine(line string, defaultYear int) (ParsedTransaction, bool) {
	date, rest, ok := extractDate(line, defaultYear)
	if !ok {
		return ParsedTransaction{}, false
	}

	amount, sign, rest, ok := extractAmount(rest)
	if !ok || amount == 0 {
		return ParsedTransaction{}, false
	}

	description := strings.Trim(strings.TrimSpace(rest), "-–—:; ")
	if description == "" {
		return ParsedTransaction{}, false
	}

	tx := ParsedTransaction{
		Date:        date,
		Description: description,
		Amount:      amount,
		Type:        classify(description, sign),
		RawLine:     line,
	}
	if m := pixNameRe.FindStringSubmatch(description); m != nil {
		if name := strings.TrimSpace(m[1]); name != "" {
			tx.PixBeneficiary = &name
		}
	}
	return tx, true
}

// extractDate reconhece datas no início da linha em formato numérico
// (dd/mm, dd/mm/yy ou dd/mm/yyyy) ou por extenso abreviada (ex.: "01 mar"),
// usado por alguns extratos como o do Nubank.
func extractDate(line string, defaultYear int) (date string, rest string, ok bool) {
	if m := dateNumericRe.FindStringSubmatch(line); m != nil {
		day, month, yearPart, tail := m[1], m[2], m[3], m[4]
		year := defaultYear
		if yearPart != "" {
			if len(yearPart) == 2 {
				yearPart = "20" + yearPart
			}
			if y, err := strconv.Atoi(yearPart); err == nil {
				year = y
			}
		}
		d, errD := strconv.Atoi(day)
		mo, errM := strconv.Atoi(month)
		if errD != nil || errM != nil || d < 1 || d > 31 || mo < 1 || mo > 12 {
			return "", "", false
		}
		return fmt.Sprintf("%04d-%02d-%02d", year, mo, d), tail, true
	}

	if m := dateMonthRe.FindStringSubmatch(line); m != nil {
		day, monthAbbr, yearPart, tail := m[1], strings.ToLower(m[2]), m[3], m[4]
		mo, known := monthAbbrev[monthAbbr]
		if !known {
			return "", "", false
		}
		year := defaultYear
		if yearPart != "" {
			if y, err := strconv.Atoi(yearPart); err == nil {
				year = y
			}
		}
		d, err := strconv.Atoi(day)
		if err != nil || d < 1 || d > 31 {
			return "", "", false
		}
		return fmt.Sprintf("%04d-%s-%02d", year, mo, d), tail, true
	}

	return "", "", false
}

// extractAmount localiza o primeiro valor monetário da linha (o lançamento
// em si) e ignora eventuais valores seguintes (ex.: saldo acumulado da
// linha). O sinal é 1 (crédito), -1 (débito) ou 0 quando não há marcador
// explícito de sinal/D-C na linha.
func extractAmount(s string) (amount float64, sign int, description string, ok bool) {
	loc := amountRe.FindStringSubmatchIndex(s)
	if loc == nil {
		return 0, 0, s, false
	}

	group := func(i int) string {
		if loc[2*i] < 0 {
			return ""
		}
		return s[loc[2*i]:loc[2*i+1]]
	}

	leadingSign := group(1)
	numeric := group(2)
	trailingSign := group(3)
	suffix := strings.ToUpper(group(4))

	normalized := strings.NewReplacer(".", "", ",", ".").Replace(numeric)
	value, err := strconv.ParseFloat(normalized, 64)
	if err != nil {
		return 0, 0, s, false
	}

	switch {
	case leadingSign == "-" || trailingSign == "-" || suffix == "D" || suffix == "DB":
		sign = -1
	case suffix == "C" || suffix == "CR":
		sign = 1
	}

	return value, sign, s[:loc[0]], true
}

// classify decide receita/despesa: usa o sinal explícito quando disponível
// e, na ausência dele, palavras-chave comuns em descrições de extratos.
func classify(description string, sign int) string {
	if sign > 0 {
		return "income"
	}
	if sign < 0 {
		return "expense"
	}

	lower := strings.ToLower(description)
	for _, h := range incomeHints {
		if strings.Contains(lower, h) {
			return "income"
		}
	}
	for _, h := range expenseHints {
		if strings.Contains(lower, h) {
			return "expense"
		}
	}
	return "expense"
}
