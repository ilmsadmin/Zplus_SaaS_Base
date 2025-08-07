package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// RoleRepository implements the role repository using PostgreSQL
type RoleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// Create creates a new role
func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// GetByName retrieves a role by name and tenant ID
func (r *RoleRepository) GetByName(ctx context.Context, name string, tenantID *uuid.UUID) (*domain.Role, error) {
	var role domain.Role
	query := r.db.WithContext(ctx).Preload("Permissions")

	if tenantID == nil {
		// System role
		query = query.Where("name = ? AND tenant_id IS NULL", name)
	} else {
		// Tenant role
		query = query.Where("name = ? AND tenant_id = ?", name, *tenantID)
	}

	err := query.First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

// Update updates a role
func (r *RoleRepository) Update(ctx context.Context, role *domain.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

// Delete deletes a role (only if not system role)
func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ? AND is_system = ?", id, false).
		Delete(&domain.Role{}).Error
}

// List lists roles with pagination
func (r *RoleRepository) List(ctx context.Context, tenantID *uuid.UUID, limit, offset int) ([]*domain.Role, error) {
	var roles []*domain.Role
	query := r.db.WithContext(ctx).Preload("Permissions")

	if tenantID == nil {
		// List system roles
		query = query.Where("tenant_id IS NULL")
	} else {
		// List tenant roles
		query = query.Where("tenant_id = ?", *tenantID)
	}

	err := query.Limit(limit).Offset(offset).Find(&roles).Error
	return roles, err
}

// ListSystemRoles lists all system roles
func (r *RoleRepository) ListSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	var roles []*domain.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("tenant_id IS NULL AND is_system = ?", true).
		Find(&roles).Error
	return roles, err
}

// ListTenantRoles lists all roles for a specific tenant
func (r *RoleRepository) ListTenantRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	var roles []*domain.Role
	err := r.db.WithContext(ctx).
		Preload("Permissions").
		Where("tenant_id = ?", tenantID).
		Find(&roles).Error
	return roles, err
}

// Count counts roles
func (r *RoleRepository) Count(ctx context.Context, tenantID *uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&domain.Role{})

	if tenantID == nil {
		query = query.Where("tenant_id IS NULL")
	} else {
		query = query.Where("tenant_id = ?", *tenantID)
	}

	err := query.Count(&count).Error
	return count, err
}
