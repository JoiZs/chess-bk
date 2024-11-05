package ws

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	id      uuid.UUID
	conn    *websocket.Conn
	manager *Manager
	ingress chan Event
}

func NewClient(m *Manager, c *websocket.Conn) *Client {
	uid, err := uuid.NewV1()
	if err != nil {
		fmt.Printf("Err at creating new ws client, %v", err)
	}

	return &Client{
		id:      uid,
		conn:    c,
		manager: m,
		ingress: make(chan Event),
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
		fmt.Println(request.Type)

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
		// if !ok {
		// 	if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
		// 		log.Println("Connection closed.")
		// 	}
		// 	return
		// }

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
}
