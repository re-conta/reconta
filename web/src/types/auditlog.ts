export interface AuditLogEntry {
  id: number;
  userId: number | null;
  userName: string;
  userEmail: string;
  action: string;
  entity: string;
  entityId: number | null;
  details: string;
  ip: string;
  userAgent: string;
  createdAt: string;
}
