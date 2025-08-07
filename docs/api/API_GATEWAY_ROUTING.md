# API Gateway & Routing Implementation Guide

## Overview

This document describes the complete implementation of the API Gateway & Routing system for the Zplus SaaS Base multi-tenant architecture. The system provides dynamic routing, SSL termination, rate limiting, health checks, and domain management capabilities.

## Architecture Components

### 1. Traefik API Gateway
- **Purpose**: Reverse proxy and load balancer with automatic SSL
- **Features**: Dynamic routing, SSL/TLS termination, rate limiting, health checks
- **Configuration**: Static config in `traefik/traefik.yml`, dynamic routing in `traefik/dynamic/routing.yml`

### 2. Domain Management Service
- **Purpose**: Handles custom domain addition, verification, and management
- **Components**: `DomainService`, `DNSValidationService`, domain handlers and routes
- **Database**: Enhanced `tenant_domains` table with SSL, validation, and routing config

### 3. DNS Validation System
- **Methods**: DNS TXT record validation, HTTP file validation
- **Providers**: Cloudflare DNS, Let's Encrypt ACME
- **Security**: Domain ownership verification, rate limiting

### 4. SSL Certificate Management
- **Automated**: Let's Encrypt integration with Traefik
- **Monitoring**: Certificate expiration tracking and auto-renewal
- **Storage**: Persistent certificate storage with backup

## Database Schema

### Enhanced tenant_domains Table
```sql
-- Core domain information
id                    UUID PRIMARY KEY
tenant_id             VARCHAR(50) NOT NULL
domain                VARCHAR(255) UNIQUE NOT NULL
is_custom             BOOLEAN DEFAULT FALSE
verified              BOOLEAN DEFAULT FALSE
ssl_enabled           BOOLEAN DEFAULT FALSE

-- Verification and SSL
verification_token    VARCHAR(255)
verification_method   VARCHAR(20) DEFAULT 'dns'
verified_at           TIMESTAMP
ssl_cert_issued_at    TIMESTAMP
ssl_cert_expires_at   TIMESTAMP

-- DNS and SSL configuration
dns_provider          VARCHAR(50) DEFAULT 'auto'
dns_zone_id           VARCHAR(100)
ssl_issuer            VARCHAR(50) DEFAULT 'letsencrypt'
ssl_auto_renew        BOOLEAN DEFAULT TRUE

-- Routing configuration
routing_priority      INTEGER DEFAULT 100
rate_limit_config     JSONB
security_config       JSONB
health_check_config   JSONB

-- Status and monitoring
status               VARCHAR(20) DEFAULT 'active'
metrics_enabled      BOOLEAN DEFAULT TRUE
last_health_check    TIMESTAMP
last_ssl_check       TIMESTAMP
```

### Supporting Tables
- `domain_validation_logs`: Tracks verification attempts
- `ssl_certificates`: Stores certificate details
- `domain_routing_cache`: Performance cache for routing decisions
- `domain_metrics`: Domain performance metrics

## API Endpoints

### Tenant Domain Management
```
GET    /api/v1/tenants/{tenant_id}/domains
POST   /api/v1/tenants/{tenant_id}/domains
POST   /api/v1/tenants/{tenant_id}/domains/{id}/verify
GET    /api/v1/tenants/{tenant_id}/domains/{id}/instructions
DELETE /api/v1/tenants/{tenant_id}/domains/{id}
```

### Public Domain Status
```
GET    /api/v1/domains/{domain}/status
GET    /api/v1/domains/{domain}/metrics
```

### Admin Management
```
GET    /api/v1/admin/domains
```

## Domain Verification Process

