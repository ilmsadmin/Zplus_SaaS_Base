package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/auth"
)

// PermissionConfig defines configuration for permission middleware
type PermissionConfig struct {
	// Next defines a function to skip this middleware when returned true.
	Next func(c *fiber.Ctx) bool

	// ContextKey is the key to store the user in the context
	ContextKey string

	// Skipper is a function to skip this middleware
	Skipper func(c *fiber.Ctx) bool

	// CasbinService for authorization
	CasbinService *auth.CasbinService

	// ErrorHandler defines a function which is executed for an invalid token.
	ErrorHandler fiber.ErrorHandler
}

// ConfigDefault is the default config
var PermissionConfigDefault = PermissionConfig{
	Next:       nil,
	ContextKey: "user_id",
	Skipper:    nil,
	ErrorHandler: func(c *fiber.Ctx, err error) error {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":   "Forbidden",
			"message": err.Error(),
		})
	},
}

// RequirePermission middleware that checks if user has required permission
func RequirePermission(resource, action string, config ...PermissionConfig) fiber.Handler {
	// Set default config
	cfg := PermissionConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default values
	if cfg.ContextKey == "" {
		cfg.ContextKey = PermissionConfigDefault.ContextKey
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = PermissionConfigDefault.ErrorHandler
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Skip middleware if Skipper returns true
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		// Get user ID from context
		userIDInterface := c.Locals(cfg.ContextKey)
		if userIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("user not authenticated"))
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID format"))
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID: %w", err))
		}

		// Get tenant ID from context
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("tenant not found"))
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID format"))
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID: %w", err))
		}

		// Check permission using Casbin
		allowed, err := cfg.CasbinService.Enforce(userID, resource, action, tenantID)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("authorization check failed: %w", err))
		}

		if !allowed {
			return cfg.ErrorHandler(c, fmt.Errorf("insufficient permissions for %s:%s", resource, action))
		}

		return c.Next()
	}
}

// RequireSystemRole middleware that checks if user has a system role
func RequireSystemRole(roles []string, config ...PermissionConfig) fiber.Handler {
	// Set default config
	cfg := PermissionConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default values
	if cfg.ContextKey == "" {
		cfg.ContextKey = PermissionConfigDefault.ContextKey
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = PermissionConfigDefault.ErrorHandler
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Skip middleware if Skipper returns true
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		// Get user ID from context
		userIDInterface := c.Locals(cfg.ContextKey)
		if userIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("user not authenticated"))
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID format"))
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID: %w", err))
		}

		// Check if user has any of the required system roles
		hasRole := false
		for _, role := range roles {
			allowed, err := cfg.CasbinService.EnforceSystemRole(userID, role)
			if err != nil {
				continue
			}
			if allowed {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return cfg.ErrorHandler(c, fmt.Errorf("insufficient system role, required one of: %s", strings.Join(roles, ", ")))
		}

		return c.Next()
	}
}

// RequireTenantRole middleware that checks if user has a specific role in current tenant
func RequireTenantRole(roles []string, config ...PermissionConfig) fiber.Handler {
	// Set default config
	cfg := PermissionConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default values
	if cfg.ContextKey == "" {
		cfg.ContextKey = PermissionConfigDefault.ContextKey
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = PermissionConfigDefault.ErrorHandler
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Skip middleware if Skipper returns true
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		// Get user ID from context
		userIDInterface := c.Locals(cfg.ContextKey)
		if userIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("user not authenticated"))
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID format"))
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID: %w", err))
		}

		// Get tenant ID from context
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("tenant not found"))
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID format"))
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID: %w", err))
		}

		// Get user roles in the tenant
		userRoles, err := cfg.CasbinService.GetRolesForUser(userID, tenantID)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("failed to get user roles: %w", err))
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, userRole := range userRoles {
			for _, requiredRole := range roles {
				if userRole == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			return cfg.ErrorHandler(c, fmt.Errorf("insufficient tenant role, required one of: %s", strings.Join(roles, ", ")))
		}

		return c.Next()
	}
}

// RequireTenantAccess middleware that checks if user has access to current tenant
func RequireTenantAccess(config ...PermissionConfig) fiber.Handler {
	// Set default config
	cfg := PermissionConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default values
	if cfg.ContextKey == "" {
		cfg.ContextKey = PermissionConfigDefault.ContextKey
	}
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = PermissionConfigDefault.ErrorHandler
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Skip middleware if Skipper returns true
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		// Get user ID from context
		userIDInterface := c.Locals(cfg.ContextKey)
		if userIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("user not authenticated"))
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID format"))
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid user ID: %w", err))
		}

		// Get tenant ID from context
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("tenant not found"))
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID format"))
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID: %w", err))
		}

		// Check tenant access
		hasAccess, err := cfg.CasbinService.ValidateTenantAccess(context.Background(), userID, tenantID)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("failed to validate tenant access: %w", err))
		}

		if !hasAccess {
			return cfg.ErrorHandler(c, fmt.Errorf("no access to tenant"))
		}

		return c.Next()
	}
}

// TenantIsolation middleware that ensures tenant isolation
func TenantIsolation(config ...PermissionConfig) fiber.Handler {
	// Set default config
	cfg := PermissionConfigDefault

	// Override config if provided
	if len(config) > 0 {
		cfg = config[0]
	}

	// Set default values
	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = PermissionConfigDefault.ErrorHandler
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Skip middleware if Skipper returns true
		if cfg.Skipper != nil && cfg.Skipper(c) {
			return c.Next()
		}

		// Get tenant ID from context (set by TenantMiddleware)
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return cfg.ErrorHandler(c, fmt.Errorf("tenant not found in request"))
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID format"))
		}

		// Validate tenant ID format
		_, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return cfg.ErrorHandler(c, fmt.Errorf("invalid tenant ID: %w", err))
		}

		// Check if there are any path parameters that should match tenant context
		// This is to prevent cross-tenant data access
		pathTenantID := c.Params("tenantId")
		if pathTenantID != "" && pathTenantID != tenantIDStr {
			return cfg.ErrorHandler(c, fmt.Errorf("tenant ID mismatch in path"))
		}

		// Add tenant isolation header for database queries
		c.Set("X-Tenant-Isolation", tenantIDStr)

		return c.Next()
	}
}

// AdminOnly middleware that restricts access to admin routes
func AdminOnly(config ...PermissionConfig) fiber.Handler {
	return RequireTenantRole([]string{
		domain.RoleTenantAdmin,
		domain.RoleSystemAdmin,
	}, config...)
}

// SystemAdminOnly middleware that restricts access to system admin only
func SystemAdminOnly(config ...PermissionConfig) fiber.Handler {
	return RequireSystemRole([]string{
		domain.RoleSystemAdmin,
	}, config...)
}
