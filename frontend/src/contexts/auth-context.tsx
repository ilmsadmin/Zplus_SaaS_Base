'use client';

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { useMutation, useQuery } from '@apollo/client';
import { 
  LOGIN_MUTATION, 
  LOGOUT_MUTATION, 
  GET_CURRENT_USER,
  REFRESH_TOKEN_MUTATION 
} from '@/graphql/auth';
import { LoginInput } from '@/graphql/auth';

interface User {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  phone?: string;
  avatar?: string;
  roles: Array<{
    id: string;
    name: string;
    tenantId?: string;
    permissions: string[];
  }>;
  preferences?: Record<string, any>;
  lastLoginAt?: string;
  isActive: boolean;
  createdAt: string;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (input: LoginInput) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<boolean>;
  isAuthenticated: boolean;
  hasRole: (role: string, tenantId?: string) => boolean;
  hasPermission: (permission: string, tenantId?: string) => boolean;
  getUserRole: (tenantId?: string) => string | null;
  isSystemAdmin: boolean;
  isTenantAdmin: (tenantId?: string) => boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  // Token management
  const getStoredToken = (): string | null => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('auth_token');
  };

  const getStoredRefreshToken = (): string | null => {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem('refresh_token');
  };

  const setTokens = (token: string, refreshToken: string, expiresAt: string) => {
    localStorage.setItem('auth_token', token);
    localStorage.setItem('refresh_token', refreshToken);
    localStorage.setItem('token_expires_at', expiresAt);
  };

  const clearTokens = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('token_expires_at');
  };

  const isTokenExpired = (): boolean => {
    const expiresAt = localStorage.getItem('token_expires_at');
    if (!expiresAt) return true;
    return new Date().getTime() > new Date(expiresAt).getTime();
  };

  // GraphQL operations
  const [loginMutation] = useMutation(LOGIN_MUTATION);
  const [logoutMutation] = useMutation(LOGOUT_MUTATION);
  const [refreshTokenMutation] = useMutation(REFRESH_TOKEN_MUTATION);
  
  const { data: currentUserData, loading: currentUserLoading, refetch: refetchUser } = useQuery(
    GET_CURRENT_USER,
    {
      skip: !getStoredToken(),
      errorPolicy: 'ignore'
    }
  );

  // Auth functions
  const login = async (input: LoginInput): Promise<{ success: boolean; error?: string }> => {
    try {
      const { data } = await loginMutation({
        variables: { input }
      });

      if (data?.login) {
        const { token, refreshToken, expiresAt, user: userData } = data.login;
        setTokens(token, refreshToken, expiresAt);
        setUser(userData);
        
        // Redirect based on role
        const userRole = userData.roles[0]?.name;
        if (userRole === 'system_admin') {
          router.push('/admin');
        } else if (userRole === 'tenant_admin') {
          router.push('/admin/dashboard');
        } else {
          router.push('/dashboard');
        }
        
        return { success: true };
      }
    } catch (error: any) {
      console.error('Login error:', error);
      return { 
        success: false, 
        error: error.message || 'Login failed' 
      };
    }
    
    return { success: false, error: 'Invalid credentials' };
  };

  const logout = async (): Promise<void> => {
    try {
      await logoutMutation();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      clearTokens();
      setUser(null);
      router.push('/');
    }
  };

  const refreshToken = async (): Promise<boolean> => {
    const refreshTokenValue = getStoredRefreshToken();
    if (!refreshTokenValue) return false;

    try {
      const { data } = await refreshTokenMutation({
        variables: { refreshToken: refreshTokenValue }
      });

      if (data?.refreshToken) {
        const { token, refreshToken: newRefreshToken, expiresAt } = data.refreshToken;
        setTokens(token, newRefreshToken, expiresAt);
        return true;
      }
    } catch (error) {
      console.error('Token refresh error:', error);
      clearTokens();
      setUser(null);
    }

    return false;
  };

  // Role and permission checks
  const hasRole = (role: string, tenantId?: string): boolean => {
    if (!user) return false;
    return user.roles.some(userRole => 
      userRole.name === role && 
      (!tenantId || userRole.tenantId === tenantId)
    );
  };

  const hasPermission = (permission: string, tenantId?: string): boolean => {
    if (!user) return false;
    return user.roles.some(role => 
      role.permissions.includes(permission) &&
      (!tenantId || role.tenantId === tenantId)
    );
  };

  const getUserRole = (tenantId?: string): string | null => {
    if (!user) return null;
    const role = user.roles.find(role => 
      !tenantId || role.tenantId === tenantId
    );
    return role?.name || null;
  };

  const isSystemAdmin = hasRole('system_admin');
  const isTenantAdmin = (tenantId?: string) => hasRole('tenant_admin', tenantId);

  // Auto-refresh token
  useEffect(() => {
    const checkAndRefreshToken = async () => {
      const token = getStoredToken();
      if (token && isTokenExpired()) {
        const refreshed = await refreshToken();
        if (!refreshed) {
          clearTokens();
          setUser(null);
        }
      }
    };

    const interval = setInterval(checkAndRefreshToken, 5 * 60 * 1000); // Check every 5 minutes
    return () => clearInterval(interval);
  }, []);

  // Set user from GraphQL query
  useEffect(() => {
    if (currentUserData?.getCurrentUser) {
      setUser(currentUserData.getCurrentUser);
    }
    setLoading(currentUserLoading);
  }, [currentUserData, currentUserLoading]);

  // Initial auth check
  useEffect(() => {
    const token = getStoredToken();
    if (token && !isTokenExpired()) {
      refetchUser();
    } else {
      setLoading(false);
    }
  }, []);

  const value: AuthContextType = {
    user,
    loading,
    login,
    logout,
    refreshToken,
    isAuthenticated: !!user,
    hasRole,
    hasPermission,
    getUserRole,
    isSystemAdmin,
    isTenantAdmin
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

// Helper hooks
export function useRequireAuth(redirectTo: string = '/login') {
  const { isAuthenticated, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !isAuthenticated) {
      router.push(redirectTo);
    }
  }, [isAuthenticated, loading, redirectTo, router]);

  return { isAuthenticated, loading };
}

export function useRequireRole(role: string, tenantId?: string, redirectTo: string = '/unauthorized') {
  const { hasRole, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !hasRole(role, tenantId)) {
      router.push(redirectTo);
    }
  }, [hasRole, loading, role, tenantId, redirectTo, router]);

  return { hasRole: hasRole(role, tenantId), loading };
}
