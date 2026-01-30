package ws

import (
	// "log"

	"github.com/Atharv-3105/Graph-Duel/internal/logger"
)

//Hub will act as the Central-Brain of the WebSocket system
type Hub struct {
	clients 	map[*Client]bool
	register	chan *Client
	unregister	chan *Client

	OnConnect 	chan *Client
	OnDisconnect chan *Client 

	log *logger.Logger
}


func NewHub(log *logger.Logger) *Hub{
	return &Hub{
		clients: 	make(map[*Client]bool),	//A map where keys are pointers to Client objects
		register:	make(chan *Client), //A channel used to signal a new client has connected and added to the hub
		unregister:	make(chan *Client),	//A channel used to signal a new client has disconnected and should be removed from hub
		OnConnect:  make(chan *Client),
		OnDisconnect: make(chan *Client),
		log:		 log,
	}
}


func (h *Hub) Run(){
	//AN Infinite loop that waits for communication on the channels
	for {
		select {
		
		//Case when client is received through the register channel
		case client := <-h.register:
			h.clients[client] = true
			h.log.Info("[WS] client connected", "active", len(h.clients))
			h.OnConnect <- client
		

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.log.Info("[WS] client disconnected", "active", len(h.clients))
				h.OnDisconnect <- client
			}
			
		}
	}
}