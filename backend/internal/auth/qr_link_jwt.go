package auth

import (
	"fmt"
	"time"

	"wsapi/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

const qrLinkTokenExpiry = 600 * time.Second // 10 minutos

// GenerateQRLinkToken genera un JWT de corta duración para el QR link de un teléfono.
func GenerateQRLinkToken(empresaID, phoneID int64, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret no puede estar vacío")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      empresaID,
		"phone_id": phoneID,
		"scope":    "qr_link",
		"iat":      now.Unix(),
		"exp":      now.Add(qrLinkTokenExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error al firmar QR link token: %w", err)
	}
	return signed, nil
}

// ParseQRLinkToken valida y parsea un JWT de QR link. Rechaza tokens con scope distinto a "qr_link".
func ParseQRLinkToken(tokenString, secret string) (*domain.QRLinkClaims, error) {
	if secret == "" {
		return nil, fmt.Errorf("secret no puede estar vacío")
	}
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("QR link token inválido: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token no válido")
	}
	m, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("claims con formato inesperado")
	}
	var empresaID, phoneID int64
	if v, ok := m["sub"].(float64); ok {
		empresaID = int64(v)
	}
	if v, ok := m["phone_id"].(float64); ok {
		phoneID = int64(v)
	}
	scope, _ := m["scope"].(string)
	if scope != "qr_link" {
		return nil, fmt.Errorf("scope inválido para QR link: %q", scope)
	}
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id ausente o inválido en token QR")
	}
	if phoneID <= 0 {
		return nil, fmt.Errorf("phone_id ausente o inválido en token QR")
	}
	return &domain.QRLinkClaims{EmpresaID: empresaID, PhoneID: phoneID, Scope: scope}, nil
}
