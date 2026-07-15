import type { HealthScore, HealthSettings } from "../types/health";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function getFinancialHealth(month: number, year: number): Promise<HealthScore> {
  const params = new URLSearchParams({ month: String(month), year: String(year) });
  return fetch(`/api/financial-health?${params}`, { credentials: "include" }).then((res) =>
    parseResponse<HealthScore>(res),
  );
}

export function getHealthSettings(): Promise<HealthSettings> {
  return fetch("/api/admin/financial-health", { credentials: "include" }).then((res) =>
    parseResponse<HealthSettings>(res),
  );
}

export function updateHealthSettings(settings: HealthSettings): Promise<HealthSettings> {
  return fetch("/api/admin/financial-health", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(settings),
  }).then((res) => parseResponse<HealthSettings>(res));
}
