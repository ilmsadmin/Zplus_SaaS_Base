# Thiết Kế Hệ Thống Cơ Sở Cho Zplus_SaaS_Base (Multi-Tenant)

## 1. Tổng Quan

**Zplus_SaaS_Base** là một nền tảng SaaS quy mô lớn, sử dụng kiến trúc **microservices** và mô hình **multi-tenant** để phục vụ nhiều khách hàng (tenant). Mỗi tenant được truy cập qua subdomain (`tenant_name.zplus.io`) hoặc domain riêng (custom domain). Hệ thống cung cấp các màn hình đăng nhập riêng biệt cho **system admin/manager**, **tenant admin/manager**, và **user**, với các module service được quản lý tập trung trong dashboard tenant.

### Mục tiêu thiết kế
- **Cô lập tenant**: Dữ liệu và quyền truy cập phân tách hoàn toàn, hỗ trợ subdomain và custom domain.
- **Giao diện quản lý**: Màn hình đăng nhập riêng cho từng vai trò (system admin, tenant admin, user).
- **Module hóa**: Mỗi module service là một màn hình riêng, tích hợp trong dashboard tenant.
- **Hiệu suất**: Tối ưu hóa truy vấn, caching, và SSR.
- **Bảo mật**: HTTPS, RBAC, rate limiting, ngăn rò rỉ dữ liệu giữa tenant.
- **Khả năng mở rộng**: Hỗ trợ hàng nghìn tenant với tự động hóa cao.

### Các thành phần chính
- **Backend**: Go 1.21+, Fiber v2, GORM, gqlgen, Casbin.
- **Frontend**: Next.js 14, Apollo Client, TypeScript, Tailwind CSS.
- **Cơ sở dữ liệu**:
  - PostgreSQL 16 (schema-per-tenant).
  - MongoDB 7 (database-per-tenant).
  - Redis 7 (key prefix per tenant).
- **API Gateway**: Traefik v3, routing theo subdomain/custom domain.
- **GraphQL Gateway**: GraphQL Federation.
- **Auth & RBAC**: Keycloak (single realm, client scope per tenant), Casbin.
- **CI/CD**: GitHub Actions, ArgoCD, Helm.
- **Infra**: Kubernetes (EKS), Prometheus, Grafana, Loki.
- **Bảo mật**: HTTPS, HSTS, CSP, rate limiting, Argon2.

---

## 2. Mô hình Multi-Tenant

### 2.1. Phân tách dữ liệu
- **PostgreSQL**: Schema riêng cho mỗi tenant (e.g., `tenant1.users`, `tenant2.orders`).
- **MongoDB**: Database riêng (e.g., `tenant1_metadata`, `tenant2_metadata`).
- **Redis**: Key prefix (e.g., `tenant1:session:abc123`).

### 2.2. Subdomain và Custom Domain
- **Subdomain**: Mỗi tenant có subdomain dạng `tenant_name.zplus.io`.
  - Ví dụ: `acme.zplus.io`, `beta.zplus.io`.
- **Custom Domain**: Tenant admin có thể thiết lập domain riêng (e.g., `app.acme.com`) trỏ đến subdomain thông qua CNAME record.
  - Cấu hình DNS: Tenant admin thêm CNAME record (e.g., `app.acme.com` → `acme.zplus.io`).
  - Traefik xử lý routing dựa trên `Host` header, ánh xạ custom domain đến `tenant_id`.
- **Database lưu cấu hình domain**:
  - Bảng `tenant_domains` trong PostgreSQL:
    ```sql
    CREATE TABLE tenant_domains (
        id SERIAL PRIMARY KEY,
        tenant_id VARCHAR(50) NOT NULL,
        domain VARCHAR(255) NOT NULL,
        is_custom BOOLEAN DEFAULT FALSE,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );
    ```

### 2.3. Routing
- **Traefik**:
  - Định tuyến dựa trên `Host` header (subdomain hoặc custom domain).
  - Ví dụ rule: `Host("acme.zplus.io") || Host("app.acme.com")` → service `ilms-api`.
  - Middleware Traefik: Gắn `tenant_id` vào header `X-Tenant-Id`.
