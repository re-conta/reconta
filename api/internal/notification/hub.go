package notification

import "sync"

// Hub distribui eventos de notificação em tempo real para os clientes SSE
// conectados de cada usuário.
type Hub struct {
	mu      sync.Mutex
	clients map[int64]map[chan []byte]bool
}

func NewHub() *Hub {
	return &Hub{clients: make(map[int64]map[chan []byte]bool)}
}

// Subscribe registra um novo canal de eventos para o usuário e retorna uma
// função de limpeza que deve ser chamada quando o cliente desconectar.
func (h *Hub) Subscribe(userID int64) (chan []byte, func()) {
	ch := make(chan []byte, 8)

	h.mu.Lock()
	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan []byte]bool)
	}
	h.clients[userID][ch] = true
	h.mu.Unlock()

	unsubscribe := func() {
		h.mu.Lock()
		delete(h.clients[userID], ch)
		if len(h.clients[userID]) == 0 {
			delete(h.clients, userID)
		}
		h.mu.Unlock()
		close(ch)
	}
	return ch, unsubscribe
}

// Publish envia o payload para todos os clientes conectados do usuário,
// descartando o envio se o canal do cliente estiver cheio (cliente lento).
func (h *Hub) Publish(userID int64, payload []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients[userID] {
		select {
		case ch <- payload:
		default:
		}
	}
}
