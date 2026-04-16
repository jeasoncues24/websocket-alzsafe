package storage

import (
	"database/sql"
	"fmt"
	"time"

	"wsapi/internal/domain"
)

type ModuleStore struct {
	db *sql.DB
}

func NewModuleStore(db *sql.DB) *ModuleStore {
	return &ModuleStore{db: db}
}

func (s *ModuleStore) Create(module *domain.Module) (int64, error) {
	query := `INSERT INTO modules (name, description, slug) VALUES (?, ?, ?)`

	result, err := s.db.Exec(query, module.Name, module.Description, module.Slug)
	if err != nil {
		return 0, fmt.Errorf("error al crear módulo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error al obtener ID: %w", err)
	}

	module.ID = id
	module.CreatedAt = time.Now()

	return id, nil
}

func (s *ModuleStore) GetByID(id int64) (*domain.Module, error) {
	query := `SELECT id, name, description, slug, created_at FROM modules WHERE id = ?`

	module := &domain.Module{}
	err := s.db.QueryRow(query, id).Scan(
		&module.ID, &module.Name, &module.Description, &module.Slug, &module.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener módulo: %w", err)
	}

	return module, nil
}

func (s *ModuleStore) GetBySlug(slug string) (*domain.Module, error) {
	query := `SELECT id, name, description, slug, created_at FROM modules WHERE slug = ?`

	module := &domain.Module{}
	err := s.db.QueryRow(query, slug).Scan(
		&module.ID, &module.Name, &module.Description, &module.Slug, &module.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error al obtener módulo: %w", err)
	}

	return module, nil
}

func (s *ModuleStore) GetAll() ([]domain.Module, error) {
	query := `SELECT id, name, description, slug, created_at FROM modules ORDER BY id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error al obtener módulos: %w", err)
	}
	defer rows.Close()

	var modules []domain.Module
	for rows.Next() {
		var m domain.Module
		err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.Slug, &m.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error al escanear módulo: %w", err)
		}
		modules = append(modules, m)
	}

	return modules, nil
}

func (s *ModuleStore) Update(module *domain.Module) error {
	query := `UPDATE modules SET name = ?, description = ? WHERE id = ?`

	_, err := s.db.Exec(query, module.Name, module.Description, module.ID)
	if err != nil {
		return fmt.Errorf("error al actualizar módulo: %w", err)
	}

	return nil
}

func (s *ModuleStore) Delete(id int64) error {
	query := `DELETE FROM modules WHERE id = ?`

	_, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error al eliminar módulo: %w", err)
	}

	return nil
}
