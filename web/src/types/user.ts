export type UserRole = "pessoa_fisica" | "pessoa_juridica" | "contador" | "admin" | "super_admin";

export type Permission = "admin_panel" | "manage_users" | "manage_permissions";

export const roleLabels: Record<UserRole, string> = {
  pessoa_fisica: "Pessoa Física",
  pessoa_juridica: "Pessoa Jurídica",
  contador: "Contador / Técnico Contábil",
  admin: "Administrador",
  super_admin: "Super Administrador",
};

export const permissionLabels: Record<Permission, string> = {
  admin_panel: "Acessar painel de admin",
  manage_users: "Gerenciar cargos de usuários",
  manage_permissions: "Editar permissões dos cargos",
};

export interface User {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  cnpj: string;
  avatarUrl: string;
  hasPassword: boolean;
  permissions: Permission[];
  createdAt: string;
}

export function canAccessAdmin(user: User | null): boolean {
  if (!user) return false;
  return user.role === "super_admin" || user.permissions?.includes("admin_panel");
}

export interface CreateUserInput {
  name: string;
  email: string;
  password: string;
  role: UserRole;
  cnpj?: string;
}

export interface UpdateProfileInput {
  name: string;
  email: string;
}

export interface UpdatePasswordInput {
  currentPassword: string;
  newPassword: string;
}

export interface RolePermissions {
  roles: UserRole[];
  available: Permission[];
  permissions: Record<string, Permission[]>;
}
