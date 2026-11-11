package config

import (
	"strings"
	"testing"
	"time"
)

const validSecret = "0123456789abcdef0123456789abcdef" // exactly 32 chars

func env(m map[string]string) func(string) string {
	return func(key string) string { return m[key] }
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := load(env(map[string]string{
		"DATABASE_URL": "postgres://localhost/db",
		"JWT_SECRET":   validSecret,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.JWTTTL != time.Hour {
		t.Errorf("JWTTTL = %v, want 1h", cfg.JWTTTL)
	}
}

func TestLoad_Overrides(t *testing.T) {
	cfg, err := load(env(map[string]string{
		"PORT":         "9090",
		"DATABASE_URL": "postgres://localhost/db",
		"JWT_SECRET":   validSecret,
		"JWT_TTL":      "15m",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.JWTTTL != 15*time.Minute {
		t.Errorf("JWTTTL = %v, want 15m", cfg.JWTTTL)
	}
}

func TestLoad_Errors(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "missing database url",
			env:     map[string]string{"JWT_SECRET": validSecret},
			wantErr: "DATABASE_URL",
		},
		{
			name:    "missing jwt secret",
			env:     map[string]string{"DATABASE_URL": "x"},
			wantErr: "JWT_SECRET",
		},
		{
			name:    "jwt secret one char too short",
			env:     map[string]string{"DATABASE_URL": "x", "JWT_SECRET": validSecret[:31]},
			wantErr: "JWT_SECRET",
		},
		{
			name:    "jwt ttl not a duration",
			env:     map[string]string{"DATABASE_URL": "x", "JWT_SECRET": validSecret, "JWT_TTL": "hour"},
			wantErr: "JWT_TTL",
		},
		{
			name:    "jwt ttl negative",
			env:     map[string]string{"DATABASE_URL": "x", "JWT_SECRET": validSecret, "JWT_TTL": "-5m"},
			wantErr: "JWT_TTL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := load(env(tt.env))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not mention %q", err, tt.wantErr)
			}
		})
	}
}
