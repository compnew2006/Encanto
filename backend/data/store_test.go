package data

import "testing"

func TestNewStoreInitializesFields(t *testing.T) {
	store := NewStore(nil)
	if store == nil {
		t.Fatal("expected store")
	}
	if store.Pool != nil {
		t.Fatalf("expected nil pool, got %#v", store.Pool)
	}
	if store.Queries == nil {
		t.Fatal("expected queries to be initialized")
	}
}