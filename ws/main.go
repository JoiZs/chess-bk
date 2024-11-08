package ws

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/JoiZs/chess-bk/cachedb"
	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

var ErrEventNotSupported = errors.New("this event type is not supported")

var upg = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Manager struct {
	clients     map[*Client]bool
	clientsByID map[uuid.UUID]*Client
	handlers    map[EventType]EventHandler
	matchQ      *game.MatchMakingQ
	gameSess    map[uuid.UUID][2]game.Player
	rdClient    *cachedb.RdCache
	mu          sync.RWMutex
}

func InitManager(ctx context.Context) *Manager {
	mq := game.NewMatchMakingQ()
	rd := cachedb.NewRdCache(ctx)

	m := &Manager{
		clients:     make(map[*Client]bool),
		mu:          sync.RWMutex{},
		handlers:    make(map[EventType]EventHandler),
		matchQ:      mq,
		gameSess:    make(map[uuid.UUID][2]game.Player),
		clientsByID: make(map[uuid.UUID]*Client),
		rdClient:    rd,
	}
	m.setupEventHandlers()

	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[SendMessage] = SendMessageEventHandler
	m.handlers[FindMatch] = FindMatchEventHandler
	m.handlers[RematchReq] = RematchReqEventHandler
	m.handlers[RematchRes] = RematchResEventHandler
}

func (m *Manager) routeEvent(event Event, c *Client) error {
	if handler, ok := m.handlers[event.Type]; ok {
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

	fmt.Printf("Connected: %v\n", conn.RemoteAddr())

	uid, err := uuid.NewV1()
	if err != nil {
		fmt.Printf("Err at creating new ws client, %v", err)
	}
	client := NewClient(m, uid, conn)
	m.clients[client] = true
	m.clientsByID[uid] = client
	go client.ReadMsg()
	go client.WriteMsg()
}
