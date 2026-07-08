import type { Notification } from "../types/notification";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listNotifications(): Promise<Notification[]> {
  return fetch("/api/notifications", { credentials: "include" }).then((res) =>
    parseResponse<Notification[]>(res),
  );
}

export function getUnreadCount(): Promise<number> {
  return fetch("/api/notifications/unread-count", { credentials: "include" })
    .then((res) => parseResponse<{ count: number }>(res))
    .then((body) => body.count);
}

export function markNotificationRead(id: number): Promise<void> {
  return fetch(`/api/notifications/${id}/read`, {
    method: "POST",
    credentials: "include",
  }).then((res) => parseResponse<void>(res));
}

export function markAllNotificationsRead(): Promise<void> {
  return fetch("/api/notifications/read-all", {
    method: "POST",
    credentials: "include",
  }).then((res) => parseResponse<void>(res));
}
