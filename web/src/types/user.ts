export type UserRole = "user" | "admin" | "super_admin";

export interface User {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  createdAt: string;
}

export interface CreateUserInput {
  name: string;
  email: string;
  password: string;
}
