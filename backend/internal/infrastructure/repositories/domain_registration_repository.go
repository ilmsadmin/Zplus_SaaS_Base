package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// DomainRegistrationRepositoryImpl implements DomainRegistrationRepository
type DomainRegistrationRepositoryImpl struct {
	db *gorm.DB
}

func NewDomainRegistrationRepository(db *gorm.DB) *DomainRegistrationRepositoryImpl {
	return &DomainRegistrationRepositoryImpl{db: db}
}

func (r *DomainRegistrationRepositoryImpl) Create(ctx context.Context, registration *domain.DomainRegistration) error {
	return r.db.WithContext(ctx).Create(registration).Error
}

func (r *DomainRegistrationRepositoryImpl) GetByID(ctx context.Context, id uuid.UUID) (*domain.DomainRegistration, error) {
	var registration domain.DomainRegistration
	err := r.db.WithContext(ctx).First(&registration, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *DomainRegistrationRepositoryImpl) GetByDomain(ctx context.Context, domainName string) (*domain.DomainRegistration, error) {
	var registration domain.DomainRegistration
	// Note: This would need a join with TenantDomain table
	err := r.db.WithContext(ctx).
		Joins("JOIN tenant_domains ON domain_registrations.domain_id = tenant_domains.id").
		Where("tenant_domains.domain = ?", domainName).
		First(&registration).Error
	if err != nil {
		return nil, err
	}
	return &registration, nil
}

func (r *DomainRegistrationRepositoryImpl) GetByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*domain.DomainRegistration, error) {
	var registrations []*domain.DomainRegistration
	err := r.db.WithContext(ctx).
		Joins("JOIN tenant_domains ON domain_registrations.domain_id = tenant_domains.id").
		Where("tenant_domains.tenant_id = ?", tenantID).
		Find(&registrations).Error
	return registrations, err
}

func (r *DomainRegistrationRepositoryImpl) Update(ctx context.Context, registration *domain.DomainRegistration) error {
	return r.db.WithContext(ctx).Save(registration).Error
}

func (r *DomainRegistrationRepositoryImpl) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&domain.DomainRegistration{}, "id = ?", id).Error
}

func (r *DomainRegistrationRepositoryImpl) GetExpiringRegistrations(ctx context.Context, days int) ([]*domain.DomainRegistration, error) {
	expiryDate := time.Now().AddDate(0, 0, days)
	var registrations []*domain.DomainRegistration
	err := r.db.WithContext(ctx).Where("expiration_date <= ? AND expiration_date > ? AND auto_renew = ?",
		expiryDate, time.Now(), true).Find(&registrations).Error
	return registrations, err
}

func (r *DomainRegistrationRepositoryImpl) GetByStatus(ctx context.Context, status string) ([]*domain.DomainRegistration, error) {
	var registrations []*domain.DomainRegistration
	err := r.db.WithContext(ctx).Where("registration_status = ?", status).Find(&registrations).Error
	return registrations, err
}

func (r *DomainRegistrationRepositoryImpl) GetByProvider(ctx context.Context, provider string) ([]*domain.DomainRegistration, error) {
	var registrations []*domain.DomainRegistration
	err := r.db.WithContext(ctx).Where("registrar_provider = ?", provider).Find(&registrations).Error
	return registrations, err
}

func (r *DomainRegistrationRepositoryImpl) UpdateAutoRenew(ctx context.Context, id uuid.UUID, autoRenew bool) error {
	return r.db.WithContext(ctx).Model(&domain.DomainRegistration{}).Where("id = ?", id).Update("auto_renew", autoRenew).Error
}

func (r *DomainRegistrationRepositoryImpl) List(ctx context.Context, filter domain.DomainRegistrationFilter) ([]*domain.DomainRegistration, int64, error) {
	query := r.db.WithContext(ctx).Model(&domain.DomainRegistration{})

	// Apply filters
	if filter.TenantID != nil {
		query = query.Joins("JOIN tenant_domains ON domain_registrations.domain_id = tenant_domains.id").
			Where("tenant_domains.tenant_id = ?", *filter.TenantID)
	}
	if filter.Status != nil {
		query = query.Where("registration_status = ?", *filter.Status)
	}
	if filter.Provider != nil {
		query = query.Where("registrar_provider = ?", *filter.Provider)
	}
	if filter.Search != nil {
		query = query.Joins("JOIN tenant_domains ON domain_registrations.domain_id = tenant_domains.id").
			Where("tenant_domains.domain ILIKE ?", "%"+*filter.Search+"%")
	}
	if filter.AutoRenew != nil {
		query = query.Where("auto_renew = ?", *filter.AutoRenew)
	}

	// Count total
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination and sorting
	if filter.SortBy != "" {
		order := filter.SortBy
		if filter.SortOrder == "desc" {
			order += " DESC"
		}
		query = query.Order(order)
	} else {
		query = query.Order("created_at DESC")
	}

	offset := (filter.Page - 1) * filter.Limit
	query = query.Offset(offset).Limit(filter.Limit)

	var registrations []*domain.DomainRegistration
	err := query.Find(&registrations).Error
	return registrations, total, err
}
