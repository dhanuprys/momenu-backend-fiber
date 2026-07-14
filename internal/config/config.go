package config

import (
	"log"
	"os"

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
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:20261/api/v1/auth/google/callback"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:20260"),
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
