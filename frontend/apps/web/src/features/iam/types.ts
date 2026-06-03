// Canonical 8-role backend enum (matches internal/modules/iam/domain/model.go).
// Distinct from `UserRole` in lib/types which retains FE compat aliases "admin" and
// "reviewer" still used by templates/documents/taxonomy features. New IAM admin code
// (PR-12) routes through this narrower type to stay aligned with the codegen schema.
export type IamRole =
  | "system_admin"
  | "approver"
  | "author"
  | "editor"
  | "viewer"
  | "signer"
  | "area_admin"
  | "qms_admin";
