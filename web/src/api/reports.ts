import type {
  ChartImagePayload,
  ExportFormat,
  ImportBackupResult,
  ReportScopeParams,
} from "../types/report";

export class ApiError extends Error {}

async function parseResponse<T>(response: Response): Promise<T> {
  const body = await response.json().catch(() => null);
  if (!response.ok) {
    const message = body?.error ?? "Erro inesperado ao comunicar com o servidor";
    throw new ApiError(message);
  }
  return body as T;
}

function filenameFromDisposition(disposition: string | null, fallback: string): string {
  const match = disposition?.match(/filename="?([^"]+)"?/);
  return match?.[1] ?? fallback;
}

export async function exportReport(
  format: ExportFormat,
  scope: ReportScopeParams,
  charts: ChartImagePayload[],
): Promise<{ blob: Blob; filename: string }> {
  const response = await fetch("/api/reports/export", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ format, ...scope, charts }),
  });

  if (!response.ok) {
    const body = await response.json().catch(() => null);
    throw new ApiError(body?.error ?? "Falha ao gerar o relatório");
  }

  const blob = await response.blob();
  const filename = filenameFromDisposition(
    response.headers.get("Content-Disposition"),
    `relatorio-gastos.${format}`,
  );
  return { blob, filename };
}

export function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

export function importBackup(file: File): Promise<ImportBackupResult> {
  const formData = new FormData();
  formData.append("file", file);
  return fetch("/api/reports/import", {
    method: "POST",
    credentials: "include",
    body: formData,
  }).then((res) => parseResponse<ImportBackupResult>(res));
}
