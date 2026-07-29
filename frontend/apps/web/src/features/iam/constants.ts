import {
  AREA_ROLES,
  USER_ROLES,
  getRoleLabel,
  type AreaRole,
  type UserRole,
} from "../../lib/iam/role-vocabulary";

/**
 * Role options offered when assigning a *tenant* role (invite, user filter,
 * people drawer). Derived from the generated `UserRole` enum — see
 * `src/lib/iam/role-vocabulary.ts`, which is the single runtime projection of
 * the contract vocabulary and is exhaustiveness-guarded against it.
 *
 * F-QA4-2: this used to be a hand-written list of 8 [code, label] pairs, so a
 * contract enum change silently left the UI stale and three labels had drifted
 * away from the canonical PT-BR ones ("Area admin" vs "Administrador de área").
 * Labels now come from the same presenter every other surface uses.
 */
export const ROLE_OPTIONS: ReadonlyArray<readonly [UserRole, string]> =
  USER_ROLES.map((code) => [code, getRoleLabel(code)] as const);

/**
 * Role options offered when granting an *area* membership.
 *
 * F-QA4-2: this previously aliased ROLE_OPTIONS (all 8 roles) on the belief
 * that `GrantAreaMembershipRequest.role` accepted the full `UserRole` enum. It
 * does not: an area membership is a `public.user_process_areas` row, and that
 * table's role CHECK accepts exactly the seven area-assignable roles —
 * `system_admin` is tenant-wide tier-1 and can never be an area membership.
 * Offering it produced a request the database rejected. The contract now
 * declares `AreaRole` for that field and this list derives from it, so the UI
 * cannot offer an ungrantable role again.
 */
export const IAM_AREA_ROLES: ReadonlyArray<readonly [AreaRole, string]> =
  AREA_ROLES.map((code) => [code, getRoleLabel(code)] as const);

export const SP_DATE_TIME_FORMATTER = new Intl.DateTimeFormat("pt-BR", {
  timeZone: "America/Sao_Paulo",
  dateStyle: "short",
  timeStyle: "short",
});

export const SP_DATE_FORMATTER = new Intl.DateTimeFormat("pt-BR", {
  timeZone: "America/Sao_Paulo",
  dateStyle: "short",
});
