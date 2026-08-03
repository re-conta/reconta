package advisor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Record é a última análise salva para um usuário/mês/ano.
type Record struct {
	Stars       int
	Status      string // "ready" ou "error"
	Error       string
	Items       []Recommendation
	GeneratedAt time.Time
}

func (r *Repository) Get(ctx context.Context, userID int64, month, year int) (*Record, error) {
	var rec Record
	var status, errMsg, content sql.NullString
	var createdAt string

	err := r.db.QueryRowContext(ctx, `
		SELECT stars, status, error, content, created_at
		FROM advisor_recommendations WHERE user_id = ? AND month = ? AND year = ?`,
		userID, month, year,
	).Scan(&rec.Stars, &status, &errMsg, &content, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lendo recomendações: %w", err)
	}

	rec.Status = status.String
	rec.Error = errMsg.String
	rec.GeneratedAt = parseTimestamp(createdAt)
	if content.Valid && content.String != "" {
		if err := json.Unmarshal([]byte(content.String), &rec.Items); err != nil {
			return nil, fmt.Errorf("interpretando recomendações salvas: %w", err)
		}
	}
	return &rec, nil
}

// Save grava uma análise concluída com sucesso, substituindo qualquer
// resultado anterior do mesmo usuário/mês/ano.
func (r *Repository) Save(ctx context.Context, userID int64, month, year, stars int, items []Recommendation) error {
	raw, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("serializando recomendações: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO advisor_recommendations (user_id, month, year, stars, status, error, content, created_at)
		VALUES (?, ?, ?, ?, 'ready', NULL, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		ON CONFLICT (user_id, month, year) DO UPDATE SET
			stars = excluded.stars, status = 'ready', error = NULL, content = excluded.content, created_at = excluded.created_at`,
		userID, month, year, stars, string(raw),
	)
	if err != nil {
		return fmt.Errorf("salvando recomendações: %w", err)
	}
	return nil
}

// SaveError registra que a última tentativa de análise falhou, para que o
// handler saiba tentar de novo (com um pequeno backoff) na próxima consulta.
func (r *Repository) SaveError(ctx context.Context, userID int64, month, year, stars int, message string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO advisor_recommendations (user_id, month, year, stars, status, error, content, created_at)
		VALUES (?, ?, ?, ?, 'error', ?, '[]', strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		ON CONFLICT (user_id, month, year) DO UPDATE SET
			stars = excluded.stars, status = 'error', error = excluded.error, created_at = excluded.created_at`,
		userID, month, year, stars, message,
	)
	if err != nil {
		return fmt.Errorf("salvando erro de recomendações: %w", err)
	}
	return nil
}

// LogAPICall registra, no banco, o instante de uma chamada real à API do
// Groq. Persistir isso (em vez de manter só em memória) garante que os
// limites por hora/dia/mês sobrevivam a reinícios do processo.
func (r *Repository) LogAPICall(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO advisor_api_calls (created_at) VALUES (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))`)
	if err != nil {
		return fmt.Errorf("registrando chamada ao groq: %w", err)
	}
	return nil
}

func (r *Repository) LastCallAt(ctx context.Context) (time.Time, error) {
	var raw sql.NullString
	err := r.db.QueryRowContext(ctx, `SELECT created_at FROM advisor_api_calls ORDER BY id DESC LIMIT 1`).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) || !raw.Valid {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("lendo última chamada ao groq: %w", err)
	}
	return parseTimestamp(raw.String), nil
}

// CountCallsSince conta quantas chamadas ao Groq ocorreram desde o instante
// informado e retorna a mais antiga dentro dessa janela — usada para
// calcular quanto tempo falta até liberar uma nova chamada.
func (r *Repository) CountCallsSince(ctx context.Context, since time.Time) (count int, oldest time.Time, err error) {
	sinceStr := since.UTC().Format("2006-01-02T15:04:05.000Z")

	if err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM advisor_api_calls WHERE created_at >= ?`, sinceStr).Scan(&count); err != nil {
		return 0, time.Time{}, fmt.Errorf("contando chamadas ao groq: %w", err)
	}
	if count == 0 {
		return 0, time.Time{}, nil
	}

	var oldestRaw string
	if err = r.db.QueryRowContext(ctx, `SELECT MIN(created_at) FROM advisor_api_calls WHERE created_at >= ?`, sinceStr).Scan(&oldestRaw); err != nil {
		return 0, time.Time{}, fmt.Errorf("lendo chamada mais antiga: %w", err)
	}
	return count, parseTimestamp(oldestRaw), nil
}

func parseTimestamp(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}
