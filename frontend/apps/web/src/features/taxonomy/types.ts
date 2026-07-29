import type { components } from "../../lib/api-types";
import type { AreaRole, UserRole } from "../../lib/iam/role-vocabulary";

// App-level (camelCase) taxonomy types. These are the shapes consumed across
// features (documents, templates, approval) and by this feature's own UI.
// They are intentionally NOT a 1:1 re-export of the generated (snake_case)
// schemas — `taxonomy/api/taxonomy.ts` maps the wire shape to these at the
// transport boundary, the same way `features/documents/api` maps its DTOs.
// Field names below are a structural subset of the generated schema so any
// backend contract drift still fails typecheck at the mapping site.

export interface DocumentProfile {
  code: string;
  // Not present on the generated DocumentProfileItem schema (profiles are
  // tenant-scoped implicitly via the authenticated session, not echoed back
  // per-row) — kept optional so existing fixtures/tests that set it still
  // typecheck without asserting a value the wire response never sends.
  tenantId?: string;
  familyCode: string;
  name: string;
  description: string;
  reviewIntervalDays: number;
  defaultTemplateVersionId: string | null;
  ownerUserId: string | null;
  editableByRole: string;
  governanceClass: components["schemas"]["DocumentProfileItem"]["governance_class"];
  // Approval-route readiness. False means the profile has no active approval
  // route (approval_routes, subject_kind=document) and creating a controlled
  // document under it is rejected with 409 `state.approval_route_missing`.
  // Wire-required on DocumentProfileItem — surfaced as the admin INCOMPLETO
  // badge and as the wizard's fail-closed profile gate.
  hasActiveRoute: boolean;
  // Template-route readiness (ADR 0086). False means the profile has no active
  // TEMPLATE approval route (approval_routes, subject_kind=template,
  // subject_key=profile code) and `POST /templates` under it is rejected with
  // 409 `APPROVAL_ROUTE_MISSING`. Orthogonal to `hasActiveRoute`: a profile can
  // be ready for documents and not for templates, and vice-versa.
  hasActiveTemplateRoute: boolean;
  archivedAt: string | null;
  createdAt: string;
}

export interface ProcessArea {
  code: string;
  tenantId?: string;
  name: string;
  description: string;
  parentCode: string | null;
  ownerUserId: string | null;
  defaultApproverRole: string | null;
  archivedAt: string | null;
  createdAt: string;
}

export interface DocumentFamily {
  code: string;
  name: string;
  description: string;
  isActive: boolean;
  createdAt: string;
}

// Generated request schemas, aliased to the names this feature already uses
// at call sites. `CreateXRequest`/`UpdateXRequest` intentionally narrow the
// generated upsert schema to the fields each operation actually accepts.
type ProfileUpsertRequest = components["schemas"]["TaxonomyProfileUpsertRequest"];
type AreaUpsertRequest = components["schemas"]["TaxonomyAreaUpsertRequest"];
type FamilyUpsertRequest = components["schemas"]["TaxonomyFamilyUpsertRequest"];

export interface CreateProfileRequest {
  code: string;
  familyCode: string;
  name: string;
  description?: string;
  reviewIntervalDays: number;
  // Required by the contract (TaxonomyProfileUpsertRequest) and validated
  // fail-closed by the backend — no phantom default (F-QA4-2).
  editableByRole: UserRole;
}

export interface UpdateProfileRequest {
  familyCode: string;
  name?: string;
  description?: string;
  editableByRole: UserRole;
  reviewIntervalDays?: number;
}

export type SetDefaultTemplateRequest = {
  templateVersionId: components["schemas"]["SetTaxonomyProfileDefaultTemplateRequest"]["template_version_id"];
};

export interface CreateAreaRequest {
  code: string;
  name: string;
  description?: string;
  parentCode?: string;
  defaultApproverRole?: AreaRole;
}

export interface UpdateAreaRequest {
  name?: string;
  description?: string;
  parentCode?: string | null;
  defaultApproverRole?: AreaRole | null;
}

export interface CreateFamilyRequest {
  code: string;
  name: string;
  description?: string;
}

export interface UpdateFamilyRequest {
  name: string;
  description?: string;
}

// Re-exported so `api/taxonomy.ts` can build wire payloads without importing
// generated types directly in more than one place.
export type { ProfileUpsertRequest, AreaUpsertRequest, FamilyUpsertRequest };
