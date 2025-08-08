package services

import (
	"context"
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// ===========================
// User Service DTOs
// ===========================

// CreateUserRequest represents request to create a new user
type CreateUserRequest struct {
	Email         string `json:"email" validate:"required,email"`
	Username      string `json:"username"`
	FirstName     string `json:"first_name" validate:"required"`
	LastName      string `json:"last_name" validate:"required"`
	Phone         string `json:"phone"`
	EmailVerified bool   `json:"email_verified"`
	PhoneVerified bool   `json:"phone_verified"`
}

// UpdateUserRequest represents request to update a user
type UpdateUserRequest struct {
	FirstName     *string                `json:"first_name"`
	LastName      *string                `json:"last_name"`
	Phone         *string                `json:"phone"`
	Status        *string                `json:"status"`
	EmailVerified *bool                  `json:"email_verified"`
	PhoneVerified *bool                  `json:"phone_verified"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// UpdateProfileRequest represents request to update user profile
type UpdateProfileRequest struct {
	FirstName string                 `json:"first_name"`
	LastName  string                 `json:"last_name"`
	Phone     string                 `json:"phone"`
	Bio       string                 `json:"bio"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ListUsersRequest represents request to list users
type ListUsersRequest struct {
	Page     int    `json:"page" validate:"min=1"`
	Limit    int    `json:"limit" validate:"min=1,max=100"`
	Sort     string `json:"sort"`      // e.g., "created_at", "-created_at"
	Filter   string `json:"filter"`    // General filter
	Status   string `json:"status"`    // Filter by status
	Role     string `json:"role"`      // Filter by role
	TenantID string `json:"tenant_id"` // Filter by tenant
}

// UserListResponse represents response for user list
type UserListResponse struct {
	Users      []*UserResponse `json:"users"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	Limit      int             `json:"limit"`
	TotalPages int             `json:"total_pages"`
}

// UserResponse represents user data in responses
type UserResponse struct {
	ID            uuid.UUID              `json:"id"`
	Email         string                 `json:"email"`
	Username      string                 `json:"username"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Phone         string                 `json:"phone"`
	Avatar        *string                `json:"avatar"`
	AvatarURL     *string                `json:"avatar_url"`
	Status        string                 `json:"status"`
	EmailVerified bool                   `json:"email_verified"`
	PhoneVerified bool                   `json:"phone_verified"`
	LastLoginAt   *time.Time             `json:"last_login_at"`
	LoginCount    int                    `json:"login_count"`
	Preferences   map[string]interface{} `json:"preferences"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	Roles         []*UserRoleResponse    `json:"roles,omitempty"`
	TenantUsers   []*TenantUserResponse  `json:"tenant_users,omitempty"`
}

// UserProfile represents user profile data
type UserProfile struct {
	ID            uuid.UUID              `json:"id"`
	Email         string                 `json:"email"`
	Username      string                 `json:"username"`
	FirstName     string                 `json:"first_name"`
	LastName      string                 `json:"last_name"`
	Phone         string                 `json:"phone"`
	Avatar        *string                `json:"avatar"`
	AvatarURL     *string                `json:"avatar_url"`
	Status        string                 `json:"status"`
	EmailVerified bool                   `json:"email_verified"`
	PhoneVerified bool                   `json:"phone_verified"`
	LastLoginAt   *time.Time             `json:"last_login_at"`
	LoginCount    int                    `json:"login_count"`
	Preferences   map[string]interface{} `json:"preferences"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// UserRoleResponse represents user role data
type UserRoleResponse struct {
	ID        uuid.UUID `json:"id"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// TenantUserResponse represents tenant-user relationship data
type TenantUserResponse struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	TenantName string    `json:"tenant_name"`
	Role       string    `json:"role"`
	Status     string    `json:"status"`
	JoinedAt   time.Time `json:"joined_at"`
}

// ===========================
// File Service DTOs
// ===========================

// FileUpload represents file upload data
type FileUpload struct {
	File     multipart.File        `json:"-"`
	Header   *multipart.FileHeader `json:"-"`
	FileName string                `json:"file_name"`
	Size     int64                 `json:"size"`
	MimeType string                `json:"mime_type"`
}

// FileUploadRequest represents file upload request
type FileUploadRequest struct {
	TenantID     *uuid.UUID `json:"tenant_id"`
	UserID       uuid.UUID  `json:"user_id"`
	File         *FileUpload
	Category     string   `json:"category"`
	Tags         []string `json:"tags"`
	IsPublic     bool     `json:"is_public"`
	ExpiresAt    *time.Time
	MaxSize      int64    `json:"max_size"`
	AllowedTypes []string `json:"allowed_types"`
}

// FileResponse represents file data in responses
type FileResponse struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     *uuid.UUID             `json:"tenant_id"`
	UserID       uuid.UUID              `json:"user_id"`
	FileName     string                 `json:"file_name"`
	OriginalName string                 `json:"original_name"`
	MimeType     string                 `json:"mime_type"`
	Size         int64                  `json:"size"`
	Path         string                 `json:"path"`
	URL          string                 `json:"url"`
	StorageType  string                 `json:"storage_type"`
	Category     string                 `json:"category"`
	Tags         []string               `json:"tags"`
	IsPublic     bool                   `json:"is_public"`
	ExpiresAt    *time.Time             `json:"expires_at"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ===========================
// Preference DTOs
// ===========================

// UpdatePreferencesRequest represents request to update user preferences
type UpdatePreferencesRequest struct {
	Preferences map[string]interface{} `json:"preferences" validate:"required"`
}

// SetPreferenceRequest represents request to set a single preference
type SetPreferenceRequest struct {
	Category string      `json:"category" validate:"required"`
	Key      string      `json:"key" validate:"required"`
	Value    interface{} `json:"value" validate:"required"`
}

// PreferenceResponse represents preference data
type PreferenceResponse struct {
	ID        uuid.UUID   `json:"id"`
	UserID    uuid.UUID   `json:"user_id"`
	TenantID  *uuid.UUID  `json:"tenant_id"`
	Category  string      `json:"category"`
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

// ===========================
// Role Assignment DTOs
// ===========================

// AssignRoleRequest represents request to assign role to user
type AssignRoleRequest struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
	RoleID   uuid.UUID `json:"role_id" validate:"required"`
}

// UpdateUserRolesRequest represents request to update user roles
type UpdateUserRolesRequest struct {
	RoleIDs []uuid.UUID `json:"role_ids" validate:"required"`
}

// ===========================
// Session DTOs
// ===========================

// SessionResponse represents session data
type SessionResponse struct {
	ID             uuid.UUID  `json:"id"`
	UserID         uuid.UUID  `json:"user_id"`
	TenantID       *uuid.UUID `json:"tenant_id"`
	SessionToken   string     `json:"session_token"`
	IPAddress      string     `json:"ip_address"`
	UserAgent      string     `json:"user_agent"`
	ExpiresAt      time.Time  `json:"expires_at"`
	LastAccessedAt time.Time  `json:"last_accessed_at"`
	CreatedAt      time.Time  `json:"created_at"`
	IsActive       bool       `json:"is_active"`
}

// ===========================
// Generic Service Interfaces
// ===========================

// FileService defines file operations interface
type FileService interface {
	UploadFile(ctx context.Context, req *FileUploadRequest) (*FileResponse, error)
	DeleteFile(ctx context.Context, fileID uuid.UUID) error
	GetFile(ctx context.Context, fileID uuid.UUID) (*FileResponse, error)
	GetFileByPath(ctx context.Context, path string) (*FileResponse, error)
}

// AuditService defines audit logging interface
type AuditService interface {
	LogEvent(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, action, resource, resourceID string, details map[string]interface{}) error
}

// UserRoleRepository interface (missing from domain)
type UserRoleRepository interface {
	Create(ctx context.Context, userRole *UserRole) error
	GetByID(ctx context.Context, id uuid.UUID) (*UserRole, error)
	GetByUserTenantRole(ctx context.Context, userID, tenantID, roleID uuid.UUID) (*UserRole, error)
	Update(ctx context.Context, userRole *UserRole) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*UserRole, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*UserRole, error)
	ListByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]*UserRole, error)
	DeleteByUserAndTenant(ctx context.Context, userID, tenantID uuid.UUID) error
}

// UserRole represents user role (temporary definition, should be in domain)
type UserRole struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
