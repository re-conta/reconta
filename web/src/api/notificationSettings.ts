import type { NotificationSettings } from "../types/notification";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function getNotificationSettings(): Promise<NotificationSettings> {
  return fetch("/api/notification-settings", { credentials: "include" }).then((res) =>
    parseResponse<NotificationSettings>(res),
  );
}

export function updateNotificationSettings(
  input: NotificationSettings,
): Promise<NotificationSettings> {
  return fetch("/api/notification-settings", {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<NotificationSettings>(res));
}
