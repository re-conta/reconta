package statement

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
)

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
			parts := make([]string, len(row.Content))
			for j, text := range row.Content {
				parts[j] = text.S
			}
			buf.WriteString(strings.Join(parts, " "))
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
