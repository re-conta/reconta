package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/re-conta/reconta/api/internal/auth"
	"github.com/re-conta/reconta/api/internal/email"
	"github.com/re-conta/reconta/api/internal/notification"
	"github.com/re-conta/reconta/api/internal/user"
)

// Estágios de lembrete de renovação, em dias antes do fim do período.
var reminderStages = []int{7, 3, 1}

// Kinds das notificações de assinatura exibidas no sino do site.
const (
	KindSubscriptionExpiring = "subscription_expiring"
	KindSubscriptionExpired  = "subscription_expired"
	KindSubscriptionActive   = "subscription_active"
)

type Handler struct {
	repo          *Repository
	auth          *auth.Handler
	users         *user.Repository
	notifications *notification.Repository
	hub           *notification.Hub
	mailQueue     *email.Queue
	gateway       *Gateway
	internalToken string
	webhookSecret string
	appURL        string
}

func NewHandler(
	repo *Repository,
	authHandler *auth.Handler,
	users *user.Repository,
	notifications *notification.Repository,
	hub *notification.Hub,
	mailQueue *email.Queue,
	gateway *Gateway,
	internalToken, webhookSecret, appURL string,
) *Handler {
	return &Handler{
		repo: repo, auth: authHandler, users: users, notifications: notifications,
		hub: hub, mailQueue: mailQueue, gateway: gateway,
		internalToken: internalToken, webhookSecret: webhookSecret, appURL: appURL,
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/plans", h.listPlans)
	mux.HandleFunc("GET /api/billing/subscription", h.auth.RequireUser(h.getSubscription))
	mux.HandleFunc("POST /api/billing/subscribe", h.auth.RequireUser(h.subscribe))
	mux.HandleFunc("GET /api/billing/payments/{id}", h.auth.RequireUser(h.paymentStatus))
	mux.HandleFunc("POST /api/billing/subscription/cancel", h.auth.RequireUser(h.cancel))

	mux.HandleFunc("POST /api/billing/webhook", h.webhook)
	mux.HandleFunc("POST /api/internal/billing/scan", h.requireInternalToken(h.scan))

	mux.HandleFunc("GET /api/admin/plans", h.requirePermission(h.adminListPlans, user.PermManagePlans))
	mux.HandleFunc("PUT /api/admin/plans/{id}", h.requirePermission(h.adminUpdatePlan, user.PermManagePlans))
}

// --- Planos (público) ---

func (h *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.repo.ListPlans(r.Context())
	if err != nil {
		log.Printf("erro ao listar planos: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, plans)
}

// --- Assinatura do usuário ---

type subscriptionResponse struct {
	PlanCode     string        `json:"planCode"`
	Subscription *Subscription `json:"subscription"`
}

func (h *Handler) getSubscription(w http.ResponseWriter, r *http.Request, userID int64) {
	sub, err := h.repo.CurrentSubscription(r.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusOK, subscriptionResponse{PlanCode: PlanFree})
			return
		}
		log.Printf("erro ao buscar assinatura do usuário %d: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	code := sub.PlanCode
	if sub.Status == StatusPending {
		// Checkout ainda não concluído: para o site o usuário segue no gratuito.
		code = PlanFree
	}
	writeJSON(w, http.StatusOK, subscriptionResponse{PlanCode: code, Subscription: sub})
}

// --- Checkout ---

type subscribeRequest struct {
	PlanCode        string `json:"planCode"`
	Cycle           string `json:"cycle"`
	Method          string `json:"method"`
	Token           string `json:"token"`
	PaymentMethodID string `json:"paymentMethodId"`
	IssuerID        string `json:"issuerId"`
	Installments    int    `json:"installments"`
	DocType         string `json:"docType"`
	DocNumber       string `json:"docNumber"`
	ZipCode         string `json:"zipCode"`
	StreetName      string `json:"streetName"`
	StreetNumber    string `json:"streetNumber"`
	Neighborhood    string `json:"neighborhood"`
	City            string `json:"city"`
	FederalUnit     string `json:"federalUnit"`
}

type subscribeResponse struct {
	Payment      *Payment      `json:"payment"`
	Subscription *Subscription `json:"subscription"`
}

func (h *Handler) subscribe(w http.ResponseWriter, r *http.Request, userID int64) {
	if h.gateway == nil {
		writeError(w, http.StatusServiceUnavailable, "pagamentos indisponíveis: integração com Mercado Pago não configurada")
		return
	}

	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if !ValidCycle(req.Cycle) {
		writeError(w, http.StatusBadRequest, "ciclo inválido: use monthly ou yearly")
		return
	}
	if !ValidMethod(req.Method) {
		writeError(w, http.StatusBadRequest, "método de pagamento inválido")
		return
	}

	ctx := r.Context()
	plan, err := h.repo.GetPlanByCode(ctx, req.PlanCode)
	if err != nil {
		writeError(w, http.StatusNotFound, "plano não encontrado")
		return
	}
	if plan.IsFree() {
		writeError(w, http.StatusUnprocessableEntity, "o plano gratuito não precisa de assinatura")
		return
	}
	amount := plan.PriceFor(req.Cycle)
	if amount <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "plano sem preço configurado para este ciclo")
		return
	}

	switch req.Method {
	case MethodDebit, MethodCredit:
		if req.Token == "" || req.PaymentMethodID == "" {
			writeError(w, http.StatusBadRequest, "dados do cartão incompletos")
			return
		}
	case MethodBoleto:
		if req.DocNumber == "" || req.ZipCode == "" || req.StreetName == "" || req.City == "" || req.FederalUnit == "" {
			writeError(w, http.StatusBadRequest, "boleto exige CPF/CNPJ e endereço completo")
			return
		}
	}

	u, err := h.users.GetByID(ctx, userID)
	if err != nil {
		log.Printf("erro ao buscar usuário %d para checkout: %v", userID, err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	// Renovação reaproveita a assinatura ativa; troca de plano (ou primeira
	// compra) cria uma pendente que só entra em vigor com o pagamento aprovado.
	sub, err := h.repo.CurrentSubscription(ctx, userID)
	isRenewal := err == nil && sub.Status == StatusActive && sub.PlanID == plan.ID && sub.Cycle == req.Cycle
	if !isRenewal {
		sub, err = h.repo.CreateSubscription(ctx, userID, plan.ID, req.Cycle, req.Method)
		if err != nil {
			log.Printf("erro ao criar assinatura: %v", err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
	}

	cycleLabel := "mensal"
	if req.Cycle == CycleYearly {
		cycleLabel = "anual"
	}
	firstName, lastName := splitName(u.Name)
	result, err := h.gateway.CreatePayment(ctx, PaymentInput{
		Amount:            amount,
		Description:       fmt.Sprintf("Reconta — Plano %s (%s)", plan.Name, cycleLabel),
		Method:            req.Method,
		PaymentMethodID:   req.PaymentMethodID,
		Token:             req.Token,
		Installments:      req.Installments,
		IssuerID:          req.IssuerID,
		PayerEmail:        u.Email,
		PayerFirstName:    firstName,
		PayerLastName:     lastName,
		DocType:           req.DocType,
		DocNumber:         req.DocNumber,
		ZipCode:           req.ZipCode,
		StreetName:        req.StreetName,
		StreetNumber:      req.StreetNumber,
		Neighborhood:      req.Neighborhood,
		City:              req.City,
		FederalUnit:       req.FederalUnit,
		ExternalReference: fmt.Sprintf("reconta-sub-%d", sub.ID),
		NotificationURL:   h.appURL + "/api/billing/webhook",
	})
	if err != nil {
		log.Printf("erro ao criar pagamento no Mercado Pago: %v", err)
		writeError(w, http.StatusBadGateway, "o Mercado Pago recusou a criação do pagamento; confira os dados e tente novamente")
		return
	}

	pay, err := h.repo.CreatePayment(ctx, Payment{
		SubscriptionID: sub.ID,
		UserID:         userID,
		MPPaymentID:    &result.ID,
		Amount:         amount,
		Method:         req.Method,
		Status:         result.Status,
		StatusDetail:   result.StatusDetail,
		PixQR:          result.QRCode,
		PixQRBase64:    result.QRCodeBase64,
		TicketURL:      result.TicketURL,
	})
	if err != nil {
		log.Printf("erro ao registrar pagamento local: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	if result.Status == "approved" {
		sub, err = h.activateForPayment(ctx, pay)
		if err != nil {
			log.Printf("erro ao ativar assinatura %d: %v", pay.SubscriptionID, err)
		}
	} else if result.Status == "rejected" || result.Status == "cancelled" {
		// Cartão recusado na hora: devolve o motivo sem esperar polling.
	}

	writeJSON(w, http.StatusOK, subscribeResponse{Payment: pay, Subscription: sub})
}

// paymentStatus é a rota de polling do modal de checkout: enquanto o
// pagamento estiver pendente, reconsulta o Mercado Pago (cobre também o
// ambiente local, onde o webhook não alcança o servidor).
func (h *Handler) paymentStatus(w http.ResponseWriter, r *http.Request, userID int64) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}
	pay, err := h.repo.GetUserPayment(r.Context(), userID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "pagamento não encontrado")
		return
	}

	if isPendingStatus(pay.Status) && pay.MPPaymentID != nil {
		if updated, err := h.refreshPayment(r.Context(), pay); err == nil {
			pay = updated
		} else {
			log.Printf("erro ao atualizar pagamento %d: %v", pay.ID, err)
		}
	}
	writeJSON(w, http.StatusOK, pay)
}

// --- Cancelamento ---

type cancelRequest struct {
	// Mode: "refund" cancela agora com reembolso parcial do tempo não usado;
	// "end_of_cycle" mantém o acesso até o fim do período, sem reembolso.
	Mode string `json:"mode"`
}

type cancelResponse struct {
	Subscription *Subscription `json:"subscription"`
	RefundAmount float64       `json:"refundAmount"`
}

func (h *Handler) cancel(w http.ResponseWriter, r *http.Request, userID int64) {
	var req cancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if req.Mode != "refund" && req.Mode != "end_of_cycle" {
		writeError(w, http.StatusBadRequest, "modo inválido: use refund ou end_of_cycle")
		return
	}

	ctx := r.Context()
	sub, err := h.repo.CurrentSubscription(ctx, userID)
	if err != nil || sub.Status != StatusActive {
		writeError(w, http.StatusUnprocessableEntity, "você não possui assinatura ativa para cancelar")
		return
	}

	refundAmount := 0.0
	if req.Mode == "refund" {
		refundAmount = h.processRefund(ctx, sub)
		if err := h.repo.Cancel(ctx, sub.ID, true, &refundAmount); err != nil {
			log.Printf("erro ao cancelar assinatura %d: %v", sub.ID, err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
	} else {
		if err := h.repo.Cancel(ctx, sub.ID, false, nil); err != nil {
			log.Printf("erro ao agendar cancelamento da assinatura %d: %v", sub.ID, err)
			writeError(w, http.StatusInternalServerError, "erro interno")
			return
		}
	}

	updated, err := h.repo.GetSubscription(ctx, sub.ID)
	if err != nil {
		updated = sub
	}
	writeJSON(w, http.StatusOK, cancelResponse{Subscription: updated, RefundAmount: refundAmount})
}

// processRefund calcula o valor proporcional ao tempo restante do ciclo e
// solicita o reembolso parcial ao Mercado Pago. Falha no reembolso não impede
// o cancelamento — o valor devolvido fica 0 e o erro vai para o log.
func (h *Handler) processRefund(ctx context.Context, sub *Subscription) float64 {
	if sub.CurrentPeriodEnd == nil {
		return 0
	}
	pay, err := h.repo.LastApprovedPayment(ctx, sub.ID)
	if err != nil || pay.MPPaymentID == nil {
		return 0
	}

	end := *sub.CurrentPeriodEnd
	var start time.Time
	if sub.Cycle == CycleYearly {
		start = end.AddDate(-1, 0, 0)
	} else {
		start = end.AddDate(0, -1, 0)
	}
	now := time.Now().UTC()
	total := end.Sub(start)
	remaining := end.Sub(now)
	if total <= 0 || remaining <= 0 {
		return 0
	}

	amount := math.Floor(pay.Amount*(remaining.Seconds()/total.Seconds())*100) / 100
	if amount < 0.01 {
		return 0
	}
	if amount > pay.Amount {
		amount = pay.Amount
	}

	if err := h.gateway.PartialRefund(ctx, *pay.MPPaymentID, amount); err != nil {
		log.Printf("erro ao reembolsar pagamento %d (assinatura %d): %v", *pay.MPPaymentID, sub.ID, err)
		return 0
	}
	return amount
}

// --- Webhook do Mercado Pago ---

// webhook recebe as notificações de pagamento do Mercado Pago. O status
// nunca é confiado ao corpo recebido: o pagamento é sempre reconsultado na
// API antes de qualquer mudança local.
func (h *Handler) webhook(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	dataID := query.Get("data.id")
	if dataID == "" {
		dataID = query.Get("id")
	}

	var body struct {
		Type string `json:"type"`
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if dataID == "" {
		dataID = body.Data.ID
	}

	if h.webhookSecret != "" && !h.validSignature(r, dataID) {
		writeError(w, http.StatusUnauthorized, "assinatura do webhook inválida")
		return
	}

	if t := query.Get("type"); (t != "" && t != "payment") || (body.Type != "" && body.Type != "payment") {
		w.WriteHeader(http.StatusOK)
		return
	}

	mpID, err := strconv.ParseInt(dataID, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	pay, err := h.repo.GetPaymentByMPID(r.Context(), mpID)
	if err != nil {
		// Pagamento desconhecido (ex.: teste do painel do MP): confirma
		// recebimento para o Mercado Pago não reenviar indefinidamente.
		w.WriteHeader(http.StatusOK)
		return
	}
	if _, err := h.refreshPayment(r.Context(), pay); err != nil {
		log.Printf("erro ao processar webhook do pagamento %d: %v", mpID, err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// validSignature valida o header x-signature (ts=...,v1=...) conforme a
// documentação de webhooks do Mercado Pago: HMAC-SHA256 do manifesto
// "id:{data.id};request-id:{x-request-id};ts:{ts};" com a assinatura secreta.
func (h *Handler) validSignature(r *http.Request, dataID string) bool {
	var ts, v1 string
	for part := range strings.SplitSeq(r.Header.Get("x-signature"), ",") {
		key, value, found := strings.Cut(strings.TrimSpace(part), "=")
		if !found {
			continue
		}
		switch strings.TrimSpace(key) {
		case "ts":
			ts = strings.TrimSpace(value)
		case "v1":
			v1 = strings.TrimSpace(value)
		}
	}
	if ts == "" || v1 == "" {
		return false
	}

	manifest := fmt.Sprintf("id:%s;request-id:%s;ts:%s;", strings.ToLower(dataID), r.Header.Get("x-request-id"), ts)
	mac := hmac.New(sha256.New, []byte(h.webhookSecret))
	mac.Write([]byte(manifest))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(v1))
}

// refreshPayment reconsulta o pagamento no Mercado Pago, sincroniza o status
// local e ativa a assinatura quando o pagamento foi aprovado.
func (h *Handler) refreshPayment(ctx context.Context, pay *Payment) (*Payment, error) {
	result, err := h.gateway.GetPayment(ctx, *pay.MPPaymentID)
	if err != nil {
		return nil, err
	}
	if result.Status == pay.Status {
		return pay, nil
	}
	if err := h.repo.UpdatePaymentStatus(ctx, pay.ID, result.Status, result.StatusDetail); err != nil {
		return nil, err
	}

	if result.Status == "approved" {
		updated, err := h.repo.GetPayment(ctx, pay.ID)
		if err != nil {
			return nil, err
		}
		if _, err := h.activateForPayment(ctx, updated); err != nil {
			return nil, err
		}
		return updated, nil
	}
	return h.repo.GetPayment(ctx, pay.ID)
}

// activateForPayment ativa a assinatura de um pagamento aprovado e avisa o
// usuário no site e por e-mail.
func (h *Handler) activateForPayment(ctx context.Context, pay *Payment) (*Subscription, error) {
	sub, err := h.repo.Activate(ctx, pay.SubscriptionID, pay.Method)
	if err != nil {
		return nil, err
	}

	title := fmt.Sprintf("Assinatura do plano %s ativa", sub.PlanName)
	message := fmt.Sprintf(
		"Pagamento confirmado! Sua assinatura do plano %s está ativa até %s.",
		sub.PlanName, formatDate(sub.CurrentPeriodEnd),
	)
	h.notifyUser(ctx, sub.UserID, KindSubscriptionActive, title, message, sub.CurrentPeriodEnd, 0, true)
	return sub, nil
}

// --- Varredura de renovação (timer systemd) ---

func (h *Handler) requireInternalToken(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h.internalToken == "" || r.Header.Get("X-Internal-Token") != h.internalToken {
			writeError(w, http.StatusUnauthorized, "não autorizado")
			return
		}
		next(w, r)
	}
}

// scan roda pelo mesmo timer systemd da varredura de notificações: envia os
// lembretes de renovação (7, 3 e 1 dia antes do vencimento, no site e por
// e-mail) e expira assinaturas cujo período terminou.
func (h *Handler) scan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().UTC()

	subs, err := h.repo.ListActiveEndingBy(ctx, now.AddDate(0, 0, reminderStages[0]))
	if err != nil {
		log.Printf("erro ao listar assinaturas a vencer: %v", err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}

	reminded, expired := 0, 0
	for i := range subs {
		sub := &subs[i]
		if sub.CurrentPeriodEnd == nil {
			continue
		}

		if !sub.CurrentPeriodEnd.After(now) {
			if err := h.repo.Expire(ctx, sub.ID); err != nil {
				log.Printf("erro ao expirar assinatura %d: %v", sub.ID, err)
				continue
			}
			expired++
			if !sub.CancelAtPeriodEnd {
				title := fmt.Sprintf("Assinatura do plano %s expirou", sub.PlanName)
				message := fmt.Sprintf(
					"Sua assinatura do plano %s expirou em %s e sua conta voltou ao plano Gratuito. Renove em %s/planos.",
					sub.PlanName, formatDate(sub.CurrentPeriodEnd), h.appURL,
				)
				h.notifyUser(ctx, sub.UserID, KindSubscriptionExpired, title, message, sub.CurrentPeriodEnd, 0, true)
			}
			continue
		}

		// Quem pediu cancelamento ao fim do ciclo não quer lembrete de renovação.
		if sub.CancelAtPeriodEnd {
			continue
		}

		// O estágio aplicável é o mais urgente já alcançado (ex.: faltando 2
		// dias, o estágio é o de 3). Cada estágio dispara no máximo uma vez.
		daysLeft := int(math.Ceil(sub.CurrentPeriodEnd.Sub(now).Hours() / 24))
		stage := 0
		for _, s := range reminderStages {
			if daysLeft <= s {
				stage = s
			}
		}
		if stage == 0 || (sub.LastReminderDays != nil && *sub.LastReminderDays <= stage) {
			continue
		}

		title := fmt.Sprintf("Assinatura do plano %s vence em %d dia(s)", sub.PlanName, daysLeft)
		message := fmt.Sprintf(
			"Sua assinatura %s do plano %s vence em %s. Renove em %s/planos para não perder os benefícios.",
			cycleLabel(sub.Cycle), sub.PlanName, formatDate(sub.CurrentPeriodEnd), h.appURL,
		)
		h.notifyUser(ctx, sub.UserID, KindSubscriptionExpiring, title, message, sub.CurrentPeriodEnd, stage*1440, true)
		if err := h.repo.SetReminderDays(ctx, sub.ID, stage); err != nil {
			log.Printf("erro ao registrar lembrete da assinatura %d: %v", sub.ID, err)
		}
		reminded++
	}

	writeJSON(w, http.StatusOK, map[string]int{"reminded": reminded, "expired": expired})
}

// notifyUser cria a notificação no site (sino + SSE) e, quando withEmail,
// envia também por e-mail. Avisos de cobrança são transacionais, então não
// dependem das preferências de notificação de contas fixas.
func (h *Handler) notifyUser(ctx context.Context, userID int64, kind, title, message string, dueDate *time.Time, offsetMinutes int, withEmail bool) {
	due := ""
	if dueDate != nil {
		due = dueDate.Format("2006-01-02")
	}
	notif, err := h.notifications.CreateGeneral(ctx, userID, kind, title, message, due, offsetMinutes)
	if err != nil {
		log.Printf("erro ao criar notificação de assinatura para usuário %d: %v", userID, err)
	} else if payload, err := json.Marshal(notif); err == nil {
		h.hub.Publish(userID, payload)
	}

	if !withEmail {
		return
	}
	u, err := h.users.GetByID(ctx, userID)
	if err != nil {
		log.Printf("erro ao buscar usuário %d para e-mail de assinatura: %v", userID, err)
		return
	}
	h.mailQueue.Enqueue(u.Email, title, message)
}

// --- Admin ---

func (h *Handler) requirePermission(next http.HandlerFunc, perm string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := h.auth.CurrentUser(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "não autenticado")
			return
		}
		if !u.HasPermission(perm) {
			writeError(w, http.StatusForbidden, "acesso negado")
			return
		}
		next(w, r)
	}
}

func (h *Handler) adminListPlans(w http.ResponseWriter, r *http.Request) {
	h.listPlans(w, r)
}

type updatePlanRequest struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	PriceMonthly float64  `json:"priceMonthly"`
	PriceYearly  float64  `json:"priceYearly"`
	Benefits     []string `json:"benefits"`
	Highlight    bool     `json:"highlight"`
}

func (h *Handler) adminUpdatePlan(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id inválido")
		return
	}

	var req updatePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "corpo da requisição inválido")
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		writeError(w, http.StatusUnprocessableEntity, "o nome do plano é obrigatório")
		return
	}
	if req.PriceMonthly < 0 || req.PriceYearly < 0 {
		writeError(w, http.StatusUnprocessableEntity, "preços não podem ser negativos")
		return
	}

	current, err := h.repo.GetPlanByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "plano não encontrado")
		return
	}
	if current.IsFree() {
		// O plano gratuito nunca vira pago por engano no painel.
		req.PriceMonthly, req.PriceYearly = 0, 0
	} else if req.PriceMonthly <= 0 || req.PriceYearly <= 0 {
		writeError(w, http.StatusUnprocessableEntity, "planos pagos precisam de preço mensal e anual maiores que zero")
		return
	}

	benefits := make([]string, 0, len(req.Benefits))
	for _, b := range req.Benefits {
		if b = strings.TrimSpace(b); b != "" {
			benefits = append(benefits, b)
		}
	}

	updated, err := h.repo.UpdatePlan(r.Context(), Plan{
		ID:           id,
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		PriceMonthly: req.PriceMonthly,
		PriceYearly:  req.PriceYearly,
		Benefits:     benefits,
		Highlight:    req.Highlight,
	})
	if err != nil {
		log.Printf("erro ao atualizar plano %d: %v", id, err)
		writeError(w, http.StatusInternalServerError, "erro interno")
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// --- Helpers ---

func isPendingStatus(status string) bool {
	return status == "pending" || status == "in_process" || status == "authorized" || status == ""
}

func splitName(name string) (first, last string) {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "Cliente", "Reconta"
	}
	if len(parts) == 1 {
		return parts[0], parts[0]
	}
	return parts[0], strings.Join(parts[1:], " ")
}

func cycleLabel(cycle string) string {
	if cycle == CycleYearly {
		return "anual"
	}
	return "mensal"
}

func formatDate(t *time.Time) string {
	if t == nil {
		return "-"
	}
	return t.Format("02/01/2006")
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
