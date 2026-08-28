package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken string
	TelegramAdminIDs []int64
	DatabaseURL      string
	RedisURL         string
	JWTSecret        string
	AppPort          string
	LogLevel         string
}

func Load() (*Config, error) {
	// Try loading .env file if it exists
	_ = godotenv.Load()

	cfg := &Config{
		TelegramBotToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		RedisURL:         os.Getenv("REDIS_URL"),
		JWTSecret:        getEnv("JWT_SECRET", "super_secret_key_123"),
		AppPort:          getEnv("APP_PORT", "8080"),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}

	adminIDsStr := os.Getenv("TELEGRAM_ADMIN_IDS")
	if adminIDsStr != "" {
		parts := strings.Split(adminIDsStr, ",")
		for _, p := range parts {
			if id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64); err == nil {
				cfg.TelegramAdminIDs = append(cfg.TelegramAdminIDs, id)
			}
		}
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultValue
}
