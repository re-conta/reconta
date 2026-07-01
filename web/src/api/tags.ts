import type { Tag, TagInput } from "../types/tag";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listTags(): Promise<Tag[]> {
  return fetch("/api/tags", { credentials: "include" }).then((res) => parseResponse<Tag[]>(res));
}

export function createTag(input: TagInput): Promise<Tag> {
  return fetch("/api/tags", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Tag>(res));
}

export function updateTag(id: number, input: TagInput): Promise<Tag> {
  return fetch(`/api/tags/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Tag>(res));
}

export function deleteTag(id: number): Promise<void> {
  return fetch(`/api/tags/${id}`, { method: "DELETE", credentials: "include" }).then((res) =>
    parseResponse<void>(res),
  );
}
