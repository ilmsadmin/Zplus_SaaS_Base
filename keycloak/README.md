# Keycloak Authentication Setup

Keycloak được sử dụng làm Identity & Access Management (IAM) solution cho Zplus SaaS Base, hỗ trợ multi-tenant authentication và authorization.

## 🎯 Tổng quan

### Architecture
- **Single Realm**: `zplus` realm cho tất cả tenants
- **Multi-Client**: Mỗi frontend application có riêng client
- **Role-based Access Control**: System admin, tenant admin, tenant user
- **Tenant Isolation**: Mỗi user được gán tenant_id và permissions

### Roles & Permissions
- **system_admin**: Quản lý toàn hệ thống
- **tenant_admin**: Quản lý tenant riêng
- **tenant_user**: Sử dụng services của tenant

## 🚀 Quick Setup

### 1. Start Services
```bash
# Start all development services
make dev-up

# Or start only database services first
make db-up
```

### 2. Setup Keycloak
```bash
# Setup Keycloak realm và configuration
make keycloak-setup
```

### 3. Access Keycloak
```bash
# Open admin console
make keycloak-admin

# Or manually visit: http://localhost:8081
# Username: admin
# Password: admin123
```

## 🔧 Configuration

### Environment Variables
Thêm vào `backend/.env`:
```bash
KEYCLOAK_URL=http://keycloak:8080
KEYCLOAK_REALM=zplus
KEYCLOAK_BACKEND_CLIENT_ID=zplus-backend
KEYCLOAK_BACKEND_SECRET=zplus-backend-secret-2024
KEYCLOAK_ADMIN_CLIENT_ID=zplus-admin-frontend
KEYCLOAK_ADMIN_SECRET=zplus-admin-frontend-secret-2024
KEYCLOAK_TENANT_CLIENT_ID=zplus-tenant-frontend
KEYCLOAK_TENANT_SECRET=zplus-tenant-frontend-secret-2024
```

### Clients Configuration

#### 1. zplus-backend (Backend API)
- **Type**: Confidential
- **Protocol**: OpenID Connect
- **Valid Redirect URIs**: `http://localhost:8082/*`, `http://api.localhost/*`
- **Service Accounts**: Enabled
- **Direct Access Grants**: Enabled

#### 2. zplus-admin-frontend (System Admin)
- **Type**: Public (SPA)
- **Protocol**: OpenID Connect
- **Valid Redirect URIs**: `http://localhost:3000/*`, `http://admin.localhost/*`
- **PKCE**: Enabled

#### 3. zplus-tenant-frontend (Tenant App)
- **Type**: Public (SPA)
- **Protocol**: OpenID Connect
- **Valid Redirect URIs**: `http://localhost:3001/*`, `http://*.localhost/*`
- **PKCE**: Enabled

## 👤 Default Users

Setup script tự động tạo test users:

### System Administrator
- **Username**: `system.admin`
- **Email**: `admin@zplus.io`
- **Password**: `Admin123!`
- **Roles**: `system_admin`
- **Permissions**: Full system access

### Tenant Administrator
- **Username**: `tenant.admin`
- **Email**: `admin@acme.example.com`
- **Password**: `TenantAdmin123!`
- **Tenant**: `acme_corp`
- **Roles**: `tenant_admin`

### Tenant User
- **Username**: `john.doe`
- **Email**: `john.doe@acme.example.com`
- **Password**: `User123!`
- **Tenant**: `acme_corp`
- **Roles**: `tenant_user`

## 🔐 Integration với Go Backend

