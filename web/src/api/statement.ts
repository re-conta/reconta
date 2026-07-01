import type { Bank, ConfirmImportRow, ImportPreview, ImportResult } from "../types/statement";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listBanks(): Promise<Bank[]> {
  return fetch("/api/transactions/import/banks", { credentials: "include" }).then((res) => parseResponse<Bank[]>(res));
}

export function previewStatementImport(file: File, bank?: string): Promise<ImportPreview> {
  const formData = new FormData();
  formData.append("file", file);
  if (bank) formData.append("bank", bank);
  return fetch("/api/transactions/import/preview", {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then((res) => parseResponse<ImportPreview>(res));
}

export function confirmStatementImport(
  bank: string,
  accountId: number | null,
  transactions: ConfirmImportRow[],
): Promise<ImportResult> {
  return fetch("/api/transactions/import/confirm", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ bank, accountId, transactions }),
  }).then((res) => parseResponse<ImportResult>(res));
}
