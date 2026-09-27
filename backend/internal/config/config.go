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
	JWTSecret       string
	MaxBotToken     string
	ShutdownTimeout time.Duration
}

func Load() *Config {
	_ = godotenv.Load()

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		panic("JWT_SECRET is required")
	}

	return &Config{
		Env:             getEnv("ENV", "dev"),
		Addr:            getEnv("ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5433/maxapp?sslmode=disable"),
		JWTSecret:       jwtSecret,
		MaxBotToken:     os.Getenv("MAX_BOT_TOKEN"),
		ShutdownTimeout: 10 * time.Second,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
