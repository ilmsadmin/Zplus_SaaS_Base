'use client';

import { useAuth } from '@/contexts/rest-auth-context';
import { useRouter } from 'next/navigation';
import { useEffect, ReactNode } from 'react';

interface AuthRedirectProps {
  children: ReactNode;
  redirectTo: string;
  allowedRoles?: string[];
  redirectIf?: 'authenticated' | 'unauthenticated';
}

export function AuthRedirect({ 
  children, 
  redirectTo, 
  allowedRoles,
  redirectIf = 'authenticated' 
}: AuthRedirectProps) {
  const { isAuthenticated, loading, user } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (loading) return;

    if (redirectIf === 'authenticated' && isAuthenticated) {
      // Redirect authenticated users away from login pages
      if (allowedRoles && user) {
        // Check if user has allowed role
        if (allowedRoles.includes(user.role)) {
          router.push(redirectTo);
        } else {
          // User doesn't have required role, redirect to appropriate dashboard
          if (user.role === 'system_admin' || user.role === 'tenant_admin') {
            router.push('/admin/dashboard');
          } else {
            router.push('/dashboard');
          }
        }
      } else {
        router.push(redirectTo);
      }
    } else if (redirectIf === 'unauthenticated' && !isAuthenticated) {
      // Redirect unauthenticated users to login
      router.push(redirectTo);
    }
  }, [isAuthenticated, loading, user, router, redirectTo, allowedRoles, redirectIf]);

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  // Don't render children if redirecting
  if (redirectIf === 'authenticated' && isAuthenticated) {
    return null;
  }
  
  if (redirectIf === 'unauthenticated' && !isAuthenticated) {
    return null;
  }

  return <>{children}</>;
}
