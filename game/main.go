package game

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/notnil/chess"
)

type Player struct {
	priority time.Time
	client   uuid.UUID
	index    int
}

type ChessGame struct {
	Players [2]Player
	Game    chess.Game
}

func NewGame(p1 Player, p2 Player) *ChessGame {
	var players [2]Player

	players[0] = p1
	players[1] = p2

	return &ChessGame{
		Players: players,
		Game:    *chess.NewGame(),
	}
}
