package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// UserSessionRepositoryImpl implements domain.UserSessionRepository
type UserSessionRepositoryImpl struct {
	db *gorm.DB
}

// NewUserSessionRepository creates a new user session repository
func NewUserSessionRepository(db *gorm.DB) domain.UserSessionRepository {
	return &UserSessionRepositoryImpl{db: db}
}

// Create creates a new user session
func (r *UserSessionRepositoryImpl) Create(ctx context.Context, session *domain.UserSession) error {
	return r.db.WithContext(ctx).Create(session).Error
}

// GetByID retrieves a user session by ID
func (r *UserSessionRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserSession, error) {
	var session domain.UserSession
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		First(&session, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByToken retrieves a user session by token
func (r *UserSessionRepositoryImpl) GetByToken(ctx context.Context, token string) (*domain.UserSession, error) {
	var session domain.UserSession
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Tenant").
		Where("session_token = ? AND expires_at > ?", token, time.Now()).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update updates a user session
func (r *UserSessionRepositoryImpl) Update(ctx context.Context, session *domain.UserSession) error {
	return r.db.WithContext(ctx).Save(session).Error
}

// Delete deletes a user session
func (r *UserSessionRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserSession{}, "id = ?", id).Error
}

// DeleteByUserID deletes all sessions for a user
func (r *UserSessionRepositoryImpl) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.UserSession{}, "user_id = ?", userID).Error
}

// DeleteExpired deletes expired sessions
func (r *UserSessionRepositoryImpl) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Delete(&domain.UserSession{}, "expires_at < ?", time.Now()).Error
}

// ListByUser retrieves sessions by user with pagination
func (r *UserSessionRepositoryImpl) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.UserSession, error) {
	var sessions []*domain.UserSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Limit(limit).
		Offset(offset).
		Order("last_accessed_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// ListByTenant retrieves sessions by tenant with pagination
func (r *UserSessionRepositoryImpl) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.UserSession, error) {
	var sessions []*domain.UserSession
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND expires_at > ?", tenantID, time.Now()).
		Limit(limit).
		Offset(offset).
		Order("last_accessed_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// CountActiveByUser counts active sessions for a user
func (r *UserSessionRepositoryImpl) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&domain.UserSession{}).
		Where("user_id = ? AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}

// UpdateLastAccessed updates the last accessed time for a session
func (r *UserSessionRepositoryImpl) UpdateLastAccessed(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).
		Model(&domain.UserSession{}).
		Where("session_token = ?", token).
		Update("last_accessed_at", time.Now()).Error
}
