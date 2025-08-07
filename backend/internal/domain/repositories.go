package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Tenant, error)
	Count(ctx context.Context) (int64, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int64, error)
}

// TenantUserRepository defines the interface for tenant-user relationship operations
type TenantUserRepository interface {
	Create(ctx context.Context, tenantUser *TenantUser) error
	GetByID(ctx context.Context, id uuid.UUID) (*TenantUser, error)
	GetByTenantAndUser(ctx context.Context, tenantID, userID uuid.UUID) (*TenantUser, error)
	Update(ctx context.Context, tenantUser *TenantUser) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*TenantUser, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*TenantUser, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// TenantDomainRepository defines the interface for tenant domain operations
type TenantDomainRepository interface {
	Create(ctx context.Context, domain *TenantDomain) error
	GetByID(ctx context.Context, id uuid.UUID) (*TenantDomain, error)
	GetByDomain(ctx context.Context, domain string) (*TenantDomain, error)
	Update(ctx context.Context, domain *TenantDomain) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*TenantDomain, error)
	SetPrimary(ctx context.Context, tenantID, domainID uuid.UUID) error
}

// APIKeyRepository defines the interface for API key operations
type APIKeyRepository interface {
	Create(ctx context.Context, apiKey *APIKey) error
	GetByID(ctx context.Context, id uuid.UUID) (*APIKey, error)
	GetByKey(ctx context.Context, key string) (*APIKey, error)
	Update(ctx context.Context, apiKey *APIKey) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*APIKey, error)
	UpdateLastUsed(ctx context.Context, id uuid.UUID) error
}

// AuditLogRepository defines the interface for audit log operations
type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*AuditLog, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*AuditLog, error)
	ListByResource(ctx context.Context, tenantID uuid.UUID, resource string, limit, offset int) ([]*AuditLog, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	DeleteOldLogs(ctx context.Context, beforeDate time.Time) error
}

// RoleRepository defines the interface for role operations
type RoleRepository interface {
	Create(ctx context.Context, role *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string, tenantID *uuid.UUID) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, tenantID *uuid.UUID, limit, offset int) ([]*Role, error)
	ListSystemRoles(ctx context.Context) ([]*Role, error)
	ListTenantRoles(ctx context.Context, tenantID uuid.UUID) ([]*Role, error)
	Count(ctx context.Context, tenantID *uuid.UUID) (int64, error)
}

// PermissionRepository defines the interface for permission operations
type PermissionRepository interface {
	Create(ctx context.Context, permission *Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)
	Update(ctx context.Context, permission *Permission) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Permission, error)
	ListByResource(ctx context.Context, resource string) ([]*Permission, error)
	Count(ctx context.Context) (int64, error)
}

// UserRoleRepository defines the interface for user-role assignment operations
type UserRoleRepository interface {
	Create(ctx context.Context, userRole *UserRole) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserRole, error)
	GetByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]*UserRole, error)
	Update(ctx context.Context, userRole *UserRole) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserAndRole(ctx context.Context, userID, roleID, tenantID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	ListByRole(ctx context.Context, roleID uuid.UUID, limit, offset int) ([]*UserRole, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UserRole, error)
	CountByRole(ctx context.Context, roleID uuid.UUID) (int64, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
}
