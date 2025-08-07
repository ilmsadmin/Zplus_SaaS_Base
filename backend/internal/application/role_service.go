package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/auth"
)

// RoleService provides role management functionality
type RoleService struct {
	roleRepo       domain.RoleRepository
	permissionRepo domain.PermissionRepository
	userRoleRepo   domain.UserRoleRepository
	casbinService  *auth.CasbinService
}

// NewRoleService creates a new role service
func NewRoleService(
	roleRepo domain.RoleRepository,
	permissionRepo domain.PermissionRepository,
	userRoleRepo domain.UserRoleRepository,
	casbinService *auth.CasbinService,
) *RoleService {
	return &RoleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRoleRepo:   userRoleRepo,
		casbinService:  casbinService,
	}
}

// CreateRoleInput represents the input for creating a role
type CreateRoleInput struct {
	Name        string      `json:"name" validate:"required,min=3,max=100"`
	Description string      `json:"description" validate:"max=500"`
	TenantID    *uuid.UUID  `json:"tenant_id"`
	Permissions []uuid.UUID `json:"permissions"`
}

// UpdateRoleInput represents the input for updating a role
type UpdateRoleInput struct {
	Name        *string     `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Description *string     `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions []uuid.UUID `json:"permissions,omitempty"`
}

// AssignRoleInput represents the input for assigning a role to a user
type AssignRoleInput struct {
	UserID   uuid.UUID `json:"user_id" validate:"required"`
	RoleID   uuid.UUID `json:"role_id" validate:"required"`
	TenantID uuid.UUID `json:"tenant_id" validate:"required"`
}

// CreateRole creates a new role
func (s *RoleService) CreateRole(ctx context.Context, input CreateRoleInput) (*domain.Role, error) {
	// Check if role name already exists in the same context (system or tenant)
	existing, err := s.roleRepo.GetByName(ctx, input.Name, input.TenantID)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("role with name '%s' already exists", input.Name)
	}

	// Create the role
	role := &domain.Role{
		Name:        input.Name,
		Description: input.Description,
		IsSystem:    input.TenantID == nil,
		TenantID:    input.TenantID,
	}

	// Get permissions if provided
	if len(input.Permissions) > 0 {
		for _, permID := range input.Permissions {
			perm, err := s.permissionRepo.GetByID(ctx, permID)
			if err != nil {
				return nil, fmt.Errorf("permission %s not found: %w", permID, err)
			}
			role.Permissions = append(role.Permissions, *perm)
		}
	}

	// Save the role
	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Sync permissions to Casbin
	if err := s.casbinService.SyncRolePermissions(ctx, role.ID); err != nil {
		return nil, fmt.Errorf("failed to sync role permissions: %w", err)
	}

	return role, nil
}

// GetRole retrieves a role by ID
func (s *RoleService) GetRole(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	return role, nil
}

// UpdateRole updates a role
func (s *RoleService) UpdateRole(ctx context.Context, roleID uuid.UUID, input UpdateRoleInput) (*domain.Role, error) {
	// Get existing role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("role not found: %w", err)
	}

	// Check if it's a system role (cannot be modified)
	if role.IsSystem {
		return nil, fmt.Errorf("system roles cannot be modified")
	}

	// Update fields
	if input.Name != nil {
		// Check for name conflicts
		existing, err := s.roleRepo.GetByName(ctx, *input.Name, role.TenantID)
		if err == nil && existing != nil && existing.ID != roleID {
			return nil, fmt.Errorf("role with name '%s' already exists", *input.Name)
		}
		role.Name = *input.Name
	}

	if input.Description != nil {
		role.Description = *input.Description
	}

	// Update permissions if provided
	if input.Permissions != nil {
		role.Permissions = []domain.Permission{}
		for _, permID := range input.Permissions {
			perm, err := s.permissionRepo.GetByID(ctx, permID)
			if err != nil {
				return nil, fmt.Errorf("permission %s not found: %w", permID, err)
			}
			role.Permissions = append(role.Permissions, *perm)
		}
	}

	// Save the role
	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Sync permissions to Casbin
	if err := s.casbinService.SyncRolePermissions(ctx, role.ID); err != nil {
		return nil, fmt.Errorf("failed to sync role permissions: %w", err)
	}

	return role, nil
}

