import type { Account, AccountInput } from "../types/account";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listAccounts(): Promise<Account[]> {
  return fetch("/api/accounts", { credentials: "include" }).then((res) => parseResponse<Account[]>(res));
}

export function createAccount(input: AccountInput): Promise<Account> {
  return fetch("/api/accounts", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Account>(res));
}

export function updateAccount(id: number, input: AccountInput): Promise<Account> {
  return fetch(`/api/accounts/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Account>(res));
}

export function deleteAccount(id: number): Promise<void> {
  return fetch(`/api/accounts/${id}`, { method: "DELETE", credentials: "include" }).then((res) =>
    parseResponse<void>(res),
  );
}
