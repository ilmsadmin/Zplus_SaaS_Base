package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Name      string     `json:"name" gorm:"not null"`
	Subdomain string     `json:"subdomain" gorm:"unique;not null"`
	Status    string     `json:"status" gorm:"not null;default:'active'"`
	Settings  *Settings  `json:"settings" gorm:"type:jsonb"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}

// TenantDomain represents custom domains for tenants
type TenantDomain struct {
	ID                 uuid.UUID              `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID           uuid.UUID              `json:"tenant_id" gorm:"type:uuid;not null"`
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
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email     string     `json:"email" gorm:"unique;not null"`
	Username  string     `json:"username" gorm:"unique"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Avatar    *string    `json:"avatar"`
	Status    string     `json:"status" gorm:"not null;default:'active'"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	TenantUsers []TenantUser `json:"tenant_users" gorm:"foreignKey:UserID"`
	UserRoles   []UserRole   `json:"user_roles" gorm:"foreignKey:UserID"`
}

// TenantUser represents user-tenant relationships with roles
type TenantUser struct {
	ID       uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
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
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null"`
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

// Constants for status values
const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusSuspended = "suspended"
	StatusPending   = "pending"
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
