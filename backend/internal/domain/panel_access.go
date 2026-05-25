package domain

import "context"

// PanelAccess normaliza el scope efectivo para endpoints del panel.
// - Admin JWT: acceso global a cualquier empresa del panel.
// - IsRoot se conserva para operaciones privilegiadas independientes del scope.
type PanelAccess struct {
	IsRoot     bool
	IsAdminJWT bool
}

func GetPanelAccess(ctx context.Context) (PanelAccess, bool) {
	if claims, ok := GetAdminJWTClaims(ctx); ok && claims != nil {
		return PanelAccess{IsRoot: claims.IsRoot, IsAdminJWT: true}, true
	}
	return PanelAccess{}, false
}

func (a PanelAccess) CanAccessEmpresa(empresaID int64) bool {
	return a.IsRoot || a.IsAdminJWT
}
