# API Documentation

## Overview

Zplus SaaS Base cung cấp APIs thông qua GraphQL Federation pattern với multiple microservices. Tất cả APIs đều support multi-tenant architecture với subdomain và custom domain, require authentication theo role (system admin, tenant admin, user).

## Table of Contents

- [Authentication](#authentication)
- [Domain & Tenant Management](#domain--tenant-management)
- [GraphQL Endpoints](#graphql-endpoints)
- [REST Endpoints](#rest-endpoints)
- [Rate Limiting](#rate-limiting)
- [Error Handling](#error-handling)
- [Examples](#examples)

## Authentication

### JWT Token Structure

```json
{
  "sub": "user123",
  "tenant_id": "tenant1", 
  "roles": ["tenant_admin", "user"],
  "scope": "tenant1_scope",
  "permissions": ["read", "write"],
  "exp": 1640995200,
  "iat": 1640908800
}
```

### Headers Required

```http
Authorization: Bearer <jwt_token>
X-Tenant-ID: tenant1
Host: tenant1.zplus.io
Content-Type: application/json
```

### Role-based Access

- **System Admin**: Access via `admin.zplus.io`
- **Tenant Admin**: Access via `tenant.zplus.io/admin`  
- **User**: Access via `tenant.zplus.io`

## Domain & Tenant Management

### Tenant Domain Configuration

API endpoints để quản lý subdomain và custom domain:

```http
GET /api/v1/tenants/{tenant_id}/domains
POST /api/v1/tenants/{tenant_id}/domains
DELETE /api/v1/tenants/{tenant_id}/domains/{domain_id}
```

### Custom Domain Setup

1. Tenant admin thêm custom domain qua API
2. Hệ thống validate domain ownership
3. Automatic SSL certificate generation
4. DNS CNAME record verification

## GraphQL Endpoints

### Main Gateway
- **URL**: `https://api.zplus.io/graphql`
- **Tenant-specific**: `https://{tenant}.zplus.io/graphql`
- **Custom domain**: `https://{custom-domain}/graphql`
- **Playground**: Available in development only

### Service-specific Endpoints
- **User Service**: `https://api.zplus.io/user/graphql`
- **Tenant Service**: `https://api.zplus.io/tenant/graphql`
- **File Service**: `https://api.zplus.io/file/graphql` 
- **POS Service**: `https://api.zplus.io/pos/graphql`

### Admin Endpoints
- **System Admin**: `https://admin.zplus.io/graphql`

## Schema Documentation

### User Schema
See [User API](./user.md)

### File Schema  
See [File API](./file.md)

### POS Schema
See [POS API](./pos.md)

### Tenant Schema
See [Tenant API](./tenant.md)

## REST Endpoints

### Health Check
```http
GET /health
```

Response:
```json
{
  "status": "ok",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### File Upload
```http
POST /api/v1/files/upload
Content-Type: multipart/form-data
```

### Webhook Endpoints
```http
POST /api/v1/webhooks/payment
POST /api/v1/webhooks/keycloak
```

## Rate Limiting

- **Per User**: 1000 requests/hour
- **Per Tenant**: 10000 requests/hour
- **Per IP**: 100 requests/minute

Headers returned:
```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## Error Handling

### GraphQL Errors
```json
{
  "errors": [
    {
      "message": "User not found",
      "locations": [{"line": 2, "column": 3}],
      "path": ["user"],
      "extensions": {
        "code": "USER_NOT_FOUND",
        "tenant_id": "tenant1"
      }
    }
  ],
  "data": null
}
```

### REST Errors
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid email format",
    "details": {
      "field": "email",
      "value": "invalid-email"
    }
  }
}
```

### Common Error Codes
- `UNAUTHENTICATED`: Missing or invalid token
- `UNAUTHORIZED`: Insufficient permissions
- `TENANT_NOT_FOUND`: Invalid tenant
- `VALIDATION_ERROR`: Input validation failed
- `RATE_LIMIT_EXCEEDED`: Too many requests

## Examples

### GraphQL Query Example
```graphql
query GetUsers($tenantId: String!) {
  users(tenantId: $tenantId) {
    id
    name
    email
    createdAt
  }
}
```

### GraphQL Mutation Example
```graphql
mutation CreateUser($input: CreateUserInput!) {
  createUser(input: $input) {
    id
    name
    email
  }
}
```

### cURL Examples
```bash
# GraphQL Query
curl -X POST https://api.zplus.io/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: tenant1" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { users { id name email } }"}'

# File Upload
curl -X POST https://api.zplus.io/api/v1/files/upload \
  -H "Authorization: Bearer $TOKEN" \
  -H "X-Tenant-ID: tenant1" \
  -F "file=@document.pdf" \
  -F "category=documents"
```

## SDKs and Clients

### JavaScript/TypeScript
```bash
npm install @zplus/api-client
```

```typescript
import { ZplusClient } from '@zplus/api-client';

const client = new ZplusClient({
  endpoint: 'https://api.zplus.io/graphql',
  token: 'your-jwt-token',
  tenantId: 'tenant1'
});

const users = await client.user.list();
```

### Go
```bash
go get github.com/zplus/go-client
```

```go
import "github.com/zplus/go-client"

client := zplus.NewClient(&zplus.Config{
    Endpoint: "https://api.zplus.io/graphql",
    Token:    "your-jwt-token", 
    TenantID: "tenant1",
})

users, err := client.User.List()
```

## Testing

### Postman Collection
Download: [Zplus API Collection](./postman/zplus-api.json)

### GraphQL Introspection
```graphql
query IntrospectionQuery {
  __schema {
    queryType { name }
    mutationType { name }
    subscriptionType { name }
  }
}
```

## Versioning

API versions được quản lý through:
- GraphQL schema evolution (backward compatible)
- REST API versioning: `/api/v1/`, `/api/v2/`
- Deprecation notices in response headers

## Support

- **API Issues**: [GitHub Issues](https://github.com/ilmsadmin/Zplus_SaaS_Base/issues)
- **Documentation**: [API Docs](https://docs.zplus.io)
- **Status Page**: [status.zplus.io](https://status.zplus.io)
