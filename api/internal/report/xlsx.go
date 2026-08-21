package report

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/re-conta/reconta/api/internal/transaction"
)

const currencyFmt = `"R$" #,##0.00;[RED]-"R$" #,##0.00`

// BuildXLSX gera uma planilha XLSX com os lançamentos, o resumo de totais e
// os gráficos (imagens PNG) recebidos do frontend.
func BuildXLSX(scope Scope, txs []transaction.Transaction, totals transaction.Totals, charts []ChartImage, accountNames map[int64]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheetTx = "Lançamentos"
	f.SetSheetName("Sheet1", sheetTx)

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#1C1712"}},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	cellStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	amountStyle, err := f.NewStyle(&excelize.Style{
		CustomNumFmt: strPtr(currencyFmt),
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	stripeStyle, err := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F4F0EC"}},
		Alignment: &excelize.Alignment{Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}
	stripeAmountStyle, err := f.NewStyle(&excelize.Style{
		Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#F4F0EC"}},
		CustomNumFmt: strPtr(currencyFmt),
		Alignment:    &excelize.Alignment{Horizontal: "right", Vertical: "center"},
	})
	if err != nil {
		return nil, err
	}

	// Valor é a última coluna, para acompanhar a leitura natural da linha (dados
	// descritivos primeiro, o número que fecha o registro por último).
	headers := []string{"Data", "Descrição", "Categoria", "Conta", "Tipo", "Tags", "Observações", "Valor"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetTx, cell, h)
	}
	f.SetRowHeight(sheetTx, 1, 20)
	headerRange, _ := excelize.CoordinatesToCellName(1, 1)
	headerRangeEnd, _ := excelize.CoordinatesToCellName(len(headers), 1)
	f.SetCellStyle(sheetTx, headerRange, headerRangeEnd, headerStyle)

	for i, tx := range txs {
		row := i + 2
		category := ""
		if tx.CategoryName != nil {
			category = *tx.CategoryName
		}
		typeLabel := "Despesa"
		if tx.Type == "income" {
			typeLabel = "Receita"
		}
		amount := tx.Amount
		if tx.Type == "expense" {
			amount = -amount
		}
		tagNames := ""
		for i, t := range tx.Tags {
			if i > 0 {
				tagNames += ", "
			}
			tagNames += t.Name
		}
		notes := ""
		if tx.Notes != nil {
			notes = *tx.Notes
		}

		f.SetCellValue(sheetTx, fmt.Sprintf("A%d", row), tx.Date)
		f.SetCellValue(sheetTx, fmt.Sprintf("B%d", row), tx.Description)
		f.SetCellValue(sheetTx, fmt.Sprintf("C%d", row), category)
		f.SetCellValue(sheetTx, fmt.Sprintf("D%d", row), accountName(tx.AccountID, accountNames))
		f.SetCellValue(sheetTx, fmt.Sprintf("E%d", row), typeLabel)
		f.SetCellValue(sheetTx, fmt.Sprintf("F%d", row), tagNames)
		f.SetCellValue(sheetTx, fmt.Sprintf("G%d", row), notes)
		f.SetCellValue(sheetTx, fmt.Sprintf("H%d", row), amount)

		rowStyle, rowAmountStyle := cellStyle, amountStyle
		if i%2 == 1 {
			rowStyle, rowAmountStyle = stripeStyle, stripeAmountStyle
		}
		f.SetCellStyle(sheetTx, fmt.Sprintf("A%d", row), fmt.Sprintf("G%d", row), rowStyle)
		f.SetCellStyle(sheetTx, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), rowAmountStyle)
	}
	f.SetColWidth(sheetTx, "A", "A", 12)
	f.SetColWidth(sheetTx, "B", "B", 32)
	f.SetColWidth(sheetTx, "C", "D", 18)
	f.SetColWidth(sheetTx, "E", "E", 12)
	f.SetColWidth(sheetTx, "F", "G", 24)
	f.SetColWidth(sheetTx, "H", "H", 16)
	f.SetPanes(sheetTx, &excelize.Panes{Freeze: true, Split: false, XSplit: 0, YSplit: 1, TopLeftCell: "A2", ActivePane: "bottomLeft"})

	if err := buildSummarySheet(f, scope, totals, charts); err != nil {
		return nil, err
	}

	f.SetActiveSheet(0)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("gerando xlsx: %w", err)
	}
	return buf.Bytes(), nil
}

