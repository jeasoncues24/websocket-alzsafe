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
	userStore       *storage.AdminUserStore
	empresaStore    domain.EmpresaStoreInterface
	blacklistStore  *storage.TokenBlacklistStore
	jwtConfig       *config.JWTConfig
	userModuleStore *storage.UserModuleStore
	roleStore       *storage.RoleStore
	moduleStore     *storage.ModuleStore
}

func NewAuthHandler(
	userStore *storage.AdminUserStore,
	empresaStore domain.EmpresaStoreInterface,
	blacklistStore *storage.TokenBlacklistStore,
	jwtConfig *config.JWTConfig,
	userModuleStore *storage.UserModuleStore,
	roleStore *storage.RoleStore,
	moduleStore *storage.ModuleStore,
) *AuthHandler {
	return &AuthHandler{
		userStore:       userStore,
		empresaStore:    empresaStore,
		blacklistStore:  blacklistStore,
		jwtConfig:       jwtConfig,
		userModuleStore: userModuleStore,
		roleStore:       roleStore,
		moduleStore:     moduleStore,
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

	allowedModules := h.resolveAllowedModules(user)

	response := map[string]interface{}{
		"ok": true,
		"user": map[string]interface{}{
			"id":              user.ID,
			"username":        user.Username,
			"email":           user.Email,
			"role_id":         user.RoleID,
			"is_root":         user.IsRoot,
			"activo":          user.Activo,
			"allowed_modules": allowedModules,
		},
	}

	writeHandlerJSON(w, http.StatusOK, response)
}

// resolveAllowedModules determina los slugs de módulos efectivos para el usuario.
// Precedencia: is_root → user_modules override → role.permissions["all"] → role.permissions específicos → fallback dashboard.
func (h *AuthHandler) resolveAllowedModules(user *domain.AdminUser) []string {
	if user.IsRoot {
		return h.getAllModuleSlugs()
	}

	if h.userModuleStore != nil {
		modules, err := h.userModuleStore.GetByUserID(user.ID)
		if err == nil && len(modules) > 0 {
			slugs := make([]string, 0, len(modules))
			for _, m := range modules {
				slugs = append(slugs, m.Slug)
			}
			return slugs
		}
	}

	if user.RoleID != nil && h.roleStore != nil {
		role, err := h.roleStore.GetByID(*user.RoleID)
		if err == nil && role != nil {
			for _, p := range role.Permissions {
				if p == "all" {
					return h.getAllModuleSlugs()
				}
			}
			if len(role.Permissions) > 0 {
				return role.Permissions
			}
		}
	}

	return []string{"dashboard"}
}

func (h *AuthHandler) getAllModuleSlugs() []string {
	if h.moduleStore == nil {
		return []string{"dashboard"}
	}
	modules, err := h.moduleStore.GetAll()
	if err != nil || len(modules) == 0 {
		return []string{"dashboard"}
	}
	slugs := make([]string, 0, len(modules))
	for _, m := range modules {
		slugs = append(slugs, m.Slug)
	}
	return slugs
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

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	claims, ok := domain.GetAdminJWTClaims(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "Formato de petición inválido")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if req.Username == "" || req.Email == "" {
		writeAPIError(w, http.StatusBadRequest, "Nombre de usuario y correo electrónico son obligatorios")
		return
	}

	taken, err := h.userStore.IsUsernameTaken(req.Username, claims.UserID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al validar nombre de usuario")
		return
	}
	if taken {
		writeAPIError(w, http.StatusConflict, "El nombre de usuario ya está en uso")
		return
	}

	takenEmail, err := h.userStore.IsEmailTaken(req.Email, claims.UserID)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al validar correo electrónico")
		return
	}
	if takenEmail {
		writeAPIError(w, http.StatusConflict, "El correo electrónico ya está en uso")
		return
	}

	if err := h.userStore.UpdateProfile(claims.UserID, req.Username, req.Email); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al actualizar perfil")
		return
	}

	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}

func (h *AuthHandler) UpdateMePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		writeAPIError(w, http.StatusMethodNotAllowed, "Método no permitido")
		return
	}

	claims, ok := domain.GetAdminJWTClaims(r.Context())
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "Formato de petición inválido")
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeAPIError(w, http.StatusBadRequest, "Las contraseñas actual y nueva son obligatorias")
		return
	}

	user, err := h.userStore.GetByID(claims.UserID)
	if err != nil || user == nil {
		// Mitigación de Timing Attack
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$10$fG6T6T6T6T6T6T6T6T6T6eDummyDummyDummyDummyDummy"), []byte(req.CurrentPassword))
		writeAPIError(w, http.StatusNotFound, "Usuario no encontrado")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		writeAPIError(w, http.StatusUnauthorized, "Contraseña actual incorrecta")
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al cifrar nueva contraseña")
		return
	}

	if err := h.userStore.UpdatePassword(claims.UserID, string(newHash)); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "Error al actualizar contraseña")
		return
	}

	writeHandlerJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
}
