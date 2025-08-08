package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// DomainManagementService handles all domain-related operations
type DomainManagementService struct {
	domainRegistrationRepo domain.DomainRegistrationRepository
	dnsRecordRepo          domain.DNSRecordRepository
	domainOwnershipRepo    domain.DomainOwnershipVerificationRepository
	sslCertRequestRepo     domain.SSLCertificateRequestRepository
	domainHealthCheckRepo  domain.DomainHealthCheckRepository
	domainEventRepo        domain.DomainRegistrationEventRepository
	tenantDomainRepo       domain.TenantDomainRepository
}

func NewDomainManagementService(
	domainRegistrationRepo domain.DomainRegistrationRepository,
	dnsRecordRepo domain.DNSRecordRepository,
	domainOwnershipRepo domain.DomainOwnershipVerificationRepository,
	sslCertRequestRepo domain.SSLCertificateRequestRepository,
	domainHealthCheckRepo domain.DomainHealthCheckRepository,
	domainEventRepo domain.DomainRegistrationEventRepository,
	tenantDomainRepo domain.TenantDomainRepository,
) *DomainManagementService {
	return &DomainManagementService{
		domainRegistrationRepo: domainRegistrationRepo,
		dnsRecordRepo:          dnsRecordRepo,
		domainOwnershipRepo:    domainOwnershipRepo,
		sslCertRequestRepo:     sslCertRequestRepo,
		domainHealthCheckRepo:  domainHealthCheckRepo,
		domainEventRepo:        domainEventRepo,
		tenantDomainRepo:       tenantDomainRepo,
	}
}

// Domain Registration Operations

func (s *DomainManagementService) CheckDomainAvailability(ctx context.Context, req *DomainAvailabilityRequest) (*DomainAvailabilityResponse, error) {
	// TODO: Integrate with domain registrar APIs (Namecheap, GoDaddy, etc.)
	// For now, return a mock response

	domain := strings.ToLower(req.Domain)

	// Basic domain validation
	if !isValidDomain(domain) {
		return &DomainAvailabilityResponse{
			Domain:    domain,
			Available: false,
			Reason:    "Invalid domain format",
		}, nil
	}

	// Check if domain already registered in our system
	existingDomain, err := s.tenantDomainRepo.GetByDomain(ctx, domain)
	if err == nil && existingDomain != nil {
		return &DomainAvailabilityResponse{
			Domain:    domain,
			Available: false,
			Reason:    "Domain already registered in our system",
		}, nil
	}

	// Mock external availability check
	return &DomainAvailabilityResponse{
		Domain:    domain,
		Available: true,
		Price:     12.99,
		Currency:  "USD",
		Premium:   false,
	}, nil
}

