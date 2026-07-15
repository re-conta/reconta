package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var ErrNotFound = errors.New("registro não encontrado")

const timeLayout = "2006-01-02T15:04:05.000Z"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// --- Planos ---

const selectPlan = `SELECT id, code, name, description, price_monthly, price_yearly, benefits, highlight, sort_order FROM plans`

func (r *Repository) ListPlans(ctx context.Context) ([]Plan, error) {
	rows, err := r.db.QueryContext(ctx, selectPlan+` ORDER BY sort_order, id`)
	if err != nil {
		return nil, fmt.Errorf("listando planos: %w", err)
	}
	defer rows.Close()

	plans := []Plan{}
	for rows.Next() {
		p, err := scanPlan(rows)
		if err != nil {
			return nil, err
		}
		plans = append(plans, *p)
	}
	return plans, rows.Err()
}

func (r *Repository) GetPlanByCode(ctx context.Context, code string) (*Plan, error) {
	return scanPlan(r.db.QueryRowContext(ctx, selectPlan+` WHERE code = ?`, code))
}

func (r *Repository) GetPlanByID(ctx context.Context, id int64) (*Plan, error) {
	return scanPlan(r.db.QueryRowContext(ctx, selectPlan+` WHERE id = ?`, id))
}

// UpdatePlan salva nome, descrição, preços, benefícios e destaque de um plano
// (edição feita no painel de admin). O código do plano nunca muda.
func (r *Repository) UpdatePlan(ctx context.Context, p Plan) (*Plan, error) {
	benefits, err := json.Marshal(p.Benefits)
	if err != nil {
		return nil, fmt.Errorf("codificando benefícios: %w", err)
	}
	res, err := r.db.ExecContext(ctx, `
		UPDATE plans
		SET name = ?, description = ?, price_monthly = ?, price_yearly = ?, benefits = ?, highlight = ?,
		    updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		WHERE id = ?`,
		p.Name, p.Description, p.PriceMonthly, p.PriceYearly, string(benefits), p.Highlight, p.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("atualizando plano: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, ErrNotFound
	}
	return r.GetPlanByID(ctx, p.ID)
}

func scanPlan(s scanner) (*Plan, error) {
	var p Plan
	var benefitsJSON string
	if err := s.Scan(&p.ID, &p.Code, &p.Name, &p.Description, &p.PriceMonthly, &p.PriceYearly, &benefitsJSON, &p.Highlight, &p.SortOrder); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo plano: %w", err)
	}
	if err := json.Unmarshal([]byte(benefitsJSON), &p.Benefits); err != nil || p.Benefits == nil {
		p.Benefits = []string{}
	}
	return &p, nil
}

// --- Assinaturas ---

const selectSubscription = `
	SELECT s.id, s.user_id, s.plan_id, p.code, p.name, s.cycle, s.status, s.payment_method,
	       s.started_at, s.current_period_end, s.cancel_at_period_end, s.canceled_at,
	       s.refund_amount, s.last_reminder_days, s.created_at
	FROM subscriptions s
	JOIN plans p ON p.id = s.plan_id`

// CurrentSubscription retorna a assinatura vigente do usuário: a ativa mais
// recente ou, na falta dela, a pendente mais recente. Sem nenhuma das duas o
// usuário está no plano gratuito e o retorno é ErrNotFound.
func (r *Repository) CurrentSubscription(ctx context.Context, userID int64) (*Subscription, error) {
	return scanSubscription(r.db.QueryRowContext(ctx,
		selectSubscription+`
		WHERE s.user_id = ? AND s.status IN ('active', 'pending')
		ORDER BY CASE s.status WHEN 'active' THEN 0 ELSE 1 END, s.id DESC
		LIMIT 1`, userID,
	))
}

func (r *Repository) GetSubscription(ctx context.Context, id int64) (*Subscription, error) {
	return scanSubscription(r.db.QueryRowContext(ctx, selectSubscription+` WHERE s.id = ?`, id))
}

