import type { components } from '../../../lib/api-types';

export type ApprovalState =
  | 'draft'
  | 'under_review'
  | 'approved'
  | 'scheduled'
  | 'published'
  | 'superseded'
  | 'rejected'
  | 'obsolete'
  | 'cancelled';

export type SignatureMethod = 'password_reauth' | 'icp_brasil';

// Route-admin DTOs (Route, RouteStage, *RouteRequest/*RouteResponse, ListRoutesResponse)
// previously lived here as hand-rolled types. They were deleted in the
// canonical FE rewrite — consumers now import the codegen types from
// `lib/api-types` via `features/approval/api/routeAdminApi.ts`.

// CON-10: the approval-instance read shape (id/document_id/route_id/tenant_id/
// status/stages/actors/etag) is now generated from
// `components['schemas']['ApprovalInstanceByDocumentResponse']` and friends —
// never hand-rolled. The prior hand-written unions drifted from the runtime
// values emitted by `mapInstanceResponse` (internal/modules/documents/approval/
// http/doc_approval_handler.go): instance status is
// `in_progress | approved | rejected | cancelled` (there is no `completed`
// value), and the `actors` array was missing entirely. See
// wiki/backlog/editor.md:254.
export type ApprovalInstance = components['schemas']['ApprovalInstanceByDocumentResponse'];
export type StageInstance = components['schemas']['ApprovalStageInstanceResponse'];
export type StageActor = components['schemas']['ApprovalStageActorResponse'];
export type Signoff = components['schemas']['ApprovalSignoffRecordResponse'];

export interface InboxItem {
  instance_id: string;
  document_id: string;
  controlled_document_id: string;
  document_title: string;
  area_code: string;
  submitted_by: string;
  submitted_at: string;
  stage_label: string;
  quorum_progress: string;
}

// FE-03: SubmitRequest/SubmitResponse (formerly hand-rolled here) moved to
// approvalApi.ts as aliases of the generated
// components['schemas']['SubmitDocumentRequest']/['SubmitDocumentResponse'] —
// that route's contract is not drifted, unlike the six below.

export interface SignoffRequest {
  decision: 'approve' | 'reject';
  reason?: string;
  password: string;
  content_hash: string;
}

export interface SignoffResponse {
  signoff_id: string;
  was_replay: boolean;
}

export interface PublishRequest {
  content_hash: string;
}

export interface PublishResponse {
  document_id: string;
}

export interface SchedulePublishRequest {
  effective_from: string;
  superseded_document_id?: string;
}

export interface SchedulePublishResponse {
  document_id: string;
  scheduled_at: string;
}

export interface SupersedeRequest {
  superseded_document_id: string;
}

export interface SupersedeResponse {
  document_id: string;
}

export interface ObsoleteRequest {
  reason: string;
}

export interface ObsoleteResponse {
  document_id: string;
}

export interface CancelRequest {
  reason: string;
}

export interface CancelResponse {
  document_id: string;
}

export interface ListInboxParams {
  area_code?: string;
  limit?: number;
  offset?: number;
}

export interface ListInboxResponse {
  items: InboxItem[];
  total: number;
}
