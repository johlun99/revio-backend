package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabaseURL         string
	AppEnv              string
	AppPort             string
	JWTSecret           string
	CORSAllowedOrigins  string
}

func Load() *Config {
	cfg := &Config{
		DatabaseURL:        requireEnv("DATABASE_URL"),
		JWTSecret:          requireEnv("JWT_SECRET"),
		AppEnv:             getEnv("APP_ENV", "development"),
		AppPort:            getEnv("APP_PORT", "8080"),
		CORSAllowedOrigins: getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:5173"),
	}

	if len(cfg.JWTSecret) < 32 {
		panic("JWT_SECRET must be at least 32 characters")
	}

	return cfg
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