// CreateSubscription cria uma assinatura pendente e cancela pendências
// anteriores do usuário (checkouts abandonados) para não acumular lixo.
func (r *Repository) CreateSubscription(ctx context.Context, userID, planID int64, cycle, method string) (*Subscription, error) {
	if _, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET status = 'canceled', canceled_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
		 updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
		 WHERE user_id = ? AND status = 'pending'`, userID,
	); err != nil {
		return nil, fmt.Errorf("cancelando assinaturas pendentes: %w", err)
	}

	res, err := r.db.ExecContext(ctx,
		`INSERT INTO subscriptions (user_id, plan_id, cycle, status, payment_method) VALUES (?, ?, ?, 'pending', ?)`,
		userID, planID, cycle, method,
	)
	if err != nil {
		return nil, fmt.Errorf("criando assinatura: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id da assinatura: %w", err)
	}
	return r.GetSubscription(ctx, id)
}

// Activate ativa (ou renova) a assinatura após um pagamento aprovado: estende
// o período vigente a partir do fim atual (ou de agora, se já venceu) e zera
// os marcadores de lembrete e cancelamento agendado. Outras assinaturas
// ativas do usuário são encerradas — só existe um plano vigente por vez.
func (r *Repository) Activate(ctx context.Context, subID int64, method string) (*Subscription, error) {
	sub, err := r.GetSubscription(ctx, subID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	base := now
	if sub.CurrentPeriodEnd != nil && sub.CurrentPeriodEnd.After(now) {
		base = *sub.CurrentPeriodEnd
	}
	var periodEnd time.Time
	if sub.Cycle == CycleYearly {
		periodEnd = base.AddDate(1, 0, 0)
	} else {
		periodEnd = base.AddDate(0, 1, 0)
	}

	if _, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET status = 'canceled', canceled_at = ?, updated_at = ?
		 WHERE user_id = ? AND status = 'active' AND id != ?`,
		now.Format(timeLayout), now.Format(timeLayout), sub.UserID, subID,
	); err != nil {
		return nil, fmt.Errorf("encerrando assinaturas anteriores: %w", err)
	}

	if _, err := r.db.ExecContext(ctx, `
		UPDATE subscriptions
		SET status = 'active',
		    payment_method = ?,
		    started_at = COALESCE(started_at, ?),
		    current_period_end = ?,
		    cancel_at_period_end = 0,
		    canceled_at = NULL,
		    last_reminder_days = NULL,
		    updated_at = ?
		WHERE id = ?`,
		method, now.Format(timeLayout), periodEnd.Format(timeLayout), now.Format(timeLayout), subID,
	); err != nil {
		return nil, fmt.Errorf("ativando assinatura: %w", err)
	}
	return r.GetSubscription(ctx, subID)
}

// Cancel encerra a assinatura. Com immediate=true o acesso termina agora
// (cancelamento com reembolso parcial); caso contrário a assinatura segue
// ativa até o fim do ciclo e apenas fica marcada para não renovar.
func (r *Repository) Cancel(ctx context.Context, subID int64, immediate bool, refundAmount *float64) error {
	now := time.Now().UTC().Format(timeLayout)
	var err error
	if immediate {
		_, err = r.db.ExecContext(ctx,
			`UPDATE subscriptions SET status = 'canceled', canceled_at = ?, refund_amount = ?, updated_at = ? WHERE id = ?`,
			now, refundAmount, now, subID,
		)
	} else {
		_, err = r.db.ExecContext(ctx,
			`UPDATE subscriptions SET cancel_at_period_end = 1, canceled_at = ?, updated_at = ? WHERE id = ?`,
			now, now, subID,
		)
	}
	if err != nil {
		return fmt.Errorf("cancelando assinatura: %w", err)
	}
	return nil
}

// ListActiveEndingBy retorna as assinaturas ativas cujo período vigente
// termina até o instante informado — insumo da varredura de renovação.
func (r *Repository) ListActiveEndingBy(ctx context.Context, until time.Time) ([]Subscription, error) {
	rows, err := r.db.QueryContext(ctx,
		selectSubscription+` WHERE s.status = 'active' AND s.current_period_end IS NOT NULL AND s.current_period_end <= ?`,
		until.UTC().Format(timeLayout),
	)
	if err != nil {
		return nil, fmt.Errorf("listando assinaturas a vencer: %w", err)
	}
	defer rows.Close()

	subs := []Subscription{}
	for rows.Next() {
		sub, err := scanSubscription(rows)
		if err != nil {
			return nil, err
		}
		subs = append(subs, *sub)
	}
	return subs, rows.Err()
}

// SetReminderDays registra o menor estágio de lembrete já enviado (em dias
// antes do vencimento), deduplicando os avisos entre varreduras.
func (r *Repository) SetReminderDays(ctx context.Context, subID int64, days int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET last_reminder_days = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ?`,
		days, subID,
	)
	if err != nil {
		return fmt.Errorf("registrando lembrete de renovação: %w", err)
	}
	return nil
}

// Expire marca como expirada uma assinatura cujo período terminou.
func (r *Repository) Expire(ctx context.Context, subID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE subscriptions SET status = 'expired', updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ? AND status = 'active'`,
		subID,
	)
	if err != nil {
		return fmt.Errorf("expirando assinatura: %w", err)
	}
	return nil
}

