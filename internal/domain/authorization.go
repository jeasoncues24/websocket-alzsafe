package domain

import (
	"context"
	"errors"
)

// AuthorizationUserStore interface for authorization service
type AuthorizationUserStore interface {
	GetByID(id int64) (*AdminUser, error)
}

// AuthorizationModuleStore interface for authorization service
type AuthorizationModuleStore interface {
	GetAll() ([]Module, error)
	GetByID(id int64) (*Module, error)
}

// AuthorizationUserModuleStore interface for authorization service
type AuthorizationUserModuleStore interface {
	GetByUserID(userID int64) ([]Module, error)
	HasModuleAccess(userID int64, moduleSlug string) (bool, error)
}

// AuthorizationService maneja la autorización de acceso a módulos
type AuthorizationService struct {
	userStore    AuthorizationUserStore
	moduleStore  AuthorizationModuleStore
	userModStore AuthorizationUserModuleStore
}

func NewAuthorizationService(
	userStore AuthorizationUserStore,
	moduleStore AuthorizationModuleStore,
	userModStore AuthorizationUserModuleStore,
) *AuthorizationService {
	return &AuthorizationService{
		userStore:    userStore,
		moduleStore:  moduleStore,
		userModStore: userModStore,
	}
}

// CanAccess verifica si el usuario puede acceder a un módulo específico
func (as *AuthorizationService) CanAccess(ctx context.Context, moduleSlug string) bool {
	claims, ok := GetTokenClaims(ctx)
	if !ok {
		return false
	}

	// ROOT BYPASS: acceso total sin verificar módulos
	if claims.IsRoot {
		return true
	}

	// Consultar permisos del usuario para el módulo
	hasAccess, err := as.userModStore.HasModuleAccess(claims.UserID, moduleSlug)
	if err != nil {
		return false
	}

	return hasAccess
}

// GetUserModules devuelve los módulos accesibles por el usuario
func (as *AuthorizationService) GetUserModules(ctx context.Context) ([]Module, error) {
	claims, ok := GetTokenClaims(ctx)
	if !ok {
		return nil, errors.New("claims no encontrados en contexto")
	}

	// ROOT BYPASS: devolver todos los módulos
	if claims.IsRoot {
		return as.moduleStore.GetAll()
	}

	// Obtener módulos del usuario (override)
	userModules, err := as.userModStore.GetByUserID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// Si hay override a nivel de usuario, devolver esos
	if len(userModules) > 0 {
		return userModules, nil
	}

	// Por defecto, usuario sin módulos asignados no tiene acceso
	return []Module{}, nil
}

// RequireModule crea un middleware que verifica acceso a un módulo
func (as *AuthorizationService) RequireModule(moduleSlug string) func(ctx context.Context) bool {
	return func(ctx context.Context) bool {
		return as.CanAccess(ctx, moduleSlug)
	}
}
