package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	redis "github.com/redis/go-redis/v9"
)

type RedisInstance struct {
	Client *redis.Client
}

// SetupRedisConnection creates a Redis client using the provided config
func Init() (*RedisInstance, error) {
	cfg := config.LoadRedis()

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
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &RedisInstance{client}, nil
}

func (r RedisInstance) Close() error {
	return r.Client.Close()
}

func addr(cfg config.RedisConfig) string {
	return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}
