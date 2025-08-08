package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id uuid.UUID) (*Tenant, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*Tenant, error)
	Count(ctx context.Context) (int64, error)

	// Advanced tenant operations
	GetByDomain(ctx context.Context, domain string) (*Tenant, error)
	SearchTenants(ctx context.Context, query string, limit, offset int) ([]*Tenant, error)
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*Tenant, error)
	ListByPlan(ctx context.Context, plan string, limit, offset int) ([]*Tenant, error)

	// Onboarding operations
	UpdateOnboardingStatus(ctx context.Context, tenantID uuid.UUID, status string, step int) error
	UpdateOnboardingData(ctx context.Context, tenantID uuid.UUID, data map[string]interface{}) error
	GetTenantsByOnboardingStatus(ctx context.Context, status string, limit, offset int) ([]*Tenant, error)

	// Configuration management
	UpdateSettings(ctx context.Context, tenantID uuid.UUID, settings *TenantSettings) error
	UpdateConfiguration(ctx context.Context, tenantID uuid.UUID, config *TenantConfiguration) error
	UpdateBranding(ctx context.Context, tenantID uuid.UUID, branding *TenantBranding) error
	UpdateBilling(ctx context.Context, tenantID uuid.UUID, billing *TenantBilling) error

	// Tenant lifecycle
	ActivateTenant(ctx context.Context, tenantID uuid.UUID) error
	SuspendTenant(ctx context.Context, tenantID uuid.UUID) error
	CancelTenant(ctx context.Context, tenantID uuid.UUID) error

	// Subdomain management
	IsSubdomainAvailable(ctx context.Context, subdomain string) (bool, error)
	ReserveSubdomain(ctx context.Context, subdomain string, tenantID uuid.UUID) error
	ReleaseSubdomain(ctx context.Context, subdomain string) error

	// Metrics and analytics
	GetTenantMetrics(ctx context.Context, tenantID uuid.UUID, metricType string, from, to time.Time) ([]*TenantUsageMetrics, error)
	RecordUsageMetric(ctx context.Context, metric *TenantUsageMetrics) error
	GetTenantStats(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error)
}

// TenantOnboardingRepository defines operations for tenant onboarding tracking
type TenantOnboardingRepository interface {
	Create(ctx context.Context, log *TenantOnboardingLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*TenantOnboardingLog, error)
	Update(ctx context.Context, log *TenantOnboardingLog) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*TenantOnboardingLog, error)
	GetByTenantAndStep(ctx context.Context, tenantID uuid.UUID, step int) (*TenantOnboardingLog, error)
	UpdateStepStatus(ctx context.Context, tenantID uuid.UUID, step int, status string, data map[string]interface{}) error
	GetCurrentStep(ctx context.Context, tenantID uuid.UUID) (*TenantOnboardingLog, error)
	CompleteStep(ctx context.Context, tenantID uuid.UUID, step int, data map[string]interface{}) error
}

// TenantInvitationRepository defines operations for tenant invitations
type TenantInvitationRepository interface {
	Create(ctx context.Context, invitation *TenantInvitation) error
	GetByID(ctx context.Context, id uuid.UUID) (*TenantInvitation, error)
	GetByToken(ctx context.Context, token string) (*TenantInvitation, error)
	Update(ctx context.Context, invitation *TenantInvitation) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*TenantInvitation, error)
	ListByEmail(ctx context.Context, email string) ([]*TenantInvitation, error)
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error
	RevokeInvitation(ctx context.Context, id uuid.UUID) error
	CleanupExpiredInvitations(ctx context.Context) error
}

