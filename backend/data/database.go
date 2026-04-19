package data

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	var pingErr error
	for range 30 {
		pingErr = pool.Ping(ctx)
		if pingErr == nil {
			return pool, nil
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("ping database: %w", ctx.Err())
		case <-time.After(time.Second):
		}
	}

	pool.Close()
	return nil, fmt.Errorf("ping database: %w", pingErr)
}
