'use client';

import { ProtectedRoute } from '@/components/auth/protected-route';
import { useAuth } from '@/contexts/rest-auth-context';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { 
  Users, 
  Settings, 
  FileText, 
  ShoppingCart, 
  BarChart3,
  Shield,
  Globe,
  CreditCard
} from 'lucide-react';

export default function TenantAdminDashboard() {
  const { logout } = useAuth();

  const handleLogout = () => {
    logout();
  };

  return (
    <ProtectedRoute requiredRole="tenant_admin" redirectTo="/admin/tenant-login">
      <div className="min-h-screen bg-gray-50">
        {/* Header */}
        <header className="bg-white shadow">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center py-6">
              <div className="flex items-center">
                <Shield className="h-8 w-8 text-blue-600 mr-3" />
                <h1 className="text-3xl font-bold text-gray-900">Admin Dashboard</h1>
              </div>
              <div className="flex items-center space-x-4">
                <Button asChild variant="outline">
                  <Link href="/admin/profile">Profile</Link>
                </Button>
                <Button onClick={handleLogout} variant="outline">
                  Logout
                </Button>
              </div>
            </div>
          </div>
        </header>

        {/* Main Content */}
        <main className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
          <div className="px-4 py-6 sm:px-0">
            <div className="border-4 border-dashed border-gray-200 rounded-lg p-6">
              {/* Stats Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Users</CardTitle>
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">45</div>
                    <p className="text-xs text-muted-foreground">+3 from last month</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Monthly Revenue</CardTitle>
                    <CreditCard className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">$12,234</div>
                    <p className="text-xs text-muted-foreground">+15% from last month</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Orders</CardTitle>
                    <ShoppingCart className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">89</div>
                    <p className="text-xs text-muted-foreground">+7 this week</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Active Sessions</CardTitle>
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">12</div>
                    <p className="text-xs text-muted-foreground">Current active users</p>
                  </CardContent>
                </Card>
              </div>

              {/* Management Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Users className="mr-2 h-5 w-5" />
                      User Management
                    </CardTitle>
                    <CardDescription>
                      Manage users, roles, and permissions for your organization
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/users">Manage Users</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <ShoppingCart className="mr-2 h-5 w-5" />
                      POS Management
                    </CardTitle>
                    <CardDescription>
                      Manage products, inventory, orders, and sales operations
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/pos">POS System</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <FileText className="mr-2 h-5 w-5" />
                      File Management
                    </CardTitle>
                    <CardDescription>
                      Upload, organize, and manage files and documents
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/files">Manage Files</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <BarChart3 className="mr-2 h-5 w-5" />
                      Analytics & Reports
                    </CardTitle>
                    <CardDescription>
                      View sales reports, user analytics, and business insights
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/analytics">View Reports</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Globe className="mr-2 h-5 w-5" />
                      Domain Settings
                    </CardTitle>
                    <CardDescription>
                      Configure custom domain and branding settings
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/domain">Domain Settings</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Settings className="mr-2 h-5 w-5" />
                      Organization Settings
                    </CardTitle>
                    <CardDescription>
                      Manage organization settings, preferences, and configurations
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/settings">Settings</Link>
                    </Button>
                  </CardContent>
                </Card>
              </div>
            </div>
          </div>
        </main>
      </div>
    </ProtectedRoute>
  );
}
