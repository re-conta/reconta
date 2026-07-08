export type ExportFormat = "xlsx" | "ods" | "pdf" | "json";

export type ReportScopeKind = "month" | "year" | "range" | "all";

export interface ReportScopeParams {
  scope: ReportScopeKind;
  month?: number;
  year?: number;
  dateFrom?: string;
  dateTo?: string;
}

export interface ChartImagePayload {
  title: string;
  pngBase64: string;
}

export interface ImportBackupResult {
  imported: number;
  skipped: number;
  total: number;
}
