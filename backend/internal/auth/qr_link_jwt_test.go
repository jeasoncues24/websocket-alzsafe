package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestQRLinkToken_GenerateAndParse_Success(t *testing.T) {
	secret := "test-secret"
	empresaID := int64(123)
	phoneID := int64(456)

	token, err := GenerateQRLinkToken(empresaID, phoneID, secret)
	if err != nil {
		t.Fatalf("Error inesperado al generar token: %v", err)
	}

	claims, err := ParseQRLinkToken(token, secret)
	if err != nil {
		t.Fatalf("Error inesperado al parsear token: %v", err)
	}

	if claims.EmpresaID != empresaID {
		t.Errorf("EmpresaID esperado %d, obtenido %d", empresaID, claims.EmpresaID)
	}
	if claims.PhoneID != phoneID {
		t.Errorf("PhoneID esperado %d, obtenido %d", phoneID, claims.PhoneID)
	}
	if claims.Scope != "qr_link" {
		t.Errorf("Scope esperado qr_link, obtenido %s", claims.Scope)
	}
}

func TestQRLinkToken_Parse_InvalidSecret(t *testing.T) {
	secret := "test-secret"
	token, _ := GenerateQRLinkToken(1, 2, secret)

	_, err := ParseQRLinkToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("Se esperaba error al parsear con secret incorrecto")
	}
}

func TestQRLinkToken_Parse_InvalidScope(t *testing.T) {
	secret := "test-secret"
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      float64(1),
		"phone_id": float64(2),
		"scope":    "wrong_scope",
		"iat":      now.Unix(),
		"exp":      now.Add(600 * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Error inesperado al firmar token de test: %v", err)
	}

	_, err = ParseQRLinkToken(signed, secret)
	if err == nil {
		t.Fatal("Se esperaba error al parsear token con scope inválido")
	}
}

func TestQRLinkToken_Parse_MissingPhoneID(t *testing.T) {
	secret := "test-secret"
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   float64(1),
		"scope": "qr_link",
		"iat":   now.Unix(),
		"exp":   now.Add(600 * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("Error inesperado al firmar token de test: %v", err)
	}

	_, err = ParseQRLinkToken(signed, secret)
	if err == nil {
		t.Fatal("Se esperaba error al parsear token sin phone_id")
	}
}
