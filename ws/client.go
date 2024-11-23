package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
	"github.com/notnil/chess"
)

type Client struct {
	conn          *websocket.Conn
	manager       *Manager
	ingress       chan Event
	Playerprofile *game.Player
	id            uuid.UUID
}

var (
	pongWaitTime = time.Second * 15
	pongInterval = pongWaitTime * 9 / 10
)

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
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWaitTime)); err != nil {
		log.Println(err)
		return
	}

	c.conn.SetPongHandler(c.pongHandler)

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
			time.Sleep(time.Second * 10)
			// break
		}

		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Err handling message route event,", err)
		}
	}
	defer c.BreakConn()
}

func (c *Client) WriteMsg() {
	ticker := time.NewTicker(pongInterval)

	for {
		select {
		case message, ok := <-c.ingress:

			if !ok {
				if err := c.conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Println("connection closed: ", err)
				}
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				log.Println(err)
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Println(err)
			}
			log.Println("Message Sent")

		case <-ticker.C:
			// log.Println("ping")
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Printf("Err at ping: %v", err)
				return
			}

		}
	}
}

func (c *Client) BreakConn() {
	c.manager.mu.Lock()
	defer c.manager.mu.Unlock()
	c.conn.Close()
	delete(c.manager.chessGames, *c.Playerprofile.MatchID)
	delete(c.manager.gameSess, *c.Playerprofile.MatchID)
	delete(c.manager.clients, c)
	delete(c.manager.clientsByID, c.id)
}

func (c *Client) pongHandler(msg string) error {
	// log.Print("pong")
	return c.conn.SetReadDeadline(time.Now().Add(pongWaitTime))
}

func (c *Client) IsValidPlayer() bool {
	player := c.Playerprofile

	gamesess := c.manager.rdClient.RetrieveGame(*player.MatchID)

	_, ok := c.manager.chessGames[*player.MatchID]
	if !ok {
		return false
	} else if gamesess == nil {
		return false
	} else if gamesess.Outcome != chess.NoOutcome {
		return false
	}

	return true
}

func (c *Client) GetOpponent() *game.Player {
	players := c.manager.gameSess[*c.Playerprofile.MatchID]
	var opponent game.Player

	for _, p := range players {
		if p.Client != c.Playerprofile.Client {
			opponent = p
		}
	}

	return &opponent
}