- **Middleware Go Fiber**:
  - Trích xuất `tenant_id` từ `X-Tenant-Id` hoặc hostname.
  - Kiểm tra quyền truy cập qua Casbin.

### 2.4. Cô lập tenant
- **Dữ liệu**: Schema/database riêng, không có truy vấn cross-tenant.
- **Quyền truy cập**: Casbin policy kiểm tra `tenant_id`, `user_id`, `resource`, `action`.
- **Backup**: Backup riêng cho mỗi tenant (PostgreSQL schema, MongoDB database).
- **Logging**: Log gắn `tenant_id` (e.g., `{ "tenant_id": "acme", "message": "User login" }`).

---

## 3. Hệ thống Auth & RBAC

### 3.1. Vai trò và giao diện đăng nhập
- **System Admin/Manager**:
  - Đăng nhập qua: `admin.zplus.io`.
  - Quyền: Quản lý tất cả tenant, tạo/xóa tenant, cấu hình hệ thống.
  - Keycloak role: `system_admin`, `system_manager`.
- **Tenant Admin/Manager**:
  - Đăng nhập qua: `tenant_name.zplus.io/admin`.
  - Quyền: Quản lý user, cấu hình tenant (e.g., custom domain, theme).
  - Keycloak role: `tenant_admin`, `tenant_manager`.
- **User**:
  - Đăng nhập qua: `tenant_name.zplus.io`.
  - Quyền: Sử dụng dịch vụ (e.g., POS, file upload).
  - Keycloak role: `user`.

### 3.2. Keycloak
- **Single Realm**: `zplus_realm`.
- **Client Scope**: Mỗi tenant có scope riêng (e.g., `acme_scope`).
- **JWT Claims**:
  ```json
  {
    "sub": "user_id",
    "tenant_id": "acme",
    "scope": "acme_scope",
    "roles": ["tenant_admin"]
  }
  ```
- **Flow**: OAuth2 Authorization Code Flow, refresh token trong HttpOnly cookie.

### 3.3. Casbin
- **Policy**: `sub, obj, act, tenant_id` (e.g., `alice, /file, read, acme`).
- **Middleware**:
```go
package middleware

import (
    "github.com/casbin/casbin/v2"
    "github.com/gofiber/fiber/v2"
)

func Authorize(enforcer *casbin.Enforcer) fiber.Handler {
    return func(c *fiber.Ctx) error {
        tenantID := c.Get("X-Tenant-Id")
        userID := c.Locals("user_id").(string)
        resource := c.Path()
        action := c.Method()

        allowed, err := enforcer.Enforce(userID, resource, action, tenantID)
        if err != nil || !allowed {
            return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
        }
        return c.Next()
    }
}
```

### 3.4. Giao diện đăng nhập
- **System Admin/Manager**:
  - URL: `admin.zplus.io/login`.
  - Component: `components/auth/SystemLogin.tsx`.
- **Tenant Admin/Manager**:
  - URL: `tenant_name.zplus.io/admin/login`.
  - Component: `components/auth/TenantAdminLogin.tsx`.
- **User**:
  - URL: `tenant_name.zplus.io/login`.
  - Component: `components/auth/UserLogin.tsx`.

---

## 4. Cấu Trúc Backend

### 4.1. Thư mục dự án
```
backend/
├── cmd/
│   └── api/main.go
├── internal/
│   ├── tenant/              # Quản lý tenant (CRUD, custom domain)
│   │   ├── service.go
│   │   ├── repository.go
│   │   └── model.go
│   ├── user/                # Quản lý user
│   ├── auth/                # Xác thực và phân quyền
│   ├── file/                # Upload/download file
│   ├── pos/                 # Point of Sale
│   └── common/              # Logging, config, errors
├── pkg/
│   ├── middleware/          # Tenant resolver, auth
│   ├── graphql/             # GraphQL schema, resolver
│   └── db/                  # DB connections
├── schema/
│   ├── auth.graphql
│   ├── user.graphql
│   ├── file.graphql
│   └── pos.graphql
└── scripts/
    └── create_tenant.sh
```

