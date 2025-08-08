'use client';

// REST API Client for backend communication
class ApiClient {
  private baseURL: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8090/api/v1';
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
    const tenantId = typeof window !== 'undefined' ? localStorage.getItem('tenant_id') : null;

    const url = `${this.baseURL}${endpoint}`;
    const config: RequestInit = {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token && { Authorization: `Bearer ${token}` }),
        ...(tenantId && { 'X-Tenant-ID': tenantId }),
        ...options.headers,
      },
    };

    try {
      const response = await fetch(url, config);
      
      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
      }

      return await response.json();
    } catch (error) {
      console.error('API request failed:', error);
      throw error;
    }
  }

  // Auth methods
  async login(email: string, password: string) {
    return this.request<{ token: string; user: any }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    });
  }

  // Helper method to extract user ID from token
  private getUserIdFromToken(): string | null {
    const token = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
    if (!token) return null;
    
    console.log('Token:', token);
    
    // Token format: token_{user_id}_{timestamp}
    const parts = token.split('_');
    console.log('Token parts:', parts);
    
    if (parts.length >= 3) {
      // Join back the UUID parts (user_id might contain hyphens)
      const userId = parts.slice(1, -1).join('-');
      console.log('Extracted user ID:', userId);
      return userId;
    }
    return null;
  }

  // Profile methods
  async updateProfile(profileData: {
    email: string;
    username: string;
    first_name: string;
    last_name: string;
    phone?: string;
  }) {
    const userId = this.getUserIdFromToken();
    if (!userId) {
      throw new Error('User not authenticated');
    }

    // Get current user info from localStorage to preserve role and tenant_id
    const userStr = typeof window !== 'undefined' ? localStorage.getItem('user') : null;
    let currentUser = null;
    try {
      currentUser = userStr ? JSON.parse(userStr) : null;
    } catch (e) {
      console.error('Failed to parse user from localStorage:', e);
    }

    const updateData = {
      email: profileData.email,
      username: profileData.username,
      first_name: profileData.first_name,
      last_name: profileData.last_name,
      phone: profileData.phone || '',
      role: currentUser?.role || 'user',
      tenant_id: currentUser?.tenant_id || '',
    };

    console.log('Updating profile with data:', updateData);
    console.log('User ID:', userId);

    const result = await this.request<{ message: string }>(`/users/${userId}`, {
      method: 'PUT',
      body: JSON.stringify(updateData),
    });

    // Return updated user data along with success message
    return {
      ...result,
      updatedUser: {
        id: userId,
        ...updateData,
        status: currentUser?.status || 'active'
      }
    };
  }

  async changePassword(passwords: {
    current_password: string;
    new_password: string;
  }) {
    const userId = this.getUserIdFromToken();
    if (!userId) {
      throw new Error('User not authenticated');
    }

    console.log('Changing password for user:', userId);

    return this.request<{ message: string }>(`/users/${userId}/password`, {
      method: 'PUT',
      body: JSON.stringify({
        current_password: passwords.current_password,
        new_password: passwords.new_password,
      }),
    });
  }

  // User methods
  async getUsers(params?: { limit?: number; offset?: number }) {
    const query = new URLSearchParams();
    if (params?.limit) query.append('limit', params.limit.toString());
    if (params?.offset) query.append('offset', params.offset.toString());
    
    return this.request<{ users: any[] }>(`/users?${query}`);
  }

  async createUser(userData: {
    email: string;
    username: string;
    first_name: string;
    last_name: string;
    password: string;
    role: string;
    tenant_id?: string;
  }) {
    return this.request<{ id: string; message: string }>('/users', {
      method: 'POST',
      body: JSON.stringify(userData),
    });
  }

  // Tenant methods
  async getTenants() {
    return this.request<{ tenants: any[] }>('/tenants');
  }

  // Role methods
  async getRoles(tenantId?: string) {
    const query = tenantId ? `?tenant_id=${tenantId}` : '';
    return this.request<{ roles: any[] }>(`/roles${query}`);
  }

  async createRole(roleData: {
    name: string;
    description: string;
    is_system: boolean;
    tenant_id?: string;
  }) {
    return this.request<{ id: string; message: string }>('/roles', {
      method: 'POST',
      body: JSON.stringify(roleData),
    });
  }

  // Health check
  async healthCheck() {
    return this.request<{ status: string; message: string }>('/health');
  }
}

export const apiClient = new ApiClient();
