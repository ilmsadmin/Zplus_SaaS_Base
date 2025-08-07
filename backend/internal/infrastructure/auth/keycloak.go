package auth

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

// KeycloakConfig holds the configuration for Keycloak integration
type KeycloakConfig struct {
	URL      string
	Realm    string
	ClientID string
	Secret   string
}

// TokenClaims represents the claims in a Keycloak JWT token
type TokenClaims struct {
	jwt.RegisteredClaims
	Subject           string                `json:"sub"`
	RealmAccess       RealmAccess           `json:"realm_access"`
	ResourceAccess    map[string]ClientRole `json:"resource_access"`
	PreferredUsername string                `json:"preferred_username"`
	Email             string                `json:"email"`
	EmailVerified     bool                  `json:"email_verified"`
	Name              string                `json:"name"`
	GivenName         string                `json:"given_name"`
	FamilyName        string                `json:"family_name"`
	TenantID          string                `json:"tenant_id"`
	TenantDomain      string                `json:"tenant_domain"`
	TenantPermissions json.RawMessage       `json:"tenant_permissions"`
}

// RealmAccess represents realm-level roles
type RealmAccess struct {
	Roles []string `json:"roles"`
}

// ClientRole represents client-specific roles
type ClientRole struct {
	Roles []string `json:"roles"`
}

// TenantPermissions represents tenant-specific permissions
type TenantPermissions map[string]bool

// KeycloakValidator handles JWT token validation with Keycloak
type KeycloakValidator struct {
	config   KeycloakConfig
	jwkSet   jwk.Set
	lastSync time.Time
}

// NewKeycloakValidator creates a new Keycloak JWT validator
func NewKeycloakValidator(config KeycloakConfig) *KeycloakValidator {
	return &KeycloakValidator{
		config: config,
	}
}

// ValidateToken validates a JWT token against Keycloak
func (kv *KeycloakValidator) ValidateToken(tokenString string) (*TokenClaims, error) {
	// Remove Bearer prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	// Sync JWK set if needed
	if err := kv.syncJWKSet(); err != nil {
		return nil, fmt.Errorf("failed to sync JWK set: %w", err)
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		// Find the key in JWK set
		key, found := kv.jwkSet.LookupKeyID(kid)
		if !found {
			return nil, fmt.Errorf("key with kid %s not found", kid)
		}

		// Convert to RSA public key
		var rsaKey rsa.PublicKey
		if err := key.Raw(&rsaKey); err != nil {
			return nil, fmt.Errorf("failed to convert key to RSA: %w", err)
		}

		return &rsaKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(*TokenClaims)
	if !ok {
		return nil, fmt.Errorf("failed to parse claims")
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", kv.config.URL, kv.config.Realm)
	if claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	return claims, nil
}

// ValidateTokenString validates a JWT token string (alias for ValidateToken)
func (kv *KeycloakValidator) ValidateTokenString(tokenString string) (*TokenClaims, error) {
	return kv.ValidateToken(tokenString)
}

// syncJWKSet fetches the latest JWK set from Keycloak
func (kv *KeycloakValidator) syncJWKSet() error {
	// Sync every 5 minutes or if not synced yet
	if time.Since(kv.lastSync) < 5*time.Minute && kv.jwkSet != nil {
		return nil
	}

	jwkURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", kv.config.URL, kv.config.Realm)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	set, err := jwk.Fetch(ctx, jwkURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWK set: %w", err)
	}

	kv.jwkSet = set
	kv.lastSync = time.Now()

	return nil
}

