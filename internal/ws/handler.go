package ws
import(
	"net/http"
	 "log"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {
	hub.log.Info("[WS] websocket upgrade request", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
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


func (c *Client) handleMessage(m IncomingMessage) {
	log.Println("[WS] message received:", m.Type, m.Word)
}


