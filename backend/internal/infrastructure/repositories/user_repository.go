package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// UserRepositoryImpl implements domain.UserRepository
type UserRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) domain.UserRepository {
	return &UserRepositoryImpl{db: db}
}

// Create creates a new user
func (r *UserRepositoryImpl) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by ID
func (r *UserRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		Preload("Sessions").
		First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepositoryImpl) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		First(&user, "email = ?", email).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepositoryImpl) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		First(&user, "username = ?", username).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByKeycloakUserID retrieves a user by Keycloak user ID
func (r *UserRepositoryImpl) GetByKeycloakUserID(ctx context.Context, keycloakUserID string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		First(&user, "keycloak_user_id = ?", keycloakUserID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete hard deletes a user
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&domain.User{}, "id = ?", id).Error
}

// SoftDelete soft deletes a user
func (r *UserRepositoryImpl) SoftDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.User{}, "id = ?", id).Error
}

// List retrieves users with pagination
func (r *UserRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// ListByTenant retrieves users by tenant with pagination
func (r *UserRepositoryImpl) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		Joins("JOIN tenant_users ON users.id = tenant_users.user_id").
		Where("tenant_users.tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Order("users.created_at DESC").
		Find(&users).Error
	return users, err
}

// ListByRole retrieves users by role with pagination
func (r *UserRepositoryImpl) ListByRole(ctx context.Context, role string, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("roles.name = ?", role).
		Limit(limit).
		Offset(offset).
		Order("users.created_at DESC").
		Find(&users).Error
	return users, err
}

// ListSystemAdmins retrieves system admin users
func (r *UserRepositoryImpl) ListSystemAdmins(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	return r.ListByRole(ctx, domain.RoleSystemAdmin, limit, offset)
}

// ListTenantAdmins retrieves tenant admin users for a specific tenant
func (r *UserRepositoryImpl) ListTenantAdmins(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		Joins("JOIN tenant_users ON users.id = tenant_users.user_id").
		Where("tenant_users.tenant_id = ? AND tenant_users.role = ?", tenantID, domain.RoleTenantAdmin).
		Limit(limit).
		Offset(offset).
		Order("users.created_at DESC").
		Find(&users).Error
	return users, err
}

// Search searches users by query
func (r *UserRepositoryImpl) Search(ctx context.Context, query string, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	searchPattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ? OR username ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	return users, err
}

// SearchByTenant searches users by tenant and query
func (r *UserRepositoryImpl) SearchByTenant(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*domain.User, error) {
	var users []*domain.User
	searchPattern := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Preload("TenantUsers").
		Preload("UserRoles").
		Joins("JOIN tenant_users ON users.id = tenant_users.user_id").
		Where("tenant_users.tenant_id = ?", tenantID).
		Where("users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.email ILIKE ? OR users.username ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern).
		Limit(limit).
		Offset(offset).
		Order("users.created_at DESC").
		Find(&users).Error
	return users, err
}

// Count counts all users
func (r *UserRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&count).Error
	return count, err
}

// CountByTenant counts users by tenant
func (r *UserRepositoryImpl) CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Joins("JOIN tenant_users ON users.id = tenant_users.user_id").
		Where("tenant_users.tenant_id = ?", tenantID).
		Count(&count).Error
	return count, err
}

// CountByRole counts users by role
func (r *UserRepositoryImpl) CountByRole(ctx context.Context, role string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.User{}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id").
		Where("roles.name = ?", role).
		Count(&count).Error
	return count, err
}

// UpdateProfile updates user profile fields
func (r *UserRepositoryImpl) UpdateProfile(ctx context.Context, userID uuid.UUID, profile map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(profile).Error
}

// UpdateAvatar updates user avatar URL
func (r *UserRepositoryImpl) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("avatar_url", avatarURL).Error
}

// UpdatePreferences updates user preferences
func (r *UserRepositoryImpl) UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences map[string]interface{}) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("preferences", preferences).Error
}

// UpdateLastLogin updates user's last login time
func (r *UserRepositoryImpl) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"last_login_at": now,
			"login_count":   gorm.Expr("login_count + 1"),
		}).Error
}

// IncrementLoginCount increments user's login count
func (r *UserRepositoryImpl) IncrementLoginCount(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("login_count", gorm.Expr("login_count + 1")).Error
}

// UpdateStatus updates user status
func (r *UserRepositoryImpl) UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error {
	return r.db.WithContext(ctx).
		Model(&domain.User{}).
		Where("id = ?", userID).
		Update("status", status).Error
}

// ActivateUser activates a user
func (r *UserRepositoryImpl) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	return r.UpdateStatus(ctx, userID, domain.StatusActive)
}

// DeactivateUser deactivates a user
func (r *UserRepositoryImpl) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	return r.UpdateStatus(ctx, userID, domain.StatusInactive)
}

// SuspendUser suspends a user
func (r *UserRepositoryImpl) SuspendUser(ctx context.Context, userID uuid.UUID) error {
	return r.UpdateStatus(ctx, userID, domain.StatusSuspended)
}
