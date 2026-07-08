import type {
  BulkUpdateFields,
  Period,
  Transaction,
  TransactionFilters,
  TransactionInput,
  TransactionListResult,
} from "../types/transaction";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listTransactions(filters: TransactionFilters = {}): Promise<TransactionListResult> {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters)) {
    if (value !== undefined && value !== null && value !== "") {
      params.set(key, String(value));
    }
  }
  const query = params.toString();
  return fetch(`/api/transactions${query ? `?${query}` : ""}`, { credentials: "include" }).then(
    (res) => parseResponse<TransactionListResult>(res),
  );
}

export function listPeriods(): Promise<Period[]> {
  return fetch("/api/transactions/periods", { credentials: "include" }).then((res) =>
    parseResponse<Period[]>(res),
  );
}

export function getTransaction(id: number): Promise<Transaction> {
  return fetch(`/api/transactions/${id}`, { credentials: "include" }).then((res) =>
    parseResponse<Transaction>(res),
  );
}

export function createTransaction(input: TransactionInput): Promise<Transaction> {
  return fetch("/api/transactions", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Transaction>(res));
}

export function updateTransaction(id: number, input: TransactionInput): Promise<Transaction> {
  return fetch(`/api/transactions/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Transaction>(res));
}

export function deleteTransaction(id: number): Promise<void> {
  return fetch(`/api/transactions/${id}`, { method: "DELETE", credentials: "include" }).then(
    (res) => parseResponse<void>(res),
  );
}

export function bulkUpdateTransactions(
  ids: number[],
  fields: BulkUpdateFields,
): Promise<{ updated: number }> {
  return fetch("/api/transactions", {
    method: "PATCH",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ ids, fields }),
  }).then((res) => parseResponse<{ updated: number }>(res));
}

export type BulkDeleteScope = "month" | "year" | "all";

export function bulkDeleteTransactions(
  scope: BulkDeleteScope,
  month?: number,
  year?: number,
): Promise<{ deleted: number }> {
  return fetch("/api/transactions", {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ scope, month, year }),
  }).then((res) => parseResponse<{ deleted: number }>(res));
}

export function autoCategorize(): Promise<{ updated: number; checked: number }> {
  return fetch("/api/transactions/auto-categorize", {
    method: "POST",
    credentials: "include",
  }).then((res) => parseResponse<{ updated: number; checked: number }>(res));
}

export function getOpeningBalance(month: number, year: number): Promise<{ amount: number }> {
  return fetch(`/api/transactions/opening-balance?month=${month}&year=${year}`, {
    credentials: "include",
  }).then((res) => parseResponse<{ amount: number }>(res));
}

export function setOpeningBalance(
  month: number,
  year: number,
  amount: number,
): Promise<{ amount: number }> {
  return fetch("/api/transactions/opening-balance", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ month, year, amount }),
  }).then((res) => parseResponse<{ amount: number }>(res));
}