// TenantUsageRepository defines operations for tenant usage tracking
type TenantUsageRepository interface {
	Create(ctx context.Context, metric *TenantUsageMetrics) error
	GetByID(ctx context.Context, id uuid.UUID) (*TenantUsageMetrics, error)
	Update(ctx context.Context, metric *TenantUsageMetrics) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*TenantUsageMetrics, error)
	GetMetricsByType(ctx context.Context, tenantID uuid.UUID, metricType string, from, to time.Time) ([]*TenantUsageMetrics, error)
	GetLatestMetric(ctx context.Context, tenantID uuid.UUID, metricType string) (*TenantUsageMetrics, error)
	RecordMetric(ctx context.Context, tenantID uuid.UUID, metricType string, value float64, unit string, additionalData map[string]interface{}) error
	AggregateMetrics(ctx context.Context, tenantID uuid.UUID, metricType string, from, to time.Time, aggregationType string) (float64, error)
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByKeycloakUserID(ctx context.Context, keycloakUserID string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error

	// List operations
	List(ctx context.Context, limit, offset int) ([]*User, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*User, error)
	ListByRole(ctx context.Context, role string, limit, offset int) ([]*User, error)
	ListSystemAdmins(ctx context.Context, limit, offset int) ([]*User, error)
	ListTenantAdmins(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*User, error)

	// Search operations
	Search(ctx context.Context, query string, limit, offset int) ([]*User, error)
	SearchByTenant(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*User, error)

	// Count operations
	Count(ctx context.Context) (int64, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	CountByRole(ctx context.Context, role string) (int64, error)

	// Profile operations
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile map[string]interface{}) error
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
	UpdatePreferences(ctx context.Context, userID uuid.UUID, preferences map[string]interface{}) error

	// Session operations
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	IncrementLoginCount(ctx context.Context, userID uuid.UUID) error

	// Status operations
	UpdateStatus(ctx context.Context, userID uuid.UUID, status string) error
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	SuspendUser(ctx context.Context, userID uuid.UUID) error
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

// DomainRegistrationRepository defines the interface for domain registration operations
type DomainRegistrationRepository interface {
	Create(ctx context.Context, registration *DomainRegistration) error
	GetByID(ctx context.Context, id uuid.UUID) (*DomainRegistration, error)
	GetByDomainID(ctx context.Context, domainID uuid.UUID) (*DomainRegistration, error)
	Update(ctx context.Context, registration *DomainRegistration) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*DomainRegistration, error)
	ListExpiring(ctx context.Context, days int) ([]*DomainRegistration, error)
	ListByProvider(ctx context.Context, provider string) ([]*DomainRegistration, error)
	GetRegistrationStats(ctx context.Context) (map[string]interface{}, error)
}

