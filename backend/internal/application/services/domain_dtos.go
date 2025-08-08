package services

import (
	"time"

	"github.com/google/uuid"
)

// Domain Registration DTOs

type RegisterDomainRequest struct {
	Domain             string                 `json:"domain" validate:"required,fqdn"`
	TenantID           uuid.UUID              `json:	DomainID uuid.UUID `json:"domain_id" validate:"required"`tenant_id" validate:"required"`
	RegistrarProvider  string                 `json:"registrar_provider" validate:"required"`
	RegistrationPeriod int                    `json:"registration_period" validate:"min=1,max=10"` // years
	AutoRenew          bool                   `json:"auto_renew"`
	PrivacyProtection  bool                   `json:"privacy_protection"`
	ContactInfo        map[string]interface{} `json:"contact_info" validate:"required"`
	NameServers        []string               `json:"name_servers,omitempty"`
	TransferLock       bool                   `json:"transfer_lock"`
}

type DomainRegistrationResponse struct {
	ID                 uuid.UUID              `json:"id"`
	Domain             string                 `json:"domain"`
	RegistrarProvider  string                 `json:"registrar_provider"`
	RegistrationStatus string                 `json:"registration_status"`
	RegistrationDate   *time.Time             `json:"registration_date,omitempty"`
	ExpirationDate     *time.Time             `json:"expiration_date,omitempty"`
	AutoRenew          bool                   `json:"auto_renew"`
	RegistrationPrice  float64                `json:"registration_price"`
	RenewalPrice       float64                `json:"renewal_price"`
	Currency           string                 `json:"currency"`
	NameServers        []string               `json:"name_servers"`
	ContactInfo        map[string]interface{} `json:"contact_info"`
	PrivacyProtection  bool                   `json:"privacy_protection"`
	TransferLock       bool                   `json:"transfer_lock"`
	Notes              string                 `json:"notes,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type UpdateDomainRegistrationRequest struct {
	AutoRenew         *bool                  `json:"auto_renew,omitempty"`
	PrivacyProtection *bool                  `json:"privacy_protection,omitempty"`
	TransferLock      *bool                  `json:"transfer_lock,omitempty"`
	NameServers       []string               `json:"name_servers,omitempty"`
	ContactInfo       map[string]interface{} `json:"contact_info,omitempty"`
	Notes             *string                `json:"notes,omitempty"`
}

type DomainAvailabilityRequest struct {
	Domain string `json:"domain" validate:"required,fqdn"`
}

type DomainAvailabilityResponse struct {
	Domain      string   `json:"domain"`
	Available   bool     `json:"available"`
	Price       float64  `json:"price,omitempty"`
	Currency    string   `json:"currency,omitempty"`
	Premium     bool     `json:"premium"`
	Reason      string   `json:"reason,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// DNS Management DTOs

type CreateDNSRecordRequest struct {
	DomainID uuid.UUID    `json:"domain_id" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=A AAAA CNAME MX TXT NS PTR SRV CAA"`
	Name     string `json:"name" validate:"required"`
	Value    string `json:"value" validate:"required"`
	TTL      *int   `json:"ttl,omitempty"`
	Priority *int   `json:"priority,omitempty"`
	Weight   *int   `json:"weight,omitempty"`
	Port     *int   `json:"port,omitempty"`
	Purpose  string `json:"purpose,omitempty"`
	Notes    string `json:"notes,omitempty"`
}

type UpdateDNSRecordRequest struct {
	Value    *string `json:"value,omitempty"`
	TTL      *int    `json:"ttl,omitempty"`
	Priority *int    `json:"priority,omitempty"`
	Weight   *int    `json:"weight,omitempty"`
	Port     *int    `json:"port,omitempty"`
	Purpose  *string `json:"purpose,omitempty"`
	Notes    *string `json:"notes,omitempty"`
}

type DNSRecordResponse struct {
	ID             uuid.UUID              `json:"id"`
	DomainID       int                    `json:"domain_id"`
	Type           string                 `json:"type"`
	Name           string                 `json:"name"`
	Value          string                 `json:"value"`
	TTL            int                    `json:"ttl"`
	Priority       *int                   `json:"priority,omitempty"`
	Weight         *int                   `json:"weight,omitempty"`
	Port           *int                   `json:"port,omitempty"`
	IsManaged      bool                   `json:"is_managed"`
	Purpose        string                 `json:"purpose,omitempty"`
	Status         string                 `json:"status"`
	DNSProviderID  string                 `json:"dns_provider_id,omitempty"`
	LastCheckedAt  *time.Time             `json:"last_checked_at,omitempty"`
	ValidationData map[string]interface{} `json:"validation_data,omitempty"`
	Notes          string                 `json:"notes,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type BulkDNSOperationRequest struct {
	DomainID uuid.UUID                      `json:"domain_id" validate:"required"`
	Records  []CreateDNSRecordRequest `json:"records" validate:"required,min=1"`
}

type DNSZoneResponse struct {
	DomainID uuid.UUID                 `json:"domain_id"`
	Domain   string              `json:"domain"`
	Records  []DNSRecordResponse `json:"records"`
	SOA      *DNSRecordResponse  `json:"soa,omitempty"`
	NS       []DNSRecordResponse `json:"ns,omitempty"`
}

// Domain Ownership Verification DTOs

type InitiateDomainVerificationRequest struct {
	DomainID           int    `json:"domain_id" validate:"required"`
	VerificationMethod string `json:"verification_method" validate:"required,oneof=dns_txt dns_cname file_upload email meta_tag"`
}

type DomainVerificationResponse struct {
	ID                 uuid.UUID              `json:"id"`
	DomainID           int                    `json:"domain_id"`
	VerificationMethod string                 `json:"verification_method"`
	Status             string                 `json:"status"`
	VerificationToken  string                 `json:"verification_token"`
	ExpectedValue      string                 `json:"expected_value,omitempty"`
	Instructions       map[string]interface{} `json:"instructions"`
	AttemptCount       int                    `json:"attempt_count"`
	MaxAttempts        int                    `json:"max_attempts"`
	NextAttemptAt      *time.Time             `json:"next_attempt_at,omitempty"`
	LastAttemptAt      *time.Time             `json:"last_attempt_at,omitempty"`
	VerifiedAt         *time.Time             `json:"verified_at,omitempty"`
	ExpiresAt          *time.Time             `json:"expires_at,omitempty"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type VerifyDomainOwnershipRequest struct {
	VerificationID uuid.UUID `json:"verification_id" validate:"required"`
}

// SSL Certificate Management DTOs

type RequestSSLCertificateRequest struct {
	DomainID        int      `json:"domain_id" validate:"required"`
	CertificateType string   `json:"certificate_type" validate:"oneof=domain_validated organization_validated extended_validation"`
	ChallengeType   string   `json:"challenge_type" validate:"required,oneof=http-01 dns-01 tls-alpn-01"`
	KeyType         string   `json:"key_type" validate:"oneof=rsa ecdsa"`
	KeySize         *int     `json:"key_size,omitempty"`
	Domains         []string `json:"domains,omitempty"` // Additional domains for SAN certificate
}

type SSLCertificateRequestResponse struct {
	ID                uuid.UUID              `json:"id"`
	DomainID          int                    `json:"domain_id"`
	RequestType       string                 `json:"request_type"`
	ChallengeType     string                 `json:"challenge_type"`
	Status            string                 `json:"status"`
	CertificateType   string                 `json:"certificate_type"`
	KeyType           string                 `json:"key_type"`
	KeySize           int                    `json:"key_size"`
	Domains           []string               `json:"domains"`
	ChallengeData     map[string]interface{} `json:"challenge_data,omitempty"`
	ValidationRecords map[string]interface{} `json:"validation_records,omitempty"`
	AttemptCount      int                    `json:"attempt_count"`
	MaxAttempts       int                    `json:"max_attempts"`
	LastAttemptAt     *time.Time             `json:"last_attempt_at,omitempty"`
	CompletedAt       *time.Time             `json:"completed_at,omitempty"`
	ExpiresAt         *time.Time             `json:"expires_at,omitempty"`
	ErrorMessage      string                 `json:"error_message,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type SSLCertificateResponse struct {
	ID                 uint       `json:"id"`
	DomainID           int        `json:"domain_id"`
	Issuer             string     `json:"issuer"`
	SerialNumber       string     `json:"serial_number"`
	Fingerprint        string     `json:"fingerprint"`
	Subject            string     `json:"subject"`
	SAN                []string   `json:"san"`
	IssuedAt           time.Time  `json:"issued_at"`
	ExpiresAt          time.Time  `json:"expires_at"`
	AutoRenew          bool       `json:"auto_renew"`
	RenewalAttempts    int        `json:"renewal_attempts"`
	LastRenewalAttempt *time.Time `json:"last_renewal_attempt,omitempty"`
	Status             string     `json:"status"`
	DaysUntilExpiry    int        `json:"days_until_expiry"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

type RenewSSLCertificateRequest struct {
	CertificateID uint   `json:"certificate_id" validate:"required"`
	ForceRenewal  bool   `json:"force_renewal"`
	ChallengeType string `json:"challenge_type,omitempty"`
}

// Domain Health Check DTOs

type CreateDomainHealthCheckRequest struct {
	DomainID         int    `json:"domain_id" validate:"required"`
	CheckType        string `json:"check_type" validate:"required,oneof=http https dns ssl ping"`
	CheckFrequency   *int   `json:"check_frequency,omitempty"`   // seconds
	TimeoutThreshold *int   `json:"timeout_threshold,omitempty"` // seconds
	AlertsEnabled    *bool  `json:"alerts_enabled,omitempty"`
}

type UpdateDomainHealthCheckRequest struct {
	CheckFrequency   *int  `json:"check_frequency,omitempty"`
	TimeoutThreshold *int  `json:"timeout_threshold,omitempty"`
	AlertsEnabled    *bool `json:"alerts_enabled,omitempty"`
}

type DomainHealthCheckResponse struct {
	ID               uuid.UUID              `json:"id"`
	DomainID         int                    `json:"domain_id"`
	CheckType        string                 `json:"check_type"`
	Status           string                 `json:"status"`
	ResponseTime     int                    `json:"response_time"`
	StatusCode       *int                   `json:"status_code,omitempty"`
	ErrorMessage     string                 `json:"error_message,omitempty"`
	CheckFrequency   int                    `json:"check_frequency"`
	TimeoutThreshold int                    `json:"timeout_threshold"`
	HealthData       map[string]interface{} `json:"health_data,omitempty"`
	AlertsEnabled    bool                   `json:"alerts_enabled"`
	LastAlertAt      *time.Time             `json:"last_alert_at,omitempty"`
	NextCheckAt      *time.Time             `json:"next_check_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// Domain Management Analytics DTOs

type DomainStatsResponse struct {
	TotalDomains      int                    `json:"total_domains"`
	CustomDomains     int                    `json:"custom_domains"`
	VerifiedDomains   int                    `json:"verified_domains"`
	ActiveSSLCerts    int                    `json:"active_ssl_certs"`
	ExpiringSSLCerts  int                    `json:"expiring_ssl_certs"`
	ExpiringDomains   int                    `json:"expiring_domains"`
	HealthyDomains    int                    `json:"healthy_domains"`
	UnhealthyDomains  int                    `json:"unhealthy_domains"`
	DomainsByProvider map[string]int         `json:"domains_by_provider"`
	SSLCertsByIssuer  map[string]int         `json:"ssl_certs_by_issuer"`
	HealthCheckStats  map[string]interface{} `json:"health_check_stats"`
}

type DomainMetricsRequest struct {
	DomainID    int       `json:"domain_id" validate:"required"`
	MetricTypes []string  `json:"metric_types" validate:"required"`
	From        time.Time `json:"from" validate:"required"`
	To          time.Time `json:"to" validate:"required"`
}

type DomainMetricsResponse struct {
	DomainID    int                    `json:"domain_id"`
	Metrics     map[string]interface{} `json:"metrics"`
	LastUpdated time.Time              `json:"last_updated"`
}

// List and Filter DTOs

type ListDomainsRequest struct {
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	IsCustom  *bool      `json:"is_custom,omitempty"`
	Verified  *bool      `json:"verified,omitempty"`
	Status    *string    `json:"status,omitempty"`
	Search    *string    `json:"search,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order" validate:"oneof=asc desc"`
}

type DomainListResponse struct {
	Domains    []TenantDomainResponse `json:"domains"`
	Total      int64                  `json:"total"`
	Page       int                    `json:"page"`
	Limit      int                    `json:"limit"`
	TotalPages int                    `json:"total_pages"`
}

type TenantDomainResponse struct {
	ID                 uuid.UUID              `json:"id"`
	TenantID           uuid.UUID              `json:"tenant_id"`
	Domain             string                 `json:"domain"`
	IsCustom           bool                   `json:"is_custom"`
	Verified           bool                   `json:"verified"`
	IsPrimary          bool                   `json:"is_primary"`
	SSLEnabled         bool                   `json:"ssl_enabled"`
	VerificationMethod string                 `json:"verification_method,omitempty"`
	VerifiedAt         *time.Time             `json:"verified_at,omitempty"`
	SSLCertIssuedAt    *time.Time             `json:"ssl_cert_issued_at,omitempty"`
	SSLCertExpiresAt   *time.Time             `json:"ssl_cert_expires_at,omitempty"`
	DNSProvider        string                 `json:"dns_provider,omitempty"`
	ValidationErrors   map[string]interface{} `json:"validation_errors,omitempty"`
	Status             string                 `json:"status"`
	LastHealthCheck    *time.Time             `json:"last_health_check,omitempty"`
	LastSSLCheck       *time.Time             `json:"last_ssl_check,omitempty"`
	Notes              string                 `json:"notes,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	// Related data
	Registration   *DomainRegistrationResponse `json:"registration,omitempty"`
	SSLCertificate *SSLCertificateResponse     `json:"ssl_certificate,omitempty"`
	HealthStatus   string                      `json:"health_status,omitempty"`
}
