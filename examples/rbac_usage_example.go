package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/application"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/auth"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/database/postgres"
	"gorm.io/gorm"
)

// Example demonstrating RBAC system usage
func main() {
	// Initialize database connection (assumes you have this setup)
	db := initializeDatabase() // Your database initialization

	// Initialize repositories
	roleRepo := postgres.NewRoleRepository(db)
	permissionRepo := postgres.NewPermissionRepository(db)
	userRoleRepo := postgres.NewUserRoleRepository(db)

	// Initialize Casbin service
	casbinService, err := auth.NewCasbinService(db, roleRepo, permissionRepo, userRoleRepo)
	if err != nil {
		log.Fatal("Failed to initialize Casbin service:", err)
	}

	// Initialize role service
	roleService := application.NewRoleService(roleRepo, permissionRepo, userRoleRepo, casbinService)

	ctx := context.Background()

	// Example 1: Create a custom tenant role
	fmt.Println("=== Example 1: Creating Custom Tenant Role ===")

	tenantID := uuid.New() // Assume this is a valid tenant ID

	// Get permissions for the role
	fileManagePermission, _ := permissionRepo.GetByName(ctx, domain.PermUserManageFiles)
	fileViewPermission, _ := permissionRepo.GetByName(ctx, domain.PermUserViewFiles)

	customRole, err := roleService.CreateRole(ctx, application.CreateRoleInput{
		Name:        "file_manager",
		Description: "File Manager with file management permissions",
		TenantID:    &tenantID,
		Permissions: []uuid.UUID{
			fileManagePermission.ID,
			fileViewPermission.ID,
		},
	})

	if err != nil {
		log.Printf("Error creating role: %v", err)
	} else {
		fmt.Printf("Created role: %s (ID: %s)\n", customRole.Name, customRole.ID)
	}

	// Example 2: Assign role to user
	fmt.Println("\n=== Example 2: Assigning Role to User ===")

	userID := uuid.New() // Assume this is a valid user ID

	err = roleService.AssignRoleToUser(ctx, application.AssignRoleInput{
		UserID:   userID,
		RoleID:   customRole.ID,
		TenantID: tenantID,
	})

	if err != nil {
		log.Printf("Error assigning role: %v", err)
	} else {
		fmt.Printf("Assigned role '%s' to user %s in tenant %s\n",
			customRole.Name, userID, tenantID)
	}

	// Example 3: Check user permissions
	fmt.Println("\n=== Example 3: Checking User Permissions ===")

	hasFileManagePermission, err := roleService.HasPermission(ctx, userID, "file", "manage", tenantID)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
	} else {
		fmt.Printf("User has file manage permission: %t\n", hasFileManagePermission)
	}

	hasUserManagePermission, err := roleService.HasPermission(ctx, userID, "user", "manage", tenantID)
	if err != nil {
		log.Printf("Error checking permission: %v", err)
	} else {
		fmt.Printf("User has user manage permission: %t\n", hasUserManagePermission)
	}

	// Example 4: Get user permissions
	fmt.Println("\n=== Example 4: Getting All User Permissions ===")

	permissions, err := roleService.GetUserPermissions(ctx, userID, tenantID)
	if err != nil {
		log.Printf("Error getting permissions: %v", err)
	} else {
		fmt.Printf("User permissions in tenant:\n")
		for _, perm := range permissions {
			fmt.Printf("  - %s\n", perm)
		}
	}

	// Example 5: Initialize tenant with default roles
	fmt.Println("\n=== Example 5: Initialize Tenant with Default Roles ===")

	newTenantID := uuid.New()
	err = roleService.InitializeTenantRoles(ctx, newTenantID)
	if err != nil {
		log.Printf("Error initializing tenant roles: %v", err)
	} else {
		fmt.Printf("Initialized default roles for tenant %s\n", newTenantID)
	}

	// Example 6: List tenant roles
	fmt.Println("\n=== Example 6: Listing Tenant Roles ===")

	tenantRoles, err := roleService.GetTenantRoles(ctx, newTenantID)
	if err != nil {
		log.Printf("Error getting tenant roles: %v", err)
	} else {
		fmt.Printf("Tenant roles:\n")
		for _, role := range tenantRoles {
			fmt.Printf("  - %s: %s (%d permissions)\n",
				role.Name, role.Description, len(role.Permissions))
		}
	}

	// Example 7: System admin operations
	fmt.Println("\n=== Example 7: System Admin Operations ===")

	systemRoles, err := roleService.GetSystemRoles(ctx)
	if err != nil {
		log.Printf("Error getting system roles: %v", err)
	} else {
		fmt.Printf("System roles:\n")
		for _, role := range systemRoles {
			fmt.Printf("  - %s: %s\n", role.Name, role.Description)
		}
	}

	// Example 8: Tenant isolation demonstration
	fmt.Println("\n=== Example 8: Tenant Isolation Test ===")

	otherTenantID := uuid.New()

	// User should not have permissions in other tenant
	hasPermissionInOtherTenant, err := roleService.HasPermission(ctx, userID, "file", "manage", otherTenantID)
	if err != nil {
		log.Printf("Error checking cross-tenant permission: %v", err)
	} else {
		fmt.Printf("User has file manage permission in other tenant: %t (should be false)\n",
			hasPermissionInOtherTenant)
	}

	fmt.Println("\n=== RBAC Examples Completed ===")
}