### 4.2. Ví dụ: Tenant Service (Custom Domain)
```go
package tenant

type Tenant struct {
    ID           string `gorm:"primaryKey"`
    Name         string
    CustomDomain string
}

type Repository struct {
    db *gorm.DB
}

func (r *Repository) AddCustomDomain(tenantID, domain string) error {
    return r.db.Model(&Tenant{ID: tenantID}).Update("custom_domain", domain).Error
}
```

### 4.3. GraphQL Schema
```graphql
type Tenant {
    id: ID!
    name: String!
    customDomain: String
}

type Query {
    getTenant(id: ID!): Tenant!
}

type Mutation {
    addCustomDomain(tenantId: ID!, domain: String!): Tenant!
}
```

---

## 5. Cấu Trúc Frontend

### 5.1. Thư mục dự án
```
frontend/
├── pages/
│   ├── _app.tsx
│   ├── index.tsx
│   ├── login.tsx            # User login
│   ├── admin/
│   │   ├── login.tsx       # System admin login
│   │   └── dashboard.tsx   # System admin dashboard
│   └── [tenantId]/
│       ├── login.tsx       # Tenant user login
│       ├── admin/
│       │   ├── login.tsx   # Tenant admin login
│       │   └── dashboard.tsx # Tenant admin dashboard
│       └── dashboard.tsx   # Tenant user dashboard
├── components/
│   ├── auth/
│   │   ├── SystemLogin.tsx
│   │   ├── TenantAdminLogin.tsx
│   │   └── UserLogin.tsx
│   ├── file/
│   │   └── FileManager.tsx
│   └── pos/
│       └── Checkout.tsx
├── lib/
│   ├── apolloClient.ts
│   ├── auth.ts
│   └── tenant.ts
├── styles/
│   └── globals.css
└── public/
```

### 5.2. Tenant Dashboard
- **URL**: `tenant_name.zplus.io/dashboard`.
- **Chức năng**: Hiển thị danh sách module (e.g., User, File, POS) dưới dạng card hoặc menu.
- **Component**:
```tsx
// components/dashboard/TenantDashboard.tsx
import { useRouter } from 'next/router';
import Link from 'next/link';

export default function TenantDashboard() {
  const router = useRouter();
  const { tenantId } = router.query;

  const modules = [
    { name: 'User Management', path: `/${tenantId}/users` },
    { name: 'File Management', path: `/${tenantId}/files` },
    { name: 'POS', path: `/${tenantId}/pos` },
  ];

  return (
    <div className="p-4">
      <h1 className="text-2xl">Welcome to {tenantId} Dashboard</h1>
      <div className="grid grid-cols-3 gap-4">
        {modules.map((module) => (
          <Link key={module.name} href={module.path}>
            <div className="p-4 bg-blue-500 text-white rounded">
              {module.name}
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
```

### 5.3. Tenant Admin Dashboard
- **URL**: `tenant_name.zplus.io/admin/dashboard`.
- **Chức năng**: Quản lý user, custom domain, theme.
- **Component**:
```tsx
// pages/[tenantId]/admin/dashboard.tsx
import { useMutation } from '@apollo/client';
import { gql } from '@apollo/client';

const ADD_CUSTOM_DOMAIN = gql`
  mutation AddCustomDomain($tenantId: ID!, $domain: String!) {
    addCustomDomain(tenantId: $tenantId, domain: $domain) {
      id
      customDomain
    }
  }
`;

export default function TenantAdminDashboard({ tenantId }: { tenantId: string }) {
  const [addCustomDomain] = useMutation(ADD_CUSTOM_DOMAIN);

  const handleAddDomain = async (domain: string) => {
    await addCustomDomain({ variables: { tenantId, domain } });
  };

  return (
    <div className="p-4">
      <h1 className="text-2xl">Tenant Admin Dashboard</h1>
      <div>
        <input
          type="text"
          placeholder="Enter custom domain"
          onBlur={(e) => handleAddDomain(e.target.value)}
          className="border p-2"
        />
      </div>
    </div>
  );
}
```

---

## 6. CI/CD và Triển khai

### 6.1. Pipeline
- **GitHub Actions**:
  - Build Docker images: `zplus/api`, `zplus/ui`.
  - Test: Unit test, integration test.
  - Push: Đẩy image lên AWS ECR.
