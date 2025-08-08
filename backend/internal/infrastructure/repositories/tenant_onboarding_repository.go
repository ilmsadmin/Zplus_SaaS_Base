package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// TenantOnboardingRepositoryImpl implements the TenantOnboardingRepository interface
type TenantOnboardingRepositoryImpl struct {
	db *gorm.DB
}

// NewTenantOnboardingRepository creates a new tenant onboarding repository
func NewTenantOnboardingRepository(db *gorm.DB) domain.TenantOnboardingRepository {
	return &TenantOnboardingRepositoryImpl{db: db}
}

// Create creates a new onboarding log
func (r *TenantOnboardingRepositoryImpl) Create(ctx context.Context, log *domain.TenantOnboardingLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetByID gets an onboarding log by ID
func (r *TenantOnboardingRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.TenantOnboardingLog, error) {
	var log domain.TenantOnboardingLog
	err := r.db.WithContext(ctx).First(&log, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// Update updates an onboarding log
func (r *TenantOnboardingRepositoryImpl) Update(ctx context.Context, log *domain.TenantOnboardingLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// Delete deletes an onboarding log
func (r *TenantOnboardingRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.TenantOnboardingLog{}, "id = ?", id).Error
}

// ListByTenant lists onboarding logs for a tenant
func (r *TenantOnboardingRepositoryImpl) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*domain.TenantOnboardingLog, error) {
	var logs []*domain.TenantOnboardingLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("step ASC").
		Find(&logs).Error
	return logs, err
}

// GetByTenantAndStep gets an onboarding log by tenant and step
func (r *TenantOnboardingRepositoryImpl) GetByTenantAndStep(ctx context.Context, tenantID uuid.UUID, step int) (*domain.TenantOnboardingLog, error) {
	var log domain.TenantOnboardingLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND step = ?", tenantID, step).
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// UpdateStepStatus updates the status of an onboarding step
func (r *TenantOnboardingRepositoryImpl) UpdateStepStatus(ctx context.Context, tenantID uuid.UUID, step int, status string, data map[string]interface{}) error {
	now := time.Now()

	// Try to find existing log entry
	var log domain.TenantOnboardingLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND step = ?", tenantID, step).
		First(&log).Error

	if err == gorm.ErrRecordNotFound {
		// Create new log entry
		log = domain.TenantOnboardingLog{
			ID:        uuid.New(),
			TenantID:  tenantID,
			Step:      step,
			Status:    status,
			Data:      data,
			StartedAt: &now,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if status == "completed" {
			log.CompletedAt = &now
		}

		return r.db.WithContext(ctx).Create(&log).Error
	} else if err != nil {
		return err
	} else {
		// Update existing log entry
		updates := map[string]interface{}{
			"status":     status,
			"updated_at": now,
		}

		if data != nil {
			updates["data"] = data
		}

		if status == "completed" && log.CompletedAt == nil {
			updates["completed_at"] = now
		}

		if status == "in_progress" && log.StartedAt == nil {
			updates["started_at"] = now
		}

		return r.db.WithContext(ctx).
			Model(&log).
			Updates(updates).Error
	}
}

// GetCurrentStep gets the current onboarding step for a tenant
func (r *TenantOnboardingRepositoryImpl) GetCurrentStep(ctx context.Context, tenantID uuid.UUID) (*domain.TenantOnboardingLog, error) {
	var log domain.TenantOnboardingLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND status IN (?, ?)", tenantID, "in_progress", "pending").
		Order("step ASC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// CompleteStep completes an onboarding step
func (r *TenantOnboardingRepositoryImpl) CompleteStep(ctx context.Context, tenantID uuid.UUID, step int, data map[string]interface{}) error {
	return r.UpdateStepStatus(ctx, tenantID, step, "completed", data)
}
