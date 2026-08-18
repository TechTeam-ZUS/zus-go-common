package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	redis "github.com/redis/go-redis/v9"
)

type CacheInstance struct {
	Client *redis.Client
}

// Init creates a cache client using the provided config. Uses the go-redis
// driver, which speaks the Redis protocol and is compatible with Redis and
// Valkey servers alike.
func Init() (*CacheInstance, error) {
	cfg := config.LoadCache()

	client := redis.NewClient(&redis.Options{
		Addr:     addr(cfg),
		Password: cfg.Password,
		Username: cfg.Username,
	})

	client.AddHook(NewHook(cfg.Prefix))

	//ping timeout for 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping cache: %w", err)
	}

	return &CacheInstance{client}, nil
}

func (c CacheInstance) Close() error {
	return c.Client.Close()
}

func addr(cfg config.CacheConfig) string {
	return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}
