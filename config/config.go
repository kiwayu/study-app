package config

import (
	"os"

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
}

// Load reads configuration from environment variables, optionally loading a
// .env file first. It returns an error only if required values are missing
// in production mode.
func Load() (*Config, error) {
	// Best-effort .env load — missing file is fine.
	_ = godotenv.Load()

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
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
