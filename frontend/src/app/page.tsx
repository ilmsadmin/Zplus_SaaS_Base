import { Metadata } from 'next';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

export const metadata: Metadata = {
  title: 'Dashboard',
  description: 'Zplus SaaS Platform Dashboard',
};

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800">
      <div className="container mx-auto px-4 py-8">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-4xl font-bold text-gradient">
                Welcome to Zplus SaaS
              </h1>
              <p className="mt-2 text-lg text-gray-600 dark:text-gray-300">
                Multi-tenant SaaS platform with advanced features
              </p>
            </div>
            <Badge variant="success" className="px-3 py-1">
              Production Ready
            </Badge>
          </div>
        </div>

        {/* Status Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-8">
          <Card className="p-6">
            <div className="flex items-center space-x-3">
              <div className="w-3 h-3 bg-success-500 rounded-full animate-pulse"></div>
              <h3 className="text-lg font-semibold">Backend API</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-300 mt-2">
              GraphQL API with multi-tenant support
            </p>
            <Badge variant="success" size="sm" className="mt-3">
              ✅ Operational
            </Badge>
          </Card>

          <Card className="p-6">
            <div className="flex items-center space-x-3">
              <div className="w-3 h-3 bg-success-500 rounded-full animate-pulse"></div>
              <h3 className="text-lg font-semibold">Authentication</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-300 mt-2">
              Keycloak with RBAC support
            </p>
            <Badge variant="success" size="sm" className="mt-3">
              ✅ Operational
            </Badge>
          </Card>

          <Card className="p-6">
            <div className="flex items-center space-x-3">
              <div className="w-3 h-3 bg-success-500 rounded-full animate-pulse"></div>
              <h3 className="text-lg font-semibold">Database</h3>
            </div>
            <p className="text-sm text-gray-600 dark:text-gray-300 mt-2">
              PostgreSQL with tenant isolation
            </p>
            <Badge variant="success" size="sm" className="mt-3">
              ✅ Operational
            </Badge>
          </Card>
        </div>

        {/* Feature Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-8 mb-8">
          <Card className="p-6">
            <h3 className="text-xl font-semibold mb-4">Core Features</h3>
            <ul className="space-y-3">
              {[
                'Multi-tenant Architecture',
                'Role-based Access Control',
                'API Gateway & Routing',
                'Custom Domain Support',
                'File Management System',
                'POS Module',
                'Reporting & Analytics',
              ].map((feature, index) => (
                <li key={index} className="flex items-center space-x-3">
                  <div className="w-2 h-2 bg-brand-500 rounded-full"></div>
                  <span className="text-gray-700 dark:text-gray-300">
                    {feature}
                  </span>
                </li>
              ))}
            </ul>
          </Card>

          <Card className="p-6">
            <h3 className="text-xl font-semibold mb-4">Technical Stack</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <h4 className="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Backend
                </h4>
                <ul className="text-sm space-y-1 text-gray-600 dark:text-gray-400">
                  <li>• Go (Fiber)</li>
                  <li>• GraphQL</li>
                  <li>• PostgreSQL</li>
                  <li>• Redis</li>
                  <li>• MongoDB</li>
                </ul>
              </div>
              <div>
                <h4 className="font-medium text-gray-700 dark:text-gray-300 mb-2">
                  Frontend
                </h4>
                <ul className="text-sm space-y-1 text-gray-600 dark:text-gray-400">
                  <li>• Next.js 14</li>
                  <li>• TypeScript</li>
                  <li>• Tailwind CSS</li>
                  <li>• Apollo Client</li>
                  <li>• React Hook Form</li>
                </ul>
              </div>
            </div>
          </Card>
        </div>

        {/* Action Buttons */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Button size="lg" className="gradient-brand text-white">
            Get Started
          </Button>
          <Button variant="outline" size="lg">
            View Documentation
          </Button>
        </div>

        {/* Footer */}
        <footer className="mt-16 text-center text-gray-600 dark:text-gray-400">
          <p>&copy; 2025 Zplus SaaS. All rights reserved.</p>
        </footer>
      </div>
    </div>
  );
}
