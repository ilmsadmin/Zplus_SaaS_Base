'use client';

import { ProtectedRoute } from '@/components/auth/protected-route';
import { useAuth } from '@/contexts/rest-auth-context';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { 
  User, 
  FileText, 
  ShoppingCart, 
  Settings,
  Activity,
  Heart,
  CreditCard,
  Bell
} from 'lucide-react';

export default function UserDashboard() {
  const { logout } = useAuth();

  const handleLogout = () => {
    logout();
  };

  return (
    <ProtectedRoute redirectTo="/login">
      <div className="min-h-screen bg-gray-50">
        {/* Header */}
        <header className="bg-white shadow">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex justify-between items-center py-6">
              <div className="flex items-center">
                <User className="h-8 w-8 text-blue-600 mr-3" />
                <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
              </div>
              <div className="flex items-center space-x-4">
                <Button asChild variant="outline">
                  <Link href="/profile">Profile</Link>
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
              {/* Welcome Section */}
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-2">Welcome back!</h2>
                <p className="text-gray-600">Here's what's happening with your account today.</p>
              </div>

              {/* Stats Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">My Orders</CardTitle>
                    <ShoppingCart className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">12</div>
                    <p className="text-xs text-muted-foreground">3 pending delivery</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">My Files</CardTitle>
                    <FileText className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">24</div>
                    <p className="text-xs text-muted-foreground">2.3 GB used</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Wishlist</CardTitle>
                    <Heart className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">8</div>
                    <p className="text-xs text-muted-foreground">2 on sale</p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">Notifications</CardTitle>
                    <Bell className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">3</div>
                    <p className="text-xs text-muted-foreground">1 unread</p>
                  </CardContent>
                </Card>
              </div>

              {/* Services Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <ShoppingCart className="mr-2 h-5 w-5" />
                      Shopping
                    </CardTitle>
                    <CardDescription>
                      Browse products, manage cart, and track your orders
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/shop">Start Shopping</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <FileText className="mr-2 h-5 w-5" />
                      My Files
                    </CardTitle>
                    <CardDescription>
                      Upload, organize, and share your documents and files
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/files">Manage Files</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <CreditCard className="mr-2 h-5 w-5" />
                      Orders & Billing
                    </CardTitle>
                    <CardDescription>
                      View order history, invoices, and payment methods
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/orders">View Orders</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Heart className="mr-2 h-5 w-5" />
                      Wishlist
                    </CardTitle>
                    <CardDescription>
                      Save your favorite items and get notified when they're on sale
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/wishlist">View Wishlist</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Activity className="mr-2 h-5 w-5" />
                      Activity
                    </CardTitle>
                    <CardDescription>
                      Track your account activity and usage statistics
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/activity">View Activity</Link>
                    </Button>
                  </CardContent>
                </Card>

                <Card className="hover:shadow-lg transition-shadow cursor-pointer">
                  <CardHeader>
                    <CardTitle className="flex items-center">
                      <Settings className="mr-2 h-5 w-5" />
                      Account Settings
                    </CardTitle>
                    <CardDescription>
                      Manage your profile, preferences, and security settings
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <Button asChild className="w-full">
                      <Link href="/profile">Manage Account</Link>
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
