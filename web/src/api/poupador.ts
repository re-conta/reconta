import type { PoupadorSnapshot } from "../types/poupador";

export class PoupadorApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    throw new PoupadorApiError(body?.error ?? "Erro ao salvar o resultado");
  }
  return body as T;
}

export function createPoupadorSnapshot(snapshot: Omit<PoupadorSnapshot, "id" | "createdAt">) {
  return fetch("/api/poupador/snapshots", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(snapshot),
  }).then((response) => parseResponse<PoupadorSnapshot>(response));
}

export function getPoupadorSnapshot(id: string) {
  return fetch(`/api/poupador/snapshots/${encodeURIComponent(id)}`).then((response) =>
    parseResponse<PoupadorSnapshot>(response),
  );
}
