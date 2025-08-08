'use client';

import { ProfileManagement } from '@/components/auth/profile-management';
import { ProtectedRoute } from '@/components/auth/protected-route';

export default function AdminProfilePage() {
  return (
    <ProtectedRoute requiredRole="system_admin" redirectTo="/admin/login">
      <ProfileManagement
        canEditProfile={true}
        canChangePassword={true}
        showPreferences={true}
      />
    </ProtectedRoute>
  );
}
