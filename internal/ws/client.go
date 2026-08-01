package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = pongWait * 9 / 10
)

type MessageHandler func(clientID string, msg ClientMessage)

type Client struct {
	ID    string
	conn  *websocket.Conn
	send  chan []byte
	onMsg MessageHandler
}

func NewClient(conn *websocket.Conn) *Client {
	return &Client{
		conn: conn,
		send: make(chan []byte, 256),
	}
}

// ReadPump reads messages from the WebSocket connection and dispatches them to onMsg.
// Runs in its own goroutine per client. Handles heartbeat pong resets.
func (c *Client) ReadPump(unregister func(*Client)) {
	defer func() {
		unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var envelope ClientMessage
		if err := json.Unmarshal(msg, &envelope); err != nil {
			log.Println("[WS] invalid message format")
			continue
		}

		if c.onMsg != nil {
			c.onMsg(c.ID, envelope)
		}
	}
}

// WritePump drains the send channel and sends messages to the WebSocket connection.
// Also sends periodic ping messages to keep the connection alive.
// Runs in its own goroutine per client.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, msg)

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) SetMessageHandler(h MessageHandler) {
	c.onMsg = h
}

// Send serialises an event and queues it for writing.
// The recover() guard protects against sends to a closed channel (client disconnected).
// Drop-on-full (default case) prevents slow clients from blocking the server.
func (c *Client) Send(eventType ServerEventType, payload any) {
	defer func() {
		if r := recover(); r != nil {
			// Channel was closed — client already disconnected, safe to ignore.
		}
	}()

	msg := EventMessage{
		Type:    eventType,
		Payload: payload,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return
	}

	select {
	case c.send <- b:
	default:
		// Client's send buffer is full; drop the message rather than block.
	}
}