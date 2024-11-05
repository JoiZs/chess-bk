package game

import (
	"fmt"
	"log"
	"time"

	"github.com/gofrs/uuid"
	"github.com/notnil/chess"
)

type Player struct {
	priority time.Time
	client   uuid.UUID
	index    int
	match    chan *ChessGame
}

type ChessGame struct {
	id   uuid.UUID
	Game chess.Game
}

func NewGame() *ChessGame {
	gid, err := uuid.NewV1()
	if err != nil {
		fmt.Printf("Err at creating chess game id: %v", err)
	}

	return &ChessGame{
		id:   gid,
		Game: *chess.NewGame(),
	}
}

func (p *Player) WaitGame() *uuid.UUID {
	waitTime := time.NewTimer(time.Second * 10)

	select {
	case gid := <-p.match:
		log.Println("Matched..")
		return &gid.id
	case <-waitTime.C:
		log.Println("No Match found..")
		return nil
	}
}
