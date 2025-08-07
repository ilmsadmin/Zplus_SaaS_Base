package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// PermissionRepository implements the permission repository using PostgreSQL
type PermissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a new permission repository
func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

// Create creates a new permission
func (r *PermissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Create(permission).Error
}

// GetByID retrieves a permission by ID
func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).First(&permission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// GetByName retrieves a permission by name
func (r *PermissionRepository) GetByName(ctx context.Context, name string) (*domain.Permission, error) {
	var permission domain.Permission
	err := r.db.WithContext(ctx).First(&permission, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// Update updates a permission
func (r *PermissionRepository) Update(ctx context.Context, permission *domain.Permission) error {
	return r.db.WithContext(ctx).Save(permission).Error
}

// Delete deletes a permission
func (r *PermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Permission{}, "id = ?", id).Error
}

// List lists permissions with pagination
func (r *PermissionRepository) List(ctx context.Context, limit, offset int) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&permissions).Error
	return permissions, err
}

// ListByResource lists permissions by resource
func (r *PermissionRepository) ListByResource(ctx context.Context, resource string) ([]*domain.Permission, error) {
	var permissions []*domain.Permission
	err := r.db.WithContext(ctx).
		Where("resource = ?", resource).
		Find(&permissions).Error
	return permissions, err
}

// Count counts permissions
func (r *PermissionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Permission{}).Count(&count).Error
	return count, err
}
