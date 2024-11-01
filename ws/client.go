package ws

import (
	"fmt"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	id      uuid.UUID
	conn    *websocket.Conn
	manager *Manager
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

		fmt.Println(string(pl))
		c.WriteMsg(pl)
	}
	defer c.BreakConn()
}

func (c *Client) WriteMsg(data []byte) {
	for client := range c.manager.clients {
		err := client.conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			fmt.Printf("Err at writing messages, %v", err)
		}
	}
}

func (c *Client) BreakConn() {
	c.manager.mu.Lock()
	c.conn.Close()
	delete(c.manager.clients, c)

	defer c.manager.mu.Unlock()
}
