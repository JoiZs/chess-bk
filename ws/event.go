package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type EventType int

const (
	SendMessage EventType = iota
	FindMatch
	MakeMove
	GetMatchInfo
)

type Event struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(e Event, c *Client) error

type ReceivedMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type FindMatchEvent struct {
	From string `json:"from"`
}

type NewMessageEvent struct {
	ReceivedMessageEvent
	At time.Time `json:"at"`
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

func FindMatchEventHandler(event Event, c *Client) error {
	p := c.manager.matchQ.AddPlayer(c.id)
	log.Printf("Added a player, %v \n", c.id)

	go c.manager.matchQ.MatchingPlayers()

	matchid := p.WaitGame()

	if matchid != nil {
		msgEvt := makeMsgEvt(fmt.Sprintf("Match found: %v", matchid))

		c.ingress <- msgEvt
		return nil
	}

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
