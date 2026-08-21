import type { AuditLogEntry } from "../types/auditlog";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function getActionLogs(limit = 200): Promise<AuditLogEntry[]> {
  return fetch(`/api/admin/logs?${new URLSearchParams({ limit: String(limit) })}`, {
    credentials: "include",
  }).then((res) => parseResponse<AuditLogEntry[]>(res).then((v) => v ?? []));
}
