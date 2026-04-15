package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv  string
	AppPort string
	DBHost  string
	DBPort  string
	DBName  string
	DBUser  string
	DBPass  string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv:  getEnv("APP_ENV", "development"),
		AppPort: getEnv("APP_PORT", "8080"),
		DBHost:  getEnv("DB_HOST", "localhost"),
		DBPort:  getEnv("DB_PORT", "3306"),
		DBName:  getEnv("DB_NAME", "wsapi"),
		DBUser:  getEnv("DB_USER", "root"),
		DBPass:  getEnv("DB_PASS", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
