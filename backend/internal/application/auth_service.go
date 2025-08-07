package application

import (
	"context"
	"fmt"

	"github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/auth"
	"go.uber.org/zap"
)

// AuthService handles authentication logic
type AuthService struct {
	keycloakClient    *auth.KeycloakClient
	keycloakValidator *auth.KeycloakValidator
	logger            *zap.Logger
}

// NewAuthService creates a new authentication service
func NewAuthService(
	keycloakClient *auth.KeycloakClient,
	keycloakValidator *auth.KeycloakValidator,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{
		keycloakClient:    keycloakClient,
		keycloakValidator: keycloakValidator,
		logger:            logger,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	ClientID string `json:"client_id,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	TokenType    string          `json:"token_type"`
	ExpiresIn    int             `json:"expires_in"`
	User         *UserInfo       `json:"user"`
	RedirectURL  string          `json:"redirect_url"`
	Permissions  map[string]bool `json:"permissions"`
}

// UserInfo represents user information
type UserInfo struct {
	ID            string   `json:"id"`
	Username      string   `json:"username"`
	Email         string   `json:"email"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	TenantID      string   `json:"tenant_id,omitempty"`
	TenantDomain  string   `json:"tenant_domain,omitempty"`
	Roles         []string `json:"roles"`
	IsSystemAdmin bool     `json:"is_system_admin"`
	IsTenantAdmin bool     `json:"is_tenant_admin"`
}

// SystemAdminLogin handles system admin login via admin.zplus.io
func (s *AuthService) SystemAdminLogin(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	s.logger.Info("System admin login attempt", zap.String("username", req.Username))

	// Use admin frontend client ID for system admin login
	clientID := "zplus-admin-frontend"
	if req.ClientID != "" {
		clientID = req.ClientID
	}

	// Get token from Keycloak
	tokenResp, err := s.keycloakClient.GetUserToken(req.Username, req.Password, clientID)
	if err != nil {
		s.logger.Error("System admin login failed", zap.Error(err), zap.String("username", req.Username))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Validate and parse token
	claims, err := s.keycloakValidator.ValidateTokenString(tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Check if user has system admin role
	if !claims.IsSystemAdmin() {
		s.logger.Warn("User without system admin role attempted system admin login",
			zap.String("username", req.Username),
			zap.Strings("roles", claims.RealmAccess.Roles))
		return nil, fmt.Errorf("insufficient permissions: system admin role required")
	}

	// Build user info
	userInfo := &UserInfo{
		ID:            claims.Subject,
		Username:      claims.PreferredUsername,
		Email:         claims.Email,
		FirstName:     claims.GivenName,
		LastName:      claims.FamilyName,
		Roles:         claims.RealmAccess.Roles,
		IsSystemAdmin: claims.IsSystemAdmin(),
		IsTenantAdmin: claims.IsTenantAdmin(),
	}

	// Determine redirect URL based on role
	redirectURL := s.getSystemAdminRedirectURL(claims)

	// Get permissions
	permissions := s.extractPermissions(claims)

	return &LoginResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpiresIn,
		User:         userInfo,
		RedirectURL:  redirectURL,
		Permissions:  permissions,
	}, nil
}

// TenantAdminLogin handles tenant admin login via tenant.zplus.io/admin
func (s *AuthService) TenantAdminLogin(ctx context.Context, req LoginRequest, tenantDomain string) (*LoginResponse, error) {
	s.logger.Info("Tenant admin login attempt",
		zap.String("username", req.Username),
		zap.String("tenant_domain", tenantDomain))

	// Use tenant frontend client ID for tenant admin login
	clientID := "zplus-tenant-frontend"
	if req.ClientID != "" {
		clientID = req.ClientID
	}

	// Get token from Keycloak
	tokenResp, err := s.keycloakClient.GetUserToken(req.Username, req.Password, clientID)
	if err != nil {
		s.logger.Error("Tenant admin login failed", zap.Error(err),
			zap.String("username", req.Username),
			zap.String("tenant_domain", tenantDomain))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Validate and parse token
	claims, err := s.keycloakValidator.ValidateTokenString(tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Check if user has admin role (system admin or tenant admin)
	if !claims.IsSystemAdmin() && !claims.IsTenantAdmin() {
		s.logger.Warn("User without admin role attempted tenant admin login",
			zap.String("username", req.Username),
			zap.String("tenant_domain", tenantDomain),
			zap.Strings("roles", claims.RealmAccess.Roles))
		return nil, fmt.Errorf("insufficient permissions: admin role required")
	}

	// For tenant admin, validate tenant access
	if claims.IsTenantAdmin() && !claims.IsSystemAdmin() {
		if claims.TenantDomain != tenantDomain {
			s.logger.Warn("Tenant admin attempted to access different tenant",
				zap.String("username", req.Username),
				zap.String("user_tenant", claims.TenantDomain),
				zap.String("requested_tenant", tenantDomain))
			return nil, fmt.Errorf("insufficient permissions: no access to this tenant")
		}
	}

	// Build user info
	userInfo := &UserInfo{
		ID:            claims.Subject,
		Username:      claims.PreferredUsername,
		Email:         claims.Email,
		FirstName:     claims.GivenName,
		LastName:      claims.FamilyName,
		TenantID:      claims.TenantID,
		TenantDomain:  claims.TenantDomain,
		Roles:         claims.RealmAccess.Roles,
		IsSystemAdmin: claims.IsSystemAdmin(),
		IsTenantAdmin: claims.IsTenantAdmin(),
	}

	// Determine redirect URL based on role and tenant
	redirectURL := s.getTenantAdminRedirectURL(claims, tenantDomain)

	// Get permissions
	permissions := s.extractPermissions(claims)

	return &LoginResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpiresIn,
		User:         userInfo,
		RedirectURL:  redirectURL,
		Permissions:  permissions,
	}, nil
}

// UserLogin handles regular user login via tenant.zplus.io
func (s *AuthService) UserLogin(ctx context.Context, req LoginRequest, tenantDomain string) (*LoginResponse, error) {
	s.logger.Info("User login attempt",
		zap.String("username", req.Username),
		zap.String("tenant_domain", tenantDomain))

	// Use tenant frontend client ID for user login
	clientID := "zplus-tenant-frontend"
	if req.ClientID != "" {
		clientID = req.ClientID
	}

	// Get token from Keycloak
	tokenResp, err := s.keycloakClient.GetUserToken(req.Username, req.Password, clientID)
	if err != nil {
		s.logger.Error("User login failed", zap.Error(err),
			zap.String("username", req.Username),
			zap.String("tenant_domain", tenantDomain))
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Validate and parse token
	claims, err := s.keycloakValidator.ValidateTokenString(tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Validate tenant access (except for system admin)
	if !claims.IsSystemAdmin() {
		if claims.TenantDomain != tenantDomain {
			s.logger.Warn("User attempted to access different tenant",
				zap.String("username", req.Username),
				zap.String("user_tenant", claims.TenantDomain),
				zap.String("requested_tenant", tenantDomain))
			return nil, fmt.Errorf("insufficient permissions: no access to this tenant")
		}
	}

	// Build user info
	userInfo := &UserInfo{
		ID:            claims.Subject,
		Username:      claims.PreferredUsername,
		Email:         claims.Email,
		FirstName:     claims.GivenName,
		LastName:      claims.FamilyName,
		TenantID:      claims.TenantID,
		TenantDomain:  claims.TenantDomain,
		Roles:         claims.RealmAccess.Roles,
		IsSystemAdmin: claims.IsSystemAdmin(),
		IsTenantAdmin: claims.IsTenantAdmin(),
	}

	// Determine redirect URL based on role and tenant
	redirectURL := s.getUserRedirectURL(claims, tenantDomain)

	// Get permissions
	permissions := s.extractPermissions(claims)

	return &LoginResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpiresIn,
		User:         userInfo,
		RedirectURL:  redirectURL,
		Permissions:  permissions,
	}, nil
}

// RefreshToken handles token refresh
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string, clientID string) (*LoginResponse, error) {
	s.logger.Debug("Token refresh attempt", zap.String("client_id", clientID))

	// Refresh token via Keycloak
	tokenResp, err := s.keycloakClient.RefreshToken(refreshToken, clientID)
	if err != nil {
		s.logger.Error("Token refresh failed", zap.Error(err))
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}

	// Validate and parse new token
	claims, err := s.keycloakValidator.ValidateTokenString(tokenResp.AccessToken)
	if err != nil {
		s.logger.Error("Refreshed token validation failed", zap.Error(err))
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Build user info
	userInfo := &UserInfo{
		ID:            claims.Subject,
		Username:      claims.PreferredUsername,
		Email:         claims.Email,
		FirstName:     claims.GivenName,
		LastName:      claims.FamilyName,
		TenantID:      claims.TenantID,
		TenantDomain:  claims.TenantDomain,
		Roles:         claims.RealmAccess.Roles,
		IsSystemAdmin: claims.IsSystemAdmin(),
		IsTenantAdmin: claims.IsTenantAdmin(),
	}

	// Get permissions
	permissions := s.extractPermissions(claims)

	return &LoginResponse{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    tokenResp.ExpiresIn,
		User:         userInfo,
		Permissions:  permissions,
	}, nil
}

// Logout handles user logout
func (s *AuthService) Logout(ctx context.Context, refreshToken string, clientID string) error {
	s.logger.Debug("Logout attempt", zap.String("client_id", clientID))

	// Logout via Keycloak
	err := s.keycloakClient.Logout(refreshToken, clientID)
	if err != nil {
		s.logger.Error("Logout failed", zap.Error(err))
		return fmt.Errorf("logout failed: %w", err)
	}

	s.logger.Info("User logged out successfully")
	return nil
}

// Helper methods

func (s *AuthService) getSystemAdminRedirectURL(claims *auth.TokenClaims) string {
	// System admin redirect to admin dashboard
	return "/admin/dashboard"
}

func (s *AuthService) getTenantAdminRedirectURL(claims *auth.TokenClaims, tenantDomain string) string {
	// Tenant admin redirect to tenant admin dashboard
	if claims.IsSystemAdmin() {
		// System admin accessing tenant admin
		return fmt.Sprintf("/tenant/%s/admin/dashboard", tenantDomain)
	}
	// Regular tenant admin
	return "/admin/dashboard"
}

func (s *AuthService) getUserRedirectURL(claims *auth.TokenClaims, tenantDomain string) string {
	// Check role and redirect accordingly
	if claims.IsSystemAdmin() {
		return "/admin/dashboard"
	}
	if claims.IsTenantAdmin() {
		return "/admin/dashboard"
	}
	// Regular user
	return "/dashboard"
}

func (s *AuthService) extractPermissions(claims *auth.TokenClaims) map[string]bool {
	permissions := make(map[string]bool)

	// System-level permissions
	if claims.IsSystemAdmin() {
		permissions["system:manage"] = true
		permissions["tenant:create"] = true
		permissions["tenant:manage"] = true
		permissions["user:manage_all"] = true
	}

	// Tenant-level permissions
	if claims.IsTenantAdmin() {
		permissions["tenant:manage_own"] = true
		permissions["user:manage_tenant"] = true
		permissions["domain:manage"] = true
	}

	// Extract from token attributes
	if len(claims.TenantPermissions) > 0 {
		// Parse JSON permissions from token
		tenantPerms, err := claims.GetTenantPermissions()
		if err == nil {
			// Merge tenant permissions
			for perm, value := range tenantPerms {
				permissions[perm] = value
			}
		}
	}

	// Default permissions for all users
	permissions["tenant:access"] = true
	permissions["profile:manage_own"] = true

	return permissions
}

// TokenInfo represents token information
type TokenInfo struct {
	Valid       bool              `json:"valid"`
	Claims      *auth.TokenClaims `json:"claims,omitempty"`
	User        *UserInfo         `json:"user,omitempty"`
	Permissions map[string]bool   `json:"permissions,omitempty"`
	Error       string            `json:"error,omitempty"`
}

// ValidateToken validates a token and returns token info
func (s *AuthService) ValidateToken(ctx context.Context, token string) *TokenInfo {
	claims, err := s.keycloakValidator.ValidateTokenString(token)
	if err != nil {
		return &TokenInfo{
			Valid: false,
			Error: err.Error(),
		}
	}

	userInfo := &UserInfo{
		ID:            claims.Subject,
		Username:      claims.PreferredUsername,
		Email:         claims.Email,
		FirstName:     claims.GivenName,
		LastName:      claims.FamilyName,
		TenantID:      claims.TenantID,
		TenantDomain:  claims.TenantDomain,
		Roles:         claims.RealmAccess.Roles,
		IsSystemAdmin: claims.IsSystemAdmin(),
		IsTenantAdmin: claims.IsTenantAdmin(),
	}

	permissions := s.extractPermissions(claims)

	return &TokenInfo{
		Valid:       true,
		Claims:      claims,
		User:        userInfo,
		Permissions: permissions,
	}
}
