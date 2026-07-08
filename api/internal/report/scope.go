package report

import (
	"fmt"
	"time"
)

var monthNames = [...]string{
	"Janeiro", "Fevereiro", "Março", "Abril", "Maio", "Junho",
	"Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro",
}

// ResolveScope traduz o tipo de período (month|year|range|all) e seus
// parâmetros em um intervalo de datas e um rótulo legível para o relatório.
func ResolveScope(kind string, month, year int, dateFrom, dateTo string) (Scope, error) {
	switch kind {
	case "month":
		if month < 1 || month > 12 || year <= 0 {
			return Scope{}, fmt.Errorf("mês e ano inválidos")
		}
		first := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		last := first.AddDate(0, 1, -1)
		return Scope{
			DateFrom: first.Format("2006-01-02"),
			DateTo:   last.Format("2006-01-02"),
			Label:    fmt.Sprintf("%s/%d", monthNames[month-1], year),
		}, nil
	case "year":
		if year <= 0 {
			return Scope{}, fmt.Errorf("ano inválido")
		}
		return Scope{
			DateFrom: fmt.Sprintf("%04d-01-01", year),
			DateTo:   fmt.Sprintf("%04d-12-31", year),
			Label:    fmt.Sprintf("%d", year),
		}, nil
	case "range":
		if dateFrom == "" || dateTo == "" {
			return Scope{}, fmt.Errorf("intervalo de datas inválido")
		}
		return Scope{
			DateFrom: dateFrom,
			DateTo:   dateTo,
			Label:    fmt.Sprintf("%s a %s", formatDate(dateFrom), formatDate(dateTo)),
		}, nil
	case "all":
		return Scope{Label: "Todo o período"}, nil
	default:
		return Scope{}, fmt.Errorf("escopo inválido: %s", kind)
	}
}

func formatDate(s string) string {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return s
	}
	return t.Format("02/01/2006")
}
