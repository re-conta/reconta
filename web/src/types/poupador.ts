export type PoupadorKind = "income" | "expense";
export type PoupadorFrequency = "monthly" | "weekly" | "biweekly" | "yearly" | "one-time";
export type PoupadorFuelType = "gasoline" | "diesel";
export type PoupadorDistancePeriod = "daily" | "monthly";

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

export interface PoupadorFuelInput {
  fuelType: PoupadorFuelType;
  fuelPrice: number;
  distance: number;
  distancePeriod: PoupadorDistancePeriod;
  consumption: number;
}

export interface PoupadorSnapshot {
  id: string;
  incomes: PoupadorEntry[];
  expenses: PoupadorEntry[];
  fuel?: PoupadorFuelInput;
  createdAt?: string;
}
