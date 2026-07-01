package statement

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// buildTestPDF monta um PDF mínimo e válido com uma página de texto simples,
// para exercitar a extração real via github.com/ledongthuc/pdf (e não apenas
// a lógica de parsing sobre uma string já extraída).
func buildTestPDF(lines []string) []byte {
	var content strings.Builder
	content.WriteString("BT /F1 12 Tf\n")
	y := 750
	for _, line := range lines {
		escaped := strings.NewReplacer("(", `\(`, ")", `\)`).Replace(line)
		fmt.Fprintf(&content, "1 0 0 1 50 %d Tm (%s) Tj\n", y, escaped)
		y -= 20
	}
	content.WriteString("ET")
	stream := content.String()

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /Resources << /Font << /F1 4 0 R >> >> /MediaBox [0 0 612 792] /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}

	xrefStart := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objects)+1)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefStart)

	return buf.Bytes()
}

// buildPerGlyphTestPDF monta um PDF onde cada caractere é emitido em uma
// operação Tj própria, com Tm explícito por glifo — reproduzindo o layout de
// extratos (ex.: Sicredi) em que o gerador não agrupa o texto em runs, e que
// antes fazia a extração juntar tudo com espaço, quebrando cada caractere.
func buildPerGlyphTestPDF(row string, colStart int) []byte {
	var content strings.Builder
	content.WriteString("BT /F1 12 Tf\n")
	x := 50
	for i, ch := range row {
		gap := 15
		if i == colStart {
			gap = 150 // salto de coluna, ex.: fim da data para início da descrição
		}
		if i > 0 {
			x += gap
		}
		escaped := strings.NewReplacer("(", `\(`, ")", `\)`).Replace(string(ch))
		fmt.Fprintf(&content, "1 0 0 1 %d 750 Tm (%s) Tj\n", x, escaped)
	}
	content.WriteString("ET")
	stream := content.String()

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /Resources << /Font << /F1 4 0 R >> >> /MediaBox [0 0 612 792] /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(stream), stream),
	}

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := make([]int, len(objects)+1)
	for i, obj := range objects {
		offsets[i+1] = buf.Len()
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}

	xrefStart := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(objects)+1)
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i <= len(objects); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF", len(objects)+1, xrefStart)

	return buf.Bytes()
}

func TestExtractJoinsPerGlyphFragmentsWithoutExtraSpaces(t *testing.T) {
	row := "01/12/2025PAGAMENTO PIX - JOAO SILVA - R$ 149,99"
	colStart := len("01/12/2025")
	pdfBytes := buildPerGlyphTestPDF(row, colStart)

	text, err := ExtractText(pdfBytes)
	if err != nil {
		t.Fatalf("ExtractText falhou: %v", err)
	}
	if !strings.Contains(text, "01/12/2025 PAGAMENTO PIX - JOAO SILVA - R$ 149,99") {
		t.Fatalf("texto extraído não reconstituiu a linha esperada: %q", text)
	}

	parsed := ParseStatement("generic", text)
	if len(parsed) != 1 {
		t.Fatalf("esperava 1 lançamento, obteve %d: %+v", len(parsed), parsed)
	}
	if parsed[0].Date != "2025-12-01" || parsed[0].Amount != 149.99 || parsed[0].Type != "expense" {
		t.Errorf("lançamento inesperado: %+v", parsed[0])
	}
}

func TestExtractAndParseRealPDF(t *testing.T) {
	pdfBytes := buildTestPDF([]string{
		"01/03/2024 PIX RECEBIDO - JOAO SILVA 500,00 C",
		"02/03/2024 COMPRA CARTAO - MERCADO SAO PAULO 89,90 D",
	})

	text, err := ExtractText(pdfBytes)
	if err != nil {
		t.Fatalf("ExtractText falhou: %v", err)
	}
	if !strings.Contains(text, "JOAO SILVA") || !strings.Contains(text, "500,00") {
		t.Fatalf("texto extraído não contém o esperado: %q", text)
	}

	parsed := ParseStatement("generic", text)
	if len(parsed) != 2 {
		t.Fatalf("esperava 2 lançamentos a partir do PDF real, obteve %d: %+v", len(parsed), parsed)
	}
	if parsed[0].Type != "income" || parsed[0].Amount != 500 {
		t.Errorf("lançamento 0 inesperado: %+v", parsed[0])
	}
	if parsed[1].Type != "expense" || parsed[1].Amount != 89.90 {
		t.Errorf("lançamento 1 inesperado: %+v", parsed[1])
	}
}
