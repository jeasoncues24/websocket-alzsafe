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
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"ok": false, "error": "JSON inválido"}`, http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, `{"ok": false, "error": "Usuario y contraseña requeridos"}`, http.StatusBadRequest)
		return
	}

	// Get user from DB
	user, err := h.userStore.GetByUsername(req.Username)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error interno"}`, http.StatusInternalServerError)
		return
	}
	if user == nil || !user.Activo {
		http.Error(w, `{"ok": false, "error": "Credenciales inválidas"}`, http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"ok": false, "error": "Credenciales inválidas"}`, http.StatusUnauthorized)
		return
	}

	// Get empresa info if user has empresa
	var empresaRUC, empresaNombre *string
	if user.EmpresaID != nil {
		empresa, err := h.empresaStore.GetByID(*user.EmpresaID)
		if err == nil && empresa != nil {
			empresaRUC = &empresa.RUC
			empresaNombre = &empresa.Nombre
		}
	}

	// Generate JWT
	token, err := h.generateToken(user, empresaRUC, empresaNombre)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al generar token"}`, http.StatusInternalServerError)
		return
	}

	// Update last login
	h.userStore.UpdateLastLogin(user.ID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.LoginResponse{
		OK:      true,
		Token:   token,
		Message: "Login exitoso",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true, "logged_out": true})
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, `{"ok": false, "error": "Token requerido"}`, http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, `{"ok": false, "error": "Formato de token inválido"}`, http.StatusUnauthorized)
		return
	}

	tokenString := parts[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtConfig.Secret), nil
	})

	if err != nil || !token.Valid {
		http.Error(w, `{"ok": false, "error": "Token inválido"}`, http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, `{"ok": false, "error": "Token inválido"}`, http.StatusUnauthorized)
		return
	}

	// Check if token is in blacklist
	if jti, ok := claims["jti"].(string); ok {
		blacklisted, _ := h.blacklistStore.IsBlacklisted(jti)
		if blacklisted {
			http.Error(w, `{"ok": false, "error": "Token invalidado"}`, http.StatusUnauthorized)
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
		http.Error(w, `{"ok": false, "error": "Usuario no encontrado"}`, http.StatusUnauthorized)
		return
	}

	var empresaRUC, empresaNombre *string
	if user.EmpresaID != nil {
		empresa, _ := h.empresaStore.GetByID(*user.EmpresaID)
		if empresa != nil {
			empresaRUC = &empresa.RUC
			empresaNombre = &empresa.Nombre
		}
	}

	newToken, err := h.generateToken(user, empresaRUC, empresaNombre)
	if err != nil {
		http.Error(w, `{"ok": false, "error": "Error al generar token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(domain.LoginResponse{
		OK:      true,
		Token:   newToken,
		Message: "Token refrescado",
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"ok": false, "error": "Método no permitido"}`, http.StatusMethodNotAllowed)
		return
	}

	claims, ok := domain.GetAdminJWTClaims(r.Context())
	if !ok {
		http.Error(w, `{"ok": false, "error": "No autenticado"}`, http.StatusUnauthorized)
		return
	}

	// Get full user info
	user, err := h.userStore.GetByID(claims.UserID)
	if err != nil || user == nil {
		http.Error(w, `{"ok": false, "error": "Usuario no encontrado"}`, http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"ok": true,
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"rol":      user.Rol,
			"activo":   user.Activo,
		},
	}

	if user.EmpresaID != nil {
		empresa, _ := h.empresaStore.GetByID(*user.EmpresaID)
		if empresa != nil {
			response["empresa"] = map[string]interface{}{
				"id":     empresa.ID,
				"ruc":    empresa.RUC,
				"nombre": empresa.Nombre,
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) generateToken(user *domain.AdminUser, empresaRUC, empresaNombre *string) (string, error) {
	now := time.Now()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	jti := hex.EncodeToString(b)

	claims := jwt.MapClaims{
		"jti":            jti,
		"user_id":        float64(user.ID),
		"username":       user.Username,
		"rol":            string(user.Rol),
		"is_root":        user.IsRoot,
		"iat":            now.Unix(),
		"exp":            now.Add(h.jwtConfig.Expiry).Unix(),
		"iss":            h.jwtConfig.Issuer,
		"empresa_ruc":    empresaRUC,
		"empresa_nombre": empresaNombre,
	}

	if user.EmpresaID != nil {
		claims["empresa_id"] = float64(*user.EmpresaID)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtConfig.Secret))
}
