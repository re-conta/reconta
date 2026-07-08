package report

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/xuri/excelize/v2"

	"github.com/re-conta/reconta/api/internal/transaction"
)

// BuildXLSX gera uma planilha XLSX com os lançamentos, o resumo de totais e
// os gráficos (imagens PNG) recebidos do frontend.
func BuildXLSX(scope Scope, txs []transaction.Transaction, totals transaction.Totals, charts []ChartImage, accountNames map[int64]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	const sheetTx = "Lançamentos"
	f.SetSheetName("Sheet1", sheetTx)

	headers := []string{"Data", "Descrição", "Categoria", "Conta", "Tipo", "Valor", "Tags", "Observações"}
	for col, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(col+1, 1)
		f.SetCellValue(sheetTx, cell, h)
	}

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
		f.SetCellValue(sheetTx, fmt.Sprintf("F%d", row), amount)
		f.SetCellValue(sheetTx, fmt.Sprintf("G%d", row), tagNames)
		f.SetCellValue(sheetTx, fmt.Sprintf("H%d", row), notes)
	}
	f.SetColWidth(sheetTx, "A", "A", 12)
	f.SetColWidth(sheetTx, "B", "B", 32)
	f.SetColWidth(sheetTx, "C", "D", 18)
	f.SetColWidth(sheetTx, "F", "F", 14)
	f.SetColWidth(sheetTx, "G", "H", 24)

	const sheetSummary = "Resumo"
	f.NewSheet(sheetSummary)
	f.SetCellValue(sheetSummary, "A1", "Relatório de gastos")
	f.SetCellValue(sheetSummary, "A2", "Período")
	f.SetCellValue(sheetSummary, "B2", scope.Label)
	f.SetCellValue(sheetSummary, "A3", "Receitas")
	f.SetCellValue(sheetSummary, "B3", totals.Income)
	f.SetCellValue(sheetSummary, "A4", "Despesas")
	f.SetCellValue(sheetSummary, "B4", totals.Expense)
	f.SetCellValue(sheetSummary, "A5", "Saldo")
	f.SetCellValue(sheetSummary, "B5", totals.Balance)
	f.SetCellValue(sheetSummary, "A6", "Lançamentos")
	f.SetCellValue(sheetSummary, "B6", totals.Count)
	f.SetColWidth(sheetSummary, "A", "A", 16)
	f.SetColWidth(sheetSummary, "B", "B", 20)

	row := 8
	for _, chart := range charts {
		data, err := base64.StdEncoding.DecodeString(chart.PNGBase64)
		if err != nil {
			continue
		}
		f.SetCellValue(sheetSummary, fmt.Sprintf("A%d", row), chart.Title)
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

	f.SetActiveSheet(0)

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, fmt.Errorf("gerando xlsx: %w", err)
	}
	return buf.Bytes(), nil
}

// accountName resolve o nome de uma conta pelo id, retornando string vazia quando ausente.
func accountName(accountID *int64, names map[int64]string) string {
	if accountID == nil {
		return ""
	}
	return names[*accountID]
}
