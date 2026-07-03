import { etagCache } from './etagCache';
import { mutate, type MutateOptions } from './mutationClient';
import { apiFetch, requestRaw } from '../../../lib/api';
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

export async function getInstance(documentId: string): Promise<ApprovalInstance> {
  // requestRaw goes through the shared transport (credentials, RFC 9457 Problem
  // decoding, authn.expired dispatch) but returns the Response so the ETag header
  // can seed the cache for a subsequent If-Match write (F-FE-RAWFETCH).
  const res = await requestRaw(`${BASE}/documents/${documentId}/approval-instance`);
  const data = (await res.json()) as ApprovalInstance;
  const etag = res.headers.get('ETag');
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
  return apiFetch<ListInboxResponse>(url);
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

// F-FE-10: the active-document fetch is owned by the controlled-documents feature
// (features/controlled-documents/api/controlledDocuments.ts — fetchActiveDocumentInstance).
// Approval consumers import that hook/fn directly instead of forking a second copy
// under a second cache root; see useActiveDocumentContextQuery (deleted) and
// QK.controlledDocuments.activeDocument.
