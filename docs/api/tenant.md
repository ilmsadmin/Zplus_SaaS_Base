# Tenant API Documentation

## Overview

Tenant API cung cấp các endpoints để quản lý tenant, domain configuration, và settings.

## Table of Contents

- [Tenant Management](#tenant-management)
- [Domain Management](#domain-management)
- [Tenant Configuration](#tenant-configuration)
- [GraphQL Schema](#graphql-schema)
- [Examples](#examples)

## Tenant Management

### Create Tenant (System Admin Only)

```graphql
mutation CreateTenant($input: CreateTenantInput!) {
  createTenant(input: $input) {
    id
    name
    subdomain
    status
    createdAt
  }
}
```

**Input:**
```typescript
interface CreateTenantInput {
  name: string;
  subdomain: string;
  adminEmail: string;
  adminName: string;
  plan?: string;
}
```

### Get Tenant Information

```graphql
query GetTenant($id: ID!) {
  tenant(id: $id) {
    id
    name
    subdomain
    customDomains {
      id
      domain
      verified
      sslEnabled
    }
    settings {
      theme
      logo
      features
    }
    status
    createdAt
  }
}
```

### Update Tenant

```graphql
mutation UpdateTenant($id: ID!, $input: UpdateTenantInput!) {
  updateTenant(id: $id, input: $input) {
    id
    name
    settings {
      theme
      logo
      features
    }
  }
}
```

## Domain Management

### Add Custom Domain

```graphql
mutation AddCustomDomain($tenantId: ID!, $domain: String!) {
  addCustomDomain(tenantId: $tenantId, domain: $domain) {
    id
    domain
    verified
    sslEnabled
    verificationRecord {
      type
      name
      value
    }
  }
}
```

### Verify Domain

```graphql
mutation VerifyDomain($domainId: ID!) {
  verifyDomain(domainId: $domainId) {
    id
    domain
    verified
    sslEnabled
    verifiedAt
  }
}
```

### Remove Custom Domain

```graphql
mutation RemoveCustomDomain($domainId: ID!) {
  removeCustomDomain(domainId: $domainId) {
    success
    message
  }
}
```

## Tenant Configuration

### Update Tenant Settings

```graphql
mutation UpdateTenantSettings($tenantId: ID!, $settings: TenantSettingsInput!) {
  updateTenantSettings(tenantId: $tenantId, settings: $settings) {
    theme {
      primaryColor
      secondaryColor
      logo
    }
    features {
      fileUpload
      userManagement
      analytics
      customBranding
    }
    limits {
      maxUsers
      maxStorageMB
      maxApiCallsPerHour
    }
  }
}
```

## GraphQL Schema

### Types

```graphql
type Tenant {
  id: ID!
  name: String!
  subdomain: String!
  customDomains: [CustomDomain!]!
  settings: TenantSettings!
  status: TenantStatus!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type CustomDomain {
  id: ID!
  tenantId: ID!
  domain: String!
  verified: Boolean!
  sslEnabled: Boolean!
  verificationRecord: DNSRecord
  verifiedAt: DateTime
  createdAt: DateTime!
}

type DNSRecord {
  type: String!
  name: String!
  value: String!
}

type TenantSettings {
  theme: ThemeSettings!
  features: FeatureSettings!
  limits: LimitSettings!
}

type ThemeSettings {
  primaryColor: String
  secondaryColor: String
  logo: String
  favicon: String
}

type FeatureSettings {
  fileUpload: Boolean!
  userManagement: Boolean!
  analytics: Boolean!
  customBranding: Boolean!
}

type LimitSettings {
  maxUsers: Int!
  maxStorageMB: Int!
  maxApiCallsPerHour: Int!
}

enum TenantStatus {
  ACTIVE
  SUSPENDED
  PENDING
  CANCELLED
}
```

### Input Types

```graphql
input CreateTenantInput {
  name: String!
  subdomain: String!
  adminEmail: String!
  adminName: String!
  plan: String
}

input UpdateTenantInput {
  name: String
  settings: TenantSettingsInput
}

input TenantSettingsInput {
  theme: ThemeSettingsInput
  features: FeatureSettingsInput
  limits: LimitSettingsInput
}

input ThemeSettingsInput {
  primaryColor: String
  secondaryColor: String
  logo: String
  favicon: String
}

input FeatureSettingsInput {
  fileUpload: Boolean
  userManagement: Boolean
  analytics: Boolean
  customBranding: Boolean
}

input LimitSettingsInput {
  maxUsers: Int
  maxStorageMB: Int
  maxApiCallsPerHour: Int
}
```

## Examples

### Complete Tenant Setup Flow

```typescript
// 1. Create new tenant (System Admin)
const createTenantResult = await client.mutate({
  mutation: CREATE_TENANT,
  variables: {
    input: {
      name: "ACME Corporation",
      subdomain: "acme",
      adminEmail: "admin@acme.com",
      adminName: "John Admin",
      plan: "enterprise"
    }
  }
});

// 2. Add custom domain (Tenant Admin)
const addDomainResult = await client.mutate({
  mutation: ADD_CUSTOM_DOMAIN,
  variables: {
    tenantId: "tenant_123",
    domain: "app.acme.com"
  }
});

// 3. Get DNS verification record
const domain = addDomainResult.data.addCustomDomain;
console.log(`Add DNS record: ${domain.verificationRecord.type} ${domain.verificationRecord.name} ${domain.verificationRecord.value}`);

// 4. Verify domain after DNS setup
const verifyResult = await client.mutate({
  mutation: VERIFY_DOMAIN,
  variables: {
    domainId: domain.id
  }
});

// 5. Update tenant theme
const updateSettingsResult = await client.mutate({
  mutation: UPDATE_TENANT_SETTINGS,
  variables: {
    tenantId: "tenant_123",
    settings: {
      theme: {
        primaryColor: "#007bff",
        secondaryColor: "#6c757d",
        logo: "https://cdn.acme.com/logo.png"
      },
      features: {
        customBranding: true,
        analytics: true
      }
    }
  }
});
```

### Domain Verification Process

```bash
# DNS Record Setup for Custom Domain
# Add CNAME record to verify domain ownership

# For domain: app.acme.com
# Add CNAME record:
_zplus-verify.app.acme.com -> verify-token-abc123.zplus.io

# After DNS propagation, call verify API
curl -X POST https://api.zplus.io/graphql \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { verifyDomain(domainId: \"domain_456\") { verified sslEnabled } }"
  }'
```

### REST API Alternatives

```http
# Get tenant domains
GET /api/v1/tenants/{tenant_id}/domains
Authorization: Bearer <token>

# Add custom domain
POST /api/v1/tenants/{tenant_id}/domains
Content-Type: application/json
Authorization: Bearer <token>

{
  "domain": "app.acme.com"
}

# Verify domain
POST /api/v1/domains/{domain_id}/verify
Authorization: Bearer <token>
```

## Error Handling

### Common Errors

```json
{
  "errors": [
    {
      "message": "Domain already exists",
      "extensions": {
        "code": "DOMAIN_ALREADY_EXISTS",
        "domain": "app.acme.com"
      }
    }
  ]
}
```

### Error Codes

- `TENANT_NOT_FOUND`: Tenant không tồn tại
- `SUBDOMAIN_TAKEN`: Subdomain đã được sử dụng
- `DOMAIN_ALREADY_EXISTS`: Domain đã được thêm bởi tenant khác
- `DOMAIN_VERIFICATION_FAILED`: Không thể verify domain
- `INSUFFICIENT_PERMISSIONS`: Không đủ quyền truy cập
- `TENANT_LIMIT_EXCEEDED`: Vượt quá giới hạn tenant

## Rate Limiting

- **Tenant Creation**: 10 requests/hour (System Admin)
- **Domain Operations**: 50 requests/hour per tenant
- **Settings Update**: 100 requests/hour per tenant

## Security Notes

- Custom domain verification required trước khi SSL activation
- CNAME records phải point đến verified subdomain
- SSL certificates được auto-provision sau khi verify
- Domain ownership phải được verify định kỳ (30 days)
