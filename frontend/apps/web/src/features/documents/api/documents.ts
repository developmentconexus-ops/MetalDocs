// All routes under /api/v1/documents*. All requests rely on IAM cookies +
// tenant/role headers stamped by the middleware chain; we do not set X-* from
// the client.

import { apiFetch } from '../../../lib/api';
import type { components, paths } from '../../../lib/api-types';

export type DocumentRow = components['schemas']['DocumentSummary'];
export type DocumentListResponse = components['schemas']['DocumentListResponse'];
export type AcquireWriter = components['schemas']['DocumentSessionWriterResponse'];
export type AcquireReadonly = components['schemas']['DocumentSessionReadonlyResponse'];
export type AcquireResult = paths['/api/v1/documents/{id}/session/acquire']['post']['responses'][200]['content']['application/json'] | paths['/api/v1/documents/{id}/session/acquire']['post']['responses'][201]['content']['application/json'];
export type PresignResult = components['schemas']['DocumentAutosavePresignResponse'];
export type CommitAutosaveRequest = paths['/api/v1/documents/{id}/autosave/commit']['post']['requestBody']['content']['application/json'];
export type CommitResult = paths['/api/v1/documents/{id}/autosave/commit']['post']['responses'][200]['content']['application/json'];
export type Checkpoint = components['schemas']['DocumentCheckpoint'];
export type FinalizeDocumentResult = paths['/api/v1/documents/{id}/finalize']['post']['responses'][201]['content']['application/json'];
export type FinalizeDocumentRequest = components['schemas']['FinalizeDocumentRequest'];
export type DocumentDetail = components['schemas']['DocumentDetailResponse'];
export type DocumentRevisionHistoryResponse = components['schemas']['DocumentRevisionHistoryResponse'];
export type DocumentComment = components['schemas']['DocumentCommentResponse'];
export type DocumentCommentCreateRequest = components['schemas']['DocumentCommentCreateRequest'];
export type DocumentCommentUpdateRequest = components['schemas']['DocumentCommentUpdateRequest'];
export type ApprovalInstanceResponse = components['schemas']['ApprovalInstanceByDocumentResponse'];

export async function listDocuments(): Promise<DocumentRow[]> {
  const response = await apiFetch<DocumentListResponse>('/api/v1/documents');
  return response.items;
}
export async function getDocument(id: string): Promise<DocumentDetail> {
  return apiFetch<DocumentDetail>(`/api/v1/documents/${id}`);
}
export async function renameDocument(id: string, name: string): Promise<void> {
  await apiFetch<void>(`/api/v1/documents/${id}`, {
    method: 'PATCH',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ name }),
  });
}
export async function finalizeDocument(id: string, body: FinalizeDocumentRequest): Promise<FinalizeDocumentResult> {
  return apiFetch<FinalizeDocumentResult>(`/api/v1/documents/${id}/finalize`, {
    method: 'POST',
    idempotencyKey: crypto.randomUUID(),
    body: JSON.stringify(body),
  });
}
export async function archiveDocument(id: string) {
  return apiFetch(`/api/v1/documents/${id}/archive`, { method: 'POST' });
}

export async function getDocumentRevisionHistory(id: string): Promise<DocumentRevisionHistoryResponse> {
  return apiFetch<DocumentRevisionHistoryResponse>(`/api/v1/documents/${id}/revision-history`);
}

export async function acquireSession(id: string): Promise<AcquireResult> {
  return apiFetch(`/api/v1/documents/${id}/session/acquire`, { method: 'POST' });
}
export async function heartbeatSession(id: string, sessionID: string): Promise<void> {
  await apiFetch<void>(`/api/v1/documents/${id}/session/heartbeat`, {
    method: 'POST', headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ session_id: sessionID }),
  });
}
export async function releaseSession(id: string, sessionID: string) {
  return apiFetch(`/api/v1/documents/${id}/session/release`, {
    method: 'POST', headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ session_id: sessionID }),
  });
}
export async function forceReleaseSession(id: string, sessionID: string) {
  return apiFetch(`/api/v1/documents/${id}/session/force-release`, {
    method: 'POST', headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ session_id: sessionID }),
  });
}

export async function presignAutosave(id: string, req: { session_id: string; base_revision_id: string; content_hash: string }): Promise<PresignResult> {
  return apiFetch(`/api/v1/documents/${id}/autosave/presign`, {
    method: 'POST', headers: { 'content-type': 'application/json' },
    body: JSON.stringify(req),
  });
}
// Server is authoritative for content_hash -- it re-computes SHA256 from S3 on
// commit. Client does NOT forward a client-computed hash.
export async function commitAutosave(id: string, req: CommitAutosaveRequest): Promise<CommitResult> {
  return apiFetch(`/api/v1/documents/${id}/autosave/commit`, {
    method: 'POST', headers: { 'content-type': 'application/json' },
    body: JSON.stringify(req),
  });
}

export async function listCheckpoints(id: string): Promise<Checkpoint[]> {
  return apiFetch(`/api/v1/documents/${id}/checkpoints`);
}
export async function createCheckpoint(id: string, label: string): Promise<Checkpoint> {
  return apiFetch(`/api/v1/documents/${id}/checkpoints`, {
    method: 'POST', headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ label }),
  });
}

export type RestoreCheckpointResult = paths['/api/v1/documents/{id}/checkpoints/{version}/restore']['post']['responses'][200]['content']['application/json'];
export async function restoreCheckpoint(id: string, versionNum: number): Promise<RestoreCheckpointResult> {
  return apiFetch(`/api/v1/documents/${id}/checkpoints/${versionNum}/restore`, { method: 'POST' });
}

export function signedRevisionURL(documentID: string, revisionID: string): string {
  return `/api/v1/documents/${documentID}/revisions/${revisionID}/url`;
}

export async function listComments(documentID: string): Promise<DocumentComment[]> {
  return apiFetch<DocumentComment[]>(`/api/v1/documents/${documentID}/comments`);
}

export async function createComment(
  documentID: string,
  body: DocumentCommentCreateRequest,
): Promise<DocumentComment> {
  return apiFetch<DocumentComment>(`/api/v1/documents/${documentID}/comments`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function patchComment(
  documentID: string,
  libraryID: number,
  body: DocumentCommentUpdateRequest,
): Promise<DocumentComment> {
  return apiFetch<DocumentComment>(`/api/v1/documents/${documentID}/comments/${libraryID}`, {
    method: 'PATCH',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function deleteComment(documentID: string, libraryID: number): Promise<void> {
  await apiFetch<void>(`/api/v1/documents/${documentID}/comments/${libraryID}`, { method: 'DELETE' });
}

export async function getApprovalInstance(documentId: string): Promise<ApprovalInstanceResponse> {
  return apiFetch<ApprovalInstanceResponse>(`/api/v1/documents/${documentId}/approval-instance`);
}
