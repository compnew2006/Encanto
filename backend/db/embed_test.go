package db

import (
	"testing"
)

func TestMigrationsFSContainsSQLFiles(t *testing.T) {
	entries, err := MigrationsFS.ReadDir("migrations")
	if err != nil {
		t.Fatalf("ReadDir returned error: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected embedded migrations to contain at least one file")
	}

	foundSQL := false
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 4 && entry.Name()[len(entry.Name())-4:] == ".sql" {
			foundSQL = true
			break
		}
	}
	if !foundSQL {
		t.Fatal("expected at least one embedded .sql migration file")
	}
}
