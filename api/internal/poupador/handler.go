package poupador

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

const maxEntriesPerKind = 100

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/poupador/snapshots", h.create)
	mux.HandleFunc("GET /api/poupador/snapshots/{id}", h.get)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var snapshot Snapshot
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&snapshot); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := validateSnapshot(snapshot); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	snapshot.ID = uuid.NewString()
	created, err := h.repo.Create(r.Context(), snapshot)
	if err != nil {
		log.Printf("erro ao salvar resultado do poupador: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeError(w, http.StatusNotFound, "resultado não encontrado")
		return
	}

	snapshot, err := h.repo.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "resultado não encontrado")
		return
	}
	if err != nil {
		log.Printf("erro ao buscar resultado do poupador: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}

func validateSnapshot(snapshot Snapshot) error {
	if len(snapshot.Incomes) > maxEntriesPerKind || len(snapshot.Expenses) > maxEntriesPerKind {
		return errors.New("o limite é de 100 receitas e 100 gastos")
	}
	for _, entry := range append(snapshot.Incomes, snapshot.Expenses...) {
		if strings.TrimSpace(entry.ID) == "" || strings.TrimSpace(entry.Name) == "" || len([]rune(entry.Name)) > 60 {
			return errors.New("cada entrada deve ter nome e identificador válidos")
		}
		if math.IsNaN(entry.Amount) || math.IsInf(entry.Amount, 0) || entry.Amount <= 0 {
			return errors.New("o valor de cada entrada deve ser maior que zero")
		}
		if entry.Month < 1 || entry.Month > 12 {
			return errors.New("o mês de cada entrada deve estar entre 1 e 12")
		}
		switch entry.Frequency {
		case "monthly", "weekly", "biweekly", "yearly", "one-time":
		default:
			return errors.New("a frequência de cada entrada é inválida")
		}
	}
	if err := validateFuel(snapshot.Fuel); err != nil {
		return err
	}
	return nil
}

func validateFuel(fuel *Fuel) error {
	if fuel == nil {
		return nil
	}

	values := []float64{fuel.FuelPrice, fuel.Distance, fuel.Consumption}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
			return errors.New("os valores de combustível devem ser maiores ou iguais a zero")
		}
	}

	switch fuel.FuelType {
	case "gasoline", "diesel":
	default:
		return errors.New("o tipo de combustível é inválido")
	}

	switch fuel.DistancePeriod {
	case "daily", "monthly":
	default:
		return errors.New("o período da distância é inválido")
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("erro ao codificar resposta JSON: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
