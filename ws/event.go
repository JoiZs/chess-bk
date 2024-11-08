package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
)

type EventType int

const (
	SendMessage EventType = iota
	FindMatch
	RematchReq
	RematchRes
	Resign
	MakeMove
	GetMatchInfo
)

type Event struct {
	Payload json.RawMessage `json:"payload"`
	Type    EventType       `json:"type"`
}

type EventHandler func(e Event, c *Client) error

var (
	ErrGameSessionNotFound = errors.New("not found the current game sessin")
	ErrGameSessionFull     = errors.New("game session is already assigned for 2 players")
	ErrGameAreadyFinding   = errors.New("already in finding pool, cannot request...")
)

type ReceivedMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type FindMatchEvent struct {
	From string `json:"from"`
}

type NewMessageEvent struct {
	At time.Time `json:"at"`
	ReceivedMessageEvent
}

type ReceivedRematchEvent struct {
	From     string `json:"from"`
	GameSess string `json:"gamesess"`
	IsGame   bool   `json:"isgame"`
}

type RematchReqEvent struct {
	At time.Time `json:"at"`
	ReceivedRematchEvent
}

func makeMsgEvt(msg string) Event {
	statusMsg := NewMessageEvent{
		At: time.Now(),
		ReceivedMessageEvent: ReceivedMessageEvent{
			Message: msg,
			From:    "Server",
		},
	}

	data, err := json.Marshal(statusMsg)
	if err != nil {
		fmt.Printf("Err at marchshalling data, %v", err)
	}

	msgEvt := Event{
		Payload: data,
		Type:    FindMatch,
	}

	return msgEvt
}

func RematchReqEventHandler(event Event, c *Client) error {
	var ReceivedRematchMsg ReceivedRematchEvent

	err := json.Unmarshal(event.Payload, &ReceivedRematchMsg)
	if err != nil {
		log.Println("Err at unmarshaling rematch received event.")
		return err
	}

	gsuid, err := uuid.FromString(ReceivedRematchMsg.GameSess)
	if err != nil {
		log.Println("Err at parsing rematch uuid.")
		return err

	}

	currGameSession, ok := c.manager.gameSess[gsuid]
	if !ok {
		return ErrGameSessionNotFound
	}

	for _, p := range currGameSession {
		if p.Client == c.id {
			p.Rematch = true
		} else {
			var msgRem RematchReqEvent

			msgRem.At = time.Now()
			msgRem.From = ReceivedRematchMsg.From
			msgRem.GameSess = ReceivedRematchMsg.GameSess
			msgRem.IsGame = ReceivedRematchMsg.IsGame

			data, err := json.Marshal(msgRem)
			if err != nil {
				log.Println("Err at marshaling json rematch event")
				return err
			}

			var evt Event
			evt.Payload = data
			evt.Type = RematchReq

			otherClient, ok := c.manager.clientsByID[p.Client]
			if !ok {
				log.Printf("Other client err.. %v ---- %v", p.Client, otherClient)
			}

			otherClient.ingress <- evt
		}
	}

	return nil
}

func RematchResEventHandler(event Event, c *Client) error {
	var ReceivedRematchMsg ReceivedRematchEvent

	err := json.Unmarshal(event.Payload, &ReceivedRematchMsg)
	if err != nil {
		log.Println("Err at unmarshaling rematch received event.")
		return err
	}

	gsuid, err := uuid.FromString(ReceivedRematchMsg.GameSess)
	if err != nil {
		log.Println("Err at parsing rematch uuid.")
		return err

	}

	_, ok := c.manager.gameSess[gsuid]
	if !ok {
		return ErrGameSessionNotFound
	}

	return nil
}

func FindMatchEventHandler(event Event, c *Client) error {
	log.Println("called find....")

	var p *game.Player

	if c.Playerprofile != nil {
		p = c.Playerprofile
	} else {
		p = game.NewPlayer(c.id, c.manager.matchQ.PlayerSize())
		c.Playerprofile = p
	}

	c.manager.matchQ.AddPlayer(p)

	go c.manager.matchQ.MatchingPlayers()

	currMatch := p.WaitGame()

	if currMatch != nil {
		msgEvt := makeMsgEvt(fmt.Sprintf("Match found: %v", currMatch.Id))
		c.ingress <- msgEvt

		c.manager.rdClient.StoreGame(currMatch.Id, currMatch.Game)

		c.manager.mu.Lock()
		// Add game session to manager
		currPlayers, ok := c.manager.gameSess[currMatch.Id]
		if !ok {
			var newPlayers [2]game.Player
			newPlayers[0] = *p
			c.manager.gameSess[currMatch.Id] = newPlayers
		} else {
			currPlayers[1] = *p
			c.manager.gameSess[currMatch.Id] = currPlayers
		}
		c.manager.mu.Unlock()

		log.Printf("game session stored(%v) - %v players: ", c.manager.gameSess, currPlayers)

		return nil
	}

	c.manager.matchQ.RemoveTimeoutPlayers()

	log.Printf("Curr pool: %v", c.manager.matchQ.PlayerSize())

	msgEvt := makeMsgEvt("Match Not Found")
	c.ingress <- msgEvt
	return nil
}

func SendMessageEventHandler(event Event, c *Client) error {
	var tempPayload ReceivedMessageEvent

	if err := json.Unmarshal(event.Payload, &tempPayload); err != nil {
		fmt.Printf("Err at event json unmarshal, %v", err)
	}

	var tempBcMsg NewMessageEvent

	tempBcMsg.Message = tempPayload.Message
	tempBcMsg.From = tempPayload.From
	tempBcMsg.At = time.Now()

	data, err := json.Marshal(tempBcMsg)
	if err != nil {
		return fmt.Errorf("err at marshaling data, %v", err)
	}

	var BroadcastEvt Event

	BroadcastEvt.Payload = data
	BroadcastEvt.Type = SendMessage

	for el := range c.manager.clients {
		el.ingress <- BroadcastEvt
	}

	return nil
}
