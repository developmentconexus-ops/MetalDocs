import { etagCache } from './etagCache';
import { mutate, type MutateOptions } from './mutationClient';
import { apiFetch, ApiError } from '../../../lib/api';
import type { components } from '../../../lib/api-types';
import type {
  ApprovalInstance,
  CancelRequest,
  CancelResponse,
  ListInboxParams,
  ListInboxResponse,
  ObsoleteRequest,
  ObsoleteResponse,
  PublishRequest,
  PublishResponse,
  SchedulePublishRequest,
  SchedulePublishResponse,
  SignoffRequest,
  SignoffResponse,
  SubmitRequest,
  SubmitResponse,
  SupersedeRequest,
  SupersedeResponse,
} from './approvalTypes';

const BASE = '/api/v1';

async function getJSON<T>(url: string): Promise<{ data: T; etag?: string }> {
  const res = await fetch(url);
  if (!res.ok) {
    throw Object.assign(new Error(`http_${res.status}`), { status: res.status });
  }
  const data = (await res.json()) as T;
  const etag = res.headers.get('ETag') ?? undefined;
  return { data, etag };
}

export async function getInstance(documentId: string): Promise<ApprovalInstance> {
  const { data, etag } = await getJSON<ApprovalInstance>(
    `${BASE}/documents/${documentId}/approval-instance`,
  );
  if (etag) {
    etagCache.set(documentId, etag);
  }
  return data;
}

export async function listInbox(params: ListInboxParams = {}): Promise<ListInboxResponse> {
  const qs = new URLSearchParams();
  if (params.area_code) {
    qs.set('area_code', params.area_code);
  }
  if (params.limit != null) {
    qs.set('limit', String(params.limit));
  }
  if (params.offset != null) {
    qs.set('offset', String(params.offset));
  }
  const url = `${BASE}/approval/inbox${qs.toString() ? `?${qs}` : ''}`;
  const { data } = await getJSON<ListInboxResponse>(url);
  return data;
}

export function submit(
  documentId: string,
  body: SubmitRequest,
  opts?: MutateOptions,
): Promise<SubmitResponse> {
  return mutate('POST', `${BASE}/documents/${documentId}/submit`, body, {
    resourceId: documentId,
    ...opts,
  });
}

export function signoff(
  documentId: string,
  body: SignoffRequest,
  opts?: MutateOptions,
): Promise<SignoffResponse> {
  return mutate('POST', `${BASE}/documents/${documentId}/signoff`, body, {
    resourceId: documentId,
    ...opts,
  });
}

export function publish(
  documentId: string,
  body: PublishRequest,
  opts?: MutateOptions,
): Promise<PublishResponse> {
  return mutate('POST', `${BASE}/documents/${documentId}/publish`, body, {
    resourceId: documentId,
    ...opts,
  });
}

export function schedulePublish(
  documentId: string,
  body: SchedulePublishRequest,
  opts?: MutateOptions,
): Promise<SchedulePublishResponse> {
  return mutate('POST', `${BASE}/documents/${documentId}/schedule-publish`, body, {
    resourceId: documentId,
    ...opts,
  });
}

export function supersede(
  documentId: string,
  body: SupersedeRequest,
  opts?: MutateOptions,
): Promise<SupersedeResponse> {
  return mutate('POST', `${BASE}/documents/${documentId}/supersede`, body, {
    resourceId: documentId,
    ...opts,
  });
}

export function obsolete(
  documentId: string,
  body: ObsoleteRequest,
  opts?: MutateOptions,
): Promise<ObsoleteResponse> {
  return mutate('POST', `${BASE}/documents/${documentId}/obsolete`, body, {
    resourceId: documentId,
    ...opts,
  });
}

export function cancel(
  documentId: string,
  body: CancelRequest,
  opts?: MutateOptions,
): Promise<CancelResponse> {
  return mutate('POST', `${BASE}/documents/${documentId}/cancel`, body, {
    resourceId: documentId,
    ...opts,
  });
}

export type ActiveDocumentContext = components['schemas']['ActiveDocumentResponse'];

export async function getActiveDocumentContext(
  controlledDocumentId: string,
): Promise<ActiveDocumentContext | null> {
  try {
    return await apiFetch<ActiveDocumentContext>(
      `/api/v1/controlled-documents/${encodeURIComponent(controlledDocumentId)}/active-document`,
    );
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      return null;
    }
    throw err;
  }
}
