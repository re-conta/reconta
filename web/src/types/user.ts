export type UserRole = "user" | "admin" | "super_admin";

export interface User {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  avatarUrl: string;
  hasPassword: boolean;
  createdAt: string;
}

export interface CreateUserInput {
  name: string;
  email: string;
  password: string;
}

export interface UpdateProfileInput {
  name: string;
  email: string;
}

export interface UpdatePasswordInput {
  currentPassword: string;
  newPassword: string;
}
