// THE canonical role vocabulary for the web app (F-QA4-2).
//
// Both role enums are single-sourced from the CONTRACT: the types below are the
// generated `components['schemas']` unions, and every runtime list in this file
// is exhaustiveness-guarded against them. The generated types are type-only, so
// a runtime array of values has to be written out once — the `satisfies`
// guards make that single literal list impossible to drift:
//
//   * miss a role  -> "Property 'x' is missing in type ..." (tsc fails)
//   * add a phantom-> "Object literal may only specify known properties" (tsc fails)
//
// Regenerate the types with `pnpm gen:api` after editing api/openapi/v1/openapi.yaml;
// tsc then tells you exactly which lists to resync. Do NOT hand-roll role arrays
// anywhere else — import from here.
//
// Labels live here too (see ROLE_LABELS below) because role vocabulary is
// cross-cutting; features/iam/presenters/role-label-presenter.ts re-exports
// getRoleLabel so IAM code keeps its existing import path.
import type { components } from '../api-types';

/** Canonical tenant role. 8 values; no `admin` or `reviewer` phantoms. */
export type UserRole = components['schemas']['UserRole'];

/**
 * Role assignable as an AREA membership (public.user_process_areas): every
 * UserRole EXCEPT `system_admin`, which is tenant-wide tier-1 and never an area
 * membership. Mirrors internal/platform/iamtypes.areaRoles.
 */
export type AreaRole = components['schemas']['AreaRole'];

// Compile-time proof that AreaRole really is a subset of UserRole. If the spec
// ever lets the two enums diverge, this alias resolves to `never` and the
// assignment below stops compiling.
type AreaRoleIsSubsetOfUserRole = AreaRole extends UserRole ? true : never;
const _AREA_ROLE_SUBSET: AreaRoleIsSubsetOfUserRole = true;
void _AREA_ROLE_SUBSET;

// Display order for pickers: broadest authority first, read-only last. The
// `satisfies` clause is the exhaustiveness guard described above.
const USER_ROLE_ORDER = {
  system_admin: 0,
  qms_admin: 1,
  area_admin: 2,
  approver: 3,
  signer: 4,
  author: 5,
  editor: 6,
  viewer: 7,
} as const satisfies Record<UserRole, number>;

const AREA_ROLE_ORDER = {
  qms_admin: 0,
  area_admin: 1,
  approver: 2,
  signer: 3,
  author: 4,
  editor: 5,
  viewer: 6,
} as const satisfies Record<AreaRole, number>;

function orderedKeys<T extends string>(order: Record<T, number>): readonly T[] {
  return (Object.keys(order) as T[]).sort((a, b) => order[a] - order[b]);
}

/** All 8 canonical tenant roles, in display order. */
export const USER_ROLES: readonly UserRole[] = orderedKeys(USER_ROLE_ORDER);

/** The 7 area-assignable roles (UserRole minus system_admin), in display order. */
export const AREA_ROLES: readonly AreaRole[] = orderedKeys(AREA_ROLE_ORDER);

// PT-BR display labels. These live here rather than in features/iam because
// role vocabulary is cross-cutting: taxonomy, approval and iam all render role
// pickers, and hard-rule #8 forbids feature-to-feature imports. The canonical
// presenter (features/iam/presenters/role-label-presenter) re-exports
// getRoleLabel, so IAM code keeps its existing import path and there is still
// exactly ONE label definition.
const ROLE_LABELS = {
  system_admin: 'Administrador do sistema',
  qms_admin: 'Administrador do SGQ',
  area_admin: 'Administrador de área',
  approver: 'Aprovador',
  signer: 'Assinante',
  author: 'Autor',
  editor: 'Editor',
  viewer: 'Visualizador',
} as const satisfies Record<UserRole, string>;

/**
 * PT-BR label for a role code. Accepts `string` so callers can render values
 * arriving from the wire (including retired phantoms); an unknown code falls
 * back to the raw value rather than rendering blank.
 */
export function getRoleLabel(role: string): string {
  return (ROLE_LABELS as Record<string, string>)[role] ?? role;
}

/** Runtime narrowing for values arriving from the wire or the URL. */
export function isUserRole(value: unknown): value is UserRole {
  return typeof value === 'string' && (USER_ROLES as readonly string[]).includes(value);
}

/** Runtime narrowing for area-scoped role values. */
export function isAreaRole(value: unknown): value is AreaRole {
  return typeof value === 'string' && (AREA_ROLES as readonly string[]).includes(value);
}
