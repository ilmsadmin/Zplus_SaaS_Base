'use client';

import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api-client';
import { toast } from 'react-hot-toast';

interface User {
  id: string;
  email: string;
  username: string;
  first_name: string;
  last_name: string;
  phone?: string;
  role: string;
  tenant_id: string;
  status: string;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<{ success: boolean; error?: string }>;
  logout: () => void;
  updateUser: (userData: User) => void;
  isAuthenticated: boolean;
  hasRole: (role: string) => boolean;
  isSystemAdmin: boolean;
  isTenantAdmin: boolean;
  getUserRole: () => string | null;
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

  const setToken = (token: string) => {
    localStorage.setItem('auth_token', token);
  };

  const clearToken = () => {
    localStorage.removeItem('auth_token');
    localStorage.removeItem('tenant_id');
    localStorage.removeItem('user_data');
  };

  // Load user from stored token
  useEffect(() => {
    const token = getStoredToken();
    if (token) {
      // In a real app, you'd verify the token with the server
      // For now, we'll assume the token is valid and decode user info from it
      try {
        // This is a simple token format: token_userId_timestamp
        const parts = token.split('_');
        if (parts.length === 3 && parts[0] === 'token') {
          // Try to get user info from localStorage (set during login)
          const storedUser = localStorage.getItem('user_data');
          if (storedUser) {
            setUser(JSON.parse(storedUser));
          }
          setLoading(false);
        } else {
          clearToken();
          setLoading(false);
        }
      } catch (error) {
        clearToken();
        setLoading(false);
      }
    } else {
      setLoading(false);
    }
  }, []);

  // Auth functions
  const login = async (email: string, password: string): Promise<{ success: boolean; error?: string }> => {
    try {
      setLoading(true);
      const response = await apiClient.login(email, password);
      
      if (response.token && response.user) {
        setToken(response.token);
        setUser(response.user);
        
        // Store user data for persistence
        localStorage.setItem('user_data', JSON.stringify(response.user));
        
        // Store tenant ID if user has one
        if (response.user.tenant_id) {
          localStorage.setItem('tenant_id', response.user.tenant_id);
        }
        
        toast.success('Login successful!');
        
        // Auto redirect based on user role - removed as login form handles this
        // if (response.user.role === 'system_admin') {
        //   router.push('/admin/dashboard');
        // } else if (response.user.role === 'tenant_admin') {
        //   router.push('/admin/dashboard');
        // } else {
        //   router.push('/dashboard');
        // }
        
        return { success: true };
      } else {
        return { success: false, error: 'Invalid response from server' };
      }
    } catch (error: any) {
      console.error('Login failed:', error);
      const errorMessage = error.message || 'Login failed';
      toast.error(errorMessage);
      return { success: false, error: errorMessage };
    } finally {
      setLoading(false);
    }
  };

  const logout = (): void => {
    try {
      console.log('Logout function called');
      
      // Clear all stored data
      clearToken();
      setUser(null);
      
      console.log('User data cleared');
      
      // Show success message
      toast.success('Logged out successfully');
      
      console.log('Toast shown, redirecting...');
      
      // Redirect to login page
      router.push('/login');
      
      console.log('Redirect completed');
    } catch (error) {
      console.error('Logout error:', error);
      // Still try to clear data even if there's an error
      clearToken();
      setUser(null);
      router.push('/login');
    }
  };

  const updateUser = (userData: User): void => {
    try {
      console.log('Updating user state:', userData);
      setUser(userData);
      
      // Update localStorage as well
      localStorage.setItem('user_data', JSON.stringify(userData));
      
      console.log('User state updated successfully');
    } catch (error) {
      console.error('Failed to update user state:', error);
    }
  };

  // Role checking functions
  const hasRole = (role: string): boolean => {
    return user?.role === role;
  };

  const isSystemAdmin = (): boolean => {
    return user?.role === 'system_admin';
  };

  const isTenantAdmin = (): boolean => {
    return user?.role === 'tenant_admin';
  };

  const getUserRole = (): string | null => {
    return user?.role || null;
  };

  const isAuthenticated = !!user && !!getStoredToken();

  const value: AuthContextType = {
    user,
    loading,
    login,
    logout,
    updateUser,
    isAuthenticated,
    hasRole,
    isSystemAdmin: isSystemAdmin(),
    isTenantAdmin: isTenantAdmin(),
    getUserRole,
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
