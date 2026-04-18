package domain

import "context"

type ApiKeyClaims struct {
	ApiKeyID   int64    `json:"api_key_id"`
	EmpresaID  int64    `json:"empresa_id"`
	TelefonoID int64    `json:"telefono_id"`
	KeyPrefix  string   `json:"key_prefix"`
	Scopes     []string `json:"scopes,omitempty"`
}

type apiKeyClaimsKey struct{}

func WithApiKeyClaims(ctx context.Context, claims *ApiKeyClaims) context.Context {
	return context.WithValue(ctx, apiKeyClaimsKey{}, claims)
}

func GetApiKeyClaims(ctx context.Context) (*ApiKeyClaims, bool) {
	claims, ok := ctx.Value(apiKeyClaimsKey{}).(*ApiKeyClaims)
	return claims, ok
}
