import type { IamRole } from "./types";

export const ROLE_OPTIONS: ReadonlyArray<readonly [IamRole, string]> = [
  ["viewer", "Visualizador"],
  ["author", "Autor"],
  ["editor", "Editor"],
  ["approver", "Aprovador"],
  ["signer", "Assinante"],
  ["area_admin", "Area admin"],
  ["qms_admin", "QMS admin"],
  ["system_admin", "System admin"],
];

// Single source for the role options offered when granting an *area* membership.
// The contract (`GrantAreaMembershipRequest.role`) accepts the full `UserRole`
// enum, so this mirrors all 8 roles in ROLE_OPTIONS — kept as a dedicated
// export so area-membership UIs never re-declare the list and so the area set
// can diverge from the invite tenant-role set later without churn.
export const IAM_AREA_ROLES: ReadonlyArray<readonly [IamRole, string]> = ROLE_OPTIONS;

export const SP_DATE_TIME_FORMATTER = new Intl.DateTimeFormat("pt-BR", {
  timeZone: "America/Sao_Paulo",
  dateStyle: "short",
  timeStyle: "short",
});

export const SP_DATE_FORMATTER = new Intl.DateTimeFormat("pt-BR", {
  timeZone: "America/Sao_Paulo",
  dateStyle: "short",
});
