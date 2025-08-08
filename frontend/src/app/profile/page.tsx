'use client';

import { ProfileManagement } from '@/components/auth/profile-management';
import { ProtectedRoute } from '@/components/auth/protected-route';

export default function UserProfilePage() {
  return (
    <ProtectedRoute redirectTo="/login">
      <ProfileManagement
        canEditProfile={true}
        canChangePassword={true}
        showPreferences={true}
      />
    </ProtectedRoute>
  );
}
