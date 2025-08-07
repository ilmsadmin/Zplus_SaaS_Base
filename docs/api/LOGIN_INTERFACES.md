# Login Interfaces Implementation Guide

## Overview

Đã triển khai thành công hệ thống Login Interfaces cho Zplus SaaS Base với hỗ trợ multi-tenant và role-based access control. Hệ thống bao gồm 3 loại login interface chính:

1. **System Admin Login** (`admin.zplus.io`)
2. **Tenant Admin Login** (`tenant.zplus.io/admin`)
3. **User Login** (`tenant.zplus.io`)

## Architecture

### 1. Authentication Flow

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   Keycloak      │
│                 │    │                 │    │                 │
├─────────────────┤    ├─────────────────┤    ├─────────────────┤
│ Login Interface │───▶│ Auth Handler    │───▶│ Token Endpoint  │
│                 │    │                 │    │                 │
│ Role Detection  │◀───│ JWT Validation  │◀───│ JWT Signing     │
│                 │    │                 │    │                 │
│ Auto Redirect   │◀───│ Role-based      │    │ User Claims     │
│                 │    │ Authorization   │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 2. Components Structure

```
backend/
├── internal/
│   ├── application/
│   │   └── auth_service.go          # Business logic cho authentication
│   ├── infrastructure/
│   │   ├── auth/
│   │   │   └── keycloak.go         # Keycloak integration (updated)
│   │   └── middleware/
│   │       └── auth.go             # Authentication middleware
│   └── interfaces/
│       ├── auth_handler.go         # HTTP handlers cho auth endpoints
│       └── auth_routes.go          # Route definitions
└── scripts/
    └── test-login-interfaces.sh    # Testing script
```

## Features Implemented

### 1. Authentication Endpoints

| Endpoint | Method | Description | Client ID |
|----------|--------|-------------|-----------|
| `/auth/system-admin/login` | POST | System admin login | `zplus-admin-frontend` |
| `/auth/tenant-admin/login` | POST | Tenant admin login | `zplus-tenant-frontend` |
| `/auth/user/login` | POST | Regular user login | `zplus-tenant-frontend` |
| `/auth/refresh` | POST | Token refresh | Auto-detected |
| `/auth/logout` | POST | User logout | Auto-detected |
| `/auth/validate` | GET | Token validation | N/A |
| `/auth/profile` | GET | User profile | N/A |

### 2. Login Interface Pages

| URL Pattern | Description | Target Users |
|-------------|-------------|--------------|
| `/admin/login` | System admin login page | System administrators |
| `/admin/login.html` | HTML login form | System administrators |
| `/tenant/:slug/admin/login` | Tenant admin login API | Tenant administrators |
| `/tenant/:slug/admin/login.html` | Tenant admin HTML form | Tenant administrators |
| `/tenant/:slug/login` | User login API | Regular users |
| `/tenant/:slug/login.html` | User HTML form | Regular users |

### 3. Role-based Redirects

| Role | Login Success Redirect | Default Dashboard |
|------|----------------------|-------------------|
| System Admin | `/admin/dashboard` | System management |
| Tenant Admin | `/tenant/:slug/admin/dashboard` | Tenant management |
| User | `/tenant/:slug/dashboard` | User services |

### 4. Protected Routes

#### System Admin Routes
- `/admin/dashboard` - System admin dashboard
- `/admin/tenants` - Tenant management
- `/admin/users` - Global user management
- `/api/v1/admin/*` - Admin API endpoints

#### Tenant Admin Routes
- `/tenant/:slug/admin/dashboard` - Tenant admin dashboard
- `/tenant/:slug/admin/users` - Tenant user management
- `/tenant/:slug/admin/settings` - Tenant settings
- `/tenant/:slug/admin/domains` - Custom domain management

#### User Routes
- `/tenant/:slug/dashboard` - User dashboard
- `/tenant/:slug/profile` - User profile
- `/tenant/:slug/modules/*` - Module access (files, POS, etc.)

## Security Features

### 1. JWT Token Management

