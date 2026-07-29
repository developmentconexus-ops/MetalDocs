import { etagCache } from './etagCache';
import { mutate, type MutateOptions } from './mutationClient';
import { apiFetch, requestRaw } from '../../../lib/api';
import type { components } from '../../../lib/api-types';
import type {
  ApprovalDelegation,
  ApprovalInstance,
  CreateApprovalDelegationRequest,
  FastForwardRequest,
  FastForwardResponse,
  ListInboxParams,
  ListInboxResponse,
  ReviewVerdictRequest,
  ReviewVerdictResponse,
} from './approvalTypes';

// Local argument wrapper for reviewVerdict — the URL is instance/stage-scoped
// but the If-Match version is the instance's, passed explicitly by the caller
// (the instance etag cached by getInstance). The body/response stay generated.
interface ReviewVerdictArgs {
  instanceId: string;
  stageId: string;
  etag: string;
  body: ReviewVerdictRequest;
}

// G3 (unit 2.4): the fast-forward request/args wrapper mirrors ReviewVerdictArgs
// exactly — stage_id is the active REVIEW stage, etag is the instance's current
// version, both supplied by the caller (same shape as reviewVerdict).
interface FastForwardArgs {
  instanceId: string;
  stageId: string;
  etag: string;
  body: FastForwardRequest;
}

// Contract-first: every approval mutation's request/response body is the codegen
// type from api/openapi/v1/openapi.yaml. ADR 0085 (Stage B) deleted the three
// manual publication endpoints (publish / schedule-publish / supersede) — release
// is now an approval-driven coordinator outcome, and the publication PLAN travels
// on SubmitDocumentRequest (planned_effective_from / effective_to / review_due_at /
// superseded_document_id). There is no client-callable publication mutation.
type SubmitRequest = components['schemas']['SubmitDocumentRequest'];
type SubmitResponse = components['schemas']['SubmitDocumentResponse'];
type SignoffRequest = components['schemas']['SignoffDocumentRequest'];
type SignoffResponse = components['schemas']['SignoffDocumentResponse'];
type CancelRequest = components['schemas']['CancelDocumentApprovalRequest'];
type CancelResponse = components['schemas']['CancelDocumentApprovalResponse'];

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
  // F8/M2c filters (W4/P2/P3/P8): only forwarded when defined so the query key
  // (which includes the full params object) caches distinct filter sets apart.
  if (params.stage_kind) {
    qs.set('stage_kind', params.stage_kind);
  }
  if (params.due_before) {
    qs.set('due_before', params.due_before);
  }
  if (params.scope) {
    qs.set('scope', params.scope);
  }
  const url = `${BASE}/approval/inbox${qs.toString() ? `?${qs}` : ''}`;
  return apiFetch<ListInboxResponse>(url);
}

// F2 (M2c): the review sidebar (F4) posts a ready/request-changes verdict. The
// endpoint requires If-Match (optimistic concurrency on the instance version)
// and an Idempotency-Key — both supplied by `mutate`: the caller passes the
// cached instance etag as `ifMatch`, the key is auto-generated. Error mapping is
// the shared problem+json path (same as signoff).
export function reviewVerdict({
  instanceId,
  stageId,
  etag,
  body,
}: ReviewVerdictArgs): Promise<ReviewVerdictResponse> {
  return mutate(
    'POST',
    `${BASE}/approval/instances/${instanceId}/stages/${stageId}/review-verdict`,
    body,
    { ifMatch: etag },
  );
}

// G3 (unit 2.4): fast-forward records the verdict leg + the approve-signoff
// leg on the now-active approval stage in ONE tx. Same optimistic-concurrency
// (If-Match) and Idempotency-Key handling as reviewVerdict/signoff.
export function fastForward({
  instanceId,
  stageId,
  etag,
  body,
}: FastForwardArgs): Promise<FastForwardResponse> {
  return mutate(
    'POST',
    `${BASE}/approval/instances/${instanceId}/stages/${stageId}/fast-forward`,
    body,
    { ifMatch: etag },
  );
}

// F2 (M2c): delegation of eligibility. delegator_id is session-derived
// server-side — never sent in the body. Create carries an auto Idempotency-Key.
export function createDelegation(
  body: CreateApprovalDelegationRequest,
): Promise<ApprovalDelegation> {
  return mutate('POST', `${BASE}/approval/delegations`, body, {});
}

// Revoke returns 204 No Content. `mutate` short-circuits a 204 to `undefined`
// (mutationClient.ts:55 — `if (res.status === 204) return undefined`) BEFORE any
// res.json(), so an empty body does not throw; no requestRaw detour is needed.
export function revokeDelegation(id: string): Promise<void> {
  return mutate('DELETE', `${BASE}/approval/delegations/${id}`, undefined, {});
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
