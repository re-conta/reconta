package poupador

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/re-conta/reconta/api/internal/db"
)

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	conn, err := db.Open(filepath.Join(t.TempDir(), "reconta.db"))
	if err != nil {
		t.Fatalf("abrindo banco de teste: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	mux := http.NewServeMux()
	NewHandler(NewRepository(conn)).RegisterRoutes(mux)
	return mux
}

func TestSnapshotCanBeCreatedAndRetrieved(t *testing.T) {
	handler := newTestHandler(t)
	body := []byte(`{"incomes":[{"id":"income-1","name":"Salário","amount":5000,"frequency":"monthly","month":1}],"expenses":[{"id":"expense-1","name":"Aluguel","amount":1500,"frequency":"monthly","month":1}],"fuel":{"fuelType":"gasoline","fuelPrice":6.2,"distance":30,"distancePeriod":"daily","consumption":12}}`)
	create := httptest.NewRequest(http.MethodPost, "/api/poupador/snapshots", bytes.NewReader(body))
	create.Header.Set("Content-Type", "application/json")
	created := httptest.NewRecorder()
	handler.ServeHTTP(created, create)
	if created.Code != http.StatusCreated {
		t.Fatalf("status de criação = %d, corpo: %s", created.Code, created.Body.String())
	}

	var saved Snapshot
	if err := json.NewDecoder(created.Body).Decode(&saved); err != nil {
		t.Fatalf("decodificando criação: %v", err)
	}
	if saved.ID == "" || len(saved.Incomes) != 1 || len(saved.Expenses) != 1 || saved.Fuel == nil {
		t.Fatalf("snapshot salvo inesperado: %#v", saved)
	}

	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, "/api/poupador/snapshots/"+saved.ID, nil))
	if get.Code != http.StatusOK {
		t.Fatalf("status de leitura = %d, corpo: %s", get.Code, get.Body.String())
	}
	var loaded Snapshot
	if err := json.NewDecoder(get.Body).Decode(&loaded); err != nil {
		t.Fatalf("decodificando leitura: %v", err)
	}
	if loaded.Incomes[0].Name != "Salário" || loaded.Expenses[0].Amount != 1500 || loaded.Fuel.Consumption != 12 {
		t.Fatalf("snapshot recuperado diferente: %#v", loaded)
	}
}

func TestSnapshotRejectsInvalidEntryAndUnknownID(t *testing.T) {
	handler := newTestHandler(t)
	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, "/api/poupador/snapshots", bytes.NewBufferString(`{"incomes":[{"id":"a","name":"","amount":0,"frequency":"daily","month":13}],"expenses":[]}`)))
	if invalid.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status inválido = %d, corpo: %s", invalid.Code, invalid.Body.String())
	}

	missing := httptest.NewRecorder()
	handler.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/api/poupador/snapshots/00000000-0000-0000-0000-000000000000", nil))
	if missing.Code != http.StatusNotFound {
		t.Fatalf("status ausente = %d, corpo: %s", missing.Code, missing.Body.String())
	}
}
