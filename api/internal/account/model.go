package account

import "time"

// Account representa uma conta bancária/carteira do usuário.
type Account struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"createdAt"`
}
