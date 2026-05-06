package domain

import "context"

// PanelAccess normaliza el scope efectivo para endpoints del panel.
// - Admin JWT: acceso global a cualquier empresa del panel.
// - Empresa JWT: acceso limitado a su propia empresa.
// - IsRoot se conserva para operaciones privilegiadas independientes del scope.
type PanelAccess struct {
	EmpresaID  *int64
	IsRoot     bool
	IsAdminJWT bool
}

func GetPanelAccess(ctx context.Context) (PanelAccess, bool) {
	if claims, ok := GetAdminJWTClaims(ctx); ok && claims != nil {
		return PanelAccess{IsRoot: claims.IsRoot, IsAdminJWT: true}, true
	}
	if claims, ok := GetEmpresaJWTClaims(ctx); ok && claims != nil {
		eid := claims.EmpresaID
		return PanelAccess{EmpresaID: &eid}, true
	}
	return PanelAccess{}, false
}

func (a PanelAccess) CanAccessEmpresa(empresaID int64) bool {
	if a.IsRoot || a.IsAdminJWT {
		return true
	}
	if a.EmpresaID == nil {
		return false
	}
	return *a.EmpresaID == empresaID
}

func (a PanelAccess) CompanyID() (int64, bool) {
	if a.EmpresaID == nil || *a.EmpresaID <= 0 {
		return 0, false
	}
	return *a.EmpresaID, true
}