- **ArgoCD**:
  - Đồng bộ Helm chart với EKS cluster.
  - Rolling update cho zero downtime.

### 6.2. Helm Chart
- Cấu hình Traefik cho custom domain:
```yaml
# helm/traefik/values.yaml
ingressRoute:
  dashboard:
    enabled: true
  routes:
    - match: Host(`admin.zplus.io`)
      kind: Rule
      services:
        - name: ilms-admin
          port: 80
    - match: HostRegexp(`{tenant:[a-z0-9]+}.zplus.io`) || HostRegexp(`{custom:.+}`)
      kind: Rule
      middlewares:
        - name: tenant-resolver
      services:
        - name: ilms-api
          port: 80
```

### 6.3. Script tạo tenant
```bash
# scripts/create_tenant.sh
#!/bin/bash
TENANT_ID=$1
psql -U postgres -d zplus -c "CREATE SCHEMA $TENANT_ID;"
mongosh --eval "db.getSiblingDB('$TENANT_ID_metadata').createCollection('config')"
helm upgrade --install $TENANT_ID helm/ilms-api --set tenant.id=$TENANT_ID
kubectl apply -f - <<EOF
apiVersion: traefik.io/v1alpha1
kind: IngressRoute
metadata:
  name: $TENANT_ID-route
spec:
  routes:
  - match: Host(\`$TENANT_ID.zplus.io\`)
    kind: Rule
    services:
    - name: ilms-api
      port: 80
EOF
```

---

## 7. Monitoring & Logging

- **Prometheus + Grafana**: Metrics per tenant (e.g., `requests_total{tenant_id="acme"}`).
- **Loki**: Log với label `tenant_id`.
- **Alertmanager**: Cảnh báo CPU, memory, request timeout.

---

## 8. Bảo mật

- **HTTPS**: Let's Encrypt tự động cấp SSL cho subdomain và custom domain.
- **HSTS**: `Strict-Transport-Security: max-age=31536000`.
- **Rate Limiting**: Traefik (100 req/s/tenant), Fiber middleware.
- **CORS**: Chỉ cho phép `*.zplus.io` và custom domain.

---

## 9. Backup và Disaster Recovery

- **PostgreSQL**: `pg_dump` mỗi schema, lưu trên S3.
- **MongoDB**: `mongodump` mỗi database, lưu trên S3.
- **Redis**: Export key theo prefix, lưu trên S3.
- **HA**: Patroni (PostgreSQL), replica sets (MongoDB), multi-AZ EKS.

---

## 10. Khả năng mở rộng

- **Thêm module**: Tạo folder trong `internal/`, thêm schema GraphQL, component trong `components/`.
- **Thêm tenant**: Chạy `create_tenant.sh`.
- **Scale ngang**: HPA dựa trên CPU/memory.
- **White-label**: Theme và logo lưu trong MongoDB (`tenant_config`).

---

## 11. Công nghệ sử dụng

| Layer         | Stack                                    |
|---------------|------------------------------------------|
| Backend       | Go 1.21+, Fiber v2, GORM, gqlgen, Casbin |
| Frontend      | Next.js 14, Apollo, TypeScript, Tailwind |
| Auth          | Keycloak                                 |
| Gateway       | Traefik v3                               |
| DB            | PostgreSQL 16, MongoDB 7, Redis 7        |
| Infra         | Kubernetes (EKS), Helm, ArgoCD           |
| Observability | Prometheus, Grafana, Loki                |

---

## 12. Kết luận

**Zplus_SaaS_Base** là một nền tảng SaaS mạnh mẽ, hỗ trợ:
- **Subdomain và custom domain**: Linh hoạt cho branding tenant.
- **Đăng nhập riêng biệt**: Tách biệt rõ ràng giữa system admin, tenant admin, và user.
- **Module hóa**: Dashboard tập trung, dễ mở rộng.
- **Bảo mật và hiệu suất**: Đáp ứng yêu cầu production.
Hệ thống sẵn sàng cho hàng nghìn tenant với tự động hóa và khả năng mở rộng cao.