package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID                 uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name               string                 `json:"name" gorm:"not null"`
	Subdomain          string                 `json:"subdomain" gorm:"unique;not null"`
	Status             string                 `json:"status" gorm:"not null;default:'pending'"`
	Plan               string                 `json:"plan" gorm:"not null;default:'starter'"`
	Settings           *TenantSettings        `json:"settings" gorm:"type:jsonb"`
	Configuration      *TenantConfiguration   `json:"configuration" gorm:"type:jsonb"`
	Branding           *TenantBranding        `json:"branding" gorm:"type:jsonb"`
	Billing            *TenantBilling         `json:"billing" gorm:"type:jsonb"`
	Features           []string               `json:"features" gorm:"type:text[]"`
	Integrations       map[string]interface{} `json:"integrations" gorm:"type:jsonb"`
	Metadata           map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	OnboardingStatus   string                 `json:"onboarding_status" gorm:"default:'pending'"`
	OnboardingStep     int                    `json:"onboarding_step" gorm:"default:0"`
	OnboardingData     map[string]interface{} `json:"onboarding_data" gorm:"type:jsonb"`
	ContactEmail       string                 `json:"contact_email"`
	ContactName        string                 `json:"contact_name"`
	ContactPhone       string                 `json:"contact_phone"`
	CompanySize        string                 `json:"company_size"`
	Industry           string                 `json:"industry"`
	Region             string                 `json:"region" gorm:"default:'us-west-1'"`
	Timezone           string                 `json:"timezone" gorm:"default:'UTC'"`
	Language           string                 `json:"language" gorm:"default:'en'"`
	Currency           string                 `json:"currency" gorm:"default:'USD'"`
	TrialStartsAt      *time.Time             `json:"trial_starts_at"`
	TrialEndsAt        *time.Time             `json:"trial_ends_at"`
	SubscriptionID     string                 `json:"subscription_id"`
	SubscriptionStatus string                 `json:"subscription_status" gorm:"default:'trial'"`
	LastActivityAt     *time.Time             `json:"last_activity_at"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
	DeletedAt          *time.Time             `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Domains        []TenantDomain        `json:"domains" gorm:"foreignKey:TenantID"`
	Users          []User                `json:"users" gorm:"many2many:tenant_users;"`
	OnboardingLogs []TenantOnboardingLog `json:"onboarding_logs" gorm:"foreignKey:TenantID"`
}

// TenantDomain represents custom domains for tenants
type TenantDomain struct {
	ID                 uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID           string                 `json:"tenant_id" gorm:"type:varchar(50);not null"`
	Domain             string                 `json:"domain" gorm:"unique;not null"`
	IsCustom           bool                   `json:"is_custom" gorm:"default:false"`
	Verified           bool                   `json:"verified" gorm:"default:false"`
	IsVerified         bool                   `json:"is_verified" gorm:"default:false"` // Alias for compatibility
	IsPrimary          bool                   `json:"is_primary" gorm:"default:false"`
	SSLEnabled         bool                   `json:"ssl_enabled" gorm:"default:false"`
	VerificationToken  string                 `json:"verification_token"`
	VerificationMethod string                 `json:"verification_method" gorm:"default:'dns'"`
	VerifiedAt         *time.Time             `json:"verified_at"`
	SSLCertIssuedAt    *time.Time             `json:"ssl_cert_issued_at"`
	SSLCertExpiresAt   *time.Time             `json:"ssl_cert_expires_at"`
	DNSProvider        string                 `json:"dns_provider" gorm:"default:'auto'"`
	DNSZoneID          string                 `json:"dns_zone_id"`
	ValidationErrors   map[string]interface{} `json:"validation_errors" gorm:"type:jsonb;default:'[]'"`
	SSLIssuer          string                 `json:"ssl_issuer" gorm:"default:'letsencrypt'"`
	SSLCertSubject     string                 `json:"ssl_cert_subject"`
	SSLCertSAN         []string               `json:"ssl_cert_san" gorm:"type:text[]"`
	SSLAutoRenew       bool                   `json:"ssl_auto_renew" gorm:"default:true"`
	RoutingPriority    int                    `json:"routing_priority" gorm:"default:100"`
	RateLimitConfig    map[string]interface{} `json:"rate_limit_config" gorm:"type:jsonb"`
	SecurityConfig     map[string]interface{} `json:"security_config" gorm:"type:jsonb"`
	HealthCheckConfig  map[string]interface{} `json:"health_check_config" gorm:"type:jsonb"`
	MetricsEnabled     bool                   `json:"metrics_enabled" gorm:"default:true"`
	LastHealthCheck    *time.Time             `json:"last_health_check"`
	LastSSLCheck       *time.Time             `json:"last_ssl_check"`
	Status             string                 `json:"status" gorm:"default:'active'"`
	Notes              string                 `json:"notes"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`

	// Relationships - Note: Using string TenantID instead of UUID for flexibility
	ValidationLogs  []DomainValidationLog `json:"validation_logs" gorm:"foreignKey:DomainID"`
	SSLCertificates []SSLCertificate      `json:"ssl_certificates" gorm:"foreignKey:DomainID"`
}

// DomainValidationLog tracks domain validation attempts
type DomainValidationLog struct {
	ID             uint                   `json:"id" gorm:"primaryKey"`
	DomainID       uuid.UUID              `json:"domain_id" gorm:"type:uuid;not null"`
	ValidationType string                 `json:"validation_type" gorm:"not null"`
	Status         string                 `json:"status" gorm:"not null"`
	ValidationData map[string]interface{} `json:"validation_data" gorm:"type:jsonb;default:'{}'"`
	ErrorMessage   string                 `json:"error_message"`
	Attempts       int                    `json:"attempts" gorm:"default:1"`
	NextRetryAt    *time.Time             `json:"next_retry_at"`
	CreatedAt      time.Time              `json:"created_at"`
	CompletedAt    *time.Time             `json:"completed_at"`

	// Relationships
	Domain TenantDomain `json:"domain" gorm:"foreignKey:DomainID"`
}

