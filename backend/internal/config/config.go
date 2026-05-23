package config

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv            string
	AppPort           string
	Version           string
	DBHost            string
	DBPort            string
	DBName            string
	DBUser            string
	DBPass            string
	WhatsAppSQLiteDir string

	WhatsAppBootstrapEnabled        bool
	WhatsAppBootstrapMaxConcurrency int
	WhatsAppBootstrapTimeoutSec     int

	WhatsAppDebugLogDir        string
	WhatsAppDebugLogPerAccount bool
	WhatsAppDebugLogLevel      string
	WhatsAppConsoleLogLevel    string
}

func Load() *Config {
	_ = godotenv.Load()
	return &Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		AppPort:           getEnv("APP_PORT", ""),
		Version:           getEnv("APP_VERSION", "dev"),
		DBHost:            getEnv("DB_HOST", ""),
		DBPort:            getEnv("DB_PORT", ""),
		DBName:            getEnv("DB_NAME", ""),
		DBUser:            getEnv("DB_USER", ""),
		DBPass:            getEnv("DB_PASS", ""),
		WhatsAppSQLiteDir: getEnv("WHATSAPP_SQLITE_DIR", "sessions/whatsappmeow"),

		WhatsAppBootstrapEnabled:        getEnvBool("WHATSAPP_BOOTSTRAP_ENABLED", true),
		WhatsAppBootstrapMaxConcurrency: getEnvInt("WHATSAPP_BOOTSTRAP_MAX_CONCURRENCY", 4),
		WhatsAppBootstrapTimeoutSec:     getEnvInt("WHATSAPP_BOOTSTRAP_TIMEOUT_SEC", 60),

		WhatsAppDebugLogDir:        getEnv("WHATSAPP_DEBUG_LOG_DIR", "debug_log"),
		WhatsAppDebugLogPerAccount: getEnvBool("WHATSAPP_DEBUG_LOG_PER_ACCOUNT", true),
		WhatsAppDebugLogLevel:      strings.ToUpper(getEnv("WHATSAPP_DEBUG_LOG_LEVEL", "DEBUG")),
		WhatsAppConsoleLogLevel:    strings.ToUpper(getEnv("WHATSAPP_CONSOLE_LOG_LEVEL", "INFO")),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		defaultStr := "false"
		if defaultValue {
			defaultStr = "true"
		}
		println("[WARN] env " + key + ": valor no reconocido como booleano: " + value + ", usando default: " + defaultStr)
		return defaultValue
	}
	return parsed
}

func getEnvInt(key string, defaultValue int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
