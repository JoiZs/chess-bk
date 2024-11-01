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
