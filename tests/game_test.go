package tests

import (
	"fmt"
	"testing"

	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
)

func TestMatchMaking(t *testing.T) {
	mq := game.NewMatchMakingQ()

	p1, err := uuid.NewV1()
	if err != nil {
		fmt.Println("Err at adding player 1")
	}

	p2, err := uuid.NewV1()
	if err != nil {
		fmt.Println("Err at adding player 2")
	}

	mq.AddPlayer(p1)
	mq.AddPlayer(p2)

	expSize := 2
	playersize := mq.PlayerSize()

	if playersize != expSize {
		t.Errorf("Error at adding player to matchmaking: %v", playersize)
	}
}
