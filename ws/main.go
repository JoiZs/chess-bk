package ws

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var ErrEventNotSupported = errors.New("this event type is not supported")

var upg = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Manager struct {
	clients  map[*Client]bool
	mu       sync.RWMutex
	handlers map[EventType]EventHandler
}

func InitManager() *Manager {
	m := &Manager{
		clients:  make(map[*Client]bool),
		mu:       sync.RWMutex{},
		handlers: make(map[EventType]EventHandler),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[SendMessage] = SendMessageEventHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	// Check if Handler is present in Map
	if handler, ok := m.handlers[event.Type]; ok {
		// Execute the handler and return any err
		if err := handler(event, c); err != nil {
			return err
		}
		return nil
	} else {
		return ErrEventNotSupported
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
	go client.ReadMsg()
	go client.WriteMsg()
}
