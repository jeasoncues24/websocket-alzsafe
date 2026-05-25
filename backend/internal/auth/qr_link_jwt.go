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
	claims := rawQRLinkClaims{
		EmpresaID: empresaID,
		PhoneID:   phoneID,
		Scope:     "qr_link",
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(qrLinkTokenExpiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error al firmar QR link token: %w", err)
	}
	return signed, nil
}

type rawQRLinkClaims struct {
	EmpresaID int64  `json:"sub"`
	PhoneID   int64  `json:"phone_id"`
	Scope     string `json:"scope"`
	jwt.RegisteredClaims
}

// ParseQRLinkToken valida y parsea un JWT de QR link. Rechaza tokens con scope distinto a "qr_link".
func ParseQRLinkToken(tokenString, secret string) (*domain.QRLinkClaims, error) {
	if secret == "" {
		return nil, fmt.Errorf("secret no puede estar vacío")
	}
	var claims rawQRLinkClaims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, fmt.Errorf("QR link token inválido: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("token no válido")
	}
	if claims.Scope != "qr_link" {
		return nil, fmt.Errorf("scope inválido para QR link: %q", claims.Scope)
	}
	if claims.EmpresaID <= 0 {
		return nil, fmt.Errorf("empresa_id ausente o inválido en token QR")
	}
	if claims.PhoneID <= 0 {
		return nil, fmt.Errorf("phone_id ausente o inválido en token QR")
	}
	return &domain.QRLinkClaims{EmpresaID: claims.EmpresaID, PhoneID: claims.PhoneID, Scope: claims.Scope}, nil
}
