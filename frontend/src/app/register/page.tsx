import { RegisterForm } from '@/components/auth/register-form';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Create Account - Zplus',
  description: 'Create your new account',
};

export default function UserRegisterPage() {
  return (
    <RegisterForm
      role="user"
      title="Create Your Account"
      description="Join us today and get started"
      showLoginLink={true}
    />
  );
}
