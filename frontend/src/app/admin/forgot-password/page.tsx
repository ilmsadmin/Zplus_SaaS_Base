import { ForgotPasswordForm } from '@/components/auth/forgot-password-form';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Reset Admin Password - Zplus',
  description: 'Reset your admin password',
};

export default function AdminForgotPasswordPage() {
  return (
    <ForgotPasswordForm
      role="system_admin"
      mode="request"
    />
  );
}
