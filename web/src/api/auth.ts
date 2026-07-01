import type { User } from "../types/user";
import { ApiError } from "./users";

export interface LoginInput {
  email: string;
  password: string;
}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function login(input: LoginInput): Promise<User> {
  return fetch("/api/auth/login", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<User>(res));
}

export async function logout(): Promise<void> {
  await fetch("/api/auth/logout", { method: "POST", credentials: "include" });
}

export async function fetchCurrentUser(): Promise<User | null> {
  const res = await fetch("/api/auth/me", { credentials: "include" });
  if (res.status === 401) return null;
  return parseResponse<User>(res);
}
