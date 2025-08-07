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
	ID         uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	TenantID   uuid.UUID `json:"tenant_id" gorm:"type:uuid;not null"`
	Domain     string    `json:"domain" gorm:"unique;not null"`
	IsVerified bool      `json:"is_verified" gorm:"default:false"`
	IsPrimary  bool      `json:"is_primary" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relationships
	Tenant Tenant `json:"tenant" gorm:"foreignKey:TenantID"`
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
