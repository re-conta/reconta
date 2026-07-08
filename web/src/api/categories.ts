import type { Category, CategoryInput } from "../types/category";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listCategories(): Promise<Category[]> {
  return fetch("/api/categories", { credentials: "include" }).then((res) =>
    parseResponse<Category[]>(res),
  );
}

export function createCategory(input: CategoryInput): Promise<Category> {
  return fetch("/api/categories", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Category>(res));
}

export function updateCategory(id: number, input: CategoryInput): Promise<Category> {
  return fetch(`/api/categories/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Category>(res));
}

export function deleteCategory(id: number): Promise<void> {
  return fetch(`/api/categories/${id}`, { method: "DELETE", credentials: "include" }).then((res) =>
    parseResponse<void>(res),
  );
}
