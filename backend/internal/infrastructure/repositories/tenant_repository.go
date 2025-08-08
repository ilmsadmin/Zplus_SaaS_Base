package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"gorm.io/gorm"
)

// TenantRepositoryImpl implements the TenantRepository interface
type TenantRepositoryImpl struct {
	db *gorm.DB
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *gorm.DB) domain.TenantRepository {
	return &TenantRepositoryImpl{db: db}
}

// Basic CRUD Operations

// Create creates a new tenant
func (r *TenantRepositoryImpl) Create(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

// GetByID gets a tenant by ID
func (r *TenantRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).
		First(&tenant, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetBySubdomain gets a tenant by subdomain
func (r *TenantRepositoryImpl) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := r.db.WithContext(ctx).
		First(&tenant, "subdomain = ?", subdomain).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetByDomain gets a tenant by domain
func (r *TenantRepositoryImpl) GetByDomain(ctx context.Context, domainName string) (*domain.Tenant, error) {
	var tenantDomain domain.TenantDomain
	err := r.db.WithContext(ctx).
		First(&tenantDomain, "domain = ?", domainName).Error
	if err != nil {
		return nil, err
	}

	// Parse TenantID string to UUID
	tenantUUID, err := uuid.Parse(tenantDomain.TenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID format: %w", err)
	}

	return r.GetByID(ctx, tenantUUID)
}

// Update updates a tenant
func (r *TenantRepositoryImpl) Update(ctx context.Context, tenant *domain.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}

// Delete deletes a tenant
func (r *TenantRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.Tenant{}, "id = ?", id).Error
}

// List lists tenants with pagination
func (r *TenantRepositoryImpl) List(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	var tenants []*domain.Tenant
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tenants).Error
	return tenants, err
}

// Count counts total tenants
func (r *TenantRepositoryImpl) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Tenant{}).Count(&count).Error
	return count, err
}

// SearchTenants searches tenants by query
func (r *TenantRepositoryImpl) SearchTenants(ctx context.Context, query string, limit, offset int) ([]*domain.Tenant, error) {
	var tenants []*domain.Tenant
	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).
		Where("name ILIKE ? OR plan ILIKE ?", searchQuery, searchQuery).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tenants).Error
	return tenants, err
}

// ListByStatus lists tenants by status
func (r *TenantRepositoryImpl) ListByStatus(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error) {
	var tenants []*domain.Tenant
	err := r.db.WithContext(ctx).
		Where("status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tenants).Error
	return tenants, err
}

// ListByPlan lists tenants by plan
func (r *TenantRepositoryImpl) ListByPlan(ctx context.Context, plan string, limit, offset int) ([]*domain.Tenant, error) {
	var tenants []*domain.Tenant
	err := r.db.WithContext(ctx).
		Where("plan = ?", plan).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tenants).Error
	return tenants, err
}

// Onboarding operations

// UpdateOnboardingStatus updates tenant onboarding status
func (r *TenantRepositoryImpl) UpdateOnboardingStatus(ctx context.Context, tenantID uuid.UUID, status string, step int) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"onboarding_status": status,
			"onboarding_step":   step,
			"updated_at":        now,
		}).Error
}

// UpdateOnboardingData updates tenant onboarding data
func (r *TenantRepositoryImpl) UpdateOnboardingData(ctx context.Context, tenantID uuid.UUID, data map[string]interface{}) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"onboarding_data": data,
			"updated_at":      now,
		}).Error
}

// GetTenantsByOnboardingStatus gets tenants by onboarding status
func (r *TenantRepositoryImpl) GetTenantsByOnboardingStatus(ctx context.Context, status string, limit, offset int) ([]*domain.Tenant, error) {
	var tenants []*domain.Tenant
	err := r.db.WithContext(ctx).
		Where("onboarding_status = ?", status).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&tenants).Error
	return tenants, err
}

// Configuration management