// SSLCertificate tracks SSL certificates for domains
type SSLCertificate struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	DomainID           uuid.UUID  `json:"domain_id" gorm:"type:uuid;not null"`
	RequestID          *uuid.UUID `json:"request_id" gorm:"type:uuid"` // Reference to SSLCertificateRequest
	Issuer             string     `json:"issuer" gorm:"not null;default:'letsencrypt'"`
	CertificatePEM     string     `json:"certificate_pem" gorm:"type:text"`
	PrivateKeyPEM      string     `json:"private_key_pem" gorm:"type:text"` // Should be encrypted
	ChainPEM           string     `json:"chain_pem" gorm:"type:text"`
	SerialNumber       string     `json:"serial_number"`
	Fingerprint        string     `json:"fingerprint"`
	Subject            string     `json:"subject"`
	SAN                []string   `json:"san" gorm:"type:text[]"`
	IssuedAt           time.Time  `json:"issued_at" gorm:"not null"`
	ExpiresAt          time.Time  `json:"expires_at" gorm:"not null"`
	AutoRenew          bool       `json:"auto_renew" gorm:"default:true"`
	RenewalAttempts    int        `json:"renewal_attempts" gorm:"default:0"`
	LastRenewalAttempt *time.Time `json:"last_renewal_attempt"`
	Status             string     `json:"status" gorm:"default:'active'"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Relationships
	Domain  TenantDomain           `json:"domain" gorm:"foreignKey:DomainID"`
	Request *SSLCertificateRequest `json:"request" gorm:"foreignKey:RequestID"`
}

// DomainRegistration represents domain registration tracking
type DomainRegistration struct {
	ID                 uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DomainID           uuid.UUID              `json:"domain_id" gorm:"type:uuid;not null"`
	RegistrarProvider  string                 `json:"registrar_provider" gorm:"not null"` // namecheap, godaddy, cloudflare, etc.
	RegistrarDomainID  string                 `json:"registrar_domain_id"`
	RegistrationStatus string                 `json:"registration_status" gorm:"not null;default:'pending'"` // pending, registered, failed, expired
	RegistrationDate   *time.Time             `json:"registration_date"`
	ExpirationDate     *time.Time             `json:"expiration_date"`
	AutoRenew          bool                   `json:"auto_renew" gorm:"default:true"`
	RegistrationPrice  float64                `json:"registration_price"`
	RenewalPrice       float64                `json:"renewal_price"`
	Currency           string                 `json:"currency" gorm:"default:'USD'"`
	NameServers        []string               `json:"name_servers" gorm:"type:text[]"`
	ContactInfo        map[string]interface{} `json:"contact_info" gorm:"type:jsonb;default:'{}'"`
	PrivacyProtection  bool                   `json:"privacy_protection" gorm:"default:true"`
	TransferLock       bool                   `json:"transfer_lock" gorm:"default:true"`
	RegistrationConfig map[string]interface{} `json:"registration_config" gorm:"type:jsonb;default:'{}'"`
	LastCheckedAt      *time.Time             `json:"last_checked_at"`
	Notes              string                 `json:"notes"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`

	// Relationships
	Domain             TenantDomain              `json:"domain" gorm:"foreignKey:DomainID"`
	RegistrationEvents []DomainRegistrationEvent `json:"registration_events" gorm:"foreignKey:RegistrationID"`
}

// DomainRegistrationEvent tracks registration events and status changes
type DomainRegistrationEvent struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	RegistrationID uuid.UUID              `json:"registration_id" gorm:"type:uuid;not null"`
	EventType      string                 `json:"event_type" gorm:"not null"` // registration_started, registered, renewed, transfer_initiated, etc.
	Status         string                 `json:"status" gorm:"not null"`     // success, failed, pending
	Message        string                 `json:"message"`
	EventData      map[string]interface{} `json:"event_data" gorm:"type:jsonb;default:'{}'"`
	ErrorCode      string                 `json:"error_code"`
	CreatedAt      time.Time              `json:"created_at"`

	// Relationships
	Registration DomainRegistration `json:"registration" gorm:"foreignKey:RegistrationID"`
}

// DNSRecord represents DNS record management
type DNSRecord struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DomainID       int                    `json:"domain_id" gorm:"not null"`
	RecordType     string                 `json:"record_type" gorm:"not null"` // A, AAAA, CNAME, MX, TXT, etc.
	Name           string                 `json:"name" gorm:"not null"`        // subdomain or @ for root
	Value          string                 `json:"value" gorm:"not null"`
	TTL            int                    `json:"ttl" gorm:"default:3600"`
	Priority       *int                   `json:"priority"`                       // For MX records
	Weight         *int                   `json:"weight"`                         // For SRV records
	Port           *int                   `json:"port"`                           // For SRV records
	IsManaged      bool                   `json:"is_managed" gorm:"default:true"` // Managed by system or manual
	Purpose        string                 `json:"purpose"`                        // verification, email, www, api, etc.
	Status         string                 `json:"status" gorm:"default:'active'"`
	DNSProviderID  string                 `json:"dns_provider_id"` // External DNS provider record ID
	LastCheckedAt  *time.Time             `json:"last_checked_at"`
	ValidationData map[string]interface{} `json:"validation_data" gorm:"type:jsonb;default:'{}'"`
	Notes          string                 `json:"notes"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`

	// Relationships
	Domain TenantDomain `json:"domain" gorm:"foreignKey:DomainID"`
}

// DomainOwnershipVerification represents domain ownership verification methods
type DomainOwnershipVerification struct {
	ID                 uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DomainID           uuid.UUID              `json:"domain_id" gorm:"type:uuid;not null"`
	VerificationMethod string                 `json:"verification_method" gorm:"not null"`      // dns_txt, dns_cname, file_upload, email, meta_tag
	Status             string                 `json:"status" gorm:"not null;default:'pending'"` // pending, verified, failed, expired
	VerificationToken  string                 `json:"verification_token" gorm:"not null"`
	ExpectedValue      string                 `json:"expected_value"` // Expected TXT record value, filename, etc.
	ActualValue        string                 `json:"actual_value"`   // What was actually found during verification
	VerificationData   map[string]interface{} `json:"verification_data" gorm:"type:jsonb;default:'{}'"`
	AttemptCount       int                    `json:"attempt_count" gorm:"default:0"`
	MaxAttempts        int                    `json:"max_attempts" gorm:"default:10"`
	NextAttemptAt      *time.Time             `json:"next_attempt_at"`
	LastAttemptAt      *time.Time             `json:"last_attempt_at"`
	VerifiedAt         *time.Time             `json:"verified_at"`
	ExpiresAt          *time.Time             `json:"expires_at"`
	ErrorMessage       string                 `json:"error_message"`
	Instructions       map[string]interface{} `json:"instructions" gorm:"type:jsonb;default:'{}'"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`

	// Relationships
	Domain TenantDomain `json:"domain" gorm:"foreignKey:DomainID"`
}

// SSLCertificateRequest represents SSL certificate requests and ACME challenges
type SSLCertificateRequest struct {
	ID                uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DomainID          uuid.UUID              `json:"domain_id" gorm:"type:uuid;not null"`
	RequestType       string                 `json:"request_type" gorm:"not null;default:'new'"`         // new, renewal, revoke
	ChallengeType     string                 `json:"challenge_type" gorm:"not null"`                     // http-01, dns-01, tls-alpn-01
	Status            string                 `json:"status" gorm:"not null;default:'pending'"`           // pending, processing, completed, failed
	CertificateType   string                 `json:"certificate_type" gorm:"default:'domain_validated'"` // domain_validated, organization_validated, extended_validation
	KeyType           string                 `json:"key_type" gorm:"default:'rsa'"`                      // rsa, ecdsa
	KeySize           int                    `json:"key_size" gorm:"default:2048"`
	Domains           []string               `json:"domains" gorm:"type:text[]"`   // All domains in certificate (SAN)
	CSR               string                 `json:"csr" gorm:"type:text"`         // Certificate Signing Request
	PrivateKey        string                 `json:"private_key" gorm:"type:text"` // Should be encrypted
	ACMEOrderURL      string                 `json:"acme_order_url"`
	ACMEOrderStatus   string                 `json:"acme_order_status"`
	ChallengeData     map[string]interface{} `json:"challenge_data" gorm:"type:jsonb;default:'{}'"`
	ValidationRecords map[string]interface{} `json:"validation_records" gorm:"type:jsonb;default:'{}'"`
	CertificateChain  string                 `json:"certificate_chain" gorm:"type:text"`
	CertificateURL    string                 `json:"certificate_url"`
	AttemptCount      int                    `json:"attempt_count" gorm:"default:0"`
	MaxAttempts       int                    `json:"max_attempts" gorm:"default:3"`
	LastAttemptAt     *time.Time             `json:"last_attempt_at"`
	CompletedAt       *time.Time             `json:"completed_at"`
	ExpiresAt         *time.Time             `json:"expires_at"`
	ErrorMessage      string                 `json:"error_message"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`

	// Relationships
	Domain      TenantDomain    `json:"domain" gorm:"foreignKey:DomainID"`
	Certificate *SSLCertificate `json:"certificate" gorm:"foreignKey:RequestID"`
}

// DomainHealthCheck represents domain health monitoring
type DomainHealthCheck struct {
	ID               uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DomainID         uuid.UUID              `json:"domain_id" gorm:"type:uuid;not null"`
	CheckType        string                 `json:"check_type" gorm:"not null"` // http, https, dns, ssl, ping
	Status           string                 `json:"status" gorm:"not null"`     // healthy, unhealthy, unknown
	ResponseTime     int                    `json:"response_time"`              // milliseconds
	StatusCode       *int                   `json:"status_code"`                // HTTP status code
	ResponseBody     string                 `json:"response_body"`
	ErrorMessage     string                 `json:"error_message"`
	CheckFrequency   int                    `json:"check_frequency" gorm:"default:300"`  // seconds
	TimeoutThreshold int                    `json:"timeout_threshold" gorm:"default:30"` // seconds
	HealthData       map[string]interface{} `json:"health_data" gorm:"type:jsonb;default:'{}'"`
	AlertsEnabled    bool                   `json:"alerts_enabled" gorm:"default:true"`
	LastAlertAt      *time.Time             `json:"last_alert_at"`
	NextCheckAt      *time.Time             `json:"next_check_at"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`

	// Relationships
	Domain TenantDomain `json:"domain" gorm:"foreignKey:DomainID"`
}

// DomainRoutingCache caches domain routing information for performance
type DomainRoutingCache struct {
	Domain         string                 `json:"domain" gorm:"primaryKey"`
	TenantID       string                 `json:"tenant_id" gorm:"not null"`
	BackendService string                 `json:"backend_service" gorm:"not null;default:'ilms-api'"`
	RoutingConfig  map[string]interface{} `json:"routing_config" gorm:"type:jsonb;default:'{}'"`
	CacheExpiresAt time.Time              `json:"cache_expires_at" gorm:"not null"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	KeycloakUserID string                 `json:"keycloak_user_id" gorm:"unique"`
	Email          string                 `json:"email" gorm:"unique;not null"`
	Username       string                 `json:"username" gorm:"unique"`
	FirstName      string                 `json:"first_name"`
	LastName       string                 `json:"last_name"`
	Phone          string                 `json:"phone"`
	Avatar         *string                `json:"avatar"`
	AvatarURL      *string                `json:"avatar_url"`
	Status         string                 `json:"status" gorm:"not null;default:'active'"`
	EmailVerified  bool                   `json:"email_verified" gorm:"default:false"`
	PhoneVerified  bool                   `json:"phone_verified" gorm:"default:false"`
	LastLoginAt    *time.Time             `json:"last_login_at"`
	LoginCount     int                    `json:"login_count" gorm:"default:0"`
	Preferences    map[string]interface{} `json:"preferences" gorm:"type:jsonb;default:'{}'"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	DeletedAt      *time.Time             `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	TenantUsers []TenantUser  `json:"tenant_users" gorm:"foreignKey:UserID"`
	UserRoles   []UserRole    `json:"user_roles" gorm:"foreignKey:UserID"`
	Sessions    []UserSession `json:"sessions" gorm:"foreignKey:UserID"`
}

// UserSession represents user session management
type UserSession struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null"`
	TenantID       *uuid.UUID `json:"tenant_id" gorm:"type:uuid"`
	SessionToken   string     `json:"session_token" gorm:"unique;not null"`
	RefreshToken   string     `json:"refresh_token"`
	IPAddress      string     `json:"ip_address"`
	UserAgent      string     `json:"user_agent"`
	ExpiresAt      time.Time  `json:"expires_at" gorm:"not null"`
	LastAccessedAt time.Time  `json:"last_accessed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Tenant *Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// TenantUser represents user-tenant relationships with roles
type TenantUser struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID string    `json:"tenant_id" gorm:"type:varchar(50);not null"`
	UserID   uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Role     string    `json:"role" gorm:"not null"`
	Status   string    `json:"status" gorm:"not null;default:'active'"`
	JoinedAt time.Time `json:"joined_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
	User   User   `json:"user" gorm:"foreignKey:UserID"`
}

// APIKey represents API keys for tenant authentication
type APIKey struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
	Name        string     `json:"name" gorm:"not null"`
	Key         string     `json:"key" gorm:"unique;not null"`
	Permissions []string   `json:"permissions" gorm:"type:text[]"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// AuditLog represents audit logging for tenant activities
type AuditLog struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   string     `json:"tenant_id" gorm:"type:varchar(50);not null"`
	UserID     *uuid.UUID `json:"user_id" gorm:"type:uuid"`
	Action     string     `json:"action" gorm:"not null"`
	Resource   string     `json:"resource" gorm:"not null"`
	ResourceID *string    `json:"resource_id"`
	Details    *Details   `json:"details" gorm:"type:jsonb"`
	IPAddress  string     `json:"ip_address"`
	UserAgent  string     `json:"user_agent"`
	CreatedAt  time.Time  `json:"created_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
	User   *User  `json:"user" gorm:"foreignKey:UserID"`
}

