package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	"github.com/ilmsadmin/zplus-saas-base/internal/domain"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/auth"
	"go.uber.org/zap"
)

// AuthMiddleware creates a middleware for JWT authentication with Keycloak
func AuthMiddleware(validator *auth.KeycloakValidator, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing Authorization header")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "missing_authorization_header",
				"message": "Authorization header is required",
			})
		}

		// Validate token
		claims, err := validator.ValidateToken(authHeader)
		if err != nil {
			logger.Warn("Token validation failed", zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "invalid_token",
				"message": "Token validation failed",
			})
		}

		// Store claims in context
		c.Locals("user", claims)
		c.Locals("user_id", claims.Subject)
		c.Locals("username", claims.PreferredUsername)
		c.Locals("email", claims.Email)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("tenant_domain", claims.TenantDomain)

		return c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware for optional JWT authentication
func OptionalAuthMiddleware(validator *auth.KeycloakValidator, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No auth header, continue without authentication
			return c.Next()
		}

		// Try to validate token
		claims, err := validator.ValidateToken(authHeader)
		if err != nil {
			logger.Debug("Optional token validation failed", zap.Error(err))
			// Continue without authentication on validation failure
			return c.Next()
		}

		// Store claims in context if validation succeeds
		c.Locals("user", claims)
		c.Locals("user_id", claims.Subject)
		c.Locals("username", claims.PreferredUsername)
		c.Locals("email", claims.Email)
		c.Locals("tenant_id", claims.TenantID)
		c.Locals("tenant_domain", claims.TenantDomain)

		return c.Next()
	}
}

// RequireRole creates a middleware that requires specific roles
func RequireRole(roles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthenticated",
				"message": "Authentication required",
			})
		}

		claims, ok := user.(*auth.TokenClaims)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user data in context",
			})
		}

		// Check if user has any of the required roles
		hasRole := false
		for _, role := range roles {
			if claims.HasRole(role) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":          "insufficient_permissions",
				"message":        "Required role not found",
				"required_roles": roles,
			})
		}

		return c.Next()
	}
}

// RequirePermission creates a middleware that requires specific permissions
func RequirePermission(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthenticated",
				"message": "Authentication required",
			})
		}

		claims, ok := user.(*auth.TokenClaims)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user data in context",
			})
		}

		// Check if user has any of the required permissions
		hasPermission := false
		for _, permission := range permissions {
			if claims.HasPermission(permission) {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":                "insufficient_permissions",
				"message":              "Required permission not found",
				"required_permissions": permissions,
			})
		}

		return c.Next()
	}
}

// RequireSystemAdmin creates a middleware that requires system admin role
func RequireSystemAdmin() fiber.Handler {
	return RequireRole("system_admin")
}

// RequireTenantAdmin creates a middleware that requires tenant admin role or higher
func RequireTenantAdmin() fiber.Handler {
	return RequireRole("system_admin", "tenant_admin")
}

// TenantIsolationMiddleware ensures users can only access their own tenant data
func TenantIsolationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		user := c.Locals("user")
		if user == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthenticated",
				"message": "Authentication required",
			})
		}

		claims, ok := user.(*auth.TokenClaims)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user data in context",
			})
		}

		// System admins can access any tenant
		if claims.IsSystemAdmin() {
			return c.Next()
		}

		// Get tenant from URL parameter or query
		requestedTenant := c.Params("tenant_id")
		if requestedTenant == "" {
			requestedTenant = c.Query("tenant_id")
		}

		// If no tenant specified in request, allow (will be handled by business logic)
		if requestedTenant == "" {
			return c.Next()
		}

		// Check if user can access the requested tenant
		if claims.TenantID != requestedTenant {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":            "tenant_access_denied",
				"message":          "Access to this tenant is not allowed",
				"user_tenant":      claims.TenantID,
				"requested_tenant": requestedTenant,
			})
		}

		return c.Next()
	}
}

// GetUserFromContext extracts user claims from Fiber context
func GetUserFromContext(c *fiber.Ctx) (*auth.TokenClaims, bool) {
	user := c.Locals("user")
	if user == nil {
		return nil, false
	}

	claims, ok := user.(*auth.TokenClaims)
	return claims, ok
}

// GetUserIDFromContext extracts user ID from Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (string, bool) {
	userID := c.Locals("user_id")
	if userID == nil {
		return "", false
	}

	id, ok := userID.(string)
	return id, ok
}

// GetTenantIDFromContext extracts tenant ID from Fiber context
func GetTenantIDFromContext(c *fiber.Ctx) (string, bool) {
	tenantID := c.Locals("tenant_id")
	if tenantID == nil {
		return "", false
	}

	id, ok := tenantID.(string)
	return id, ok
}

// CORSConfig provides CORS configuration for Keycloak integration
func CORSConfig() cors.Config {
	return cors.Config{
		AllowOrigins: strings.Join([]string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://admin.localhost",
			"http://*.localhost",
			"https://admin.zplus.io",
			"https://*.zplus.io",
		}, ","),
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodPut,
			fiber.MethodPatch,
			fiber.MethodDelete,
			fiber.MethodOptions,
			fiber.MethodHead,
		}, ","),
		AllowHeaders: strings.Join([]string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Tenant-ID",
		}, ","),
		AllowCredentials: true,
		ExposeHeaders: strings.Join([]string{
			"Content-Length",
			"Content-Type",
			"X-Total-Count",
		}, ","),
		MaxAge: 86400, // 24 hours
	}
}

