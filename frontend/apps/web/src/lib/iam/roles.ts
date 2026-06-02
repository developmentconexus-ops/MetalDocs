// IAM role catalogue.
//
// ADR 0018 "IAM roles source" — Option B fallback: until a canonical
// `GET /api/v1/iam/roles` lands, the catalogue is mirrored here in the same
// wire shape the future endpoint will return, so callers depend on a stable
// hook-shape (`useIamRolesQuery`) and the swap is a one-file change.
//
// Keep this list in sync with `internal/modules/iam/domain/model.go:10-16`.

export interface IamRole {
  /** Stable identifier sent to the backend. */
  code: string;
  /** Human-readable PT-BR label for UI surfaces. */
  label: string;
  /** One-line description, surfaced as helper text or tooltip. */
  description: string;
}

export interface IamRolesResponse {
  roles: IamRole[];
}

export const IAM_ROLES: ReadonlyArray<IamRole> = [
  {
    code: 'system_admin',
    label: 'Administrador do sistema',
    description: 'Acesso total — administração de tenant, rotas e usuários.',
  },
  {
    code: 'approver',
    label: 'Aprovador',
    description: 'Pode assinar etapas de aprovação de documentos controlados.',
  },
  {
    code: 'editor',
    label: 'Editor',
    description: 'Edita conteúdo de documentos antes da submissão.',
  },
  {
    code: 'author',
    label: 'Autor',
    description: 'Cria documentos e versões, submete para revisão.',
  },
  {
    code: 'viewer',
    label: 'Leitor',
    description: 'Apenas leitura de documentos publicados.',
  },
];

export function getIamRolesFallback(): IamRolesResponse {
  return { roles: [...IAM_ROLES] };
}
