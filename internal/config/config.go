// Package config loads service settings from environment variables.
package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// MinJWTSecretLen is the minimum accepted length of JWT_SECRET.
const MinJWTSecretLen = 32

// Config holds all runtime settings of the service.
type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	JWTTTL      time.Duration
}

// Load reads the configuration from the process environment.
func Load() (Config, error) {
	return load(os.Getenv)
}

// load is the testable core of Load: getenv is injected so tests
// don't have to touch real environment variables.
func load(getenv func(string) string) (Config, error) {
	cfg := Config{
		Port:        getenv("PORT"),
		DatabaseURL: getenv("DATABASE_URL"),
		JWTSecret:   getenv("JWT_SECRET"),
		JWTTTL:      time.Hour,
	}

	if cfg.Port == "" {
		cfg.Port = "8080"
	}

	var errs []error
	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}
	if len(cfg.JWTSecret) < MinJWTSecretLen {
		errs = append(errs, fmt.Errorf("JWT_SECRET is required and must be at least %d characters", MinJWTSecretLen))
	}
	if raw := getenv("JWT_TTL"); raw != "" {
		ttl, err := time.ParseDuration(raw)
		if err != nil || ttl <= 0 {
			errs = append(errs, fmt.Errorf("JWT_TTL must be a positive duration like 15m or 1h, got %q", raw))
		} else {
			cfg.JWTTTL = ttl
		}
	}

	return cfg, errors.Join(errs...)
}
