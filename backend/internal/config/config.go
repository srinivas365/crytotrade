package config

import "os"

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
}

func Load() Config {
	return Config{
		DatabaseURL: env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/cryptotrade?sslmode=disable"),
		JWTSecret:   env("JWT_SECRET", "change-me-in-production"),
		Port:        env("PORT", "8080"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
