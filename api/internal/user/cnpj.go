package user

import "strings"

// NormalizeCNPJ remove máscara/formatação, mantendo apenas os dígitos.
func NormalizeCNPJ(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// IsValidCNPJ valida um CNPJ (já normalizado ou não) pelo algoritmo oficial
// dos dígitos verificadores.
func IsValidCNPJ(s string) bool {
	cnpj := NormalizeCNPJ(s)
	if len(cnpj) != 14 {
		return false
	}

	// CNPJs com todos os dígitos iguais passam nos dígitos verificadores, mas
	// são inválidos (ex.: 00.000.000/0000-00).
	allEqual := true
	for i := 1; i < len(cnpj); i++ {
		if cnpj[i] != cnpj[0] {
			allEqual = false
			break
		}
	}
	if allEqual {
		return false
	}

	digit := func(weights []int) byte {
		sum := 0
		for i, w := range weights {
			sum += int(cnpj[i]-'0') * w
		}
		rest := sum % 11
		if rest < 2 {
			return '0'
		}
		return byte('0' + 11 - rest)
	}

	first := digit([]int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	second := digit([]int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return cnpj[12] == first && cnpj[13] == second
}
