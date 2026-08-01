package ws

import (
	"github.com/Atharv-3105/Semantic-Duel/internal/logger"
	"github.com/Atharv-3105/Semantic-Duel/internal/metrics"
)

// Hub is the central brain of the WebSocket system.
// All client registration and deregistration flows through here in a single goroutine,
// keeping the clients map free of data races without a mutex.
type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client

	// Buffered so Hub.Run() never blocks while the consumer goroutines are busy.
	OnConnect    chan *Client
	OnDisconnect chan *Client

	log *logger.Logger
}

func NewHub(log *logger.Logger) *Hub {
	return &Hub{
		clients:      make(map[*Client]bool),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		OnConnect:    make(chan *Client, 64),
		OnDisconnect: make(chan *Client, 64),
		log:          log,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			metrics.IncConnections()
			h.log.Info("[WS] client connected", "active", len(h.clients))
			h.OnConnect <- client

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				metrics.DecConnections()
				h.log.Info("[WS] client disconnected", "active", len(h.clients))
				h.OnDisconnect <- client
			}
		}
	}
}