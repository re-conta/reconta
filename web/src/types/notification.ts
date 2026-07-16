export type NotificationKind =
  | "bill_due_soon"
  | "bill_overdue"
  | "share_invited"
  | "share_accepted"
  | "share_rejected"
  | "share_cancelled";

export interface Notification {
  id: number;
  fixedBillId: number | null;
  fixedBillName?: string;
  shareId: number | null;
  kind: NotificationKind;
  title: string;
  message: string;
  dueDate: string;
  readAt: string | null;
  createdAt: string;
}

export interface NotificationSettings {
  siteEnabled: boolean;
  emailEnabled: boolean;
  offsets: number[];
  afterOffsets: number[];
}

// Unidades disponíveis para montar um lembrete (dropdown na tela de
// configurações). O valor é o multiplicador em minutos.
export const OFFSET_UNIT_OPTIONS: { value: number; label: string }[] = [
  { value: 60, label: "Hora(s)" },
  { value: 1440, label: "Dia(s)" },
];

// Momento do lembrete em relação ao vencimento.
export const OFFSET_TIMING_OPTIONS: { value: "before" | "after"; label: string }[] = [
  { value: "before", label: "antes" },
  { value: "after", label: "depois" },
];

function formatAmountUnit(minutes: number): string {
  if (minutes % 1440 === 0) {
    const days = minutes / 1440;
    return `${days} dia${days > 1 ? "s" : ""}`;
  }
  if (minutes % 60 === 0) {
    const hours = minutes / 60;
    return `${hours} hora${hours > 1 ? "s" : ""}`;
  }
  return `${minutes} min`;
}

export function formatOffsetLabel(minutes: number): string {
  if (minutes === 0) return "No vencimento";
  return `${formatAmountUnit(minutes)} antes`;
}

export function formatAfterOffsetLabel(minutes: number): string {
  return `${formatAmountUnit(minutes)} depois`;
}
