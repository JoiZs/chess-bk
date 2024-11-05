package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
)

func TestAll(t *testing.T) {
	t.Run("Add 1 Player and wait timeout",
		func(t *testing.T) {
			t.Parallel()
			mq := game.NewMatchMakingQ()

			p1, err := uuid.NewV1()
			if err != nil {
				fmt.Println("Err at adding player 1")
			}

			mq.AddPlayer(p1)

			time.Sleep(time.Second * 3)

			mq.RemoveTimeoutPlayers()
			expSize := 0
			playersize := mq.PlayerSize()

			if playersize != expSize {
				t.Errorf("Error at adding player to matchmaking: %v", playersize)
			}
		})

	t.Run("Multiple Players and pair them", func(t *testing.T) {
		t.Parallel()
		mq := game.NewMatchMakingQ()
		pairCount := 0

		for i := 1; i <= 50; i++ {

			player, err := uuid.NewV1()
			if err != nil {
				fmt.Printf("Err at adding player %v, %v\n", i, err)
			}
			mq.AddPlayer(player)
		}

		for mq.PlayerSize() >= 2 {
			p1, p2 := mq.MatchingPlayers()
			if p1 != nil && p2 != nil {
				pairCount++
			}
		}

		if pairCount != 25 {
			t.Errorf("Err at multiple players-%v matchmaking(%v)", pairCount, 25)
		}
	})
}
