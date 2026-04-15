package config

import (
	"os"
	"time"
)

type JWTConfig struct {
	Secret        string
	Expiry        time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

func LoadJWT() *JWTConfig {
	return &JWTConfig{
		Secret:        getEnv("JWT_SECRET", "wsapi-secret-key-change-in-production"),
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
