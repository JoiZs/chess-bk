package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/JoiZs/chess-bk/cachedb"
	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
)

type EventType int

const (
	SendMessage EventType = iota
	FindMatch
	MatchInfo
	RematchReq
	RematchRes
	Resign
	MakeMove
)

type Event struct {
	Payload json.RawMessage `json:"payload"`
	Type    EventType       `json:"type"`
}

type EventHandler func(e Event, c *Client) error

var (
	ErrGameSessionNotFound = errors.New("not found the current game sessin")
	ErrGameSessionFull     = errors.New("game session is already assigned for 2 players")
	ErrGameAreadyFinding   = errors.New("already in finding pool, cannot request")
	ErrWaitYourTurn        = errors.New("wait for your turn")
	ErrInvalidMove         = errors.New("invalid move")
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

type NewGameInfoEvent struct {
	cachedb.CacheGame
	At time.Time `json:"at"`
}

type ReceivedMoveEvent struct {
	ReceivedMatchEvent
	Move string `json:"move"`
}

type ReceivedMatchEvent struct {
	From     string `json:"from"`
	GameSess string `json:"gamesess"`
}

type RematchReqEvent struct {
	At time.Time `json:"at"`
	ReceivedMatchEvent
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
	var ReceivedRematchMsg ReceivedMatchEvent

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
	var ReceivedRematchMsg ReceivedMatchEvent

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
		msgEvt := makeMsgEvt(fmt.Sprintf("Match found: %v, Color: %v ", currMatch.Id, c.Playerprofile.Color))

		c.ingress <- msgEvt
		c.manager.mu.Lock()
		c.manager.chessGames[currMatch.Id] = currMatch.Game
		c.manager.rdClient.StoreGame(currMatch.Id, *currMatch.Game)

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
		log.Printf("game session stored(%v) -  players: %v", currMatch.Id, currPlayers)

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
		return fmt.Errorf("err at event json unmarshal, %v", err)
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

func GetMatchEventHandler(event Event, c *Client) error {
	if !c.IsValidPlayer() {
		return ErrGameSessionNotFound
	}

	var tempPayload ReceivedMoveEvent

	err := json.Unmarshal(event.Payload, &tempPayload)
	if err != nil {
		return fmt.Errorf("err at event json unmarshal, %v", err)
	}

	var tempBcMsg NewGameInfoEvent

	cgame := c.manager.rdClient.RetrieveGame(*c.Playerprofile.MatchID)

	tempBcMsg.CacheGame = *cgame
	tempBcMsg.At = time.Now()

	data, err := json.Marshal(tempBcMsg)
	if err != nil {
		return fmt.Errorf("err at marshaling data, %v", err)
	}

	var BroadcastEvt Event

	BroadcastEvt.Payload = data
	BroadcastEvt.Type = MatchInfo

	c.ingress <- BroadcastEvt
	return nil
}

func ResignEventHandler(event Event, c *Client) error {
	if !c.IsValidPlayer() {
		return ErrGameSessionNotFound
	}

	var tempPayload ReceivedMatchEvent

	if err := json.Unmarshal(event.Payload, &tempPayload); err != nil {
		return fmt.Errorf("err at event json unmarshal, %v", err)
	}

	currGame, ok := c.manager.chessGames[*c.Playerprofile.MatchID]

	if !ok {
		return ErrGameSessionNotFound
	}

	currGame.Resign(c.Playerprofile.Color)
	c.manager.rdClient.StoreGame(*c.Playerprofile.MatchID, *currGame)

	msgEvt := makeMsgEvt(fmt.Sprintf("Resigned the match: %v", c.Playerprofile.Match))

	c.ingress <- msgEvt
	return nil
}

func MakeMoveHandler(event Event, c *Client) error {
	if !c.IsValidPlayer() {
		return ErrGameSessionNotFound
	}

	var tempPayload ReceivedMoveEvent

	err := json.Unmarshal(event.Payload, &tempPayload)
	if err != nil {
		return fmt.Errorf("err at event json unmarshal, %v", err)
	}

	currGame, ok := c.manager.chessGames[*c.Playerprofile.MatchID]

	if !ok {
		return ErrGameSessionNotFound
	}

	if c.Playerprofile.Color != currGame.Position().Turn() {
		msgEvt := makeMsgEvt(fmt.Sprintf("Wait your turn: %v", tempPayload.Move))
		c.ingress <- msgEvt
		return ErrWaitYourTurn
	}

	err = currGame.MoveStr(tempPayload.Move)
	if err != nil {
		msgEvt := makeMsgEvt(fmt.Sprintf("Invalid Move: %v", tempPayload.Move))
		c.ingress <- msgEvt

		return ErrInvalidMove
	}

	// tempCG := cachedb.CreateCacheGame(*currGame)

	var tempBcMsg NewGameInfoEvent

	var moveStrs []string
	for _, move := range currGame.Moves() {
		moveStrs = append(moveStrs, move.String()) // Use the String() method for each move
	}

	tempBcMsg.CacheGame = cachedb.CacheGame{
		Moves:   moveStrs,
		Turn:    currGame.Position().Turn(),
		Outcome: currGame.Outcome(),
		Fen:     currGame.FEN(),
	}
	tempBcMsg.At = time.Now()

	data, err := json.Marshal(tempBcMsg)
	log.Println(tempBcMsg.CacheGame)
	log.Println(string(data))
	if err != nil {
		return fmt.Errorf("err at marshaling data, %v", err)
	}

	var BroadcastEvt Event

	BroadcastEvt.Payload = data
	BroadcastEvt.Type = MatchInfo

	for _, p := range c.manager.gameSess[*c.Playerprofile.MatchID] {
		cl, ok := c.manager.clientsByID[p.Client]
		if !ok {
			return fmt.Errorf("no client found, %v", err)
		}
		cl.ingress <- BroadcastEvt
	}

	return nil
}
