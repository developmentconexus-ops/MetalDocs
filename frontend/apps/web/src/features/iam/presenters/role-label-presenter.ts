// THE role-label entrypoint for IAM code.
//
// F-QA4-2: the label map itself moved to lib/iam/role-vocabulary, next to the
// role codes it is keyed by. Role vocabulary is cross-cutting — taxonomy and
// approval render role pickers too, and hard-rule #8 forbids them importing
// from features/iam. This module stays as the IAM-facing name so every existing
// `getRoleLabel` import keeps working; there is still exactly one label
// definition, and it is exhaustiveness-guarded against the generated UserRole
// enum.
export { getRoleLabel } from "../../../lib/iam/role-vocabulary";
