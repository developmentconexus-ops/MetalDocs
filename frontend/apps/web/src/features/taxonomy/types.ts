import type { components } from "../../lib/api-types";

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
  editableByRole?: string;
}

export interface UpdateProfileRequest {
  familyCode: string;
  name?: string;
  description?: string;
  editableByRole?: string;
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
  defaultApproverRole?: string;
}

export interface UpdateAreaRequest {
  name?: string;
  description?: string;
  parentCode?: string | null;
  defaultApproverRole?: string | null;
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
