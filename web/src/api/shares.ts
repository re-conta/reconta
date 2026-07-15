import type { Account } from "../types/account";
import type { Category } from "../types/category";
import type { CreateShareInput, Share } from "../types/share";
import type { Transaction, TransactionFilters, TransactionInput, TransactionListResult } from "../types/transaction";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function createShare(input: CreateShareInput): Promise<Share> {
  return fetch("/api/shares", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Share>(res));
}

export function listSentShares(): Promise<Share[]> {
  return fetch("/api/shares/sent", { credentials: "include" }).then((res) =>
    parseResponse<Share[]>(res),
  );
}

export function listReceivedShares(): Promise<Share[]> {
  return fetch("/api/shares/received", { credentials: "include" }).then((res) =>
    parseResponse<Share[]>(res),
  );
}

export function acceptShare(id: number): Promise<Share> {
  return fetch(`/api/shares/${id}/accept`, { method: "POST", credentials: "include" }).then(
    (res) => parseResponse<Share>(res),
  );
}

export function rejectShare(id: number): Promise<Share> {
  return fetch(`/api/shares/${id}/reject`, { method: "POST", credentials: "include" }).then(
    (res) => parseResponse<Share>(res),
  );
}

export function cancelShare(id: number): Promise<void> {
  return fetch(`/api/shares/${id}`, { method: "DELETE", credentials: "include" }).then((res) =>
    parseResponse<void>(res),
  );
}

export function getShareAccounts(shareId: number): Promise<Account[]> {
  return fetch(`/api/shares/${shareId}/accounts`, { credentials: "include" }).then((res) =>
    parseResponse<Account[]>(res),
  );
}

export function getShareCategories(shareId: number): Promise<Category[]> {
  return fetch(`/api/shares/${shareId}/categories`, { credentials: "include" }).then((res) =>
    parseResponse<Category[]>(res),
  );
}

export function listSharedTransactions(
  shareId: number,
  filters: TransactionFilters = {},
): Promise<TransactionListResult> {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  }
  const query = params.toString();
  return fetch(`/api/shares/${shareId}/transactions${query ? `?${query}` : ""}`, {
    credentials: "include",
  }).then((res) => parseResponse<TransactionListResult>(res));
}

export function createSharedTransaction(
  shareId: number,
  input: TransactionInput,
): Promise<Transaction> {
  return fetch(`/api/shares/${shareId}/transactions`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Transaction>(res));
}

export function updateSharedTransaction(
  shareId: number,
  txId: number,
  input: TransactionInput,
): Promise<Transaction> {
  return fetch(`/api/shares/${shareId}/transactions/${txId}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Transaction>(res));
}

export function deleteSharedTransaction(shareId: number, txId: number): Promise<void> {
  return fetch(`/api/shares/${shareId}/transactions/${txId}`, {
    method: "DELETE",
    credentials: "include",
  }).then((res) => parseResponse<void>(res));
}
