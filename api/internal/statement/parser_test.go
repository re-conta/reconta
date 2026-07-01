package statement

import "testing"

func TestParseStatement(t *testing.T) {
	text := `
EXTRATO DE 01/03/2024 A 31/03/2024
DATA       HISTORICO                          VALOR       SALDO
01/03/2024 PIX RECEBIDO - JOAO SILVA           500,00 C    1.500,00 C
02/03/2024 COMPRA CARTAO - MERCADO SAO PAULO   89,90 D     1.410,10 C
03/03/2024 PIX ENVIADO - MARIA OLIVEIRA        150,00 D    1.260,10 C
04/03/2024 TARIFA MANUTENCAO CONTA             30,00 D     1.230,10 C
01 MAR PIX ENVIADO Carlos Souza -R$ 75,00
02 MAR Pagamento recebido de Ana Paula R$ 200,00
`

	got := ParseStatement("generic", text)
	if len(got) != 6 {
		t.Fatalf("esperava 6 lançamentos, obteve %d: %+v", len(got), got)
	}

	if got[0].Date != "2024-03-01" || got[0].Type != "income" || got[0].Amount != 500 {
		t.Errorf("linha 0 inesperada: %+v", got[0])
	}
	if got[1].Type != "expense" || got[1].Amount != 89.90 {
		t.Errorf("linha 1 inesperada: %+v", got[1])
	}
	if got[2].Type != "expense" || got[2].PixBeneficiary == nil || *got[2].PixBeneficiary != "MARIA OLIVEIRA" {
		t.Errorf("linha 2 inesperada: %+v", got[2])
	}
	if got[3].Type != "expense" {
		t.Errorf("linha 3 (tarifa) deveria ser despesa: %+v", got[3])
	}
	if got[4].Date != "2024-03-01" || got[4].Type != "expense" || got[4].Amount != 75 {
		t.Errorf("linha 4 (nubank) inesperada: %+v", got[4])
	}
	if got[5].Type != "income" || got[5].Amount != 200 {
		t.Errorf("linha 5 (mercado pago) inesperada: %+v", got[5])
	}
}

func TestDetectBank(t *testing.T) {
	cases := map[string]string{
		"Banco do Brasil S.A. - Extrato de conta corrente": "bb",
		"SICREDI - Cooperativa de Crédito":                 "sicredi",
		"Nubank - Nu Pagamentos S.A.":                      "nubank",
		"Mercado Pago - Comprovante":                       "mercadopago",
		"Itaú Unibanco S.A.":                               "itau",
		"Alguma coisa desconhecida":                        "generic",
	}
	for text, want := range cases {
		if got := DetectBank(text).Key; got != want {
			t.Errorf("DetectBank(%q) = %q, want %q", text, got, want)
		}
	}
}
