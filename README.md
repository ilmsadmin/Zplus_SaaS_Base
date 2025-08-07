# Zplus SaaS Base - Multi-Tenant SaaS Platform

[![Build Status](https://github.com/ilmsadmin/Zplus_SaaS_Base/actions/workflows/ci.yml/badge.svg)](https://github.com/ilmsadmin/Zplus_SaaS_Base/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/ilmsadmin/Zplus_SaaS_Base)](https://golang.org/)

**Zplus_SaaS_Base** là một nền tảng SaaS quy mô lớn, sử dụng kiến trúc **microservices** và mô hình **multi-tenant** để phục vụ nhiều khách hàng (tenant). Mỗi tenant được truy cập qua subdomain (`tenant_name.zplus.io`) hoặc domain riêng (custom domain). Hệ thống cung cấp các màn hình đăng nhập riêng biệt cho **system admin/manager**, **tenant admin/manager**, và **user**, với các module service được quản lý tập trung trong dashboard tenant.

## 🚀 Tính năng chính

- **Multi-tenant Architecture**: Cô lập dữ liệu hoàn toàn giữa các tenant với subdomain và custom domain
- **Role-based Access**: Màn hình đăng nhập riêng cho system admin, tenant admin, và user
- **Modular Dashboard**: Module service tích hợp trong dashboard tenant
- **Custom Domain Support**: Tenant có thể sử dụng domain riêng
- **GraphQL Federation**: API Gateway thống nhất cho tất cả services
- **Real-time**: WebSocket support cho các tính năng real-time
- **Authentication & Authorization**: Keycloak + Casbin RBAC
- **Auto-scaling**: Kubernetes HPA và VPA
- **Multi-database**: PostgreSQL, MongoDB, Redis
- **Observability**: Prometheus, Grafana, Loki

## 🏗️ Kiến trúc hệ thống

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │  API Gateway    │    │   Backend       │
│   (Next.js)     │◄──►│   (Traefik)     │◄──►│   (Go Fiber)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │                        │
                       ┌─────────────────┐    ┌─────────────────┐
                       │   Auth Service  │    │   Databases     │
                       │   (Keycloak)    │    │ PostgreSQL/Mongo│
                       └─────────────────┘    └─────────────────┘
```

## 📋 Yêu cầu hệ thống

### Phát triển (Development)
- **Go**: 1.21+
- **Node.js**: 18+
- **Docker**: 24+
- **Docker Compose**: 2.0+

### Production
- **Kubernetes**: 1.28+
- **PostgreSQL**: 16+
- **MongoDB**: 7+
- **Redis**: 7+

## 🚀 Bắt đầu nhanh

### 1. Clone dự án
```bash
git clone https://github.com/ilmsadmin/Zplus_SaaS_Base.git
cd Zplus_SaaS_Base
```

### 2. Khởi động môi trường development
```bash
# Khởi động tất cả services
make dev-up

# Hoặc sử dụng Docker Compose
docker-compose -f docker-compose.dev.yml up -d
```

### 3. Cài đặt dependencies
```bash
# Backend
cd backend && go mod download

# Frontend  
cd frontend && npm install
```

### 4. Chạy migration và seed data
```bash
make migrate-up
make seed-data
```

### 5. Truy cập ứng dụng
- **System Admin**: http://admin.localhost
- **Tenant App**: http://tenant1.localhost (hoặc custom domain)
- **Tenant Admin**: http://tenant1.localhost/admin
- **API Gateway**: http://localhost:8080
- **GraphQL Playground**: http://localhost:8080/graphql
- **Keycloak Admin**: http://localhost:8081

## 📁 Cấu trúc dự án

```
Zplus_SaaS_Base/
├── backend/              # Go microservices
│   ├── cmd/             # Applications
│   ├── internal/        # Private application code
│   ├── pkg/             # Public libraries
│   └── schema/          # GraphQL schemas
├── frontend/            # Next.js application
│   ├── components/      # React components
│   ├── pages/          # Next.js pages
│   └── lib/            # Utilities and configs
├── infrastructure/      # Kubernetes, Helm charts
│   ├── k8s/            # Kubernetes manifests
│   ├── helm/           # Helm charts
│   └── terraform/      # Infrastructure as Code
├── scripts/            # Automation scripts
├── docs/               # Documentation
└── .github/            # GitHub Actions workflows
```

## 🔧 Development

### Commands có sẵn
```bash
# Development
make dev-up              # Khởi động môi trường dev
make dev-down            # Dừng môi trường dev
make dev-logs            # Xem logs

# Database
make migrate-up          # Chạy migrations
make migrate-down        # Rollback migrations
make seed-data          # Seed dữ liệu mẫu

# Testing
make test               # Chạy tất cả tests
make test-backend       # Test backend only
make test-frontend      # Test frontend only

# Build
make build              # Build tất cả services
make build-backend      # Build backend only
make build-frontend     # Build frontend only

# Deployment
make deploy-staging     # Deploy lên staging
make deploy-prod        # Deploy lên production
```

### Workflow phát triển
1. Tạo feature branch từ `main`
2. Phát triển tính năng mới
3. Viết tests và đảm bảo coverage > 80%
4. Chạy `make lint` và `make test`
5. Tạo Pull Request
6. Code review và merge

## 📚 Documentation

- [System Design](./docs/Zplus_SaaS_Base_System_Design.md) - Thiết kế tổng quan hệ thống
- [API Documentation](./docs/api/README.md) - API reference và examples
- [Deployment Guide](./docs/deployment/README.md) - Hướng dẫn triển khai
- [Development Guide](./docs/development/README.md) - Hướng dẫn phát triển
- [Architecture Decision Records](./docs/adr/README.md) - Quyết định kiến trúc

## 🔐 Bảo mật

- HTTPS enforced với HSTS
- JWT tokens với refresh mechanism
- Rate limiting per tenant
- RBAC với Casbin policies
- Secrets management với Kubernetes secrets
- Regular security audits

## 🚀 Deployment

### Staging
```bash
make deploy-staging
```

### Production
```bash
make deploy-prod
```

Chi tiết về deployment có trong [Deployment Guide](./docs/deployment/README.md).

## 🤝 Contributing

1. Fork dự án
2. Tạo feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Tạo Pull Request

Xem [Contributing Guidelines](./CONTRIBUTING.md) để biết thêm chi tiết.

## 📄 License

Dự án này được phân phối dưới MIT License. Xem `LICENSE` file để biết thêm chi tiết.

## 🆘 Support

- 📧 Email: support@zplus.io
- 💬 Slack: [Zplus Workspace](https://zplus.slack.com)
- 📖 Wiki: [Zplus Wiki](https://wiki.zplus.io)
- 🐛 Issues: [GitHub Issues](https://github.com/ilmsadmin/Zplus_SaaS_Base/issues)

## 👥 Team

- **Tech Lead**: [Your Name](mailto:tech-lead@zplus.io)
- **Backend Lead**: [Backend Lead](mailto:backend-lead@zplus.io)
- **Frontend Lead**: [Frontend Lead](mailto:frontend-lead@zplus.io)
- **DevOps Lead**: [DevOps Lead](mailto:devops-lead@zplus.io)

---

Made with ❤️ by Zplus Team
