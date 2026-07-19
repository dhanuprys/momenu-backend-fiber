package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort         string
	Env                string
	DBUrl              string
	JWTSecret          string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	FrontendURL        string
	TurnstileSecret           string
	DefaultProjectDiskQuotaMB int64
	ImageOptimizationQuality  int64
	CacheWarmerBaseURL        string
}

var AppConfig *Config

// Load loads the configuration from .env file or environment variables.
func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	AppConfig = &Config{
		ServerPort:         getEnv("SERVER_PORT", "3000"),
		Env:                getEnv("APP_ENV", "development"),
		DBUrl:              getEnv("DATABASE_URL", "host=localhost user=default password=default dbname=momenu port=5432 sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "supersecretkey"),
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:         getEnv("GOOGLE_REDIRECT_URL", "http://localhost:20261/api/v1/auth/google/callback"),
		FrontendURL:               getEnv("FRONTEND_URL", "http://localhost:20260"),
		TurnstileSecret:           getEnv("TURNSTILE_SECRET_KEY", ""),
		DefaultProjectDiskQuotaMB: getEnvAsInt64("DEFAULT_PROJECT_DISK_QUOTA_MB", 100),
		ImageOptimizationQuality:  getEnvAsInt64("IMAGE_OPTIMIZATION_QUALITY", 80),
		CacheWarmerBaseURL:        getEnv("CACHE_WARMER_BASE_URL", ""),
	}

	if AppConfig.Env == "production" && AppConfig.JWTSecret == "supersecretkey" {
		log.Fatal("JWT_SECRET environment variable is required in production environment")
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

func getEnvAsInt64(key string, fallback int64) int64 {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
			return value
		}
	}
	return fallback
}
