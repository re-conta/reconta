package email

import (
	"log"
	"sync"
	"time"
)

// rateLimit define o intervalo mínimo entre dois e-mails para o mesmo
// destinatário, imposto pela Queue. A Apple (e outros provedores) limitam a
// taxa de envio por destinatário/IP, então esse intervalo evita bloqueios.
const rateLimit = 60 * time.Second

// queueSize é o tamanho do buffer de jobs pendentes. Se a fila encher, novos
// e-mails são descartados (com log) em vez de bloquear quem os enfileira.
const queueSize = 500

type job struct {
	To      string
	Subject string
	Body    string
}

// Queue enfileira e-mails e os envia em uma goroutine dedicada, garantindo no
// máximo um envio por destinatário a cada rateLimit. E-mails para
// destinatários diferentes não esperam uns pelos outros: quando um job
// precisa aguardar sua janela, ele é reagendado sem travar o processamento
// dos demais.
type Queue struct {
	mailer *Mailer
	jobs   chan job

	mu       sync.Mutex
	lastSent map[string]time.Time
}

// NewQueue cria a fila e inicia a goroutine de envio em background.
func NewQueue(mailer *Mailer) *Queue {
	q := &Queue{
		mailer:   mailer,
		jobs:     make(chan job, queueSize),
		lastSent: make(map[string]time.Time),
	}
	go q.run()
	return q
}

// Enqueue agenda o envio de um e-mail. Não bloqueia: se a fila estiver cheia,
// o e-mail é descartado e um aviso é registrado no log.
func (q *Queue) Enqueue(to, subject, body string) {
	select {
	case q.jobs <- job{To: to, Subject: subject, Body: body}:
	default:
		log.Printf("fila de e-mail cheia: descartando envio para %s", to)
	}
}

func (q *Queue) run() {
	for j := range q.jobs {
		wait := q.reserveSlot(j.To)
		if wait > 0 {
			// Reagenda sem bloquear o processamento de outros destinatários.
			time.AfterFunc(wait, func() { q.jobs <- j })
			continue
		}

		if err := q.mailer.Send(j.To, j.Subject, j.Body); err != nil {
			log.Printf("erro ao enviar e-mail da fila para %s: %v", j.To, err)
		}
	}
}

// reserveSlot verifica se to já pode receber um e-mail agora. Se puder,
// registra o envio imediatamente (evitando corrida com outros jobs do mesmo
// destinatário) e retorna 0. Caso contrário, retorna quanto falta esperar.
func (q *Queue) reserveSlot(to string) time.Duration {
	q.mu.Lock()
	defer q.mu.Unlock()

	if last, ok := q.lastSent[to]; ok {
		if wait := rateLimit - time.Since(last); wait > 0 {
			return wait
		}
	}
	q.lastSent[to] = time.Now()
	return 0
}