### 1. Add Custom Domain
```json
POST /api/v1/tenants/acme/domains
{
  "domain": "app.acme.com",
  "verification_method": "dns",
  "auto_ssl": true,
  "priority": 150
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "domain": "app.acme.com",
    "verification_token": "zplus-verify-abc123def456",
    "verification_method": "dns",
    "dns_record": {
      "type": "TXT",
      "name": "_zplus-verify.app.acme.com",
      "value": "zplus-verify-abc123def456",
      "ttl": 300
    },
    "instructions": "Add a TXT record to your DNS...",
    "expires_at": "2025-08-08T10:00:00Z"
  }
}
```

### 2. DNS Configuration
Customer adds TXT record:
- **Type**: TXT
- **Name**: `_zplus-verify.app.acme.com`
- **Value**: `zplus-verify-abc123def456`
- **TTL**: 300 seconds

### 3. Domain Verification
```json
POST /api/v1/tenants/acme/domains/{domain_id}/verify
```

**Response:**
```json
{
  "success": true,
  "data": {
    "domain": "app.acme.com",
    "verified": true,
    "ssl_enabled": false,
    "status": "active",
    "last_checked": "2025-08-07T15:30:00Z",
    "health_status": "healthy"
  }
}
```

### 4. SSL Certificate Issuance
- Automatic Let's Encrypt certificate request
- Domain validation via ACME protocol
- Certificate storage and deployment
- Auto-renewal 30 days before expiration

## Traefik Configuration

### Static Configuration (`traefik/traefik.yml`)
```yaml
# Entry points
entryPoints:
  web:
    address: ":80"
    http:
      redirections:
        entrypoint:
          to: websecure
          scheme: https
  websecure:
    address: ":443"

# Certificate resolvers
certificatesResolvers:
  letsencrypt:
    acme:
      tlsChallenge: {}
      email: admin@zplus.io
      storage: /letsencrypt/acme.json
  
  cloudflare:
    acme:
      dnsChallenge:
        provider: cloudflare
      email: admin@zplus.io
      storage: /letsencrypt/cloudflare.json
```

### Dynamic Configuration (`traefik/dynamic/routing.yml`)
```yaml
http:
  middlewares:
    # Rate limiting per tenant type
    admin-ratelimit:
      rateLimit:
        average: 50
        burst: 100
        period: 1m
    
    tenant-admin-ratelimit:
      rateLimit:
        average: 25
        burst: 50
        period: 1m
    
    user-ratelimit:
      rateLimit:
        average: 15
        burst: 30
        period: 1m
    
    # Security headers
    security-headers:
      headers:
        frameDeny: true
        contentTypeNosniff: true
        browserXssFilter: true
        forceSTSHeader: true
        stsIncludeSubdomains: true
        stsPreload: true
        stsSeconds: 31536000

  routers:
    # Admin domain routing
    admin-router:
      rule: "Host(`admin.zplus.io`)"
      entryPoints:
        - websecure
      middlewares:
        - admin-ratelimit
        - security-headers
      service: ilms-api
      tls:
        certResolver: letsencrypt
    
    # Tenant subdomain routing
    tenant-router:
      rule: "HostRegexp(`{subdomain:[a-z0-9-]+}.zplus.io`)"
      entryPoints:
        - websecure
      middlewares:
        - tenant-admin-ratelimit
        - security-headers
        - tenant-extractor
      service: ilms-api
      tls:
        certResolver: cloudflare
        domains:
          - main: "*.zplus.io"
    
    # Custom domain routing
    custom-domain-router:
      rule: "HostRegexp(`{domain:.+}`)"
      entryPoints:
        - websecure
      middlewares:
        - user-ratelimit
        - security-headers
        - domain-validator
      service: ilms-api
      tls:
        certResolver: letsencrypt
      priority: 100

  services:
    ilms-api:
      loadBalancer:
        servers:
          - url: "http://zplus-backend:8080"
        healthCheck:
          path: "/health"
          interval: "30s"
          timeout: "10s"
```

## Rate Limiting Configuration

