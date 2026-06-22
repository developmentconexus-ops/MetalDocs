// Distribution endpoints under /api/v1/documents/{id}/distribution*.
// Types reference generated api-types — never hand-roll a duplicate shape.

import { apiFetch } from '../../../lib/api';
import type { components } from '../../../lib/api-types';

export type DistributionSummaryResponse = components['schemas']['DistributionSummaryResponse'];
export type DistributionRecipient = components['schemas']['DistributionRecipient'];
export type DistributionRecipientsResponse = components['schemas']['DistributionRecipientsResponse'];
export type DistributionAreaCoverage = components['schemas']['DistributionAreaCoverage'];

export async function getDistributionSummary(id: string): Promise<DistributionSummaryResponse> {
  return apiFetch<DistributionSummaryResponse>(`/api/v1/documents/${id}/distribution`);
}

export async function listDistributionRecipients(
  id: string,
  params: { cursor?: string; limit?: number } = {},
): Promise<DistributionRecipientsResponse> {
  const qs = new URLSearchParams();
  if (params.cursor) qs.set('cursor', params.cursor);
  if (params.limit != null) qs.set('limit', String(params.limit));
  const query = qs.toString();
  return apiFetch<DistributionRecipientsResponse>(
    `/api/v1/documents/${id}/distribution/recipients${query ? `?${query}` : ''}`,
  );
}

export async function getDistributionCoverage(id: string): Promise<DistributionAreaCoverage[]> {
  return apiFetch<DistributionAreaCoverage[]>(`/api/v1/documents/${id}/distribution/coverage`);
}
