package core

import (
	"slices"
	"testing"
)

func TestNewPermissionCatalogLoadsSharedCatalog(t *testing.T) {
	t.Parallel()

	catalog, err := NewPermissionCatalog()
	if err != nil {
		t.Fatalf("NewPermissionCatalog() error = %v", err)
	}

	if !catalog.Has("messages.send") {
		t.Fatalf("catalog should include messages.send")
	}
	if !catalog.Has("roles.manage") {
		t.Fatalf("catalog should include roles.manage")
	}

	keys := catalog.Keys()
	if len(keys) == 0 {
		t.Fatal("catalog keys should not be empty")
	}
	if !slices.IsSorted(keys) {
		t.Fatalf("catalog keys should be sorted, got %v", keys)
	}
}
