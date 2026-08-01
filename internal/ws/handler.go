package ws

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// Allow all origins for development. Restrict this in production.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	hub.log.Info("[WS] upgrade request", "remote", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		hub.log.Error("[WS] upgrade failed", "err", err)
		return
	}

	client := NewClient(conn)
	client.ID = uuid.NewString()
	hub.register <- client

	go client.WritePump()
	go client.ReadPump(func(c *Client) {
		hub.unregister <- c
	})
}
