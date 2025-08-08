import { ForgotPasswordForm } from '@/components/auth/forgot-password-form';
import { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Reset Password - Zplus',
  description: 'Reset your password',
};

export default function ForgotPasswordPage() {
  return (
    <ForgotPasswordForm
      role="user"
      mode="request"
    />
  );
}
