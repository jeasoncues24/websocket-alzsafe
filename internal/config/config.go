package config

import "os"

type Config struct {
	AppEnv  string
	DBHost  string
	DBPort  string
	DBName  string
	DBUser  string
	DBPass  string
}

func Load() *Config {
	return &Config{
		AppEnv: os.Getenv("APP_ENV"),
		DBHost: os.Getenv("DB_HOST"),
		DBPort: os.Getenv("DB_PORT"),
		DBName: os.Getenv("DB_NAME"),
		DBUser: os.Getenv("DB_USER"),
		DBPass: os.Getenv("DB_PASS"),
	}
}
