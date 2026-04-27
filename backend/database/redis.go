package database

import (
	"context"
	"fmt"
	"time"

	"github.com/irham/topup-backend/config"
	"github.com/redis/go-redis/v9"
)

func NewRedis(ctx context.Context, cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Pass,
		DB:       0,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if _, err := client.Ping(pingCtx).Result(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return client, nil
}
