# ADR-009: Custom Domain Support for Multi-Tenant Architecture

## Status
Accepted

## Context

Zplus SaaS Base cần hỗ trợ custom domain cho các tenant để:
- Cho phép tenant sử dụng domain riêng (white-label)
- Tăng brand identity cho tenant
- Cải thiện SEO và user experience
- Đáp ứng yêu cầu enterprise customers

Hiện tại hệ thống chỉ hỗ trợ subdomain pattern (`tenant.zplus.io`). Cần mở rộng để hỗ trợ custom domain như `app.acme.com` point đến tenant `acme`.

## Decision

Chúng tôi quyết định implement custom domain support với:

### 1. Database Schema
```sql
CREATE TABLE tenant_domains (
    id SERIAL PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    domain VARCHAR(255) NOT NULL UNIQUE,
    is_custom BOOLEAN DEFAULT FALSE,
    verified BOOLEAN DEFAULT FALSE,
    ssl_enabled BOOLEAN DEFAULT FALSE,
    verification_token VARCHAR(255),
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. Domain Verification Process
- DNS TXT record verification: `_zplus-verify.{domain}` với verification token
- Automatic SSL certificate generation sau khi verify
- Domain ownership re-verification mỗi 30 ngày

### 3. Traefik Configuration
```yaml
# Dynamic routing rule
- match: HostRegexp(`{tenant:[a-z0-9]+}.zplus.io`) || HostRegexp(`{custom:.+}`)
  kind: Rule
  middlewares:
    - name: tenant-resolver
  services:
    - name: ilms-api
      port: 80
```

### 4. Middleware Implementation
- `tenant-resolver` middleware để extract tenant_id từ custom domain
- Database lookup để map custom domain → tenant_id
- Cache mapping trong Redis cho performance

### 5. SSL Certificate Management
- Traefik automatic HTTPS với Let's Encrypt
- Wildcard certificate cho `*.zplus.io`
- Individual certificates cho custom domains

## Consequences

### Positive
- **Improved Branding**: Tenant có thể sử dụng domain riêng
- **Enterprise Ready**: Đáp ứng requirement của enterprise customers
- **SEO Benefits**: Better SEO cho tenant websites
- **Scalability**: Support unlimited custom domains per tenant

### Negative
- **Complexity**: Tăng complexity của routing và SSL management
- **Monitoring**: Cần monitor SSL certificate expiration
- **Security**: Risk của domain hijacking nếu verification không chặt chẽ
- **Performance**: Additional database lookup cho domain resolution

### Mitigation
- Implement robust domain verification process
- Cache domain mappings trong Redis
- Monitor SSL certificate status
- Rate limiting cho domain operations
- Regular security audits

## Alternatives Considered

### 1. CNAME-only Approach
**Pros**: Simpler implementation
**Cons**: Less flexible, không support root domain

### 2. Separate Load Balancer per Domain
**Pros**: Better isolation
**Cons**: Higher cost, management overhead

### 3. Third-party Domain Service
**Pros**: Less implementation effort
**Cons**: Vendor lock-in, higher cost

## Implementation Plan

### Phase 1: Core Infrastructure
1. Database schema migration
2. Domain verification API
3. Basic DNS validation

### Phase 2: Traefik Integration
1. Dynamic routing configuration
2. SSL certificate automation
3. Domain mapping middleware

### Phase 3: Frontend Integration
1. Domain management UI
2. DNS setup instructions
3. Verification status tracking

### Phase 4: Monitoring & Security
1. Certificate monitoring
2. Domain ownership re-verification
3. Security hardening

## References

- [RFC 1035: Domain Names](https://tools.ietf.org/html/rfc1035)
- [Let's Encrypt ACME Protocol](https://letsencrypt.org/docs/acme-protocol/)
- [Traefik Custom Domain Configuration](https://doc.traefik.io/traefik/routing/routers/)
- [Multi-tenant SaaS Best Practices](https://docs.aws.amazon.com/wellarchitected/latest/saas-lens/tenant-isolation.html)

## Security Considerations

1. **Domain Verification**
   - DNS TXT record validation
   - Protection against subdomain takeover
   - Rate limiting domain addition

2. **SSL Certificate Management**
   - Automatic certificate renewal
   - Certificate transparency monitoring
   - Secure certificate storage

3. **Access Control**
   - Domain ownership verification
   - Tenant isolation validation
   - Admin-only domain management

## Monitoring Requirements

1. **Certificate Expiration**: Alert 30 days before expiration
2. **Domain Resolution**: Monitor DNS resolution times
3. **Verification Status**: Track verification success/failure rates
4. **Security Events**: Log domain-related security events

---

**Date**: August 7, 2025  
**Reviewers**: Tech Lead, DevOps Lead, Security Lead  
**Next Review**: September 7, 2025
