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
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:svg="urn:oasis:names:tc:opendocument:xmlns:svg-compatible:1.0"
  xmlns:xlink="http://www.w3.org/1999/xlink"
  office:version="1.2">
 <office:body>
  <office:spreadsheet>
   <table:table table:name="Relatório">
`)

	writeODSRow(&b, cellStr("Relatório de gastos — "+scope.Label))
	writeODSRow(&b)
	writeODSRow(&b, cellStr("Data"), cellStr("Descrição"), cellStr("Categoria"), cellStr("Conta"), cellStr("Tipo"), cellStr("Valor"), cellStr("Tags"), cellStr("Observações"))

	for _, tx := range txs {
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
			cellStr(tx.Date), cellStr(tx.Description), cellStr(category),
			cellStr(accountName(tx.AccountID, accountNames)), cellStr(typeLabel),
			cellFloat(amount), cellStr(tagNames), cellStr(notes),
		)
	}

	writeODSRow(&b)
	writeODSRow(&b, cellStr("Receitas"), cellFloat(totals.Income))
	writeODSRow(&b, cellStr("Despesas"), cellFloat(totals.Expense))
	writeODSRow(&b, cellStr("Saldo"), cellFloat(totals.Balance))
	writeODSRow(&b, cellStr("Lançamentos"), cellFloat(float64(totals.Count)))

	for _, img := range images {
		writeODSRow(&b)
		writeODSRow(&b, cellStr(img.title))
		b.WriteString(`    <table:table-row>
     <table:table-cell>
      <draw:frame draw:name="` + html.EscapeString(img.title) + `" svg:width="12cm" svg:height="7cm">
       <draw:image xlink:href="` + img.path + `" xlink:type="simple" xlink:show="embed" xlink:actuate="onLoad"/>
      </draw:frame>
     </table:table-cell>
    </table:table-row>
`)
	}

	b.WriteString(`   </table:table>
  </office:spreadsheet>
 </office:body>
</office:document-content>
`)

	return b.String()
}

type odsCell struct {
	valueType string
	text      string
}

func cellStr(v string) odsCell { return odsCell{valueType: "string", text: v} }
func cellFloat(v float64) odsCell {
	return odsCell{valueType: "float", text: fmt.Sprintf("%.2f", v)}
}

func writeODSRow(b *bytes.Buffer, cells ...odsCell) {
	b.WriteString("    <table:table-row>\n")
	for _, c := range cells {
		if c.valueType == "float" {
			b.WriteString(fmt.Sprintf(`     <table:table-cell office:value-type="float" office:value="%s"><text:p>%s</text:p></table:table-cell>`+"\n", c.text, html.EscapeString(c.text)))
		} else {
			b.WriteString(`     <table:table-cell office:value-type="string"><text:p>` + html.EscapeString(c.text) + `</text:p></table:table-cell>` + "\n")
		}
	}
	b.WriteString("    </table:table-row>\n")
}