### Per Tenant Type
- **Admin** (`admin.zplus.io`): 50 requests/minute, burst 100
- **Tenant Admin** (`*.zplus.io`): 25 requests/minute, burst 50  
- **Users** (custom domains): 15 requests/minute, burst 30

### Per API Endpoint
- Domain management: 10 requests/minute
- Domain verification: 5 requests/minute
- Metrics endpoints: 100 requests/minute

## Health Checks

### Backend Health Check
```yaml
healthCheck:
  path: "/health"
  interval: "30s"
  timeout: "10s"
  retries: 3
```

### Domain Health Monitoring
- SSL certificate expiration (30-day warning)
- DNS resolution validation
- HTTP/HTTPS connectivity tests
- Response time monitoring

## SSL/TLS Configuration

### Automatic HTTPS
- All HTTP traffic redirected to HTTPS
- HSTS headers with 1-year max-age
- Secure cipher suites only

### Certificate Management
- Let's Encrypt for individual domains
- Cloudflare DNS challenge for wildcard `*.zplus.io`
- 30-day expiration warnings
- Automatic renewal process

### Certificate Storage
```
/letsencrypt/
├── acme.json          # Let's Encrypt certificates
├── cloudflare.json    # Wildcard certificates
└── backup/            # Certificate backups
```

## Monitoring and Metrics

### Prometheus Metrics
```
# Traefik metrics endpoint
http://metrics.zplus.io/metrics

# Key metrics
traefik_http_requests_total
traefik_http_request_duration_seconds
traefik_config_reloads_total
traefik_tls_certs_not_after
```

### Domain Metrics
```json
{
  "domain": "app.acme.com",
  "metrics": {
    "requests": {
      "total": 12500,
      "2xx": 11875,
      "4xx": 500,
      "5xx": 125
    },
    "response_time": {
      "avg": 120,
      "p50": 100,
      "p95": 180,
      "p99": 250
    },
    "ssl": {
      "enabled": true,
      "expires_at": "2025-11-05T10:00:00Z",
      "auto_renew": true,
      "issuer": "Let's Encrypt"
    }
  }
}
```

## Security Considerations

### Domain Validation Security
1. **DNS Verification**: TXT record validation prevents subdomain takeover
2. **Rate Limiting**: Prevents abuse of domain addition/verification
3. **Token Expiration**: Verification tokens expire in 24 hours
4. **Ownership Re-verification**: Periodic re-validation every 30 days

### SSL Security
1. **HSTS Headers**: Force HTTPS with long max-age
2. **Certificate Transparency**: Monitor CT logs for unauthorized certificates
3. **Secure Storage**: Encrypted certificate storage with access controls
4. **Key Rotation**: Regular certificate renewal and key rotation

### Access Control
1. **Tenant Isolation**: Strict tenant-based domain access control
2. **Admin Privileges**: System admin access for cross-tenant operations
3. **API Authentication**: JWT token validation for all operations
4. **Audit Logging**: Complete audit trail of domain operations

## Performance Optimization

### Caching Strategy
```sql
-- Domain routing cache (1-hour TTL)
CREATE TABLE domain_routing_cache (
    domain VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(50) NOT NULL,
    routing_config JSONB,
    cache_expires_at TIMESTAMP NOT NULL
);
```

### Database Optimization
- Indexed domain lookups
- Materialized views for complex queries
- Connection pooling
- Read replicas for metrics

### CDN Integration
- Cloudflare for DNS and CDN
- Geographic load balancing
- DDoS protection
- Edge caching

## Deployment Guide

### 1. Prerequisites
```bash
# Install Docker and Docker Compose
# Configure Cloudflare DNS API credentials
# Set up monitoring and alerting
```

### 2. Environment Configuration
```bash
# Copy environment template
cp .env.example .env

# Configure required variables
CLOUDFLARE_EMAIL=admin@zplus.io
CLOUDFLARE_API_KEY=your-api-key
CLOUDFLARE_DNS_API_TOKEN=your-dns-token
ACME_EMAIL=admin@zplus.io
```

