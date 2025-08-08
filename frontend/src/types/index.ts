// Base types for the application

export interface User {
  id: string;
  email: string;
  username: string;
  firstName?: string;
  lastName?: string;
  avatar?: string;
  roles: Role[];
  tenantId?: string;
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Role {
  id: string;
  name: string;
  description?: string;
  permissions: Permission[];
}

export interface Permission {
  id: string;
  name: string;
  action: string;
  resource: string;
}

export interface Tenant {
  id: string;
  name: string;
  subdomain: string;
  customDomain?: string;
  settings: TenantSettings;
  status: TenantStatus;
  createdAt: string;
  updatedAt: string;
}

export interface TenantSettings {
  theme: {
    primaryColor: string;
    logoUrl?: string;
    faviconUrl?: string;
  };
  features: {
    enablePOS: boolean;
    enableReports: boolean;
    enableFileManagement: boolean;
  };
  billing: {
    plan: string;
    status: string;
  };
}

export enum TenantStatus {
  ACTIVE = 'ACTIVE',
  INACTIVE = 'INACTIVE',
  SUSPENDED = 'SUSPENDED',
  TRIAL = 'TRIAL',
}

export interface ApiResponse<T> {
  data: T;
  message?: string;
  status: 'success' | 'error';
  errors?: ApiError[];
}

export interface ApiError {
  field?: string;
  message: string;
  code: string;
}

export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
  hasNext: boolean;
  hasPrev: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationInfo;
}

// UI Component Props
export interface BaseComponentProps {
  className?: string;
  children?: React.ReactNode;
}

export interface LoadingState {
  isLoading: boolean;
  error?: string | null;
}

// Form types
export interface FormField {
  name: string;
  label: string;
  type: 'text' | 'email' | 'password' | 'select' | 'textarea' | 'checkbox' | 'radio';
  required?: boolean;
  placeholder?: string;
  options?: Array<{ label: string; value: string }>;
  validation?: {
    min?: number;
    max?: number;
    pattern?: string;
    message?: string;
  };
}

// Navigation types
export interface NavItem {
  label: string;
  href: string;
  icon?: React.ComponentType<any>;
  children?: NavItem[];
  requiresAuth?: boolean;
  roles?: string[];
}

// Theme types
export type ThemeMode = 'light' | 'dark' | 'system';

export interface ThemeConfig {
  mode: ThemeMode;
  primaryColor: string;
  fontFamily: string;
}

// File types
export interface FileUpload {
  id: string;
  name: string;
  size: number;
  type: string;
  url: string;
  thumbnailUrl?: string;
  uploadedAt: string;
}

// Dashboard types
export interface DashboardStats {
  totalUsers: number;
  totalTenants: number;
  totalRevenue: number;
  totalOrders: number;
  growth: {
    users: number;
    tenants: number;
    revenue: number;
    orders: number;
  };
}

// Chart data types
export interface ChartDataPoint {
  label: string;
  value: number;
  color?: string;
}

export interface TimeSeriesData {
  timestamp: string;
  value: number;
}

// Notification types
export interface Notification {
  id: string;
  title: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error';
  read: boolean;
  createdAt: string;
}

// Module specific types
export interface POSProduct {
  id: string;
  name: string;
  description?: string;
  price: number;
  currency: string;
  category: string;
  imageUrl?: string;
  stock: number;
  isActive: boolean;
}

export interface POSOrder {
  id: string;
  orderNumber: string;
  items: POSOrderItem[];
  total: number;
  currency: string;
  status: 'pending' | 'processing' | 'completed' | 'cancelled';
  customerId?: string;
  createdAt: string;
}

export interface POSOrderItem {
  id: string;
  productId: string;
  product: POSProduct;
  quantity: number;
  price: number;
  total: number;
}

// Report types
export interface Report {
  id: string;
  name: string;
  type: 'sales' | 'users' | 'revenue' | 'custom';
  dateRange: {
    start: string;
    end: string;
  };
  data: any;
  generatedAt: string;
}

export interface ReportFilter {
  dateRange?: {
    start: string;
    end: string;
  };
  tenantId?: string;
  userId?: string;
  status?: string;
  category?: string;
}
