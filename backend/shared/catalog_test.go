package shared

import "testing"

func TestLoadPermissionCatalog(t *testing.T) {
	catalog, err := LoadPermissionCatalog()
	if err != nil {
		t.Fatalf("LoadPermissionCatalog() returned error: %v", err)
	}

	if len(catalog.Permissions) == 0 {
		t.Fatal("LoadPermissionCatalog() returned an empty catalog")
	}
}