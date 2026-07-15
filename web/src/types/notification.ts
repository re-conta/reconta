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
}

// Opções de antecedência oferecidas na tela de configurações (em minutos).
export const OFFSET_OPTIONS: { value: number; label: string }[] = [
  { value: 4320, label: "3 dias antes" },
  { value: 2880, label: "2 dias antes" },
  { value: 1440, label: "1 dia antes" },
  { value: 720, label: "12 horas antes" },
  { value: 360, label: "6 horas antes" },
  { value: 120, label: "2 horas antes" },
  { value: 60, label: "1 hora antes" },
  { value: 0, label: "No vencimento" },
];
