package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// New connects to Redis and returns a client
func New(ctx context.Context, url string) (*goredis.Client, error) {
	opts, err := goredis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	client := goredis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
