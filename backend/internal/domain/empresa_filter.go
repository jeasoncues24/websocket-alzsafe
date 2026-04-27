package domain

import (
	"context"
	"strconv"
	"strings"
)

type empresaContextKey string

const ctxEmpresaID empresaContextKey = "empresa_id"

func WithEmpresaID(ctx context.Context, empresaID int64) context.Context {
	return context.WithValue(ctx, ctxEmpresaID, empresaID)
}

func GetEmpresaID(ctx context.Context) (int64, bool) {
	if v := ctx.Value(ctxEmpresaID); v != nil {
		if id, ok := v.(int64); ok {
			return id, true
		}
	}
	return 0, false
}

type EmpresaFilter struct {
	EmpresaID *int64
	IsRoot    bool
}

func GetEmpresaFilter(ctx context.Context, headerEmpresaID string) (*EmpresaFilter, bool) {
	claims, ok := GetAdminJWTClaims(ctx)
	if !ok {
		return nil, false
	}

	filter := &EmpresaFilter{
		IsRoot:    claims.IsRoot,
		EmpresaID: claims.EmpresaID,
	}

	if claims.IsRoot && headerEmpresaID != "" {
		if id, err := strconv.ParseInt(headerEmpresaID, 10, 64); err == nil && id > 0 {
			filter.EmpresaID = &id
		}
	}

	return filter, true
}

func GetEmpresaIDsFromFilter(filter *EmpresaFilter, empresaStore EmpresaStoreInterface) ([]string, error) {
	if filter == nil {
		return []string{}, nil
	}

	if filter.EmpresaID == nil {
		if filter.IsRoot {
			return []string{}, nil
		}
		return []string{}, nil
	}

	empresa, err := empresaStore.GetByID(*filter.EmpresaID)
	if err != nil {
		return nil, err
	}
	if empresa == nil {
		return []string{}, nil
	}

	return []string{empresa.RUC}, nil
}

type EmpresaStoreInterface interface {
	GetByID(id int64) (*Empresa, error)
	GetByRUC(ruc string) (*Empresa, error)
	GetAll(page, limit int, search string, activo *bool) ([]Empresa, int, error)
	Create(empresa *Empresa) (int64, error)
	Update(empresa *Empresa) error
	Delete(id int64) error
	IncrementTokenVersion(id int64) (int, error)
}

func GetAllEmpresaRUCs(empresaStore EmpresaStoreInterface) ([]string, error) {
	empresas, _, err := empresaStore.GetAll(1, 1000, "", nil)
	if err != nil {
		return nil, err
	}

	rucs := make([]string, len(empresas))
	for i, e := range empresas {
		rucs[i] = e.RUC
	}

	return rucs, nil
}

func GetRUCFromContext(ctx context.Context, filter *EmpresaFilter, empresaStore EmpresaStoreInterface) (string, error) {
	if filter == nil {
		return "", nil
	}

	if filter.IsRoot {
		headerID := ctx.Value("x-empresa-id")
		if headerID != nil {
			if id, err := strconv.ParseInt(headerID.(string), 10, 64); err == nil && id > 0 {
				empresa, err := empresaStore.GetByID(id)
				if err != nil {
					return "", err
				}
				if empresa != nil {
					return empresa.RUC, nil
				}
			}
		}
		return "", nil
	}

	if filter.EmpresaID == nil {
		return "", nil
	}

	empresa, err := empresaStore.GetByID(*filter.EmpresaID)
	if err != nil {
		return "", err
	}
	if empresa == nil {
		return "", nil
	}

	return empresa.RUC, nil
}

func GetEmpresaIDFromContext(ctx context.Context, filter *EmpresaFilter) (int64, bool) {
	if filter == nil {
		return 0, false
	}

	if filter.IsRoot {
		headerID := ctx.Value("x-empresa-id")
		if headerID != nil {
			if id, err := strconv.ParseInt(headerID.(string), 10, 64); err == nil && id > 0 {
				return id, true
			}
		}
	}

	if filter.EmpresaID != nil {
		return *filter.EmpresaID, true
	}

	return 0, false
}

func NormalizeAccountID(id string) string {
	id = strings.TrimSpace(id)
	id = strings.ToLower(id)
	if len(id) > 12 && !strings.HasSuffix(id, "@c.us") {
		id = id + "@c.us"
	}
	return id
}
