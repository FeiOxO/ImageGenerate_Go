package config

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort         string
	JWTSecret       string
	JWTExpireHours  int
	MySQLDSN        string
	ImageAPIBaseURL string
	ImageAPIAppID   string
	ImageAPIKey     string
	ImageOutputDir  string
	WorkerPoolSize  int
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppPort:         getEnv("APP_PORT", "3000"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		MySQLDSN:        os.Getenv("MYSQL_DSN"),
		ImageAPIBaseURL: strings.TrimRight(os.Getenv("IMAGE_API_BASE_URL"), "/"),
		ImageAPIAppID:   os.Getenv("IMAGE_API_APP_ID"),
		ImageAPIKey:     os.Getenv("IMAGE_API_KEY"),
		ImageOutputDir:  getEnv("IMAGE_OUTPUT_DIR", "./storage/images"),
	}

	cfg.JWTExpireHours = getEnvInt("JWT_EXPIRE_HOURS", 168)
	cfg.WorkerPoolSize = getEnvInt("WORKER_POOL_SIZE", 8)

	if cfg.JWTSecret == "" {
		return cfg, errors.New("JWT_SECRET is required")
	}
	if cfg.MySQLDSN == "" {
		return cfg, errors.New("MYSQL_DSN is required")
	}
	if cfg.ImageAPIBaseURL == "" {
		return cfg, errors.New("IMAGE_API_BASE_URL is required")
	}
	if cfg.ImageAPIKey == "" {
		return cfg, errors.New("IMAGE_API_KEY is required")
	}
	if cfg.WorkerPoolSize < 1 {
		cfg.WorkerPoolSize = 8
	}

	return cfg, nil
}

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return number
}
