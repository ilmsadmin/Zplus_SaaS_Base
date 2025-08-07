# Zplus SaaS Base - Backend

Backend service for the Zplus SaaS platform built with Go, implementing Domain-Driven Design (DDD) pattern with multi-tenant architecture.

## 🏗️ Architecture

This backend follows Domain-Driven Design (DDD) principles with a clean architecture pattern:

```
backend/
├── cmd/api/                 # Application entry point
├── internal/
│   ├── domain/             # Domain layer (entities, repositories)
│   ├── application/        # Application layer (use cases, services)
│   ├── infrastructure/     # Infrastructure layer (database, external services)
│   └── interfaces/         # Interface layer (HTTP handlers, controllers)
├── pkg/                    # Shared packages
│   ├── config/            # Configuration management
│   └── logger/            # Logging utilities
├── graph/                  # GraphQL schema and resolvers
└── database/              # Database migrations and seeds
```

## 🚀 Tech Stack

- **Framework**: [Fiber v2](https://gofiber.io/) - Fast HTTP web framework
- **Database**: PostgreSQL (primary), MongoDB (documents), Redis (cache)
- **ORM**: [GORM](https://gorm.io/) with multi-tenant support
- **GraphQL**: [gqlgen](https://gqlgen.com/) for GraphQL API
- **Logging**: [Zap](https://github.com/uber-go/zap) structured logging
- **Configuration**: Environment-based configuration
- **Authentication**: JWT + Keycloak (planned)

## 🛠️ Features

### Multi-Tenancy
- **Schema-based isolation** for PostgreSQL (tenant_xxx schemas)
- **Database-based isolation** for MongoDB (tenant_xxx databases)  
- **Key prefixing** for Redis (tenant:xxx:key)
- **Automatic tenant detection** from subdomain/custom domain

### Security
- **JWT authentication** with role-based access control
- **Tenant isolation** at database and application level
- **API key authentication** for service-to-service communication
- **Audit logging** for all tenant operations

### APIs
- **GraphQL API** for flexible data queries
- **REST API** for standard operations
- **Health checks** and monitoring endpoints
- **Real-time subscriptions** (planned)

## 🏃‍♂️ Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 14+
- MongoDB 6.0+
- Redis 7.0+
- Make (optional)

### 1. Setup Environment
```bash
# Copy environment file
cp .env.example .env

# Update .env with your database credentials
```

### 2. Install Dependencies
```bash
# Using Make
make install

# Or directly with Go
go mod download
```

### 3. Setup Databases
```bash
# Start databases with Docker Compose (recommended for development)
make dev-db-up

# Or configure your own databases and update .env
```

### 4. Run Database Migrations
```bash
make db-migrate-up
```

### 5. Seed Initial Data
```bash
make db-seed
```

### 6. Start Development Server
```bash
# With hot reload (requires air)
make dev

# Or run directly
make run
```

The server will start on `http://localhost:8080`

## 📖 API Documentation

### Health Check
```
GET /health
```

### GraphQL
```
POST /graphql          # GraphQL endpoint
GET /playground        # GraphQL playground (development only)
```

### REST API v1
```
GET /api/v1/tenant/info           # Get tenant information
GET /api/v1/admin/tenants         # List tenants (admin only)
```

## 🗃️ Database

### Multi-Tenant Strategy

#### PostgreSQL (Primary Data)
- Uses schema-based multi-tenancy
- Each tenant gets a dedicated schema: `tenant_{tenant_id}`
- Shared tables in `public` schema for system-wide data

#### MongoDB (Documents & Metadata)  
- Uses database-based multi-tenancy
- Each tenant gets a dedicated database: `tenant_{tenant_id}`
- Shared collections in main database for system-wide data

#### Redis (Cache & Sessions)
- Uses key prefixing for multi-tenancy
- Format: `tenant:{tenant_id}:{key}`
- Automatic cleanup per tenant

### Migrations
```bash
# Create new migration
make db-migrate-create NAME=your_migration_name

# Run migrations
make db-migrate-up

# Rollback migrations  
make db-migrate-down
```

## 🔧 Development

### Available Make Commands
```bash
make help              # Show all available commands
make install           # Install dependencies
make build             # Build application
make run               # Run application
make dev               # Run with hot reload
make test              # Run tests
make test-coverage     # Run tests with coverage
make lint              # Run linter
make fmt               # Format code
make check             # Run all code quality checks
```

### Code Generation
```bash
# Generate GraphQL code
make gen-gql

# Generate all code
make generate
```

### Testing
```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run integration tests
make test-integration
```

## 🐳 Docker

### Build Docker Image
```bash
make docker-build
```

### Run with Docker
```bash
make docker-run
```

### Development with Docker Compose
```bash
# Start all services
docker-compose -f ../docker-compose.dev.yml up

# Start only databases
make dev-db-up
```

## 📝 Environment Variables

Key environment variables (see `.env.example` for complete list):

```bash
# Application
APP_ENV=development
APP_DEBUG=true
SERVER_PORT=8080

# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=zplus_saas

# Redis  
REDIS_HOST=localhost
REDIS_PORT=6379

# MongoDB
MONGO_URI=mongodb://localhost:27017
MONGO_DATABASE=zplus_saas

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# Logging
LOG_LEVEL=debug
LOG_FORMAT=json
```

## 🔒 Security

- **Input validation** on all endpoints
- **SQL injection protection** via GORM  
- **Tenant isolation** at database level
- **Rate limiting** per tenant (planned)
- **API key rotation** support
- **Audit trail** for all operations

## 📊 Monitoring

- **Health check endpoint**: `/health`
- **Metrics endpoint**: `/metrics` (planned)
- **Structured logging** with correlation IDs
- **Database connection monitoring**
- **Performance metrics** (planned)

## 🚀 Deployment

### Production Build
```bash
make build
```

### Docker Deployment
```bash
# Build production image
docker build -t zplus-saas-backend .

# Run container
docker run -p 8080:8080 --env-file .env zplus-saas-backend
```

### Kubernetes (Planned)
- Helm charts in `/deploy` directory
- Environment-specific configurations
- Horizontal Pod Autoscaling
- Rolling deployments

## 🤝 Contributing

1. Follow the existing code structure and patterns
2. Write tests for new functionality  
3. Update documentation as needed
4. Follow Go best practices and conventions
5. Use conventional commit messages

## 📚 Additional Resources

- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Multi-Tenant Architecture Patterns](https://docs.microsoft.com/en-us/azure/sql-database/saas-tenancy-app-design-patterns)
- [Fiber Framework Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)

---

**Last Updated**: August 7, 2025
**Go Version**: 1.21+
**API Version**: v1.0.0
