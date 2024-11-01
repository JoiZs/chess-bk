package ws

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upg = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Manager struct {
	clients map[*Client]bool
	mu      sync.RWMutex
}

func InitManager() *Manager {
	return &Manager{
		clients: make(map[*Client]bool),
		mu:      sync.RWMutex{},
	}
}

func (m *Manager) WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upg.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Err at upgrading ws conn, %v", err)
	}

	fmt.Printf("Connected: %v", conn.RemoteAddr())

	client := NewClient(m, conn)

	m.clients[client] = true
	client.ReadMsg()
}
