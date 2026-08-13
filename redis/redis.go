package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	"github.com/redis/go-redis/v9"
)

type RedisInstance struct {
	Client *redis.Client
	prefix string
}

// SetupRedisConnection creates a Redis client using the provided config
func Init() (*RedisInstance, error) {
	cfg := config.LoadRedis()

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		Username: cfg.Username,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisInstance{client, cfg.Prefix}, nil
}

func (r RedisInstance) Close() error {
	return r.Client.Close()
}
