package report

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/go-pdf/fpdf"

	"github.com/re-conta/reconta/api/internal/transaction"
)

// BuildPDF gera um relatório em PDF: capa com totais, imagens dos gráficos e
// uma tabela paginada com os lançamentos do período.
func BuildPDF(scope Scope, txs []transaction.Transaction, totals transaction.Totals, charts []ChartImage, accountNames map[int64]string) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(12, 15, 12)
	pdf.SetAutoPageBreak(true, 15)
	pdf.AddPage()

	pdf.SetFont("Helvetica", "B", 18)
	pdf.CellFormat(0, 10, "Relatorio de gastos", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(0, 7, sanitize(scope.Label), "", 1, "L", false, 0, "")
	pdf.Ln(4)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Resumo", "", 1, "L", false, 0, "")
	pdf.SetFont("Helvetica", "", 11)
	pdf.CellFormat(60, 7, "Receitas", "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, formatCurrency(totals.Income), "1", 1, "R", false, 0, "")
	pdf.CellFormat(60, 7, "Despesas", "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, formatCurrency(totals.Expense), "1", 1, "R", false, 0, "")
	pdf.CellFormat(60, 7, "Saldo", "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, formatCurrency(totals.Balance), "1", 1, "R", false, 0, "")
	pdf.CellFormat(60, 7, "Lancamentos", "1", 0, "L", false, 0, "")
	pdf.CellFormat(60, 7, fmt.Sprintf("%d", totals.Count), "1", 1, "R", false, 0, "")
	pdf.Ln(6)

	for i, chart := range charts {
		data, err := base64.StdEncoding.DecodeString(chart.PNGBase64)
		if err != nil {
			continue
		}
		name := fmt.Sprintf("chart-%d", i)
		pdf.RegisterImageOptionsReader(name, fpdf.ImageOptions{ImageType: "PNG"}, bytes.NewReader(data))

		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(0, 7, sanitize(chart.Title), "", 1, "L", false, 0, "")

		imgInfo := pdf.GetImageInfo(name)
		width := 180.0
		height := width
		if imgInfo != nil && imgInfo.Width() > 0 {
			height = width * imgInfo.Height() / imgInfo.Width()
		}
		if pdf.GetY()+height > 270 {
			pdf.AddPage()
		}
		pdf.ImageOptions(name, pdf.GetX(), pdf.GetY(), width, height, false, fpdf.ImageOptions{ImageType: "PNG"}, 0, "")
		pdf.Ln(height + 6)
	}

	pdf.AddPage()
	pdf.SetFont("Helvetica", "B", 12)
	pdf.CellFormat(0, 8, "Lancamentos", "", 1, "L", false, 0, "")

	colWidths := []float64{20, 55, 30, 25, 20, 26}
	headers := []string{"Data", "Descricao", "Categoria", "Conta", "Tipo", "Valor"}
	pdf.SetFont("Helvetica", "B", 9)
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 7, h, "1", 0, "L", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Helvetica", "", 8)
	for _, tx := range txs {
		if pdf.GetY()+7 > 280 {
			pdf.AddPage()
			pdf.SetFont("Helvetica", "B", 9)
			for i, h := range headers {
				pdf.CellFormat(colWidths[i], 7, h, "1", 0, "L", false, 0, "")
			}
			pdf.Ln(-1)
			pdf.SetFont("Helvetica", "", 8)
		}
		category := ""
		if tx.CategoryName != nil {
			category = *tx.CategoryName
		}
		typeLabel := "Despesa"
		amount := -tx.Amount
		if tx.Type == "income" {
			typeLabel = "Receita"
			amount = tx.Amount
		}
		pdf.CellFormat(colWidths[0], 6, tx.Date, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[1], 6, truncate(sanitize(tx.Description), 34), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[2], 6, truncate(sanitize(category), 18), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[3], 6, truncate(sanitize(accountName(tx.AccountID, accountNames)), 15), "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[4], 6, typeLabel, "1", 0, "L", false, 0, "")
		pdf.CellFormat(colWidths[5], 6, formatCurrency(amount), "1", 0, "R", false, 0, "")
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("gerando pdf: %w", err)
	}
	return buf.Bytes(), nil
}

func formatCurrency(v float64) string {
	return fmt.Sprintf("R$ %.2f", v)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

// sanitize converte uma string UTF-8 para Latin-1/CP1252 byte a byte — é a
// codificação (WinAnsiEncoding) que as fontes padrão do fpdf (Helvetica etc.)
// esperam, e cobre todos os caracteres acentuados do pt-BR.
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
