package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = pongWait * 9 / 10
)

type MessageHandler func(clientID string, msg ClientMessage)

type Client struct {
	ID string 
	conn *websocket.Conn
	send chan []byte
	onMsg  MessageHandler
}

type IncomingMessage struct {
	Type string `json:"type"`
	Word string `json:"word"`
}

func NewClient(conn *websocket.Conn) *Client{
	return &Client{
		conn: conn,
		send: make(chan []byte, 256), 
	}
}

//Function responsible for reading messages from the Client
func (c *Client) ReadPump(unregister func(*Client)) {
	//Defer func to ensure if the loop breaks; the client is unregistered from the sytem
	defer func() {
		unregister(c)
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	//HeartBeat(Pong) if no response from client within pongWait(60 seconds) connection is considered dead
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
		if err := json.Unmarshal(msg, &envelope); err != nil{
			log.Println("[WS] invalid message format")
			continue
		}

		if c.onMsg != nil {
			c.onMsg(c.ID, envelope)
		}

		// c.handleMessage(m)
	}
}


//Function Responsible for sending messages and heartbeats to the Client
func (c *Client) WritePump(){
	
	ticker := time.NewTicker(pingPeriod)
	defer func(){
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select{
		//Case when message arrives in the Send channel
		case msg, ok := <-c.send:
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok{
					c.conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				c.conn.WriteMessage(websocket.TextMessage, msg)
		
		//Case when the Ticker fires; it sends a PingMessage
		case <-ticker.C:
				c.conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return 
				}
		}
	}
}

func (c *Client) SetMessageHandler(h MessageHandler){
	c.onMsg = h
}

func (c *Client) Send(eventType ServerEventType, payload any) {
	defer func() {
		if r := recover(); r != nil {
			//Ignore
		}
	}()
	
	msg := EventMessage {
		Type: eventType,
		Payload: payload,
	}

	b, err := json.Marshal(msg)
	if err != nil {
		return 
	}

	select {
	case c.send <- b:
	default:
	}
}