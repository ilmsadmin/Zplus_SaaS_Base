import { gql } from '@apollo/client';

// Auth Types
export interface LoginInput {
  email: string;
  password: string;
  role?: 'system_admin' | 'tenant_admin' | 'user';
  tenantId?: string;
  subdomain?: string;
}

export interface RegisterInput {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
  phone?: string;
  role: 'system_admin' | 'tenant_admin' | 'user';
  tenantId?: string;
}

export interface ResetPasswordInput {
  email: string;
  role?: 'system_admin' | 'tenant_admin' | 'user';
  tenantId?: string;
}

export interface ChangePasswordInput {
  currentPassword: string;
  newPassword: string;
}

export interface UpdateProfileInput {
  firstName?: string;
  lastName?: string;
  phone?: string;
  avatar?: string;
  preferences?: Record<string, any>;
}

// GraphQL Queries
export const GET_CURRENT_USER = gql`
  query GetCurrentUser {
    getCurrentUser {
      id
      email
      firstName
      lastName
      phone
      avatar
      roles {
        id
        name
        tenantId
        permissions
      }
      preferences
      lastLoginAt
      isActive
      createdAt
    }
  }
`;

export const GET_USER_PROFILE = gql`
  query GetUserProfile($userId: ID!) {
    getUserProfile(userId: $userId) {
      id
      email
      firstName
      lastName
      phone
      avatar
      preferences
      isActive
      createdAt
      updatedAt
    }
  }
`;

// GraphQL Mutations
export const LOGIN_MUTATION = gql`
  mutation Login($input: LoginInput!) {
    login(input: $input) {
      token
      refreshToken
      expiresAt
      user {
        id
        email
        firstName
        lastName
        roles {
          id
          name
          tenantId
          permissions
        }
      }
    }
  }
`;

export const REGISTER_MUTATION = gql`
  mutation Register($input: RegisterInput!) {
    register(input: $input) {
      token
      refreshToken
      expiresAt
      user {
        id
        email
        firstName
        lastName
        roles {
          id
          name
          tenantId
        }
      }
    }
  }
`;

export const LOGOUT_MUTATION = gql`
  mutation Logout {
    logout {
      success
      message
    }
  }
`;

export const REFRESH_TOKEN_MUTATION = gql`
  mutation RefreshToken($refreshToken: String!) {
    refreshToken(refreshToken: $refreshToken) {
      token
      refreshToken
      expiresAt
    }
  }
`;

export const RESET_PASSWORD_MUTATION = gql`
  mutation ResetPassword($input: ResetPasswordInput!) {
    resetPassword(input: $input) {
      success
      message
    }
  }
`;

export const CHANGE_PASSWORD_MUTATION = gql`
  mutation ChangePassword($input: ChangePasswordInput!) {
    changePassword(input: $input) {
      success
      message
    }
  }
`;

export const UPDATE_PROFILE_MUTATION = gql`
  mutation UpdateProfile($input: UpdateProfileInput!) {
    updateProfile(input: $input) {
      id
      email
      firstName
      lastName
      phone
      avatar
      preferences
      updatedAt
    }
  }
`;

export const UPLOAD_AVATAR_MUTATION = gql`
  mutation UploadAvatar($file: Upload!) {
    uploadAvatar(file: $file) {
      id
      url
      fileName
      fileSize
    }
  }
`;

export const VERIFY_EMAIL_MUTATION = gql`
  mutation VerifyEmail($token: String!) {
    verifyEmail(token: $token) {
      success
      message
    }
  }
`;

export const RESEND_VERIFICATION_MUTATION = gql`
  mutation ResendVerification($email: String!) {
    resendVerification(email: $email) {
      success
      message
    }
  }
`;
