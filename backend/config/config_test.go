package config

import (
	"os"
	"testing"
	"time"
)

func setTestEnv(t *testing.T, key, value string) {
	t.Helper()

	oldValue, hadOldValue := os.LookupEnv(key)
	t.Setenv(key, value)
	t.Cleanup(func() {
		if hadOldValue {
			_ = os.Setenv(key, oldValue)
			return
		}
		_ = os.Unsetenv(key)
	})
}

func unsetConfigEnv(t *testing.T) {
	t.Helper()

	for _, key := range []string{
		"PORT",
		"FRONTEND_ORIGIN",
		"DATABASE_URL",
		"REDIS_URL",
		"ACCESS_TOKEN_SECRET",
		"ACCESS_TOKEN_TTL",
		"REFRESH_TOKEN_TTL",
		"WORKER_POLL_INTERVAL",
		"AUTO_MIGRATE",
		"AUTO_SEED",
	} {
		setTestEnv(t, key, "")
	}
}

func TestLoadDefaults(t *testing.T) {
	unsetConfigEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.WorkerPollInterval, 10*time.Second; got != want {
		t.Fatalf("WorkerPollInterval = %v, want %v", got, want)
	}
}

func TestLoadParsesWorkerPollInterval(t *testing.T) {
	unsetConfigEnv(t)
	setTestEnv(t, "WORKER_POLL_INTERVAL", "45s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if got, want := cfg.WorkerPollInterval, 45*time.Second; got != want {
		t.Fatalf("WorkerPollInterval = %v, want %v", got, want)
	}
}