func scanSubscription(s scanner) (*Subscription, error) {
	var sub Subscription
	var startedAt, periodEnd, canceledAt sql.NullString
	var refund sql.NullFloat64
	var reminderDays sql.NullInt64
	var createdAt string

	if err := s.Scan(
		&sub.ID, &sub.UserID, &sub.PlanID, &sub.PlanCode, &sub.PlanName, &sub.Cycle, &sub.Status,
		&sub.PaymentMethod, &startedAt, &periodEnd, &sub.CancelAtPeriodEnd, &canceledAt,
		&refund, &reminderDays, &createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo assinatura: %w", err)
	}

	sub.StartedAt = parseNullTime(startedAt)
	sub.CurrentPeriodEnd = parseNullTime(periodEnd)
	sub.CanceledAt = parseNullTime(canceledAt)
	if refund.Valid {
		sub.RefundAmount = &refund.Float64
	}
	if reminderDays.Valid {
		d := int(reminderDays.Int64)
		sub.LastReminderDays = &d
	}
	sub.CreatedAt = parseTime(createdAt)
	return &sub, nil
}

// --- Pagamentos ---

const selectPayment = `
	SELECT id, subscription_id, user_id, mp_payment_id, amount, method, status, status_detail,
	       COALESCE(pix_qr, ''), COALESCE(pix_qr_base64, ''), COALESCE(ticket_url, ''), created_at
	FROM subscription_payments`

func (r *Repository) CreatePayment(ctx context.Context, p Payment) (*Payment, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO subscription_payments (subscription_id, user_id, mp_payment_id, amount, method, status, status_detail, pix_qr, pix_qr_base64, ticket_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.SubscriptionID, p.UserID, p.MPPaymentID, p.Amount, p.Method, p.Status, p.StatusDetail, p.PixQR, p.PixQRBase64, p.TicketURL,
	)
	if err != nil {
		return nil, fmt.Errorf("criando pagamento: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("obtendo id do pagamento: %w", err)
	}
	return r.GetPayment(ctx, id)
}

func (r *Repository) GetPayment(ctx context.Context, id int64) (*Payment, error) {
	return scanPayment(r.db.QueryRowContext(ctx, selectPayment+` WHERE id = ?`, id))
}

func (r *Repository) GetUserPayment(ctx context.Context, userID, id int64) (*Payment, error) {
	return scanPayment(r.db.QueryRowContext(ctx, selectPayment+` WHERE id = ? AND user_id = ?`, id, userID))
}

func (r *Repository) GetPaymentByMPID(ctx context.Context, mpPaymentID int64) (*Payment, error) {
	return scanPayment(r.db.QueryRowContext(ctx, selectPayment+` WHERE mp_payment_id = ?`, mpPaymentID))
}

// LastApprovedPayment retorna o pagamento aprovado mais recente da
// assinatura — base do cálculo de reembolso parcial no cancelamento.
func (r *Repository) LastApprovedPayment(ctx context.Context, subID int64) (*Payment, error) {
	return scanPayment(r.db.QueryRowContext(ctx,
		selectPayment+` WHERE subscription_id = ? AND status = 'approved' ORDER BY id DESC LIMIT 1`, subID,
	))
}

func (r *Repository) UpdatePaymentStatus(ctx context.Context, id int64, status, statusDetail string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE subscription_payments SET status = ?, status_detail = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now') WHERE id = ?`,
		status, statusDetail, id,
	)
	if err != nil {
		return fmt.Errorf("atualizando status do pagamento: %w", err)
	}
	return nil
}

func scanPayment(s scanner) (*Payment, error) {
	var p Payment
	var mpID sql.NullInt64
	var createdAt string
	if err := s.Scan(
		&p.ID, &p.SubscriptionID, &p.UserID, &mpID, &p.Amount, &p.Method, &p.Status, &p.StatusDetail,
		&p.PixQR, &p.PixQRBase64, &p.TicketURL, &createdAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("lendo pagamento: %w", err)
	}
	if mpID.Valid {
		p.MPPaymentID = &mpID.Int64
	}
	p.CreatedAt = parseTime(createdAt)
	return &p, nil
}

// --- Helpers ---

type scanner interface {
	Scan(dest ...any) error
}

func parseTime(s string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.999Z", s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseNullTime(s sql.NullString) *time.Time {
	if !s.Valid {
		return nil
	}
	t := parseTime(s.String)
	if t.IsZero() {
		return nil
	}
	return &t
}
