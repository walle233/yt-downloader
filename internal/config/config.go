package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv             string
	AppDomain          string
	AppTimezone        string
	AppLocation        *time.Location
	APIPort            string
	ClerkSecretKey     string
	DatabaseURL        string
	RedisAddr          string
	DownloadRoot       string
	YTDLPCookiesFile   string
	YTDLPJSRuntimes    string
	YTDLPRemoteComponents string
	DownloadTTLHours   int
	WorkerConcurrency  int
	WorkerPollInterval time.Duration
	R2                 R2Config
}

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Endpoint        string
	PublicBaseURL   string
}

func Load() (Config, error) {
	cfg := Config{
		AppEnv:             getenv("APP_ENV", "development"),
		AppDomain:          getenv("APP_DOMAIN", "localhost"),
		AppTimezone:        getenv("APP_TIMEZONE", "Asia/Shanghai"),
		APIPort:            getenv("API_PORT", "8080"),
		ClerkSecretKey:     getenv("CLERK_SECRET_KEY", ""),
		DatabaseURL:        getenv("DATABASE_URL", "postgres://ytvideos:ytvideos@postgres:5432/ytvideos?sslmode=disable"),
		RedisAddr:          getenv("REDIS_ADDR", "redis:6379"),
		DownloadRoot:       getenv("DOWNLOAD_ROOT", "/data/jobs"),
		YTDLPCookiesFile:   getenv("YTDLP_COOKIES_FILE", ""),
		YTDLPJSRuntimes:    getenv("YTDLP_JS_RUNTIMES", "node"),
		YTDLPRemoteComponents: getenv("YTDLP_REMOTE_COMPONENTS", "ejs:github"),
		DownloadTTLHours:   getenvInt("DOWNLOAD_TTL_HOURS", 24),
		WorkerConcurrency:  getenvInt("WORKER_CONCURRENCY", 2),
		WorkerPollInterval: getenvDuration("WORKER_POLL_INTERVAL", 5*time.Second),
		R2: R2Config{
			AccountID:       getenv("R2_ACCOUNT_ID", ""),
			AccessKeyID:     getenv("R2_ACCESS_KEY_ID", ""),
			SecretAccessKey: getenv("R2_SECRET_ACCESS_KEY", ""),
			Bucket:          getenv("R2_BUCKET", ""),
			Endpoint:        getenv("R2_ENDPOINT", ""),
			PublicBaseURL:   getenv("R2_PUBLIC_BASE_URL", ""),
		},
	}

	if cfg.APIPort == "" {
		return Config{}, fmt.Errorf("API_PORT must not be empty")
	}

	location, err := loadLocation(cfg.AppTimezone)
	if err != nil {
		return Config{}, fmt.Errorf("load app timezone: %w", err)
	}
	cfg.AppLocation = location

	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func loadLocation(name string) (*time.Location, error) {
	location, err := time.LoadLocation(name)
	if err == nil {
		return location, nil
	}

	switch name {
	case "Asia/Shanghai":
		return time.FixedZone("Asia/Shanghai", 8*60*60), nil
	default:
		return nil, err
	}
}
