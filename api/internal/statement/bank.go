package statement

import "strings"

// Bank identifica um banco/instituição suportado na detecção de extratos.
type Bank struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

// SupportedBanks lista os bancos com detecção dedicada. "generic" é o
// fallback usado quando nenhum dos anteriores é identificado no texto.
var SupportedBanks = []Bank{
	{"bb", "Banco do Brasil"},
	{"sicredi", "Sicredi"},
	{"nubank", "Nubank"},
	{"mercadopago", "Mercado Pago"},
	{"itau", "Itaú"},
	{"generic", "Outro / genérico"},
}

// DetectBank tenta identificar o banco a partir do texto extraído do PDF.
func DetectBank(text string) Bank {
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "banco do brasil") || strings.Contains(lower, "bb.com.br"):
		return BankByKey("bb")
	case strings.Contains(lower, "sicredi"):
		return BankByKey("sicredi")
	case strings.Contains(lower, "nubank") || strings.Contains(lower, "nu pagamentos"):
		return BankByKey("nubank")
	case strings.Contains(lower, "mercado pago") || strings.Contains(lower, "mercadopago"):
		return BankByKey("mercadopago")
	case strings.Contains(lower, "itaú") || strings.Contains(lower, "itau unibanco") || strings.Contains(lower, "banco itau"):
		return BankByKey("itau")
	default:
		return BankByKey("generic")
	}
}

// BankByKey retorna o banco correspondente à chave, ou o genérico se não existir.
func BankByKey(key string) Bank {
	for _, b := range SupportedBanks {
		if b.Key == key {
			return b
		}
	}
	return SupportedBanks[len(SupportedBanks)-1]
}