// UpdateSettings updates tenant settings
func (r *TenantRepositoryImpl) UpdateSettings(ctx context.Context, tenantID uuid.UUID, settings *domain.TenantSettings) error {
	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the tenant's updated_at timestamp
		if err := tx.Model(&domain.Tenant{}).
			Where("id = ?", tenantID).
			Updates(map[string]interface{}{
				"settings":   settings,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdateConfiguration updates tenant configuration
func (r *TenantRepositoryImpl) UpdateConfiguration(ctx context.Context, tenantID uuid.UUID, config *domain.TenantConfiguration) error {
	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the tenant's updated_at timestamp
		if err := tx.Model(&domain.Tenant{}).
			Where("id = ?", tenantID).
			Updates(map[string]interface{}{
				"configuration": config,
				"updated_at":    now,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdateBranding updates tenant branding
func (r *TenantRepositoryImpl) UpdateBranding(ctx context.Context, tenantID uuid.UUID, branding *domain.TenantBranding) error {
	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the tenant's updated_at timestamp
		if err := tx.Model(&domain.Tenant{}).
			Where("id = ?", tenantID).
			Updates(map[string]interface{}{
				"branding":   branding,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdateBilling updates tenant billing
func (r *TenantRepositoryImpl) UpdateBilling(ctx context.Context, tenantID uuid.UUID, billing *domain.TenantBilling) error {
	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update the tenant's updated_at timestamp
		if err := tx.Model(&domain.Tenant{}).
			Where("id = ?", tenantID).
			Updates(map[string]interface{}{
				"billing":    billing,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		return nil
	})
}

// Tenant lifecycle

// ActivateTenant activates a tenant
func (r *TenantRepositoryImpl) ActivateTenant(ctx context.Context, tenantID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"status":     "active",
			"updated_at": now,
		}).Error
}

// SuspendTenant suspends a tenant
func (r *TenantRepositoryImpl) SuspendTenant(ctx context.Context, tenantID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"status":     "suspended",
			"updated_at": now,
		}).Error
}

// CancelTenant cancels a tenant
func (r *TenantRepositoryImpl) CancelTenant(ctx context.Context, tenantID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"status":     "cancelled",
			"updated_at": now,
		}).Error
}

// Subdomain management

// IsSubdomainAvailable checks if a subdomain is available
func (r *TenantRepositoryImpl) IsSubdomainAvailable(ctx context.Context, subdomain string) (bool, error) {
	var count int64

	// Check both subdomain field in tenants and domain field in tenant_domains
	err := r.db.WithContext(ctx).
		Model(&domain.Tenant{}).
		Where("subdomain = ?", subdomain).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	if count > 0 {
		return false, nil
	}

	err = r.db.WithContext(ctx).
		Model(&domain.TenantDomain{}).
		Where("domain = ?", subdomain+".example.com").
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count == 0, nil
}

// ReserveSubdomain reserves a subdomain for a tenant
func (r *TenantRepositoryImpl) ReserveSubdomain(ctx context.Context, subdomain string, tenantID uuid.UUID) error {
	tenantDomain := &domain.TenantDomain{
		ID:         uuid.New(),
		TenantID:   tenantID.String(),
		Domain:     subdomain + ".example.com", // Adjust domain suffix as needed
		IsCustom:   false,
		Verified:   true,
		IsVerified: true,
		IsPrimary:  false,
		SSLEnabled: true,
		Status:     "active",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	return r.db.WithContext(ctx).Create(tenantDomain).Error
}

// ReleaseSubdomain releases a subdomain
func (r *TenantRepositoryImpl) ReleaseSubdomain(ctx context.Context, subdomain string) error {
	return r.db.WithContext(ctx).
		Delete(&domain.TenantDomain{}, "domain LIKE ?", subdomain+"%").Error
}

// Metrics and analytics

// GetTenantMetrics gets detailed tenant metrics
func (r *TenantRepositoryImpl) GetTenantMetrics(ctx context.Context, tenantID uuid.UUID, metricType string, from, to time.Time) ([]*domain.TenantUsageMetrics, error) {
	var metrics []*domain.TenantUsageMetrics
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND metric_type = ? AND recorded_at BETWEEN ? AND ?", tenantID, metricType, from, to).
		Order("recorded_at DESC").
		Find(&metrics).Error
	return metrics, err
}

// RecordUsageMetric records a usage metric
func (r *TenantRepositoryImpl) RecordUsageMetric(ctx context.Context, metric *domain.TenantUsageMetrics) error {
	return r.db.WithContext(ctx).Create(metric).Error
}

// GetTenantStats gets basic tenant statistics
func (r *TenantRepositoryImpl) GetTenantStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get user count
	var userCount int64
	if err := r.db.WithContext(ctx).
		Table("tenant_users").
		Where("tenant_id = ?", tenantID).
		Count(&userCount).Error; err == nil {
		stats["user_count"] = int(userCount)
	}

	// Get domain count
	var domainCount int64
	if err := r.db.WithContext(ctx).
		Model(&domain.TenantDomain{}).
		Where("tenant_id = ?", tenantID).
		Count(&domainCount).Error; err == nil {
		stats["domain_count"] = int(domainCount)
	}

	// Additional stats can be added here based on available data
	stats["storage_used"] = int64(0)  // Placeholder
	stats["api_calls_this_month"] = 0 // Placeholder
	stats["uploaded_files"] = 0       // Placeholder

	return stats, nil
}