func (s *DomainManagementService) RegisterDomain(ctx context.Context, req *RegisterDomainRequest) (*DomainRegistrationResponse, error) {
	// Check if domain is already registered
	existingDomain, err := s.tenantDomainRepo.GetByDomain(ctx, req.Domain)
	if err == nil && existingDomain != nil {
		return nil, fmt.Errorf("domain %s is already registered", req.Domain)
	}

	// Create tenant domain first
	tenantDomain := &domain.TenantDomain{
		ID:                 uuid.New(),
		TenantID:           req.TenantID.String(),
		Domain:             req.Domain,
		IsCustom:           true,
		Verified:           false,
		IsPrimary:          false,
		SSLEnabled:         false,
		VerificationMethod: "dns_txt", // Default verification method
		Status:             "pending_registration",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.tenantDomainRepo.Create(ctx, tenantDomain); err != nil {
		return nil, fmt.Errorf("failed to create tenant domain: %w", err)
	}

	// Create domain registration
	registration := &domain.DomainRegistration{
		ID:                 uuid.New(),
		DomainID:           tenantDomain.ID,
		RegistrarProvider:  req.RegistrarProvider,
		RegistrationStatus: "pending",
		AutoRenew:          req.AutoRenew,
		RegistrationPrice:  0, // Will be updated after actual registration
		RenewalPrice:       0, // Will be updated after actual registration
		Currency:           "USD",
		NameServers:        req.NameServers,
		ContactInfo:        req.ContactInfo,
		PrivacyProtection:  req.PrivacyProtection,
		TransferLock:       req.TransferLock,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.domainRegistrationRepo.Create(ctx, registration); err != nil {
		return nil, fmt.Errorf("failed to create domain registration: %w", err)
	}

	// Log registration event
	event := &domain.DomainRegistrationEvent{
		ID:             uuid.New(),
		RegistrationID: registration.ID,
		EventType:      "registration_initiated",
		EventData: map[string]interface{}{
			"registrar_provider": req.RegistrarProvider,
			"auto_renew":         req.AutoRenew,
			"privacy_protection": req.PrivacyProtection,
		},
		CreatedAt: time.Now(),
	}

	if err := s.domainEventRepo.Create(ctx, event); err != nil {
		// Log error but don't fail the registration
		fmt.Printf("Failed to create registration event: %v\n", err)
	}

	// TODO: Initiate actual domain registration with registrar API
	// For now, return the pending registration

	return s.mapDomainRegistrationResponse(registration), nil
}

func (s *DomainManagementService) GetDomainRegistration(ctx context.Context, id uuid.UUID) (*DomainRegistrationResponse, error) {
	registration, err := s.domainRegistrationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain registration: %w", err)
	}

	return s.mapDomainRegistrationResponse(registration), nil
}

func (s *DomainManagementService) UpdateDomainRegistration(ctx context.Context, id uuid.UUID, req *UpdateDomainRegistrationRequest) (*DomainRegistrationResponse, error) {
	registration, err := s.domainRegistrationRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain registration: %w", err)
	}

	// Update fields
	if req.AutoRenew != nil {
		registration.AutoRenew = *req.AutoRenew
	}
	if req.PrivacyProtection != nil {
		registration.PrivacyProtection = *req.PrivacyProtection
	}
	if req.TransferLock != nil {
		registration.TransferLock = *req.TransferLock
	}
	if req.NameServers != nil {
		registration.NameServers = req.NameServers
	}
	if req.ContactInfo != nil {
		registration.ContactInfo = req.ContactInfo
	}
	if req.Notes != nil {
		registration.Notes = *req.Notes
	}

	registration.UpdatedAt = time.Now()

	if err := s.domainRegistrationRepo.Update(ctx, registration); err != nil {
		return nil, fmt.Errorf("failed to update domain registration: %w", err)
	}

	return s.mapDomainRegistrationResponse(registration), nil
}

func (s *DomainManagementService) ListDomainRegistrations(ctx context.Context, filter domain.DomainRegistrationFilter) ([]*DomainRegistrationResponse, int64, error) {
	// TODO: Implement proper listing with pagination
	// For now, return empty list
	return []*DomainRegistrationResponse{}, 0, nil
}

// DNS Record Operations

func (s *DomainManagementService) CreateDNSRecord(ctx context.Context, req *CreateDNSRecordRequest) (*DNSRecordResponse, error) {
	// Create DNS record
	record := &domain.DNSRecord{
		ID:         uuid.New(),
		DomainID:   req.DomainID,
		RecordType: req.Type,
		Name:       req.Name,
		Value:      req.Value,
		TTL:        300, // Default TTL
		IsManaged:  true,
		Purpose:    req.Purpose,
		Status:     "pending",
		Notes:      req.Notes,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if req.TTL != nil {
		record.TTL = *req.TTL
	}
	if req.Priority != nil {
		record.Priority = req.Priority
	}
	if req.Weight != nil {
		record.Weight = req.Weight
	}
	if req.Port != nil {
		record.Port = req.Port
	}

	if err := s.dnsRecordRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	// TODO: Create record in actual DNS provider

	return s.mapDNSRecordResponse(record), nil
}

func (s *DomainManagementService) GetDNSRecord(ctx context.Context, id uuid.UUID) (*DNSRecordResponse, error) {
	record, err := s.dnsRecordRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}

	return s.mapDNSRecordResponse(record), nil
}

