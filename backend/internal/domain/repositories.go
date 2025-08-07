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
	GetByTenantID(ctx context.Context, tenantID string) ([]*TenantDomain, error)
	ListActive(ctx context.Context) ([]*TenantDomain, error)
	ListExpiringSSL(ctx context.Context, days int) ([]*TenantDomain, error)
}

// DomainValidationLogRepository defines the interface for domain validation log operations
type DomainValidationLogRepository interface {
	Create(ctx context.Context, log *DomainValidationLog) error
	GetByID(ctx context.Context, id uint) (*DomainValidationLog, error)
	GetByDomainID(ctx context.Context, domainID uuid.UUID) ([]*DomainValidationLog, error)
	Update(ctx context.Context, log *DomainValidationLog) error
	Delete(ctx context.Context, id uint) error
	ListPendingRetries(ctx context.Context) ([]*DomainValidationLog, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*DomainValidationLog, error)
}

// SSLCertificateRepository defines the interface for SSL certificate operations
type SSLCertificateRepository interface {
	Create(ctx context.Context, cert *SSLCertificate) error
	GetByID(ctx context.Context, id uint) (*SSLCertificate, error)
	GetByDomainID(ctx context.Context, domainID uuid.UUID) (*SSLCertificate, error)
	Update(ctx context.Context, cert *SSLCertificate) error
	Delete(ctx context.Context, id uint) error
	ListExpiring(ctx context.Context, days int) ([]*SSLCertificate, error)
	ListByStatus(ctx context.Context, status string) ([]*SSLCertificate, error)
	GetActiveByDomain(ctx context.Context, domain string) (*SSLCertificate, error)
}

// DomainRoutingCacheRepository defines the interface for domain routing cache operations
type DomainRoutingCacheRepository interface {
	Upsert(ctx context.Context, cache *DomainRoutingCache) error
	GetByDomain(ctx context.Context, domain string) (*DomainRoutingCache, error)
	GetByTenantID(ctx context.Context, tenantID string) ([]*DomainRoutingCache, error)
	Delete(ctx context.Context, domain string) error
	DeleteByDomain(ctx context.Context, domain string) error
	DeleteExpired(ctx context.Context) error
	ListAll(ctx context.Context) ([]*DomainRoutingCache, error)
	RefreshCache(ctx context.Context, domain string) error
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

// ===========================
// GraphQL Federation Repository Interfaces
// ===========================

// GraphQLSchemaRepository defines the interface for GraphQL schema operations
type GraphQLSchemaRepository interface {
	Create(ctx context.Context, schema *GraphQLSchema) error
	GetByID(ctx context.Context, id uuid.UUID) (*GraphQLSchema, error)
	GetByServiceAndVersion(ctx context.Context, serviceName, version string) (*GraphQLSchema, error)
	GetLatestByService(ctx context.Context, serviceName string) (*GraphQLSchema, error)
	GetActiveSchemas(ctx context.Context) ([]*GraphQLSchema, error)
	Update(ctx context.Context, schema *GraphQLSchema) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByService(ctx context.Context, serviceName string, limit, offset int) ([]*GraphQLSchema, error)
	ListAll(ctx context.Context, limit, offset int) ([]*GraphQLSchema, error)
	MarkAsActive(ctx context.Context, id uuid.UUID) error
	MarkAsInactive(ctx context.Context, id uuid.UUID) error
	GetByHash(ctx context.Context, hash string) (*GraphQLSchema, error)
}

// FederationServiceRepository defines the interface for federation service operations
type FederationServiceRepository interface {
	Register(ctx context.Context, service *FederationService) error
	GetByName(ctx context.Context, serviceName string) (*FederationService, error)
	GetByID(ctx context.Context, id uuid.UUID) (*FederationService, error)
	Update(ctx context.Context, service *FederationService) error
	Deregister(ctx context.Context, serviceName string) error
	ListActive(ctx context.Context) ([]*FederationService, error)
	ListAll(ctx context.Context, limit, offset int) ([]*FederationService, error)
	UpdateStatus(ctx context.Context, serviceName, status string) error
	UpdateHealthCheck(ctx context.Context, serviceName string, timestamp time.Time) error
	GetHealthyServices(ctx context.Context) ([]*FederationService, error)
	GetByStatus(ctx context.Context, status string) ([]*FederationService, error)
}

// FederationCompositionRepository defines the interface for schema composition operations
type FederationCompositionRepository interface {
	Create(ctx context.Context, composition *FederationComposition) error
	GetByID(ctx context.Context, id uuid.UUID) (*FederationComposition, error)
	GetByNameAndVersion(ctx context.Context, name, version string) (*FederationComposition, error)
	GetLatestByName(ctx context.Context, name string) (*FederationComposition, error)
	GetActive(ctx context.Context) (*FederationComposition, error)
	Update(ctx context.Context, composition *FederationComposition) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByName(ctx context.Context, name string, limit, offset int) ([]*FederationComposition, error)
	ListAll(ctx context.Context, limit, offset int) ([]*FederationComposition, error)
	MarkAsDeployed(ctx context.Context, id uuid.UUID) error
}

// GraphQLQueryMetricsRepository interface for query metrics data access
type GraphQLQueryMetricsRepository interface {
	Create(ctx context.Context, metrics *GraphQLQueryMetrics) error
	Record(ctx context.Context, metrics *GraphQLQueryMetrics) error
	GetByID(ctx context.Context, id uuid.UUID) (*GraphQLQueryMetrics, error)
	GetByQueryHash(ctx context.Context, queryHash string, limit int) ([]*GraphQLQueryMetrics, error)
	GetMetricsByService(ctx context.Context, serviceName string, from, to time.Time) ([]*GraphQLQueryMetrics, error)
	GetAverageExecutionTime(ctx context.Context, queryHash string, hours int) (time.Duration, error)
	GetQueryComplexityStats(ctx context.Context, from, to time.Time) (map[string]interface{}, error)
	GetByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*GraphQLQueryMetrics, error)
	GetSlowQueries(ctx context.Context, thresholdMs int, limit, offset int) ([]*GraphQLQueryMetrics, error)
	GetQueryStats(ctx context.Context, queryHash string, from, to time.Time) (*QueryStats, error)
	GetTenantStats(ctx context.Context, tenantID uuid.UUID, from, to time.Time) (*TenantQueryStats, error)
	GetServiceStats(ctx context.Context, serviceName string, from, to time.Time) (*ServiceQueryStats, error)
	DeleteOldMetrics(ctx context.Context, olderThan time.Time) error
}

