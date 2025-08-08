package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// DomainService handles domain management operations
type DomainService struct {
	domainRepo        domain.TenantDomainRepository
	validationLogRepo domain.DomainValidationLogRepository
	sslCertRepo       domain.SSLCertificateRepository
	routingCacheRepo  domain.DomainRoutingCacheRepository
	auditLogRepo      domain.AuditLogRepository
	logger            *zap.Logger
}

// NewDomainService creates a new domain service
func NewDomainService(
	domainRepo domain.TenantDomainRepository,
	validationLogRepo domain.DomainValidationLogRepository,
	sslCertRepo domain.SSLCertificateRepository,
	routingCacheRepo domain.DomainRoutingCacheRepository,
	auditLogRepo domain.AuditLogRepository,
	logger *zap.Logger,
) *DomainService {
	return &DomainService{
		domainRepo:        domainRepo,
		validationLogRepo: validationLogRepo,
		sslCertRepo:       sslCertRepo,
		routingCacheRepo:  routingCacheRepo,
		auditLogRepo:      auditLogRepo,
		logger:            logger,
	}
}

// AddCustomDomainRequest represents the request to add a custom domain
type AddCustomDomainRequest struct {
	TenantID           string `json:"tenant_id" validate:"required"`
	Domain             string `json:"domain" validate:"required,fqdn"`
	VerificationMethod string `json:"verification_method,omitempty"` // dns, http, email
	AutoSSL            bool   `json:"auto_ssl,omitempty"`
	Priority           int    `json:"priority,omitempty"`
}

// DomainVerificationResponse contains domain verification information
type DomainVerificationResponse struct {
	Domain             string                  `json:"domain"`
	VerificationToken  string                  `json:"verification_token"`
	VerificationMethod string                  `json:"verification_method"`
	DNSRecord          *DNSVerificationRecord  `json:"dns_record,omitempty"`
	HTTPRecord         *HTTPVerificationRecord `json:"http_record,omitempty"`
	Instructions       string                  `json:"instructions"`
	ExpiresAt          time.Time               `json:"expires_at"`
}

// DNSVerificationRecord contains DNS verification details
type DNSVerificationRecord struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int    `json:"ttl"`
}

// HTTPVerificationRecord contains HTTP verification details
type HTTPVerificationRecord struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	URL     string `json:"url"`
}

