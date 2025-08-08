package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// UserPreferenceRepositoryImpl implements domain.UserPreferenceRepository
type UserPreferenceRepositoryImpl struct {
	db *gorm.DB
}

// NewUserPreferenceRepository creates a new user preference repository
func NewUserPreferenceRepository(db *gorm.DB) domain.UserPreferenceRepository {
	return &UserPreferenceRepositoryImpl{db: db}
}

// Create creates a new user preference
func (r *UserPreferenceRepositoryImpl) Create(ctx context.Context, preference *domain.UserPreference) error {
	return r.db.WithContext(ctx).Create(preference).Error
}

// GetByID retrieves a user preference by ID
func (r *UserPreferenceRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserPreference, error) {
	var preference domain.UserPreference
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		First(&preference, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &preference, nil
}

// GetByUserAndKey retrieves a user preference by user and key
func (r *UserPreferenceRepositoryImpl) GetByUserAndKey(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string) (*domain.UserPreference, error) {
	var preference domain.UserPreference
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ? AND key = ?", userID, category, key)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	err := query.First(&preference).Error
	if err != nil {
		return nil, err
	}
	return &preference, nil
}

// Update updates a user preference
func (r *UserPreferenceRepositoryImpl) Update(ctx context.Context, preference *domain.UserPreference) error {
	return r.db.WithContext(ctx).Save(preference).Error
}

// Delete deletes a user preference
func (r *UserPreferenceRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserPreference{}, "id = ?", id).Error
}

// UpsertPreference creates or updates a user preference
func (r *UserPreferenceRepositoryImpl) UpsertPreference(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string, value interface{}) error {
	preference := &domain.UserPreference{
		UserID:   userID,
		TenantID: tenantID,
		Category: category,
		Key:      key,
		Value:    value,
	}

	return r.db.WithContext(ctx).
		Where("user_id = ? AND category = ? AND key = ? AND (tenant_id = ? OR (tenant_id IS NULL AND ? IS NULL))",
			userID, category, key, tenantID, tenantID).
		Assign(map[string]interface{}{
			"value":      value,
			"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
		}).
		FirstOrCreate(preference).Error
}

// ListByUser retrieves all preferences for a user
func (r *UserPreferenceRepositoryImpl) ListByUser(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*domain.UserPreference, error) {
	var preferences []*domain.UserPreference
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	err := query.Find(&preferences).Error
	return preferences, err
}

// ListByCategory retrieves preferences for a user by category
func (r *UserPreferenceRepositoryImpl) ListByCategory(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category string) ([]*domain.UserPreference, error) {
	var preferences []*domain.UserPreference
	query := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ?", userID, category)

	if tenantID != nil {
		query = query.Where("tenant_id = ?", *tenantID)
	} else {
		query = query.Where("tenant_id IS NULL")
	}

	err := query.Find(&preferences).Error
	return preferences, err
}

// DeleteByUser deletes all preferences for a user
func (r *UserPreferenceRepositoryImpl) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&domain.UserPreference{}, "user_id = ?", userID).Error
}

// GetUserPreferences retrieves all preferences for a user as a map
func (r *UserPreferenceRepositoryImpl) GetUserPreferences(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) (map[string]interface{}, error) {
	preferences, err := r.ListByUser(ctx, userID, tenantID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, pref := range preferences {
		categoryMap, exists := result[pref.Category].(map[string]interface{})
		if !exists {
			categoryMap = make(map[string]interface{})
			result[pref.Category] = categoryMap
		}
		categoryMap[pref.Key] = pref.Value
	}

	return result, nil
}
