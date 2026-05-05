// Stubbed pending frontend rewrite. No legacy notification or stream calls.

export type OperationsStreamSnapshot = {
  pendingNotifications: number;
  pendingApprovals: number;
  documentsInReview: number;
  totalDocuments: number;
  generatedAt: string;
};

export async function listNotifications(_params?: URLSearchParams): Promise<{ items: never[] }> {
  return { items: [] };
}

export async function markNotificationRead(_id: string): Promise<void> {}

export async function listAuditEvents(_params?: URLSearchParams): Promise<{ items: never[] }> {
  return { items: [] };
}

export function subscribeOperationsStream(
  _onSnapshot: (snapshot: OperationsStreamSnapshot) => void,
  _onError?: (error: Event) => void,
): () => void {
  return () => {};
}
