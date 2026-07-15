package user

import "testing"

func TestIsValidCNPJ(t *testing.T) {
	valid := []string{
		"11.222.333/0001-81",
		"11222333000181",
		"06.990.590/0001-23", // Google Brasil
		"33.000.167/0001-01", // Petrobras
	}
	for _, c := range valid {
		if !IsValidCNPJ(c) {
			t.Errorf("IsValidCNPJ(%q) = false, esperado true", c)
		}
	}

	invalid := []string{
		"",
		"123",
		"11.222.333/0001-82",
		"00.000.000/0000-00",
		"11111111111111",
		"1122233300018",   // 13 dígitos
		"112223330001811", // 15 dígitos
		"abc.def.ghi/jklm-no",
	}
	for _, c := range invalid {
		if IsValidCNPJ(c) {
			t.Errorf("IsValidCNPJ(%q) = true, esperado false", c)
		}
	}
}

func TestNormalizeCNPJ(t *testing.T) {
	if got := NormalizeCNPJ("11.222.333/0001-81"); got != "11222333000181" {
		t.Errorf("NormalizeCNPJ = %q, esperado 11222333000181", got)
	}
}
