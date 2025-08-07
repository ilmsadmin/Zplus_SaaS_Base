package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

type TenantService struct {
	tenantRepo       domain.TenantRepository
	tenantDomainRepo domain.TenantDomainRepository
	auditLogRepo     domain.AuditLogRepository
}

func NewTenantService(
	tenantRepo domain.TenantRepository,
	tenantDomainRepo domain.TenantDomainRepository,
	auditLogRepo domain.AuditLogRepository,
) *TenantService {
	return &TenantService{
		tenantRepo:       tenantRepo,
		tenantDomainRepo: tenantDomainRepo,
		auditLogRepo:     auditLogRepo,
	}
}

func (s *TenantService) CreateTenant(ctx context.Context, tenant *domain.Tenant) error {
	// Validate subdomain uniqueness
	existingTenant, err := s.tenantRepo.GetBySubdomain(ctx, tenant.Subdomain)
	if err == nil && existingTenant != nil {
		return fmt.Errorf("subdomain %s already exists", tenant.Subdomain)
	}

	// Create tenant
	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		TenantID: tenant.ID,
		Action:   domain.ActionCreate,
		Resource: "tenant",
		Details: &domain.Details{
			After: map[string]interface{}{
				"name":      tenant.Name,
				"subdomain": tenant.Subdomain,
				"status":    tenant.Status,
			},
		},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (s *TenantService) GetTenant(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	return s.tenantRepo.GetByID(ctx, id)
}

func (s *TenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*domain.Tenant, error) {
	return s.tenantRepo.GetBySubdomain(ctx, subdomain)
}

func (s *TenantService) UpdateTenant(ctx context.Context, tenant *domain.Tenant) error {
	// Get existing tenant for audit
	existingTenant, err := s.tenantRepo.GetByID(ctx, tenant.ID)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Update tenant
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		TenantID: tenant.ID,
		Action:   domain.ActionUpdate,
		Resource: "tenant",
		Details: &domain.Details{
			Before: map[string]interface{}{
				"name":      existingTenant.Name,
				"subdomain": existingTenant.Subdomain,
				"status":    existingTenant.Status,
			},
			After: map[string]interface{}{
				"name":      tenant.Name,
				"subdomain": tenant.Subdomain,
				"status":    tenant.Status,
			},
		},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (s *TenantService) DeleteTenant(ctx context.Context, id uuid.UUID) error {
	// Get tenant for audit
	tenant, err := s.tenantRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("tenant not found: %w", err)
	}

	// Delete tenant
	if err := s.tenantRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		TenantID: id,
		Action:   domain.ActionDelete,
		Resource: "tenant",
		Details: &domain.Details{
			Before: map[string]interface{}{
				"name":      tenant.Name,
				"subdomain": tenant.Subdomain,
				"status":    tenant.Status,
			},
		},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return nil
}

func (s *TenantService) ListTenants(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	return s.tenantRepo.List(ctx, limit, offset)
}

func (s *TenantService) AddCustomDomain(ctx context.Context, tenantID uuid.UUID, domainName string) (*domain.TenantDomain, error) {
	// Check if domain already exists
	existingDomain, err := s.tenantDomainRepo.GetByDomain(ctx, domainName)
	if err == nil && existingDomain != nil {
		return nil, fmt.Errorf("domain %s already exists", domainName)
	}

	// Create domain
	domain := &domain.TenantDomain{
		TenantID:   tenantID,
		Domain:     domainName,
		IsVerified: false,
		IsPrimary:  false,
	}

	if err := s.tenantDomainRepo.Create(ctx, domain); err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		TenantID: tenantID,
		Action:   domain.ActionCreate,
		Resource: "tenant_domain",
		Details: &domain.Details{
			After: map[string]interface{}{
				"domain":      domainName,
				"is_verified": false,
				"is_primary":  false,
			},
		},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return domain, nil
}
