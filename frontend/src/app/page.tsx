import { Metadata } from 'next';
import HomePage from './home';

export const metadata: Metadata = {
  title: 'Zplus SaaS Platform - Complete Multi-Tenant Solution',
  description: 'Build, deploy, and scale your business with our comprehensive SaaS platform. Advanced multi-tenancy, user management, POS system, and analytics.',
};

export default function Page() {
  return <HomePage />;
}
