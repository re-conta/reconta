package share

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/re-conta/reconta/api/internal/account"
	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/category"
	"github.com/re-conta/reconta/api/internal/notification"
	"github.com/re-conta/reconta/api/internal/tag"
	"github.com/re-conta/reconta/api/internal/transaction"
	"github.com/re-conta/reconta/api/internal/user"
)

type Handler struct {
	repo          *Repository
	transactions  *transaction.Repository
	accounts      *account.Repository
	categories    *category.Repository
	tags          *tag.Repository
	users         *user.Repository
	notifications *notification.Repository
	hub           *notification.Hub
	auth          *auth.Handler
}

func NewHandler(
	repo *Repository,
	transactions *transaction.Repository,
	accounts *account.Repository,
	categories *category.Repository,
	tags *tag.Repository,
	users *user.Repository,
	notifications *notification.Repository,
	hub *notification.Hub,
	authHandler *auth.Handler,
) *Handler {
	return &Handler{
		repo: repo, transactions: transactions, accounts: accounts, categories: categories,
		tags: tags, users: users, notifications: notifications, hub: hub, auth: authHandler,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/shares", h.auth.RequireUser(h.create))
	mux.HandleFunc("GET /api/shares/sent", h.auth.RequireUser(h.listSent))
	mux.HandleFunc("GET /api/shares/received", h.auth.RequireUser(h.listReceived))
	mux.HandleFunc("POST /api/shares/{id}/accept", h.auth.RequireUser(h.accept))
	mux.HandleFunc("POST /api/shares/{id}/reject", h.auth.RequireUser(h.reject))
	mux.HandleFunc("DELETE /api/shares/{id}", h.auth.RequireUser(h.cancel))
	mux.HandleFunc("GET /api/shares/{id}/accounts", h.auth.RequireUser(h.accountsForShare))
	mux.HandleFunc("GET /api/shares/{id}/categories", h.auth.RequireUser(h.categoriesForShare))
	mux.HandleFunc("GET /api/shares/{id}/transactions", h.auth.RequireUser(h.listTransactions))
	mux.HandleFunc("POST /api/shares/{id}/transactions", h.auth.RequireUser(h.createTransaction))
	mux.HandleFunc("PUT /api/shares/{id}/transactions/{txId}", h.auth.RequireUser(h.updateTransaction))
	mux.HandleFunc("DELETE /api/shares/{id}/transactions/{txId}", h.auth.RequireUser(h.deleteTransaction))
}

type createRequest struct {
	RecipientEmail string  `json:"recipientEmail"`
	AccountIDs     []int64 `json:"accountIds"`
	CanEdit        bool    `json:"canEdit"`
	IncludeFuture  bool    `json:"includeFuture"`
	PeriodStart    *string `json:"periodStart"`
	PeriodEnd      *string `json:"periodEnd"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request, userID int64) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}

	if req.RecipientEmail == "" || len(req.AccountIDs) == 0 {
		writeError(w, http.StatusUnprocessableEntity, "e-mail do convidado e ao menos uma conta são obrigatórios")
		return
	}
	if !req.IncludeFuture && (req.PeriodEnd == nil || *req.PeriodEnd == "") {
		writeError(w, http.StatusUnprocessableEntity, "período final é obrigatório quando transações futuras não são incluídas")
		return
	}

	recipient, err := h.users.GetByEmail(r.Context(), req.RecipientEmail)
	if err != nil {
		writeError(w, http.StatusNotFound, "nenhum usuário cadastrado com esse e-mail")
		return
	}
	if recipient.ID == userID {
		writeError(w, http.StatusUnprocessableEntity, "não é possível compartilhar consigo mesmo")
		return
	}

	for _, accountID := range req.AccountIDs {
		if _, err := h.accounts.GetByID(r.Context(), userID, accountID); err != nil {
			writeError(w, http.StatusUnprocessableEntity, fmt.Sprintf("conta %d não encontrada", accountID))
			return
		}
	}

	s, err := h.repo.Create(r.Context(), userID, recipient.ID, req.AccountIDs, req.CanEdit, req.IncludeFuture, req.PeriodStart, req.PeriodEnd)
	if err != nil {
		log.Printf("erro ao criar compartilhamento: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	title := "Convite de compartilhamento"
	message := fmt.Sprintf("%s quer compartilhar transações com você.", s.OwnerName)
	h.notifyShare(r.Context(), recipient.ID, s.ID, notification.KindShareInvited, title, message)

	writeJSON(w, http.StatusCreated, s)
}

func (h *Handler) listSent(w http.ResponseWriter, r *http.Request, userID int64) {
	items, err := h.repo.ListSent(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao listar compartilhamentos enviados: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) listReceived(w http.ResponseWriter, r *http.Request, userID int64) {
	items, err := h.repo.ListReceived(r.Context(), userID)
	if err != nil {
		log.Printf("erro ao listar compartilhamentos recebidos: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) accept(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	s, err := h.repo.Accept(r.Context(), userID, id)
	if err != nil {
		writeShareError(w, err)
		return
	}

	if err := h.notifications.MarkReadByShare(r.Context(), s.ID, notification.KindShareInvited); err != nil {
		log.Printf("erro ao marcar convite como lido: %v", err)
	}
	title := "Compartilhamento aceito"
	message := fmt.Sprintf("%s aceitou seu convite de compartilhamento.", s.RecipientName)
	h.notifyShare(r.Context(), s.OwnerID, s.ID, notification.KindShareAccepted, title, message)

	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) reject(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	s, err := h.repo.Reject(r.Context(), userID, id)
	if err != nil {
		writeShareError(w, err)
		return
	}

	if err := h.notifications.MarkReadByShare(r.Context(), s.ID, notification.KindShareInvited); err != nil {
		log.Printf("erro ao marcar convite como lido: %v", err)
	}
	title := "Compartilhamento rejeitado"
	message := fmt.Sprintf("%s rejeitou seu convite de compartilhamento.", s.RecipientName)
	h.notifyShare(r.Context(), s.OwnerID, s.ID, notification.KindShareRejected, title, message)

	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	s, err := h.repo.Cancel(r.Context(), userID, id)
	if err != nil {
		writeShareError(w, err)
		return
	}

	// Se o convite ainda estava pendente, a notificação original também
	// precisa ser liberada — senão fica presa para sempre no estado "requer ação".
	if err := h.notifications.MarkReadByShare(r.Context(), s.ID, notification.KindShareInvited); err != nil {
		log.Printf("erro ao marcar convite como lido: %v", err)
	}
	title := "Compartilhamento cancelado"
	message := fmt.Sprintf("%s cancelou o compartilhamento de transações com você.", s.OwnerName)
	h.notifyShare(r.Context(), s.RecipientID, s.ID, notification.KindShareCancelled, title, message)

	writeJSON(w, http.StatusOK, s)
}

func (h *Handler) notifyShare(ctx context.Context, userID, shareID int64, kind, title, message string) {
	notif, err := h.notifications.CreateForShare(ctx, userID, shareID, kind, title, message)
	if err != nil {
		log.Printf("erro ao criar notificação de compartilhamento: %v", err)
		return
	}
	if payload, err := json.Marshal(notif); err == nil {
		h.hub.Publish(userID, payload)
	}
}

// --- dados compartilhados (visão do convidado) ---

func (h *Handler) requireGrant(w http.ResponseWriter, r *http.Request, userID int64) (*AccessGrant, bool) {
	id, err := parseID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return nil, false
	}
	grant, err := h.repo.GetActiveGrant(r.Context(), userID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "compartilhamento não encontrado ou não aceito")
			return nil, false
		}
		log.Printf("erro ao resolver acesso ao compartilhamento: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return nil, false
	}
	return grant, true
}

func (h *Handler) accountsForShare(w http.ResponseWriter, r *http.Request, userID int64) {
	grant, ok := h.requireGrant(w, r, userID)
	if !ok {
		return
	}
	all, err := h.accounts.List(r.Context(), grant.OwnerID)
	if err != nil {
		log.Printf("erro ao listar contas do compartilhamento: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	allowed := map[int64]bool{}
	for _, id := range grant.AccountIDs {
		allowed[id] = true
	}
	filtered := make([]any, 0, len(grant.AccountIDs))
	for _, a := range all {
		if allowed[a.ID] {
			filtered = append(filtered, a)
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (h *Handler) categoriesForShare(w http.ResponseWriter, r *http.Request, userID int64) {
	grant, ok := h.requireGrant(w, r, userID)
	if !ok {
		return
	}
	cats, err := h.categories.List(r.Context(), grant.OwnerID)
	if err != nil {
		log.Printf("erro ao listar categorias do compartilhamento: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, cats)
}

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request, userID int64) {
	grant, ok := h.requireGrant(w, r, userID)
	if !ok {
		return
	}

	q := r.URL.Query()
	filters := transaction.ListFilters{
		AccountIDs: grant.AccountIDs,
		Type:       q.Get("type"),
		Search:     q.Get("search"),
	}
	filters.Month, _ = strconv.Atoi(q.Get("month"))
	filters.Year, _ = strconv.Atoi(q.Get("year"))
	filters.CategoryID, _ = strconv.ParseInt(q.Get("categoryId"), 10, 64)
	filters.Page, _ = strconv.Atoi(q.Get("page"))
	filters.Limit, _ = strconv.Atoi(q.Get("limit"))

	filters.DateFrom = intersectFrom(grant.DateFrom, q.Get("dateFrom"))
	filters.DateTo = intersectTo(grant.DateTo, q.Get("dateTo"))

	result, err := h.transactions.List(r.Context(), grant.OwnerID, filters)
	if err != nil {
		log.Printf("erro ao listar transações compartilhadas: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// intersectFrom/intersectTo combinam o limite do grant com um filtro opcional
// pedido pelo cliente, sempre respeitando o mais restritivo dos dois.
func intersectFrom(grantFrom *string, requested string) string {
	if requested != "" && (grantFrom == nil || requested > *grantFrom) {
		return requested
	}
	if grantFrom != nil {
		return *grantFrom
	}
	return ""
}

func intersectTo(grantTo *string, requested string) string {
	if requested != "" && (grantTo == nil || requested < *grantTo) {
		return requested
	}
	if grantTo != nil {
		return *grantTo
	}
	return ""
}

type transactionRequest struct {
	Date        string  `json:"date"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	CategoryID  *int64  `json:"categoryId"`
	AccountID   *int64  `json:"accountId"`
	Notes       *string `json:"notes"`
	TagIDs      []int64 `json:"tagIds"`
}

// validateWithinGrant garante que a conta e a data informadas estão dentro
// do que foi efetivamente compartilhado, para impedir que um convidado com
// permissão de edição escreva fora do escopo concedido.
func validateWithinGrant(grant *AccessGrant, accountID *int64, date string) error {
	if accountID == nil {
		return errors.New("conta é obrigatória")
	}
	allowed := false
	for _, id := range grant.AccountIDs {
		if id == *accountID {
			allowed = true
			break
		}
	}
	if !allowed {
		return errors.New("conta fora do compartilhamento")
	}
	if grant.DateFrom != nil && date < *grant.DateFrom {
		return errors.New("data fora do período compartilhado")
	}
	if grant.DateTo != nil && date > *grant.DateTo {
		return errors.New("data fora do período compartilhado")
	}
	return nil
}

func (h *Handler) createTransaction(w http.ResponseWriter, r *http.Request, userID int64) {
	grant, ok := h.requireGrant(w, r, userID)
	if !ok {
		return
	}
	if !grant.CanEdit {
		writeError(w, http.StatusForbidden, "compartilhamento não permite edição")
		return
	}

	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if req.Date == "" || req.Description == "" || req.Amount == 0 || req.Type == "" {
		writeError(w, http.StatusBadRequest, "campos obrigatórios faltando")
		return
	}
	if err := validateWithinGrant(grant, req.AccountID, req.Date); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	tx, err := h.transactions.Create(r.Context(), grant.OwnerID, transaction.Input{
		Date: req.Date, Description: req.Description, Amount: absFloat(req.Amount), Type: req.Type,
		CategoryID: req.CategoryID, AccountID: req.AccountID, Notes: req.Notes,
	})
	if err != nil {
		log.Printf("erro ao criar transação compartilhada: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if len(req.TagIDs) > 0 {
		owned, err := h.tags.FilterOwnedIDs(r.Context(), grant.OwnerID, req.TagIDs)
		if err == nil {
			h.tags.SetTransactionTags(r.Context(), tx.ID, owned)
			tx.Tags, _ = h.tags.ListByTransactionID(r.Context(), tx.ID)
		}
	}

	writeJSON(w, http.StatusCreated, tx)
}

func (h *Handler) updateTransaction(w http.ResponseWriter, r *http.Request, userID int64) {
	grant, ok := h.requireGrant(w, r, userID)
	if !ok {
		return
	}
	if !grant.CanEdit {
		writeError(w, http.StatusForbidden, "compartilhamento não permite edição")
		return
	}
	txID, err := strconv.ParseInt(r.PathValue("txId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id da transação inválido")
		return
	}

	var req transactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if err := validateWithinGrant(grant, req.AccountID, req.Date); err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	tx, err := h.transactions.Update(r.Context(), grant.OwnerID, txID, transaction.Input{
		Date: req.Date, Description: req.Description, Amount: absFloat(req.Amount), Type: req.Type,
		CategoryID: req.CategoryID, AccountID: req.AccountID, Notes: req.Notes,
	})
	if err != nil {
		if errors.Is(err, transaction.ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrado")
			return
		}
		log.Printf("erro ao atualizar transação compartilhada: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if req.TagIDs != nil {
		owned, err := h.tags.FilterOwnedIDs(r.Context(), grant.OwnerID, req.TagIDs)
		if err == nil {
			h.tags.SetTransactionTags(r.Context(), tx.ID, owned)
		}
	}
	tx.Tags, _ = h.tags.ListByTransactionID(r.Context(), tx.ID)

	writeJSON(w, http.StatusOK, tx)
}

func (h *Handler) deleteTransaction(w http.ResponseWriter, r *http.Request, userID int64) {
	grant, ok := h.requireGrant(w, r, userID)
	if !ok {
		return
	}
	if !grant.CanEdit {
		writeError(w, http.StatusForbidden, "compartilhamento não permite edição")
		return
	}
	txID, err := strconv.ParseInt(r.PathValue("txId"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id da transação inválido")
		return
	}

	if err := h.transactions.Delete(r.Context(), grant.OwnerID, txID); err != nil {
		if errors.Is(err, transaction.ErrNotFound) {
			writeError(w, http.StatusNotFound, "não encontrado")
			return
		}
		log.Printf("erro ao remover transação compartilhada: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"success": true})
}

func writeShareError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "compartilhamento não encontrado")
		return
	}
	if errors.Is(err, ErrInvalidState) {
		writeError(w, http.StatusConflict, "compartilhamento não está mais pendente")
		return
	}
	log.Printf("erro ao processar compartilhamento: %v", err)
	writeError(w, http.StatusInternalServerError, "erro interno")
}

func parseID(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
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