// buildSummarySheet monta a aba "Resumo": título, período, um bloco de
// totais com rótulo/valor lado a lado (estilizado, sem números soltos) e os
// gráficos recebidos do frontend abaixo.
func buildSummarySheet(f *excelize.File, scope Scope, totals transaction.Totals, charts []ChartImage) error {
	const sheetSummary = "Resumo"
	f.NewSheet(sheetSummary)

	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 16, Color: "#1C1712"},
	})
	if err != nil {
		return err
	}
	labelStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "#5C5044"},
	})
	if err != nil {
		return err
	}
	periodStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Color: "#1C1712"},
	})
	if err != nil {
		return err
	}
	cardLabelStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 10, Color: "#5C5044"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FFFAEB"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "top", Color: "#E6DED5", Style: 1},
			{Type: "left", Color: "#E6DED5", Style: 1},
			{Type: "right", Color: "#E6DED5", Style: 1},
		},
	})
	if err != nil {
		return err
	}
	cardValueStyle := func(color string) (int, error) {
		return f.NewStyle(&excelize.Style{
			Font:         &excelize.Font{Bold: true, Size: 14, Color: color},
			Fill:         excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FFFAEB"}},
			CustomNumFmt: strPtr(currencyFmt),
			Alignment:    &excelize.Alignment{Horizontal: "center"},
			Border: []excelize.Border{
				{Type: "bottom", Color: "#E6DED5", Style: 1},
				{Type: "left", Color: "#E6DED5", Style: 1},
				{Type: "right", Color: "#E6DED5", Style: 1},
			},
		})
	}
	incomeStyle, err := cardValueStyle("#228B57")
	if err != nil {
		return err
	}
	expenseStyle, err := cardValueStyle("#D63163")
	if err != nil {
		return err
	}
	balanceStyle, err := cardValueStyle("#1C1712")
	if err != nil {
		return err
	}
	countValueStyle, err := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 14, Color: "#1C1712"},
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FFFAEB"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "bottom", Color: "#E6DED5", Style: 1},
			{Type: "left", Color: "#E6DED5", Style: 1},
			{Type: "right", Color: "#E6DED5", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	f.SetCellValue(sheetSummary, "A1", "Relatório de gastos")
	f.SetCellStyle(sheetSummary, "A1", "A1", titleStyle)
	f.SetCellValue(sheetSummary, "A2", "Período")
	f.SetCellStyle(sheetSummary, "A2", "A2", labelStyle)
	f.SetCellValue(sheetSummary, "B2", scope.Label)
	f.SetCellStyle(sheetSummary, "B2", "B2", periodStyle)

	// Bloco de totais: uma linha de rótulos (row 4) sobre uma linha de valores
	// (row 5), como cartões — cada card ocupa uma coluna, lado a lado.
	cards := []struct {
		col   string
		label string
		style int
		value float64
	}{
		{"A", "Receitas", incomeStyle, totals.Income},
		{"B", "Despesas", expenseStyle, totals.Expense},
		{"C", "Saldo", balanceStyle, totals.Balance},
	}
	for _, card := range cards {
		f.SetCellValue(sheetSummary, card.col+"4", card.label)
		f.SetCellStyle(sheetSummary, card.col+"4", card.col+"4", cardLabelStyle)
		f.SetCellValue(sheetSummary, card.col+"5", card.value)
		f.SetCellStyle(sheetSummary, card.col+"5", card.col+"5", card.style)
	}
	f.SetCellValue(sheetSummary, "D4", "Lançamentos")
	f.SetCellStyle(sheetSummary, "D4", "D4", cardLabelStyle)
	f.SetCellValue(sheetSummary, "D5", totals.Count)
	f.SetCellStyle(sheetSummary, "D5", "D5", countValueStyle)

	f.SetRowHeight(sheetSummary, 1, 24)
	f.SetRowHeight(sheetSummary, 5, 22)
	f.SetColWidth(sheetSummary, "A", "D", 20)

	row := 8
	for _, chart := range charts {
		data, err := base64.StdEncoding.DecodeString(chart.PNGBase64)
		if err != nil {
			continue
		}
		f.SetCellValue(sheetSummary, fmt.Sprintf("A%d", row), chart.Title)
		f.SetCellStyle(sheetSummary, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), periodStyle)
		cell := fmt.Sprintf("A%d", row+1)
		if err := f.AddPictureFromBytes(sheetSummary, cell, &excelize.Picture{
			Extension: ".png",
			File:      data,
			Format:    &excelize.GraphicOptions{AutoFit: true},
		}); err != nil {
			continue
		}
		row += 18
	}

	return nil
}

// accountName resolve o nome de uma conta pelo id, retornando string vazia quando ausente.
func accountName(accountID *int64, names map[int64]string) string {
	if accountID == nil {
		return ""
	}
	return names[*accountID]
}
