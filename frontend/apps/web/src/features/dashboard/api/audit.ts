import { api } from '../../../lib/api/client';
import { ApiError } from '../../../lib/api/errors';
import type { components } from '../../../lib/api-types';

// Types come from the generated contract — `/audit/events` IS in the OpenAPI
// bundle (operations.listAuditEvents / ListAuditEventsResponse), so we never
// hand-roll the shape (frontend-structure §3 rule 4). Mirrors `library.ts`.
export type AuditEventItem = components['schemas']['AuditEventItem'];
export type ListAuditEventsResponse = components['schemas']['ListAuditEventsResponse'];

function asApiError(error: unknown, fallbackCode: string): ApiError {
  if (error instanceof ApiError) return error;
  const env = error as { code?: string; message?: string; details?: unknown } | null;
  return new ApiError(env?.code ?? fallbackCode, 0, env?.message ?? 'Erro inesperado', env?.details);
}

export async function fetchRecentAuditEvents(limit = 8): Promise<ListAuditEventsResponse> {
  const { data, error } = await api.GET('/audit/events', {
    params: { query: { limit } },
  });

  if (error) throw asApiError(error, 'dashboard.activity_failed');
  if (!data) throw new ApiError('dashboard.empty_response', 0, 'Resposta vazia ao carregar atividade.');
  return data;
}
