package statement

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

// columnGapThreshold é o salto horizontal (em pontos) acima do qual dois
// fragmentos de texto na mesma linha são considerados colunas diferentes de
// uma tabela, e não caracteres/palavras adjacentes. Calibrado observando que
// o espaçamento normal entre caracteres fica abaixo de ~45pt, enquanto saltos
// de coluna passam de 130pt.
const columnGapThreshold = 60

// ExtractText lê o texto de um PDF a partir dos bytes do arquivo, agrupando
// os fragmentos de texto por linha (mesma posição vertical) para preservar o
// layout tabular típico de extratos bancários. Isso é mais confiável do que
// a extração ingênua de texto puro, que só quebra linha quando o PDF usa
// certos operadores de conteúdo (BT/T*) — nem todo gerador de PDF os usa.
func ExtractText(data []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("lendo pdf: %w", err)
	}

	var buf bytes.Buffer
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		rows, err := page.GetTextByRow()
		if err != nil {
			continue
		}
		for _, row := range rows {
			var line strings.Builder
			var prevEndX float64
			for j, text := range row.Content {
				// Alguns geradores de PDF (ex.: extratos Sicredi) emitem cada
				// glifo como um fragmento de texto isolado, já incluindo os
				// espaços "reais" do texto como fragmentos próprios — então
				// unir tudo com espaço (como antes) duplicava espaços e
				// quebrava cada caractere isoladamente. Só sintetizamos um
				// espaço quando o salto horizontal é grande demais para ser
				// espaçamento normal de caractere (indicando um pulo de
				// coluna da tabela, ex.: de "Data" para "Descrição").
				if j > 0 && text.X-prevEndX > columnGapThreshold {
					line.WriteString(" ")
				}
				line.WriteString(text.S)
				prevEndX = text.X + text.W
			}
			buf.WriteString(line.String())
			buf.WriteString("\n")
		}
	}

	if buf.Len() > 0 {
		return buf.String(), nil
	}

	// Sem linhas reconhecidas por posição (ex.: PDF sem operadores Tm) — cai
	// para a extração de texto puro como última tentativa.
	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extraindo texto do pdf: %w", err)
	}
	text, err := io.ReadAll(textReader)
	if err != nil {
		return "", fmt.Errorf("lendo texto extraído do pdf: %w", err)
	}
	return string(text), nil
}
