package poupador

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("resultado do poupador não encontrado")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, snapshot Snapshot) (*Snapshot, error) {
	data, err := json.Marshal(struct {
		Incomes  []Entry `json:"incomes"`
		Expenses []Entry `json:"expenses"`
		Fuel     *Fuel   `json:"fuel,omitempty"`
	}{
		Incomes:  snapshot.Incomes,
		Expenses: snapshot.Expenses,
		Fuel:     snapshot.Fuel,
	})
	if err != nil {
		return nil, fmt.Errorf("codificando resultado do poupador: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `INSERT INTO poupador_snapshots (id, data) VALUES (?, ?)`, snapshot.ID, data); err != nil {
		return nil, fmt.Errorf("salvando resultado do poupador: %w", err)
	}
	return r.Get(ctx, snapshot.ID)
}

func (r *Repository) Get(ctx context.Context, id string) (*Snapshot, error) {
	var (
		data      []byte
		createdAt string
	)
	if err := r.db.QueryRowContext(ctx, `SELECT data, created_at FROM poupador_snapshots WHERE id = ?`, id).Scan(&data, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("buscando resultado do poupador: %w", err)
	}

	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("decodificando resultado do poupador: %w", err)
	}
	created, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, fmt.Errorf("decodificando data do resultado do poupador: %w", err)
	}
	snapshot.ID = id
	snapshot.CreatedAt = created
	return &snapshot, nil
}
