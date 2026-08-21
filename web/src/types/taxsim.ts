export interface TaxBracketResult {
  rate: number;
  upTo: number;
  deduction: number;
  taxableInBand: number;
  isIncomeBracket: boolean;
}

export interface TaxSimulation {
  enabled: boolean;
  year: number;
  totalIncome: number;
  estimatedTax: number;
  effectiveRate: number;
  marginalRate: number;
  brackets: TaxBracketResult[];
  source: string;
}
