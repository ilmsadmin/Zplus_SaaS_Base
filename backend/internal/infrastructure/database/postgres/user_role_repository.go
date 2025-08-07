package postgres

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// UserRoleRepository implements the user role repository using PostgreSQL
type UserRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(db *gorm.DB) *UserRoleRepository {
	return &UserRoleRepository{db: db}
}

// Create creates a new user role assignment
func (r *UserRoleRepository) Create(ctx context.Context, userRole *domain.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

// GetByID retrieves a user role by ID
func (r *UserRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserRole, error) {
	var userRole domain.UserRole
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Role").
		Preload("Tenant").
		First(&userRole, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &userRole, nil
}

// GetByUserAndTenant retrieves user roles by user and tenant
func (r *UserRoleRepository) GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.UserRole, error) {
	var userRoles []*domain.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Role.Permissions").
		Where("user_id = ? AND tenant_id = ?", userID, tenantID).
		Find(&userRoles).Error
	return userRoles, err
}

// Update updates a user role
func (r *UserRoleRepository) Update(ctx context.Context, userRole *domain.UserRole) error {
	return r.db.WithContext(ctx).Save(userRole).Error
}

// Delete deletes a user role by ID
func (r *UserRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserRole{}, "id = ?", id).Error
}

// DeleteByUserAndRole deletes a user role by user, role, and tenant
func (r *UserRoleRepository) DeleteByUserAndRole(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ? AND tenant_id = ?", userID, roleID, tenantID).
		Delete(&domain.UserRole{}).Error
}

// ListByUser lists user roles by user
func (r *UserRoleRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.UserRole, error) {
	var userRoles []*domain.UserRole
	err := r.db.WithContext(ctx).
		Preload("Role").
		Preload("Role.Permissions").
		Preload("Tenant").
		Where("user_id = ?", userID).
		Find(&userRoles).Error
	return userRoles, err
}

// ListByRole lists user roles by role with pagination
func (r *UserRoleRepository) ListByRole(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*domain.UserRole, error) {
	var userRoles []*domain.UserRole
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("role_id = ?", roleID).
		Limit(limit).
		Offset(offset).
		Find(&userRoles).Error
	return userRoles, err
}

// ListByTenant lists user roles by tenant with pagination
func (r *UserRoleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.UserRole, error) {
	var userRoles []*domain.UserRole
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Role").
		Where("tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Find(&userRoles).Error
	return userRoles, err
}

// CountByRole counts user roles by role
func (r *UserRoleRepository) CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	return count, err
}

// CountByTenant counts user roles by tenant
func (r *UserRoleRepository) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserRole{}).
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}
