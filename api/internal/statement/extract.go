package statement

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
)

// columnGapMultiplier define, em múltiplos do tamanho da fonte do fragmento,
// o salto horizontal acima do qual dois fragmentos na mesma linha são
// considerados colunas diferentes de uma tabela (e não caracteres/palavras
// adjacentes). Um limite relativo à fonte é necessário porque documentos
// diferentes (e até seções diferentes do mesmo PDF) usam tamanhos de fonte
// distintos — um valor fixo em pontos calibrado para um extrato não
// generaliza para outro. Calibrado observando extratos Sicredi reais: o
// espaçamento normal entre caracteres fica abaixo de ~0.9x o tamanho da
// fonte, enquanto saltos de coluna passam de 2.7x.
const columnGapMultiplier = 1.5

// columnGapFloor é o piso absoluto (em pontos) usado quando o fragmento não
// informa um tamanho de fonte (FontSize == 0), evitando dividir por zero.
const columnGapFloor = 6.0

// rowYTolerance agrupa fragmentos de texto na mesma linha visual mesmo que
// suas coordenadas Y não sejam idênticas bit a bit (variações de arredondamento
// de ponto flutuante entre glifos de uma mesma linha).
const rowYTolerance = 2.0

// ExtractText lê o texto de um PDF a partir dos bytes do arquivo, agrupando
// os fragmentos de texto por linha (mesma posição vertical) para preservar o
// layout tabular típico de extratos bancários.
//
// Usamos page.Content() (que rastreia a matriz de texto completa, incluindo
// os operadores Td/TD/T*) em vez de page.GetTextByRow(): esta última só
// atualiza a posição Y no operador Tm, então extratos gerados com Td/T*
// (como as versões mais novas do extrato do Sicredi) fazem toda a página cair
// numa única "linha", e nenhuma data/valor é reconhecido.
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
		for _, row := range groupIntoRows(page.Content().Text) {
			buf.WriteString(joinRow(row))
			buf.WriteString("\n")
		}
	}

	if buf.Len() > 0 {
		return buf.String(), nil
	}

	// Sem texto posicionado (ex.: PDF sem operadores de texto reconhecidos) —
	// cai para a extração de texto puro como última tentativa.
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

// groupIntoRows agrupa fragmentos de texto pela coordenada Y (linha visual),
// tolerando pequenas variações de ponto flutuante, e ordena cada linha da
// esquerda para a direita pela coordenada X.
func groupIntoRows(texts []pdf.Text) [][]pdf.Text {
	sorted := make([]pdf.Text, len(texts))
	copy(sorted, texts)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].Y > sorted[j].Y
	})

	var rows [][]pdf.Text
	for _, text := range sorted {
		if n := len(rows); n > 0 {
			last := rows[n-1]
			if last[0].Y-text.Y <= rowYTolerance {
				rows[n-1] = append(last, text)
				continue
			}
		}
		rows = append(rows, []pdf.Text{text})
	}

	for _, row := range rows {
		sort.SliceStable(row, func(i, j int) bool { return row[i].X < row[j].X })
	}
	return rows
}

// joinRow concatena os fragmentos de uma linha, sintetizando um espaço apenas
// quando o salto horizontal indica um pulo de coluna da tabela (ver
// columnGapMultiplier) — necessário porque alguns geradores de PDF (ex.:
// extratos Sicredi) emitem cada glifo como um fragmento de texto isolado, já
// incluindo os espaços "reais" do texto como fragmentos próprios.
func joinRow(row []pdf.Text) string {
	var line strings.Builder
	var prevEndX float64
	for j, text := range row {
		if j > 0 {
			threshold := text.FontSize * columnGapMultiplier
			if threshold < columnGapFloor {
				threshold = columnGapFloor
			}
			if text.X-prevEndX > threshold {
				line.WriteString(" ")
			}
		}
		line.WriteString(text.S)
		prevEndX = text.X + text.W
	}
	return line.String()
}
