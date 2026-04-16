package config

import (
	"fmt"
	"os"
	"time"
)

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

const defaultJWTSecret = "wsapi-secret-key-change-in-production"

func LoadJWT() *JWTConfig {
	secret := getEnv("JWT_SECRET", "")
	if secret == "" {
		fmt.Println("[WARN] JWT_SECRET no configurado — usando secret por defecto, NO SEGURO para producción")
		secret = defaultJWTSecret
	}
	return &JWTConfig{
		Secret:        secret,
		Expiry:        getEnvDuration("JWT_EXPIRY", 24*time.Hour),
		RefreshExpiry: getEnvDuration("JWT_REFRESH_EXPIRY", 7*24*time.Hour),
		Issuer:        getEnv("JWT_ISSUER", "wsapi"),
	}
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		// Try parsing as duration string
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		// Try parsing as integer hours
		var n int
		for _, c := range value {
			if c < '0' || c > '9' {
				break
			}
			n = n*10 + int(c-'0')
		}
		if n > 0 {
			return time.Duration(n) * time.Hour
		}
	}
	return defaultValue
}
