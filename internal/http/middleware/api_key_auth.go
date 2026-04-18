package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

			started := time.Now()
			rw := &apiKeyResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r.WithContext(ctx))

			_ = m.apiKeyStore.RecordUsageEvent(&domain.ApiKeyUsageEvent{
				ApiKeyID:      key.ID,
				EmpresaID:     key.EmpresaID,
				TelefonoID:    key.TelefonoID,
				Method:        r.Method,
				Endpoint:      r.URL.Path,
				StatusCode:    rw.statusCode,
				LatencyMS:     int(time.Since(started).Milliseconds()),
				RequestUnits:  1,
				ResponseUnits: 0,
				RequestID:     strings.TrimSpace(r.Header.Get("X-Correlation-ID")),
			})
			_ = m.apiKeyStore.UpsertDailyUsage(&domain.ApiKeyUsageDaily{
				Day:          started.Format("2006-01-02"),
				ApiKeyID:     key.ID,
				EmpresaID:    key.EmpresaID,
				TelefonoID:   key.TelefonoID,
				RequestCount: 1,
				SuccessCount: func() int {
					if rw.statusCode < 400 {
						return 1
					}
					return 0
				}(),
				ErrorCount: func() int {
					if rw.statusCode >= 400 {
						return 1
					}
					return 0
				}(),
				LatencyAvgMS:   int(time.Since(started).Milliseconds()),
				MessagesSent:   0,
				BroadcastsSent: 0,
				BytesIn:        0,
				BytesOut:       0,
			})
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

type apiKeyResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *apiKeyResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
