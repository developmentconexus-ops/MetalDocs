// Canonical 8-role backend enum (matches internal/modules/iam/domain/model.go).
// Identical to `UserRole` in lib/types after PR-12b removed the phantoms.
// Kept as a feature-local alias so IAM code remains decoupled from the shared
// types module and lines up 1:1 with the codegen schema.
export type IamRole =
  | "system_admin"
  | "approver"
  | "author"
  | "editor"
  | "viewer"
  | "signer"
  | "area_admin"
  | "qms_admin";
