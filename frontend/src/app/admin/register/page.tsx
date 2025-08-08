import { RegisterForm } from '@/components/auth/register-form';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'System Admin Registration - Zplus',
  description: 'Create a system administrator account',
};

export default function SystemAdminRegisterPage() {
  return (
    <RegisterForm
      role="system_admin"
      title="Create System Admin Account"
      description="Register as a system administrator"
      showLoginLink={true}
    />
  );
}