// Settings represents tenant-specific settings
type Settings struct {
	Theme        string                 `json:"theme,omitempty"`
	Language     string                 `json:"language,omitempty"`
	Timezone     string                 `json:"timezone,omitempty"`
	Features     []string               `json:"features,omitempty"`
	Integrations map[string]interface{} `json:"integrations,omitempty"`
	Branding     *Branding              `json:"branding,omitempty"`
}

// Branding represents tenant branding settings
type Branding struct {
	Logo           string `json:"logo,omitempty"`
	FaviconURL     string `json:"favicon_url,omitempty"`
	PrimaryColor   string `json:"primary_color,omitempty"`
	SecondaryColor string `json:"secondary_color,omitempty"`
}

// Details represents flexible JSON details for audit logs
type Details struct {
	Before   interface{}            `json:"before,omitempty"`
	After    interface{}            `json:"after,omitempty"`
	Changes  map[string]interface{} `json:"changes,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// UserPreference represents user preferences per tenant
type UserPreference struct {
	ID        uuid.UUID   `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID   `json:"user_id" gorm:"type:uuid;not null"`
	TenantID  *uuid.UUID  `json:"tenant_id" gorm:"type:uuid"` // NULL for global preferences
	Category  string      `json:"category" gorm:"not null"`   // e.g., "ui", "notifications", "privacy"
	Key       string      `json:"key" gorm:"not null"`        // e.g., "theme", "language", "timezone"
	Value     interface{} `json:"value" gorm:"type:jsonb"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`

	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Tenant *Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// FileStorageConfig represents storage configuration per tenant
type FileStorageConfig struct {
	ID               uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID         *uuid.UUID             `json:"tenant_id" gorm:"type:uuid"`
	StorageType      string                 `json:"storage_type" gorm:"not null;default:'local'"` // local, s3, minio, azure
	Config           map[string]interface{} `json:"config" gorm:"type:jsonb;default:'{}'"`
	IsActive         bool                   `json:"is_active" gorm:"default:true"`
	MaxFileSize      int64                  `json:"max_file_size" gorm:"default:104857600"` // 100MB
	AllowedMimeTypes []string               `json:"allowed_mime_types" gorm:"type:text[]"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`

	// Relationships
	Tenant *Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// File represents uploaded files with enhanced features
type File struct {
	ID               uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID         *uuid.UUID             `json:"tenant_id" gorm:"type:uuid"`
	UserID           uuid.UUID              `json:"user_id" gorm:"type:uuid;not null"`
	FileName         string                 `json:"file_name" gorm:"not null"`
	OriginalName     string                 `json:"original_name" gorm:"not null"`
	FilePath         string                 `json:"file_path" gorm:"column:file_path;not null"`
	StoragePath      string                 `json:"storage_path" gorm:"not null"`
	StorageType      string                 `json:"storage_type" gorm:"default:'local'"` // local, s3, minio, azure
	MimeType         string                 `json:"mime_type" gorm:"not null"`
	Size             int64                  `json:"size" gorm:"not null"`
	URL              string                 `json:"url"`
	Category         string                 `json:"category" gorm:"default:'general'"` // avatar, document, image, video, etc.
	Tags             []string               `json:"tags" gorm:"type:text[]"`
	IsPublic         bool                   `json:"is_public" gorm:"default:false"`
	IsProcessed      bool                   `json:"is_processed" gorm:"default:false"`
	ProcessingStatus string                 `json:"processing_status" gorm:"default:'pending'"` // pending, processing, completed, failed
	Status           string                 `json:"status" gorm:"default:'pending'"`            // pending, available, quarantined, deleted
	Checksum         string                 `json:"checksum"`                                   // SHA256 hash
	VirusScanned     bool                   `json:"virus_scanned" gorm:"default:false"`
	VirusScanStatus  string                 `json:"virus_scan_status" gorm:"default:'pending'"` // pending, scanning, clean, infected, error
	VirusScanResult  *string                `json:"virus_scan_result"`
	VirusScanDetails map[string]interface{} `json:"virus_scan_details" gorm:"type:jsonb;default:'{}'"`
	VersionNumber    int                    `json:"version_number" gorm:"default:1"`
	ExpiryDate       *time.Time             `json:"expiry_date"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	DeletedAt        *time.Time             `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	User     User          `json:"user" gorm:"foreignKey:UserID"`
	Tenant   *Tenant       `json:"tenant" gorm:"foreignKey:TenantID"`
	Versions []FileVersion `json:"versions" gorm:"foreignKey:FileID"`
	Shares   []FileShare   `json:"shares" gorm:"foreignKey:FileID"`
}

// FileVersion represents file version history
type FileVersion struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FileID        uuid.UUID              `json:"file_id" gorm:"type:uuid;not null"`
	VersionNumber int                    `json:"version_number" gorm:"not null;default:1"`
	FilePath      string                 `json:"file_path" gorm:"not null"`
	Size          int64                  `json:"size" gorm:"not null"`
	Checksum      string                 `json:"checksum"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedBy     uuid.UUID              `json:"created_by" gorm:"type:uuid;not null"`
	CreatedAt     time.Time              `json:"created_at"`

	// Relationships
	File          File `json:"file" gorm:"foreignKey:FileID"`
	CreatedByUser User `json:"created_by_user" gorm:"foreignKey:CreatedBy"`
}

// FileShare represents file sharing permissions
type FileShare struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FileID        uuid.UUID              `json:"file_id" gorm:"type:uuid;not null"`
	SharedBy      uuid.UUID              `json:"shared_by" gorm:"type:uuid;not null"`
	SharedWith    *uuid.UUID             `json:"shared_with" gorm:"type:uuid"`     // NULL for public shares
	ShareType     string                 `json:"share_type" gorm:"default:'read'"` // read, write, download
	AccessToken   string                 `json:"access_token" gorm:"unique"`
	PasswordHash  string                 `json:"password_hash"`                   // for password-protected shares
	MaxDownloads  int                    `json:"max_downloads" gorm:"default:-1"` // -1 for unlimited
	DownloadCount int                    `json:"download_count" gorm:"default:0"`
	ExpiresAt     *time.Time             `json:"expires_at"`
	IsActive      bool                   `json:"is_active" gorm:"default:true"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`

	// Relationships
	File           File  `json:"file" gorm:"foreignKey:FileID"`
	SharedByUser   User  `json:"shared_by_user" gorm:"foreignKey:SharedBy"`
	SharedWithUser *User `json:"shared_with_user" gorm:"foreignKey:SharedWith"`
}

// FileUploadSession represents chunked upload sessions
type FileUploadSession struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionToken   string                 `json:"session_token" gorm:"unique;not null"`
	TenantID       *uuid.UUID             `json:"tenant_id" gorm:"type:uuid"`
	UserID         uuid.UUID              `json:"user_id" gorm:"type:uuid;not null"`
	FileName       string                 `json:"file_name" gorm:"not null"`
	FileSize       int64                  `json:"file_size" gorm:"not null"`
	ChunkSize      int                    `json:"chunk_size" gorm:"default:1048576"` // 1MB chunks
	TotalChunks    int                    `json:"total_chunks" gorm:"not null"`
	UploadedChunks int                    `json:"uploaded_chunks" gorm:"default:0"`
	UploadPath     string                 `json:"upload_path"`
	Status         string                 `json:"status" gorm:"default:'pending'"` // pending, uploading, completed, failed, cancelled
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ExpiresAt      time.Time              `json:"expires_at"`

	// Relationships
	User   User    `json:"user" gorm:"foreignKey:UserID"`
	Tenant *Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// FileProcessingJob represents async file processing jobs
type FileProcessingJob struct {
	ID                  uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FileID              uuid.UUID              `json:"file_id" gorm:"type:uuid;not null"`
	JobType             string                 `json:"job_type" gorm:"not null"`                  // resize, crop, compress, thumbnail, virus_scan
	Parameters          []byte                 `json:"parameters" gorm:"type:jsonb;default:'{}'"` // JSON parameters for job
	Status              string                 `json:"status" gorm:"default:'pending'"`           // pending, processing, completed, failed
	Progress            int                    `json:"progress" gorm:"default:0"`                 // 0-100
	RetryCount          int                    `json:"retry_count" gorm:"default:0"`
	LastError           string                 `json:"last_error"`
	Result              map[string]interface{} `json:"result" gorm:"type:jsonb;default:'{}'"`
	ProcessingStartedAt *time.Time             `json:"processing_started_at"`
	CompletedAt         *time.Time             `json:"completed_at"`
	FailedAt            *time.Time             `json:"failed_at"`
	CreatedAt           time.Time              `json:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at"`

	// Relationships
	File File `json:"file" gorm:"foreignKey:FileID"`
}

// FileAccessLog represents audit trail for file operations
type FileAccessLog struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FileID     uuid.UUID              `json:"file_id" gorm:"type:uuid;not null"`
	UserID     *uuid.UUID             `json:"user_id" gorm:"type:uuid"`
	Action     string                 `json:"action" gorm:"not null"` // upload, download, view, share, delete
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	ShareToken string                 `json:"share_token"` // if accessed via share
	Metadata   map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt  time.Time              `json:"created_at"`

	// Relationships
	File File  `json:"file" gorm:"foreignKey:FileID"`
	User *User `json:"user" gorm:"foreignKey:UserID"`
}

// TenantSettings represents tenant-specific system settings
type TenantSettings struct {
	Theme                string                 `json:"theme,omitempty"`
	Language             string                 `json:"language,omitempty"`
	Timezone             string                 `json:"timezone,omitempty"`
	DateFormat           string                 `json:"date_format,omitempty"`
	TimeFormat           string                 `json:"time_format,omitempty"`
	NumberFormat         string                 `json:"number_format,omitempty"`
	Features             []string               `json:"features,omitempty"`
	APILimits            map[string]interface{} `json:"api_limits,omitempty"`
	SecuritySettings     map[string]interface{} `json:"security_settings,omitempty"`
	NotificationSettings map[string]interface{} `json:"notification_settings,omitempty"`
	IntegrationSettings  map[string]interface{} `json:"integration_settings,omitempty"`
}

