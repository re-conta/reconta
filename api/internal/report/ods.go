package report

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"html"

	"github.com/re-conta/reconta/api/internal/transaction"
)

const odsManifest = `<?xml version="1.0" encoding="UTF-8"?>
<manifest:manifest xmlns:manifest="urn:oasis:names:tc:opendocument:xmlns:manifest:1.0" manifest:version="1.2">
 <manifest:file-entry manifest:full-path="/" manifest:version="1.2" manifest:media-type="application/vnd.oasis.opendocument.spreadsheet"/>
 <manifest:file-entry manifest:full-path="content.xml" manifest:media-type="text/xml"/>
%s</manifest:manifest>
`

// odsStyles define os estilos usados no content.xml: título, cabeçalho de
// tabela (negrito, centralizado, fundo escuro, texto branco), células de
// moeda (alinhadas à direita) e o bloco de cartões do resumo.
const odsStyles = `
  <style:style style:name="colWide" style:family="table-column">
   <style:table-column-properties style:column-width="4.2cm"/>
  </style:style>
  <style:style style:name="colMed" style:family="table-column">
   <style:table-column-properties style:column-width="2.6cm"/>
  </style:style>
  <style:style style:name="colNarrow" style:family="table-column">
   <style:table-column-properties style:column-width="1.8cm"/>
  </style:style>
  <style:style style:name="title" style:family="table-cell">
   <style:text-properties fo:font-weight="bold" fo:font-size="16pt" fo:color="#1c1712"/>
  </style:style>
  <style:style style:name="label" style:family="table-cell">
   <style:text-properties fo:color="#5c5044"/>
  </style:style>
  <style:style style:name="header" style:family="table-cell">
   <style:table-cell-properties fo:background-color="#1c1712" style:text-align-source="fix" style:vertical-align="middle"/>
   <style:paragraph-properties fo:text-align="center"/>
   <style:text-properties fo:font-weight="bold" fo:color="#ffffff"/>
  </style:style>
  <style:style style:name="cell" style:family="table-cell">
   <style:table-cell-properties style:vertical-align="middle"/>
  </style:style>
  <style:style style:name="cellStripe" style:family="table-cell">
   <style:table-cell-properties fo:background-color="#f4f0ec" style:vertical-align="middle"/>
  </style:style>
  <style:style style:name="currency" style:family="table-cell" style:data-style-name="currencyFmt">
   <style:table-cell-properties style:vertical-align="middle"/>
   <style:paragraph-properties fo:text-align="end"/>
  </style:style>
  <style:style style:name="currencyStripe" style:family="table-cell" style:data-style-name="currencyFmt">
   <style:table-cell-properties fo:background-color="#f4f0ec" style:vertical-align="middle"/>
   <style:paragraph-properties fo:text-align="end"/>
  </style:style>
  <style:style style:name="cardLabel" style:family="table-cell">
   <style:table-cell-properties fo:background-color="#fffaeb" fo:border="0.5pt solid #e6ded5"/>
   <style:paragraph-properties fo:text-align="center"/>
   <style:text-properties fo:font-size="9pt" fo:color="#5c5044"/>
  </style:style>
  <style:style style:name="cardValue" style:family="table-cell" style:data-style-name="currencyFmt">
   <style:table-cell-properties fo:background-color="#fffaeb" fo:border="0.5pt solid #e6ded5"/>
   <style:paragraph-properties fo:text-align="center"/>
   <style:text-properties fo:font-weight="bold" fo:font-size="13pt" fo:color="#1c1712"/>
  </style:style>
  <style:style style:name="cardValueIncome" style:family="table-cell" style:data-style-name="currencyFmt">
   <style:table-cell-properties fo:background-color="#fffaeb" fo:border="0.5pt solid #e6ded5"/>
   <style:paragraph-properties fo:text-align="center"/>
   <style:text-properties fo:font-weight="bold" fo:font-size="13pt" fo:color="#228b57"/>
  </style:style>
  <style:style style:name="cardValueExpense" style:family="table-cell" style:data-style-name="currencyFmt">
   <style:table-cell-properties fo:background-color="#fffaeb" fo:border="0.5pt solid #e6ded5"/>
   <style:paragraph-properties fo:text-align="center"/>
   <style:text-properties fo:font-weight="bold" fo:font-size="13pt" fo:color="#d63163"/>
  </style:style>
  <style:style style:name="cardValueCount" style:family="table-cell">
   <style:table-cell-properties fo:background-color="#fffaeb" fo:border="0.5pt solid #e6ded5"/>
   <style:paragraph-properties fo:text-align="center"/>
   <style:text-properties fo:font-weight="bold" fo:font-size="13pt" fo:color="#1c1712"/>
  </style:style>
`