// DomainRegistrationEventRepository defines the interface for domain registration event operations
type DomainRegistrationEventRepository interface {
	Create(ctx context.Context, event *DomainRegistrationEvent) error
	GetByID(ctx context.Context, id uuid.UUID) (*DomainRegistrationEvent, error)
	ListByRegistrationID(ctx context.Context, registrationID uuid.UUID) ([]*DomainRegistrationEvent, error)
	ListByEventType(ctx context.Context, eventType string, limit, offset int) ([]*DomainRegistrationEvent, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// DNSRecordRepository defines the interface for DNS record operations
type DNSRecordRepository interface {
	Create(ctx context.Context, record *DNSRecord) error
	GetByID(ctx context.Context, id uuid.UUID) (*DNSRecord, error)
	Update(ctx context.Context, record *DNSRecord) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByDomainID(ctx context.Context, domainID uuid.UUID) ([]*DNSRecord, error)
	ListByType(ctx context.Context, domainID uuid.UUID, recordType string) ([]*DNSRecord, error)
	ListByPurpose(ctx context.Context, domainID uuid.UUID, purpose string) ([]*DNSRecord, error)
	FindByNameAndType(ctx context.Context, domainID uuid.UUID, name, recordType string) (*DNSRecord, error)
	ListManagedRecords(ctx context.Context, domainID uuid.UUID) ([]*DNSRecord, error)
	BulkCreate(ctx context.Context, records []*DNSRecord) error
	BulkUpdate(ctx context.Context, records []*DNSRecord) error
	BulkDelete(ctx context.Context, recordIDs []uuid.UUID) error
}

// DomainOwnershipVerificationRepository defines the interface for domain ownership verification operations
type DomainOwnershipVerificationRepository interface {
	Create(ctx context.Context, verification *DomainOwnershipVerification) error
	GetByID(ctx context.Context, id uuid.UUID) (*DomainOwnershipVerification, error)
	GetByDomainID(ctx context.Context, domainID uuid.UUID) ([]*DomainOwnershipVerification, error)
	GetActiveByDomainID(ctx context.Context, domainID uuid.UUID) (*DomainOwnershipVerification, error)
	Update(ctx context.Context, verification *DomainOwnershipVerification) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*DomainOwnershipVerification, error)
	ListPendingVerifications(ctx context.Context) ([]*DomainOwnershipVerification, error)
	ListReadyForRetry(ctx context.Context) ([]*DomainOwnershipVerification, error)
	ListExpiring(ctx context.Context, hours int) ([]*DomainOwnershipVerification, error)
}

// SSLCertificateRequestRepository defines the interface for SSL certificate request operations
type SSLCertificateRequestRepository interface {
	Create(ctx context.Context, request *SSLCertificateRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*SSLCertificateRequest, error)
	GetByDomainID(ctx context.Context, domainID uuid.UUID) ([]*SSLCertificateRequest, error)
	GetActiveByDomainID(ctx context.Context, domainID uuid.UUID) (*SSLCertificateRequest, error)
	Update(ctx context.Context, request *SSLCertificateRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByStatus(ctx context.Context, status string, limit, offset int) ([]*SSLCertificateRequest, error)
	ListPendingRequests(ctx context.Context) ([]*SSLCertificateRequest, error)
	ListExpiredRequests(ctx context.Context) ([]*SSLCertificateRequest, error)
	ListFailedRequests(ctx context.Context, maxAttempts int) ([]*SSLCertificateRequest, error)
}

// DomainHealthCheckRepository defines the interface for domain health check operations
type DomainHealthCheckRepository interface {
	Create(ctx context.Context, healthCheck *DomainHealthCheck) error
	GetByID(ctx context.Context, id uuid.UUID) (*DomainHealthCheck, error)
	GetByDomainID(ctx context.Context, domainID uuid.UUID) ([]*DomainHealthCheck, error)
	Update(ctx context.Context, healthCheck *DomainHealthCheck) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByStatus(ctx context.Context, status string) ([]*DomainHealthCheck, error)
	ListByCheckType(ctx context.Context, checkType string) ([]*DomainHealthCheck, error)
	ListReadyForCheck(ctx context.Context) ([]*DomainHealthCheck, error)
	GetHealthStats(ctx context.Context, domainID uuid.UUID, hours int) (map[string]interface{}, error)
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

// UserSessionRepository defines the interface for user session operations
type UserSessionRepository interface {
	Create(ctx context.Context, session *UserSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserSession, error)
	GetByToken(ctx context.Context, token string) (*UserSession, error)
	Update(ctx context.Context, session *UserSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*UserSession, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*UserSession, error)
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	UpdateLastAccessed(ctx context.Context, token string) error
}

// UserPreferenceRepository defines the interface for user preference operations
type UserPreferenceRepository interface {
	Create(ctx context.Context, preference *UserPreference) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserPreference, error)
	GetByUserAndKey(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string) (*UserPreference, error)
	Update(ctx context.Context, preference *UserPreference) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpsertPreference(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string, value interface{}) error
	ListByUser(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*UserPreference, error)
	ListByCategory(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category string) ([]*UserPreference, error)
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	GetUserPreferences(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) (map[string]interface{}, error)
}

// FileStorageConfigRepository defines the interface for file storage configuration operations
type FileStorageConfigRepository interface {
	Create(ctx context.Context, config *FileStorageConfig) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileStorageConfig, error)
	GetByTenantAndType(ctx context.Context, tenantID uuid.UUID, storageType string) (*FileStorageConfig, error)
	GetActiveByTenant(ctx context.Context, tenantID uuid.UUID) (*FileStorageConfig, error)
	Update(ctx context.Context, config *FileStorageConfig) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*FileStorageConfig, error)
	SetActive(ctx context.Context, id uuid.UUID, tenantID uuid.UUID) error
}

// FileRepository defines the interface for file operations
type FileRepository interface {
	Create(ctx context.Context, file *File) error
	GetByID(ctx context.Context, id uuid.UUID) (*File, error)
	GetByPath(ctx context.Context, path string) (*File, error)
	GetByChecksum(ctx context.Context, checksum string) (*File, error)
	Update(ctx context.Context, file *File) error
	Delete(ctx context.Context, id uuid.UUID) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*File, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*File, error)
	ListByCategory(ctx context.Context, category string, limit, offset int) ([]*File, error)
	ListByUserAndCategory(ctx context.Context, userID uuid.UUID, category string, limit, offset int) ([]*File, error)
	ListByTags(ctx context.Context, tags []string, limit, offset int) ([]*File, error)
	ListPublic(ctx context.Context, limit, offset int) ([]*File, error)
	ListPendingProcessing(ctx context.Context, limit, offset int) ([]*File, error)
	ListPendingVirusScan(ctx context.Context, limit, offset int) ([]*File, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	GetTotalSizeByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	GetTotalSizeByTenant(ctx context.Context, tenantID uuid.UUID) (int64, error)
	DeleteExpired(ctx context.Context) error
	Search(ctx context.Context, query string, tenantID *uuid.UUID, limit, offset int) ([]*File, error)
	UpdateProcessingStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateVirusScanStatus(ctx context.Context, id uuid.UUID, status string, result map[string]interface{}) error
}

