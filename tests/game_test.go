package tests

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JoiZs/chess-bk/cachedb"
	"github.com/JoiZs/chess-bk/game"
	"github.com/gofrs/uuid"
	"github.com/joho/godotenv"
)

func TestAll(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Err at loading .env ")
	}
	t.Run("Create and Retrieve Game", func(t *testing.T) {
		t.Parallel()
		cg := game.NewGame()
		log.Println("New Game....")
		ctx := context.Background()

		cache := cachedb.NewRdCache(ctx)
		log.Println("New Cache")

		err := cache.StoreGame(cg.Id, cg.Game)
		if err != nil {
			t.Fatal(err)
		}

		rtG := cache.RetrieveGame(cg.Id)

		if rtG == nil {
			t.Error("Cannot RetrieveGame from redis")
		}

		log.Printf("Game: %v", rtG)
	})
	t.Run("Add 1 Player and wait timeout",
		func(t *testing.T) {
			t.Parallel()
			mq := game.NewMatchMakingQ()

			p1id, err := uuid.NewV1()
			if err != nil {
				fmt.Println("Err at adding player 1")
			}

			p1 := game.NewPlayer(p1id, mq.PlayerSize())

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

			player := game.NewPlayer(playerid, mq.PlayerSize())
			mq.AddPlayer(player)
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
				} else {
					log.Println("Unable to match")
				}
			}(p)
		}

		wg.Wait()

		if pairCount != 0 {
			t.Errorf("Err at multiple players-%v matchmaking(%v)", pairCount, 0)
		}
	})
}
