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

export type RecommendationKind = "cut" | "invest";

export interface Recommendation {
  kind: RecommendationKind;
  title: string;
  description: string;
  impact?: string;
}

export type RecommendationsStatus = "disabled" | "no_data" | "pending" | "ready";

export interface RecommendationsResponse {
  status: RecommendationsStatus;
  stars?: number;
  savingsRate?: number;
  generatedAt?: string;
  recommendations?: Recommendation[];
}