### JWT Token Validation
```go
package main

import (
    "github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/auth"
    "github.com/ilmsadmin/zplus-saas-base/internal/infrastructure/middleware"
    "github.com/ilmsadmin/zplus-saas-base/pkg/config"
)

func main() {
    // Load config
    authConfig := config.LoadAuthConfig()
    
    // Create validator
    validator := auth.NewKeycloakValidator(auth.KeycloakConfig{
        URL:      authConfig.Keycloak.URL,
        Realm:    authConfig.Keycloak.Realm,
        ClientID: authConfig.Keycloak.BackendClientID,
        Secret:   authConfig.Keycloak.BackendSecret,
    })
    
    // Use middleware
    app.Use(middleware.AuthMiddleware(validator, logger))
}
```

### Protected Routes
```go
// System admin only
app.Get("/admin/*", 
    middleware.AuthMiddleware(validator, logger),
    middleware.RequireSystemAdmin(),
    adminHandler,
)

// Tenant admin or higher
app.Get("/tenant/:tenant_id/admin/*",
    middleware.AuthMiddleware(validator, logger),
    middleware.RequireTenantAdmin(),
    middleware.TenantIsolationMiddleware(),
    tenantAdminHandler,
)

// Any authenticated user
app.Get("/tenant/:tenant_id/dashboard",
    middleware.AuthMiddleware(validator, logger),
    middleware.TenantIsolationMiddleware(),
    dashboardHandler,
)
```

### Get User Information
```go
func handler(c *fiber.Ctx) error {
    // Get user from context
    user, ok := middleware.GetUserFromContext(c)
    if !ok {
        return c.Status(401).JSON(fiber.Map{"error": "unauthenticated"})
    }
    
    // Check permissions
    if !user.CanManageTenant() {
        return c.Status(403).JSON(fiber.Map{"error": "forbidden"})
    }
    
    // Get tenant info
    tenantID, _ := middleware.GetTenantIDFromContext(c)
    
    return c.JSON(fiber.Map{
        "user_id": user.Subject,
        "username": user.PreferredUsername,
        "email": user.Email,
        "tenant_id": tenantID,
        "roles": user.RealmAccess.Roles,
    })
}
```

## 🔧 Troubleshooting

### 1. Keycloak không khởi động
```bash
# Check logs
make keycloak-logs

# Restart service
make keycloak-restart

# Check database connection
make db-status
```

### 2. Realm import failed
```bash
# Reset và re-import
make keycloak-restart
sleep 30
make keycloak-setup
```

### 3. Token validation errors
```bash
# Check JWK endpoint
curl http://localhost:8081/realms/zplus/protocol/openid-connect/certs

# Check realm configuration
curl http://localhost:8081/realms/zplus/.well-known/openid_configuration
```

### 4. CORS issues
Đảm bảo frontend origins được thêm vào client configuration:
- `http://localhost:3000` (admin frontend)
- `http://localhost:3001` (tenant frontend)
- `http://*.localhost` (wildcard subdomains)

## 📚 Tài liệu thêm

- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [OpenID Connect Specification](https://openid.net/connect/)
- [JWT Token Format](https://tools.ietf.org/html/rfc7519)

## 🔄 Development Workflow

### 1. Thêm user mới
```bash
# Sử dụng Keycloak Admin Console
make keycloak-admin

# Hoặc dùng Admin API thông qua Go client
```

### 2. Thêm client mới
1. Vào Admin Console
2. Tạo client mới với appropriate settings
3. Cập nhật environment variables
4. Restart backend service

### 3. Testing authentication
```bash
# Test realm endpoint
make auth-test

# Test với curl
curl -X POST http://localhost:8081/realms/zplus/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "username=john.doe" \
  -d "password=User123!" \
  -d "grant_type=password" \
  -d "client_id=zplus-backend" \
  -d "client_secret=zplus-backend-secret-2024"
```

## 📊 Monitoring

### Health Checks
- Keycloak: `http://localhost:8081/health/ready`
- Realm: `http://localhost:8081/realms/zplus`
- Metrics: `http://localhost:8081/metrics`

### Logs
```bash
# Real-time logs
make keycloak-logs

# Container logs
docker logs zplus-keycloak-dev -f
```
