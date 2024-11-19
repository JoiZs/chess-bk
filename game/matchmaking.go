package game

import (
	"container/heap"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/notnil/chess"
)

type PlayerPoolQ []*Player

type MatchMakingQ struct {
	pq PlayerPoolQ
	mu sync.RWMutex
}

func (pq PlayerPoolQ) Len() int {
	return len(pq)
}

func (pq PlayerPoolQ) Less(i, j int) bool {
	return pq[i].priority.Before(pq[j].priority)
}

func (pq PlayerPoolQ) Peek() Player {
	return *pq[0]
}

func (pq PlayerPoolQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PlayerPoolQ) Push(x interface{}) {
	n := pq.Len()
	player := x.(*Player)
	player.index = n
	*pq = append(*pq, player)
}

func (pq *PlayerPoolQ) Pop() interface{} {
	oldPq := *pq
	n := pq.Len()
	player := oldPq[n-1]

	player.index = -1

	oldPq[n-1] = nil
	*pq = oldPq[0 : n-1]
	return player
}

func NewMatchMakingQ() *MatchMakingQ {
	pq := &PlayerPoolQ{}

	heap.Init(pq)

	return &MatchMakingQ{
		pq: *pq,
		mu: sync.RWMutex{},
	}
}

func (mq *MatchMakingQ) AddPlayer(p *Player) error {
	log.Println("add player called...")
	mq.mu.Lock()
	defer mq.mu.Unlock()
	heap.Push(&mq.pq, p)

	log.Printf("Added a player: %v\n", p.Client)

	return nil
}

func (mq *MatchMakingQ) PlayerSize() int {
	return mq.pq.Len()
}

func (mq *MatchMakingQ) MatchingPlayers() {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for mq.pq.Len() >= 2 {
		p1 := heap.Pop(&mq.pq).(*Player)
		p2 := heap.Pop(&mq.pq).(*Player)

		nGame := NewGame()

		firstColor := getRandomChessColor()

		p1.Match <- nGame
		p1.Color <- firstColor
		p2.Match <- nGame
		p2.Color <- firstColor.Other()

		log.Printf("Paired: %v & %v\n", p1.Client, p2.Client)
	}

	log.Println("Remaining: ", mq.pq.Len())
}

func (mq *MatchMakingQ) RemoveTimeoutPlayers() {
	timeOutDuration := time.Second * 2
	mq.mu.Lock()
	defer mq.mu.Unlock()

	log.Println("Timeout Called...")

	for mq.pq.Len() > 0 {
		player := mq.pq.Peek()
		if time.Since(player.priority) >= timeOutDuration {
			heap.Pop(&mq.pq)
			log.Printf("Removed, %v\n", player.Client)
		} else {
			return
		}
	}
}

func getRandomChessColor() chess.Color {
	colors := [2]chess.Color{chess.White, chess.Black}
	return colors[rand.Intn(2)]
}
