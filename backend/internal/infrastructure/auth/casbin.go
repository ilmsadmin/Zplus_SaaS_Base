package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
)

// CasbinService provides authorization services using Casbin
type CasbinService struct {
	enforcer       *casbin.Enforcer
	db             *gorm.DB
	roleRepo       domain.RoleRepository
	permissionRepo domain.PermissionRepository
	userRoleRepo   domain.UserRoleRepository
}

// NewCasbinService creates a new Casbin service
func NewCasbinService(
	db *gorm.DB,
	roleRepo domain.RoleRepository,
	permissionRepo domain.PermissionRepository,
	userRoleRepo domain.UserRoleRepository,
) (*CasbinService, error) {
	// Initialize Gorm adapter
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin adapter: %w", err)
	}

	// Define the model
	modelText := `
[request_definition]
r = sub, obj, act, tenant

[policy_definition]
p = sub, obj, act, tenant

[role_definition]
g = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub, r.tenant) && r.obj == p.obj && r.act == p.act && r.tenant == p.tenant
`

	// Create model
	m, err := model.NewModelFromString(modelText)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin model: %w", err)
	}

	// Create enforcer
	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	// Enable auto-save
	enforcer.EnableAutoSave(true)

	service := &CasbinService{
		enforcer:       enforcer,
		db:             db,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRoleRepo:   userRoleRepo,
	}

	// Initialize default policies
	if err := service.initializeDefaultPolicies(); err != nil {
		return nil, fmt.Errorf("failed to initialize default policies: %w", err)
	}

	return service, nil
}

// Enforce checks if a user has permission to perform an action on a resource in a tenant
func (s *CasbinService) Enforce(userID uuid.UUID, resource, action string, tenantID uuid.UUID) (bool, error) {
	return s.enforcer.Enforce(userID.String(), resource, action, tenantID.String())
}

// EnforceSystemRole checks if a user has a system-level role
func (s *CasbinService) EnforceSystemRole(userID uuid.UUID, role string) (bool, error) {
	return s.enforcer.Enforce(userID.String(), "system", "access", "system")
}

// AddRoleForUser assigns a role to a user in a specific tenant
func (s *CasbinService) AddRoleForUser(userID, roleID, tenantID uuid.UUID) error {
	_, err := s.enforcer.AddRoleForUser(userID.String(), roleID.String(), tenantID.String())
	return err
}

// RemoveRoleForUser removes a role from a user in a specific tenant
func (s *CasbinService) RemoveRoleForUser(userID, roleID, tenantID uuid.UUID) error {
	_, err := s.enforcer.DeleteRoleForUser(userID.String(), roleID.String(), tenantID.String())
	return err
}

// GetRolesForUser gets all roles for a user in a specific tenant
func (s *CasbinService) GetRolesForUser(userID, tenantID uuid.UUID) ([]string, error) {
	return s.enforcer.GetRolesForUser(userID.String(), tenantID.String())
}

// GetUsersForRole gets all users with a specific role in a tenant
func (s *CasbinService) GetUsersForRole(roleID, tenantID uuid.UUID) ([]string, error) {
	return s.enforcer.GetUsersForRole(roleID.String(), tenantID.String())
}

// AddPermissionForRole adds a permission to a role
func (s *CasbinService) AddPermissionForRole(roleID uuid.UUID, resource, action string, tenantID uuid.UUID) error {
	_, err := s.enforcer.AddPermissionForUser(roleID.String(), resource, action, tenantID.String())
	return err
}

// RemovePermissionForRole removes a permission from a role
func (s *CasbinService) RemovePermissionForRole(roleID uuid.UUID, resource, action string, tenantID uuid.UUID) error {
	_, err := s.enforcer.DeletePermissionForUser(roleID.String(), resource, action, tenantID.String())
	return err
}

// GetPermissionsForRole gets all permissions for a role
func (s *CasbinService) GetPermissionsForRole(roleID uuid.UUID) ([][]string, error) {
	return s.enforcer.GetPermissionsForUser(roleID.String())
}

// SyncUserRoles synchronizes user roles from database to Casbin
func (s *CasbinService) SyncUserRoles(ctx context.Context, userID uuid.UUID) error {
	// Get user roles from database
	userRoles, err := s.userRoleRepo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user roles: %w", err)
	}

	// Clear existing roles for user
	s.enforcer.DeleteUser(userID.String())

	// Add roles to enforcer
	for _, ur := range userRoles {
		if ur.Status == domain.StatusActive {
			err := s.AddRoleForUser(userID, ur.RoleID, ur.TenantID)
			if err != nil {
				return fmt.Errorf("failed to add role for user: %w", err)
			}
		}
	}

	return nil
}

