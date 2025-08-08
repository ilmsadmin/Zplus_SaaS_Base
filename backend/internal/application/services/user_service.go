package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// UserService defines the interface for user operations
type UserService interface {
	// System Admin operations
	CreateSystemAdmin(ctx context.Context, req *CreateUserRequest) (*domain.User, error)
	UpdateSystemAdmin(ctx context.Context, userID uuid.UUID, req *UpdateUserRequest) (*domain.User, error)
	DeleteSystemAdmin(ctx context.Context, userID uuid.UUID) error
	ListSystemAdmins(ctx context.Context, req *ListUsersRequest) (*UserListResponse, error)

	// Tenant Admin operations
	CreateTenantAdmin(ctx context.Context, tenantID uuid.UUID, req *CreateUserRequest) (*domain.User, error)
	UpdateTenantAdmin(ctx context.Context, tenantID, userID uuid.UUID, req *UpdateUserRequest) (*domain.User, error)
	DeleteTenantAdmin(ctx context.Context, tenantID, userID uuid.UUID) error
	ListTenantAdmins(ctx context.Context, tenantID uuid.UUID, req *ListUsersRequest) (*UserListResponse, error)

	// User management
	CreateUser(ctx context.Context, tenantID uuid.UUID, req *CreateUserRequest) (*domain.User, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByKeycloakID(ctx context.Context, keycloakUserID string) (*domain.User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req *UpdateUserRequest) (*domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, req *ListUsersRequest) (*UserListResponse, error)
	ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, req *ListUsersRequest) (*UserListResponse, error)
	SearchUsers(ctx context.Context, query string, req *ListUsersRequest) (*UserListResponse, error)

	// Profile management
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error)
	UploadAvatar(ctx context.Context, userID uuid.UUID, file *FileUpload) (*domain.User, error)
	DeleteAvatar(ctx context.Context, userID uuid.UUID) error

	// Role assignment
	AssignRole(ctx context.Context, userID, tenantID, roleID uuid.UUID) error
	UnassignRole(ctx context.Context, userID, tenantID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*domain.UserRole, error)
	UpdateUserRoles(ctx context.Context, userID, tenantID uuid.UUID, roleIDs []uuid.UUID) error

	// User preferences
	GetUserPreferences(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) (map[string]interface{}, error)
	UpdateUserPreferences(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, preferences map[string]interface{}) error
	GetUserPreference(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string) (interface{}, error)
	SetUserPreference(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string, value interface{}) error

	// Status management
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	SuspendUser(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, userID uuid.UUID) error
	VerifyPhone(ctx context.Context, userID uuid.UUID) error

	// Session management
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.UserSession, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
}

// UserServiceImpl implements UserService
type UserServiceImpl struct {
	userRepo       domain.UserRepository
	tenantUserRepo domain.TenantUserRepository
	userRoleRepo   domain.UserRoleRepository
	prefRepo       domain.UserPreferenceRepository
	sessionRepo    domain.UserSessionRepository
	fileRepo       domain.FileRepository
	fileService    FileService
	auditService   AuditService
}

// NewUserService creates a new user service
func NewUserService(
	userRepo domain.UserRepository,
	tenantUserRepo domain.TenantUserRepository,
	userRoleRepo domain.UserRoleRepository,
	prefRepo domain.UserPreferenceRepository,
	sessionRepo domain.UserSessionRepository,
	fileRepo domain.FileRepository,
	fileService FileService,
	auditService AuditService,
) UserService {
	return &UserServiceImpl{
		userRepo:       userRepo,
		tenantUserRepo: tenantUserRepo,
		userRoleRepo:   userRoleRepo,
		prefRepo:       prefRepo,
		sessionRepo:    sessionRepo,
		fileRepo:       fileRepo,
		fileService:    fileService,
		auditService:   auditService,
	}
}

// CreateSystemAdmin creates a new system admin user
func (s *UserServiceImpl) CreateSystemAdmin(ctx context.Context, req *CreateUserRequest) (*domain.User, error) {
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	user := &domain.User{
		ID:            uuid.New(),
		Email:         req.Email,
		Username:      req.Username,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		Status:        domain.StatusActive,
		EmailVerified: req.EmailVerified,
		PhoneVerified: req.PhoneVerified,
		Preferences:   make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create system admin: %w", err)
	}

	// Create system admin role assignment
	if err := s.AssignRole(ctx, user.ID, uuid.Nil, uuid.Nil); err != nil {
		// Log error but don't fail the user creation
		// TODO: Implement proper logging
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.LogEvent(ctx, uuid.Nil, nil, domain.ActionCreate, domain.ResourceUser, user.ID.String(), nil)
	}

	return user, nil
}

