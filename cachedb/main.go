package cachedb

import (
	"context"
	"log"
	"os"

	"github.com/gofrs/uuid"
	"github.com/redis/go-redis/v9"
)

type RdCache struct {
	rdc *redis.Client
}

func NewRdCache(ctx context.Context) *RdCache {
	opt, err := redis.ParseURL(os.Getenv("REDISURI"))
	if err != nil {
		log.Printf("Err at connecting redis, %v", err)
	}

	newRdC := redis.NewClient(opt)

	testConn := newRdC.Ping(ctx)

	log.Printf("Test redis connection, ", testConn)

	return &RdCache{rdc: newRdC}
}

func (r *RdCache) StoreGame(gid uuid.UUID) error {
	return nil
}

func (r *RdCache) RetrieveGame(gid uuid.UUID) error {
	return nil
}
