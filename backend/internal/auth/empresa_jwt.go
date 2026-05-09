package auth

import (
	"fmt"
	"time"

	"wsapi/internal/domain"

	"github.com/golang-jwt/jwt/v5"
)

const empresaTokenExpiry = 5 * 365 * 24 * time.Hour // 5 años

// GenerateEmpresaJWT genera un JWT de larga duración (5 años) para una empresa.
// Claims: sub=empresa_id, ver=token_version, ruc, nombre, permissions, iss, iat, exp.
func GenerateEmpresaJWT(empresa *domain.Empresa, secret, issuer string) (string, error) {
	now := time.Now()

	permissions := empresa.Permissions
	if permissions == nil {
		permissions = []string{}
	}

	claims := jwt.MapClaims{
		"sub":         empresa.ID,
		"ver":         empresa.TokenVersion,
		"ruc":         empresa.RUC,
		"nombre":      empresa.Nombre,
		"permissions": permissions,
		"iss":         issuer,
		"iat":         now.Unix(),
		"exp":         now.Add(empresaTokenExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error al firmar JWT de empresa: %w", err)
	}
	return signed, nil
}

const qrLinkTokenExpirySeconds = 600 // 10 minutos

// GenerateQRLinkToken genera un JWT de corta duración (10 min) para compartir
// el enlace de QR de un teléfono específico sin acceso completo a la API de empresa.
func GenerateQRLinkToken(empresaID, phoneID int64, secret string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      empresaID,
		"phone_id": phoneID,
		"scope":    "qr_link",
		"iat":      now.Unix(),
		"exp":      now.Add(qrLinkTokenExpirySeconds * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("error al firmar QR link token: %w", err)
	}
	return signed, nil
}

// ParseEmpresaJWT valida y parsea un JWT de empresa.
// Retorna EmpresaJWTClaims o error si el JWT es inválido/expirado.
func ParseEmpresaJWT(tokenString, secret string) (*domain.EmpresaJWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritmo de firma inesperado: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("JWT inválido: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("JWT no válido")
	}

	mapClaims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("claims con formato inesperado")
	}

	claims, err := extractEmpresaJWTClaims(mapClaims)
	if err != nil {
		return nil, fmt.Errorf("error al extraer claims: %w", err)
	}
	return claims, nil
}

// extractEmpresaJWTClaims convierte jwt.MapClaims → EmpresaJWTClaims
func extractEmpresaJWTClaims(m jwt.MapClaims) (*domain.EmpresaJWTClaims, error) {
	subRaw, ok := m["sub"]
	if !ok {
		return nil, fmt.Errorf("claim 'sub' ausente")
	}

	var empresaID int64
	switch v := subRaw.(type) {
	case float64:
		empresaID = int64(v)
	case int64:
		empresaID = v
	default:
		return nil, fmt.Errorf("claim 'sub' con tipo inesperado")
	}

	var tokenVersion int
	if v, ok := m["ver"].(float64); ok {
		tokenVersion = int(v)
	}

	ruc, _ := m["ruc"].(string)
	nombre, _ := m["nombre"].(string)

	var permissions []string
	if raw, ok := m["permissions"]; ok {
		if arr, ok := raw.([]interface{}); ok {
			for _, p := range arr {
				if s, ok := p.(string); ok {
					permissions = append(permissions, s)
				}
			}
		}
	}

	claims := &domain.EmpresaJWTClaims{
		EmpresaID:     empresaID,
		TokenVersion:  tokenVersion,
		EmpresaRUC:    ruc,
		EmpresaNombre: nombre,
		Permissions:   permissions,
	}

	// Campos opcionales para tokens provisionales (QR link)
	if v, ok := m["scope"].(string); ok {
		claims.Scope = v
	}
	if v, ok := m["phone_id"].(float64); ok {
		claims.PhoneID = int64(v)
	}

	return claims, nil
}