const odsNumberStyles = `
  <number:currency-style style:name="currencyFmt">
   <number:text>R$&#160;</number:text>
   <number:number number:decimal-places="2" number:min-integer-digits="1" number:grouping="true"/>
  </number:currency-style>
`

// BuildODS gera uma planilha OpenDocument (.ods) manualmente: um zip com
// mimetype, manifest.xml, content.xml (uma tabela de lançamentos + resumo) e
// as imagens dos gráficos embutidas em Pictures/.
func BuildODS(scope Scope, txs []transaction.Transaction, totals transaction.Totals, charts []ChartImage, accountNames map[int64]string) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	mimeWriter, err := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	if err != nil {
		return nil, fmt.Errorf("criando mimetype do ods: %w", err)
	}
	if _, err := mimeWriter.Write([]byte("application/vnd.oasis.opendocument.spreadsheet")); err != nil {
		return nil, err
	}

	manifestEntries := ""
	var images []odsImage
	for i, chart := range charts {
		data, err := base64.StdEncoding.DecodeString(chart.PNGBase64)
		if err != nil {
			continue
		}
		path := fmt.Sprintf("Pictures/chart%d.png", i+1)
		w, err := zw.Create(path)
		if err != nil {
			return nil, fmt.Errorf("escrevendo imagem do gráfico: %w", err)
		}
		if _, err := w.Write(data); err != nil {
			return nil, err
		}
		manifestEntries += fmt.Sprintf(" <manifest:file-entry manifest:full-path=\"%s\" manifest:media-type=\"image/png\"/>\n", path)
		images = append(images, odsImage{title: chart.Title, path: path})
	}

	content := buildODSContent(scope, txs, totals, images, accountNames)
	cw, err := zw.Create("content.xml")
	if err != nil {
		return nil, fmt.Errorf("criando content.xml do ods: %w", err)
	}
	if _, err := cw.Write([]byte(content)); err != nil {
		return nil, err
	}

	mw, err := zw.Create("META-INF/manifest.xml")
	if err != nil {
		return nil, fmt.Errorf("criando manifest.xml do ods: %w", err)
	}
	if _, err := mw.Write([]byte(fmt.Sprintf(odsManifest, manifestEntries))); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("fechando ods: %w", err)
	}
	return buf.Bytes(), nil
}

type odsImage struct {
	title string
	path  string
}

