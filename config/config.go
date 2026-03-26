package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	DatabaseURL        string
	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
	JWTSecret          string
	BaseURL            string
	Port               string
	Env                string // "development" or "production"
	RateLimit          float64
	RateBurst          int
}

// Load reads configuration from environment variables, optionally loading a
// .env file first. It returns an error only if required values are missing
// in production mode.
func Load() (*Config, error) {
	// Best-effort .env load — missing file is fine.
	_ = godotenv.Load()

	rateLimit, _ := strconv.ParseFloat(getenv("RATE_LIMIT", "10"), 64)
	if rateLimit <= 0 {
		rateLimit = 10
	}
	rateBurst, _ := strconv.Atoi(getenv("RATE_BURST", "20"))
	if rateBurst <= 0 {
		rateBurst = 20
	}

	cfg := &Config{
		DatabaseURL:        getenv("DATABASE_URL", ""),
		GoogleClientID:     getenv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getenv("GOOGLE_CLIENT_SECRET", ""),
		GitHubClientID:     getenv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret: getenv("GITHUB_CLIENT_SECRET", ""),
		JWTSecret:          getenv("JWT_SECRET", ""),
		BaseURL:            getenv("BASE_URL", "http://localhost:8080"),
		Port:               getenv("PORT", "8080"),
		Env:                getenv("ENV", "development"),
		RateLimit:          rateLimit,
		RateBurst:          rateBurst,
	}

	// In production, require critical configuration values.
	if cfg.Env == "production" {
		var missing []string
		if cfg.DatabaseURL == "" {
			missing = append(missing, "DATABASE_URL")
		}
		if cfg.JWTSecret == "" {
			missing = append(missing, "JWT_SECRET")
		}
		if cfg.GoogleClientID == "" {
			missing = append(missing, "GOOGLE_CLIENT_ID")
		}
		if cfg.GoogleClientSecret == "" {
			missing = append(missing, "GOOGLE_CLIENT_SECRET")
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("production mode requires these env vars: %v", missing)
		}
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