// FileVersionRepository defines the interface for file version operations
type FileVersionRepository interface {
	Create(ctx context.Context, version *FileVersion) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileVersion, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*FileVersion, error)
	GetLatestByFileID(ctx context.Context, fileID uuid.UUID) (*FileVersion, error)
	GetByFileAndVersion(ctx context.Context, fileID uuid.UUID, versionNumber int) (*FileVersion, error)
	Update(ctx context.Context, version *FileVersion) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) error
	CountByFileID(ctx context.Context, fileID uuid.UUID) (int64, error)
}

// FileShareRepository defines the interface for file sharing operations
type FileShareRepository interface {
	Create(ctx context.Context, share *FileShare) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileShare, error)
	GetByAccessToken(ctx context.Context, token string) (*FileShare, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID) ([]*FileShare, error)
	GetBySharedBy(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FileShare, error)
	GetBySharedWith(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FileShare, error)
	Update(ctx context.Context, share *FileShare) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByFileID(ctx context.Context, fileID uuid.UUID) error
	IncrementDownloadCount(ctx context.Context, id uuid.UUID) error
	DeactivateExpired(ctx context.Context) error
	ValidateAccess(ctx context.Context, token string, password *string) (*FileShare, error)
}

// FileUploadSessionRepository defines the interface for upload session operations
type FileUploadSessionRepository interface {
	Create(ctx context.Context, session *FileUploadSession) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileUploadSession, error)
	GetByToken(ctx context.Context, token string) (*FileUploadSession, error)
	GetByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FileUploadSession, error)
	Update(ctx context.Context, session *FileUploadSession) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateProgress(ctx context.Context, id uuid.UUID, uploadedChunks int) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	DeleteExpired(ctx context.Context) error
	CleanupCompleted(ctx context.Context, olderThan time.Time) error
}

// FileProcessingJobRepository interface for file processing job operations
type FileProcessingJobRepository interface {
	Create(ctx context.Context, job *FileProcessingJob) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileProcessingJob, error)
	Update(ctx context.Context, job *FileProcessingJob) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByFileID(ctx context.Context, fileID uuid.UUID) ([]*FileProcessingJob, error)
	FindByStatus(ctx context.Context, status string) ([]*FileProcessingJob, error)
	FindByJobType(ctx context.Context, jobType string) ([]*FileProcessingJob, error)
	FindByDateRange(ctx context.Context, start, end time.Time) ([]*FileProcessingJob, error)
	FindFailedJobs(ctx context.Context) ([]*FileProcessingJob, error)
	CleanupOldJobs(ctx context.Context, olderThan time.Time) error
} // FileAccessLogRepository defines the interface for file access logging operations
type FileAccessLogRepository interface {
	Create(ctx context.Context, log *FileAccessLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*FileAccessLog, error)
	GetByFileID(ctx context.Context, fileID uuid.UUID, limit, offset int) ([]*FileAccessLog, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*FileAccessLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*FileAccessLog, error)
	GetActivitySummary(ctx context.Context, fileID uuid.UUID, since time.Time) (map[string]int, error)
	CleanupOldLogs(ctx context.Context, olderThan time.Time) error
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
	GetByUserTenantRole(ctx context.Context, userID, tenantID, roleID uuid.UUID) (*UserRole, error)
	Update(ctx context.Context, userRole *UserRole) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUserAndRole(ctx context.Context, userID, roleID, tenantID uuid.UUID) error
	DeleteByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) error
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

// ===========================
// Reporting & Analytics Repositories
// ===========================

