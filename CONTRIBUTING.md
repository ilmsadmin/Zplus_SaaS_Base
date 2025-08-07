# Contributing to Zplus SaaS Base

Cảm ơn bạn đã quan tâm đến việc đóng góp cho Zplus SaaS Base! Chúng tôi rất hoan nghênh mọi đóng góp từ cộng đồng.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Code Standards](#code-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Issue Guidelines](#issue-guidelines)

## 📜 Code of Conduct

Dự án này tuân thủ [Contributor Covenant Code of Conduct](./CODE_OF_CONDUCT.md). Bằng cách tham gia, bạn đồng ý tuân theo những quy tắc này.

## 🚀 Getting Started

### Prerequisites

- **Go**: 1.21+
- **Node.js**: 18+
- **Docker**: 24+
- **Docker Compose**: 2.0+
- **Git**: 2.30+

### Development Setup

1. **Fork repository**
   ```bash
   # Fork trên GitHub UI, sau đó clone
   git clone https://github.com/YOUR_USERNAME/Zplus_SaaS_Base.git
   cd Zplus_SaaS_Base
   ```

2. **Add upstream remote**
   ```bash
   git remote add upstream https://github.com/ilmsadmin/Zplus_SaaS_Base.git
   ```

3. **Setup development environment**
   ```bash
   # Copy environment files
   cp .env.example .env
   cp backend/.env.example backend/.env
   cp frontend/.env.example frontend/.env
   
   # Start development environment
   make dev-up
   ```

4. **Install dependencies**
   ```bash
   # Backend
   cd backend && go mod download
   
   # Frontend
   cd frontend && npm install
   ```

5. **Run tests để đảm bảo setup thành công**
   ```bash
   make test
   ```

## 🛠️ How to Contribute

### Types of Contributions

1. **Bug Reports**: Báo cáo bugs và issues
2. **Feature Requests**: Đề xuất tính năng mới
3. **Code Contributions**: Sửa bugs, thêm features
4. **Documentation**: Cải thiện documentation
5. **Testing**: Viết tests, cải thiện test coverage

### Workflow

1. **Check existing issues**: Tìm kiếm issues hiện có trước khi tạo mới
2. **Create or assign issue**: Tạo issue mới hoặc được assign issue
3. **Create feature branch**: Tạo branch từ `main`
4. **Develop**: Code và test thay đổi
5. **Create PR**: Tạo Pull Request với mô tả chi tiết
6. **Code Review**: Đợi review và chỉnh sửa nếu cần
7. **Merge**: Sau khi approved, PR sẽ được merge

## 💻 Code Standards

### Backend (Go)

#### Code Style
- Sử dụng `gofmt` để format code
- Tuân thủ [Effective Go](https://golang.org/doc/effective_go.html)
- Sử dụng `golangci-lint` cho static analysis

#### Naming Conventions
```go
// Package names: lowercase, single word
package user

// Functions: PascalCase for exported, camelCase for unexported
func CreateUser() {}
func parseRequest() {}

// Variables: camelCase
var userID string
var isActive bool

// Constants: PascalCase hoặc UPPER_CASE
const DefaultTimeout = 30 * time.Second
const API_VERSION = "v1"

// Interfaces: end with -er
type UserCreator interface {
    CreateUser() error
}
```

#### Project Structure
```
internal/
├── domain/           # Business logic, entities
├── service/          # Application services  
├── repository/       # Data access layer
├── handler/          # HTTP/GraphQL handlers
└── middleware/       # HTTP middleware
```

#### Error Handling
```go
// Wrap errors với context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Sử dụng custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

### Frontend (React/TypeScript)

#### Code Style
- Sử dụng ESLint và Prettier
- TypeScript strict mode
- Functional components với hooks

#### Component Structure
```tsx
// Component naming: PascalCase
interface UserCardProps {
  user: User;
  onEdit?: (user: User) => void;
}

export const UserCard: React.FC<UserCardProps> = ({ user, onEdit }) => {
  // Hooks ở đầu component
  const [isEditing, setIsEditing] = useState(false);
  
  // Event handlers
  const handleEdit = useCallback(() => {
    onEdit?.(user);
  }, [user, onEdit]);
  
  // Render
  return (
    <div className="user-card">
      {/* Component content */}
    </div>
  );
};
```

#### File Organization
```
components/
├── ui/              # Reusable UI components
├── forms/           # Form components
├── layout/          # Layout components
└── features/        # Feature-specific components
```

### Database

#### PostgreSQL
- Sử dụng snake_case cho table và column names
- Always use migrations cho schema changes
- Index naming: `idx_table_column`

#### MongoDB
- Collection names: plural, lowercase
- Field names: camelCase
- Always validate input data

## 🧪 Testing Guidelines

### Backend Testing

#### Unit Tests
- Test coverage minimum: 80%
- Test file naming: `*_test.go`
- Sử dụng table-driven tests

```go
func TestUserService_CreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateUserInput
        want    *User
        wantErr bool
    }{
        {
            name: "valid user",
            input: CreateUserInput{
                Name:  "John Doe",
                Email: "john@example.com",
            },
            want: &User{
                Name:  "John Doe", 
                Email: "john@example.com",
            },
            wantErr: false,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

#### Integration Tests
- Test real database interactions
- Use testcontainers cho database testing
- Test tenant isolation

### Frontend Testing

#### Component Tests
```tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { UserCard } from './UserCard';

describe('UserCard', () => {
  const mockUser = {
    id: '1',
    name: 'John Doe',
    email: 'john@example.com',
  };

  it('renders user information', () => {
    render(<UserCard user={mockUser} />);
    
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('john@example.com')).toBeInTheDocument();
  });
  
  it('calls onEdit when edit button is clicked', () => {
    const onEdit = jest.fn();
    render(<UserCard user={mockUser} onEdit={onEdit} />);
    
    fireEvent.click(screen.getByRole('button', { name: /edit/i }));
    expect(onEdit).toHaveBeenCalledWith(mockUser);
  });
});
```

### E2E Tests
- Sử dụng Playwright
- Test critical user flows
- Test multi-tenant scenarios

## 🔄 Pull Request Process

### Before Creating PR

1. **Update from upstream**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run tests**
   ```bash
   make test
   make lint
   ```

3. **Update documentation** nếu cần thiết

### PR Title và Description

#### PR Title Format
```
type(scope): short description

Examples:
feat(auth): add JWT refresh token mechanism
fix(user): resolve user creation validation issue  
docs(api): update GraphQL schema documentation
```

#### PR Description Template
```markdown
## What
Brief description of what this PR does.

## Why  
Explanation of why this change is needed.

## How
Technical details of how the change is implemented.

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated  
- [ ] Manual testing completed
- [ ] E2E tests passing

## Screenshots (if UI changes)
Before: 
After:

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or breaking changes documented)
```

### Review Process

1. **Automated checks**: CI/CD pipeline phải pass
2. **Code review**: Ít nhất 1 approval từ maintainer
3. **Testing**: Manual testing nếu cần thiết
4. **Documentation**: Kiểm tra documentation updates

## 🐛 Issue Guidelines

### Bug Reports

Sử dụng template sau khi tạo bug report:

```markdown
**Bug Description**
A clear and concise description of the bug.

**To Reproduce**
Steps to reproduce the behavior:
1. Go to '...'
2. Click on '....'
3. Scroll down to '....'
4. See error

**Expected Behavior**
A clear description of what you expected to happen.

**Screenshots**
If applicable, add screenshots to help explain your problem.

**Environment:**
- OS: [e.g. macOS 13.0]
- Browser: [e.g. chrome, safari]
- Version: [e.g. 22]

**Additional Context**
Add any other context about the problem here.
```

### Feature Requests

```markdown
**Feature Description**
A clear and concise description of the feature you'd like to see.

**Problem Statement**
What problem does this feature solve?

**Proposed Solution**
Describe the solution you'd like to see.

**Alternatives Considered**
Describe any alternative solutions you've considered.

**Additional Context**
Add any other context or screenshots about the feature request here.
```

## 🏷️ Labels

Chúng tôi sử dụng labels để organize issues và PRs:

- `bug`: Something isn't working
- `enhancement`: New feature or request
- `documentation`: Improvements or additions to documentation
- `good first issue`: Good for newcomers
- `help wanted`: Extra attention is needed
- `priority/high`: High priority
- `priority/medium`: Medium priority  
- `priority/low`: Low priority

## 🤝 Getting Help

- **Discord**: [Zplus Discord](https://discord.gg/zplus)
- **Email**: dev@zplus.io
- **Documentation**: Check [docs/](./docs/) folder

## 🎉 Recognition

Contributors được recognize trong:
- README.md
- Release notes
- Annual contributor report

Cảm ơn bạn đã đóng góp cho Zplus SaaS Base! 🎉
