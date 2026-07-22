package analytics

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// Visit é uma visita de página capturada pelo beacon do front-end.
type Visit struct {
	VisitorID      string
	SessionID      string
	UserID         *int64
	Path           string
	Referrer       string
	IP             string
	Country        string
	Region         string
	City           string
	Latitude       *float64
	Longitude      *float64
	UserAgent      string
	Browser        string
	BrowserVersion string
	OS             string
	DeviceType     string
	IsBot          bool
}

func (r *Repository) InsertVisit(ctx context.Context, v Visit) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO page_visits (
			visitor_id, session_id, user_id, path, referrer, ip,
			country, region, city, latitude, longitude,
			user_agent, browser, browser_version, os, device_type, is_bot
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		v.VisitorID, v.SessionID, v.UserID, v.Path, v.Referrer, v.IP,
		v.Country, v.Region, v.City, v.Latitude, v.Longitude,
		v.UserAgent, v.Browser, v.BrowserVersion, v.OS, v.DeviceType, boolToInt(v.IsBot),
	)
	if err != nil {
		return fmt.Errorf("registrando visita: %w", err)
	}
	return nil
}

type DayCount struct {
	Date           string `json:"date"`
	Visits         int    `json:"visits"`
	UniqueVisitors int    `json:"uniqueVisitors"`
}

type Overview struct {
	TotalVisits       int        `json:"totalVisits"`
	UniqueVisitors    int        `json:"uniqueVisitors"`
	NewVisitors       int        `json:"newVisitors"`
	ReturningVisitors int        `json:"returningVisitors"`
	ByDay             []DayCount `json:"byDay"`
}

// Overview resume o período: totais, únicos, novos vs recorrentes (um
// visitante é "novo" se sua primeira visita registrada cai dentro do
// período) e a série diária de visitas/visitantes únicos.
func (r *Repository) Overview(ctx context.Context, from, to time.Time) (Overview, error) {
	var o Overview

	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(DISTINCT visitor_id)
		FROM page_visits
		WHERE created_at >= ? AND created_at < ? AND is_bot = 0`,
		fmtTime(from), fmtTime(to),
	).Scan(&o.TotalVisits, &o.UniqueVisitors)
	if err != nil {
		return o, fmt.Errorf("calculando totais do período: %w", err)
	}

	err = r.db.QueryRowContext(ctx, `
		WITH period_visitors AS (
			SELECT DISTINCT visitor_id FROM page_visits
			WHERE created_at >= ? AND created_at < ? AND is_bot = 0
		),
		first_seen AS (
			SELECT visitor_id, MIN(created_at) AS first_at FROM page_visits
			WHERE is_bot = 0 GROUP BY visitor_id
		)
		SELECT
			COUNT(*) FILTER (WHERE fs.first_at >= ?),
			COUNT(*) FILTER (WHERE fs.first_at < ?)
		FROM period_visitors pv
		JOIN first_seen fs ON fs.visitor_id = pv.visitor_id`,
		fmtTime(from), fmtTime(to), fmtTime(from), fmtTime(from),
	).Scan(&o.NewVisitors, &o.ReturningVisitors)
	if err != nil {
		return o, fmt.Errorf("calculando novos vs recorrentes: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT substr(created_at, 1, 10) AS day, COUNT(*), COUNT(DISTINCT visitor_id)
		FROM page_visits
		WHERE created_at >= ? AND created_at < ? AND is_bot = 0
		GROUP BY day
		ORDER BY day`,
		fmtTime(from), fmtTime(to),
	)
	if err != nil {
		return o, fmt.Errorf("calculando série diária: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d DayCount
		if err := rows.Scan(&d.Date, &d.Visits, &d.UniqueVisitors); err != nil {
			return o, fmt.Errorf("lendo série diária: %w", err)
		}
		o.ByDay = append(o.ByDay, d)
	}
	if err := rows.Err(); err != nil {
		return o, fmt.Errorf("lendo série diária: %w", err)
	}

	return o, nil
}

type PathCount struct {
	Path           string `json:"path"`
	Visits         int    `json:"visits"`
	UniqueVisitors int    `json:"uniqueVisitors"`
}

func (r *Repository) TopPages(ctx context.Context, from, to time.Time, limit int) ([]PathCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT path, COUNT(*), COUNT(DISTINCT visitor_id)
		FROM page_visits
		WHERE created_at >= ? AND created_at < ? AND is_bot = 0
		GROUP BY path
		ORDER BY COUNT(*) DESC
		LIMIT ?`,
		fmtTime(from), fmtTime(to), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("calculando páginas mais visitadas: %w", err)
	}
	defer rows.Close()

	var out []PathCount
	for rows.Next() {
		var p PathCount
		if err := rows.Scan(&p.Path, &p.Visits, &p.UniqueVisitors); err != nil {
			return nil, fmt.Errorf("lendo páginas mais visitadas: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

type ReferrerCount struct {
	Referrer string `json:"referrer"`
	Visits   int    `json:"visits"`
}

func (r *Repository) TopReferrers(ctx context.Context, from, to time.Time, limit int) ([]ReferrerCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT CASE WHEN referrer = '' THEN '(direto)' ELSE referrer END AS ref, COUNT(*)
		FROM page_visits
		WHERE created_at >= ? AND created_at < ? AND is_bot = 0
		GROUP BY ref
		ORDER BY COUNT(*) DESC
		LIMIT ?`,
		fmtTime(from), fmtTime(to), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("calculando referrers: %w", err)
	}
	defer rows.Close()

	var out []ReferrerCount
	for rows.Next() {
		var rc ReferrerCount
		if err := rows.Scan(&rc.Referrer, &rc.Visits); err != nil {
			return nil, fmt.Errorf("lendo referrers: %w", err)
		}
		out = append(out, rc)
	}
	return out, rows.Err()
}

