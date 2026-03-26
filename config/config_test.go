package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any env vars that might interfere.
	envVars := []string{
		"DATABASE_URL", "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
		"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "JWT_SECRET",
		"BASE_URL", "PORT", "ENV", "RATE_LIMIT", "RATE_BURST",
	}
	for _, v := range envVars {
		t.Setenv(v, "")
		os.Unsetenv(v)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("expected default port '8080', got %q", cfg.Port)
	}
	if cfg.Env != "development" {
		t.Errorf("expected default env 'development', got %q", cfg.Env)
	}
	if cfg.BaseURL != "http://localhost:8080" {
		t.Errorf("expected default BaseURL, got %q", cfg.BaseURL)
	}
	if cfg.RateLimit != 10 {
		t.Errorf("expected default RateLimit 10, got %f", cfg.RateLimit)
	}
	if cfg.RateBurst != 20 {
		t.Errorf("expected default RateBurst 20, got %d", cfg.RateBurst)
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/test")
	t.Setenv("JWT_SECRET", "my-secret")
	t.Setenv("BASE_URL", "https://example.com")
	t.Setenv("PORT", "3000")
	t.Setenv("ENV", "development")
	t.Setenv("RATE_LIMIT", "50")
	t.Setenv("RATE_BURST", "100")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("expected DatabaseURL 'postgres://localhost/test', got %q", cfg.DatabaseURL)
	}
	if cfg.JWTSecret != "my-secret" {
		t.Errorf("expected JWTSecret 'my-secret', got %q", cfg.JWTSecret)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Errorf("expected BaseURL 'https://example.com', got %q", cfg.BaseURL)
	}
	if cfg.Port != "3000" {
		t.Errorf("expected port '3000', got %q", cfg.Port)
	}
	if cfg.RateLimit != 50 {
		t.Errorf("expected RateLimit 50, got %f", cfg.RateLimit)
	}
	if cfg.RateBurst != 100 {
		t.Errorf("expected RateBurst 100, got %d", cfg.RateBurst)
	}
}

func TestLoad_ProductionMissingRequired(t *testing.T) {
	// Clear everything then set production mode.
	for _, v := range []string{
		"DATABASE_URL", "GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET",
		"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "JWT_SECRET",
		"BASE_URL", "PORT", "RATE_LIMIT", "RATE_BURST",
	} {
		os.Unsetenv(v)
	}
	t.Setenv("ENV", "production")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error in production mode with missing required vars")
	}
}
