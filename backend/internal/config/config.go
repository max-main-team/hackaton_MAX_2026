package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env             string
	Addr            string
	DatabaseURL     string
	MaxAppSecretKey string
	ShutdownTimeout time.Duration
}

func Load() *Config {
	_ = godotenv.Load()

	return &Config{
		Env:             getEnv("ENV", "dev"),
		Addr:            getEnv("ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/maxapp?sslmode=disable"),
		MaxAppSecretKey: os.Getenv("MAX_APP_SECRET_KEY"),
		ShutdownTimeout: 10 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
