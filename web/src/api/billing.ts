import type {
  CancelResult,
  Plan,
  SubscribeInput,
  SubscribeResult,
  SubscriptionInfo,
  SubscriptionPayment,
  UpdatePlanInput,
} from "../types/billing";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listPlans(): Promise<Plan[]> {
  return fetch("/api/plans", { credentials: "include" }).then((res) => parseResponse<Plan[]>(res));
}

export function getSubscription(): Promise<SubscriptionInfo> {
  return fetch("/api/billing/subscription", { credentials: "include" }).then((res) =>
    parseResponse<SubscriptionInfo>(res),
  );
}

export function subscribe(input: SubscribeInput): Promise<SubscribeResult> {
  return fetch("/api/billing/subscribe", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<SubscribeResult>(res));
}

export function getPaymentStatus(id: number): Promise<SubscriptionPayment> {
  return fetch(`/api/billing/payments/${id}`, { credentials: "include" }).then((res) =>
    parseResponse<SubscriptionPayment>(res),
  );
}

export function cancelSubscription(mode: "refund" | "end_of_cycle"): Promise<CancelResult> {
  return fetch("/api/billing/subscription/cancel", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ mode }),
  }).then((res) => parseResponse<CancelResult>(res));
}

export function updatePlan(id: number, input: UpdatePlanInput): Promise<Plan> {
  return fetch(`/api/admin/plans/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<Plan>(res));
}
