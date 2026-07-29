// IAM role catalogue.
//
// ADR 0018 "IAM roles source" — Option B fallback: until a canonical
// `GET /api/v1/iam/roles` lands, the catalogue is served locally in the same
// wire shape the future endpoint will return, so callers depend on a stable
// hook-shape (`useIamRolesQuery`) and the swap is a one-file change.
//
// F-QA4-2: the catalogue is no longer a hand-maintained list of roles (it used
// to carry only 5 of the 8, silently hiding signer/area_admin/qms_admin from
// every consumer). It is DERIVED:
//   * codes  -> USER_ROLES in lib/iam/role-vocabulary (generated contract enum)
//   * labels -> getRoleLabel in lib/iam/role-vocabulary (re-exported by
//              features/iam/presenters/role-label-presenter for IAM callers)
// Only the descriptions are authored here, and they are exhaustiveness-guarded
// against the generated enum, so a new role in the spec fails the build until
// it is described.

import { USER_ROLES, getRoleLabel, type UserRole } from './role-vocabulary';

export interface IamRole {
  /** Stable identifier sent to the backend. */
  code: UserRole;
  /** Human-readable PT-BR label for UI surfaces. */
  label: string;
  /** One-line description, surfaced as helper text or tooltip. */
  description: string;
}

export interface IamRolesResponse {
  roles: IamRole[];
}

// `satisfies` is the exhaustiveness guard: a role added to (or removed from)
// the OpenAPI UserRole enum breaks compilation here until this map is resynced.
const ROLE_DESCRIPTIONS = {
  system_admin: 'Acesso total — administração de tenant, rotas e usuários.',
  qms_admin: 'Administra o sistema da qualidade: taxonomia, rotas de aprovação e políticas.',
  area_admin: 'Administra usuários e permissões dentro das áreas sob sua responsabilidade.',
  approver: 'Pode assinar etapas de aprovação de documentos controlados.',
  signer: 'Aplica assinatura eletrônica nas etapas que a exigem.',
  author: 'Cria documentos e versões, submete para revisão.',
  editor: 'Edita conteúdo de documentos antes da submissão.',
  viewer: 'Apenas leitura de documentos publicados.',
} as const satisfies Record<UserRole, string>;

export const IAM_ROLES: ReadonlyArray<IamRole> = USER_ROLES.map((code) => ({
  code,
  label: getRoleLabel(code),
  description: ROLE_DESCRIPTIONS[code],
}));

export function getIamRolesFallback(): IamRolesResponse {
  return { roles: [...IAM_ROLES] };
}
