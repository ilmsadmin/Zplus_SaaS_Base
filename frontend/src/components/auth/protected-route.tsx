'use client';

import { ReactNode } from 'react';
import { useAuth } from '@/contexts/rest-auth-context';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import { Loader2 } from 'lucide-react';

interface ProtectedRouteProps {
  children: ReactNode;
  requiredRole?: string;
  tenantId?: string;
  redirectTo?: string;
  fallback?: ReactNode;
}

export function ProtectedRoute({
  children,
  requiredRole,
  tenantId,
  redirectTo = '/login',
  fallback
}: ProtectedRouteProps) {
  const { user, loading, hasRole, isAuthenticated } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading) {
      if (!isAuthenticated) {
        router.push(redirectTo);
        return;
      }

      if (requiredRole && !hasRole(requiredRole)) {
        router.push('/unauthorized');
        return;
      }
    }
  }, [loading, isAuthenticated, hasRole, requiredRole, tenantId, redirectTo, router]);

  if (loading) {
    return fallback || (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  if (requiredRole && !hasRole(requiredRole)) {
    return null;
  }

  return <>{children}</>;
}

interface RoleBasedRouteProps {
  children: ReactNode;
  allowedRoles: string[];
  tenantId?: string;
  fallback?: ReactNode;
}

export function RoleBasedRoute({
  children,
  allowedRoles,
  tenantId,
  fallback
}: RoleBasedRouteProps) {
  const { hasRole, loading } = useAuth();

  if (loading) {
    return fallback || (
      <div className="flex items-center justify-center min-h-screen">
        <Loader2 className="h-8 w-8 animate-spin" />
      </div>
    );
  }

  const hasAnyRole = allowedRoles.some(role => hasRole(role));

  if (!hasAnyRole) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-red-600">Access Denied</h1>
          <p className="text-gray-600">You don't have permission to access this resource.</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
}
