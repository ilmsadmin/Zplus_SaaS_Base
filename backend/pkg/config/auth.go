package config

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Keycloak KeycloakConfig `json:"keycloak"`
}

// KeycloakConfig holds Keycloak configuration
type KeycloakConfig struct {
	URL             string `json:"url"`
	Realm           string `json:"realm"`
	BackendClientID string `json:"backend_client_id"`
	BackendSecret   string `json:"backend_secret"`
	AdminClientID   string `json:"admin_client_id"`
	AdminSecret     string `json:"admin_secret"`
	TenantClientID  string `json:"tenant_client_id"`
	TenantSecret    string `json:"tenant_secret"`
}

// LoadAuthConfig loads authentication configuration from environment variables
func LoadAuthConfig() AuthConfig {
	return AuthConfig{
		Keycloak: KeycloakConfig{
			URL:             getEnv("KEYCLOAK_URL", "http://localhost:8081"),
			Realm:           getEnv("KEYCLOAK_REALM", "zplus"),
			BackendClientID: getEnv("KEYCLOAK_BACKEND_CLIENT_ID", "zplus-backend"),
			BackendSecret:   getEnv("KEYCLOAK_BACKEND_SECRET", "zplus-backend-secret-2024"),
			AdminClientID:   getEnv("KEYCLOAK_ADMIN_CLIENT_ID", "zplus-admin-frontend"),
			AdminSecret:     getEnv("KEYCLOAK_ADMIN_SECRET", "zplus-admin-frontend-secret-2024"),
			TenantClientID:  getEnv("KEYCLOAK_TENANT_CLIENT_ID", "zplus-tenant-frontend"),
			TenantSecret:    getEnv("KEYCLOAK_TENANT_SECRET", "zplus-tenant-frontend-secret-2024"),
		},
	}
}