// TenantConfiguration represents tenant operational configuration
type TenantConfiguration struct {
	MaxUsers            int                    `json:"max_users,omitempty"`
	MaxStorage          int64                  `json:"max_storage,omitempty"`
	MaxAPICallsPerMonth int                    `json:"max_api_calls_per_month,omitempty"`
	EnabledModules      []string               `json:"enabled_modules,omitempty"`
	CustomDomainEnabled bool                   `json:"custom_domain_enabled,omitempty"`
	SSLEnabled          bool                   `json:"ssl_enabled,omitempty"`
	BackupEnabled       bool                   `json:"backup_enabled,omitempty"`
	MonitoringEnabled   bool                   `json:"monitoring_enabled,omitempty"`
	AuditLoggingEnabled bool                   `json:"audit_logging_enabled,omitempty"`
	DataRetentionDays   int                    `json:"data_retention_days,omitempty"`
	AllowedFileTypes    []string               `json:"allowed_file_types,omitempty"`
	MaxFileSize         int64                  `json:"max_file_size,omitempty"`
	NetworkSettings     map[string]interface{} `json:"network_settings,omitempty"`
	DatabaseSettings    map[string]interface{} `json:"database_settings,omitempty"`
	CacheSettings       map[string]interface{} `json:"cache_settings,omitempty"`
}

// TenantBranding represents tenant white-label branding settings
type TenantBranding struct {
	Logo              string                 `json:"logo,omitempty"`
	LogoURL           string                 `json:"logo_url,omitempty"`
	FaviconURL        string                 `json:"favicon_url,omitempty"`
	PrimaryColor      string                 `json:"primary_color,omitempty"`
	SecondaryColor    string                 `json:"secondary_color,omitempty"`
	AccentColor       string                 `json:"accent_color,omitempty"`
	BackgroundColor   string                 `json:"background_color,omitempty"`
	TextColor         string                 `json:"text_color,omitempty"`
	FontFamily        string                 `json:"font_family,omitempty"`
	CustomCSS         string                 `json:"custom_css,omitempty"`
	CustomJavaScript  string                 `json:"custom_javascript,omitempty"`
	FooterText        string                 `json:"footer_text,omitempty"`
	TermsOfServiceURL string                 `json:"terms_of_service_url,omitempty"`
	PrivacyPolicyURL  string                 `json:"privacy_policy_url,omitempty"`
	SupportURL        string                 `json:"support_url,omitempty"`
	DocumentationURL  string                 `json:"documentation_url,omitempty"`
	CustomDomainName  string                 `json:"custom_domain_name,omitempty"`
	EmailFromName     string                 `json:"email_from_name,omitempty"`
	EmailFromAddress  string                 `json:"email_from_address,omitempty"`
	SocialMediaLinks  map[string]interface{} `json:"social_media_links,omitempty"`
	CustomMetaTags    map[string]interface{} `json:"custom_meta_tags,omitempty"`
	PWASettings       map[string]interface{} `json:"pwa_settings,omitempty"`
}

// TenantBilling represents tenant billing and subscription information
type TenantBilling struct {
	StripeCustomerID     string                 `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID string                 `json:"stripe_subscription_id,omitempty"`
	BillingEmail         string                 `json:"billing_email,omitempty"`
	BillingName          string                 `json:"billing_name,omitempty"`
	BillingAddress       map[string]interface{} `json:"billing_address,omitempty"`
	TaxInfo              map[string]interface{} `json:"tax_info,omitempty"`
	PaymentMethodID      string                 `json:"payment_method_id,omitempty"`
	DefaultPaymentMethod string                 `json:"default_payment_method,omitempty"`
	BillingCycle         string                 `json:"billing_cycle,omitempty"`
	NextBillingDate      *time.Time             `json:"next_billing_date,omitempty"`
	LastBillingDate      *time.Time             `json:"last_billing_date,omitempty"`
	CurrentPeriodStart   *time.Time             `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time             `json:"current_period_end,omitempty"`
	TrialPeriodDays      int                    `json:"trial_period_days,omitempty"`
	GracePeriodDays      int                    `json:"grace_period_days,omitempty"`
	InvoiceSettings      map[string]interface{} `json:"invoice_settings,omitempty"`
	UsageTracking        map[string]interface{} `json:"usage_tracking,omitempty"`
	CreditBalance        float64                `json:"credit_balance,omitempty"`
	AutoBilling          bool                   `json:"auto_billing,omitempty"`
	ProrationBehavior    string                 `json:"proration_behavior,omitempty"`
}

// TenantOnboardingLog represents tenant onboarding progress tracking
type TenantOnboardingLog struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID     uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null"`
	Step         int                    `json:"step" gorm:"not null"`
	StepName     string                 `json:"step_name" gorm:"not null"`
	Status       string                 `json:"status" gorm:"not null"` // 'pending', 'in_progress', 'completed', 'skipped', 'failed'
	Data         map[string]interface{} `json:"data" gorm:"type:jsonb"`
	ErrorMessage string                 `json:"error_message"`
	StartedAt    *time.Time             `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// TenantInvitation represents invitations to join a tenant
type TenantInvitation struct {
	ID         uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null"`
	Email      string                 `json:"email" gorm:"not null"`
	Role       string                 `json:"role" gorm:"not null"`
	InvitedBy  uuid.UUID              `json:"invited_by" gorm:"type:uuid;not null"`
	Token      string                 `json:"token" gorm:"unique;not null"`
	Status     string                 `json:"status" gorm:"default:'pending'"` // 'pending', 'accepted', 'expired', 'revoked'
	ExpiresAt  time.Time              `json:"expires_at"`
	AcceptedAt *time.Time             `json:"accepted_at"`
	Metadata   map[string]interface{} `json:"metadata" gorm:"type:jsonb"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`

	// Relationships
	Tenant        Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
	InvitedByUser User   `json:"invited_by_user" gorm:"foreignKey:InvitedBy"`
}

// TenantUsageMetrics represents tenant usage tracking
type TenantUsageMetrics struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID       string                 `json:"tenant_id" gorm:"type:varchar(50);not null"`
	MetricType     string                 `json:"metric_type" gorm:"not null"` // 'api_calls', 'storage_used', 'users_count', etc.
	Value          float64                `json:"value" gorm:"not null"`
	Unit           string                 `json:"unit" gorm:"not null"` // 'count', 'bytes', 'percentage', etc.
	RecordedAt     time.Time              `json:"recorded_at" gorm:"not null"`
	PeriodStart    time.Time              `json:"period_start"`
	PeriodEnd      time.Time              `json:"period_end"`
	AdditionalData map[string]interface{} `json:"additional_data" gorm:"type:jsonb"`
	CreatedAt      time.Time              `json:"created_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// Constants for status values
const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusSuspended = "suspended"
	StatusPending   = "pending"
)

// File status constants
const (
	FileStatusPending     = "pending"
	FileStatusProcessing  = "processing"
	FileStatusAvailable   = "available"
	FileStatusQuarantined = "quarantined"
	FileStatusDeleted     = "deleted"
)

// File processing job status constants
const (
	JobStatusPending    = "pending"
	JobStatusProcessing = "processing"
	JobStatusCompleted  = "completed"
	JobStatusFailed     = "failed"
)

// File processing job type constants
const (
	JobTypeVirusScan           = "virus_scan"
	JobTypeImageResize         = "image_resize"
	JobTypeImageCrop           = "image_crop"
	JobTypeThumbnailGeneration = "thumbnail_generation"
	JobTypeMetadataExtraction  = "metadata_extraction"
)

// Tenant status constants
const (
	TenantStatusPending    = "pending"
	TenantStatusActive     = "active"
	TenantStatusSuspended  = "suspended"
	TenantStatusCancelled  = "cancelled"
	TenantStatusTrialEnded = "trial_ended"
)

// Tenant plan constants
const (
	TenantPlanStarter      = "starter"
	TenantPlanProfessional = "professional"
	TenantPlanEnterprise   = "enterprise"
	TenantPlanCustom       = "custom"
)

// Onboarding status constants
const (
	OnboardingStatusPending    = "pending"
	OnboardingStatusInProgress = "in_progress"
	OnboardingStatusCompleted  = "completed"
	OnboardingStatusSkipped    = "skipped"
	OnboardingStatusFailed     = "failed"
)

// Onboarding step constants
const (
	OnboardingStepWelcome      = 1
	OnboardingStepBasicInfo    = 2
	OnboardingStepBranding     = 3
	OnboardingStepDomain       = 4
	OnboardingStepUsers        = 5
	OnboardingStepIntegrations = 6
	OnboardingStepCompleted    = 7
)

// Subscription status constants
const (
	SubscriptionStatusTrial     = "trial"
	SubscriptionStatusActive    = "active"
	SubscriptionStatusPastDue   = "past_due"
	SubscriptionStatusCancelled = "cancelled"
	SubscriptionStatusExpired   = "expired"
)

// Invitation status constants
const (
	InvitationStatusPending  = "pending"
	InvitationStatusAccepted = "accepted"
	InvitationStatusExpired  = "expired"
	InvitationStatusRevoked  = "revoked"
)

// Role represents roles in the system
type Role struct {
	ID          uuid.UUID    `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string       `json:"name" gorm:"unique;not null"`
	Description string       `json:"description"`
	IsSystem    bool         `json:"is_system" gorm:"default:false"` // System roles cannot be deleted
	TenantID    *uuid.UUID   `json:"tenant_id" gorm:"type:uuid"`     // NULL for system roles
	Permissions []Permission `json:"permissions" gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`

	// Relationships
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// Permission represents permissions in the system
type Permission struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name        string    `json:"name" gorm:"unique;not null"`
	Resource    string    `json:"resource" gorm:"not null"`
	Action      string    `json:"action" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserRole represents user-role assignments
type UserRole struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	RoleID    uuid.UUID `json:"role_id" gorm:"type:uuid;not null"`
	TenantID  uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Status    string    `json:"status" gorm:"not null;default:'active'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	User   User   `json:"user" gorm:"foreignKey:UserID"`
	Role   Role   `json:"role" gorm:"foreignKey:RoleID"`
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// Constants for system roles
const (
	RoleSystemAdmin   = "system_admin"
	RoleSystemManager = "system_manager"
	RoleTenantAdmin   = "tenant_admin"
	RoleTenantManager = "tenant_manager"
	RoleTenantUser    = "tenant_user"
	RoleUser          = "user"
	RoleViewer        = "viewer"
)

