package common

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SetupRedisConnection creates a Redis client using the provided config
// and verifies connectivity with a ping.
func SetupRedisConnection(cfg RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// SetupRedisConnectionFromEnv loads Redis settings from environment variables
// and creates a client.
func SetupRedisConnectionFromEnv() (*redis.Client, error) {
	return SetupRedisConnection(RedisConfigFromEnv())
}
