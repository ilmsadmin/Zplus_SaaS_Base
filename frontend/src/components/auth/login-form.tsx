'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/rest-auth-context';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
// import { Alert, AlertDescription } from '@/components/ui/alert';
import { Eye, EyeOff, Loader2 } from 'lucide-react';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
  rememberMe: z.boolean().optional()
});

type LoginFormData = z.infer<typeof loginSchema>;

interface LoginFormProps {
  role?: 'system_admin' | 'tenant_admin' | 'user';
  tenantId?: string;
  subdomain?: string;
  title?: string;
  description?: string;
  redirectTo?: string;
  showRegisterLink?: boolean;
  showForgotPassword?: boolean;
}

export function LoginForm({
  role,
  tenantId,
  subdomain,
  title = 'Login',
  description = 'Enter your credentials to access your account',
  redirectTo,
  showRegisterLink = true,
  showForgotPassword = true
}: LoginFormProps) {
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const { login } = useAuth();
  const router = useRouter();

  const {
    register,
    handleSubmit,
    formState: { errors }
  } = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema)
  });

  const onSubmit = async (data: LoginFormData) => {
    setIsLoading(true);
    setError('');

    try {
      const result = await login(data.email, data.password);

      if (result.success) {
        // Redirect after successful login
        if (redirectTo) {
          router.push(redirectTo);
        } else {
          // Default redirects based on role
          if (role === 'system_admin') {
            router.push('/admin/dashboard');
          } else if (role === 'tenant_admin') {
            router.push('/admin/dashboard');
          } else {
            router.push('/dashboard');
          }
        }
      } else {
        setError(result.error || 'Login failed');
      }
    } catch (err: any) {
      setError(err.message || 'An unexpected error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  const getRegisterLink = () => {
    if (role === 'system_admin') return '/admin/register';
    if (role === 'tenant_admin') return '/admin/register';
    return '/register';
  };

  const getForgotPasswordLink = () => {
    if (role === 'system_admin') return '/admin/forgot-password';
    if (role === 'tenant_admin') return '/admin/forgot-password';
    return '/forgot-password';
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <CardTitle className="text-2xl font-bold text-center">{title}</CardTitle>
          <CardDescription className="text-center">
            {description}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            {error && (
              <div className="p-3 text-sm text-red-600 bg-red-50 border border-red-200 rounded-md">
                {error}
              </div>
            )}

            <div className="space-y-2">
              <label htmlFor="email" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Email
              </label>
              <Input
                id="email"
                type="email"
                placeholder="Enter your email"
                error={errors.email?.message}
                {...register('email')}
              />
            </div>

            <div className="space-y-2">
              <label htmlFor="password" className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                Password
              </label>
              <div className="relative">
                <Input
                  id="password"
                  type={showPassword ? 'text' : 'password'}
                  placeholder="Enter your password"
                  error={errors.password?.message}
                  {...register('password')}
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                  onClick={() => setShowPassword(!showPassword)}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4" />
                  ) : (
                    <Eye className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>

            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <input
                  id="rememberMe"
                  type="checkbox"
                  className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  {...register('rememberMe')}
                />
                <label htmlFor="rememberMe" className="text-sm text-gray-600">
                  Remember me
                </label>
              </div>
              {showForgotPassword && (
                <Link
                  href={getForgotPasswordLink()}
                  className="text-sm text-blue-600 hover:text-blue-500"
                >
                  Forgot password?
                </Link>
              )}
            </div>

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isLoading ? 'Signing in...' : 'Sign in'}
            </Button>

            {showRegisterLink && (
              <div className="text-center">
                <span className="text-sm text-gray-600">
                  Don't have an account?{' '}
                  <Link
                    href={getRegisterLink()}
                    className="text-blue-600 hover:text-blue-500 font-medium"
                  >
                    Sign up
                  </Link>
                </span>
              </div>
            )}
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
