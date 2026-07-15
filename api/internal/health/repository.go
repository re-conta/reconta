package health

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Settings define os limites (em % da taxa de poupança: saldo/receitas) que
// separam cada nível de saúde financeira. Configuração global, editável no
// painel de admin — uma única linha (id = 1) na tabela.
type Settings struct {
	Enabled          bool    `json:"enabled"`
	ThresholdOtima   float64 `json:"thresholdOtima"`
	ThresholdBoa     float64 `json:"thresholdBoa"`
	ThresholdEstavel float64 `json:"thresholdEstavel"`
	ThresholdRuim    float64 `json:"thresholdRuim"`
}

// DefaultSettings são os limites usados enquanto o admin não personalizar:
// poupar ≥20% é ótimo, ≥10% bom, fechar no zero é estável e estourar até
// 10% das receitas é ruim; abaixo disso, péssimo.
func DefaultSettings() Settings {
	return Settings{
		Enabled:          true,
		ThresholdOtima:   20,
		ThresholdBoa:     10,
		ThresholdEstavel: 0,
		ThresholdRuim:    -10,
	}
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetSettings retorna a configuração global, ou os padrões se nunca foi salva.
func (r *Repository) GetSettings(ctx context.Context) (Settings, error) {
	var (
		s       Settings
		enabled int
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT enabled, threshold_otima, threshold_boa, threshold_estavel, threshold_ruim
		FROM financial_health_settings WHERE id = 1`,
	).Scan(&enabled, &s.ThresholdOtima, &s.ThresholdBoa, &s.ThresholdEstavel, &s.ThresholdRuim)
	if errors.Is(err, sql.ErrNoRows) {
		return DefaultSettings(), nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("lendo configuração de saúde financeira: %w", err)
	}
	s.Enabled = enabled == 1
	return s, nil
}

// SaveSettings grava (ou substitui) a configuração global.
func (r *Repository) SaveSettings(ctx context.Context, s Settings) error {
	enabled := 0
	if s.Enabled {
		enabled = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO financial_health_settings (id, enabled, threshold_otima, threshold_boa, threshold_estavel, threshold_ruim, updated_at)
		VALUES (1, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		ON CONFLICT (id) DO UPDATE SET
			enabled = excluded.enabled,
			threshold_otima = excluded.threshold_otima,
			threshold_boa = excluded.threshold_boa,
			threshold_estavel = excluded.threshold_estavel,
			threshold_ruim = excluded.threshold_ruim,
			updated_at = excluded.updated_at`,
		enabled, s.ThresholdOtima, s.ThresholdBoa, s.ThresholdEstavel, s.ThresholdRuim,
	)
	if err != nil {
		return fmt.Errorf("salvando configuração de saúde financeira: %w", err)
	}
	return nil
}

// MonthTotals soma receitas e despesas do usuário no mês informado.
func (r *Repository) MonthTotals(ctx context.Context, userID int64, month, year int) (income, expense float64, err error) {
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	// Dia zero do mês seguinte = último dia do mês corrente.
	end := time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	err = r.db.QueryRowContext(ctx, `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
		FROM transactions
		WHERE user_id = ? AND date >= ? AND date <= ?`,
		userID, start, end,
	).Scan(&income, &expense)
	if err != nil {
		return 0, 0, fmt.Errorf("calculando totais do mês: %w", err)
	}
	return income, expense, nil
}
