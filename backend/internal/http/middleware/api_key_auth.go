package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"wsapi/internal/domain"
	"wsapi/internal/storage"
)

type ApiKeyAuthMiddleware struct {
	apiKeyStore   *storage.ApiKeyStore
	empresaStore  domain.EmpresaStoreInterface
	telefonoStore *storage.TelefonoStore
}

func NewApiKeyAuthMiddleware(apiKeyStore *storage.ApiKeyStore, empresaStore domain.EmpresaStoreInterface, telefonoStore *storage.TelefonoStore) *ApiKeyAuthMiddleware {
	return &ApiKeyAuthMiddleware{apiKeyStore: apiKeyStore, empresaStore: empresaStore, telefonoStore: telefonoStore}
}

func (m *ApiKeyAuthMiddleware) RequireApiKeyAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawKey := extractAPIKey(r)
			if rawKey == "" {
				writeAPIKeyError(w, http.StatusUnauthorized, "API_KEY_REQUIRED", "API key requerida")
				return
			}

			key, err := m.apiKeyStore.Validate(rawKey)
			if err != nil || key == nil {
				writeAPIKeyError(w, http.StatusUnauthorized, "INVALID_API_KEY", "API key inválida o expirada")
				return
			}

			phone, err := m.telefonoStore.GetByID(key.TelefonoID)
			if err != nil || phone == nil {
				writeAPIKeyError(w, http.StatusUnauthorized, "TELEFONO_NOT_FOUND", "Teléfono no encontrado")
				return
			}
			if phone.EmpresaID != key.EmpresaID {
				writeAPIKeyError(w, http.StatusForbidden, "FORBIDDEN", "La key no corresponde al teléfono esperado")
				return
			}

			empresa, err := m.empresaStore.GetByID(key.EmpresaID)
			if err != nil || empresa == nil {
				writeAPIKeyError(w, http.StatusUnauthorized, "EMPRESA_NOT_FOUND", "Empresa no encontrada")
				return
			}
			if !empresa.Activo {
				writeAPIKeyError(w, http.StatusForbidden, "EMPRESA_INACTIVE", "Empresa inactiva")
				return
			}

			ctx := domain.WithApiKeyClaims(r.Context(), &domain.ApiKeyClaims{
				ApiKeyID:   key.ID,
				EmpresaID:  key.EmpresaID,
				TelefonoID: key.TelefonoID,
				KeyPrefix:  key.KeyPrefix,
				Scopes:     key.Scopes,
			})
			ctx = domain.WithEmpresaID(ctx, key.EmpresaID)
			ctx = domain.WithEmpresaJWTClaims(ctx, &domain.EmpresaJWTClaims{
				EmpresaID:     key.EmpresaID,
				TokenVersion:  0,
				EmpresaRUC:    empresa.RUC,
				EmpresaNombre: empresa.Nombre,
				Permissions:   key.Scopes,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractAPIKey(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-API-Key")); v != "" {
		return v
	}
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if authHeader == "" {
		return ""
	}
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if strings.EqualFold(parts[0], "ApiKey") {
		return strings.TrimSpace(parts[1])
	}
	if strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func writeAPIKeyError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	fmt.Fprintf(w, `{"ok":false,"error":%q,"message":%q}`, code, message)
}

