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
	Match    chan *ChessGame
	index    int
	Client   uuid.UUID
	Rematch  bool
}

type ChessGame struct {
	Game chess.Game
	Id   uuid.UUID
}

func NewGame() *ChessGame {
	gid, err := uuid.NewV1()
	if err != nil {
		fmt.Printf("Err at creating chess game id: %v", err)
	}

	return &ChessGame{
		Id:   gid,
		Game: *chess.NewGame(),
	}
}

func (p *Player) WaitGame() *ChessGame {
	waitTime := time.NewTimer(time.Second * 2)

	select {
	case gid := <-p.Match:
		log.Println("Matched..")
		return gid
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
