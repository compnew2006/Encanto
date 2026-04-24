package cache

import (
	"context"
	"strings"
	"testing"
)

func TestNewRejectsInvalidRedisURL(t *testing.T) {
	client, err := New(context.Background(), "://bad-url")
	if err == nil {
		t.Fatal("New() error = nil, want non-nil")
	}
	if client != nil {
		t.Fatalf("New() client = %v, want nil", client)
	}
	if !strings.Contains(err.Error(), "parse redis url") {
		t.Fatalf("New() error = %q, want parse redis url error", err)
	}
}