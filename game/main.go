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
	Client   uuid.UUID
	index    int
	Match    chan *ChessGame
	Rematch  bool
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
	waitTime := time.NewTimer(time.Second * 2)

	select {
	case gid := <-p.Match:
		log.Println("Matched..")
		return &gid.id
	case <-waitTime.C:
		log.Println("No Match found..")
		return nil
	}
}

func NewPlayer(cid uuid.UUID, idx int) *Player {
	player := &Player{
		priority: time.Now(),
		Client:   cid,
		index:    idx,
		Match:    make(chan *ChessGame),
		Rematch:  false,
	}

	log.Println("Created a new Player.")

	return player
}
