export type HealthLevel = "otima" | "boa" | "estavel" | "ruim" | "pessima";

export interface HealthScore {
  enabled: boolean;
  hasData: boolean;
  level: HealthLevel | "";
  stars: number;
  income: number;
  expense: number;
  balance: number;
  savingsRate: number;
}

export interface HealthSettings {
  enabled: boolean;
  thresholdOtima: number;
  thresholdBoa: number;
  thresholdEstavel: number;
  thresholdRuim: number;
}

export const healthLevelLabels: Record<HealthLevel, string> = {
  otima: "Ótima",
  boa: "Boa",
  estavel: "Estável",
  ruim: "Ruim",
  pessima: "Péssima",
};