// HasRole checks if the token has a specific realm role
func (tc *TokenClaims) HasRole(role string) bool {
	for _, r := range tc.RealmAccess.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasClientRole checks if the token has a specific client role
func (tc *TokenClaims) HasClientRole(clientID, role string) bool {
	if clientRoles, exists := tc.ResourceAccess[clientID]; exists {
		for _, r := range clientRoles.Roles {
			if r == role {
				return true
			}
		}
	}
	return false
}

// GetTenantPermissions parses and returns tenant permissions
func (tc *TokenClaims) GetTenantPermissions() (TenantPermissions, error) {
	if len(tc.TenantPermissions) == 0 {
		return TenantPermissions{}, nil
	}

	var permissions TenantPermissions
	if err := json.Unmarshal(tc.TenantPermissions, &permissions); err != nil {
		return nil, fmt.Errorf("failed to parse tenant permissions: %w", err)
	}

	return permissions, nil
}

// HasPermission checks if the user has a specific tenant permission
func (tc *TokenClaims) HasPermission(permission string) bool {
	permissions, err := tc.GetTenantPermissions()
	if err != nil {
		return false
	}

	return permissions[permission]
}

// IsSystemAdmin checks if the user is a system administrator
func (tc *TokenClaims) IsSystemAdmin() bool {
	return tc.HasRole("system_admin")
}

// IsTenantAdmin checks if the user is a tenant administrator
func (tc *TokenClaims) IsTenantAdmin() bool {
	return tc.HasRole("tenant_admin")
}

// IsTenantUser checks if the user is a tenant user
func (tc *TokenClaims) IsTenantUser() bool {
	return tc.HasRole("tenant_user")
}

// CanManageSystem checks if the user can manage the system
func (tc *TokenClaims) CanManageSystem() bool {
	return tc.IsSystemAdmin() && tc.HasPermission("system:manage")
}

// CanManageTenant checks if the user can manage their tenant
func (tc *TokenClaims) CanManageTenant() bool {
	return (tc.IsSystemAdmin() && tc.HasPermission("tenant:manage")) ||
		(tc.IsTenantAdmin() && tc.HasPermission("tenant:manage_own"))
}

// CanCreateTenant checks if the user can create tenants
func (tc *TokenClaims) CanCreateTenant() bool {
	return tc.IsSystemAdmin() && tc.HasPermission("tenant:create")
}

// CanManageUsers checks if the user can manage users
func (tc *TokenClaims) CanManageUsers() bool {
	return (tc.IsSystemAdmin() && tc.HasPermission("user:manage_all")) ||
		(tc.IsTenantAdmin() && tc.HasPermission("user:manage_tenant"))
}

// KeycloakClient provides methods to interact with Keycloak Admin API
type KeycloakClient struct {
	config      KeycloakConfig
	httpClient  *http.Client
	accessToken string
	tokenExpiry time.Time
}

// NewKeycloakClient creates a new Keycloak Admin API client
func NewKeycloakClient(config KeycloakConfig) *KeycloakClient {
	return &KeycloakClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TokenResponse represents a token response from Keycloak
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// getAdminToken obtains an admin access token for API calls
func (kc *KeycloakClient) getAdminToken() error {
	if time.Now().Before(kc.tokenExpiry) && kc.accessToken != "" {
		return nil
	}

	// Implementation would go here to get admin token
	// This is a placeholder for the actual implementation
	return nil
}

// GetUserToken gets a user access token using username/password
func (kc *KeycloakClient) GetUserToken(username, password, clientID string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.config.URL, kc.config.Realm)

	// Create form data
	data := fmt.Sprintf("username=%s&password=%s&grant_type=password&client_id=%s", username, password, clientID)

	// Add client secret if configured
	if kc.config.Secret != "" {
		data += fmt.Sprintf("&client_secret=%s", kc.config.Secret)
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

// RefreshToken refreshes an access token using refresh token
func (kc *KeycloakClient) RefreshToken(refreshToken, clientID string) (*TokenResponse, error) {
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", kc.config.URL, kc.config.Realm)

	// Create form data
	data := fmt.Sprintf("refresh_token=%s&grant_type=refresh_token&client_id=%s", refreshToken, clientID)

	// Add client secret if configured
	if kc.config.Secret != "" {
		data += fmt.Sprintf("&client_secret=%s", kc.config.Secret)
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &tokenResp, nil
}

// Logout logs out a user by invalidating the refresh token
func (kc *KeycloakClient) Logout(refreshToken, clientID string) error {
	logoutURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/logout", kc.config.URL, kc.config.Realm)

	// Create form data
	data := fmt.Sprintf("refresh_token=%s&client_id=%s", refreshToken, clientID)

	// Add client secret if configured
	if kc.config.Secret != "" {
		data += fmt.Sprintf("&client_secret=%s", kc.config.Secret)
	}

	req, err := http.NewRequest("POST", logoutURL, strings.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := kc.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("logout failed with status: %d", resp.StatusCode)
	}

	return nil
}

// CreateUser creates a new user in Keycloak
func (kc *KeycloakClient) CreateUser(user KeycloakUser) error {
	if err := kc.getAdminToken(); err != nil {
		return err
	}

	// Implementation would go here
	return nil
}

// KeycloakUser represents a user in Keycloak
type KeycloakUser struct {
	Username      string            `json:"username"`
	Email         string            `json:"email"`
	FirstName     string            `json:"firstName"`
	LastName      string            `json:"lastName"`
	Enabled       bool              `json:"enabled"`
	EmailVerified bool              `json:"emailVerified"`
	Attributes    map[string]string `json:"attributes"`
	Groups        []string          `json:"groups"`
	RealmRoles    []string          `json:"realmRoles"`
}
