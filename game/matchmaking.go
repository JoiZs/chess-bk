package game

import (
	"container/heap"
	"log"
	"sync"
	"time"

	"github.com/gofrs/uuid"
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

func (mq *MatchMakingQ) AddPlayer(c uuid.UUID) *Player {
	player := &Player{
		priority: time.Now(),
		client:   c,
		index:    mq.pq.Len(),
		match:    make(chan *ChessGame),
	}
	mq.mu.Lock()
	defer mq.mu.Unlock()
	heap.Push(&mq.pq, player)

	log.Printf("Added a player: %v\n", c)

	return player
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

		p1.match <- nGame
		p2.match <- nGame

		log.Printf("Paired: %v & %v\n", p1.client, p2.client)
	}
}

func (mq *MatchMakingQ) RemoveTimeoutPlayers() {
	timeOutDuration := time.Second * 2
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for mq.pq.Len() > 0 {
		player := mq.pq.Peek()
		if time.Since(player.priority) >= timeOutDuration {
			heap.Pop(&mq.pq)
			log.Printf("Removed, %v\n", player.client)
		} else {
			break
		}
	}
}
