export type PoupadorKind = "income" | "expense";
export type PoupadorFrequency = "monthly" | "weekly" | "biweekly" | "yearly" | "one-time";

export interface PoupadorEntry {
  id: string;
  name: string;
  amount: number;
  frequency: PoupadorFrequency;
  month: number;
}

export interface PoupadorEntryDraft {
  name: string;
  amount: number;
  frequency: PoupadorFrequency;
  month: number;
}
