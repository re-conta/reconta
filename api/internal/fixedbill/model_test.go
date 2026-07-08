package fixedbill

import (
	"testing"
	"time"
)

func TestNextDueDate(t *testing.T) {
	base := time.Date(2026, time.January, 31, 0, 0, 0, 0, time.UTC)

	cases := []struct {
		periodicity string
		want        time.Time
	}{
		{PeriodicityWeekly, time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC)},
		{PeriodicityBiweekly, time.Date(2026, time.February, 14, 0, 0, 0, 0, time.UTC)},
		{PeriodicityMonthly, time.Date(2026, time.March, 3, 0, 0, 0, 0, time.UTC)}, // AddDate normaliza 31/fev
		{PeriodicityBimonthly, time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)},
		{PeriodicityQuarterly, time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)}, // 31 abril não existe (abril tem 30 dias)
		{PeriodicitySemiannual, time.Date(2026, time.July, 31, 0, 0, 0, 0, time.UTC)},
		{PeriodicityAnnual, time.Date(2027, time.January, 31, 0, 0, 0, 0, time.UTC)},
		{PeriodicityBiennial, time.Date(2028, time.January, 31, 0, 0, 0, 0, time.UTC)},
	}

	for _, c := range cases {
		got := NextDueDate(base, c.periodicity)
		if !got.Equal(c.want) {
			t.Errorf("NextDueDate(%s, %s) = %s, want %s", base.Format("2006-01-02"), c.periodicity, got.Format("2006-01-02"), c.want.Format("2006-01-02"))
		}
	}
}

func TestIsValidPeriodicity(t *testing.T) {
	if !IsValidPeriodicity(PeriodicityMonthly) {
		t.Error("monthly deveria ser válido")
	}
	if IsValidPeriodicity("daily") {
		t.Error("daily não deveria ser válido")
	}
}