// Constants for resources
const (
	ResourceTenant     = "tenant"
	ResourceUser       = "user"
	ResourceFile       = "file"
	ResourceRole       = "role"
	ResourcePermission = "permission"
	ResourceAuditLog   = "audit_log"
	ResourceAPIKey     = "api_key"
	ResourceDomain     = "domain"
	ResourceSettings   = "settings"
)

// Constants for actions
const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
	ActionLogin  = "login"
	ActionLogout = "logout"
	ActionView   = "view"
	ActionManage = "manage"
	ActionAssign = "assign"
)

// Constants for permission names
const (
	// System permissions
	PermSystemManageTenants  = "system:manage_tenants"
	PermSystemManageUsers    = "system:manage_users"
	PermSystemViewAuditLogs  = "system:view_audit_logs"
	PermSystemManageSettings = "system:manage_settings"

	// Tenant permissions
	PermTenantManageUsers    = "tenant:manage_users"
	PermTenantManageRoles    = "tenant:manage_roles"
	PermTenantManageSettings = "tenant:manage_settings"
	PermTenantManageDomains  = "tenant:manage_domains"
	PermTenantViewAuditLogs  = "tenant:view_audit_logs"
	PermTenantManageAPIKeys  = "tenant:manage_api_keys"

	// User permissions
	PermUserReadProfile   = "user:read_profile"
	PermUserUpdateProfile = "user:update_profile"
	PermUserManageFiles   = "user:manage_files"
	PermUserViewFiles     = "user:view_files"
)

// ===========================
// GraphQL Federation Models
// ===========================

// GraphQLSchema represents a GraphQL schema for federation
type GraphQLSchema struct {
	ID               uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceName      string                 `json:"service_name" gorm:"not null"`
	ServiceVersion   string                 `json:"service_version" gorm:"not null"`
	SchemaSDL        string                 `json:"schema_sdl" gorm:"type:text;not null"`
	SchemaHash       string                 `json:"schema_hash" gorm:"not null"`
	Status           string                 `json:"status" gorm:"default:'active'"`
	IsActive         bool                   `json:"is_active" gorm:"default:true"`
	IsValid          bool                   `json:"is_valid" gorm:"default:true"`
	ValidationErrors []string               `json:"validation_errors" gorm:"type:jsonb;default:'[]'"`
	Metadata         map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// FederationService represents a service in the federation
type FederationService struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceName     string                 `json:"service_name" gorm:"unique;not null"`
	ServiceURL      string                 `json:"service_url" gorm:"not null"`
	HealthCheckURL  string                 `json:"health_check_url"`
	SchemaID        *uuid.UUID             `json:"schema_id" gorm:"type:uuid"`
	Status          string                 `json:"status" gorm:"default:'healthy'"`
	LastHealthCheck *time.Time             `json:"last_health_check"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	Tags            []string               `json:"tags" gorm:"type:text[]"`
	Weight          int                    `json:"weight" gorm:"default:100"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`

	// Relationships
	Schema *GraphQLSchema `json:"schema,omitempty" gorm:"foreignKey:SchemaID"`
}

// FederationComposition represents a composed federated schema
type FederationComposition struct {
	ID                 uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CompositionName    string                 `json:"composition_name" gorm:"not null"`
	CompositionVersion string                 `json:"composition_version" gorm:"not null"`
	Version            string                 `json:"version" gorm:"not null"`
	ComposedSchema     string                 `json:"composed_schema" gorm:"type:text;not null"`
	Services           []string               `json:"services" gorm:"type:jsonb;not null"`
	ServiceSchemas     []ServiceSchemaRef     `json:"service_schemas" gorm:"type:jsonb;not null"`
	Status             string                 `json:"status" gorm:"default:'active'"`
	ValidationErrors   []string               `json:"validation_errors" gorm:"type:jsonb;default:'[]'"`
	Warnings           []string               `json:"warnings" gorm:"type:jsonb;default:'[]'"`
	Configuration      map[string]interface{} `json:"configuration" gorm:"type:jsonb;default:'{}'"`
	ValidationResult   map[string]interface{} `json:"validation_result" gorm:"type:jsonb;default:'{}'"`
	CreatedAt          time.Time              `json:"created_at"`
	DeployedAt         *time.Time             `json:"deployed_at"`
}

// ServiceSchemaRef represents a reference to a service schema in a composition
type ServiceSchemaRef struct {
	ServiceName string    `json:"service_name"`
	SchemaID    uuid.UUID `json:"schema_id"`
	Version     string    `json:"version"`
}

// GraphQLQueryMetrics represents query performance metrics
type GraphQLQueryMetrics struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID        *uuid.UUID             `json:"tenant_id" gorm:"type:uuid"`
	QueryHash       string                 `json:"query_hash" gorm:"not null"`
	QueryName       string                 `json:"query_name"`
	OperationType   string                 `json:"operation_type" gorm:"not null"`
	ExecutionTime   time.Duration          `json:"execution_time" gorm:"not null"`
	ExecutionTimeMs int                    `json:"execution_time_ms" gorm:"not null"`
	QueryComplexity int                    `json:"query_complexity"`
	ComplexityScore int                    `json:"complexity_score"`
	DepthScore      int                    `json:"depth_score"`
	FieldCount      int                    `json:"field_count"`
	ServicesCalled  []string               `json:"services_called" gorm:"type:jsonb;default:'[]'"`
	ServiceCalls    []ServiceCallDetail    `json:"service_calls" gorm:"type:jsonb;default:'[]'"`
	ErrorCount      int                    `json:"error_count" gorm:"default:0"`
	CacheHit        bool                   `json:"cache_hit" gorm:"default:false"`
	Metadata        map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt       time.Time              `json:"created_at"`

	// Relationships
	Tenant *Tenant `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

// ServiceCallDetail represents details of a service call during query execution
type ServiceCallDetail struct {
	ServiceName     string `json:"service_name"`
	ExecutionTimeMs int    `json:"execution_time_ms"`
	FieldCount      int    `json:"field_count"`
	ErrorCount      int    `json:"error_count"`
}

// SchemaChangeEvent represents schema change events for auditing
type SchemaChangeEvent struct {
	ID              uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ServiceName     string                 `json:"service_name" gorm:"not null"`
	ChangeType      string                 `json:"change_type" gorm:"not null"`
	OldSchemaID     *uuid.UUID             `json:"old_schema_id" gorm:"type:uuid"`
	NewSchemaID     *uuid.UUID             `json:"new_schema_id" gorm:"type:uuid"`
	ChangeDetails   map[string]interface{} `json:"change_details" gorm:"type:jsonb;default:'{}'"`
	BreakingChanges []string               `json:"breaking_changes" gorm:"type:jsonb;default:'[]'"`
	ImpactAnalysis  map[string]interface{} `json:"impact_analysis" gorm:"type:jsonb;default:'{}'"`
	CreatedAt       time.Time              `json:"created_at"`
	ProcessedAt     *time.Time             `json:"processed_at"`

	// Relationships
	OldSchema *GraphQLSchema `json:"old_schema,omitempty" gorm:"foreignKey:OldSchemaID"`
	NewSchema *GraphQLSchema `json:"new_schema,omitempty" gorm:"foreignKey:NewSchemaID"`
}

// FederationGatewayConfig represents gateway configuration
type FederationGatewayConfig struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ConfigKey     string                 `json:"config_key" gorm:"unique;not null"`
	ConfigName    string                 `json:"config_name" gorm:"unique;not null"`
	ConfigValue   map[string]interface{} `json:"config_value" gorm:"type:jsonb;not null"`
	GatewayConfig map[string]interface{} `json:"gateway_config" gorm:"type:jsonb;not null"`
	IsActive      bool                   `json:"is_active" gorm:"default:false"`
	CreatedAt     time.Time              `json:"created_at"`
	ActivatedAt   *time.Time             `json:"activated_at"`
}

// Federation service status constants
const (
	ServiceStatusHealthy     = "healthy"
	ServiceStatusUnhealthy   = "unhealthy"
	ServiceStatusUnknown     = "unknown"
	ServiceStatusMaintenance = "maintenance"
)

// Schema status constants
const (
	SchemaStatusActive   = "active"
	SchemaStatusInactive = "inactive"
	SchemaStatusInvalid  = "invalid"
)

// Composition status constants
const (
	CompositionStatusActive   = "active"
	CompositionStatusInactive = "inactive"
	CompositionStatusInvalid  = "invalid"
	CompositionStatusFailed   = "failed"
)

// Schema change event types
const (
	ChangeTypeSchemaUpdated       = "schema_updated"
	ChangeTypeServiceRegistered   = "service_registered"
	ChangeTypeServiceDeregistered = "service_deregistered"
	ChangeTypeComposition         = "composition"
)

// GraphQL operation types
const (
	OperationTypeQuery        = "query"
	OperationTypeMutation     = "mutation"
	OperationTypeSubscription = "subscription"
)

// Filter types for domain repositories

type DomainRegistrationFilter struct {
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	Status    *string    `json:"status,omitempty"`
	Provider  *string    `json:"provider,omitempty"`
	Search    *string    `json:"search,omitempty"`
	AutoRenew *bool      `json:"auto_renew,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order" validate:"oneof=asc desc"`
}

type DNSRecordFilter struct {
	DomainID  *uuid.UUID `json:"domain_id,omitempty"`
	Type      *string    `json:"type,omitempty"`
	Name      *string    `json:"name,omitempty"`
	IsManaged *bool      `json:"is_managed,omitempty"`
	Status    *string    `json:"status,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order" validate:"oneof=asc desc"`
}

type DomainHealthCheckFilter struct {
	DomainID  *uuid.UUID `json:"domain_id,omitempty"`
	CheckType *string    `json:"check_type,omitempty"`
	Status    *string    `json:"status,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order" validate:"oneof=asc desc"`
}

// ===========================
// POS (Point of Sale) Models
// ===========================