type LocationCount struct {
	Country string `json:"country"`
	City    string `json:"city"`
	Visits  int    `json:"visits"`
}

func (r *Repository) TopLocations(ctx context.Context, from, to time.Time, limit int) ([]LocationCount, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			CASE WHEN country = '' THEN 'Desconhecido' ELSE country END,
			city, COUNT(*)
		FROM page_visits
		WHERE created_at >= ? AND created_at < ? AND is_bot = 0
		GROUP BY country, city
		ORDER BY COUNT(*) DESC
		LIMIT ?`,
		fmtTime(from), fmtTime(to), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("calculando localizações: %w", err)
	}
	defer rows.Close()

	var out []LocationCount
	for rows.Next() {
		var l LocationCount
		if err := rows.Scan(&l.Country, &l.City, &l.Visits); err != nil {
			return nil, fmt.Errorf("lendo localizações: %w", err)
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type NamedCount struct {
	Name   string `json:"name"`
	Visits int    `json:"visits"`
}

type DeviceBreakdown struct {
	Browsers []NamedCount `json:"browsers"`
	OS       []NamedCount `json:"os"`
	Devices  []NamedCount `json:"devices"`
}

func (r *Repository) DeviceBreakdown(ctx context.Context, from, to time.Time) (DeviceBreakdown, error) {
	var d DeviceBreakdown

	browsers, err := r.namedCounts(ctx, "browser", from, to)
	if err != nil {
		return d, err
	}
	d.Browsers = browsers

	oses, err := r.namedCounts(ctx, "os", from, to)
	if err != nil {
		return d, err
	}
	d.OS = oses

	devices, err := r.namedCounts(ctx, "device_type", from, to)
	if err != nil {
		return d, err
	}
	d.Devices = devices

	return d, nil
}

// namedCounts agrupa por uma coluna fixa do próprio pacote (nunca por input
// do usuário), então a interpolação do nome da coluna é segura.
func (r *Repository) namedCounts(ctx context.Context, column string, from, to time.Time) ([]NamedCount, error) {
	rows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT CASE WHEN %s = '' THEN 'Desconhecido' ELSE %s END AS name, COUNT(*)
		FROM page_visits
		WHERE created_at >= ? AND created_at < ? AND is_bot = 0
		GROUP BY name
		ORDER BY COUNT(*) DESC
		LIMIT 10`, column, column),
		fmtTime(from), fmtTime(to),
	)
	if err != nil {
		return nil, fmt.Errorf("calculando %s: %w", column, err)
	}
	defer rows.Close()

	var out []NamedCount
	for rows.Next() {
		var n NamedCount
		if err := rows.Scan(&n.Name, &n.Visits); err != nil {
			return nil, fmt.Errorf("lendo %s: %w", column, err)
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

type RecentVisit struct {
	ID        int64  `json:"id"`
	Path      string `json:"path"`
	Referrer  string `json:"referrer"`
	IP        string `json:"ip"`
	Country   string `json:"country"`
	City      string `json:"city"`
	Browser   string `json:"browser"`
	OS        string `json:"os"`
	Device    string `json:"device"`
	CreatedAt string `json:"createdAt"`
}

func (r *Repository) RecentVisits(ctx context.Context, from, to time.Time, limit int) ([]RecentVisit, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, path, referrer, ip, country, city, browser, os, device_type, created_at
		FROM page_visits
		WHERE created_at >= ? AND created_at < ? AND is_bot = 0
		ORDER BY created_at DESC
		LIMIT ?`,
		fmtTime(from), fmtTime(to), limit,
	)
	if err != nil {
		return nil, fmt.Errorf("carregando visitas recentes: %w", err)
	}
	defer rows.Close()

	var out []RecentVisit
	for rows.Next() {
		var v RecentVisit
		if err := rows.Scan(&v.ID, &v.Path, &v.Referrer, &v.IP, &v.Country, &v.City, &v.Browser, &v.OS, &v.Device, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("lendo visitas recentes: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ActiveNow conta visitantes únicos com atividade nos últimos 5 minutos.
func (r *Repository) ActiveNow(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT visitor_id) FROM page_visits
		WHERE created_at >= ? AND is_bot = 0`,
		fmtTime(time.Now().Add(-5*time.Minute)),
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("calculando visitantes ativos: %w", err)
	}
	return count, nil
}

func fmtTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
