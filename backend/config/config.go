package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port              string
	FrontendOrigin    string
	DatabaseURL       string
	RedisURL          string
	AccessTokenSecret string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	WorkerPollInterval time.Duration
	AutoMigrate       bool
	AutoSeed          bool
	AccessCookieName  string
	RefreshCookieName string
}

func Load() (Config, error) {
	accessTTL, err := parseDuration("ACCESS_TOKEN_TTL", "15m")
	if err != nil {
		return Config{}, err
	}

	refreshTTL, err := parseDuration("REFRESH_TOKEN_TTL", "168h")
	if err != nil {
		return Config{}, err
	}

	workerPollInterval, err := parseDuration("WORKER_POLL_INTERVAL", "10s")
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Port:               envOrDefault("PORT", "58080"),
		FrontendOrigin:     envOrDefault("FRONTEND_ORIGIN", "http://127.0.0.1:5173"),
		DatabaseURL:        envOrDefault("DATABASE_URL", "postgres://postgres:postgres@127.0.0.1:55432/encanto?sslmode=disable"),
		RedisURL:           envOrDefault("REDIS_URL", "redis://127.0.0.1:56379/0"),
		AccessTokenSecret:  envOrDefault("ACCESS_TOKEN_SECRET", "encanto-dev-access-secret"),
		AccessTokenTTL:     accessTTL,
		RefreshTokenTTL:    refreshTTL,
		WorkerPollInterval: workerPollInterval,
		AutoMigrate:        envBool("AUTO_MIGRATE", true),
		AutoSeed:           envBool("AUTO_SEED", true),
		AccessCookieName:   "encanto_access",
		RefreshCookieName:  "encanto_refresh",
	}

	if cfg.AccessTokenSecret == "" {
		return Config{}, fmt.Errorf("access token secret must not be empty")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value == "1" || value == "true" || value == "TRUE"
}

func parseDuration(key, fallback string) (time.Duration, error) {
	value := envOrDefault(key, fallback)
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return duration, nil
}
