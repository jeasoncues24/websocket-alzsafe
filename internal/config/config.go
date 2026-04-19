package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	AppPort           string
	DBHost            string
	DBPort            string
	DBName            string
	DBUser            string
	DBPass            string
	WhatsAppSQLiteDir string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", ""),
		DBHost:            getEnv("DB_HOST", ""),
		DBPort:            getEnv("DB_PORT", ""),
		DBName:            getEnv("DB_NAME", ""),
		DBUser:            getEnv("DB_USER", ""),
		DBPass:            getEnv("DB_PASS", ""),
		WhatsAppSQLiteDir: getEnv("WHATSAPP_SQLITE_DIR", "sessions/whatsappmeow"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
