'use client';

import { ProtectedRoute } from '@/components/auth/protected-route';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { 
  Users, 
  Building, 
  Globe, 
  Settings, 
  Activity, 
  Shield,
  BarChart3,
  Database
} from 'lucide-react';

export default function SystemAdminDashboard() {
  return (
    <ProtectedRoute requiredRole="system_admin" redirectTo="/admin/login">
      <div className="min-h-screen bg-gray-50">
        {/* Header */}
        <header className="bg-white shadow">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center py-6">
              <div className="flex items-center">
                <Shield className="h-8 w-8 text-blue-600 mr-3" />
                <h1 className="text-3xl font-bold text-gray-900">System Admin</h1>
              </div>
              <div className="flex items-center space-x-4">
                <Button asChild variant="outline">
                  <Link href="/admin/profile">Profile</Link>
                </Button>
                <Button asChild variant="outline">
                  <Link href="/logout">Logout</Link>
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
                    <CardTitle className="text-sm font-medium">Total Tenants</CardTitle>
                    <Building className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">24</div>
                    <p className="text-xs text-muted-foreground">+2 from last month</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Total Users</CardTitle>
                    <Users className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">1,234</div>
                    <p className="text-xs text-muted-foreground">+15% from last month</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Active Domains</CardTitle>
                    <Globe className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">89</div>
                    <p className="text-xs text-muted-foreground">+3 this week</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">System Health</CardTitle>
                    <Activity className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold text-green-600">99.9%</div>
                    <p className="text-xs text-muted-foreground">Uptime this month</p>
                  </CardContent>
                </Card>
              </div>

              {/* Management Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Building className="mr-2 h-5 w-5" />
                      Tenant Management
                    </CardTitle>
                    <CardDescription>
                      Manage tenants, create new organizations, and configure tenant settings
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/tenants">Manage Tenants</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Users className="mr-2 h-5 w-5" />
                      User Management
                    </CardTitle>
                    <CardDescription>
                      View and manage users across all tenants, roles, and permissions
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
                      <Globe className="mr-2 h-5 w-5" />
                      Domain Management
                    </CardTitle>
                    <CardDescription>
                      Configure domains, SSL certificates, and routing for all tenants
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/domains">Manage Domains</Link>
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
                      View system-wide analytics, usage reports, and performance metrics
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/analytics">View Analytics</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Database className="mr-2 h-5 w-5" />
                      System Monitoring
                    </CardTitle>
                    <CardDescription>
                      Monitor system health, performance, and infrastructure status
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/monitoring">System Status</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Settings className="mr-2 h-5 w-5" />
                      System Settings
                    </CardTitle>
                    <CardDescription>
                      Configure global settings, features, and platform configurations
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/admin/settings">System Settings</Link>
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
