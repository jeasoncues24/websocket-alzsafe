package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

const apiKeyPublicPrefix = "wsk_live"

func GenerateAPIKeyMaterial() (prefix string, rawKey string, err error) {
	prefix, err = randomHex(4)
	if err != nil {
		return "", "", err
	}
	secret, err := randomHex(32)
	if err != nil {
		return "", "", err
	}
	rawKey = fmt.Sprintf("%s_%s_%s", apiKeyPublicPrefix, prefix, secret)
	return prefix, rawKey, nil
}

func ParseAPIKey(raw string) (prefix string, ok bool) {
	parts := strings.Split(raw, "_")
	if len(parts) != 4 {
		return "", false
	}
	if parts[0] != "wsk" || parts[1] != "live" {
		return "", false
	}
	if len(parts[2]) == 0 || len(parts[3]) == 0 {
		return "", false
	}
	return parts[2], true
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("error generando aleatorios: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
