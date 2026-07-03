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

// Inbox read shape is generated from components['schemas']['ApprovalInboxItem']/
// ['ApprovalInboxResponse'] — never hand-rolled. The listApprovalInbox contract
// was repaired alongside the stage-signoff/cancel-instance operations (same
// defect class as 32c5066e: handler wrote contracts.InboxResponse while the
// spec declared a bare 200).
export type InboxItem = components['schemas']['ApprovalInboxItem'];

// FE-03: SubmitRequest/SubmitResponse (formerly hand-rolled here) moved to
// approvalApi.ts as aliases of the generated
// components['schemas']['SubmitDocumentRequest']/['SubmitDocumentResponse'].
// The six approval document-mutation pairs (signoff, publish, schedule-publish,
// supersede, obsolete, cancel) followed once their OpenAPI contracts were
// repaired (32c5066e declared real request/response bodies) — approvalApi.ts
// now aliases the generated SignoffDocument*/PublishDocumentResponse/
// SchedulePublishDocumentRequest/SupersedeDocument*/ObsoleteDocument*/
// CancelDocumentApproval* schemas. Publish is bodyless on the wire.

export interface ListInboxParams {
  area_code?: string;
  limit?: number;
  offset?: number;
}

export type ListInboxResponse = components['schemas']['ApprovalInboxResponse'];