// ProductCategory represents hierarchical product categories
type ProductCategory struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    string                 `json:"tenant_id" gorm:"type:varchar(50);not null"`
	Name        string                 `json:"name" gorm:"not null"`
	Description string                 `json:"description"`
	Slug        string                 `json:"slug" gorm:"not null"`
	ParentID    *uuid.UUID             `json:"parent_id" gorm:"type:uuid"`
	ImageURL    string                 `json:"image_url"`
	SortOrder   int                    `json:"sort_order" gorm:"default:0"`
	IsActive    bool                   `json:"is_active" gorm:"default:true"`
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	DeletedAt   *time.Time             `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Tenant   Tenant            `json:"tenant" gorm:"foreignKey:TenantID"`
	Parent   *ProductCategory  `json:"parent" gorm:"foreignKey:ParentID"`
	Children []ProductCategory `json:"children" gorm:"foreignKey:ParentID"`
	Products []Product         `json:"products" gorm:"foreignKey:CategoryID"`
}

// Product represents products in the POS system
type Product struct {
	ID               uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID         string     `json:"tenant_id" gorm:"type:varchar(50);not null"`
	CategoryID       *uuid.UUID `json:"category_id" gorm:"type:uuid"`
	SKU              string     `json:"sku" gorm:"not null"`
	Name             string     `json:"name" gorm:"not null"`
	Description      string     `json:"description"`
	ShortDescription string     `json:"short_description"`
	ProductType      string     `json:"product_type" gorm:"default:'simple'"` // simple, variable, grouped, external
	Status           string     `json:"status" gorm:"default:'draft'"`        // draft, published, private, pending
	Featured         bool       `json:"featured" gorm:"default:false"`

	// Pricing (using float64 for now, would need decimal package for production)
	RegularPrice float64 `json:"regular_price"`
	SalePrice    float64 `json:"sale_price"`
	CostPrice    float64 `json:"cost_price"`

	// Physical properties
	Weight float64 `json:"weight"`
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`

	// Inventory
	ManageStock       bool   `json:"manage_stock" gorm:"default:true"`
	StockQuantity     int    `json:"stock_quantity" gorm:"default:0"`
	LowStockThreshold int    `json:"low_stock_threshold" gorm:"default:5"`
	StockStatus       string `json:"stock_status" gorm:"default:'instock'"` // instock, outofstock, onbackorder
	Backorders        string `json:"backorders" gorm:"default:'no'"`        // no, notify, yes

	// SEO and visibility
	Slug            string `json:"slug" gorm:"not null"`
	MenuOrder       int    `json:"menu_order" gorm:"default:0"`
	MetaTitle       string `json:"meta_title"`
	MetaDescription string `json:"meta_description"`

	// Images and gallery
	FeaturedImage string   `json:"featured_image"`
	GalleryImages []string `json:"gallery_images" gorm:"type:text[]"`

	// Additional data
	Attributes map[string]interface{} `json:"attributes" gorm:"type:jsonb;default:'{}'"`
	Metadata   map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Tenant        Tenant             `json:"tenant" gorm:"foreignKey:TenantID"`
	Category      *ProductCategory   `json:"category" gorm:"foreignKey:CategoryID"`
	Variations    []ProductVariation `json:"variations" gorm:"foreignKey:ProductID"`
	InventoryLogs []InventoryLog     `json:"inventory_logs" gorm:"foreignKey:ProductID"`
}

// ProductVariation represents variations of configurable products
type ProductVariation struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProductID     uuid.UUID              `json:"product_id" gorm:"type:uuid;not null"`
	SKU           string                 `json:"sku" gorm:"not null"`
	RegularPrice  float64                `json:"regular_price"`
	SalePrice     float64                `json:"sale_price"`
	CostPrice     float64                `json:"cost_price"`
	StockQuantity int                    `json:"stock_quantity" gorm:"default:0"`
	StockStatus   string                 `json:"stock_status" gorm:"default:'instock'"`
	Weight        float64                `json:"weight"`
	Length        float64                `json:"length"`
	Width         float64                `json:"width"`
	Height        float64                `json:"height"`
	Image         string                 `json:"image"`
	Attributes    map[string]interface{} `json:"attributes" gorm:"type:jsonb;default:'{}'"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`

	// Relationships
	Product Product `json:"product" gorm:"foreignKey:ProductID"`
}

// InventoryLog tracks all inventory movements
type InventoryLog struct {
	ID             uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID       string                 `json:"tenant_id" gorm:"type:varchar(50);not null"`
	ProductID      uuid.UUID              `json:"product_id" gorm:"type:uuid;not null"`
	VariationID    *uuid.UUID             `json:"variation_id" gorm:"type:uuid"`
	Type           string                 `json:"type" gorm:"not null"` // 'in', 'out', 'adjustment', 'sale', 'return'
	Quantity       int                    `json:"quantity" gorm:"not null"`
	QuantityBefore int                    `json:"quantity_before" gorm:"not null"`
	QuantityAfter  int                    `json:"quantity_after" gorm:"not null"`
	Reason         string                 `json:"reason"`
	ReferenceID    *uuid.UUID             `json:"reference_id"`   // Reference to order, return, etc.
	ReferenceType  string                 `json:"reference_type"` // 'order', 'return', 'adjustment'
	CostPerUnit    float64                `json:"cost_per_unit"`
	TotalCost      float64                `json:"total_cost"`
	UserID         *uuid.UUID             `json:"user_id" gorm:"type:uuid"`
	Notes          string                 `json:"notes"`
	Metadata       map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt      time.Time              `json:"created_at"`

	// Relationships
	Tenant    Tenant            `json:"tenant" gorm:"foreignKey:TenantID"`
	Product   Product           `json:"product" gorm:"foreignKey:ProductID"`
	Variation *ProductVariation `json:"variation" gorm:"foreignKey:VariationID"`
	User      *User             `json:"user" gorm:"foreignKey:UserID"`
}

// Cart represents shopping carts
type Cart struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID      string                 `json:"tenant_id" gorm:"type:varchar(50);not null"`
	UserID        *uuid.UUID             `json:"user_id" gorm:"type:uuid"`
	SessionID     string                 `json:"session_id"`                     // For guest users
	Status        string                 `json:"status" gorm:"default:'active'"` // active, abandoned, converted
	Currency      string                 `json:"currency" gorm:"default:'USD'"`
	Subtotal      float64                `json:"subtotal" gorm:"default:0"`
	TaxTotal      float64                `json:"tax_total" gorm:"default:0"`
	ShippingTotal float64                `json:"shipping_total" gorm:"default:0"`
	DiscountTotal float64                `json:"discount_total" gorm:"default:0"`
	Total         float64                `json:"total" gorm:"default:0"`
	ExpiresAt     *time.Time             `json:"expires_at"`
	Metadata      map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`

	// Relationships
	Tenant Tenant     `json:"tenant" gorm:"foreignKey:TenantID"`
	User   *User      `json:"user" gorm:"foreignKey:UserID"`
	Items  []CartItem `json:"items" gorm:"foreignKey:CartID"`
}

// CartItem represents individual items in a cart
type CartItem struct {
	ID          uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	CartID      uuid.UUID              `json:"cart_id" gorm:"type:uuid;not null"`
	ProductID   uuid.UUID              `json:"product_id" gorm:"type:uuid;not null"`
	VariationID *uuid.UUID             `json:"variation_id" gorm:"type:uuid"`
	Quantity    int                    `json:"quantity" gorm:"not null;default:1"`
	UnitPrice   float64                `json:"unit_price" gorm:"not null"`
	TotalPrice  float64                `json:"total_price" gorm:"not null"`
	ProductData map[string]interface{} `json:"product_data" gorm:"type:jsonb"` // Snapshot of product data
	Metadata    map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`

	// Relationships
	Cart      Cart              `json:"cart" gorm:"foreignKey:CartID"`
	Product   Product           `json:"product" gorm:"foreignKey:ProductID"`
	Variation *ProductVariation `json:"variation" gorm:"foreignKey:VariationID"`
}