// SchemaChangeEventRepository defines the interface for schema change event operations
type SchemaChangeEventRepository interface {
	Create(ctx context.Context, event *SchemaChangeEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*SchemaChangeEvent, error)
	GetByService(ctx context.Context, serviceName string, limit, offset int) ([]*SchemaChangeEvent, error)
	GetUnprocessed(ctx context.Context) ([]*SchemaChangeEvent, error)
	MarkAsProcessed(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context, limit, offset int) ([]*SchemaChangeEvent, error)
	GetBreakingChanges(ctx context.Context, from time.Time) ([]*SchemaChangeEvent, error)
}

// FederationGatewayConfigRepository defines the interface for gateway configuration operations
type FederationGatewayConfigRepository interface {
	Create(ctx context.Context, config *FederationGatewayConfig) error
	Update(ctx context.Context, config *FederationGatewayConfig) error
	Upsert(ctx context.Context, config *FederationGatewayConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*FederationGatewayConfig, error)
	GetByName(ctx context.Context, configName string) (*FederationGatewayConfig, error)
	GetActive(ctx context.Context) (*FederationGatewayConfig, error)
	SetActive(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*FederationGatewayConfig, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ===========================
// Helper Types for Metrics
// ===========================

// QueryStats represents aggregated query statistics
type QueryStats struct {
	QueryHash       string  `json:"query_hash"`
	QueryName       string  `json:"query_name"`
	TotalExecutions int64   `json:"total_executions"`
	AvgExecutionMs  float64 `json:"avg_execution_ms"`
	MaxExecutionMs  int     `json:"max_execution_ms"`
	MinExecutionMs  int     `json:"min_execution_ms"`
	AvgComplexity   float64 `json:"avg_complexity"`
	ErrorRate       float64 `json:"error_rate"`
	CacheHitRate    float64 `json:"cache_hit_rate"`
}

// TenantQueryStats represents aggregated tenant query statistics
type TenantQueryStats struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	TotalQueries    int64     `json:"total_queries"`
	TotalMutations  int64     `json:"total_mutations"`
	AvgExecutionMs  float64   `json:"avg_execution_ms"`
	ErrorRate       float64   `json:"error_rate"`
	CacheHitRate    float64   `json:"cache_hit_rate"`
	TopQueries      []string  `json:"top_queries"`
	ComplexityScore float64   `json:"complexity_score"`
}

// ServiceQueryStats represents aggregated service query statistics
type ServiceQueryStats struct {
	ServiceName    string  `json:"service_name"`
	TotalCalls     int64   `json:"total_calls"`
	AvgExecutionMs float64 `json:"avg_execution_ms"`
	ErrorRate      float64 `json:"error_rate"`
	AvgFieldCount  float64 `json:"avg_field_count"`
	HealthScore    float64 `json:"health_score"`
}
