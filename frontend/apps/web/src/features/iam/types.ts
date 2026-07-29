// Canonical 8-role backend enum. Feature-local alias of the generated contract
// type so IAM code keeps its own name for it, but there is exactly ONE
// definition (F-QA4-2): the UserRole schema in api/openapi/v1/openapi.yaml.
// Runtime values live in lib/iam/role-vocabulary (USER_ROLES).
export type { UserRole as IamRole } from "../../lib/iam/role-vocabulary";