// Order represents customer orders
type Order struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    string     `json:"tenant_id" gorm:"type:varchar(50);not null"`
	OrderNumber string     `json:"order_number" gorm:"not null"`
	UserID      *uuid.UUID `json:"user_id" gorm:"type:uuid"`

	// Order status
	Status        string `json:"status" gorm:"default:'pending'"`         // pending, processing, shipped, delivered, cancelled, refunded
	PaymentStatus string `json:"payment_status" gorm:"default:'pending'"` // pending, paid, failed, refunded, partially_refunded

	// Customer information
	CustomerEmail   string                 `json:"customer_email"`
	CustomerPhone   string                 `json:"customer_phone"`
	BillingAddress  map[string]interface{} `json:"billing_address" gorm:"type:jsonb"`
	ShippingAddress map[string]interface{} `json:"shipping_address" gorm:"type:jsonb"`

	// Financial details
	Currency      string  `json:"currency" gorm:"default:'USD'"`
	Subtotal      float64 `json:"subtotal" gorm:"not null;default:0"`
	TaxTotal      float64 `json:"tax_total" gorm:"default:0"`
	ShippingTotal float64 `json:"shipping_total" gorm:"default:0"`
	DiscountTotal float64 `json:"discount_total" gorm:"default:0"`
	Total         float64 `json:"total" gorm:"not null;default:0"`

	// Payment information
	PaymentMethod      string `json:"payment_method"`
	PaymentMethodTitle string `json:"payment_method_title"`
	TransactionID      string `json:"transaction_id"`

	// Dates
	DateCreated   time.Time  `json:"date_created"`
	DateModified  time.Time  `json:"date_modified"`
	DateCompleted *time.Time `json:"date_completed"`
	DatePaid      *time.Time `json:"date_paid"`

	// Additional data
	CustomerNote string                 `json:"customer_note"`
	StaffNotes   string                 `json:"staff_notes"`
	Metadata     map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`

	// Relationships
	Tenant       Tenant               `json:"tenant" gorm:"foreignKey:TenantID"`
	User         *User                `json:"user" gorm:"foreignKey:UserID"`
	Items        []OrderItem          `json:"items" gorm:"foreignKey:OrderID"`
	Transactions []PaymentTransaction `json:"transactions" gorm:"foreignKey:OrderID"`
	Receipts     []Receipt            `json:"receipts" gorm:"foreignKey:OrderID"`
}

// OrderItem represents individual items within an order
type OrderItem struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	OrderID     uuid.UUID  `json:"order_id" gorm:"type:uuid;not null"`
	ProductID   uuid.UUID  `json:"product_id" gorm:"type:uuid;not null"`
	VariationID *uuid.UUID `json:"variation_id" gorm:"type:uuid"`
	Quantity    int        `json:"quantity" gorm:"not null"`
	UnitPrice   float64    `json:"unit_price" gorm:"not null"`
	TotalPrice  float64    `json:"total_price" gorm:"not null"`

	// Snapshot data (in case product is deleted)
	ProductName string                 `json:"product_name" gorm:"not null"`
	ProductSKU  string                 `json:"product_sku"`
	ProductData map[string]interface{} `json:"product_data" gorm:"type:jsonb"` // Full product snapshot

	Metadata  map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time              `json:"created_at"`

	// Relationships
	Order     Order             `json:"order" gorm:"foreignKey:OrderID"`
	Product   Product           `json:"product" gorm:"foreignKey:ProductID"`
	Variation *ProductVariation `json:"variation" gorm:"foreignKey:VariationID"`
}

// PaymentTransaction represents payment processing transactions
type PaymentTransaction struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID string    `json:"tenant_id" gorm:"type:varchar(50);not null"`
	OrderID  uuid.UUID `json:"order_id" gorm:"type:uuid;not null"`

	// Transaction details
	TransactionID  string `json:"transaction_id" gorm:"not null"`
	PaymentGateway string `json:"payment_gateway" gorm:"not null"` // stripe, paypal, square, etc.
	PaymentMethod  string `json:"payment_method"`                  // card, bank_transfer, cash, etc.

	// Financial details
	Amount    float64 `json:"amount" gorm:"not null"`
	Currency  string  `json:"currency" gorm:"default:'USD'"`
	Fee       float64 `json:"fee" gorm:"default:0"`
	NetAmount float64 `json:"net_amount"`

	// Status and type
	Status string `json:"status" gorm:"not null"` // pending, completed, failed, cancelled, refunded
	Type   string `json:"type" gorm:"not null"`   // payment, refund, partial_refund

	// Gateway response
	GatewayResponse map[string]interface{} `json:"gateway_response" gorm:"type:jsonb"`
	FailureReason   string                 `json:"failure_reason"`

	// Reference
	ParentTransactionID *uuid.UUID `json:"parent_transaction_id" gorm:"type:uuid"`

	// Timestamps
	ProcessedAt *time.Time `json:"processed_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Relationships
	Tenant            Tenant              `json:"tenant" gorm:"foreignKey:TenantID"`
	Order             Order               `json:"order" gorm:"foreignKey:OrderID"`
	ParentTransaction *PaymentTransaction `json:"parent_transaction" gorm:"foreignKey:ParentTransactionID"`
}

// Receipt represents generated receipts
type Receipt struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID      string    `json:"tenant_id" gorm:"type:varchar(50);not null"`
	OrderID       uuid.UUID `json:"order_id" gorm:"type:uuid;not null"`
	ReceiptNumber string    `json:"receipt_number" gorm:"not null"`

	// Receipt data
	ReceiptData    map[string]interface{} `json:"receipt_data" gorm:"type:jsonb;not null"` // Complete receipt information
	ReceiptHTML    string                 `json:"receipt_html"`                            // HTML version for display
	ReceiptPDFPath string                 `json:"receipt_pdf_path"`                        // Path to PDF file

	// Email information
	EmailSent      bool       `json:"email_sent" gorm:"default:false"`
	EmailSentAt    *time.Time `json:"email_sent_at"`
	EmailRecipient string     `json:"email_recipient"`

	// Print information
	Printed       bool       `json:"printed" gorm:"default:false"`
	PrintCount    int        `json:"print_count" gorm:"default:0"`
	LastPrintedAt *time.Time `json:"last_printed_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
	Order  Order  `json:"order" gorm:"foreignKey:OrderID"`
}

// SalesReport represents cached sales reports
type SalesReport struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID    string    `json:"tenant_id" gorm:"type:varchar(50);not null"`
	ReportType  string    `json:"report_type" gorm:"not null"` // daily, weekly, monthly, yearly, custom
	PeriodStart time.Time `json:"period_start" gorm:"not null"`
	PeriodEnd   time.Time `json:"period_end" gorm:"not null"`

	// Sales metrics
	TotalSales        float64 `json:"total_sales" gorm:"default:0"`
	TotalOrders       int     `json:"total_orders" gorm:"default:0"`
	TotalItems        int     `json:"total_items" gorm:"default:0"`
	AverageOrderValue float64 `json:"average_order_value" gorm:"default:0"`

	// Product metrics
	TopProducts           []map[string]interface{} `json:"top_products" gorm:"type:jsonb;default:'[]'"`
	CategoriesPerformance map[string]interface{}   `json:"categories_performance" gorm:"type:jsonb;default:'{}'"`

	// Customer metrics
	NewCustomers       int `json:"new_customers" gorm:"default:0"`
	ReturningCustomers int `json:"returning_customers" gorm:"default:0"`

	// Payment metrics
	PaymentMethods map[string]interface{} `json:"payment_methods" gorm:"type:jsonb;default:'{}'"`

	// Additional metrics
	RefundAmount float64 `json:"refund_amount" gorm:"default:0"`
	RefundOrders int     `json:"refund_orders" gorm:"default:0"`

	ReportData  map[string]interface{} `json:"report_data" gorm:"type:jsonb;default:'{}'"` // Full report data
	GeneratedAt time.Time              `json:"generated_at"`
	ExpiresAt   *time.Time             `json:"expires_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// Discount represents discount codes and promotional offers
type Discount struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID string    `json:"tenant_id" gorm:"type:varchar(50);not null"`

	// Basic info
	Code        string `json:"code" gorm:"not null"`
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`

	// Discount type and value
	DiscountType  string  `json:"discount_type" gorm:"not null"` // percentage, fixed_amount, free_shipping
	DiscountValue float64 `json:"discount_value" gorm:"not null"`

	// Usage restrictions
	MinimumAmount         float64 `json:"minimum_amount"`
	MaximumAmount         float64 `json:"maximum_amount"`
	UsageLimit            int     `json:"usage_limit"`
	UsageLimitPerCustomer int     `json:"usage_limit_per_customer"`
	UsedCount             int     `json:"used_count" gorm:"default:0"`

	// Product restrictions
	ApplicableProducts   []string `json:"applicable_products" gorm:"type:text[]"` // Array of product IDs
	ExcludedProducts     []string `json:"excluded_products" gorm:"type:text[]"`
	ApplicableCategories []string `json:"applicable_categories" gorm:"type:text[]"`
	ExcludedCategories   []string `json:"excluded_categories" gorm:"type:text[]"`

	// Date restrictions
	StartsAt  *time.Time `json:"starts_at"`
	ExpiresAt *time.Time `json:"expires_at"`

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	Metadata  map[string]interface{} `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
}

