import type { CreateUserInput, User } from "../types/user";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function createUser(input: CreateUserInput): Promise<User> {
  return fetch("/api/users", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(input),
  }).then((res) => parseResponse<User>(res));
}

export function listUsers(): Promise<User[]> {
  return fetch("/api/users").then((res) => parseResponse<User[]>(res));
}
