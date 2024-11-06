package tests

import (
	"fmt"
	"log"
	"os"
	"sync"
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
		logFile, err := os.OpenFile("test.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			t.Fatalf("failed to open log file: %v", err)
		}
		defer logFile.Close()

		log.SetOutput(logFile)

		t.Parallel()
		mq := game.NewMatchMakingQ()
		pairCount := 1000

		wg := sync.WaitGroup{}

		players := make(map[*game.Player]bool)

		for i := 1; i <= pairCount; i++ {
			playerid, err := uuid.NewV1()
			if err != nil {
				fmt.Printf("Err at adding player %v, %v\n", i, err)
			}
			player := mq.AddPlayer(playerid)

			players[player] = true
		}

		go mq.MatchingPlayers()

		for p := range players {
			wg.Add(1)
			go func(pl *game.Player) {
				defer wg.Done()
				gid := pl.WaitGame()
				if gid != nil {
					pairCount--
				}
			}(p)
		}

		wg.Wait()

		if pairCount != 0 {
			t.Errorf("Err at multiple players-%v matchmaking(%v)", pairCount, 0)
		}
	})
}
