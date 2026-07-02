export type DocumentStatus =
  | "DRAFT"
  | "IN_REVIEW"
  | "APPROVED"
  | "PUBLISHED"
  | "ARCHIVED"
  | "draft"
  | "under_review"
  | "approved"
  | "published"
  | "obsolete";
export type Classification = "PUBLIC" | "INTERNAL" | "CONFIDENTIAL" | "RESTRICTED";
// Canonical 8-role backend enum. The legacy FE-only phantoms "admin" and
// "reviewer" were removed in PR-12b — their last consumers were migrated to
// system_admin (DocumentPublishedPage) and to plain workflow-role strings
// (templates). features/iam uses the narrower IamRole alias declared next to
// the IAM admin code.
export type UserRole =
  | "system_admin"
  | "approver"
  | "author"
  | "editor"
  | "viewer"
  | "signer"
  | "area_admin"
  | "qms_admin";

export interface CurrentUser {
  userId: string;
  tenantId: string;
  tenantName: string;
  username: string;
  email?: string;
  displayName: string;
  mustChangePassword: boolean;
  roles: UserRole[];
  // UX hint only — backend remains the sole authz enforcer.
  // wiki/concepts/authz-tiers.md
  capabilities: string[];
}

export interface AreaMembership {
  areaId: string;
  areaCode: string;
  areaName: string;
  roleInArea: UserRole;
  grantedAt: string;
}

export interface DocumentProfileItem {
  code: string;
  familyCode: string;
  name: string;
  alias: string;
  description: string;
  reviewIntervalDays: number;
  activeSchemaVersion: number;
  workflowProfile: string;
  approvalRequired: boolean;
  retentionDays: number;
  validityDays: number;
}

export interface MetadataFieldRuleItem {
  name: string;
  type: string;
  required: boolean;
}

export interface DocumentProfileSchemaItem {
  profileCode: string;
  version: number;
  isActive: boolean;
  metadataRules: MetadataFieldRuleItem[];
  contentSchema?: Record<string, unknown>;
}

export interface DocumentListItem {
  documentId: string;
  title: string;
  documentType: string;
  documentProfile: string;
  documentFamily: string;
  documentSequence?: number;
  documentCode?: string;
  profileSchemaVersion?: number;
  processArea?: string;
  subject?: string;
  ownerId: string;
  businessUnit: string;
  department: string;
  classification: Classification;
  status: DocumentStatus;
  tags: string[];
  effectiveAt?: string;
  expiryAt?: string;
  createdAt: string;
}

export interface SearchDocumentItem extends DocumentListItem {
}

export interface WorkflowApprovalItem {
  approvalId: string;
  documentId: string;
  requestedBy: string;
  assignedReviewer: string;
  decisionBy?: string;
  status: "PENDING" | "APPROVED" | "REJECTED";
  requestReason?: string;
  decisionReason?: string;
  requestedAt: string;
  decidedAt?: string;
}

export interface DocumentTemplateSnapshotItem {
  templateKey: string;
  version: number;
  profileCode: string;
  schemaVersion: number;
  definition: Record<string, unknown>;
}

export interface AuditEventItem {
  id: string;
  occurredAt: string;
  actorId: string;
  action: string;
  resourceType: string;
  resourceId: string;
  payload: Record<string, unknown>;
  traceId: string;
}
