package fixedbill

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/re-conta/reconta/api/internal/auth"
)

type Handler struct {
	repo *Repository
	auth *auth.Handler
}

func NewHandler(repo *Repository, authHandler *auth.Handler) *Handler {
	return &Handler{repo: repo, auth: authHandler}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/fixed-bills", h.auth.RequireUser(h.list))
	mux.HandleFunc("POST /api/fixed-bills", h.auth.RequireUser(h.create))
	mux.HandleFunc("PUT /api/fixed-bills/{id}", h.auth.RequireUser(h.update))
	mux.HandleFunc("DELETE /api/fixed-bills/{id}", h.auth.RequireUser(h.delete))
	mux.HandleFunc("POST /api/fixed-bills/{id}/freeze", h.auth.RequireUser(h.freeze))
	mux.HandleFunc("POST /api/fixed-bills/{id}/reactivate", h.auth.RequireUser(h.reactivate))
	mux.HandleFunc("POST /api/fixed-bills/{id}/close", h.auth.RequireUser(h.close))
	mux.HandleFunc("POST /api/fixed-bills/{id}/pay", h.auth.RequireUser(h.pay))
	mux.HandleFunc("GET /api/fixed-bills/{id}/payments", h.auth.RequireUser(h.payments))
}

type fixedBillRequest struct {
	Name        string  `json:"name"`
	Amount      float64 `json:"amount"`
	CategoryID  *int64  `json:"categoryId"`
	AccountID   *int64  `json:"accountId"`
	Periodicity string  `json:"periodicity"`
	DueDate     string  `json:"dueDate"`
	Notes       *string `json:"notes"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request, userID int64) {
	bills, err := h.repo.List(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao listar contas fixas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, bills)
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, userID int64) {
	var req fixedBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := validateFixedBillRequest(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	b, err := h.repo.Create(r.Context(), userID, Input{
		Name: req.Name, Amount: req.Amount, CategoryID: req.CategoryID, AccountID: req.AccountID,
		Periodicity: req.Periodicity, DueDate: req.DueDate, Notes: req.Notes,
	})
	if err != nil {
		log.Printf("erro ao criar conta fixa: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req fixedBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := validateFixedBillRequest(&req); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	b, err := h.repo.Update(r.Context(), userID, id, Input{
		Name: req.Name, Amount: req.Amount, CategoryID: req.CategoryID, AccountID: req.AccountID,
		Periodicity: req.Periodicity, DueDate: req.DueDate, Notes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrada")
			return
		}
		log.Printf("erro ao atualizar conta fixa: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	if err := h.repo.Delete(r.Context(), userID, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrada")
			return
		}
		log.Printf("erro ao remover conta fixa: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) changeStatus(w http.ResponseWriter, r *http.Request, userID int64, status string) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	b, err := h.repo.UpdateStatus(r.Context(), userID, id, status)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrada")
			return
		}
		log.Printf("erro ao alterar status da conta fixa: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) freeze(w http.ResponseWriter, r *http.Request, userID int64) {
	h.changeStatus(w, r, userID, StatusFrozen)
}

func (h *Handler) reactivate(w http.ResponseWriter, r *http.Request, userID int64) {
	h.changeStatus(w, r, userID, StatusActive)
}

func (h *Handler) close(w http.ResponseWriter, r *http.Request, userID int64) {
	h.changeStatus(w, r, userID, StatusClosed)
}

type payRequest struct {
	Bank          *string  `json:"bank"`
	PaymentMethod *string  `json:"paymentMethod"`
	PaidAt        *string  `json:"paidAt"`
	AmountPaid    *float64 `json:"amountPaid"`
	AccountID     *int64   `json:"accountId"`
	Notes         *string  `json:"notes"`
}

type payResponse struct {
	Payment *Payment   `json:"payment"`
	Bill    *FixedBill `json:"bill"`
}

func (h *Handler) pay(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req payRequest
	if r.ContentLength != 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
			return
		}
	}

	payment, bill, err := h.repo.Pay(r.Context(), userID, id, PayInput{
		Bank: req.Bank, PaymentMethod: req.PaymentMethod, PaidAt: req.PaidAt,
		AmountPaid: req.AmountPaid, AccountID: req.AccountID, Notes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrada")
			return
		}
		if errors.Is(err, ErrNotActive) {
			writeError(w, http.StatusUnprocessableEntity, "conta fixa não está ativa")
			return
		}
		log.Printf("erro ao pagar conta fixa: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, payResponse{Payment: payment, Bill: bill})
}

func (h *Handler) payments(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	payments, err := h.repo.ListPayments(r.Context(), userID, id)
	if err != nil {
		log.Printf("erro ao listar pagamentos da conta fixa: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, payments)
}

func validateFixedBillRequest(req *fixedBillRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return errors.New("nome é obrigatório")
	}
	if req.Amount <= 0 {
		return errors.New("valor deve ser maior que zero")
	}
	if !IsValidPeriodicity(req.Periodicity) {
		return errors.New("periodicidade inválida")
	}
	if req.DueDate == "" {
		return errors.New("data de vencimento é obrigatória")
	}
	return nil
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("erro ao codificar resposta JSON: %v", err)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