- **Access Tokens**: Short-lived (configurable expiry)
- **Refresh Tokens**: Long-lived (7 days default)
- **HTTP-Only Cookies**: Secure token storage
- **CORS Protection**: Proper origin validation

### 2. Role-based Access Control

```go
// System Admin only
middleware.RequireSystemAdmin()

// Tenant Admin or higher
middleware.RequireTenantAdmin()

// Tenant isolation
middleware.TenantIsolationMiddleware()

// Permission-based access
middleware.RequireCasbinPermission("resource", "action", casbinService)
```

### 3. Multi-tenant Isolation

- **Tenant Context**: Extracted from subdomain/domain
- **Tenant Validation**: Users can only access their assigned tenant
- **Cross-tenant Protection**: System admins have special privileges

## API Examples

### 1. System Admin Login

```bash
curl -X POST http://localhost:8082/auth/system-admin/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "system.admin",
    "password": "Admin123!"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "user": {
    "id": "uuid",
    "username": "system.admin",
    "email": "admin@zplus.io",
    "roles": ["system_admin"],
    "is_system_admin": true,
    "is_tenant_admin": false
  },
  "redirect_url": "/admin/dashboard",
  "permissions": {
    "system:manage": true,
    "tenant:create": true,
    "tenant:manage": true,
    "user:manage_all": true
  }
}
```

### 2. Tenant Admin Login

```bash
curl -X POST http://localhost:8082/auth/tenant-admin/login \
  -H "Content-Type: application/json" \
  -H "Host: acme.zplus.io" \
  -d '{
    "username": "tenant.admin",
    "password": "TenantAdmin123!"
  }'
```

### 3. User Login

```bash
curl -X POST http://localhost:8082/auth/user/login \
  -H "Content-Type: application/json" \
  -H "Host: acme.zplus.io" \
  -d '{
    "username": "john.doe",
    "password": "User123!"
  }'
```

### 4. Token Validation

```bash
curl -X GET http://localhost:8082/auth/validate \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 5. Protected Route Access

```bash
# System admin accessing admin dashboard
curl -X GET http://localhost:8082/admin/dashboard \
  -H "Authorization: Bearer SYSTEM_ADMIN_TOKEN"

# Tenant admin accessing tenant management
curl -X GET http://localhost:8082/tenant/acme/admin/dashboard \
  -H "Authorization: Bearer TENANT_ADMIN_TOKEN"

# User accessing user dashboard
curl -X GET http://localhost:8082/tenant/acme/dashboard \
  -H "Authorization: Bearer USER_TOKEN"
```

## Configuration

### 1. Environment Variables

```bash
# Keycloak Configuration
KEYCLOAK_URL=http://localhost:8081
KEYCLOAK_REALM=zplus
KEYCLOAK_BACKEND_CLIENT_ID=zplus-backend
KEYCLOAK_BACKEND_SECRET=zplus-backend-secret-2024
KEYCLOAK_ADMIN_CLIENT_ID=zplus-admin-frontend
KEYCLOAK_ADMIN_SECRET=zplus-admin-frontend-secret-2024
KEYCLOAK_TENANT_CLIENT_ID=zplus-tenant-frontend
KEYCLOAK_TENANT_SECRET=zplus-tenant-frontend-secret-2024
```

### 2. Client Configurations

| Client ID | Type | Purpose | Redirect URIs |
|-----------|------|---------|---------------|
| `zplus-backend` | Confidential | Backend API | N/A |
| `zplus-admin-frontend` | Public (SPA) | System admin frontend | `http://admin.localhost/*` |
| `zplus-tenant-frontend` | Public (SPA) | Tenant frontend | `http://*.localhost/*` |

## Testing

### 1. Automated Testing

```bash
# Run all login interface tests
make test-login

# Or run directly
./backend/scripts/test-login-interfaces.sh
```

### 2. Manual Testing

1. **Start Services**:
   ```bash
   make dev-up          # Start all services
   make keycloak-setup  # Configure Keycloak
   ```