// DeleteRole deletes a role
func (s *RoleService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Get the role to check if it's a system role
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	if role.IsSystem {
		return fmt.Errorf("system roles cannot be deleted")
	}

	// Check if role is assigned to any users
	count, err := s.userRoleRepo.CountByRole(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to check role usage: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete role with assigned users")
	}

	// Delete the role
	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

// ListRoles lists roles with pagination
func (s *RoleService) ListRoles(ctx context.Context, tenantID *uuid.UUID, limit, offset int) ([]*domain.Role, int64, error) {
	roles, err := s.roleRepo.List(ctx, tenantID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list roles: %w", err)
	}

	count, err := s.roleRepo.Count(ctx, tenantID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count roles: %w", err)
	}

	return roles, count, nil
}

// AssignRoleToUser assigns a role to a user in a tenant
func (s *RoleService) AssignRoleToUser(ctx context.Context, input AssignRoleInput) error {
	// Check if role exists
	role, err := s.roleRepo.GetByID(ctx, input.RoleID)
	if err != nil {
		return fmt.Errorf("role not found: %w", err)
	}

	// For tenant roles, ensure the role belongs to the same tenant
	if role.TenantID != nil && *role.TenantID != input.TenantID {
		return fmt.Errorf("role does not belong to the specified tenant")
	}

	// Check if assignment already exists
	existing, err := s.userRoleRepo.GetByUserAndTenant(ctx, input.UserID, input.TenantID)
	if err != nil {
		return fmt.Errorf("failed to check existing assignments: %w", err)
	}

	for _, ur := range existing {
		if ur.RoleID == input.RoleID && ur.Status == domain.StatusActive {
			return fmt.Errorf("user already has this role")
		}
	}

	// Create the assignment
	userRole := &domain.UserRole{
		UserID:   input.UserID,
		RoleID:   input.RoleID,
		TenantID: input.TenantID,
		Status:   domain.StatusActive,
	}

	if err := s.userRoleRepo.Create(ctx, userRole); err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	// Sync user roles to Casbin
	if err := s.casbinService.SyncUserRoles(ctx, input.UserID); err != nil {
		return fmt.Errorf("failed to sync user roles: %w", err)
	}

	return nil
}

// RemoveRoleFromUser removes a role from a user in a tenant
func (s *RoleService) RemoveRoleFromUser(ctx context.Context, userID, roleID, tenantID uuid.UUID) error {
	// Delete the assignment
	if err := s.userRoleRepo.DeleteByUserAndRole(ctx, userID, roleID, tenantID); err != nil {
		return fmt.Errorf("failed to remove role: %w", err)
	}

	// Sync user roles to Casbin
	if err := s.casbinService.SyncUserRoles(ctx, userID); err != nil {
		return fmt.Errorf("failed to sync user roles: %w", err)
	}

	return nil
}

// GetUserRoles gets all roles for a user across all tenants
func (s *RoleService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.UserRole, error) {
	userRoles, err := s.userRoleRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return userRoles, nil
}

// GetUserRolesInTenant gets all roles for a user in a specific tenant
func (s *RoleService) GetUserRolesInTenant(ctx context.Context, userID, tenantID uuid.UUID) ([]*domain.UserRole, error) {
	userRoles, err := s.userRoleRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles in tenant: %w", err)
	}
	return userRoles, nil
}

// GetUserPermissions gets all permissions for a user in a tenant
func (s *RoleService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	permissions, err := s.casbinService.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return permissions, nil
}

// HasPermission checks if a user has a specific permission in a tenant
func (s *RoleService) HasPermission(ctx context.Context, userID uuid.UUID, resource, action string, tenantID uuid.UUID) (bool, error) {
	allowed, err := s.casbinService.Enforce(userID, resource, action, tenantID)
	if err != nil {
		return false, fmt.Errorf("failed to check permission: %w", err)
	}
	return allowed, nil
}

// GetSystemRoles gets all system roles
func (s *RoleService) GetSystemRoles(ctx context.Context) ([]*domain.Role, error) {
	roles, err := s.roleRepo.ListSystemRoles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get system roles: %w", err)
	}
	return roles, nil
}

// GetTenantRoles gets all roles for a specific tenant
func (s *RoleService) GetTenantRoles(ctx context.Context, tenantID uuid.UUID) ([]*domain.Role, error) {
	roles, err := s.roleRepo.ListTenantRoles(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant roles: %w", err)
	}
	return roles, nil
}

// InitializeTenantRoles initializes default roles for a new tenant
func (s *RoleService) InitializeTenantRoles(ctx context.Context, tenantID uuid.UUID) error {
	// Use the Casbin service to create default tenant roles
	if err := s.casbinService.CreateDefaultTenantRoles(ctx, tenantID); err != nil {
		return fmt.Errorf("failed to initialize tenant roles: %w", err)
	}
	return nil
}