// CasbinMiddleware provides Casbin-based authorization
func CasbinMiddleware(casbinService *auth.CasbinService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context (set by AuthMiddleware)
		userIDInterface := c.Locals("user_id")
		if userIDInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthenticated",
				"message": "User not authenticated",
			})
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID format",
			})
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID",
			})
		}

		// Get tenant ID from context (set by TenantMiddleware)
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "missing_tenant",
				"message": "Tenant not found",
			})
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID format",
			})
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID",
			})
		}

		// Validate tenant access
		hasAccess, err := casbinService.ValidateTenantAccess(context.Background(), userID, tenantID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "authorization_error",
				"message": "Failed to validate tenant access",
			})
		}

		if !hasAccess {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "no_tenant_access",
				"message": "No access to this tenant",
			})
		}

		return c.Next()
	}
}

// RequireCasbinPermission creates middleware that checks Casbin permissions
func RequireCasbinPermission(resource, action string, casbinService *auth.CasbinService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userIDInterface := c.Locals("user_id")
		if userIDInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthenticated",
				"message": "User not authenticated",
			})
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID format",
			})
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID",
			})
		}

		// Get tenant ID from context
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "missing_tenant",
				"message": "Tenant not found",
			})
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID format",
			})
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID",
			})
		}

		// Check permission using Casbin
		allowed, err := casbinService.Enforce(userID, resource, action, tenantID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "authorization_error",
				"message": "Failed to check permissions",
			})
		}

		if !allowed {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "insufficient_permissions",
				"message": fmt.Sprintf("No permission for %s:%s", resource, action),
			})
		}

		return c.Next()
	}
}

// RequireSystemRoleCasbin creates middleware that checks for system roles using Casbin
func RequireSystemRoleCasbin(roles []string, casbinService *auth.CasbinService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userIDInterface := c.Locals("user_id")
		if userIDInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthenticated",
				"message": "User not authenticated",
			})
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID format",
			})
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID",
			})
		}

		// Check if user has any of the required system roles
		hasRole := false
		for _, role := range roles {
			allowed, err := casbinService.EnforceSystemRole(userID, role)
			if err != nil {
				continue
			}
			if allowed {
				hasRole = true
				break
			}
		}

		if !hasRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "insufficient_system_role",
				"message": fmt.Sprintf("Required system role: %s", strings.Join(roles, " or ")),
			})
		}

		return c.Next()
	}
}

// RequireTenantRoleCasbin creates middleware that checks for tenant roles using Casbin
func RequireTenantRoleCasbin(roles []string, casbinService *auth.CasbinService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context
		userIDInterface := c.Locals("user_id")
		if userIDInterface == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   "unauthenticated",
				"message": "User not authenticated",
			})
		}

		userIDStr, ok := userIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID format",
			})
		}

		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_user_data",
				"message": "Invalid user ID",
			})
		}

		// Get tenant ID from context
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "missing_tenant",
				"message": "Tenant not found",
			})
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID format",
			})
		}

		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID",
			})
		}

		// Get user roles in the tenant
		userRoles, err := casbinService.GetRolesForUser(userID, tenantID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "authorization_error",
				"message": "Failed to get user roles",
			})
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
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "insufficient_tenant_role",
				"message": fmt.Sprintf("Required tenant role: %s", strings.Join(roles, " or ")),
			})
		}

		return c.Next()
	}
}

// TenantIsolationEnhanced creates enhanced tenant isolation middleware
func TenantIsolationEnhanced(casbinService *auth.CasbinService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get tenant ID from context (set by TenantMiddleware)
		tenantIDInterface := c.Locals("tenant_id")
		if tenantIDInterface == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "missing_tenant",
				"message": "Tenant not found in request",
			})
		}

		tenantIDStr, ok := tenantIDInterface.(string)
		if !ok {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID format",
			})
		}

		// Validate tenant ID format
		tenantID, err := uuid.Parse(tenantIDStr)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "invalid_tenant_data",
				"message": "Invalid tenant ID",
			})
		}

		// Check if there are any path parameters that should match tenant context
		pathTenantID := c.Params("tenantId")
		if pathTenantID != "" && pathTenantID != tenantIDStr {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "tenant_mismatch",
				"message": "Tenant ID mismatch in path",
			})
		}

		// Add tenant isolation header for database queries
		c.Set("X-Tenant-Isolation", tenantIDStr)
		c.Locals("validated_tenant_id", tenantID)

		return c.Next()
	}
}

// AdminOnlyMiddleware restricts access to admin routes
func AdminOnlyMiddleware(casbinService *auth.CasbinService) fiber.Handler {
	return RequireTenantRoleCasbin([]string{
		domain.RoleTenantAdmin,
		domain.RoleSystemAdmin,
	}, casbinService)
}

// SystemAdminOnlyMiddleware restricts access to system admin only
func SystemAdminOnlyMiddleware(casbinService *auth.CasbinService) fiber.Handler {
	return RequireSystemRoleCasbin([]string{
		domain.RoleSystemAdmin,
	}, casbinService)
}
