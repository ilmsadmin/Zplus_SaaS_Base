# Environment Configuration Guide

This document explains how to configure the Zplus SaaS Base environment variables for different deployment scenarios.

## 📁 Environment Files Structure

```
/
├── .env.example                 # Main template with all variables
├── .env.development            # Development overrides
├── .env.staging                # Staging environment config
├── .env.production             # Production environment config
├── backend/
│   └── .env.example            # Backend-specific configuration
└── frontend/
    └── .env.example            # Frontend-specific configuration
```

## 🚀 Quick Setup

### Development Environment

1. **Copy environment files:**
   ```bash
   make dev-setup
   ```
   This will create:
   - `.env` from `.env.example`
   - `backend/.env` from `backend/.env.example`
   - `frontend/.env.local` from `frontend/.env.example`

2. **Review and update values** in the created files

3. **Start development environment:**
   ```bash
   make dev-up
   ```

### Production Environment

1. **Copy production template:**
   ```bash
   cp .env.production .env
   ```

2. **Generate secure secrets:**
   ```bash
   # Generate JWT secret (256-bit)
   openssl rand -base64 32

   # Generate NextAuth secret
   openssl rand -base64 32

   # Generate encryption key (32 characters)
   openssl rand -hex 16
   ```

3. **Update all REQUIRED values** marked in `.env.production`

## 🔧 Configuration Categories

### Core Application
- `APP_NAME`: Application name
- `ENVIRONMENT`: Runtime environment (development/staging/production)
- `DEBUG`: Enable/disable debug mode

### Database Configuration
- **PostgreSQL**: Primary relational database
- **MongoDB**: Document storage and metadata
- **Redis**: Caching and session storage

### Authentication & Security
- **JWT**: Token-based authentication
- **Keycloak**: Identity and access management
- **Encryption**: Data encryption settings
- **CORS**: Cross-origin resource sharing

### Multi-Tenancy
- **Tenant Strategy**: How tenants are isolated (schema/database)
- **Domain Patterns**: Subdomain and custom domain support
- **Isolation**: Tenant data separation

### External Services
- **File Storage**: AWS S3 or local storage
- **Email**: SMTP configuration
- **Payment**: Stripe integration
- **Analytics**: Tracking and monitoring

## 🔒 Security Best Practices

### Secrets Management

**❌ Never commit secrets to git**

**✅ Use environment-specific secrets:**

```bash
# Development (weak secrets for convenience)
JWT_SECRET=dev-secret-not-secure

# Production (strong generated secrets)
JWT_SECRET=$(openssl rand -base64 32)
```

### Production Security Checklist

- [ ] Generate unique, strong secrets
- [ ] Enable TLS/SSL
- [ ] Disable debug features
- [ ] Enable security headers
- [ ] Use managed database services
- [ ] Configure proper CORS origins
- [ ] Enable rate limiting
- [ ] Set up monitoring and logging

## 🌍 Environment-Specific Configurations

### Development
- **Purpose**: Local development with Docker Compose
- **Security**: Relaxed for convenience
- **Features**: Debug tools enabled
- **Services**: All services in containers

### Staging
- **Purpose**: Pre-production testing
- **Security**: Production-like with test data
- **Features**: Some debug tools for testing
- **Services**: Managed services or production-like setup

### Production
- **Purpose**: Live application
- **Security**: Maximum security settings
- **Features**: Debug tools disabled
- **Services**: Fully managed, redundant services

## 🔧 Configuration Variables Reference

### Required for All Environments

```bash
# Application
APP_NAME=zplus-saas-base
ENVIRONMENT=development|staging|production

# Database URLs
DATABASE_URL=postgres://...
MONGODB_URL=mongodb://...
REDIS_URL=redis://...

# Authentication
JWT_SECRET=your-secret
NEXTAUTH_SECRET=your-secret
KEYCLOAK_URL=http://...
```

### Environment-Specific Overrides

#### Development
```bash
DEBUG=true
LOG_LEVEL=debug
GRAPHQL_PLAYGROUND=true
CORS_ALLOWED_ORIGINS=http://localhost:*
```

#### Production
```bash
DEBUG=false
LOG_LEVEL=warn
GRAPHQL_PLAYGROUND=false
CORS_ALLOWED_ORIGINS=https://yourdomain.com
TLS_ENABLED=true
```

## 🐳 Docker Environment Variables

Docker Compose automatically loads:
1. `.env` file in the same directory
2. Environment variables from the host system

Variables are passed to containers via the `environment` section in `docker-compose.dev.yml`.

## 🔄 Environment Loading Order

### Backend (Go)
1. System environment variables
2. `.env` file in working directory
3. Command-line flags (override all)

### Frontend (Next.js)
1. `.env.local` (loaded by Next.js)
2. `.env.development` / `.env.production`
3. `.env` (fallback)

## 📊 Monitoring Configuration

### Required for Production
```bash
# Metrics
METRICS_ENABLED=true
PROMETHEUS_ENABLED=true

# Logging
LOG_LEVEL=warn
LOG_FORMAT=json

# Tracing
TRACING_ENABLED=true
JAEGER_ENDPOINT=your-jaeger-url

# Error Tracking
SENTRY_DSN=your-sentry-dsn
```

## 🔧 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   - Check `DATABASE_URL` format
   - Verify database service is running
   - Check network connectivity

2. **Authentication Not Working**
   - Verify `JWT_SECRET` is set
   - Check `KEYCLOAK_URL` is accessible
   - Ensure secrets match between services

3. **CORS Errors**
   - Update `CORS_ALLOWED_ORIGINS`
   - Check protocol (http vs https)
   - Verify domain configuration

### Validation Commands

```bash
# Check environment loading
make info

# Validate database connections
make db-status

# Test service health
make health-check

# View current configuration
docker-compose config
```

## 📝 Example Configurations

### Local Development with Docker
```bash
# .env
DATABASE_URL=postgres://postgres:postgres123@postgres:5432/zplus
MONGODB_URL=mongodb://mongo:mongo123@mongodb:27017/zplus_metadata
REDIS_URL=redis://:redis123@redis:6379/0
KEYCLOAK_URL=http://keycloak:8080
```

### Production with AWS
```bash
# .env
DATABASE_URL=postgres://user:pass@prod-rds.amazonaws.com:5432/zplus?sslmode=require
MONGODB_URL=mongodb://user:pass@prod-documentdb.amazonaws.com:27017/zplus?ssl=true
REDIS_URL=rediss://prod-elasticache.amazonaws.com:6380/0
KEYCLOAK_URL=https://auth.yourdomain.com
```

## 🔗 Related Documentation

- [Docker Compose Guide](../deployment/README.md)
- [Database Setup](../database/Database_Management_Guide.md)
- [Security Configuration](../security/README.md)
- [Deployment Guide](../deployment/README.md)

## 💡 Tips

1. **Use environment-specific overrides** instead of maintaining separate files
2. **Document custom variables** in your team's configuration guide
3. **Validate configuration** before deployment using health checks
4. **Rotate secrets regularly** in production environments
5. **Use secret management services** (AWS Secrets Manager, Azure Key Vault) for production
