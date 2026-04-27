package middleware

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"wsapi/internal/config"
	"wsapi/internal/domain"
)

func TestValidateTokenUsesRolClaimWhenPresent(t *testing.T) {
	m := NewAuthMiddleware(&config.JWTConfig{Secret: "test-secret"}, nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  float64(2),
		"username": "operator",
		"rol":      "operador",
		"role_id":  float64(3),
		"is_root":  false,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}
	if claims.Rol != domain.RoleOperador {
		t.Fatalf("expected role %q, got %q", domain.RoleOperador, claims.Rol)
	}
}

func TestValidateTokenRejectsTokenWithoutRolClaim(t *testing.T) {
	m := NewAuthMiddleware(&config.JWTConfig{Secret: "test-secret"}, nil)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  float64(1),
		"username": "admin",
		"role_id":  float64(1),
		"is_root":  true,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	if _, err := m.ValidateToken(tokenString); err == nil {
		t.Fatal("expected token without rol claim to be rejected")
	}
}
