import type {
  AdminCreateUserInput,
  AdminUpdateUserInput,
  CreateUserInput,
  PendingSignup,
  Permission,
  RolePermissions,
  UpdatePasswordInput,
  UpdateProfileInput,
  User,
  UserRole,
} from "../types/user";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function createUser(input: CreateUserInput): Promise<PendingSignup> {
  return fetch("/api/users", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  }).then((res) => parseResponse<PendingSignup>(res));
}

export function verifySignupOtp(email: string, code: string): Promise<User> {
  return fetch("/api/users/verify-otp", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email, code }),
  }).then((res) => parseResponse<User>(res));
}

export async function resendSignupOtp(email: string): Promise<void> {
  const response = await fetch("/api/users/resend-otp", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email }),
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(body?.error ?? "Erro inesperado ao comunicar com o servidor");
  }
}

export function listUsers(): Promise<User[]> {
  return fetch("/api/users", { credentials: "include" }).then((res) => parseResponse<User[]>(res));
}

export function updateUserRole(id: number, role: UserRole): Promise<User> {
  return fetch(`/api/users/${id}/role`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ role }),
  }).then((res) => parseResponse<User>(res));
}

export function adminCreateUser(input: AdminCreateUserInput): Promise<User> {
  return fetch("/api/admin/users", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<User>(res));
}

export function adminUpdateUser(id: number, input: AdminUpdateUserInput): Promise<User> {
  return fetch(`/api/admin/users/${id}`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<User>(res));
}

export async function adminDeleteUser(id: number): Promise<void> {
  const response = await fetch(`/api/admin/users/${id}`, {
    method: "DELETE",
    credentials: "include",
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(body?.error ?? "Erro inesperado ao comunicar com o servidor");
  }
}

export function adminSetUserBanned(id: number, banned: boolean): Promise<User> {
  return fetch(`/api/admin/users/${id}/ban`, {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ banned }),
  }).then((res) => parseResponse<User>(res));
}

export function fetchRolePermissions(): Promise<RolePermissions> {
  return fetch("/api/admin/permissions", { credentials: "include" }).then((res) =>
    parseResponse<RolePermissions>(res),
  );
}

export function updateRolePermissions(
  role: UserRole,
  permissions: Permission[],
): Promise<{ permissions: Permission[] }> {
  return fetch(`/api/admin/permissions/${role}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ permissions }),
  }).then((res) => parseResponse<{ permissions: Permission[] }>(res));
}

export function updateProfile(input: UpdateProfileInput): Promise<User> {
  return fetch("/api/users/me", {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<User>(res));
}

export async function updatePassword(input: UpdatePasswordInput): Promise<void> {
  const response = await fetch("/api/users/me/password", {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  });
  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(body?.error ?? "Erro inesperado ao comunicar com o servidor");
  }
}
