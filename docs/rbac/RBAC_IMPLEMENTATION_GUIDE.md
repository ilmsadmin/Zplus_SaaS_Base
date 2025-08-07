# Role-Based Access Control (RBAC) Implementation Guide

## Overview

Hệ thống RBAC đã được triển khai đầy đủ cho Zplus SaaS Base, cung cấp khả năng kiểm soát truy cập phân cấp và cô lập tenant.

## Kiến trúc RBAC

### 1. Hierarchy của Roles

```
System Level:
├── system_admin     (Quản trị hệ thống)
└── system_manager   (Quản lý hệ thống)

Tenant Level:
├── tenant_admin     (Quản trị tenant)
├── tenant_manager   (Quản lý tenant)
├── user             (Người dùng)
└── viewer           (Chỉ xem)
```

### 2. Permissions Structure

Permissions được định nghĩa theo format: `scope:action_resource`

**System Permissions:**
- `system:manage_tenants` - Quản lý các tenant
- `system:manage_users` - Quản lý người dùng hệ thống
- `system:view_audit_logs` - Xem log audit
- `system:manage_settings` - Quản lý cài đặt hệ thống

**Tenant Permissions:**
- `tenant:manage_users` - Quản lý người dùng tenant
- `tenant:manage_roles` - Quản lý vai trò tenant
- `tenant:manage_settings` - Quản lý cài đặt tenant
- `tenant:manage_domains` - Quản lý domain
- `tenant:view_audit_logs` - Xem log audit tenant
- `tenant:manage_api_keys` - Quản lý API keys

**User Permissions:**
- `user:read_profile` - Đọc thông tin cá nhân
- `user:update_profile` - Cập nhật thông tin cá nhân
- `user:manage_files` - Quản lý file
- `user:view_files` - Xem file

## Setup và Migration

### 1. Chạy Migration

```bash
cd backend
./scripts/run_rbac_migration.sh
```

### 2. Cài đặt Dependencies

Đảm bảo đã có trong `go.mod`:

```go
require (
    github.com/casbin/casbin/v2 v2.116.0
    github.com/casbin/gorm-adapter/v3 v3.36.0
)
```

### 3. Initialize trong Application

```go
// Trong main.go hoặc setup function
casbinService, err := auth.NewCasbinService(db, roleRepo, permissionRepo, userRoleRepo)
if err != nil {
    log.Fatal("Failed to initialize Casbin service:", err)
}
```

## Sử dụng RBAC

### 1. Middleware Usage

```go
// Require specific permission
app.Get("/api/users", 
    middleware.RequireCasbinPermission("user", "read", casbinService),
    userHandler.ListUsers,
)

// Require system admin role
app.Get("/api/admin/tenants",
    middleware.SystemAdminOnlyMiddleware(casbinService),
    adminHandler.ListTenants,
)

// Require tenant admin role
app.Get("/api/tenant/settings",
    middleware.AdminOnlyMiddleware(casbinService),
    tenantHandler.GetSettings,
)

// Tenant isolation
app.Use("/api/tenant/:tenantId",
    middleware.TenantIsolationEnhanced(casbinService),
)
```

### 2. Service Layer Usage

```go
// Assign role to user
err := roleService.AssignRoleToUser(ctx, application.AssignRoleInput{
    UserID:   userID,
    RoleID:   roleID,
    TenantID: tenantID,
})

// Check permissions
hasPermission, err := roleService.HasPermission(ctx, userID, "user", "read", tenantID)

// Get user permissions
permissions, err := roleService.GetUserPermissions(ctx, userID, tenantID)
```

### 3. GraphQL Usage

```graphql
# Create a new role
mutation CreateRole($input: CreateRoleInput!) {
  createRole(input: $input) {
    id
    name
    description
    permissions {
      name
      description
    }
  }
}

# Assign role to user
mutation AssignRole($input: AssignRoleInput!) {
  assignRole(input: $input) {
    id
    user {
      email
    }
    role {
      name
    }
    tenant {
      name
    }
  }
}

# Check user permissions
query CheckPermission($userId: UUID!, $resource: String!, $action: String!, $tenantId: UUID!) {
  hasPermission(userId: $userId, resource: $resource, action: $action, tenantId: $tenantId)
}
```

## Tenant Isolation

### 1. Automatic Tenant Detection

```go
// Middleware tự động detect tenant từ:
// - Subdomain: tenant.zplus.io
// - Custom domain: app.acme.com
// - Header: X-Tenant-ID
```

### 2. Database Isolation

```go
// Tất cả queries được tự động filter theo tenant_id
// Middleware sẽ inject tenant context vào request
```

### 3. Cross-Tenant Protection

```go
// Middleware ngăn chặn truy cập cross-tenant
app.Use(middleware.TenantIsolationEnhanced(casbinService))
```

## Login Interfaces

### 1. System Admin Login

- **URL**: `admin.zplus.io/login`
- **Roles**: `system_admin`, `system_manager`
- **Access**: Toàn bộ hệ thống