// AnalyticsReportRepository defines the interface for analytics report operations
type AnalyticsReportRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, report *AnalyticsReport) error
	GetByID(ctx context.Context, id uuid.UUID) (*AnalyticsReport, error)
	Update(ctx context.Context, report *AnalyticsReport) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTenantID(ctx context.Context, tenantID string, filter *ReportFilter) ([]*AnalyticsReport, int64, error)

	// Report lifecycle operations
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, errorMessage string) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress int, stats map[string]interface{}) error
	UpdateFileInfo(ctx context.Context, id uuid.UUID, filePath, fileURL string, fileSize int64) error
	MarkCompleted(ctx context.Context, id uuid.UUID, filePath, fileURL string, fileSize int64) error
	MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) error

	// Download tracking
	IncrementDownloadCount(ctx context.Context, id uuid.UUID) error
	UpdateLastDownload(ctx context.Context, id uuid.UUID) error

	// Scheduled reports
	GetScheduledReports(ctx context.Context, before time.Time) ([]*AnalyticsReport, error)
	UpdateNextRun(ctx context.Context, id uuid.UUID, nextRunAt time.Time) error

	// Cleanup operations
	DeleteExpiredReports(ctx context.Context, before time.Time) (int64, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// Search and filtering
	SearchReports(ctx context.Context, tenantID, query string, limit, offset int) ([]*AnalyticsReport, int64, error)
	GetReportsByType(ctx context.Context, tenantID, reportType string, limit, offset int) ([]*AnalyticsReport, error)
	GetRecentReports(ctx context.Context, tenantID string, limit int) ([]*AnalyticsReport, error)
}

// UserActivityMetricsRepository defines the interface for user activity tracking
type UserActivityMetricsRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, metrics *UserActivityMetrics) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserActivityMetrics, error)
	Update(ctx context.Context, metrics *UserActivityMetrics) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Activity tracking
	RecordUserActivity(ctx context.Context, tenantID string, userID uuid.UUID, sessionID string, activityData map[string]interface{}) error
	UpdateSessionDuration(ctx context.Context, tenantID string, userID uuid.UUID, sessionID string, duration int) error
	IncrementPageViews(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time) error
	IncrementActions(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time, count int) error
	IncrementLoginCount(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time) error
	IncrementErrorCount(ctx context.Context, tenantID string, userID uuid.UUID, date time.Time, count int) error

	// Query operations
	GetByTenantAndUser(ctx context.Context, tenantID string, userID uuid.UUID, filter *ActivityMetricsFilter) ([]*UserActivityMetrics, int64, error)
	GetByTenantAndDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, filter *ActivityMetricsFilter) ([]*UserActivityMetrics, int64, error)
	GetDailyMetrics(ctx context.Context, tenantID string, date time.Time, filter *ActivityMetricsFilter) ([]*UserActivityMetrics, error)
	GetUserSummary(ctx context.Context, tenantID string, userID uuid.UUID, days int) (map[string]interface{}, error)

	// Aggregation operations
	GetActiveUsersCount(ctx context.Context, tenantID string, date time.Time) (int64, error)
	GetTopUsers(ctx context.Context, tenantID string, startDate, endDate time.Time, limit int) ([]*UserActivityMetrics, error)
	GetActivityTrends(ctx context.Context, tenantID string, startDate, endDate time.Time, groupBy string) ([]map[string]interface{}, error)
	GetDeviceStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetGeographicStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error)

	// Cleanup operations
	DeleteOldMetrics(ctx context.Context, before time.Time) (int64, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// SystemUsageMetricsRepository defines the interface for system-level metrics
type SystemUsageMetricsRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, metrics *SystemUsageMetrics) error
	GetByID(ctx context.Context, id uuid.UUID) (*SystemUsageMetrics, error)
	Update(ctx context.Context, metrics *SystemUsageMetrics) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Metrics recording
	RecordAPIUsage(ctx context.Context, tenantID string, date time.Time, hour int, calls, successes, errors int, avgResponseTime int) error
	RecordStorageUsage(ctx context.Context, tenantID string, date time.Time, hour int, storageUsed, bandwidthUsed int64, filesUp, filesDown int) error
	RecordDatabaseUsage(ctx context.Context, tenantID string, date time.Time, hour int, queries int, totalQueryTime int) error
	RecordUserMetrics(ctx context.Context, tenantID string, date time.Time, hour int, activeUsers, newUsers, sessions, logins, loginSuccesses int) error
	RecordPOSMetrics(ctx context.Context, tenantID string, date time.Time, hour int, orders int, revenue float64, payments int) error
	RecordCustomMetric(ctx context.Context, tenantID string, date time.Time, hour int, metricType, metricName string, value float64, unit string, metadata map[string]interface{}) error

	// Query operations
	GetByTenantAndDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, filter *SystemMetricsFilter) ([]*SystemUsageMetrics, int64, error)
	GetHourlyMetrics(ctx context.Context, tenantID string, date time.Time, filter *SystemMetricsFilter) ([]*SystemUsageMetrics, error)
	GetDailyAggregates(ctx context.Context, tenantID string, startDate, endDate time.Time, metricTypes []string) ([]map[string]interface{}, error)
	GetSystemOverview(ctx context.Context, tenantID string, days int) (map[string]interface{}, error)

	// Aggregation operations
	GetAPIUsageStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetStorageStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetUserActivityStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetPerformanceStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (map[string]interface{}, error)
	GetTopMetrics(ctx context.Context, tenantID string, metricType string, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error)

	// Cross-tenant analytics (for system admins)
	GetSystemWideStats(ctx context.Context, startDate, endDate time.Time) (map[string]interface{}, error)
	GetTenantRankings(ctx context.Context, metricName string, startDate, endDate time.Time, limit int) ([]map[string]interface{}, error)

	// Cleanup operations
	DeleteOldMetrics(ctx context.Context, before time.Time) (int64, error)
	DeleteByTenantID(ctx context.Context, tenantID string) error
}

