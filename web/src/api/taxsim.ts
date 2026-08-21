import type { TaxSimulation } from "../types/taxsim";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function getTaxSimulation(year: number): Promise<TaxSimulation> {
  const params = new URLSearchParams({ year: String(year) });
  return fetch(`/api/tax-simulation?${params}`, { credentials: "include" }).then((res) =>
    parseResponse<TaxSimulation>(res),
  );
}
