package sqlc

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type fakeDB struct{}

func (fakeDB) Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (fakeDB) Query(context.Context, string, ...interface{}) (pgx.Rows, error) {
	return nil, nil
}

func (fakeDB) QueryRow(context.Context, string, ...interface{}) pgx.Row {
	return nil
}

func TestNewAndWithTxCompile(t *testing.T) {
	queries := New(fakeDB{})
	if queries == nil {
		t.Fatal("expected queries")
	}

	var tx pgx.Tx
	if got := queries.WithTx(tx); got == nil {
		t.Fatal("expected WithTx to return queries")
	}
}

func TestDBTXInterfaceSatisfied(t *testing.T) {
	var _ DBTX = fakeDB{}
}
