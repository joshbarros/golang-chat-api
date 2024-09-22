package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

type RedisClientInterface interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
}
