package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"wsapi/internal/auth"
	"wsapi/internal/config"
	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

// EmpresaAuthMiddleware valida el JWT de empresa y opcionalmente verifica
// que el telefono_id del path/query pertenece a esa empresa.
type EmpresaAuthMiddleware struct {
	jwtConfig     *config.JWTConfig
	empresaStore  domain.EmpresaStoreInterface
	telefonoStore *storage.TelefonoStore
}

func NewEmpresaAuthMiddleware(
	jwtConfig *config.JWTConfig,
	empresaStore domain.EmpresaStoreInterface,
	telefonoStore *storage.TelefonoStore,
) *EmpresaAuthMiddleware {
	return &EmpresaAuthMiddleware{
		jwtConfig:     jwtConfig,
		empresaStore:  empresaStore,
		telefonoStore: telefonoStore,
	}
}

// RequireEmpresaAuth valida el Bearer JWT de empresa e inyecta EmpresaJWTClaims en el contexto.
// También verifica que token_version del claim coincide con el DB (revocación).
func (m *EmpresaAuthMiddleware) RequireEmpresaAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeEmpresaError(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "JWT de empresa requerido")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeEmpresaError(w, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "Formato de JWT inválido")
				return
			}

			claims, err := auth.ParseEmpresaJWT(parts[1], m.jwtConfig.Secret)
			if err != nil {
				writeEmpresaError(w, http.StatusUnauthorized, "INVALID_TOKEN", "JWT inválido o expirado")
				return
			}

			// Rechazar tokens provisionales (QR link) en endpoints REST.
			if claims.Scope == "qr_link" {
				writeEmpresaError(w, http.StatusUnauthorized, "INVALID_TOKEN", "JWT inválido o expirado")
				return
			}

			// Verificar token_version contra DB (revocación).
			empresa, err := m.empresaStore.GetByID(claims.EmpresaID)
			if err != nil || empresa == nil {
				writeEmpresaError(w, http.StatusUnauthorized, "EMPRESA_NOT_FOUND", "Empresa no encontrada")
				return
			}
			if !empresa.Activo {
				writeEmpresaError(w, http.StatusForbidden, "EMPRESA_INACTIVE", "Empresa inactiva")
				return
			}
			if claims.TokenVersion < empresa.TokenVersion {
				writeEmpresaError(w, http.StatusUnauthorized, "TOKEN_REVOKED", "Token revocado")
				return
			}

			ctx := domain.WithEmpresaJWTClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireOwnership verifica que el telefono_id (query param o path) pertenece a la empresa del JWT.
// Debe usarse después de RequireEmpresaAuth.
// Extrae el telefono_id de: query param "telefono_id", luego del path (último segmento numérico).
func (m *EmpresaAuthMiddleware) RequireOwnership() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := domain.GetEmpresaJWTClaims(r.Context())
			if !ok {
				writeEmpresaError(w, http.StatusUnauthorized, "TOKEN_REQUIRED", "Autenticación de empresa requerida")
				return
			}

			telefonoID, err := extractTelefonoID(r)
			if err != nil {
				writeEmpresaError(w, http.StatusBadRequest, "MISSING_TELEFONO_ID", "telefono_id requerido")
				return
			}

			belongs, err := m.telefonoStore.BelongsToEmpresa(telefonoID, claims.EmpresaID)
			if err != nil {
				writeEmpresaError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Error verificando ownership")
				return
			}
			if !belongs {
				writeEmpresaError(w, http.StatusForbidden, "FORBIDDEN", "El teléfono no pertenece a esta empresa")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// extractTelefonoID obtiene el telefono_id del query param o del último segmento del path.
func extractTelefonoID(r *http.Request) (int64, error) {
	if v := r.URL.Query().Get("telefono_id"); v != "" {
		return strconv.ParseInt(v, 10, 64)
	}
	// Intentar extraer del path: /v1/telefonos/42/...
	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(path, "/")
	for i := len(segments) - 1; i >= 0; i-- {
		if id, err := strconv.ParseInt(segments[i], 10, 64); err == nil && id > 0 {
			return id, nil
		}
	}
	return 0, fmt.Errorf("telefono_id no encontrado")
}

func writeEmpresaError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"ok":false,"error":%q,"message":%q}`, code, message)
}
