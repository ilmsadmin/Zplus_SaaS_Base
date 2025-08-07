package interfaces

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ilmsadmin/zplus-saas-base/internal/application"
	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/middleware"
	"go.uber.org/zap"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService *application.AuthService
	logger      *zap.Logger
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(
	authService *application.AuthService,
	logger *zap.Logger,
) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// SystemAdminLogin handles system admin login requests
// Route: POST /auth/system-admin/login
func (h *AuthHandler) SystemAdminLogin(c *fiber.Ctx) error {
	var req application.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("Failed to parse login request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	// Validate request (basic validation)
	if req.Username == "" || req.Password == "" {
		h.logger.Warn("Login request validation failed: missing credentials")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Username and password are required",
		})
	}

	// Perform login
	resp, err := h.authService.SystemAdminLogin(c.Context(), req)
	if err != nil {
		h.logger.Error("System admin login failed", zap.Error(err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "login_failed",
			"message": "Authentication failed",
		})
	}

	h.logger.Info("System admin login successful",
		zap.String("username", req.Username),
		zap.String("user_id", resp.User.ID))

	// Set secure HTTP-only cookies for token storage
	h.setTokenCookies(c, resp, "admin")

	return c.JSON(fiber.Map{
		"success":      true,
		"message":      "Login successful",
		"user":         resp.User,
		"redirect_url": resp.RedirectURL,
		"permissions":  resp.Permissions,
	})
}

// TenantAdminLogin handles tenant admin login requests
// Route: POST /auth/tenant-admin/login
func (h *AuthHandler) TenantAdminLogin(c *fiber.Ctx) error {
	// Extract tenant domain from Host header or context
	tenantDomain := h.extractTenantDomain(c)
	if tenantDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_tenant",
			"message": "Tenant domain not found",
		})
	}

	var req application.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("Failed to parse tenant admin login request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	// Validate request (basic validation)
	if req.Username == "" || req.Password == "" {
		h.logger.Warn("Tenant admin login request validation failed: missing credentials")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Username and password are required",
		})
	}

	// Perform login
	resp, err := h.authService.TenantAdminLogin(c.Context(), req, tenantDomain)
	if err != nil {
		h.logger.Error("Tenant admin login failed",
			zap.Error(err),
			zap.String("tenant_domain", tenantDomain))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "login_failed",
			"message": "Authentication failed",
		})
	}

	h.logger.Info("Tenant admin login successful",
		zap.String("username", req.Username),
		zap.String("user_id", resp.User.ID),
		zap.String("tenant_domain", tenantDomain))

	// Set secure HTTP-only cookies for token storage
	h.setTokenCookies(c, resp, "tenant")

	return c.JSON(fiber.Map{
		"success":      true,
		"message":      "Login successful",
		"user":         resp.User,
		"redirect_url": resp.RedirectURL,
		"permissions":  resp.Permissions,
	})
}

// UserLogin handles regular user login requests
// Route: POST /auth/user/login
func (h *AuthHandler) UserLogin(c *fiber.Ctx) error {
	// Extract tenant domain from Host header or context
	tenantDomain := h.extractTenantDomain(c)
	if tenantDomain == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "missing_tenant",
			"message": "Tenant domain not found",
		})
	}

	var req application.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Warn("Failed to parse user login request", zap.Error(err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	// Validate request (basic validation)
	if req.Username == "" || req.Password == "" {
		h.logger.Warn("User login request validation failed: missing credentials")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "validation_error",
			"message": "Username and password are required",
		})
	}

	// Perform login
	resp, err := h.authService.UserLogin(c.Context(), req, tenantDomain)
	if err != nil {
		h.logger.Error("User login failed",
			zap.Error(err),
			zap.String("tenant_domain", tenantDomain))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "login_failed",
			"message": "Authentication failed",
		})
	}

	h.logger.Info("User login successful",
		zap.String("username", req.Username),
		zap.String("user_id", resp.User.ID),
		zap.String("tenant_domain", tenantDomain))

	// Set secure HTTP-only cookies for token storage
	h.setTokenCookies(c, resp, "user")

	return c.JSON(fiber.Map{
		"success":      true,
		"message":      "Login successful",
		"user":         resp.User,
		"redirect_url": resp.RedirectURL,
		"permissions":  resp.Permissions,
	})
}

// RefreshToken handles token refresh requests
// Route: POST /auth/refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get refresh token from cookie
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "missing_refresh_token",
			"message": "Refresh token not found",
		})
	}

	// Get client ID from cookie or header
	clientID := c.Cookies("client_id")
	if clientID == "" {
		clientID = c.Get("X-Client-ID", "zplus-tenant-frontend")
	}

	// Refresh token
	resp, err := h.authService.RefreshToken(c.Context(), refreshToken, clientID)
	if err != nil {
		h.logger.Error("Token refresh failed", zap.Error(err))
		// Clear invalid cookies
		h.clearTokenCookies(c)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "refresh_failed",
			"message": "Token refresh failed",
		})
	}

	h.logger.Info("Token refreshed successfully", zap.String("user_id", resp.User.ID))

	// Update cookies with new tokens
	h.setTokenCookies(c, resp, h.getLoginType(clientID))

	return c.JSON(fiber.Map{
		"success":     true,
		"message":     "Token refreshed successfully",
		"user":        resp.User,
		"permissions": resp.Permissions,
	})
}

