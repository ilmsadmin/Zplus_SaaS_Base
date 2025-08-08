'use client';

import { LoginForm } from '@/components/auth/login-form';
import { AuthRedirect } from '@/components/auth/auth-redirect';

export default function SystemAdminLoginPage() {
  return (
    <AuthRedirect redirectTo="/admin/dashboard">
      <LoginForm
        role="system_admin"
        title="System Admin Login"
        description="Access Zplus system administration panel"
        showRegisterLink={false}
      />
    </AuthRedirect>
  );
}
