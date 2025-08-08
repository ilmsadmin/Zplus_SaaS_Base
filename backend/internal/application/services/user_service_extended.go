package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// Additional UserService methods

// UpdateSystemAdmin updates a system admin user
func (s *UserServiceImpl) UpdateSystemAdmin(ctx context.Context, userID uuid.UUID, req *UpdateUserRequest) (*domain.User, error) {
	user, err := s.UpdateUser(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// Verify user is system admin
	roles, err := s.GetUserRoles(ctx, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	isSystemAdmin := false
	for _, role := range roles {
		if role.Role.Name == domain.RoleSystemAdmin {
			isSystemAdmin = true
			break
		}
	}

	if !isSystemAdmin {
		return nil, errors.New("user is not a system admin")
	}

	return user, nil
}

// DeleteSystemAdmin deletes a system admin user
func (s *UserServiceImpl) DeleteSystemAdmin(ctx context.Context, userID uuid.UUID) error {
	// Verify user is system admin
	roles, err := s.GetUserRoles(ctx, userID, nil)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	isSystemAdmin := false
	for _, role := range roles {
		if role.Role.Name == domain.RoleSystemAdmin {
			isSystemAdmin = true
			break
		}
	}

	if !isSystemAdmin {
		return errors.New("user is not a system admin")
	}

	return s.DeleteUser(ctx, userID)
}

// ListSystemAdmins lists system admin users
func (s *UserServiceImpl) ListSystemAdmins(ctx context.Context, req *ListUsersRequest) (*UserListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit
	users, err := s.userRepo.ListSystemAdmins(ctx, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list system admins: %w", err)
	}

	total, err := s.userRepo.CountByRole(ctx, domain.RoleSystemAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to count system admins: %w", err)
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return &UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(req.Limit))),
	}, nil
}

// UpdateTenantAdmin updates a tenant admin user
func (s *UserServiceImpl) UpdateTenantAdmin(ctx context.Context, tenantID, userID uuid.UUID, req *UpdateUserRequest) (*domain.User, error) {
	// Verify user is tenant admin for this tenant
	tenantUser, err := s.tenantUserRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant user: %w", err)
	}

	if tenantUser.Role != domain.RoleTenantAdmin {
		return nil, errors.New("user is not a tenant admin")
	}

	user, err := s.UpdateUser(ctx, userID, req)
	if err != nil {
		return nil, err
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.LogEvent(ctx, tenantID, &userID, domain.ActionUpdate, domain.ResourceUser, userID.String(), nil)
	}

	return user, nil
}

// DeleteTenantAdmin deletes a tenant admin user
func (s *UserServiceImpl) DeleteTenantAdmin(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Verify user is tenant admin for this tenant
	tenantUser, err := s.tenantUserRepo.GetByTenantAndUser(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to get tenant user: %w", err)
	}

	if tenantUser.Role != domain.RoleTenantAdmin {
		return errors.New("user is not a tenant admin")
	}

	// Delete tenant-user relationship first
	if err := s.tenantUserRepo.Delete(ctx, tenantUser.ID); err != nil {
		return fmt.Errorf("failed to delete tenant user relationship: %w", err)
	}

	// Check if user has other tenant relationships
	tenantUsers, err := s.tenantUserRepo.ListByUser(ctx, userID, 1, 0)
	if err != nil {
		return fmt.Errorf("failed to check user tenant relationships: %w", err)
	}

	// If no other tenant relationships, delete user
	if len(tenantUsers) == 0 {
		if err := s.DeleteUser(ctx, userID); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.LogEvent(ctx, tenantID, &userID, domain.ActionDelete, domain.ResourceUser, userID.String(), nil)
	}

	return nil
}

// ListTenantAdmins lists tenant admin users for a specific tenant
func (s *UserServiceImpl) ListTenantAdmins(ctx context.Context, tenantID uuid.UUID, req *ListUsersRequest) (*UserListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit
	users, err := s.userRepo.ListTenantAdmins(ctx, tenantID, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenant admins: %w", err)
	}

	total, err := s.userRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count tenant users: %w", err)
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return &UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(req.Limit))),
	}, nil
}

// DeleteUser deletes a user
func (s *UserServiceImpl) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Soft delete user
	if err := s.userRepo.SoftDelete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Revoke all sessions
	if err := s.RevokeAllSessions(ctx, userID); err != nil {
		// Log error but don't fail
		fmt.Printf("Warning: failed to revoke sessions for user %s: %v\n", userID, err)
	}

	// Audit log
	if s.auditService != nil {
		s.auditService.LogEvent(ctx, uuid.Nil, &userID, domain.ActionDelete, domain.ResourceUser, user.ID.String(), nil)
	}

	return nil
}

// ListUsers lists all users with pagination and filtering
func (s *UserServiceImpl) ListUsers(ctx context.Context, req *ListUsersRequest) (*UserListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit
	users, err := s.userRepo.List(ctx, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return &UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(req.Limit))),
	}, nil
}

// ListUsersByTenant lists users by tenant
func (s *UserServiceImpl) ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, req *ListUsersRequest) (*UserListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit
	users, err := s.userRepo.ListByTenant(ctx, tenantID, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by tenant: %w", err)
	}

	total, err := s.userRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to count users by tenant: %w", err)
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return &UserListResponse{
		Users:      userResponses,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: int(math.Ceil(float64(total) / float64(req.Limit))),
	}, nil
}