// Logout handles logout requests
// Route: POST /auth/logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get refresh token from cookie
	refreshToken := c.Cookies("refresh_token")
	if refreshToken == "" {
		// Already logged out
		return c.JSON(fiber.Map{
			"success": true,
			"message": "Logout successful",
		})
	}

	// Get client ID from cookie
	clientID := c.Cookies("client_id")
	if clientID == "" {
		clientID = "zplus-tenant-frontend"
	}

	// Logout from Keycloak
	err := h.authService.Logout(c.Context(), refreshToken, clientID)
	if err != nil {
		h.logger.Error("Logout failed", zap.Error(err))
		// Continue with local logout even if Keycloak logout fails
	}

	// Clear cookies
	h.clearTokenCookies(c)

	h.logger.Info("User logged out successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logout successful",
	})
}

// ValidateToken handles token validation requests
// Route: GET /auth/validate
func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	// Get access token from Authorization header or cookie
	accessToken := c.Get("Authorization")
	if accessToken == "" {
		accessToken = c.Cookies("access_token")
		if accessToken != "" {
			accessToken = "Bearer " + accessToken
		}
	}

	if accessToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"valid": false,
			"error": "missing_token",
		})
	}

	// Validate token
	tokenInfo := h.authService.ValidateToken(c.Context(), accessToken)
	if !tokenInfo.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(tokenInfo)
	}

	return c.JSON(tokenInfo)
}

// GetProfile returns current user profile
// Route: GET /auth/profile
func (h *AuthHandler) GetProfile(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	user, ok := middleware.GetUserFromContext(c)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   "unauthenticated",
			"message": "User not authenticated",
		})
	}

	// Extract user info
	userInfo := &application.UserInfo{
		ID:            user.Subject,
		Username:      user.PreferredUsername,
		Email:         user.Email,
		FirstName:     user.GivenName,
		LastName:      user.FamilyName,
		TenantID:      user.TenantID,
		TenantDomain:  user.TenantDomain,
		Roles:         user.RealmAccess.Roles,
		IsSystemAdmin: user.IsSystemAdmin(),
		IsTenantAdmin: user.IsTenantAdmin(),
	}

	return c.JSON(fiber.Map{
		"success": true,
		"user":    userInfo,
	})
}

// Helper methods

func (h *AuthHandler) extractTenantDomain(c *fiber.Ctx) string {
	// Try to get from context (set by tenant middleware)
	if tenantCtx, exists := middleware.GetTenantFromContext(c); exists {
		// Return custom domain if available, otherwise subdomain
		if tenantCtx.CustomDomain != "" {
			return tenantCtx.CustomDomain
		}
		return tenantCtx.Subdomain + ".zplus.io"
	}

	// Try to get from header
	if domain := c.Get("X-Tenant-Domain"); domain != "" {
		return domain
	}

	// Extract from Host header
	host := c.Get("Host")
	if host == "" {
		return ""
	}

	// Parse subdomain from host (e.g., "acme.zplus.io" -> "acme.zplus.io")
	// This is a simplified extraction - in production you'd want more robust parsing
	return host
}

func (h *AuthHandler) setTokenCookies(c *fiber.Ctx, resp *application.LoginResponse, loginType string) {
	// Set access token cookie (short expiry)
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    resp.AccessToken,
		MaxAge:   resp.ExpiresIn,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	})

	// Set refresh token cookie (longer expiry)
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    resp.RefreshToken,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	})

	// Set client ID cookie for refresh operations
	clientID := h.getClientIDForLoginType(loginType)
	c.Cookie(&fiber.Cookie{
		Name:     "client_id",
		Value:    clientID,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	})

	// Set user info cookie (for client-side access)
	c.Cookie(&fiber.Cookie{
		Name:     "user_type",
		Value:    loginType,
		MaxAge:   resp.ExpiresIn,
		HTTPOnly: false, // Allow client-side access
		Secure:   true,
		SameSite: "Lax",
		Path:     "/",
	})
}

func (h *AuthHandler) clearTokenCookies(c *fiber.Ctx) {
	cookies := []string{"access_token", "refresh_token", "client_id", "user_type"}
	for _, name := range cookies {
		c.Cookie(&fiber.Cookie{
			Name:     name,
			Value:    "",
			MaxAge:   -1,
			HTTPOnly: true,
			Secure:   true,
			SameSite: "Lax",
			Path:     "/",
		})
	}
}

func (h *AuthHandler) getClientIDForLoginType(loginType string) string {
	switch loginType {
	case "admin":
		return "zplus-admin-frontend"
	case "tenant":
		return "zplus-tenant-frontend"
	case "user":
		return "zplus-tenant-frontend"
	default:
		return "zplus-tenant-frontend"
	}
}

func (h *AuthHandler) getLoginType(clientID string) string {
	switch clientID {
	case "zplus-admin-frontend":
		return "admin"
	case "zplus-tenant-frontend":
		return "tenant"
	default:
		return "user"
	}
}