### 2. Tenant Admin Login

- **URL**: `tenant.zplus.io/admin/login`
- **Roles**: `tenant_admin`, `tenant_manager`
- **Access**: Quản lý tenant cụ thể

### 3. User Login

- **URL**: `tenant.zplus.io/login`
- **Roles**: `user`, `viewer`
- **Access**: Sử dụng dịch vụ của tenant

## Best Practices

### 1. Role Assignment

```go
// Luôn assign role trong context của tenant
err := roleService.AssignRoleToUser(ctx, application.AssignRoleInput{
    UserID:   userID,
    RoleID:   roleID,
    TenantID: tenantID, // Required cho tenant isolation
})
```

### 2. Permission Checks

```go
// Check permission trước khi thực hiện action
if hasPermission, _ := roleService.HasPermission(ctx, userID, "user", "delete", tenantID); !hasPermission {
    return fiber.ErrForbidden
}
```

### 3. Error Handling

```go
// Consistent error messages cho security
return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
    "error": "insufficient_permissions",
    "message": "Access denied",
})
```

## Testing

### 1. Unit Tests

```go
func TestRoleAssignment(t *testing.T) {
    // Test role assignment trong tenant context
    err := roleService.AssignRoleToUser(ctx, assignInput)
    assert.NoError(t, err)
    
    // Verify permission
    hasPermission, err := casbinService.Enforce(userID, "user", "read", tenantID)
    assert.True(t, hasPermission)
}
```

### 2. Integration Tests

```go
func TestTenantIsolation(t *testing.T) {
    // Test user không thể access data của tenant khác
    req := httptest.NewRequest("GET", "/api/tenant/"+otherTenantID+"/users", nil)
    resp := app.Test(req)
    assert.Equal(t, 403, resp.StatusCode)
}
```

## Monitoring và Audit

### 1. Audit Logs

Tất cả actions đều được log với tenant context:

```json
{
  "user_id": "uuid",
  "tenant_id": "uuid", 
  "action": "assign_role",
  "resource": "user_role",
  "details": {
    "role_name": "tenant_admin",
    "target_user": "uuid"
  }
}
```

### 2. Metrics

- Role assignment count per tenant
- Permission check frequency
- Failed authorization attempts

## Security Considerations

### 1. Principle of Least Privilege

- Users chỉ có permissions tối thiểu cần thiết
- Roles được thiết kế hierarchical
- Regular audit của permissions

### 2. Tenant Isolation

- Hoàn toàn cô lập data giữa các tenant
- Không thể cross-tenant access
- Database level separation

### 3. System Admin Protection

- System admin roles không thể bị delete
- Require special procedures để modify system roles
- Multiple admin accounts recommendation

## Troubleshooting

### 1. Common Issues

**Permission Denied Errors:**
```bash
# Check user roles
SELECT * FROM user_roles WHERE user_id = 'uuid';

# Check role permissions  
SELECT p.name FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = 'uuid';
```

**Tenant Access Issues:**
```bash
# Verify tenant exists and is active
SELECT * FROM tenants WHERE id = 'uuid' AND status = 'active';

# Check domain mapping
SELECT * FROM tenant_domains WHERE domain = 'example.com';
```

### 2. Debug Commands

```bash
# Check Casbin policies
SELECT * FROM casbin_rule WHERE ptype = 'p';

# Verify role assignments
SELECT ur.*, u.email, r.name FROM user_roles ur
JOIN users u ON ur.user_id = u.id  
JOIN roles r ON ur.role_id = r.id
WHERE ur.tenant_id = 'uuid';
```

## Migration từ hệ thống cũ

### 1. Migration Script

```sql
-- Migrate existing roles to new RBAC system
INSERT INTO user_roles (user_id, role_id, tenant_id, status)
SELECT 
    tu.user_id,
    r.id as role_id,
    tu.tenant_id,
    'active'
FROM tenant_users tu
JOIN roles r ON r.name = tu.role
WHERE tu.status = 'active';
```

### 2. Validation

```bash
# Verify migration success
./scripts/validate_rbac_migration.sh
```

## Performance Optimization

### 1. Database Indexes

```sql
-- Key indexes đã được tạo trong migration
CREATE INDEX idx_user_roles_user_tenant ON user_roles(user_id, tenant_id);
CREATE INDEX idx_casbin_rule_subject ON casbin_rule(v0);
```

### 2. Caching Strategy

```go
// Cache user permissions trong Redis
// Cache role definitions
// TTL: 15 minutes cho permissions
```

## Roadmap

### Next Features
- [ ] Dynamic permission creation
- [ ] Role templates
- [ ] Permission inheritance
- [ ] Bulk role management
- [ ] Advanced audit analytics
- [ ] Role-based UI rendering

### Improvements
- [ ] GraphQL subscriptions for role changes
- [ ] REST API compatibility layer  
- [ ] Advanced tenant isolation strategies
- [ ] Performance optimizations
- [ ] Enhanced security features
