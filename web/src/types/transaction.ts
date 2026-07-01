import type { Tag } from "./tag";

export type TransactionType = "income" | "expense";

export interface Transaction {
  id: number;
  date: string;
  description: string;
  amount: number;
  type: TransactionType;
  categoryId: number | null;
  categoryName?: string | null;
  categoryColor?: string | null;
  accountId: number | null;
  notes: string | null;
  importedFrom: string | null;
  bank: string | null;
  pixBeneficiary: string | null;
  createdAt: string;
  tags: Tag[];
}

export interface TransactionInput {
  date: string;
  description: string;
  amount: number;
  type: TransactionType;
  categoryId: number | null;
  accountId: number | null;
  notes: string | null;
  tagIds: number[];
}

export interface Totals {
  income: number;
  expense: number;
  balance: number;
  count: number;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
}

export interface TransactionListResult {
  data: Transaction[];
  totals: Totals;
  pagination: Pagination;
}

export interface TransactionFilters {
  month?: number;
  year?: number;
  type?: TransactionType;
  categoryId?: number;
  tagId?: number;
  search?: string;
  page?: number;
  limit?: number;
}

export interface BulkUpdateFields {
  type?: TransactionType;
  categoryId?: number | "_none";
  accountId?: number | "_none";
  date?: string;
}
