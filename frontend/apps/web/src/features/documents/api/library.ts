import { api } from '../../../lib/api/client';
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

export async function fetchLibrary(params: DocumentListQueryParams): Promise<DocumentListResponse> {
  const { data, error } = await api.GET('/api/v2/documents', {
    params: {
      query: cleanParams(params),
    },
  });

  if (error) {
    throw error;
  }

  if (!data) {
    throw new Error('Library list response is empty');
  }

  return data;
}

export async function fetchLibraryStats(): Promise<DocumentStatsResponse> {
  const { data, error } = await api.GET('/api/v2/documents/stats');

  if (error) {
    throw error;
  }

  if (!data) {
    throw new Error('Library stats response is empty');
  }

  return data;
}
