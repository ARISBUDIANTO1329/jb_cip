package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jaybani/jb_cip/config"
)

var client *redis.Client

func Init(cfg *config.RedisConfig) (*redis.Client, error) {
	client = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}

func GetClient() (*redis.Client, error) {
	if client == nil {
		return nil, fmt.Errorf("redis not initialized")
	}
	return client, nil
}

func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}
