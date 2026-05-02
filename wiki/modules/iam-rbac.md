# Module: iam-rbac

> **Last verified:** 2026-05-01
> **Status:** Stub. Fill in full capability matrix when role definitions stabilize.
> **Scope:** Capabilities, roles, area-scoped grants.
> **Out of scope:** Authentication mechanism (login, sessions) — TBD.
> **Key files:**
> - `internal/modules/iam/` — backend module (verify path)
> - `internal/platform/auth/` — auth middleware

## Model

- **Capability**: fine-grained permission (e.g. `templates:author`, `templates:approve`, `documents:create`, `documents:approve`, `taxonomy:manage`, `registry:create`).
- **Role**: bundle of capabilities (e.g. `Admin`, `Template Author`, `Document Author`, `Approver`).
- **Area grant**: scopes a role to one or more areas. A `Document Author` can be granted only for `RH`, not `Qualidade`.

## Capability checks

Most checks live at the API layer (Go middleware). Some UI gates also hide buttons based on capability presence — UI gating is a UX nicety, the API is the source of truth.

## ISO segregation overlay

Independently of capabilities, the approval module enforces that the submitter of a document cannot signoff on it. See [concepts/iso-segregation.md](../concepts/iso-segregation.md).

## Capability matrix (TBD)

| Capability             | Description                                           |
|------------------------|-------------------------------------------------------|
| `taxonomy:manage`      | Create/edit profiles + areas, bind templates          |
| `templates:author`     | Create new templates and version drafts               |
| `templates:approve`    | Sign off on template versions                         |
| `templates:publish`    | Publish approved template versions                    |
| `registry:create`      | Register new controlled documents                     |
| `documents:create`     | Generate working document versions                    |
| `documents:edit`       | Edit document content while in draft                  |
| `documents:submit`     | Move draft → under_review                             |
| `documents:approve`    | Sign off on document versions                         |
| `users:manage`         | Admin user/role management                            |

(Verify against backend before relying on this.)

## See also

- [vision/target-users.md](../vision/target-users.md)
- [concepts/iso-segregation.md](../concepts/iso-segregation.md)
- [modules/approval.md](approval.md)
