package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"

	"github.com/golang-jwt/jwt/v5"
)

type AuthMiddleware struct {
	jwtConfig      *config.JWTConfig
	blacklistStore *storage.TokenBlacklistStore
}

func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, `{"ok":false,"error":%q,"message":%q}`, message, message)
}

func NewAuthMiddleware(jwtConfig *config.JWTConfig, blacklistStore *storage.TokenBlacklistStore) *AuthMiddleware {
	return &AuthMiddleware{jwtConfig: jwtConfig, blacklistStore: blacklistStore}
}

// ValidateToken validates the JWT token and extracts claims
func (m *AuthMiddleware) ValidateToken(tokenString string) (*domain.AdminJWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(m.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrSignatureInvalid
	}

	// Extract required claims with safe type assertions
	userIDRaw, ok := claims["user_id"].(float64)
	if !ok {
		return nil, jwt.ErrSignatureInvalid
	}
	userID := int64(userIDRaw)

	username, ok := claims["username"].(string)
	if !ok {
		return nil, jwt.ErrSignatureInvalid
	}

	rolRaw, ok := claims["rol"].(string)
	if !ok || rolRaw == "" {
		return nil, jwt.ErrSignatureInvalid
	}
	rol := domain.UserRole(rolRaw)

	// Extract is_root claim
	isRoot := false
	if v, ok := claims["is_root"]; ok {
		if b, ok := v.(bool); ok {
			isRoot = b
		}
	}

	jti, _ := claims["jti"].(string)

	return &domain.AdminJWTClaims{
		JTI:      jti,
		UserID:   userID,
		Username: username,
		Rol:      rol,
		IsRoot:   isRoot,
	}, nil
}

// RequireAuth returns a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeAuthError(w, http.StatusUnauthorized, "Token requerido")
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeAuthError(w, http.StatusUnauthorized, "Formato de token inválido")
				return
			}

			claims, err := m.ValidateToken(parts[1])
			if err != nil {
				writeAuthError(w, http.StatusUnauthorized, "Token inválido")
				return
			}

			// Check blacklist
			if m.blacklistStore != nil && claims.JTI != "" {
				blacklisted, _ := m.blacklistStore.IsBlacklisted(claims.JTI)
				if blacklisted {
					writeAuthError(w, http.StatusUnauthorized, "Token invalidado")
					return
				}
			}

			// Store claims in context
			ctx := r.Context()
			ctx = domain.WithAdminJWTClaims(ctx, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns a middleware that optionally extracts claims
func (m *AuthMiddleware) OptionalAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				next.ServeHTTP(w, r)
				return
			}

			claims, err := m.ValidateToken(parts[1])
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			ctx := r.Context()
			ctx = domain.WithAdminJWTClaims(ctx, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
