// Package textutil traz utilitários de comparação de texto usados em mais de
// um domínio do app (ex.: casar o nome do titular cadastrado com o nome de
// beneficiário/remetente extraído de um extrato bancário).
package textutil

import "strings"

var accentReplacements = map[rune]rune{
	'Á': 'A', 'À': 'A', 'Â': 'A', 'Ã': 'A', 'Ä': 'A',
	'É': 'E', 'È': 'E', 'Ê': 'E', 'Ë': 'E',
	'Í': 'I', 'Ì': 'I', 'Î': 'I', 'Ï': 'I',
	'Ó': 'O', 'Ò': 'O', 'Ô': 'O', 'Õ': 'O', 'Ö': 'O',
	'Ú': 'U', 'Ù': 'U', 'Û': 'U', 'Ü': 'U',
	'Ç': 'C',
}

// NormalizeName remove acentos, colapsa espaços e uniformiza a caixa de um
// nome de pessoa, para permitir comparar grafias vindas de fontes diferentes
// (ex.: "José da Silva" em um extrato vs. "Jose  da  Silva" no cadastro).
func NormalizeName(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))

	var b strings.Builder
	lastSpace := false
	for _, r := range s {
		if repl, ok := accentReplacements[r]; ok {
			r = repl
		}
		if r == ' ' {
			if lastSpace {
				continue
			}
			lastSpace = true
		} else {
			lastSpace = false
		}
		b.WriteRune(r)
	}
	return b.String()
}

// SameHolder reporta se dois nomes provavelmente se referem à mesma pessoa,
// ignorando acentuação, caixa e espaçamento.
func SameHolder(a, b string) bool {
	a, b = NormalizeName(a), NormalizeName(b)
	return a != "" && a == b
}
