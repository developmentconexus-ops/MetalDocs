import { apiFetch } from '../../../lib/api/client';
import type { ListInboxResponse } from '../../approval/api/approvalTypes';

export async function fetchDashboardInbox(): Promise<ListInboxResponse> {
  return apiFetch<ListInboxResponse>('/api/v1/approval/inbox?limit=6');
}
