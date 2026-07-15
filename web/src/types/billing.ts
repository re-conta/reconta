export type BillingCycle = "monthly" | "yearly";

export type PaymentMethod = "pix" | "boleto" | "debit_card" | "credit_card";

export type SubscriptionStatus = "pending" | "active" | "canceled" | "expired";

export interface Plan {
  id: number;
  code: string;
  name: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  benefits: string[];
  highlight: boolean;
  sortOrder: number;
}

export interface Subscription {
  id: number;
  userId: number;
  planId: number;
  planCode: string;
  planName: string;
  cycle: BillingCycle;
  status: SubscriptionStatus;
  paymentMethod: PaymentMethod;
  startedAt: string | null;
  currentPeriodEnd: string | null;
  cancelAtPeriodEnd: boolean;
  canceledAt: string | null;
  refundAmount: number | null;
  createdAt: string;
}

export interface SubscriptionInfo {
  planCode: string;
  subscription: Subscription | null;
}

export interface SubscriptionPayment {
  id: number;
  subscriptionId: number;
  userId: number;
  mpPaymentId: number | null;
  amount: number;
  method: PaymentMethod;
  status: string;
  statusDetail: string;
  pixQr?: string;
  pixQrBase64?: string;
  ticketUrl?: string;
  createdAt: string;
}

export interface SubscribeInput {
  planCode: string;
  cycle: BillingCycle;
  method: PaymentMethod;
  token?: string;
  paymentMethodId?: string;
  issuerId?: string;
  installments?: number;
  docType?: "CPF" | "CNPJ";
  docNumber?: string;
  zipCode?: string;
  streetName?: string;
  streetNumber?: string;
  neighborhood?: string;
  city?: string;
  federalUnit?: string;
}

export interface SubscribeResult {
  payment: SubscriptionPayment;
  subscription: Subscription;
}

export interface CancelResult {
  subscription: Subscription;
  refundAmount: number;
}

export interface UpdatePlanInput {
  name: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  benefits: string[];
  highlight: boolean;
}

export const paymentMethodLabels: Record<PaymentMethod, string> = {
  pix: "PIX",
  boleto: "Boleto",
  debit_card: "Cartão de débito",
  credit_card: "Cartão de crédito",
};

export function formatPrice(value: number): string {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}
