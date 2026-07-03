import { listInbox } from '../../approval/api/approvalApi';
import type { ListInboxResponse } from '../../approval/api/approvalTypes';

// Repoints to the canonical listInbox (features/approval/api/approvalApi.ts) instead
// of hand-building the query string — avoids a second, drifting URL-building copy.
export async function fetchDashboardInbox(): Promise<ListInboxResponse> {
  return listInbox({ limit: 6 });
}
