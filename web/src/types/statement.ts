import type { TransactionType } from "./transaction";

export interface Bank {
  key: string;
  label: string;
}

export interface ParsedTransaction {
  date: string;
  description: string;
  amount: number;
  type: TransactionType;
  pixBeneficiary?: string | null;
  categoryId?: number | null;
  categoryName?: string | null;
  duplicate: boolean;
}

export interface ImportPreview {
  bank: string;
  bankLabel: string;
  transactions: ParsedTransaction[];
}

export interface ImportResult {
  imported: number;
  total: number;
}

export interface ConfirmImportRow {
  date: string;
  description: string;
  amount: number;
  type: TransactionType;
  categoryId: number | null;
  pixBeneficiary: string | null;
}
