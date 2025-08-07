# Development Guide

## Overview

Hướng dẫn này cung cấp thông tin chi tiết về cách phát triển và đóng góp cho Zplus SaaS Base.

## Table of Contents

- [Environment Setup](#environment-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Coding Standards](#coding-standards)
- [Testing Guide](#testing-guide)
- [Debugging](#debugging)
- [Performance Optimization](#performance-optimization)

## Environment Setup

### Prerequisites

```bash
# Check versions
go version        # >= 1.21
node --version    # >= 18
docker --version  # >= 24.0
kubectl version   # >= 1.28
```

### Development Environment

1. **Clone và setup**
   ```bash
   git clone https://github.com/ilmsadmin/Zplus_SaaS_Base.git
   cd Zplus_SaaS_Base
   make dev-setup
   ```

2. **Environment variables**
   ```bash
   # Copy environment files
   cp .env.example .env
   cp backend/.env.example backend/.env
   cp frontend/.env.example frontend/.env
   
   # Edit environment variables
   vim .env
   ```

3. **Start development services**
   ```bash
   # Start all services với Docker Compose
   make dev-up
   
   # Hoặc start individual services
   make start-db      # PostgreSQL, MongoDB, Redis
   make start-auth    # Keycloak
   make start-gateway # Traefik
   ```

4. **Access development URLs**
   ```bash
   # Add hosts entries for local development
   echo "127.0.0.1 admin.localhost" >> /etc/hosts
   echo "127.0.0.1 tenant1.localhost" >> /etc/hosts
   echo "127.0.0.1 tenant2.localhost" >> /etc/hosts
   ```

### IDE Setup

#### VS Code
Recommended extensions:
```json
{
  "recommendations": [
    "golang.go",
    "bradlc.vscode-tailwindcss",
    "esbenp.prettier-vscode",
    "ms-vscode.vscode-typescript-next",
    "graphql.vscode-graphql",
    "ms-kubernetes-tools.vscode-kubernetes-tools"
  ]
}
```

Settings:
```json
{
  "go.toolsManagement.autoUpdate": true,
  "go.useLanguageServer": true,
  "go.lintTool": "golangci-lint",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.organizeImports": true
  }
}
```

## Project Structure

### Backend Structure
```
backend/
├── cmd/
│   ├── api/              # API server entry point
│   ├── admin/            # System admin server
│   ├── worker/           # Background worker
│   └── migrate/          # Database migration tool
├── internal/
│   ├── domain/           # Business logic layer
│   │   ├── user/        # User domain (multi-role)
│   │   ├── tenant/      # Tenant domain với custom domain
│   │   ├── auth/        # Authentication domain
│   │   └── file/        # File domain
│   ├── service/          # Application service layer
│   ├── repository/       # Data access layer
│   ├── handler/          # HTTP/GraphQL handlers
│   │   ├── api/         # Main API handlers
│   │   ├── admin/       # System admin handlers
│   │   └── tenant/      # Tenant-specific handlers
│   └── middleware/       # HTTP middleware
├── pkg/
│   ├── auth/            # Authentication utilities
│   ├── database/        # Database connections
│   ├── domain/          # Domain utilities
│   ├── logger/          # Logging utilities
│   └── validator/       # Input validation
├── schema/              # GraphQL schemas
│   ├── admin/          # System admin schemas
│   ├── tenant/         # Tenant management schemas
│   └── user/           # User schemas
├── migrations/          # Database migrations
└── scripts/            # Development scripts
```

### Frontend Structure
```
frontend/
├── pages/               # Next.js pages (App Router)
│   ├── admin/          # System admin pages
│   │   ├── login.tsx
│   │   └── dashboard.tsx
│   ├── [tenant]/       # Tenant-specific pages
│   │   ├── login.tsx
│   │   ├── dashboard.tsx
│   │   └── admin/      # Tenant admin pages
│   │       ├── login.tsx
│   │       └── dashboard.tsx
│   └── _app.tsx
├── components/
│   ├── ui/             # Reusable UI components
│   ├── forms/          # Form components
│   ├── layout/         # Layout components
│   ├── admin/          # System admin components
│   ├── tenant/         # Tenant-specific components
│   └── features/       # Feature-specific components
├── lib/
│   ├── apollo/         # Apollo Client setup
│   ├── auth/           # Authentication logic
│   ├── tenant/         # Tenant utilities
│   └── utils/          # Utility functions
├── hooks/              # Custom React hooks
├── types/              # TypeScript type definitions
└── styles/             # CSS and styling
```

## Development Workflow

### Feature Development

1. **Create feature branch**
   ```bash
   git checkout -b feature/user-management
   ```

2. **Backend development**
   ```bash
   cd backend
   
   # Generate GraphQL code
   go generate ./...
   
   # Run tests
   go test ./...
   
   # Start development server
   go run cmd/api/main.go
   ```

3. **Frontend development**
   ```bash
   cd frontend
   
   # Generate GraphQL types
   npm run codegen
   
   # Start development server
   npm run dev
   ```

4. **Database changes**
   ```bash
   # Create migration
   make migrate-create name=add_user_table
   
   # Run migrations
   make migrate-up
   
   # Rollback if needed
   make migrate-down
   ```

### Hot Reloading

Backend có hot reload với [Air](https://github.com/cosmtrek/air):
```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run với hot reload
air
```

Frontend automatic reload với Next.js dev server:
```bash
npm run dev
```

## Coding Standards

### Go Code Standards

#### Package Organization
```go
// Internal packages
package user

import (
    "context"
    "fmt"
    
    "github.com/ilmsadmin/zplus/internal/domain"
    "github.com/ilmsadmin/zplus/pkg/logger"
)
```

#### Error Handling
```go
// Wrap errors với context
func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    if err := s.validator.Validate(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    
    return user, nil
}
```

#### Logging
```go
import "github.com/ilmsadmin/zplus/pkg/logger"

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*User, error) {
    log := logger.FromContext(ctx).With(
        "operation", "CreateUser",
        "tenant_id", req.TenantID,
    )
    
    log.Info("creating user", "email", req.Email)
    
    user, err := s.repo.Create(ctx, req)
    if err != nil {
        log.Error("failed to create user", "error", err)
        return nil, err
    }
    
    log.Info("user created successfully", "user_id", user.ID)
    return user, nil
}
```

### TypeScript Code Standards

#### Component Structure
```tsx
import { FC, useState, useCallback } from 'react';

interface UserFormProps {
  onSubmit: (user: CreateUserInput) => Promise<void>;
  loading?: boolean;
}

export const UserForm: FC<UserFormProps> = ({ onSubmit, loading = false }) => {
  const [formData, setFormData] = useState<CreateUserInput>({
    name: '',
    email: '',
  });

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();
    await onSubmit(formData);
  }, [formData, onSubmit]);

  return (
    <form onSubmit={handleSubmit}>
      {/* Form content */}
    </form>
  );
};
```

#### Hook Usage
```tsx
import { useQuery, useMutation } from '@apollo/client';
import { GET_USERS, CREATE_USER } from './queries';

export const useUsers = (tenantId: string) => {
  const { data, loading, error } = useQuery(GET_USERS, {
    variables: { tenantId },
  });

  const [createUser] = useMutation(CREATE_USER, {
    update(cache, { data: { createUser } }) {
      // Update cache
    },
  });

  return {
    users: data?.users ?? [],
    loading,
    error,
    createUser,
  };
};
```

## Testing Guide

### Backend Testing

#### Unit Tests
```go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    
    mockRepo := mock.NewMockUserRepository(ctrl)
    service := NewUserService(mockRepo)
    
    // Test cases
    tests := []struct {
        name    string
        input   CreateUserRequest
        setup   func()
        want    *User
        wantErr bool
    }{
        {
            name: "successful creation",
            input: CreateUserRequest{
                Name:     "John Doe",
                Email:    "john@example.com",
                TenantID: "tenant1",
            },
            setup: func() {
                mockRepo.EXPECT().
                    Create(gomock.Any(), gomock.Any()).
                    Return(&User{ID: "123"}, nil)
            },
            want:    &User{ID: "123"},
            wantErr: false,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            got, err := service.CreateUser(context.Background(), tt.input)
            assert.Equal(t, tt.wantErr, err != nil)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

#### Integration Tests
```go
func TestUserRepository_Integration(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    repo := NewUserRepository(db)
    
    // Test với real database
    user, err := repo.Create(context.Background(), CreateUserRequest{
        Name:     "John Doe",
        Email:    "john@example.com",
        TenantID: "tenant1",
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, user.ID)
}
```

### Frontend Testing

#### Component Tests
```tsx
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { MockedProvider } from '@apollo/client/testing';
import { UserForm } from './UserForm';

const mocks = [
  {
    request: {
      query: CREATE_USER,
      variables: { input: { name: 'John', email: 'john@example.com' } },
    },
    result: {
      data: { createUser: { id: '1', name: 'John', email: 'john@example.com' } },
    },
  },
];

describe('UserForm', () => {
  it('submits form with correct data', async () => {
    const onSubmit = jest.fn();
    
    render(
      <MockedProvider mocks={mocks}>
        <UserForm onSubmit={onSubmit} />
      </MockedProvider>
    );
    
    fireEvent.change(screen.getByLabelText(/name/i), {
      target: { value: 'John' }
    });
    
    fireEvent.change(screen.getByLabelText(/email/i), {
      target: { value: 'john@example.com' }
    });
    
    fireEvent.click(screen.getByRole('button', { name: /submit/i }));
    
    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith({
        name: 'John',
        email: 'john@example.com'
      });
    });
  });
});
```

## Debugging

### Backend Debugging

#### Delve Debugger
```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug API server
dlv debug cmd/api/main.go

# Debug tests
dlv test ./internal/service
```

#### VS Code Debug Configuration
```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Debug API",
      "type": "go", 
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/backend/cmd/api",
      "env": {
        "ENV": "development"
      }
    }
  ]
}
```

### Frontend Debugging

#### Browser DevTools
- React DevTools extension
- Apollo Client DevTools
- Redux DevTools (if using Redux)

#### VS Code Debugging
```json
{
  "name": "Debug Next.js",
  "type": "node",
  "request": "launch",
  "program": "${workspaceFolder}/frontend/node_modules/.bin/next",
  "args": ["dev"],
  "console": "integratedTerminal"
}
```

## Performance Optimization

### Backend Performance

#### Database Optimization
```go
// Use GORM preloading
db.Preload("Profile").Find(&users)

// Use select specific fields
db.Select("id", "name", "email").Find(&users)

// Use pagination
db.Offset(offset).Limit(limit).Find(&users)
```

#### Caching
```go
// Redis caching
func (s *Service) GetUser(ctx context.Context, id string) (*User, error) {
    // Try cache first
    cached, err := s.cache.Get(ctx, "user:"+id)
    if err == nil {
        return cached, nil
    }
    
    // Get from database
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    s.cache.Set(ctx, "user:"+id, user, 5*time.Minute)
    return user, nil
}
```

### Frontend Performance

#### Code Splitting
```tsx
import dynamic from 'next/dynamic';

const DynamicComponent = dynamic(() => import('./HeavyComponent'), {
  loading: () => <p>Loading...</p>,
});
```

#### Apollo Client Optimization
```tsx
const GET_USERS = gql`
  query GetUsers($first: Int, $after: String) {
    users(first: $first, after: $after) {
      edges {
        node {
          id
          name
          email
        }
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`;

// Use pagination
const { data, loading, fetchMore } = useQuery(GET_USERS, {
  variables: { first: 20 },
  notifyOnNetworkStatusChange: true,
});
```

## Common Issues

### Backend Issues

#### GORM Connection Issues
```go
// Solution: Configure connection pool
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
sqlDB, err := db.DB()
sqlDB.SetMaxIdleConns(10)
sqlDB.SetMaxOpenConns(100)
sqlDB.SetConnMaxLifetime(time.Hour)
```

#### Memory Leaks
```go
// Always close resources
defer func() {
    if err := rows.Close(); err != nil {
        log.Error("failed to close rows", "error", err)
    }
}()
```

### Frontend Issues

#### Apollo Client Cache Issues
```tsx
// Reset cache on logout
const [logout] = useMutation(LOGOUT, {
  onCompleted: () => {
    client.resetStore();
  },
});
```

#### Next.js Hydration Issues
```tsx
// Use dynamic imports for client-only components
const ClientOnlyComponent = dynamic(
  () => import('./ClientOnlyComponent'),
  { ssr: false }
);
```

## Development Tools

### Useful Commands
```bash
# Backend
make dev-backend        # Start backend dev server
make test-backend       # Run backend tests
make lint-backend       # Lint backend code
make build-backend      # Build backend

# Frontend  
make dev-frontend       # Start frontend dev server
make test-frontend      # Run frontend tests
make lint-frontend      # Lint frontend code
make build-frontend     # Build frontend

# Database
make migrate-up         # Run migrations
make migrate-down       # Rollback migrations
make seed-data          # Seed test data

# Docker
make docker-build       # Build Docker images
make docker-up          # Start with Docker Compose
make docker-down        # Stop Docker Compose
```

### Monitoring Development
```bash
# Watch logs
make logs-api           # API server logs
make logs-db            # Database logs  
make logs-all           # All service logs

# Health checks
make health-check       # Check all services
curl http://localhost:8080/health
```

## Resources

- [Go Best Practices](https://golang.org/doc/effective_go.html)
- [React Best Practices](https://reactjs.org/docs/thinking-in-react.html)
- [GraphQL Best Practices](https://graphql.org/learn/best-practices/)
- [Next.js Documentation](https://nextjs.org/docs)
