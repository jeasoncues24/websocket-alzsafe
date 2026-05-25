package domain

import "context"

// QRLinkClaims representa los claims del token provisional de QR link (10 min).
type QRLinkClaims struct {
	EmpresaID int64  `json:"sub"`
	PhoneID   int64  `json:"phone_id"`
	Scope     string `json:"scope"` // siempre "qr_link"
}

type qrLinkClaimsKey struct{}

func WithQRLinkClaims(ctx context.Context, c *QRLinkClaims) context.Context {
	return context.WithValue(ctx, qrLinkClaimsKey{}, c)
}

func GetQRLinkClaims(ctx context.Context) (*QRLinkClaims, bool) {
	c, ok := ctx.Value(qrLinkClaimsKey{}).(*QRLinkClaims)
	return c, ok
}
