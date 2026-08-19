package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	"github.com/TechTeam-ZUS/zus-go-common/logger"
	"github.com/TechTeam-ZUS/zus-go-common/retry"
	redis "github.com/redis/go-redis/v9"
)

type CacheInstance struct {
	Client *redis.Client
}

// Init creates a cache client using the provided config. Uses the go-redis
// driver, which speaks the Redis protocol and is compatible with Redis and
// Valkey servers alike. Retries cfg.RetryCount times before giving up;
// exhausting retries is fatal.
func Init() (*CacheInstance, error) {
	cfg := config.LoadCache()

	var client *redis.Client
	err := retry.Do(cfg.RetryCount, retry.RetryDelay, func() error {
		client = redis.NewClient(&redis.Options{
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
			return err
		}
		return nil
	}, "Cache Connection")

	if err != nil {
		logger.Fatal("Failed to connect cache server", "error", err.Error())
	}

	return &CacheInstance{client}, nil
}

func (c *CacheInstance) Close() error {
	return c.Client.Close()
}

// Get returns the value stored at key.
func (c *CacheInstance) Get(ctx context.Context, key string) ([]byte, error) {
	return c.Client.Get(ctx, key).Bytes()
}

// Set stores value at key with the given expiration (0 for no expiration).
func (c *CacheInstance) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.Client.Set(ctx, key, value, expiration).Err()
}

// Ping checks connectivity to the cache server.
func (c *CacheInstance) Ping(ctx context.Context) error {
	return c.Client.Ping(ctx).Err()
}

// Do executes an arbitrary cache command, e.g. Do(ctx, "expire", "key", 60).
// Use this for commands not covered by Get/Set until a dedicated method exists.
func (c *CacheInstance) Do(ctx context.Context, args ...any) *redis.Cmd {
	return c.Client.Do(ctx, args...)
}

func addr(cfg config.CacheConfig) string {
	return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}
