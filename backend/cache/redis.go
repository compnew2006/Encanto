package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func New(ctx context.Context, redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(options)
	var pingErr error
	for range 30 {
		pingErr = client.Ping(ctx).Err()
		if pingErr == nil {
			return client, nil
		}
		select {
		case <-ctx.Done():
			_ = client.Close()
			return nil, fmt.Errorf("ping redis: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}

	_ = client.Close()
	return nil, fmt.Errorf("ping redis: %w", pingErr)
}
