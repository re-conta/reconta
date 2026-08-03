package report

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"

	"github.com/go-pdf/fpdf"

	"github.com/re-conta/reconta/api/internal/transaction"
)

//go:embed assets/logo.png
var logoPNG []byte

// Paleta de cores da marca (ver web/src/styles/main.css).
var (
	colorBrand600 = [3]int{242, 117, 31}  // laranja — destaques, cabeçalho
	colorBrand50  = [3]int{255, 250, 235} // laranja bem claro — fundo de faixas
	colorCoral600 = [3]int{214, 49, 99}   // rosa — despesas
	colorInk900   = [3]int{28, 23, 18}    // quase preto — títulos
	colorInk600   = [3]int{92, 80, 68}    // texto secundário
	colorInk400   = [3]int{163, 147, 127} // texto auxiliar
	colorInk200   = [3]int{230, 222, 213} // linhas/bordas
	colorInk100   = [3]int{244, 240, 236} // fundo listrado
	colorWhite    = [3]int{255, 255, 255}
)

const logoImageName = "reconta-logo"

// BuildPDF gera um relatório em PDF: capa com logo e totais, imagens dos
// gráficos e uma tabela paginada com os lançamentos do período.
func BuildPDF(scope Scope, txs []transaction.Transaction, totals transaction.Totals, charts []ChartImage, accountNames map[int64]string) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 26, 12)
	pdf.SetAutoPageBreak(true, 18)
	pdf.RegisterImageOptionsReader(logoImageName, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(logoPNG))
	pdf.AliasNbPages("")

	pdf.SetHeaderFunc(func() {
		pdf.SetFillColor(colorBrand600[0], colorBrand600[1], colorBrand600[2])
		pdf.Rect(0, 0, 210, 16, "F")
		pdf.ImageOptions(logoImageName, 12, 3, 10, 10, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		pdf.SetTextColor(colorWhite[0], colorWhite[1], colorWhite[2])
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetXY(26, 3)
		pdf.CellFormat(0, 10, "Reconta", "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetXY(0, 3)
		pdf.CellFormat(198, 10, sanitize(scope.Label), "", 0, "R", false, 0, "")
		pdf.SetTextColor(colorInk900[0], colorInk900[1], colorInk900[2])
		pdf.SetXY(12, 22)
	})
	pdf.SetFooterFunc(func() {
		pdf.SetY(-13)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(colorInk400[0], colorInk400[1], colorInk400[2])
		pdf.CellFormat(0, 8, fmt.Sprintf("Pagina %d de {nb}", pdf.PageNo()), "", 0, "C", false, 0, "")
	})

	pdf.AddPage()

	pdf.SetTextColor(colorInk900[0], colorInk900[1], colorInk900[2])
	pdf.SetFont("Helvetica", "B", 20)
	pdf.CellFormat(0, 12, "Relatorio de gastos", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.SetTextColor(colorInk600[0], colorInk600[1], colorInk600[2])
	pdf.CellFormat(0, 7, sanitize(scope.Label), "", 1, "L", false, 0, "")
	pdf.Ln(6)

	drawSummary(pdf, totals)
	pdf.Ln(4)

	for i, chart := range charts {
		data, err := base64.StdEncoding.DecodeString(chart.PNGBase64)
		if err != nil {
			continue
		}
		name := fmt.Sprintf("chart-%d", i)
		pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(data))

		imgInfo := pdf.GetImageInfo(name)
		width := 186.0
		height := width * 0.5
		if imgInfo != nil && imgInfo.Width() > 0 {
			height = width * imgInfo.Height() / imgInfo.Width()
		}
		if pdf.GetY()+height+12 > 270 {
			pdf.AddPage()
		}

		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(colorInk900[0], colorInk900[1], colorInk900[2])
		pdf.CellFormat(0, 8, sanitize(chart.Title), "", 1, "L", false, 0, "")
		pdf.ImageOptions(name, pdf.GetX(), pdf.GetY(), width, height, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		pdf.Ln(height + 8)
	}

	pdf.AddPage()
	drawTransactionsTable(pdf, txs, accountNames)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("gerando pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func drawSummary(pdf *fpdf.Fpdf, totals transaction.Totals) {
	cards := []struct {
		label string
		value string
		color [3]int
	}{
		{"Receitas", formatCurrency(totals.Income), [3]int{34, 139, 87}},
		{"Despesas", formatCurrency(totals.Expense), colorCoral600},
		{"Saldo", formatCurrency(totals.Balance), colorInk900},
		{"Lancamentos", fmt.Sprintf("%d", totals.Count), colorInk900},
	}

	cardWidth := 44.0
	gap := 3.0
	x0, y0 := pdf.GetX(), pdf.GetY()
	height := 22.0

	for i, card := range cards {
		x := x0 + float64(i)*(cardWidth+gap)
		pdf.SetXY(x, y0)
		pdf.SetFillColor(colorBrand50[0], colorBrand50[1], colorBrand50[2])
		pdf.SetDrawColor(colorInk200[0], colorInk200[1], colorInk200[2])
		pdf.RoundedRect(x, y0, cardWidth, height, 2.5, "1234", "FD")

		pdf.SetXY(x+3, y0+3.5)
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetTextColor(colorInk600[0], colorInk600[1], colorInk600[2])
		pdf.CellFormat(cardWidth-6, 5, card.label, "", 2, "L", false, 0, "")

		pdf.SetX(x + 3)
		pdf.SetFont("Helvetica", "B", 11)
		pdf.SetTextColor(card.color[0], card.color[1], card.color[2])
		pdf.CellFormat(cardWidth-6, 7, sanitize(card.value), "", 2, "L", false, 0, "")
	}

	pdf.SetXY(x0, y0+height+2)
	pdf.SetTextColor(colorInk900[0], colorInk900[1], colorInk900[2])
}

func drawTransactionsTable(pdf *fpdf.Fpdf, txs []transaction.Transaction, accountNames map[int64]string) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(colorInk900[0], colorInk900[1], colorInk900[2])
	pdf.CellFormat(0, 10, "Lancamentos", "", 1, "L", false, 0, "")

	colWidths := []float64{20, 55, 30, 25, 20, 26}
	headers := []string{"Data", "Descricao", "Categoria", "Conta", "Tipo", "Valor"}

	drawTableHeader := func() {
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetFillColor(colorInk900[0], colorInk900[1], colorInk900[2])
		pdf.SetTextColor(colorWhite[0], colorWhite[1], colorWhite[2])
		for i, h := range headers {
			align := "L"
			if h == "Valor" {
				align = "R"
			}
			pdf.CellFormat(colWidths[i], 8, sanitize(h), "", 0, align, true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetTextColor(colorInk900[0], colorInk900[1], colorInk900[2])
	}

	drawTableHeader()

	pdf.SetFont("Helvetica", "", 8)
	for i, tx := range txs {
		if pdf.GetY()+7 > 279 {
			pdf.AddPage()
			pdf.SetFont("Helvetica", "B", 14)
			pdf.CellFormat(0, 10, "Lancamentos (continuacao)", "", 1, "L", false, 0, "")
			drawTableHeader()
			pdf.SetFont("Helvetica", "", 8)
		}

		category := ""
		if tx.CategoryName != nil {
			category = *tx.CategoryName
		}
		typeLabel := "Despesa"
		amount := -tx.Amount
		valueColor := colorCoral600
		if tx.Type == "income" {
			typeLabel = "Receita"
			amount = tx.Amount
			valueColor = [3]int{34, 139, 87}
		}

		if i%2 == 1 {
			pdf.SetFillColor(colorInk100[0], colorInk100[1], colorInk100[2])
		} else {
			pdf.SetFillColor(colorWhite[0], colorWhite[1], colorWhite[2])
		}

		pdf.SetTextColor(colorInk900[0], colorInk900[1], colorInk900[2])
		pdf.CellFormat(colWidths[0], 6.5, tx.Date, "", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[1], 6.5, sanitize(truncate(tx.Description, 34)), "", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[2], 6.5, sanitize(truncate(category, 18)), "", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[3], 6.5, sanitize(truncate(accountName(tx.AccountID, accountNames), 15)), "", 0, "L", true, 0, "")
		pdf.CellFormat(colWidths[4], 6.5, typeLabel, "", 0, "L", true, 0, "")
		pdf.SetTextColor(valueColor[0], valueColor[1], valueColor[2])
		pdf.CellFormat(colWidths[5], 6.5, formatCurrency(amount), "", 0, "R", true, 0, "")
		pdf.Ln(-1)
	}
}

func formatCurrency(v float64) string {
	return fmt.Sprintf("R$ %.2f", v)
}

// truncate corta a string em no maximo max runes, preservando o texto original
// (sem introduzir caracteres multibyte) — a codificacao para exibicao no PDF
// e responsabilidade de sanitize, que deve ser aplicada depois.
func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "."
}

// sanitize converte uma string UTF-8 para Latin-1/CP1252 byte a byte — é a
// codificação (WinAnsiEncoding) que as fontes padrão do fpdf (Helvetica etc.)
// esperam, e cobre todos os caracteres acentuados do pt-BR. Deve ser a última
// transformação aplicada antes de enviar o texto ao fpdf: aplicá-la antes de
// truncate reintroduziria bytes UTF-8 multibyte (ex.: "…") no meio da string
// já convertida, corrompendo o resultado.
func sanitize(s string) string {
	b := make([]byte, 0, len(s))
	for _, r := range s {
		if r <= 0xFF {
			b = append(b, byte(r))
		} else {
			b = append(b, '?')
		}
	}
	return string(b)
}
