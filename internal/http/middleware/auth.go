package middleware

import (
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

func NewAuthMiddleware(jwtConfig *config.JWTConfig, blacklistStore *storage.TokenBlacklistStore) *AuthMiddleware {
	return &AuthMiddleware{jwtConfig: jwtConfig, blacklistStore: blacklistStore}
}

// ValidateToken validates the JWT token and extracts claims
func (m *AuthMiddleware) ValidateToken(tokenString string) (*domain.TokenClaims, error) {
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
	if !ok {
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

	var empresaID *int64
	if v, ok := claims["empresa_id"]; ok && v != nil {
		if f, ok := v.(float64); ok && f > 0 {
			eid := int64(f)
			empresaID = &eid
		}
	}

	var empresaRUC *string
	if v, ok := claims["empresa_ruc"]; ok && v != nil {
		if s, ok := v.(string); ok {
			empresaRUC = &s
		}
	}

	var empresaNombre *string
	if v, ok := claims["empresa_nombre"]; ok && v != nil {
		if s, ok := v.(string); ok {
			empresaNombre = &s
		}
	}

	return &domain.TokenClaims{
		JTI:           jti,
		UserID:        userID,
		Username:      username,
		Rol:           rol,
		IsRoot:        isRoot,
		EmpresaID:     empresaID,
		EmpresaRUC:    empresaRUC,
		EmpresaNombre: empresaNombre,
	}, nil
}

// RequireAuth returns a middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"ok": false, "error": "Token requerido"}`, http.StatusUnauthorized)
				return
			}

			// Extract token from "Bearer <token>"
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"ok": false, "error": "Formato de token inválido"}`, http.StatusUnauthorized)
				return
			}

			claims, err := m.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, `{"ok": false, "error": "Token inválido"}`, http.StatusUnauthorized)
				return
			}

			// Check blacklist
			if m.blacklistStore != nil && claims.JTI != "" {
				blacklisted, _ := m.blacklistStore.IsBlacklisted(claims.JTI)
				if blacklisted {
					http.Error(w, `{"ok": false, "error": "Token invalidado"}`, http.StatusUnauthorized)
					return
				}
			}

			// Store claims in context
			ctx := r.Context()
			ctx = domain.WithTokenClaims(ctx, claims)
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
			ctx = domain.WithTokenClaims(ctx, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
