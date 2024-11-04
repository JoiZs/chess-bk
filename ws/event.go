package ws

import (
	"encoding/json"
	"fmt"
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

func FindMatchEventHandler(event Event, c *Client) error {
	c.manager.matchQ.AddPlayer(c.id)

	count := 30

	for count > 0 {
		fmt.Println("Matching for ", c.id)
		count--
		time.Sleep(time.Second)
	}

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