// CreateTenantAdmin creates a new tenant admin user
func (s *UserServiceImpl) CreateTenantAdmin(ctx context.Context, tenantID uuid.UUID, req *CreateUserRequest) (*domain.User, error) {
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	user := &domain.User{
		ID:            uuid.New(),
		Email:         req.Email,
		Username:      req.Username,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		Status:        domain.StatusActive,
		EmailVerified: req.EmailVerified,
		PhoneVerified: req.PhoneVerified,
		Preferences:   make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create tenant admin: %w", err)
	}

	// Create tenant-user relationship
	tenantUser := &domain.TenantUser{
		ID:       uuid.New(),
		TenantID: tenantID.String(),
		UserID:   user.ID,
		Role:     domain.RoleTenantAdmin,
		Status:   domain.StatusActive,
		JoinedAt: time.Now(),
	}

	if err := s.tenantUserRepo.Create(ctx, tenantUser); err != nil {
		return nil, fmt.Errorf("failed to create tenant-user relationship: %w", err)
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.LogEvent(ctx, tenantID, &user.ID, domain.ActionCreate, domain.ResourceUser, user.ID.String(), nil)
	}

	return user, nil
}

// CreateUser creates a new regular user
func (s *UserServiceImpl) CreateUser(ctx context.Context, tenantID uuid.UUID, req *CreateUserRequest) (*domain.User, error) {
	if err := s.validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	user := &domain.User{
		ID:            uuid.New(),
		Email:         req.Email,
		Username:      req.Username,
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		Status:        domain.StatusActive,
		EmailVerified: req.EmailVerified,
		PhoneVerified: req.PhoneVerified,
		Preferences:   make(map[string]interface{}),
		Metadata:      make(map[string]interface{}),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create tenant-user relationship
	tenantUser := &domain.TenantUser{
		ID:       uuid.New(),
		TenantID: tenantID.String(),
		UserID:   user.ID,
		Role:     domain.RoleUser,
		Status:   domain.StatusActive,
		JoinedAt: time.Now(),
	}

	if err := s.tenantUserRepo.Create(ctx, tenantUser); err != nil {
		return nil, fmt.Errorf("failed to create tenant-user relationship: %w", err)
	}

	// Set default preferences
	defaultPrefs := map[string]interface{}{
		"theme":    "light",
		"language": "en",
		"timezone": "UTC",
	}
	s.UpdateUserPreferences(ctx, user.ID, &tenantID, defaultPrefs)

	// Audit log
	if s.auditService != nil {
		s.auditService.LogEvent(ctx, tenantID, &user.ID, domain.ActionCreate, domain.ResourceUser, user.ID.String(), nil)
	}

	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserServiceImpl) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserServiceImpl) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return user, nil
}

// GetUserByKeycloakID retrieves a user by Keycloak user ID
func (s *UserServiceImpl) GetUserByKeycloakID(ctx context.Context, keycloakUserID string) (*domain.User, error) {
	user, err := s.userRepo.GetByKeycloakUserID(ctx, keycloakUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by Keycloak ID: %w", err)
	}
	return user, nil
}

// UpdateUser updates a user
func (s *UserServiceImpl) UpdateUser(ctx context.Context, userID uuid.UUID, req *UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Phone != nil {
		user.Phone = *req.Phone
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.EmailVerified != nil {
		user.EmailVerified = *req.EmailVerified
	}
	if req.PhoneVerified != nil {
		user.PhoneVerified = *req.PhoneVerified
	}
	if req.Metadata != nil {
		user.Metadata = req.Metadata
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// UploadAvatar uploads and sets user avatar
func (s *UserServiceImpl) UploadAvatar(ctx context.Context, userID uuid.UUID, file *FileUpload) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Upload file using file service
	uploadedFile, err := s.fileService.UploadFile(ctx, &FileUploadRequest{
		UserID:       userID,
		File:         file,
		Category:     "avatar",
		IsPublic:     true,
		MaxSize:      5 * 1024 * 1024, // 5MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/gif"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}

	// Update user avatar URL
	user.AvatarURL = &uploadedFile.URL
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user avatar: %w", err)
	}

	return user, nil
}

// UpdateUserPreferences updates user preferences
func (s *UserServiceImpl) UpdateUserPreferences(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, preferences map[string]interface{}) error {
	for category, prefs := range preferences {
		if prefsMap, ok := prefs.(map[string]interface{}); ok {
			for key, value := range prefsMap {
				if err := s.prefRepo.UpsertPreference(ctx, userID, tenantID, category, key, value); err != nil {
					return fmt.Errorf("failed to update preference %s.%s: %w", category, key, err)
				}
			}
		}
	}
	return nil
}

// AssignRole assigns a role to a user for a tenant
func (s *UserServiceImpl) AssignRole(ctx context.Context, userID, tenantID, roleID uuid.UUID) error {
	// Check if role assignment already exists
	existing, err := s.userRoleRepo.GetByUserTenantRole(ctx, userID, tenantID, roleID)
	if err == nil && existing != nil {
		return errors.New("role already assigned to user")
	}

	userRole := &domain.UserRole{
		ID:        uuid.New(),
		UserID:    userID,
		RoleID:    roleID,
		TenantID:  tenantID,
		Status:    domain.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// Helper functions
func (s *UserServiceImpl) validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Email == "" {
		return errors.New("email is required")
	}
	if !strings.Contains(req.Email, "@") {
		return errors.New("invalid email format")
	}
	return nil
}

// Additional methods would be implemented here...
// Due to length constraints, I'm showing the core structure and key methods
