import { apiFetch } from '../../../lib/api/client';
import { ApiError } from '../../../lib/api/errors';
import type { components } from '../../../lib/api-types';

export type DocumentListResponse = components['schemas']['DocumentListResponse'];
export type DocumentStatsResponse = components['schemas']['DocumentStatsResponse'];

type DocumentListQueryParams = {
  page?: number;
  pageSize?: number;
  status?: string;
  areaCode?: string;
  profileCode?: string;
  q?: string;
  includeArchived?: boolean;
};

function buildQuery(params: DocumentListQueryParams): string {
  const search = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined) search.set(key, String(value));
  }
  const qs = search.toString();
  return qs ? `?${qs}` : '';
}

// Note: 4xx/5xx already throw `ApiError` via `apiFetch` → `assertApiResponse`.
// This guard only fires if the backend returns 2xx with an error envelope —
// rare, but if it happens we want a real `ApiError` so `resolveErrorMessage`
// works downstream rather than a raw plain object.
function asApiError(error: unknown, fallbackCode: string): ApiError {
  if (error instanceof ApiError) return error;
  const env = error as { code?: string; message?: string; details?: unknown } | null;
  return new ApiError(env?.code ?? fallbackCode, 0, env?.message ?? 'Erro inesperado', env?.details);
}

export async function fetchLibrary(params: DocumentListQueryParams): Promise<DocumentListResponse> {
  try {
    const data = await apiFetch<DocumentListResponse>(`/api/v1/documents${buildQuery(params)}`);
    if (!data) throw new ApiError('library.empty_response', 0, 'Resposta vazia ao listar documentos.');
    return data;
  } catch (error) {
    throw asApiError(error, 'library.list_failed');
  }
}

export async function fetchLibraryStats(): Promise<DocumentStatsResponse> {
  try {
    const data = await apiFetch<DocumentStatsResponse>('/api/v1/documents/stats');
    if (!data) throw new ApiError('library.empty_response', 0, 'Resposta vazia ao carregar estatísticas.');
    return data;
  } catch (error) {
    throw asApiError(error, 'library.stats_failed');
  }
}
