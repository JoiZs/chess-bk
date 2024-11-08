package ws

import (
	"encoding/json"
	"log"

	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	conn          *websocket.Conn
	manager       *Manager
	ingress       chan Event
	Playerprofile *game.Player
	id            uuid.UUID
}

func NewClient(m *Manager, id uuid.UUID, c *websocket.Conn) *Client {
	return &Client{
		id:            id,
		conn:          c,
		manager:       m,
		ingress:       make(chan Event),
		Playerprofile: nil,
	}
}

func (c *Client) ReadMsg() {
	for {
		_, pl, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure, websocket.CloseMessage, websocket.CloseGoingAway) {
				break
			}
		}

		var request Event

		if err := json.Unmarshal(pl, &request); err != nil {
			log.Printf("Err at requst event json Unmarshal, %v", err)
			break
		}
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Err handling message route event,", err)
		}
	}
	defer c.BreakConn()
}

func (c *Client) WriteMsg() {
	for message := range c.ingress {
		data, err := json.Marshal(message)
		if err != nil {
			log.Println(err)
			return
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println(err)
		}
		log.Println("Message Sent")
	}
}

func (c *Client) BreakConn() {
	c.manager.mu.Lock()
	defer c.manager.mu.Unlock()
	c.conn.Close()
	delete(c.manager.clients, c)
	delete(c.manager.clientsByID, c.id)
}