2. **Test Login Pages**:
   - System Admin: http://localhost:8082/admin/login.html
   - Tenant Admin: http://localhost:8082/tenant/acme/admin/login.html
   - User: http://localhost:8082/tenant/acme/login.html

3. **Test Credentials**:
   - System Admin: `system.admin` / `Admin123!`
   - Tenant Admin: `tenant.admin` / `TenantAdmin123!`
   - User: `john.doe` / `User123!`

### 3. API Testing with curl

```bash
# Test all endpoints
./backend/scripts/test-login-interfaces.sh

# Test specific login
curl -X POST http://localhost:8082/auth/system-admin/login \
  -H "Content-Type: application/json" \
  -d '{"username":"system.admin","password":"Admin123!"}'
```

## Error Handling

### 1. Common Error Responses

```json
// Invalid credentials
{
  "error": "login_failed",
  "message": "Authentication failed"
}

// Missing permissions
{
  "error": "insufficient_permissions",
  "message": "Required role not found"
}

// Invalid token
{
  "error": "invalid_token",
  "message": "Token validation failed"
}

// Missing tenant context
{
  "error": "missing_tenant",
  "message": "Tenant domain not found"
}
```

### 2. HTTP Status Codes

| Status | Meaning | When |
|--------|---------|------|
| 200 | Success | Login successful, token valid |
| 400 | Bad Request | Invalid request body, missing fields |
| 401 | Unauthorized | Invalid credentials, missing/invalid token |
| 403 | Forbidden | Insufficient permissions |
| 500 | Internal Error | Server error, configuration issue |

## Future Enhancements

### 1. Planned Features

- [ ] **Frontend Integration**: React/Next.js components
- [ ] **OAuth2 Flows**: Authorization Code Flow for SPAs
- [ ] **Social Login**: Google, GitHub integration
- [ ] **MFA Support**: Two-factor authentication
- [ ] **Session Management**: Advanced session handling
- [ ] **Audit Logging**: Login/logout tracking

### 2. Security Improvements

- [ ] **Rate Limiting**: Login attempt protection
- [ ] **Brute Force Protection**: Account lockout
- [ ] **IP Whitelisting**: Admin access restriction
- [ ] **Device Management**: Trusted device tracking

### 3. UX Enhancements

- [ ] **Remember Me**: Persistent login
- [ ] **Password Reset**: Self-service password reset
- [ ] **Account Recovery**: Email-based recovery
- [ ] **Theme Support**: Tenant-specific branding

## Troubleshooting

### 1. Common Issues

**Issue**: Login fails with "Authentication failed"
- **Cause**: Keycloak not configured or user doesn't exist
- **Solution**: Run `make keycloak-setup` and verify test users

**Issue**: "insufficient_permissions" error
- **Cause**: User role doesn't match required permissions
- **Solution**: Check user roles in Keycloak admin console

**Issue**: "missing_tenant" error  
- **Cause**: Tenant middleware not extracting domain correctly
- **Solution**: Verify Host header or X-Tenant-Domain header

### 2. Debug Commands

```bash
# Check Keycloak health
curl http://localhost:8081/health/ready

# Validate token manually
curl http://localhost:8081/realms/zplus/protocol/openid-connect/userinfo \
  -H "Authorization: Bearer YOUR_TOKEN"

# Check backend logs
make logs

# Test authentication
make auth-test
```

## Integration Points

### 1. Frontend Integration

```javascript
// React/Next.js example
const login = async (credentials) => {
  const response = await fetch('/auth/system-admin/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(credentials)
  });
  
  const result = await response.json();
  if (result.success) {
    // Redirect to dashboard
    window.location.href = result.redirect_url;
  }
};
```

### 2. API Gateway Integration

```yaml
# Traefik rules example
http:
  routers:
    admin:
      rule: "Host(`admin.zplus.io`)"
      service: backend
    tenant:
      rule: "Host(`{subdomain:[a-z0-9-]+}.zplus.io`)"
      service: backend
```

This implementation provides a robust, secure, and scalable authentication system that supports the multi-tenant architecture requirements of Zplus SaaS Base.