// ReportExportRepository defines the interface for report export operations
type ReportExportRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, export *ReportExport) error
	GetByID(ctx context.Context, id uuid.UUID) (*ReportExport, error)
	Update(ctx context.Context, export *ReportExport) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTenantID(ctx context.Context, tenantID string, limit, offset int) ([]*ReportExport, int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ReportExport, error)

	// Export lifecycle operations
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, progress int, errorMessage string) error
	UpdateProgress(ctx context.Context, id uuid.UUID, progress int, rowsProcessed, totalRows int) error
	UpdateFileInfo(ctx context.Context, id uuid.UUID, filePath, fileURL string, fileSize int64) error
	MarkStarted(ctx context.Context, id uuid.UUID) error
	MarkCompleted(ctx context.Context, id uuid.UUID, filePath, fileURL string, fileSize int64) error
	MarkFailed(ctx context.Context, id uuid.UUID, errorMessage string) error

	// Download tracking
	IncrementDownloadCount(ctx context.Context, id uuid.UUID) error
	UpdateLastDownload(ctx context.Context, id uuid.UUID) error

	// Queue operations
	GetPendingExports(ctx context.Context, limit int) ([]*ReportExport, error)
	GetProcessingExports(ctx context.Context) ([]*ReportExport, error)

	// Cleanup operations
	DeleteExpiredExports(ctx context.Context, before time.Time) (int64, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// Statistics
	GetExportStats(ctx context.Context, tenantID string, days int) (map[string]interface{}, error)
}

// ReportScheduleRepository defines the interface for scheduled report operations
type ReportScheduleRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, schedule *ReportSchedule) error
	GetByID(ctx context.Context, id uuid.UUID) (*ReportSchedule, error)
	Update(ctx context.Context, schedule *ReportSchedule) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByTenantID(ctx context.Context, tenantID string, limit, offset int) ([]*ReportSchedule, int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ReportSchedule, error)

	// Schedule management
	GetDueSchedules(ctx context.Context, before time.Time) ([]*ReportSchedule, error)
	UpdateLastRun(ctx context.Context, id uuid.UUID, lastRunAt time.Time, nextRunAt *time.Time) error
	UpdateNextRun(ctx context.Context, id uuid.UUID, nextRunAt time.Time) error
	IncrementRunCount(ctx context.Context, id uuid.UUID) error
	IncrementErrorCount(ctx context.Context, id uuid.UUID, errorMessage string) error

	// Status management
	Activate(ctx context.Context, id uuid.UUID) error
	Deactivate(ctx context.Context, id uuid.UUID) error
	GetActiveSchedules(ctx context.Context, tenantID string) ([]*ReportSchedule, error)

	// Cleanup operations
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// Statistics
	GetScheduleStats(ctx context.Context, tenantID string) (map[string]interface{}, error)
}

// OrderRepository interface for order operations (POS module)
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	Update(ctx context.Context, order *Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	GetByOrderNumber(ctx context.Context, tenantID string, orderNumber string) (*Order, error)
	GetByTenantID(ctx context.Context, tenantID string, filter *OrderFilter) ([]*Order, int64, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, filter *OrderFilter) ([]*Order, int64, error)
	GetByStatus(ctx context.Context, tenantID string, status string, filter *OrderFilter) ([]*Order, int64, error)
	GetByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]*Order, error)
	UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error
	UpdatePaymentStatus(ctx context.Context, orderID uuid.UUID, paymentStatus string) error
	GetRecentOrders(ctx context.Context, tenantID string, limit int) ([]*Order, error)
	GetOrderStats(ctx context.Context, tenantID string, startDate, endDate time.Time) (*OrderStats, error)
	GenerateOrderNumber(ctx context.Context, tenantID string) (string, error)
}
