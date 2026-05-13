import { api } from '../../../lib/api/client';
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

function cleanParams(params: DocumentListQueryParams): DocumentListQueryParams {
  return Object.fromEntries(
    Object.entries(params).filter(([, value]) => value !== undefined),
  ) as DocumentListQueryParams;
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
  const { data, error } = await api.GET('/api/v1/documents', {
    params: {
      query: cleanParams(params),
    },
  });

  if (error) throw asApiError(error, 'library.list_failed');
  if (!data) throw new ApiError('library.empty_response', 0, 'Resposta vazia ao listar documentos.');
  return data;
}

export async function fetchLibraryStats(): Promise<DocumentStatsResponse> {
  const { data, error } = await api.GET('/api/v1/documents/stats');

  if (error) throw asApiError(error, 'library.stats_failed');
  if (!data) throw new ApiError('library.empty_response', 0, 'Resposta vazia ao carregar estatísticas.');
  return data;
}
