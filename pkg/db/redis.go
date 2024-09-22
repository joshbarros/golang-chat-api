package repository

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func InitRedisClient(redisHost, redisPort string) *redis.Client {
  rdb := redis.NewClient(&redis.Options{
    Addr: redisHost + ":" + redisPort,
  })

  if _, err := rdb.Ping(context.Background()).Result(); err != nil {
    log.Fatalf("Failed to connect to Redis: %v", err)
  }

  log.Println("Successfully connected to Redis")
  return rdb
}