func buildODSContent(scope Scope, txs []transaction.Transaction, totals transaction.Totals, images []odsImage, accountNames map[int64]string) string {
	var b bytes.Buffer

	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-content
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
  xmlns:style="urn:oasis:names:tc:opendocument:xmlns:style:1.0"
  xmlns:fo="urn:oasis:names:tc:opendocument:xmlns:xsl-fo-compatible:1.0"
  xmlns:number="urn:oasis:names:tc:opendocument:xmlns:datastyle:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:svg="urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0"
  xmlns:xlink="http://www.w3.org/1999/xlink"
  office:version="1.2">
 <office:automatic-styles>`)
	b.WriteString(odsNumberStyles)
	b.WriteString(odsStyles)
	b.WriteString(`
 </office:automatic-styles>
 <office:body>
  <office:spreadsheet>
   <table:table table:name="Relatório">
    <table:table-column table:style-name="colNarrow"/>
    <table:table-column table:style-name="colWide"/>
    <table:table-column table:style-name="colMed"/>
    <table:table-column table:style-name="colMed"/>
    <table:table-column table:style-name="colNarrow"/>
    <table:table-column table:style-name="colMed"/>
    <table:table-column table:style-name="colWide"/>
    <table:table-column table:style-name="colMed"/>
`)

	writeODSRow(&b, styledCellStr("title", "Relatório de gastos — "+scope.Label))
	writeODSRow(&b)
	// Valor é a última coluna, seguindo a leitura natural da linha (dados
	// descritivos primeiro, o número que fecha o registro por último).
	writeODSRow(&b,
		styledCellStr("header", "Data"), styledCellStr("header", "Descrição"), styledCellStr("header", "Categoria"),
		styledCellStr("header", "Conta"), styledCellStr("header", "Tipo"), styledCellStr("header", "Tags"),
		styledCellStr("header", "Observações"), styledCellStr("header", "Valor"),
	)

	for i, tx := range txs {
		cellName, currencyName := "cell", "currency"
		if i%2 == 1 {
			cellName, currencyName = "cellStripe", "currencyStripe"
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
		writeODSRow(&b,
			styledCellStr(cellName, tx.Date), styledCellStr(cellName, tx.Description), styledCellStr(cellName, category),
			styledCellStr(cellName, accountName(tx.AccountID, accountNames)), styledCellStr(cellName, typeLabel),
			styledCellStr(cellName, tagNames), styledCellStr(cellName, notes),
			styledCellFloat(currencyName, amount),
		)
	}

	writeODSRow(&b)
	writeODSRow(&b, styledCellStr("cardLabel", "Receitas"), styledCellStr("cardLabel", "Despesas"), styledCellStr("cardLabel", "Saldo"), styledCellStr("cardLabel", "Lançamentos"))
	writeODSRow(&b,
		styledCellFloat("cardValueIncome", totals.Income),
		styledCellFloat("cardValueExpense", totals.Expense),
		styledCellFloat("cardValue", totals.Balance),
		odsCell{style: "cardValueCount", valueType: "float", text: fmt.Sprintf("%d", totals.Count)},
	)

	for _, img := range images {
		writeODSRow(&b)
		writeODSRow(&b, styledCellStr("label", img.title))
		fmt.Fprintf(&b, `    <table:table-row>
     <table:table-cell>
      <draw:frame draw:name="%s" svg:width="12cm" svg:height="7cm">
       <draw:image xlink:href="%s" xlink:type="simple" xlink:show="embed" xlink:actuate="onLoad"/>
      </draw:frame>
     </table:table-cell>
    </table:table-row>
`, html.EscapeString(img.title), img.path)
	}

	b.WriteString(`   </table:table>
  </office:spreadsheet>
 </office:body>
</office:document-content>
`)

	return b.String()
}

type odsCell struct {
	style     string
	valueType string
	text      string
}

func styledCellStr(style, v string) odsCell { return odsCell{style: style, valueType: "string", text: v} }
func styledCellFloat(style string, v float64) odsCell {
	return odsCell{style: style, valueType: "float", text: fmt.Sprintf("%.2f", v)}
}

func writeODSRow(b *bytes.Buffer, cells ...odsCell) {
	b.WriteString("    <table:table-row>\n")
	for _, c := range cells {
		styleAttr := ""
		if c.style != "" {
			styleAttr = fmt.Sprintf(` table:style-name="%s"`, c.style)
		}
		if c.valueType == "float" {
			fmt.Fprintf(b, "     <table:table-cell%s office:value-type=\"float\" office:value=\"%s\"><text:p>%s</text:p></table:table-cell>\n", styleAttr, c.text, html.EscapeString(c.text))
		} else {
			fmt.Fprintf(b, "     <table:table-cell%s office:value-type=\"string\"><text:p>%s</text:p></table:table-cell>\n", styleAttr, html.EscapeString(c.text))
		}
	}
	b.WriteString("    </table:table-row>\n")
}