// SyncRolePermissions synchronizes role permissions from database to Casbin
func (s *CasbinService) SyncRolePermissions(ctx context.Context, roleID uuid.UUID) error {
	// Get role with permissions
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Clear existing permissions for role
	s.enforcer.DeletePermissionsForUser(roleID.String())

	// Add permissions to enforcer
	for _, perm := range role.Permissions {
		tenantIDStr := "system"
		if role.TenantID != nil {
			tenantIDStr = role.TenantID.String()
		}

		err := s.AddPermissionForRole(roleID, perm.Resource, perm.Action, uuid.MustParse(tenantIDStr))
		if err != nil {
			return fmt.Errorf("failed to add permission for role: %w", err)
		}
	}

	return nil
}

// initializeDefaultPolicies sets up default system policies
func (s *CasbinService) initializeDefaultPolicies() error {
	ctx := context.Background()

	// Initialize default permissions if they don't exist
	defaultPermissions := []domain.Permission{
		{Name: domain.PermSystemManageTenants, Resource: domain.ResourceTenant, Action: domain.ActionManage, Description: "Manage tenants"},
		{Name: domain.PermSystemManageUsers, Resource: domain.ResourceUser, Action: domain.ActionManage, Description: "Manage users"},
		{Name: domain.PermSystemViewAuditLogs, Resource: domain.ResourceAuditLog, Action: domain.ActionRead, Description: "View audit logs"},
		{Name: domain.PermSystemManageSettings, Resource: domain.ResourceSettings, Action: domain.ActionManage, Description: "Manage system settings"},

		{Name: domain.PermTenantManageUsers, Resource: domain.ResourceUser, Action: domain.ActionManage, Description: "Manage tenant users"},
		{Name: domain.PermTenantManageRoles, Resource: domain.ResourceRole, Action: domain.ActionManage, Description: "Manage tenant roles"},
		{Name: domain.PermTenantManageSettings, Resource: domain.ResourceSettings, Action: domain.ActionManage, Description: "Manage tenant settings"},
		{Name: domain.PermTenantManageDomains, Resource: domain.ResourceDomain, Action: domain.ActionManage, Description: "Manage tenant domains"},
		{Name: domain.PermTenantViewAuditLogs, Resource: domain.ResourceAuditLog, Action: domain.ActionRead, Description: "View tenant audit logs"},
		{Name: domain.PermTenantManageAPIKeys, Resource: domain.ResourceAPIKey, Action: domain.ActionManage, Description: "Manage tenant API keys"},

		{Name: domain.PermUserReadProfile, Resource: domain.ResourceUser, Action: domain.ActionRead, Description: "Read user profile"},
		{Name: domain.PermUserUpdateProfile, Resource: domain.ResourceUser, Action: domain.ActionUpdate, Description: "Update user profile"},
		{Name: domain.PermUserManageFiles, Resource: domain.ResourceFile, Action: domain.ActionManage, Description: "Manage user files"},
		{Name: domain.PermUserViewFiles, Resource: domain.ResourceFile, Action: domain.ActionRead, Description: "View user files"},
	}

	for _, perm := range defaultPermissions {
		existing, err := s.permissionRepo.GetByName(ctx, perm.Name)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to check permission existence: %w", err)
		}
		if existing == nil {
			if err := s.permissionRepo.Create(ctx, &perm); err != nil {
				return fmt.Errorf("failed to create permission: %w", err)
			}
		}
	}

	// Initialize default system roles
	systemRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{
			name:        domain.RoleSystemAdmin,
			description: "System Administrator with full access",
			permissions: []string{
				domain.PermSystemManageTenants,
				domain.PermSystemManageUsers,
				domain.PermSystemViewAuditLogs,
				domain.PermSystemManageSettings,
			},
		},
		{
			name:        domain.RoleSystemManager,
			description: "System Manager with limited admin access",
			permissions: []string{
				domain.PermSystemViewAuditLogs,
				domain.PermSystemManageSettings,
			},
		},
	}

	for _, roleData := range systemRoles {
		existing, err := s.roleRepo.GetByName(ctx, roleData.name, nil)
		if err != nil && !strings.Contains(err.Error(), "not found") {
			return fmt.Errorf("failed to check role existence: %w", err)
		}
		if existing == nil {
			role := &domain.Role{
				Name:        roleData.name,
				Description: roleData.description,
				IsSystem:    true,
				TenantID:    nil,
			}

			// Get permissions
			for _, permName := range roleData.permissions {
				perm, err := s.permissionRepo.GetByName(ctx, permName)
				if err != nil {
					return fmt.Errorf("failed to get permission %s: %w", permName, err)
				}
				role.Permissions = append(role.Permissions, *perm)
			}

			if err := s.roleRepo.Create(ctx, role); err != nil {
				return fmt.Errorf("failed to create role: %w", err)
			}

			// Sync to Casbin
			if err := s.SyncRolePermissions(ctx, role.ID); err != nil {
				return fmt.Errorf("failed to sync role permissions: %w", err)
			}
		}
	}

	return nil
}

