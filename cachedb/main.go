package cachedb

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/notnil/chess"
	"github.com/redis/go-redis/v9"
)

type RdCache struct {
	rdc *redis.Client
	ctx context.Context
}

type CacheGame struct {
	Moves   []*chess.MoveHistory `json:"moves" redis:"moves"`
	Fen     string               `json:"fen" redis:"fen"`
	CurrPos chess.Position       `json:"currpos" redis:"currpos"`
	Outcome chess.Outcome        `json:"outcome" redis:"outcome"`
}

func NewRdCache(ctx context.Context) *RdCache {
	opt, err := redis.ParseURL(os.Getenv("REDISURI"))
	if err != nil {
		log.Printf("Err at connecting redis, %v", err)
	}

	newRdC := redis.NewClient(opt)

	testConn := newRdC.Ping(ctx)

	log.Printf("Test redis connection, ", testConn)

	return &RdCache{rdc: newRdC, ctx: ctx}
}

func (r *RdCache) StoreGame(gid uuid.UUID, gcg chess.Game) error {
	cacheg := CacheGame{
		Moves:   gcg.MoveHistory(),
		Fen:     gcg.FEN(),
		CurrPos: *gcg.Position(),
		Outcome: gcg.Outcome(),
	}
	data, err := json.Marshal(cacheg)
	if err != nil {
		log.Printf("Err at marshaling chessGame-JSON, %v", err)
	}

	status := r.rdc.Set(r.ctx, gid.String(), data, time.Hour*2)

	log.Printf("Added game to redis: %v", status)

	return nil
}

func (r *RdCache) RetrieveGame(gid uuid.UUID) *CacheGame {
	var crrG CacheGame
	r.rdc.Get(r.ctx, gid.String()).Scan(crrG)

	log.Printf("Get game from redis: ", crrG)

	return &crrG
}