### 3. Database Migration
```bash
# Run enhanced domain migration
docker exec zplus_postgres psql -U postgres -d zplus_saas_base \
  -f /tmp/migrations/008_enhance_tenant_domains_table.sql

# Run seeder for sample data
docker exec zplus_postgres psql -U postgres -d zplus_saas_base \
  -f /tmp/seeders/008_enhanced_domain_configurations_seeder.sql
```

### 4. Traefik Deployment
```bash
# Start Traefik with API Gateway
docker-compose -f docker-compose.traefik.yml up -d

# Verify Traefik dashboard
curl http://localhost:8080/dashboard/
```

### 5. Testing
```bash
# Run domain management tests
./backend/scripts/test-domain-management.sh

# Test DNS validation
./backend/scripts/test-dns-validation.sh

# Performance testing
./backend/scripts/test-performance.sh
```

## Troubleshooting

### Common Issues

#### 1. Domain Verification Fails
```bash
# Check DNS propagation
dig TXT _zplus-verify.yourdomain.com

# Check validation logs
docker exec zplus_postgres psql -U postgres -d zplus_saas_base \
  -c "SELECT * FROM domain_validation_logs WHERE status = 'failed' ORDER BY created_at DESC LIMIT 10;"
```

#### 2. SSL Certificate Issues
```bash
# Check certificate status
docker exec zplus_traefik ls -la /letsencrypt/

# Check Traefik logs
docker logs zplus_traefik | grep -i certificate

# Manual certificate request
docker exec zplus_traefik traefik acme --email admin@zplus.io --domain yourdomain.com
```

#### 3. Rate Limiting Issues
```bash
# Check rate limiting logs
docker logs zplus_traefik | grep -i ratelimit

# Adjust rate limits in dynamic config
# Edit traefik/dynamic/routing.yml and restart
```

### Monitoring and Alerts

#### Prometheus Alerts
```yaml
# SSL certificate expiring
- alert: SSLCertificateExpiringSoon
  expr: traefik_tls_certs_not_after - time() < 30 * 24 * 3600
  for: 1h
  labels:
    severity: warning
  annotations:
    summary: "SSL certificate expiring soon"

# High error rate
- alert: HighErrorRate
  expr: rate(traefik_http_requests_total{code=~"5.."}[5m]) > 0.1
  for: 5m
  labels:
    severity: critical
  annotations:
    summary: "High 5xx error rate detected"
```

#### Health Check Monitoring
```bash
# Check backend health
curl http://localhost:8080/health

# Check domain health
curl http://api.zplus.io/api/v1/domains/yourdomain.com/status

# Check Traefik health
curl http://localhost:8080/ping
```

## Next Steps

### Phase 2 Enhancements
1. **GraphQL Federation**: Federated GraphQL gateway
2. **Advanced Monitoring**: Custom metrics and dashboards
3. **Auto-scaling**: Dynamic backend scaling based on load
4. **Global Load Balancing**: Multi-region deployment

### Integration Points
1. **CI/CD Pipeline**: Automated deployment and testing
2. **Monitoring Stack**: Prometheus, Grafana, AlertManager
3. **Backup System**: Automated certificate and config backups
4. **Documentation**: API documentation and user guides

## Conclusion

The API Gateway & Routing implementation provides a robust, scalable foundation for the multi-tenant SaaS platform. It includes:

✅ **Dynamic routing** for subdomains and custom domains  
✅ **Automatic SSL** certificate management  
✅ **Rate limiting** per tenant type and endpoint  
✅ **Health monitoring** and metrics collection  
✅ **Domain management** API with verification  
✅ **Security hardening** with HSTS and secure headers  
✅ **Performance optimization** with caching and CDN  

The system is production-ready and includes comprehensive testing, monitoring, and troubleshooting capabilities.