// CreateDefaultTenantRoles creates default roles for a new tenant
func (s *CasbinService) CreateDefaultTenantRoles(ctx context.Context, tenantID uuid.UUID) error {
	tenantRoles := []struct {
		name        string
		description string
		permissions []string
	}{
		{
			name:        domain.RoleTenantAdmin,
			description: "Tenant Administrator with full tenant access",
			permissions: []string{
				domain.PermTenantManageUsers,
				domain.PermTenantManageRoles,
				domain.PermTenantManageSettings,
				domain.PermTenantManageDomains,
				domain.PermTenantViewAuditLogs,
				domain.PermTenantManageAPIKeys,
			},
		},
		{
			name:        domain.RoleTenantManager,
			description: "Tenant Manager with limited admin access",
			permissions: []string{
				domain.PermTenantManageUsers,
				domain.PermTenantViewAuditLogs,
			},
		},
		{
			name:        domain.RoleUser,
			description: "Standard user role",
			permissions: []string{
				domain.PermUserReadProfile,
				domain.PermUserUpdateProfile,
				domain.PermUserManageFiles,
			},
		},
		{
			name:        domain.RoleViewer,
			description: "Read-only access",
			permissions: []string{
				domain.PermUserReadProfile,
				domain.PermUserViewFiles,
			},
		},
	}

	for _, roleData := range tenantRoles {
		role := &domain.Role{
			Name:        roleData.name,
			Description: roleData.description,
			IsSystem:    false,
			TenantID:    &tenantID,
		}

		// Get permissions
		for _, permName := range roleData.permissions {
			perm, err := s.permissionRepo.GetByName(ctx, permName)
			if err != nil {
				return fmt.Errorf("failed to get permission %s: %w", permName, err)
			}
			role.Permissions = append(role.Permissions, *perm)
		}

		if err := s.roleRepo.Create(ctx, role); err != nil {
			return fmt.Errorf("failed to create tenant role: %w", err)
		}

		// Sync to Casbin
		if err := s.SyncRolePermissions(ctx, role.ID); err != nil {
			return fmt.Errorf("failed to sync tenant role permissions: %w", err)
		}
	}

	return nil
}

// ValidateTenantAccess checks if user has access to a specific tenant
func (s *CasbinService) ValidateTenantAccess(ctx context.Context, userID, tenantID uuid.UUID) (bool, error) {
	// Get user roles for the tenant
	userRoles, err := s.userRoleRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}

	// Check if user has any active role in the tenant
	for _, ur := range userRoles {
		if ur.Status == domain.StatusActive {
			return true, nil
		}
	}

	return false, nil
}

// GetUserPermissions gets all permissions for a user in a tenant
func (s *CasbinService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	// Get user roles
	userRoles, err := s.userRoleRepo.GetByUserAndTenant(ctx, userID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	permissionSet := make(map[string]bool)
	var permissions []string

	// Collect permissions from all active roles
	for _, ur := range userRoles {
		if ur.Status != domain.StatusActive {
			continue
		}

		role, err := s.roleRepo.GetByID(ctx, ur.RoleID)
		if err != nil {
			continue
		}

		for _, perm := range role.Permissions {
			if !permissionSet[perm.Name] {
				permissionSet[perm.Name] = true
				permissions = append(permissions, perm.Name)
			}
		}
	}

	return permissions, nil
}