// Wishlist represents customer wishlists
type Wishlist struct {
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   string    `json:"tenant_id" gorm:"type:varchar(50);not null"`
	UserID     uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Name       string    `json:"name" gorm:"default:'My Wishlist'"`
	IsPublic   bool      `json:"is_public" gorm:"default:false"`
	ShareToken string    `json:"share_token"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Tenant Tenant         `json:"tenant" gorm:"foreignKey:TenantID"`
	User   User           `json:"user" gorm:"foreignKey:UserID"`
	Items  []WishlistItem `json:"items" gorm:"foreignKey:WishlistID"`
}

// WishlistItem represents individual items in wishlists
type WishlistItem struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	WishlistID  uuid.UUID  `json:"wishlist_id" gorm:"type:uuid;not null"`
	ProductID   uuid.UUID  `json:"product_id" gorm:"type:uuid;not null"`
	VariationID *uuid.UUID `json:"variation_id" gorm:"type:uuid"`
	AddedAt     time.Time  `json:"added_at"`

	// Relationships
	Wishlist  Wishlist          `json:"wishlist" gorm:"foreignKey:WishlistID"`
	Product   Product           `json:"product" gorm:"foreignKey:ProductID"`
	Variation *ProductVariation `json:"variation" gorm:"foreignKey:VariationID"`
}

// POS status constants
const (
	// Product status
	ProductStatusDraft     = "draft"
	ProductStatusPublished = "published"
	ProductStatusPrivate   = "private"
	ProductStatusPending   = "pending"

	// Product types
	ProductTypeSimple   = "simple"
	ProductTypeVariable = "variable"
	ProductTypeGrouped  = "grouped"
	ProductTypeExternal = "external"

	// Stock status
	StockStatusInStock     = "instock"
	StockStatusOutOfStock  = "outofstock"
	StockStatusOnBackorder = "onbackorder"

	// Cart status
	CartStatusActive    = "active"
	CartStatusAbandoned = "abandoned"
	CartStatusConverted = "converted"

	// Order status
	OrderStatusPending    = "pending"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
	OrderStatusRefunded   = "refunded"

	// Payment status
	PaymentStatusPending           = "pending"
	PaymentStatusPaid              = "paid"
	PaymentStatusFailed            = "failed"
	PaymentStatusRefunded          = "refunded"
	PaymentStatusPartiallyRefunded = "partially_refunded"

	// Transaction types
	TransactionTypePayment       = "payment"
	TransactionTypeRefund        = "refund"
	TransactionTypePartialRefund = "partial_refund"

	// Transaction status
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusCancelled = "cancelled"
	TransactionStatusRefunded  = "refunded"

	// Inventory log types
	InventoryTypeIn         = "in"
	InventoryTypeOut        = "out"
	InventoryTypeAdjustment = "adjustment"
	InventoryTypeSale       = "sale"
	InventoryTypeReturn     = "return"

	// Discount types
	DiscountTypePercentage   = "percentage"
	DiscountTypeFixedAmount  = "fixed_amount"
	DiscountTypeFreeShipping = "free_shipping"

	// Report types
	ReportTypeDaily   = "daily"
	ReportTypeWeekly  = "weekly"
	ReportTypeMonthly = "monthly"
	ReportTypeYearly  = "yearly"
	ReportTypeCustom  = "custom"
)

// ============================
// POS Filter and Helper Structs
// ============================

// ProductCategoryFilter for filtering product categories
type ProductCategoryFilter struct {
	Name      *string    `json:"name,omitempty"`
	ParentID  *uuid.UUID `json:"parent_id,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order" validate:"oneof=asc desc"`
}

// ProductFilter for filtering products
type ProductFilter struct {
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	SKU         *string    `json:"sku,omitempty"`
	Name        *string    `json:"name,omitempty"`
	ProductType *string    `json:"product_type,omitempty"`
	Status      *string    `json:"status,omitempty"`
	Featured    *bool      `json:"featured,omitempty"`
	StockStatus *string    `json:"stock_status,omitempty"`
	MinPrice    *float64   `json:"min_price,omitempty"`
	MaxPrice    *float64   `json:"max_price,omitempty"`
	LowStock    *bool      `json:"low_stock,omitempty"`
	Page        int        `json:"page" validate:"min=1"`
	Limit       int        `json:"limit" validate:"min=1,max=100"`
	SortBy      string     `json:"sort_by"`
	SortOrder   string     `json:"sort_order" validate:"oneof=asc desc"`
}

// InventoryLogFilter for filtering inventory logs
type InventoryLogFilter struct {
	ProductID     *uuid.UUID `json:"product_id,omitempty"`
	VariationID   *uuid.UUID `json:"variation_id,omitempty"`
	Type          *string    `json:"type,omitempty"`
	ReferenceID   *uuid.UUID `json:"reference_id,omitempty"`
	ReferenceType *string    `json:"reference_type,omitempty"`
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Page          int        `json:"page" validate:"min=1"`
	Limit         int        `json:"limit" validate:"min=1,max=100"`
	SortBy        string     `json:"sort_by"`
	SortOrder     string     `json:"sort_order" validate:"oneof=asc desc"`
}

// CartFilter for filtering carts
type CartFilter struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	SessionID *string    `json:"session_id,omitempty"`
	Status    *string    `json:"status,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order" validate:"oneof=asc desc"`
}

// OrderFilter for filtering orders
type OrderFilter struct {
	UserID        *uuid.UUID `json:"user_id,omitempty"`
	Status        *string    `json:"status,omitempty"`
	PaymentStatus *string    `json:"payment_status,omitempty"`
	PaymentMethod *string    `json:"payment_method,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	MinTotal      *float64   `json:"min_total,omitempty"`
	MaxTotal      *float64   `json:"max_total,omitempty"`
	CustomerEmail *string    `json:"customer_email,omitempty"`
	CustomerPhone *string    `json:"customer_phone,omitempty"`
	Page          int        `json:"page" validate:"min=1"`
	Limit         int        `json:"limit" validate:"min=1,max=100"`
	SortBy        string     `json:"sort_by"`
	SortOrder     string     `json:"sort_order" validate:"oneof=asc desc"`
}

// OrderItemFilter for filtering order items
type OrderItemFilter struct {
	OrderID     *uuid.UUID `json:"order_id,omitempty"`
	ProductID   *uuid.UUID `json:"product_id,omitempty"`
	VariationID *uuid.UUID `json:"variation_id,omitempty"`
	ProductName *string    `json:"product_name,omitempty"`
	Page        int        `json:"page" validate:"min=1"`
	Limit       int        `json:"limit" validate:"min=1,max=100"`
	SortBy      string     `json:"sort_by"`
	SortOrder   string     `json:"sort_order" validate:"oneof=asc desc"`
}

// PaymentTransactionFilter for filtering payment transactions
type PaymentTransactionFilter struct {
	OrderID        *uuid.UUID `json:"order_id,omitempty"`
	PaymentGateway *string    `json:"payment_gateway,omitempty"`
	PaymentMethod  *string    `json:"payment_method,omitempty"`
	Status         *string    `json:"status,omitempty"`
	Type           *string    `json:"type,omitempty"`
	TransactionID  *string    `json:"transaction_id,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	EndDate        *time.Time `json:"end_date,omitempty"`
	MinAmount      *float64   `json:"min_amount,omitempty"`
	MaxAmount      *float64   `json:"max_amount,omitempty"`
	Page           int        `json:"page" validate:"min=1"`
	Limit          int        `json:"limit" validate:"min=1,max=100"`
	SortBy         string     `json:"sort_by"`
	SortOrder      string     `json:"sort_order" validate:"oneof=asc desc"`
}

// ReceiptFilter for filtering receipts
type ReceiptFilter struct {
	OrderID       *uuid.UUID `json:"order_id,omitempty"`
	ReceiptNumber *string    `json:"receipt_number,omitempty"`
	EmailSent     *bool      `json:"email_sent,omitempty"`
	Printed       *bool      `json:"printed,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Page          int        `json:"page" validate:"min=1"`
	Limit         int        `json:"limit" validate:"min=1,max=100"`
	SortBy        string     `json:"sort_by"`
	SortOrder     string     `json:"sort_order" validate:"oneof=asc desc"`
}

// SalesReportFilter for filtering sales reports
type SalesReportFilter struct {
	ReportType *string    `json:"report_type,omitempty"`
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Page       int        `json:"page" validate:"min=1"`
	Limit      int        `json:"limit" validate:"min=1,max=100"`
	SortBy     string     `json:"sort_by"`
	SortOrder  string     `json:"sort_order" validate:"oneof=asc desc"`
}

// DiscountFilter for filtering discounts
type DiscountFilter struct {
	Code         *string    `json:"code,omitempty"`
	Name         *string    `json:"name,omitempty"`
	DiscountType *string    `json:"discount_type,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	Page         int        `json:"page" validate:"min=1"`
	Limit        int        `json:"limit" validate:"min=1,max=100"`
	SortBy       string     `json:"sort_by"`
	SortOrder    string     `json:"sort_order" validate:"oneof=asc desc"`
}

// WishlistFilter for filtering wishlists
type WishlistFilter struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	IsPublic  *bool      `json:"is_public,omitempty"`
	Page      int        `json:"page" validate:"min=1"`
	Limit     int        `json:"limit" validate:"min=1,max=100"`
	SortBy    string     `json:"sort_by"`
	SortOrder string     `json:"sort_order" validate:"oneof=asc desc"`
}

// ============================
// POS Helper and Summary Structs
// ============================

// StockUpdate for bulk stock updates
type StockUpdate struct {
	ProductID   uuid.UUID  `json:"product_id"`
	VariationID *uuid.UUID `json:"variation_id,omitempty"`
	Quantity    int        `json:"quantity"`
	Type        string     `json:"type"` // set, increment, decrement
}

// InventorySummary for inventory summaries
type InventorySummary struct {
	ProductID        uuid.UUID  `json:"product_id"`
	VariationID      *uuid.UUID `json:"variation_id,omitempty"`
	TotalIn          int        `json:"total_in"`
	TotalOut         int        `json:"total_out"`
	CurrentStock     int        `json:"current_stock"`
	TotalAdjustments int        `json:"total_adjustments"`
	TotalSales       int        `json:"total_sales"`
	TotalReturns     int        `json:"total_returns"`
	TotalCost        float64    `json:"total_cost"`
}

// CartTotals for cart totals calculation
type CartTotals struct {
	Subtotal      float64 `json:"subtotal"`
	TaxTotal      float64 `json:"tax_total"`
	ShippingTotal float64 `json:"shipping_total"`
	DiscountTotal float64 `json:"discount_total"`
	Total         float64 `json:"total"`
}

// OrderStats for order statistics
type OrderStats struct {
	TotalOrders       int     `json:"total_orders"`
	TotalSales        float64 `json:"total_sales"`
	AverageOrderValue float64 `json:"average_order_value"`
	TotalItems        int     `json:"total_items"`
	PendingOrders     int     `json:"pending_orders"`
	ProcessingOrders  int     `json:"processing_orders"`
	CompletedOrders   int     `json:"completed_orders"`
	CancelledOrders   int     `json:"cancelled_orders"`
}

// ProductSales for top selling products
type ProductSales struct {
	ProductID    uuid.UUID `json:"product_id"`
	ProductName  string    `json:"product_name"`
	ProductSKU   string    `json:"product_sku"`
	TotalSold    int       `json:"total_sold"`
	TotalRevenue float64   `json:"total_revenue"`
	OrderCount   int       `json:"order_count"`
}

// SalesReportSummary for available reports
type SalesReportSummary struct {
	ID          uuid.UUID `json:"id"`
	ReportType  string    `json:"report_type"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	TotalSales  float64   `json:"total_sales"`
	TotalOrders int       `json:"total_orders"`
	GeneratedAt time.Time `json:"generated_at"`
}

// DiscountValidation for discount validation results
type DiscountValidation struct {
	IsValid        bool    `json:"is_valid"`
	ErrorMessage   string  `json:"error_message,omitempty"`
	DiscountAmount float64 `json:"discount_amount"`
	AppliedAmount  float64 `json:"applied_amount"`
	UsageCount     int     `json:"usage_count"`
	UsageLimit     int     `json:"usage_limit"`
}

// ============================
// POS Request/Response Structs
// ============================

// CustomerInfo represents customer information for order creation
type CustomerInfo struct {
	Email           string                 `json:"email"`
	Phone           string                 `json:"phone"`
	BillingAddress  map[string]interface{} `json:"billing_address"`
	ShippingAddress map[string]interface{} `json:"shipping_address"`
	Note            string                 `json:"note"`
}

// PaymentData represents payment information
type PaymentData struct {
	TransactionID   string                 `json:"transaction_id"`
	Gateway         string                 `json:"gateway"`
	Method          string                 `json:"method"`
	MethodTitle     string                 `json:"method_title"`
	GatewayResponse map[string]interface{} `json:"gateway_response"`
}
