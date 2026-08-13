// Package poupador armazena snapshots públicos e compartilháveis do
// planejador financeiro Poupador.
package poupador

import "time"

const (
	KindIncome  = "income"
	KindExpense = "expense"
)

// Entry é uma receita ou um gasto exatamente como informado no planejador.
type Entry struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Amount    float64 `json:"amount"`
	Frequency string  `json:"frequency"`
	Month     int     `json:"month"`
}

// Fuel contém os parâmetros da calculadora de consumo de combustível.
type Fuel struct {
	FuelType       string  `json:"fuelType"`
	FuelPrice      float64 `json:"fuelPrice"`
	Distance       float64 `json:"distance"`
	DistancePeriod string  `json:"distancePeriod"`
	Consumption    float64 `json:"consumption"`
}

// Snapshot contém todos os parâmetros necessários para reconstruir um
// resultado do Poupador em outro dispositivo.
type Snapshot struct {
	ID        string    `json:"id"`
	Incomes   []Entry   `json:"incomes"`
	Expenses  []Entry   `json:"expenses"`
	Fuel      *Fuel     `json:"fuel,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
}
