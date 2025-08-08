import { LoginForm } from '@/components/auth/login-form';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Admin Login - Zplus',
  description: 'Login to your tenant administration panel',
};

export default function TenantAdminLoginPage() {
  return (
    <LoginForm
      role="tenant_admin"
      title="Admin Login"
      description="Access your tenant administration panel"
      showRegisterLink={true}
    />
  );
}
