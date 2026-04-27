package domain

import (
	"context"
	"errors"
	"fmt"
)

// UserServiceUserStore interface
type UserServiceUserStore interface {
	GetByID(id int64) (*AdminUser, error)
	GetAll(page, limit int) ([]AdminUser, int, error)
	GetByUsername(username string) (*AdminUser, error)
	Create(user *AdminUser) (int64, error)
	Update(user *AdminUser) error
	Delete(id int64) error
}

type UserServiceRoleStore interface {
	GetByID(id int64) (*Role, error)
	GetByName(name string) (*Role, error)
	GetRootRole() (*Role, error)
	GetAll() ([]Role, error)
}

type UserServiceModuleStore interface {
	GetAll() ([]Module, error)
	GetByID(id int64) (*Module, error)
}

type UserServiceUserModuleStore interface {
	GetByUserID(userID int64) ([]Module, error)
	AssignModules(userID int64, moduleIDs []int64) error
	DeleteByUserID(userID int64) error
}

// UserService maneja la lógica de negocio de usuarios
type UserService struct {
	userStore    UserServiceUserStore
	roleStore    UserServiceRoleStore
	moduleStore  UserServiceModuleStore
	userModStore UserServiceUserModuleStore
}

func NewUserService(
	userStore UserServiceUserStore,
	roleStore UserServiceRoleStore,
	moduleStore UserServiceModuleStore,
	userModStore UserServiceUserModuleStore,
) *UserService {
	return &UserService{
		userStore:    userStore,
		roleStore:    roleStore,
		moduleStore:  moduleStore,
		userModStore: userModStore,
	}
}

// CreateUser crea un nuevo usuario con rol asignado
func (us *UserService) CreateUser(ctx context.Context, username, passwordHash, email string, roleID *int64) (*AdminUser, error) {
	// Verificar que el username no exista
	existing, err := us.userStore.GetByUsername(username)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("el nombre de usuario ya existe")
	}

	// Validar rol si se proporciona
	if roleID != nil {
		role, err := us.roleStore.GetByID(*roleID)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, errors.New("el rol no existe")
		}
	}

	user := &AdminUser{
		Username:     username,
		PasswordHash: passwordHash,
		Email:      email,
		RoleID:     roleID,
		Activo:     true,
	}

	id, err := us.userStore.Create(user)
	if err != nil {
		return nil, fmt.Errorf("error al crear usuario: %w", err)
	}

	user.ID = id
	return user, nil
}

// UpdateUser actualiza un usuario existente (IsRoot se deriva del rol)
func (us *UserService) UpdateUser(ctx context.Context, userID int64, email *string, roleID *int64, activo *bool) (*AdminUser, error) {
	user, err := us.userStore.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("usuario no encontrado")
	}

	// Validar rol si se proporciona
	if roleID != nil {
		role, err := us.roleStore.GetByID(*roleID)
		if err != nil {
			return nil, err
		}
		if role == nil {
			return nil, errors.New("el rol no existe")
		}
		user.RoleID = roleID
	}

	if email != nil {
		user.Email = *email
	}
	if activo != nil {
		user.Activo = *activo
	}

	err = us.userStore.Update(user)
	if err != nil {
		return nil, fmt.Errorf("error al actualizar usuario: %w", err)
	}

	return user, nil
}

// DeleteUser elimina un usuario (soft delete)
func (us *UserService) DeleteUser(ctx context.Context, userID int64) error {
	user, err := us.userStore.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("usuario no encontrado")
	}

	// No permitir auto-eliminarse si es el último root
	if user.IsRoot {
		roots, err := us.GetActiveRoots()
		if err != nil {
			return err
		}
		if len(roots) <= 1 {
			return errors.New("no se puede eliminar el último usuario root")
		}
	}

	return us.userStore.Delete(userID)
}

// PromoteToRoot promocional un usuario al rol root
func (us *UserService) PromoteToRoot(ctx context.Context, targetUserID int64, requestingUserID int64) error {
	// 1. Verificar que el solicitante es root
	requester, err := us.userStore.GetByID(requestingUserID)
	if err != nil {
		return fmt.Errorf("error al verificar solicitante: %w", err)
	}
	if requester == nil {
		return errors.New("solicitante no encontrado")
	}
	if !requester.IsRoot {
		return errors.New("solo un usuario root puede promover a otro usuario a root")
	}

	// 2. Obtener el rol root
	rootRole, err := us.roleStore.GetRootRole()
	if err != nil {
		return err
	}
	if rootRole == nil {
		return errors.New("rol root no encontrado en el sistema")
	}

	// 3. Obtener usuario objetivo
	targetUser, err := us.userStore.GetByID(targetUserID)
	if err != nil {
		return err
	}
	if targetUser == nil {
		return errors.New("usuario objetivo no encontrado")
	}

	// 4. Si el usuario objetivo ya es root, no hacer nada
	if targetUser.IsRoot {
		return nil
	}

	// 5. Actualizar usuario a root (IsRoot se deriva del rol)
	targetUser.RoleID = &rootRole.ID
	err = us.userStore.Update(targetUser)
	if err != nil {
		return fmt.Errorf("error al promover usuario a root: %w", err)
	}

	return nil
}

// GetActiveRoots devuelve todos los usuarios root activos
func (us *UserService) GetActiveRoots() ([]AdminUser, error) {
	users, _, err := us.userStore.GetAll(1, 1000)
	if err != nil {
		return nil, err
	}

	var roots []AdminUser
	for _, u := range users {
		if u.IsRoot {
			roots = append(roots, u)
		}
	}

	return roots, nil
}

// AssignModules asigna módulos específicos a un usuario (override)
func (us *UserService) AssignModules(ctx context.Context, userID int64, moduleIDs []int64) error {
	user, err := us.userStore.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("usuario no encontrado")
	}

	// Validar que los módulos existan
	for _, moduleID := range moduleIDs {
		module, err := us.moduleStore.GetByID(moduleID)
		if err != nil {
			return err
		}
		if module == nil {
			return fmt.Errorf("módulo con ID %d no encontrado", moduleID)
		}
	}

	return us.userModStore.AssignModules(userID, moduleIDs)
}

// GetUserWithModules devuelve un usuario con sus módulos asignados
func (us *UserService) GetUserWithModules(ctx context.Context, userID int64) (*AdminUser, []Module, error) {
	user, err := us.userStore.GetByID(userID)
	if err != nil {
		return nil, nil, err
	}
	if user == nil {
		return nil, nil, errors.New("usuario no encontrado")
	}

	var modules []Module

	// Si es root, devolver todos los módulos
	if user.IsRoot {
		allModules, err := us.moduleStore.GetAll()
		if err != nil {
			return nil, nil, err
		}
		return user, allModules, nil
	}

	// Obtener módulos del usuario (override)
	modules, err = us.userModStore.GetByUserID(userID)
	if err != nil {
		return nil, nil, err
	}

	return user, modules, nil
}

// GetAllUsers devuelve todos los usuarios con paginación
func (us *UserService) GetAllUsers(ctx context.Context, page, limit int) ([]AdminUser, int, error) {
	return us.userStore.GetAll(page, limit)
}

// GetAllRoles devuelve todos los roles disponibles
func (us *UserService) GetAllRoles(ctx context.Context) ([]Role, error) {
	return us.roleStore.GetAll()
}

// GetAllModules devuelve todos los módulos disponibles
func (us *UserService) GetAllModules(ctx context.Context) ([]Module, error) {
	return us.moduleStore.GetAll()
}
