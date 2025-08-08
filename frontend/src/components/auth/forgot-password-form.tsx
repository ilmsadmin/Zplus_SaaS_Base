'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Link from 'next/link';
import { useMutation } from '@apollo/client';
import { RESET_PASSWORD_MUTATION, CHANGE_PASSWORD_MUTATION } from '@/graphql/auth';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Mail, Loader2, CheckCircle, Eye, EyeOff } from 'lucide-react';

const resetPasswordSchema = z.object({
  email: z.string().email('Invalid email address')
});

const changePasswordSchema = z.object({
  newPassword: z.string()
    .min(8, 'Password must be at least 8 characters')
    .regex(/^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)/, 'Password must contain at least one uppercase letter, one lowercase letter, and one number'),
  confirmPassword: z.string()
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

type ResetPasswordFormData = z.infer<typeof resetPasswordSchema>;
type ChangePasswordFormData = z.infer<typeof changePasswordSchema>;

interface ForgotPasswordFormProps {
  role?: 'system_admin' | 'tenant_admin' | 'user';
  tenantId?: string;
  resetToken?: string;
  mode?: 'request' | 'reset';
}

export function ForgotPasswordForm({
  role,
  tenantId,
  resetToken,
  mode = 'request'
}: ForgotPasswordFormProps) {
  const [error, setError] = useState('');
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  const [resetPasswordMutation] = useMutation(RESET_PASSWORD_MUTATION);
  const [changePasswordMutation] = useMutation(CHANGE_PASSWORD_MUTATION);

  const resetForm = useForm<ResetPasswordFormData>({
    resolver: zodResolver(resetPasswordSchema)
  });

  const changeForm = useForm<ChangePasswordFormData>({
    resolver: zodResolver(changePasswordSchema)
  });

  const onResetSubmit = async (data: ResetPasswordFormData) => {
    setIsLoading(true);
    setError('');

    try {
      const { data: result } = await resetPasswordMutation({
        variables: {
          input: {
            email: data.email,
            role,
            tenantId
          }
        }
      });

      if (result?.resetPassword?.success) {
        setSuccess(true);
        resetForm.reset();
      } else {
        setError(result?.resetPassword?.message || 'Reset request failed');
      }
    } catch (err: any) {
      setError(err.message || 'An unexpected error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  const onChangeSubmit = async (data: ChangePasswordFormData) => {
    setIsLoading(true);
    setError('');

    try {
      const { data: result } = await changePasswordMutation({
        variables: {
          input: {
            currentPassword: resetToken || '', // In real app, this would be handled differently
            newPassword: data.newPassword
          }
        }
      });

      if (result?.changePassword?.success) {
        setSuccess(true);
        changeForm.reset();
      } else {
        setError(result?.changePassword?.message || 'Password change failed');
      }
    } catch (err: any) {
      setError(err.message || 'An unexpected error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  const getLoginLink = () => {
    if (role === 'system_admin') return '/admin/login';
    if (role === 'tenant_admin') return '/admin/login';
    return '/login';
  };

  if (success) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <Card className="w-full max-w-md">
          <CardHeader className="space-y-1 text-center">
            <div className="mx-auto w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mb-4">
              <CheckCircle className="w-8 h-8 text-green-600" />
            </div>
            <CardTitle className="text-2xl font-bold">
              {mode === 'request' ? 'Check Your Email' : 'Password Updated'}
            </CardTitle>
            <CardDescription>
              {mode === 'request' 
                ? 'We\'ve sent a password reset link to your email address.'
                : 'Your password has been successfully updated.'
              }
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="text-center">
              <Link
                href={getLoginLink()}
                className="text-blue-600 hover:text-blue-500 font-medium"
              >
                Return to Login
              </Link>
            </div>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
      <Card className="w-full max-w-md">
        <CardHeader className="space-y-1">
          <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center mb-4">
            <Mail className="w-8 h-8 text-blue-600" />
          </div>
          <CardTitle className="text-2xl font-bold text-center">
            {mode === 'request' ? 'Forgot Password' : 'Reset Password'}
          </CardTitle>
          <CardDescription className="text-center">
            {mode === 'request' 
              ? 'Enter your email address and we\'ll send you a reset link'
              : 'Enter your new password below'
            }
          </CardDescription>
        </CardHeader>
        <CardContent>
          {mode === 'request' ? (
            <form onSubmit={resetForm.handleSubmit(onResetSubmit)} className="space-y-4">
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <div className="space-y-2">
                <label htmlFor="email" className="text-sm font-medium">
                  Email Address
                </label>
                <Input
                  id="email"
                  type="email"
                  placeholder="Enter your email"
                  error={resetForm.formState.errors.email?.message}
                  {...resetForm.register('email')}
                />
              </div>

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {isLoading ? 'Sending...' : 'Send Reset Link'}
              </Button>

              <div className="text-center">
                <Link
                  href={getLoginLink()}
                  className="text-sm text-blue-600 hover:text-blue-500"
                >
                  Back to Login
                </Link>
              </div>
            </form>
          ) : (
            <form onSubmit={changeForm.handleSubmit(onChangeSubmit)} className="space-y-4">
              {error && (
                <Alert variant="destructive">
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              <div className="space-y-2">
                <label htmlFor="newPassword" className="text-sm font-medium">
                  New Password
                </label>
                <div className="relative">
                  <Input
                    id="newPassword"
                    type={showNewPassword ? 'text' : 'password'}
                    placeholder="Enter new password"
                    error={changeForm.formState.errors.newPassword?.message}
                    {...changeForm.register('newPassword')}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowNewPassword(!showNewPassword)}
                  >
                    {showNewPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              <div className="space-y-2">
                <label htmlFor="confirmPassword" className="text-sm font-medium">
                  Confirm New Password
                </label>
                <div className="relative">
                  <Input
                    id="confirmPassword"
                    type={showConfirmPassword ? 'text' : 'password'}
                    placeholder="Confirm new password"
                    error={changeForm.formState.errors.confirmPassword?.message}
                    {...changeForm.register('confirmPassword')}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                    onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                  >
                    {showConfirmPassword ? (
                      <EyeOff className="h-4 w-4" />
                    ) : (
                      <Eye className="h-4 w-4" />
                    )}
                  </Button>
                </div>
              </div>

              <Button type="submit" className="w-full" disabled={isLoading}>
                {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                {isLoading ? 'Updating...' : 'Update Password'}
              </Button>

              <div className="text-center">
                <Link
                  href={getLoginLink()}
                  className="text-sm text-blue-600 hover:text-blue-500"
                >
                  Back to Login
                </Link>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
