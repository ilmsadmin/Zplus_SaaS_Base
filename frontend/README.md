# Zplus SaaS Frontend

Modern, responsive frontend application built with Next.js 14, TypeScript, and Tailwind CSS for the Zplus multi-tenant SaaS platform.

## 🚀 Features

- **Next.js 15** with App Router for optimal performance
- **TypeScript** for type safety and better developer experience
- **Tailwind CSS** for utility-first styling
- **Component Library** with reusable UI components
- **Apollo Client** for GraphQL data management
- **Multi-theme Support** with dark/light mode
- **Responsive Design** optimized for all devices
- **Multi-tenant Architecture** with subdomain routing
- **Real-time Features** with WebSocket support
- **Performance Optimized** with code splitting and lazy loading

## 📁 Project Structure

```
src/
├── app/                    # Next.js 14 App Router
│   ├── layout.tsx         # Root layout
│   ├── page.tsx           # Home page
│   ├── globals.css        # Global styles
│   └── (routes)/          # Route groups
├── components/            # Reusable components
│   ├── ui/               # Base UI components
│   ├── forms/            # Form components
│   ├── layouts/          # Layout components
│   └── providers/        # Context providers
├── lib/                  # Utility libraries
│   ├── utils.ts          # Common utilities
│   ├── api.ts            # API helpers
│   └── auth.ts           # Authentication
├── hooks/                # Custom React hooks
├── types/                # TypeScript definitions
├── styles/              # Global styles
└── graphql/             # GraphQL queries/mutations
```

## 🛠️ Development Setup

### Prerequisites

- Node.js 18.18.0 or higher
- npm 10.0.0 or higher

### Installation

1. **Navigate to frontend directory:**
   ```bash
   cd frontend
   ```

2. **Install dependencies:**
   ```bash
   npm install
   ```

3. **Environment Configuration:**
   ```bash
   cp .env.example .env.local
   # Edit .env.local with your configuration
   ```

4. **Start development server:**
   ```bash
   npm run dev
   ```

5. **Open your browser:**
   ```
   http://localhost:3000
   ```

## 📝 Available Scripts

```bash
# Development
npm run dev              # Start development server
npm run build           # Create production build
npm run start           # Start production server

# Code Quality
npm run lint            # Run ESLint
npm run lint:fix        # Fix ESLint issues
npm run type-check      # TypeScript type checking
npm run format          # Format code with Prettier
npm run format:check    # Check code formatting

# Testing
npm run test            # Run tests
npm run test:watch      # Run tests in watch mode
npm run test:coverage   # Generate coverage report

# Documentation
npm run storybook       # Start Storybook
npm run build-storybook # Build Storybook
```

## 🎨 UI Components

The application uses a comprehensive component library built with:

- **Tailwind CSS** for styling
- **Class Variance Authority** for component variants
- **Headless UI** for accessible components
- **Heroicons** for consistent iconography
- **Framer Motion** for smooth animations

### Example Usage

```tsx
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

export function ExampleComponent() {
  return (
    <Card>
      <CardHeader>
        <CardTitle>Sample Card</CardTitle>
      </CardHeader>
      <CardContent>
        <Button variant="default" size="lg">
          Click me
        </Button>
        <Badge variant="success">Active</Badge>
      </CardContent>
    </Card>
  );
}
```

## 🔧 Configuration

### Tailwind CSS

Custom design system with:
- **Brand Colors**: Primary, secondary, accent colors
- **Semantic Colors**: Success, warning, error states
- **Typography**: Inter font family with custom scales
- **Components**: Pre-built component classes
- **Utilities**: Custom utility classes

### Apollo Client

Configured with:
- **Authentication**: Automatic token management
- **Caching**: Optimized cache policies
- **Error Handling**: Global error management
- **Multi-tenant**: Tenant-aware requests

### Next.js Configuration

Features enabled:
- **Typed Routes**: Type-safe routing
- **Image Optimization**: Automatic image optimization
- **Security Headers**: Enhanced security
- **API Proxying**: Backend API integration

## 🌐 Multi-tenant Support

The frontend supports multiple tenant access patterns:

1. **System Admin**: `admin.zplus.io`
2. **Tenant Admin**: `tenant.zplus.io/admin`
3. **End Users**: `tenant.zplus.io`
4. **Custom Domains**: `custom-domain.com`

### Tenant Detection

```tsx
import { useTenant } from '@/hooks/use-tenant';

export function TenantAwareComponent() {
  const { tenant, isLoading } = useTenant();
  
  if (isLoading) return <div>Loading...</div>;
  
  return (
    <div>
      <h1>Welcome to {tenant.name}</h1>
      <p>Subdomain: {tenant.subdomain}</p>
    </div>
  );
}
```

## 🔐 Authentication

Integration with Keycloak for authentication:

```tsx
import { useAuth } from '@/hooks/use-auth';

export function ProtectedComponent() {
  const { user, isAuthenticated, login, logout } = useAuth();
  
  if (!isAuthenticated) {
    return <button onClick={login}>Login</button>;
  }
  
  return (
    <div>
      <p>Welcome, {user.name}!</p>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

## 📊 State Management

- **Apollo Client**: Server state management
- **React Context**: Global UI state
- **Local Storage**: Persistent client state
- **URL State**: Shareable application state

## 🎯 Performance

### Optimization Features

- **Code Splitting**: Automatic route-based splitting
- **Image Optimization**: Next.js Image component
- **Font Optimization**: Google Fonts optimization
- **Bundle Analysis**: Webpack bundle analyzer
- **Lazy Loading**: Component-level lazy loading

### Performance Monitoring

```bash
# Analyze bundle size
npm run build && npm run analyze

# Lighthouse audit
npm run lighthouse

# Performance profiling
npm run perf
```

## 🧪 Testing

### Testing Stack

- **Jest**: Testing framework
- **React Testing Library**: Component testing
- **Playwright**: E2E testing
- **Storybook**: Component documentation

### Test Examples

```tsx
// Component Test
import { render, screen } from '@testing-library/react';
import { Button } from '@/components/ui/button';

test('renders button with text', () => {
  render(<Button>Click me</Button>);
  expect(screen.getByRole('button')).toHaveTextContent('Click me');
});

// Hook Test
import { renderHook } from '@testing-library/react';
import { useTenant } from '@/hooks/use-tenant';

test('useTenant returns tenant data', () => {
  const { result } = renderHook(() => useTenant());
  expect(result.current.tenant).toBeDefined();
});
```

## 🚀 Deployment

### Production Build

```bash
npm run build
npm run start
```

### Docker Deployment

```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
RUN npm run build
EXPOSE 3000
CMD ["npm", "start"]
```

### Environment Variables

Required for production:

```env
NEXT_PUBLIC_API_URL=https://api.zplus.io
NEXT_PUBLIC_GRAPHQL_URL=https://api.zplus.io/graphql
NEXT_PUBLIC_KEYCLOAK_URL=https://auth.zplus.io
```

## 📚 Documentation

- **Storybook**: Component documentation
- **TypeScript**: Inline type documentation
- **README**: Project documentation
- **API Docs**: GraphQL schema documentation

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

### Code Style

- Use TypeScript for all new code
- Follow ESLint configuration
- Write tests for new features
- Document components in Storybook

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support, please contact the development team or create an issue in the repository.

---

**Built with ❤️ by the Zplus Team**
