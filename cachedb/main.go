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
	Turn    chess.Color          `json:"turn" redis:"turn"`
	Outcome chess.Outcome        `json:"outcome" redis:"outcome"`
	Fen     string               `json:"fen" redis:"fen"`
}

func NewRdCache(ctx context.Context) *RdCache {
	opt, err := redis.ParseURL(os.Getenv("REDISURI"))
	if err != nil {
		log.Printf("Err at connecting redis, %v", err)
	}

	newRdC := redis.NewClient(opt)

	testConn := newRdC.Ping(ctx)

	log.Printf("Test redis connection, %v", testConn)

	return &RdCache{rdc: newRdC, ctx: ctx}
}

func CreateCacheGame(gg chess.Game) *CacheGame {
	cacheg := CacheGame{
		Moves:   make([]*chess.MoveHistory, 0),
		Fen:     gg.FEN(),
		Turn:    gg.Position().Turn(),
		Outcome: gg.Outcome(),
	}

	return &cacheg
}

func (r *RdCache) StoreGame(gid uuid.UUID, gcg chess.Game) error {
	cacheg := CacheGame{
		Moves:   gcg.MoveHistory(),
		Fen:     gcg.FEN(),
		Turn:    gcg.Position().Turn(),
		Outcome: gcg.Outcome(),
	}
	data, err := json.Marshal(cacheg)
	if err != nil {
		log.Printf("Err at marshaling chessGame-JSON, %v", err)
	}

	err = r.rdc.Set(r.ctx, gid.String(), data, time.Hour*2).Err()
	if err != nil {
		log.Printf("Err at adding game to redis: %v", gid.String())
		return err
	}

	log.Printf("Added game to redis: %v", gid.String())

	return nil
}

func (r *RdCache) RetrieveGame(gid uuid.UUID) *CacheGame {
	data, err := r.rdc.Get(r.ctx, gid.String()).Result()
	if err != nil {
		log.Printf("Err at RedisGame Scan, %v", err)
	}
	var crrG CacheGame
	err = json.Unmarshal([]byte(data), &crrG)
	if err != nil {
		log.Printf("Err at Unmarshalling game data, %v", err)
	}

	log.Printf("Get game (%v) from redis: %v", gid, crrG)

	return &crrG
}
