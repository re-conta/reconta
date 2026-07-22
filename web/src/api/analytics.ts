import type {
  AnalyticsOverview,
  DateRange,
  DeviceBreakdown,
  LocationCount,
  PathCount,
  ReferrerCount,
  RecentVisit,
} from "../types/analytics";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

function withRange(range: DateRange, extra?: Record<string, string>) {
  return new URLSearchParams({ from: range.from, to: range.to, ...extra });
}

// trackPageView reporta uma navegação de rota para o backend. Nunca deve
// atrapalhar a navegação: falhas são silenciosamente ignoradas.
export function trackPageView(path: string, referrer: string) {
  fetch("/api/track", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    keepalive: true,
    body: JSON.stringify({ path, referrer }),
  }).catch(() => {});
}

export function getOverview(range: DateRange): Promise<AnalyticsOverview> {
  return fetch(`/api/admin/analytics/overview?${withRange(range)}`, {
    credentials: "include",
  }).then((res) => parseResponse<AnalyticsOverview>(res));
}

export function getTopPages(range: DateRange): Promise<PathCount[]> {
  return fetch(`/api/admin/analytics/pages?${withRange(range)}`, {
    credentials: "include",
  }).then((res) => parseResponse<PathCount[]>(res).then((v) => v ?? []));
}

export function getTopReferrers(range: DateRange): Promise<ReferrerCount[]> {
  return fetch(`/api/admin/analytics/referrers?${withRange(range)}`, {
    credentials: "include",
  }).then((res) => parseResponse<ReferrerCount[]>(res).then((v) => v ?? []));
}

export function getTopLocations(range: DateRange): Promise<LocationCount[]> {
  return fetch(`/api/admin/analytics/locations?${withRange(range)}`, {
    credentials: "include",
  }).then((res) => parseResponse<LocationCount[]>(res).then((v) => v ?? []));
}

export function getDeviceBreakdown(range: DateRange): Promise<DeviceBreakdown> {
  return fetch(`/api/admin/analytics/devices?${withRange(range)}`, {
    credentials: "include",
  }).then((res) => parseResponse<DeviceBreakdown>(res));
}

export function getRecentVisits(range: DateRange): Promise<RecentVisit[]> {
  return fetch(`/api/admin/analytics/visitors?${withRange(range)}`, {
    credentials: "include",
  }).then((res) => parseResponse<RecentVisit[]>(res).then((v) => v ?? []));
}

export function getActiveNow(): Promise<number> {
  return fetch("/api/admin/analytics/active", { credentials: "include" })
    .then((res) => parseResponse<{ active: number }>(res))
    .then((v) => v.active);
}
