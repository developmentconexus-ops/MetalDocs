// All routes under /api/v1/documents*. All requests rely on IAM cookies +
// tenant/role headers stamped by the middleware chain; we do not set X-* from
// the client.

import { apiFetch } from '../../../lib/api';

export type DocumentRow = {
  id: string;
  name: string;
  status: 'draft' | 'finalized' | 'archived';
  template_version_id: string;
  updated_at: string;
  current_revision_id?: string;
};

export type CreateDocumentResult = { document_id: string; initial_revision_id: string; session_id: string };
export type AcquireWriter = { mode: 'writer'; session_id: string; expires_at: string; last_ack_revision_id: string };
export type AcquireReadonly = { mode: 'readonly'; held_by: string; held_until: string };
export type AcquireResult = AcquireWriter | AcquireReadonly;
export type PresignResult = { upload_url: string; pending_upload_id: string; expires_at: string };
export type CommitResult = { revision_id: string; revision_num: number; idempotent_replay?: boolean };
export type Checkpoint = { ID: string; DocumentID: string; RevisionID: string; VersionNum: number; Label: string; CreatedAt: string; CreatedBy: string };
export type FinalizeDocumentResult = { instanceId: string };
export type DocumentResponse = {
  ID?: string;
  id?: string;
  Code?: string;
  code?: string;
  Status?: string;
  status?: string;
  Name?: string;
  name?: string;
  CreatedBy?: string;
  created_by?: string;
  CurrentRevisionID?: string;
  current_revision_id?: string;
  RevisionVersion?: number;
  revision_version?: number;
  FormDataJSON?: unknown;
  form_data?: unknown;
};

export async function listDocuments(): Promise<DocumentRow[]> {
  return apiFetch('/api/v1/documents');
}
export async function getDocument(id: string): Promise<DocumentResponse> {
  return apiFetch(`/api/v1/documents/${id}`);
}
export async function renameDocument(id: string, name: string): Promise<void> {
  await apiFetch<void>(`/api/v1/documents/${id}`, {
    method: 'PATCH',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ name }),
  });
}
export async function finalizeDocument(id: string): Promise<FinalizeDocumentResult> {
  return apiFetch<FinalizeDocumentResult>(`/api/v1/documents/${id}/finalize`, {
    method: 'POST',
    idempotencyKey: crypto.randomUUID(),
  });
}
export async function archiveDocument(id: string) {
  return apiFetch(`/api/v1/documents/${id}/archive`, { method: 'POST' });
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
export async function commitAutosave(id: string, req: { session_id: string; pending_upload_id: string; form_data_snapshot?: unknown }): Promise<CommitResult> {
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

export type RestoreCheckpointResult = {
  new_revision_id: string;
  new_revision_num: number;
  source_checkpoint_version_num: number;
  idempotent: boolean;
};
export async function restoreCheckpoint(id: string, versionNum: number): Promise<RestoreCheckpointResult> {
  return apiFetch(`/api/v1/documents/${id}/checkpoints/${versionNum}/restore`, { method: 'POST' });
}

export function signedRevisionURL(documentID: string, revisionID: string): string {
  return `/api/v1/documents/${documentID}/revisions/${revisionID}/url`;
}

export type CommentRow = {
  id: string;
  library_comment_id: number;
  parent_library_id: number | null;
  author: string;
  author_id: string;
  content: unknown[];
  done: boolean;
  created_at: string;
  updated_at: string;
  resolved_at: string | null;
};

export async function listComments(documentID: string): Promise<CommentRow[]> {
  return apiFetch(`/api/v1/documents/${documentID}/comments`);
}

export async function createComment(
  documentID: string,
  body: { library_comment_id: number; parent_library_id?: number; author_display: string; content: unknown[] },
): Promise<CommentRow> {
  return apiFetch(`/api/v1/documents/${documentID}/comments`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function patchComment(
  documentID: string,
  libraryID: number,
  body: { content?: unknown[]; done?: boolean },
): Promise<CommentRow> {
  return apiFetch(`/api/v1/documents/${documentID}/comments/${libraryID}`, {
    method: 'PATCH',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(body),
  });
}

export async function deleteComment(documentID: string, libraryID: number): Promise<void> {
  await apiFetch<void>(`/api/v1/documents/${documentID}/comments/${libraryID}`, { method: 'DELETE' });
}

// Approval instance types
export type SignoffRecord = {
  actor_user_id: string;  // actually display name snapshot from backend
  status: 'pending' | 'approved' | 'rejected' | 'abstained';
  signed_at: string | null;
  comment: string | null;
};

export type StageInstance = {
  stage_id: string;
  name: string;
  status: 'pending' | 'approved' | 'rejected' | 'skipped';
  signoffs: SignoffRecord[];
};

export type ApprovalInstanceResponse = {
  id: string;
  document_id: string;
  status: 'pending' | 'approved' | 'rejected' | 'cancelled';
  stages: StageInstance[];
  created_at: string;
  updated_at: string;
};

export async function getApprovalInstance(documentId: string): Promise<ApprovalInstanceResponse> {
  return apiFetch(`/api/v1/documents/${documentId}/approval-instance`);
}