func (s *DomainManagementService) UpdateDNSRecord(ctx context.Context, id uuid.UUID, req *UpdateDNSRecordRequest) (*DNSRecordResponse, error) {
	record, err := s.dnsRecordRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}

	// Update fields
	if req.Value != nil {
		record.Value = *req.Value
	}
	if req.TTL != nil {
		record.TTL = *req.TTL
	}
	if req.Priority != nil {
		record.Priority = req.Priority
	}
	if req.Weight != nil {
		record.Weight = req.Weight
	}
	if req.Port != nil {
		record.Port = req.Port
	}
	if req.Purpose != nil {
		record.Purpose = *req.Purpose
	}
	if req.Notes != nil {
		record.Notes = *req.Notes
	}

	record.UpdatedAt = time.Now()

	if err := s.dnsRecordRepo.Update(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to update DNS record: %w", err)
	}

	// TODO: Update record in actual DNS provider

	return s.mapDNSRecordResponse(record), nil
}

func (s *DomainManagementService) DeleteDNSRecord(ctx context.Context, id uuid.UUID) error {
	// TODO: Delete record from actual DNS provider

	if err := s.dnsRecordRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	return nil
}

func (s *DomainManagementService) GetDNSRecordsByDomain(ctx context.Context, domainID uuid.UUID) ([]*DNSRecordResponse, error) {
	records, err := s.dnsRecordRepo.ListByDomainID(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS records: %w", err)
	}

	responses := make([]*DNSRecordResponse, len(records))
	for i, record := range records {
		responses[i] = s.mapDNSRecordResponse(record)
	}

	return responses, nil
}

// Helper functions

func (s *DomainManagementService) mapDomainRegistrationResponse(registration *domain.DomainRegistration) *DomainRegistrationResponse {
	return &DomainRegistrationResponse{
		ID:                 registration.ID,
		Domain:             "", // Would need to join with TenantDomain to get actual domain
		RegistrarProvider:  registration.RegistrarProvider,
		RegistrationStatus: registration.RegistrationStatus,
		RegistrationDate:   registration.RegistrationDate,
		ExpirationDate:     registration.ExpirationDate,
		AutoRenew:          registration.AutoRenew,
		RegistrationPrice:  registration.RegistrationPrice,
		RenewalPrice:       registration.RenewalPrice,
		Currency:           registration.Currency,
		NameServers:        registration.NameServers,
		ContactInfo:        registration.ContactInfo,
		PrivacyProtection:  registration.PrivacyProtection,
		TransferLock:       registration.TransferLock,
		Notes:              registration.Notes,
		CreatedAt:          registration.CreatedAt,
		UpdatedAt:          registration.UpdatedAt,
	}
}

func (s *DomainManagementService) mapDNSRecordResponse(record *domain.DNSRecord) *DNSRecordResponse {
	return &DNSRecordResponse{
		ID:             record.ID,
		DomainID:       record.DomainID,
		Type:           record.RecordType,
		Name:           record.Name,
		Value:          record.Value,
		TTL:            record.TTL,
		Priority:       record.Priority,
		Weight:         record.Weight,
		Port:           record.Port,
		IsManaged:      record.IsManaged,
		Purpose:        record.Purpose,
		Status:         record.Status,
		DNSProviderID:  record.DNSProviderID,
		LastCheckedAt:  record.LastCheckedAt,
		ValidationData: record.ValidationData,
		Notes:          record.Notes,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}
}

// isValidDomain performs basic domain validation
func isValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}

	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		// Basic character validation
		for _, char := range part {
			if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') || char == '-') {
				return false
			}
		}
	}

	return true
}
