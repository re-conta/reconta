export type BillingCycle = "monthly" | "yearly";

// Código do plano gratuito, espelhando billing.PlanFree no backend.
export const PlanFree = "gratuito";

export type PaymentMethod = "pix" | "boleto" | "debit_card" | "credit_card";

// Método usado quando um admin concede um plano manualmente pelo painel,
// sem passar pelo checkout do Mercado Pago.
export type SubscriptionMethod = PaymentMethod | "admin_grant";

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
  paymentMethod: SubscriptionMethod;
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

// GrantSubscriptionInput concede um plano a um usuário pelo painel de admin.
// cycle é obrigatório apenas para planos pagos; conceder o plano gratuito
// ("gratuito") remove a assinatura paga vigente do usuário.
export interface GrantSubscriptionInput {
  planCode: string;
  cycle?: BillingCycle;
}

export interface UpdatePlanInput {
  name: string;
  description: string;
  priceMonthly: number;
  priceYearly: number;
  benefits: string[];
  highlight: boolean;
}

export const paymentMethodLabels: Record<SubscriptionMethod, string> = {
  pix: "PIX",
  boleto: "Boleto",
  debit_card: "Cartão de débito",
  credit_card: "Cartão de crédito",
  admin_grant: "Cortesia concedida por admin",
};

export function formatPrice(value: number): string {
  return value.toLocaleString("pt-BR", { style: "currency", currency: "BRL" });
}
