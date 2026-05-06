package http

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userStore      *storage.AdminUserStore
	empresaStore   domain.EmpresaStoreInterface
	blacklistStore *storage.TokenBlacklistStore
	jwtConfig      *config.JWTConfig
}

func NewAuthHandler(userStore *storage.AdminUserStore, empresaStore domain.EmpresaStoreInterface, blacklistStore *storage.TokenBlacklistStore, jwtConfig *config.JWTConfig) *AuthHandler {
	return &AuthHandler{
		userStore:      userStore,
		empresaStore:   empresaStore,
		blacklistStore: blacklistStore,
		jwtConfig:      jwtConfig,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "JSON inválido")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeAPIError(w, http.StatusBadRequest, "Usuario y contraseña requeridos")
		return
	}

	// Get user from DB
	user, err := h.userStore.GetByUsername(req.Username)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error interno")
		return
	}
	if user == nil || !user.Activo {
		writeAPIError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeAPIError(w, http.StatusUnauthorized, "Credenciales inválidas")
		return
	}

	// Generate JWT
	token, err := h.generateToken(user)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al generar token")
		return
	}

	// Update last login
	h.userStore.UpdateLastLogin(user.ID)

	writeHandlerJSON(w, http.StatusOK, domain.LoginResponse{
		OK:      true,
		Token:   token,
		Message: "Login exitoso",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	// Get token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(h.jwtConfig.Secret), nil
			})
			if err == nil && token.Valid {
				if claims, ok := token.Claims.(jwt.MapClaims); ok {
					if jti, ok := claims["jti"].(string); ok {
						if exp, ok := claims["exp"].(float64); ok {
							expiresAt := time.Unix(int64(exp), 0)
							h.blacklistStore.Add(jti, expiresAt)
						}
					}
				}
			}
		}
	}

	writeHandlerJSON(w, http.StatusOK, map[string]bool{"ok": true, "logged_out": true})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeAPIError(w, http.StatusUnauthorized, "Token requerido")
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		writeAPIError(w, http.StatusUnauthorized, "Formato de token inválido")
		return
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		writeAPIError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "Token inválido")
		return
	}

	// Check if token is in blacklist
	if jti, ok := claims["jti"].(string); ok {
		blacklisted, _ := h.blacklistStore.IsBlacklisted(jti)
		if blacklisted {
			writeAPIError(w, http.StatusUnauthorized, "Token invalidado")
			return
		}
	}

	// Invalidar el token anterior
	if jti, ok := claims["jti"].(string); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expiresAt := time.Unix(int64(exp), 0)
			h.blacklistStore.Add(jti, expiresAt)
		}
	}

	// Get user and generate new token
	userID := int64(claims["user_id"].(float64))
	user, err := h.userStore.GetByID(userID)
	if err != nil || user == nil {
		writeAPIError(w, http.StatusUnauthorized, "Usuario no encontrado")
		return
	}

	newToken, err := h.generateToken(user)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al generar token")
		return
	}

	writeHandlerJSON(w, http.StatusOK, domain.LoginResponse{
		OK:      true,
		Token:   newToken,
		Message: "Token refrescado",
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	claims, ok := domain.GetAdminJWTClaims(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	// Get full user info
	user, err := h.userStore.GetByID(claims.UserID)
	if err != nil || user == nil {
		writeAPIError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	response := map[string]interface{}{
		"ok": true,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role_id":  user.RoleID,
			"is_root":  user.IsRoot,
			"activo":   user.Activo,
		},
	}

	writeHandlerJSON(w, http.StatusOK, response)
}

func (h *AuthHandler) generateToken(user *domain.AdminUser) (string, error) {
	now := time.Now()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	jti := hex.EncodeToString(b)

	claims := jwt.MapClaims{
		"jti":     jti,
		"user_id": float64(user.ID),
		"username": user.Username,
		"rol":     string(user.RoleName),
		"role_id": user.RoleID,
		"is_root": user.IsRoot,
		"iat":     now.Unix(),
		"exp":     now.Add(h.jwtConfig.Expiry).Unix(),
		"iss":     h.jwtConfig.Issuer,
	}

	if user.RoleID != nil {
		claims["role_id"] = float64(*user.RoleID)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtConfig.Secret))
}
