export type FixedBillPeriodicity =
  | "weekly"
  | "biweekly"
  | "monthly"
  | "bimonthly"
  | "quarterly"
  | "semiannual"
  | "annual"
  | "biennial";

export type FixedBillStatus = "active" | "frozen" | "closed";

export interface FixedBill {
  id: number;
  name: string;
  amount: number;
  categoryId: number | null;
  categoryName?: string;
  categoryColor?: string;
  accountId: number | null;
  accountName?: string;
  periodicity: FixedBillPeriodicity;
  dueDate: string;
  status: FixedBillStatus;
  notes: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface FixedBillInput {
  name: string;
  amount: number;
  categoryId: number | null;
  accountId: number | null;
  periodicity: FixedBillPeriodicity;
  dueDate: string;
  notes: string | null;
}

export interface PayFixedBillInput {
  bank?: string | null;
  paymentMethod?: string | null;
  paidAt?: string | null;
  amountPaid?: number | null;
  accountId?: number | null;
  notes?: string | null;
}

export interface FixedBillPayment {
  id: number;
  fixedBillId: number;
  dueDate: string;
  paidAt: string;
  amountPaid: number;
  bank: string | null;
  paymentMethod: string | null;
  notes: string | null;
  transactionId: number | null;
  createdAt: string;
}

export interface PayFixedBillResult {
  payment: FixedBillPayment;
  bill: FixedBill;
}

export const PERIODICITY_LABELS: Record<FixedBillPeriodicity, string> = {
  weekly: "Semanal",
  biweekly: "Quinzenal",
  monthly: "Mensal",
  bimonthly: "Bimestral",
  quarterly: "Trimestral",
  semiannual: "Semestral",
  annual: "Anual",
  biennial: "Bienal",
};

export const STATUS_LABELS: Record<FixedBillStatus, string> = {
  active: "Ativa",
  frozen: "Congelada",
  closed: "Encerrada",
};
