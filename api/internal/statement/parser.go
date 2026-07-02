package statement

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ParseStatement extrai os lançamentos do texto de um extrato bancário.
//
// Os bancos suportados (Banco do Brasil, Sicredi, Bradesco, Nubank, Mercado
// Pago, Itaú) exportam PDFs com layouts diferentes. A maioria segue um
// padrão de linha única: data no início, valor monetário (com marcador
// opcional de crédito/débito) em algum ponto da linha, e a descrição entre
// os dois. Mas Banco do Brasil e Bradesco quebram a descrição em linhas
// separadas (um rótulo antes e/ou um detalhe depois da linha com
// data+valor), já que o texto do PDF é extraído por linha visual, não por
// registro lógico. Por isso, quando a linha com data+valor não traz uma
// descrição com texto real (só número de documento/lote), buscamos a
// descrição nas linhas vizinhas sem data/valor. O bankKey só existe para
// permitir ajustes futuros específicos de um banco, se necessário.
func ParseStatement(bankKey string, text string) []ParsedTransaction {
	_ = bankKey
	year := fallbackYear(text)

	lines := strings.Split(text, "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}

	var results []ParsedTransaction
	var pendingDescription string

	for i, line := range lines {
		if line == "" {
			continue
		}

		raw, ok := parseLine(line, year)
		if !ok {
			pendingDescription = line
			continue
		}

		description := raw.description
		if !hasMeaningfulText(description) {
			description = pendingDescription
			if cont := continuationLine(lines, i, year); cont != "" {
				description = joinDescription(description, cont)
			}
		}
		pendingDescription = ""

		description = strings.Trim(description, "-–—:; ")
		if description == "" || isBalanceLine(description) {
			continue
		}

		tx := ParsedTransaction{
			Date:        raw.date,
			Description: description,
			Amount:      raw.amount,
			Type:        classify(description, raw.sign),
			RawLine:     line,
		}
		if m := pixNameRe.FindStringSubmatch(description); m != nil {
			if name := strings.TrimSpace(m[1]); name != "" {
				tx.PixBeneficiary = &name
			}
		}
		results = append(results, tx)
	}
	return results
}

// hasMeaningfulText reporta se s contém pelo menos uma letra, distinguindo
// uma descrição real de sobras como número de documento/lote ("13601
// 511058923").
func hasMeaningfulText(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// continuationLine retorna a linha seguinte a lines[i] quando ela parece ser
// um complemento de descrição (não é uma nova linha de lançamento, nem
// vazia) — usado pelo layout do Banco do Brasil, onde o detalhe da operação
// vem depois da linha com data+valor.
func continuationLine(lines []string, i int, year int) string {
	if i+1 >= len(lines) {
		return ""
	}
	next := lines[i+1]
	if next == "" {
		return ""
	}
	if _, ok := parseLine(next, year); ok {
		return ""
	}
	return next
}

func joinDescription(a, b string) string {
	a = strings.TrimSpace(a)
	b = strings.TrimSpace(b)
	switch {
	case a == "":
		return b
	case b == "":
		return a
	default:
		return a + " - " + b
	}
}

// isBalanceLine identifica linhas de saldo (não são lançamentos reais):
// "Saldo Anterior", "Saldo do dia" ou "SALDO" no Banco do Brasil, e
// "COD. LANC. 0" no Bradesco (marca o saldo de abertura entre páginas).
func isBalanceLine(description string) bool {
	lower := strings.ToLower(strings.TrimSpace(description))
	return strings.HasPrefix(lower, "saldo") || strings.HasPrefix(lower, "cod. lanc. 0")
}

var (
	dateNumericRe = regexp.MustCompile(`^(\d{2})/(\d{2})(?:/(\d{2,4}))?\s+(.*)$`)
	dateMonthRe   = regexp.MustCompile(`(?i)^(\d{1,2})\s+(jan|fev|mar|abr|mai|jun|jul|ago|set|out|nov|dez)[a-zç]*\.?\s+(\d{4})?\s*(.*)$`)
	amountRe      = regexp.MustCompile(`(?i)(-)?\s*(?:R\$\s*)?(\d{1,3}(?:\.\d{3})*,\d{2})(-)?\s*(DB|D|CR|C|\(-\)|\(\+\))?`)
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
		"rem:", // Bradesco: "REM:" identifica o remetente de um PIX recebido
	}
	expenseHints = []string{
		"pago", "pagamento de", "compra", "debito", "débito", "enviado", "saque",
		"boleto", "fatura", "tarifa", "anuidade", "encargo", "juros",
		"transferencia enviada", "transferência enviada", "ted enviada", "doc enviado",
		"des:", // Bradesco: "DES:" identifica o destinatário de um PIX enviado
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

// rawLine é o resultado de reconhecer data + valor em uma linha, antes de
// resolver a descrição final (que pode vir de linhas vizinhas) e classificar
// o tipo do lançamento.
type rawLine struct {
	date        string
	description string
	amount      float64
	sign        int
}

func parseLine(line string, defaultYear int) (rawLine, bool) {
	date, rest, ok := extractDate(line, defaultYear)
	if !ok {
		return rawLine{}, false
	}

	amount, sign, rest, ok := extractAmount(rest)
	if !ok || amount == 0 {
		return rawLine{}, false
	}

	description := strings.Trim(strings.TrimSpace(rest), "-–—:; ")
	return rawLine{date: date, description: description, amount: amount, sign: sign}, true
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
	case leadingSign == "-" || trailingSign == "-" || suffix == "D" || suffix == "DB" || suffix == "(-)":
		sign = -1
	case suffix == "C" || suffix == "CR" || suffix == "(+)":
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
