import type {
  FixedBill,
  FixedBillInput,
  FixedBillPayment,
  PayFixedBillInput,
  PayFixedBillResult,
} from "../types/fixedBill";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

export function listFixedBills(): Promise<FixedBill[]> {
  return fetch("/api/fixed-bills", { credentials: "include" }).then((res) =>
    parseResponse<FixedBill[]>(res),
  );
}

export function createFixedBill(input: FixedBillInput): Promise<FixedBill> {
  return fetch("/api/fixed-bills", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<FixedBill>(res));
}

export function updateFixedBill(id: number, input: FixedBillInput): Promise<FixedBill> {
  return fetch(`/api/fixed-bills/${id}`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<FixedBill>(res));
}

export function deleteFixedBill(id: number): Promise<void> {
  return fetch(`/api/fixed-bills/${id}`, { method: "DELETE", credentials: "include" }).then((res) =>
    parseResponse<void>(res),
  );
}

function changeStatus(id: number, action: "freeze" | "reactivate" | "close"): Promise<FixedBill> {
  return fetch(`/api/fixed-bills/${id}/${action}`, {
    method: "POST",
    credentials: "include",
  }).then((res) => parseResponse<FixedBill>(res));
}

export const freezeFixedBill = (id: number) => changeStatus(id, "freeze");
export const reactivateFixedBill = (id: number) => changeStatus(id, "reactivate");
export const closeFixedBill = (id: number) => changeStatus(id, "close");

export function payFixedBill(
  id: number,
  input: PayFixedBillInput = {},
): Promise<PayFixedBillResult> {
  return fetch(`/api/fixed-bills/${id}/pay`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify(input),
  }).then((res) => parseResponse<PayFixedBillResult>(res));
}

export function listFixedBillPayments(id: number): Promise<FixedBillPayment[]> {
  return fetch(`/api/fixed-bills/${id}/payments`, { credentials: "include" }).then((res) =>
    parseResponse<FixedBillPayment[]>(res),
  );
}
