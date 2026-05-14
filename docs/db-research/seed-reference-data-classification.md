# Seed and Reference Data Classification

## Commands

```powershell
rg -n "seed|auth|iam|capabil|permission|role|template|bootstrap|fixture" -S
rg -n "role_capabilities|iam_users|iam_user_roles|auth_sessions|auth_identities|bootstrap|seed|template_defaults|profile|process_area|reference" migrations scripts apps/api/cmd cmd -S
```

## Classification

## Product Reference Data

- IAM/auth schema foundations: `migrations/0002_init_iam_rbac.sql`, `migrations/0021_init_auth_identities_and_sessions.sql`.
- Runtime capability seeds/mutations:
  - `migrations/0165_role_capabilities_reseed.sql`
  - `migrations/0169_role_capabilities_process_areas.sql`
  - `migrations/0186_role_capabilities_typed.sql`
  - `migrations/0187_registry_lifecycle_caps_seed.sql`
  - `migrations/0189_audit_capability_seed.sql`
  - `migrations/0192_template_review_capability.sql`
- System blank template reference:
  - `migrations/0199_system_blank_template.sql`.

## Local Dev Seed Data

- Dev approver and role seed lineage:
  - `migrations/0159_seed_dev_approver_user.sql`
  - `migrations/0170_dev_approver_role_correction.sql`
  - `migrations/0158_fix_process_area_role_constraint.sql` (dev convenience rows).
- Local/e2e seed tools:
  - `apps/api/cmd/metaldocs-e2e-seed/main.go`
  - `scripts/e2e-seed.ps1`
  - `internal/test/e2e_seed.go`
  - `cmd/seed-test-document/main.go`.

## Mixed / Needs Split

- `migrations/0029_seed_metal_nobre_document_registry.sql` (tenant/business specific values mixed with structural process/profile records).
- `migrations/0055_seed_po_document_canvas_template.sql`, `migrations/0057_seed_po_browser_template.sql`, `migrations/0066_switch_po_profile_default_to_mddm.sql` (template defaults with potentially environment-opinionated content).

## Runtime Bootstrap Risk Notes

- Bootstrap local admin path currently present at startup:
  - `apps/api/cmd/metaldocs-api/main.go:150-153`
  - `internal/modules/auth/application/service.go:72-113`
  - `internal/platform/authn/config.go:86-121`
  - `deploy/compose/docker-compose.yml:117-122`.
- Dev tenant sentinel fallback appears in runtime/session behavior:
  - `internal/platform/tenant/const.go:3-4`
  - `internal/modules/auth/application/service.go:203-205`
  - `migrations/0184_auth_sessions_tenant_id.sql`
  - `migrations/0185_revoke_ambiguous_sessions.sql`.

## Open Questions

- Production guard policy for bootstrap admin outside local environments.
- Final split policy for mixed template/profile seeds.
