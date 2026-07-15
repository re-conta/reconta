export type ShareStatus = "pending" | "accepted" | "rejected" | "cancelled";

export interface Share {
  id: number;
  ownerId: number;
  ownerName: string;
  recipientId: number;
  recipientName: string;
  accountIds: number[];
  accountNames: string[];
  canEdit: boolean;
  includeFuture: boolean;
  periodStart: string | null;
  periodEnd: string | null;
  status: ShareStatus;
  createdAt: string;
  respondedAt: string | null;
}

export interface CreateShareInput {
  recipientEmail: string;
  accountIds: number[];
  canEdit: boolean;
  includeFuture: boolean;
  periodStart: string | null;
  periodEnd: string | null;
}
