package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// TenantInvitationRepositoryImpl implements the TenantInvitationRepository interface
type TenantInvitationRepositoryImpl struct {
	db *gorm.DB
}

// NewTenantInvitationRepository creates a new tenant invitation repository
func NewTenantInvitationRepository(db *gorm.DB) domain.TenantInvitationRepository {
	return &TenantInvitationRepositoryImpl{db: db}
}

// Create creates a new invitation
func (r *TenantInvitationRepositoryImpl) Create(ctx context.Context, invitation *domain.TenantInvitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// GetByID gets an invitation by ID
func (r *TenantInvitationRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.TenantInvitation, error) {
	var invitation domain.TenantInvitation
	err := r.db.WithContext(ctx).First(&invitation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

// GetByToken gets an invitation by token
func (r *TenantInvitationRepositoryImpl) GetByToken(ctx context.Context, token string) (*domain.TenantInvitation, error) {
	var invitation domain.TenantInvitation
	err := r.db.WithContext(ctx).First(&invitation, "token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

// Update updates an invitation
func (r *TenantInvitationRepositoryImpl) Update(ctx context.Context, invitation *domain.TenantInvitation) error {
	return r.db.WithContext(ctx).Save(invitation).Error
}

// Delete deletes an invitation
func (r *TenantInvitationRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.TenantInvitation{}, "id = ?", id).Error
}

// ListByTenant lists invitations for a tenant
func (r *TenantInvitationRepositoryImpl) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*domain.TenantInvitation, error) {
	var invitations []*domain.TenantInvitation
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// ListByEmail lists invitations for an email
func (r *TenantInvitationRepositoryImpl) ListByEmail(ctx context.Context, email string) ([]*domain.TenantInvitation, error) {
	var invitations []*domain.TenantInvitation
	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

// AcceptInvitation accepts an invitation
func (r *TenantInvitationRepositoryImpl) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.TenantInvitation{}).
		Where("token = ? AND status = ? AND expires_at > ?", token, "pending", now).
		Updates(map[string]interface{}{
			"status":      "accepted",
			"accepted_by": userID,
			"accepted_at": now,
			"updated_at":  now,
		}).Error
}

// RevokeInvitation revokes an invitation
func (r *TenantInvitationRepositoryImpl) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.TenantInvitation{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     "revoked",
			"updated_at": now,
		}).Error
}

// CleanupExpiredInvitations cleans up expired invitations
func (r *TenantInvitationRepositoryImpl) CleanupExpiredInvitations(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.TenantInvitation{}).
		Where("status = ? AND expires_at < ?", "pending", now).
		Updates(map[string]interface{}{
			"status":     "expired",
			"updated_at": now,
		}).Error
}