// DomainStatus represents the status of a domain
type DomainStatus struct {
	Domain       string                 `json:"domain"`
	Verified     bool                   `json:"verified"`
	SSLEnabled   bool                   `json:"ssl_enabled"`
	Status       string                 `json:"status"`
	LastChecked  time.Time              `json:"last_checked"`
	SSLExpiry    *time.Time             `json:"ssl_expiry,omitempty"`
	HealthStatus string                 `json:"health_status"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
}

// ValidateAndNormalizeDomain validates and normalizes a domain name
func (s *DomainService) ValidateAndNormalizeDomain(domain string) (string, error) {
	// Remove protocol and path
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.Split(domain, "/")[0]
	domain = strings.Split(domain, ":")[0]

	// Convert to lowercase
	domain = strings.ToLower(domain)

	// Validate domain format
	domainRegex := regexp.MustCompile(`^[a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9\-]{0,61}[a-z0-9])?)*$`)
	if !domainRegex.MatchString(domain) {
		return "", fmt.Errorf("invalid domain format: %s", domain)
	}

	// Check if domain is too long
	if len(domain) > 253 {
		return "", fmt.Errorf("domain too long: %s", domain)
	}

	// Check for reserved domains
	reservedDomains := []string{
		"localhost",
		"zplus.io",
		"admin.zplus.io",
		"api.zplus.io",
		"www.zplus.io",
	}

	for _, reserved := range reservedDomains {
		if domain == reserved || strings.HasSuffix(domain, "."+reserved) {
			return "", fmt.Errorf("domain is reserved: %s", domain)
		}
	}

	return domain, nil
}

// AddCustomDomain adds a new custom domain for a tenant
func (s *DomainService) AddCustomDomain(ctx context.Context, req *AddCustomDomainRequest) (*DomainVerificationResponse, error) {
	// Validate and normalize domain
	normalizedDomain, err := s.ValidateAndNormalizeDomain(req.Domain)
	if err != nil {
		return nil, fmt.Errorf("domain validation failed: %w", err)
	}

	// Check if domain already exists
	existingDomain, err := s.domainRepo.GetByDomain(ctx, normalizedDomain)
	if err == nil && existingDomain != nil {
		return nil, fmt.Errorf("domain already exists: %s", normalizedDomain)
	}

	// Generate verification token
	token, err := s.generateVerificationToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate verification token: %w", err)
	}

	// Set default verification method
	verificationMethod := req.VerificationMethod
	if verificationMethod == "" {
		verificationMethod = "dns"
	}

	// Set default priority
	priority := req.Priority
	if priority == 0 {
		priority = 150 // Default for custom domains
	}

	// Create domain entry
	domainEntry := &domain.TenantDomain{
		TenantID:           req.TenantID,
		Domain:             normalizedDomain,
		IsCustom:           true,
		Verified:           false,
		SSLEnabled:         false,
		VerificationToken:  token,
		VerificationMethod: verificationMethod,
		RoutingPriority:    priority,
		Status:             "inactive",
		MetricsEnabled:     true,
		SSLAutoRenew:       req.AutoSSL,
		RateLimitConfig: map[string]interface{}{
			"requests_per_minute": 100,
			"burst_limit":         200,
			"tenant_type":         "standard",
		},
		SecurityConfig: map[string]interface{}{
			"force_ssl":    true,
			"hsts_enabled": true,
			"csp_enabled":  false,
			"cors_enabled": true,
		},
		HealthCheckConfig: map[string]interface{}{
			"enabled":  true,
			"path":     "/health",
			"interval": 30,
			"timeout":  10,
			"retries":  3,
		},
	}

	// Save domain to database
	if err := s.domainRepo.Create(ctx, domainEntry); err != nil {
		return nil, fmt.Errorf("failed to create domain: %w", err)
	}

	// Create validation log entry
	validationLog := &domain.DomainValidationLog{
		DomainID:       domainEntry.ID,
		ValidationType: verificationMethod,
		Status:         "pending",
		ValidationData: map[string]interface{}{
			"verification_token": token,
			"domain":             normalizedDomain,
			"tenant_id":          req.TenantID,
		},
		Attempts:    1,
		NextRetryAt: func() *time.Time { t := time.Now().Add(5 * time.Minute); return &t }(),
	}

	if err := s.validationLogRepo.Create(ctx, validationLog); err != nil {
		s.logger.Error("Failed to create validation log", zap.Error(err))
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		TenantID: req.TenantID,
		Action:   domain.ActionCreate,
		Resource: "custom_domain",
		Details: &domain.Details{
			After: map[string]interface{}{
				"domain":              normalizedDomain,
				"verification_method": verificationMethod,
				"auto_ssl":            req.AutoSSL,
			},
		},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	// Build verification response
	response := &DomainVerificationResponse{
		Domain:             normalizedDomain,
		VerificationToken:  token,
		VerificationMethod: verificationMethod,
		ExpiresAt:          time.Now().Add(24 * time.Hour),
	}

	switch verificationMethod {
	case "dns":
		response.DNSRecord = &DNSVerificationRecord{
			Type:  "TXT",
			Name:  "_zplus-verify." + normalizedDomain,
			Value: token,
			TTL:   300,
		}
		response.Instructions = fmt.Sprintf(
			"Add a TXT record to your DNS:\nName: %s\nValue: %s\nTTL: 300",
			response.DNSRecord.Name,
			response.DNSRecord.Value,
		)
	case "http":
		response.HTTPRecord = &HTTPVerificationRecord{
			Path:    "/.well-known/zplus-verification",
			Content: token,
			URL:     fmt.Sprintf("http://%s/.well-known/zplus-verification", normalizedDomain),
		}
		response.Instructions = fmt.Sprintf(
			"Create a file at %s with content: %s",
			response.HTTPRecord.URL,
			response.HTTPRecord.Content,
		)
	}

	s.logger.Info("Custom domain added",
		zap.String("tenant_id", req.TenantID),
		zap.String("domain", normalizedDomain),
		zap.String("verification_method", verificationMethod),
	)

	return response, nil
}

// VerifyDomain verifies domain ownership
func (s *DomainService) VerifyDomain(ctx context.Context, domainID uuid.UUID) (*DomainStatus, error) {
	// Get domain entry
	domainEntry, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return nil, fmt.Errorf("domain not found: %w", err)
	}

	// Perform verification based on method
	var verified bool
	var verificationError error

	switch domainEntry.VerificationMethod {
	case "dns":
		verified, verificationError = s.verifyDNSRecord(domainEntry.Domain, domainEntry.VerificationToken)
	case "http":
		verified, verificationError = s.verifyHTTPRecord(domainEntry.Domain, domainEntry.VerificationToken)
	default:
		verificationError = fmt.Errorf("unsupported verification method: %s", domainEntry.VerificationMethod)
	}

	// Update domain status
	now := time.Now()
	if verified {
		domainEntry.Verified = true
		domainEntry.VerifiedAt = &now
		domainEntry.Status = "active"

		// Enable SSL if auto SSL is enabled
		if domainEntry.SSLAutoRenew {
			go s.requestSSLCertificate(ctx, domainEntry)
		}

		// Update routing cache
		s.updateRoutingCache(ctx, domainEntry)

		s.logger.Info("Domain verified successfully",
			zap.String("domain", domainEntry.Domain),
			zap.String("tenant_id", domainEntry.TenantID),
		)
	} else {
		s.logger.Warn("Domain verification failed",
			zap.String("domain", domainEntry.Domain),
			zap.String("tenant_id", domainEntry.TenantID),
			zap.Error(verificationError),
		)
	}

	// Update domain in database
	if err := s.domainRepo.Update(ctx, domainEntry); err != nil {
		return nil, fmt.Errorf("failed to update domain: %w", err)
	}

	// Log verification attempt
	validationLog := &domain.DomainValidationLog{
		DomainID:       domainEntry.ID,
		ValidationType: domainEntry.VerificationMethod,
		Status: func() string {
			if verified {
				return "success"
			}
			return "failed"
		}(),
		ValidationData: map[string]interface{}{
			"domain":             domainEntry.Domain,
			"verification_token": domainEntry.VerificationToken,
		},
		CompletedAt: &now,
	}

	if verificationError != nil {
		validationLog.ErrorMessage = verificationError.Error()
	}

	_ = s.validationLogRepo.Create(ctx, validationLog)

	// Create audit log
	auditLog := &domain.AuditLog{
		TenantID: domainEntry.TenantID,
		Action:   domain.ActionUpdate,
		Resource: "custom_domain",
		Details: &domain.Details{
			Before: map[string]interface{}{
				"verified": !verified,
			},
			After: map[string]interface{}{
				"verified": verified,
				"status":   domainEntry.Status,
			},
		},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	return &DomainStatus{
		Domain:       domainEntry.Domain,
		Verified:     domainEntry.Verified,
		SSLEnabled:   domainEntry.SSLEnabled,
		Status:       domainEntry.Status,
		LastChecked:  now,
		HealthStatus: "unknown",
	}, nil
}

// GetDomainsByTenant returns all domains for a tenant
func (s *DomainService) GetDomainsByTenant(ctx context.Context, tenantID string) ([]*domain.TenantDomain, error) {
	return s.domainRepo.GetByTenantID(ctx, tenantID)
}

// DeleteCustomDomain removes a custom domain
func (s *DomainService) DeleteCustomDomain(ctx context.Context, domainID uuid.UUID, tenantID string) error {
	// Get domain entry
	domainEntry, err := s.domainRepo.GetByID(ctx, domainID)
	if err != nil {
		return fmt.Errorf("domain not found: %w", err)
	}

	// Verify ownership
	if domainEntry.TenantID != tenantID {
		return fmt.Errorf("domain does not belong to tenant")
	}

	// Prevent deletion of primary subdomain
	if !domainEntry.IsCustom {
		return fmt.Errorf("cannot delete primary subdomain")
	}

	// Remove from routing cache
	if err := s.routingCacheRepo.DeleteByDomain(ctx, domainEntry.Domain); err != nil {
		s.logger.Warn("Failed to remove domain from routing cache", zap.Error(err))
	}

	// Delete domain
	if err := s.domainRepo.Delete(ctx, domainID); err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}

	// Create audit log
	auditLog := &domain.AuditLog{
		TenantID: tenantID,
		Action:   domain.ActionDelete,
		Resource: "custom_domain",
		Details: &domain.Details{
			Before: map[string]interface{}{
				"domain":      domainEntry.Domain,
				"verified":    domainEntry.Verified,
				"ssl_enabled": domainEntry.SSLEnabled,
			},
		},
	}
	_ = s.auditLogRepo.Create(ctx, auditLog)

	s.logger.Info("Custom domain deleted",
		zap.String("domain", domainEntry.Domain),
		zap.String("tenant_id", tenantID),
	)

	return nil
}

// generateVerificationToken generates a secure verification token
func (s *DomainService) generateVerificationToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "zplus-verify-" + hex.EncodeToString(bytes), nil
}

// verifyDNSRecord verifies DNS TXT record for domain ownership
func (s *DomainService) verifyDNSRecord(domain, expectedToken string) (bool, error) {
	recordName := "_zplus-verify." + domain

	txtRecords, err := net.LookupTXT(recordName)
	if err != nil {
		return false, fmt.Errorf("DNS lookup failed: %w", err)
	}

	for _, record := range txtRecords {
		if strings.TrimSpace(record) == expectedToken {
			return true, nil
		}
	}

	return false, fmt.Errorf("verification token not found in DNS records")
}

// verifyHTTPRecord verifies HTTP file for domain ownership
func (s *DomainService) verifyHTTPRecord(domain, expectedToken string) (bool, error) {
	// This would implement HTTP verification
	// For now, return false as it requires HTTP client implementation
	return false, fmt.Errorf("HTTP verification not implemented yet")
}

// requestSSLCertificate requests SSL certificate for verified domain
func (s *DomainService) requestSSLCertificate(ctx context.Context, domainEntry *domain.TenantDomain) {
	// This would integrate with Let's Encrypt or other CA
	// For now, just log the request
	s.logger.Info("SSL certificate requested",
		zap.String("domain", domainEntry.Domain),
		zap.String("tenant_id", domainEntry.TenantID),
	)
}

// updateRoutingCache updates the routing cache for a domain
func (s *DomainService) updateRoutingCache(ctx context.Context, domainEntry *domain.TenantDomain) {
	if !domainEntry.Verified || domainEntry.Status != "active" {
		return
	}

	routingEntry := &domain.DomainRoutingCache{
		Domain:         domainEntry.Domain,
		TenantID:       domainEntry.TenantID,
		BackendService: "ilms-api",
		RoutingConfig: map[string]interface{}{
			"rate_limit":   domainEntry.RateLimitConfig,
			"security":     domainEntry.SecurityConfig,
			"health_check": domainEntry.HealthCheckConfig,
			"priority":     domainEntry.RoutingPriority,
			"ssl_enabled":  domainEntry.SSLEnabled,
			"is_custom":    domainEntry.IsCustom,
		},
		CacheExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if err := s.routingCacheRepo.Upsert(ctx, routingEntry); err != nil {
		s.logger.Error("Failed to update routing cache", zap.Error(err))
	}
}
