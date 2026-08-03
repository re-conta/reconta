package advisor

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/re-conta/reconta/api/internal/health"
	"github.com/re-conta/reconta/api/internal/transaction"
)

// RateLimits define os limites de chamadas ao Groq que a Queue nunca deve
// ultrapassar. Os valores padrão (DefaultRateLimits) são deliberadamente
// conservadores: os limites reais do plano gratuito do Groq variam por
// modelo e mudam com o tempo (ver console.groq.com/settings/limits) — em vez
// de tentar acompanhar um número exato, mantemos uma margem folgada e
// permitimos ajuste fino via variáveis de ambiente sem precisar recompilar.
type RateLimits struct {
	MinInterval time.Duration // intervalo mínimo entre duas chamadas, não importa o quê
	PerMinute   int
	PerHour     int
	PerDay      int
	PerMonth    int
}

// DefaultRateLimits mantém o uso bem abaixo de qualquer plano gratuito
// razoável do Groq: no máximo 1 chamada a cada 6s, e tetos por hora/dia/mês.
func DefaultRateLimits() RateLimits {
	return RateLimits{
		MinInterval: 6 * time.Second,
		PerMinute:   10,
		PerHour:     100,
		PerDay:      500,
		PerMonth:    3000,
	}
}

type job struct {
	userID int64
	month  int
	year   int
}

func jobKey(userID int64, month, year int) string {
	return fmt.Sprintf("%d:%d:%d", userID, month, year)
}

// Queue serializa TODAS as análises pedidas por TODOS os usuários em uma
// única goroutine consumidora — portanto, no máximo uma chamada ao Groq
// acontece por vez, e sempre respeitando o intervalo mínimo e os tetos de
// RateLimits (contados via banco, então sobrevivem a reinícios do processo).
// Isso é intencional e obrigatório: o plano gratuito do Groq é compartilhado
// por toda a aplicação, então a fila não pode paralelizar entre usuários.
type Queue struct {
	repo            *Repository
	client          *GroqClient
	transactionRepo *transaction.Repository
	healthRepo      *health.Repository
	limits          RateLimits

	jobs chan job

	mu      sync.Mutex
	pending map[string]bool
}

func NewQueue(
	repo *Repository,
	client *GroqClient,
	transactionRepo *transaction.Repository,
	healthRepo *health.Repository,
	limits RateLimits,
) *Queue {
	q := &Queue{
		repo:            repo,
		client:          client,
		transactionRepo: transactionRepo,
		healthRepo:      healthRepo,
		limits:          limits,
		jobs:            make(chan job, 200),
		pending:         make(map[string]bool),
	}
	go q.run()
	return q
}

// Enqueue agenda uma análise para o usuário/mês/ano informado. Não bloqueia:
// se já houver uma análise pendente para a mesma chave, ou se a fila estiver
// cheia, a chamada é ignorada silenciosamente (o próximo carregamento da
// página tentará de novo).
func (q *Queue) Enqueue(userID int64, month, year int) {
	key := jobKey(userID, month, year)

	q.mu.Lock()
	if q.pending[key] {
		q.mu.Unlock()
		return
	}
	q.mu.Unlock()

	select {
	case q.jobs <- job{userID: userID, month: month, year: year}:
		q.mu.Lock()
		q.pending[key] = true
		q.mu.Unlock()
	default:
		log.Printf("fila de recomendações de IA cheia: descartando análise do usuário %d (%02d/%d)", userID, month, year)
	}
}

func (q *Queue) run() {
	for j := range q.jobs {
		q.waitForSlot()
		q.process(j)

		q.mu.Lock()
		delete(q.pending, jobKey(j.userID, j.month, j.year))
		q.mu.Unlock()
	}
}

// waitForSlot bloqueia a goroutine da fila (e só ela) até que uma nova
// chamada ao Groq seja permitida por todas as janelas de RateLimits.
// Bloquear aqui é o ponto central do throttling: como há uma única
// goroutine consumindo jobs, isso garante uma chamada por vez.
func (q *Queue) waitForSlot() {
	for {
		wait, ok := q.timeUntilNextSlot()
		if ok {
			return
		}
		time.Sleep(wait)
	}
}

func (q *Queue) timeUntilNextSlot() (time.Duration, bool) {
	ctx := context.Background()
	now := time.Now()

	if last, err := q.repo.LastCallAt(ctx); err != nil {
		log.Printf("erro ao verificar última chamada ao groq: %v", err)
	} else if !last.IsZero() {
		if wait := q.limits.MinInterval - now.Sub(last); wait > 0 {
			return wait, false
		}
	}

	windows := []struct {
		size time.Duration
		max  int
	}{
		{time.Minute, q.limits.PerMinute},
		{time.Hour, q.limits.PerHour},
		{24 * time.Hour, q.limits.PerDay},
		{30 * 24 * time.Hour, q.limits.PerMonth},
	}

	var maxWait time.Duration
	for _, w := range windows {
		if w.max <= 0 {
			continue
		}
		count, oldest, err := q.repo.CountCallsSince(ctx, now.Add(-w.size))
		if err != nil {
			log.Printf("erro ao contar chamadas ao groq: %v", err)
			continue
		}
		if count >= w.max {
			if wait := oldest.Add(w.size).Sub(now); wait > maxWait {
				maxWait = wait
			}
		}
	}
	if maxWait > 0 {
		return maxWait, false
	}
	return 0, true
}

func (q *Queue) process(j job) {
	ctx := context.Background()

	settings, err := q.healthRepo.GetSettings(ctx)
	if err != nil {
		log.Printf("erro ao ler configuração de saúde financeira para recomendações: %v", err)
		return
	}
	income, expense, err := q.healthRepo.MonthTotals(ctx, j.userID, j.month, j.year)
	if err != nil {
		log.Printf("erro ao calcular totais do mês para recomendações: %v", err)
		return
	}
	rate, level, stars := health.Classify(income, expense, settings)

	txs, _, err := q.transactionRepo.ListAll(ctx, j.userID, transaction.ListFilters{
		Month: j.month, Year: j.year, Type: "expense",
	})
	if err != nil {
		log.Printf("erro ao listar transações para recomendações: %v", err)
		return
	}

	signals := analyzeSignals(txs)
	prompt := buildPrompt(signals, income, expense, rate, level, stars)

	if err := q.repo.LogAPICall(ctx); err != nil {
		log.Printf("erro ao registrar chamada ao groq: %v", err)
	}

	items, err := q.client.GetRecommendations(ctx, prompt)
	if err != nil {
		log.Printf("erro ao gerar recomendações via groq para usuário %d: %v", j.userID, err)
		if err := q.repo.SaveError(ctx, j.userID, j.month, j.year, stars, err.Error()); err != nil {
			log.Printf("erro ao salvar falha de recomendações: %v", err)
		}
		return
	}

	if err := q.repo.Save(ctx, j.userID, j.month, j.year, stars, items); err != nil {
		log.Printf("erro ao salvar recomendações geradas: %v", err)
	}
}