// SearchUsers searches users by query
func (s *UserServiceImpl) SearchUsers(ctx context.Context, query string, req *ListUsersRequest) (*UserListResponse, error) {
	if req.Limit == 0 {
		req.Limit = 20
	}
	if req.Page == 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit
	users, err := s.userRepo.Search(ctx, query, req.Limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return &UserListResponse{
		Users:      userResponses,
		Total:      int64(len(users)),
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: int(math.Ceil(float64(len(users)) / float64(req.Limit))),
	}, nil
}

// GetUserProfile gets user profile
func (s *UserServiceImpl) GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &UserProfile{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		Avatar:        user.Avatar,
		AvatarURL:     user.AvatarURL,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		LastLoginAt:   user.LastLoginAt,
		LoginCount:    user.LoginCount,
		Preferences:   user.Preferences,
		Metadata:      user.Metadata,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

// UpdateUserProfile updates user profile
func (s *UserServiceImpl) UpdateUserProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*UserProfile, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update profile fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Metadata != nil {
		for key, value := range req.Metadata {
			user.Metadata[key] = value
		}
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	return s.GetUserProfile(ctx, userID)
}

// DeleteAvatar deletes user avatar
func (s *UserServiceImpl) DeleteAvatar(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	if user.AvatarURL != nil {
		// Find and delete the avatar file
		files, err := s.fileRepo.ListByUserAndCategory(ctx, userID, "avatar", 10, 0)
		if err == nil {
			for _, file := range files {
				if file.URL == *user.AvatarURL {
					s.fileService.DeleteFile(ctx, file.ID)
					break
				}
			}
		}
	}

	// Clear avatar URL
	user.AvatarURL = nil
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// UnassignRole unassigns a role from a user
func (s *UserServiceImpl) UnassignRole(ctx context.Context, userID, tenantID, roleID uuid.UUID) error {
	return s.userRoleRepo.DeleteByUserAndRole(ctx, userID, roleID, tenantID)
}

// GetUserRoles gets user roles
func (s *UserServiceImpl) GetUserRoles(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) ([]*domain.UserRole, error) {
	if tenantID != nil {
		return s.userRoleRepo.GetByUserAndTenant(ctx, userID, *tenantID)
	}
	return s.userRoleRepo.ListByUser(ctx, userID)
}

// UpdateUserRoles updates user roles for a tenant
func (s *UserServiceImpl) UpdateUserRoles(ctx context.Context, userID, tenantID uuid.UUID, roleIDs []uuid.UUID) error {
	// Delete existing roles
	if err := s.userRoleRepo.DeleteByUserAndTenant(ctx, userID, tenantID); err != nil {
		return fmt.Errorf("failed to delete existing roles: %w", err)
	}

	// Add new roles
	for _, roleID := range roleIDs {
		if err := s.AssignRole(ctx, userID, tenantID, roleID); err != nil {
			return fmt.Errorf("failed to assign role %s: %w", roleID, err)
		}
	}

	return nil
}

// GetUserPreferences gets user preferences
func (s *UserServiceImpl) GetUserPreferences(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID) (map[string]interface{}, error) {
	return s.prefRepo.GetUserPreferences(ctx, userID, tenantID)
}

// GetUserPreference gets a specific user preference
func (s *UserServiceImpl) GetUserPreference(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string) (interface{}, error) {
	pref, err := s.prefRepo.GetByUserAndKey(ctx, userID, tenantID, category, key)
	if err != nil {
		return nil, err
	}
	return pref.Value, nil
}

// SetUserPreference sets a specific user preference
func (s *UserServiceImpl) SetUserPreference(ctx context.Context, userID uuid.UUID, tenantID *uuid.UUID, category, key string, value interface{}) error {
	return s.prefRepo.UpsertPreference(ctx, userID, tenantID, category, key, value)
}

// Status management methods

// ActivateUser activates a user
func (s *UserServiceImpl) ActivateUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.ActivateUser(ctx, userID)
}

// DeactivateUser deactivates a user
func (s *UserServiceImpl) DeactivateUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.DeactivateUser(ctx, userID)
}

// SuspendUser suspends a user
func (s *UserServiceImpl) SuspendUser(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.SuspendUser(ctx, userID)
}

// VerifyEmail verifies user email
func (s *UserServiceImpl) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.EmailVerified = true
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// VerifyPhone verifies user phone
func (s *UserServiceImpl) VerifyPhone(ctx context.Context, userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.PhoneVerified = true
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(ctx, user)
}

// Session management methods

// UpdateLastLogin updates user's last login time
func (s *UserServiceImpl) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	return s.userRepo.UpdateLastLogin(ctx, userID)
}

// GetActiveSessions gets active sessions for a user
func (s *UserServiceImpl) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.UserSession, error) {
	return s.sessionRepo.ListByUser(ctx, userID, 100, 0)
}

// RevokeSession revokes a specific session
func (s *UserServiceImpl) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessionRepo.Delete(ctx, sessionID)
}

// RevokeAllSessions revokes all sessions for a user
func (s *UserServiceImpl) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	return s.sessionRepo.DeleteByUserID(ctx, userID)
}

// Helper methods

func (s *UserServiceImpl) toUserResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Phone:         user.Phone,
		Avatar:        user.Avatar,
		AvatarURL:     user.AvatarURL,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		PhoneVerified: user.PhoneVerified,
		LastLoginAt:   user.LastLoginAt,
		LoginCount:    user.LoginCount,
		Preferences:   user.Preferences,
		Metadata:      user.Metadata,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}
