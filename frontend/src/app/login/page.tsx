'use client';

import { LoginForm } from '@/components/auth/login-form';
import { AuthRedirect } from '@/components/auth/auth-redirect';

export default function UserLoginPage() {
  return (
    <AuthRedirect redirectTo="/dashboard">
      <LoginForm
        role="user"
        title="Welcome Back"
        description="Sign in to your account to continue"
        showRegisterLink={true}
      />
    </AuthRedirect>
  );
}