// Mock database initialization - replace with your actual setup
func initializeDatabase() *gorm.DB {
	// This should return your actual database connection
	// For example purposes, returning nil
	return nil
}

// Example middleware usage in HTTP handlers
func ExampleMiddlewareUsage() {
	// This shows how to use RBAC middleware in your HTTP routes

	/*
		// Initialize your Fiber app and services
		app := fiber.New()

		// Setup middleware
		app.Use(middleware.TenantMiddleware(tenantRepo, domainRepo))
		app.Use(middleware.AuthMiddleware(keycloakValidator, logger))

		// System admin routes
		adminRoutes := app.Group("/admin")
		adminRoutes.Use(middleware.SystemAdminOnlyMiddleware(casbinService))
		adminRoutes.Get("/tenants", adminHandler.ListTenants)
		adminRoutes.Post("/tenants", adminHandler.CreateTenant)

		// Tenant admin routes
		tenantAdminRoutes := app.Group("/tenant/:tenantId/admin")
		tenantAdminRoutes.Use(middleware.AdminOnlyMiddleware(casbinService))
		tenantAdminRoutes.Get("/users", tenantHandler.ListUsers)
		tenantAdminRoutes.Post("/users", tenantHandler.CreateUser)

		// Regular user routes with specific permissions
		userRoutes := app.Group("/tenant/:tenantId")
		userRoutes.Use(middleware.TenantIsolationEnhanced(casbinService))

		// File management - requires file manage permission
		userRoutes.Get("/files",
			middleware.RequireCasbinPermission("file", "read", casbinService),
			fileHandler.ListFiles)
		userRoutes.Post("/files",
			middleware.RequireCasbinPermission("file", "manage", casbinService),
			fileHandler.UploadFile)
		userRoutes.Delete("/files/:id",
			middleware.RequireCasbinPermission("file", "manage", casbinService),
			fileHandler.DeleteFile)

		// Profile management
		userRoutes.Get("/profile",
			middleware.RequireCasbinPermission("user", "read", casbinService),
			userHandler.GetProfile)
		userRoutes.Put("/profile",
			middleware.RequireCasbinPermission("user", "update", casbinService),
			userHandler.UpdateProfile)
	*/
}

// Example GraphQL resolver usage
func ExampleGraphQLResolvers() {
	/*
		// In your GraphQL resolvers

		// Check permissions before resolving
		func (r *Resolver) Users(ctx context.Context, tenantID uuid.UUID) ([]*domain.User, error) {
			// Get user from context
			userID, ok := auth.GetUserIDFromContext(ctx)
			if !ok {
				return nil, errors.New("user not authenticated")
			}

			// Check permission
			hasPermission, err := r.roleService.HasPermission(ctx, userID, "user", "read", tenantID)
			if err != nil {
				return nil, err
			}
			if !hasPermission {
				return nil, errors.New("insufficient permissions")
			}

			// Proceed with resolver logic
			return r.userService.ListByTenant(ctx, tenantID)
		}

		// Role management resolvers
		func (r *Resolver) CreateRole(ctx context.Context, input CreateRoleInput) (*domain.Role, error) {
			// Verify admin permissions
			userID, _ := auth.GetUserIDFromContext(ctx)

			if input.TenantID != nil {
				// Tenant role - check tenant admin permission
				hasPermission, err := r.roleService.HasPermission(ctx, userID, "role", "manage", *input.TenantID)
				if err != nil || !hasPermission {
					return nil, errors.New("insufficient permissions to create tenant role")
				}
			} else {
				// System role - check system admin permission
				hasPermission, err := r.casbinService.EnforceSystemRole(userID, domain.RoleSystemAdmin)
				if err != nil || !hasPermission {
					return nil, errors.New("insufficient permissions to create system role")
				}
			}

			return r.roleService.CreateRole(ctx, input)
		}
	*/
}
