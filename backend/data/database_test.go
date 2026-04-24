package data

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestConnectSignatureCompiles(t *testing.T) {
	var _ func(context.Context, string) (*pgxpool.Pool, error) = Connect
}