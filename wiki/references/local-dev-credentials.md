# Local Dev Credentials

**Last verified:** 2026-05-15

## API login

Login endpoint: `POST /api/v1/auth/login`
Body field is `identifier` (not `username`).

| identifier | password | IAM role | notes |
|---|---|---|---|
| `admin` | `AdminMetalDocs123!` | system_admin | local dev seed account; use to author documents/templates |
| `approver` | `ApproverMetalDocs123!` | approver | local dev seed account for SoD signoff flows |
| `author-test` | `AuthorTest123!` | author | smoke-test author: creates templates + docs + submits for approval |
| `approver-test` | `ApproverMetalDocs456!@` | approver | smoke-test approver: signs off docs submitted by `author-test` (ISO seg test requires distinct userIds); password reset 2026-05-01; role renamed from `reviewer` by migration 0166 |

These accounts are created by `db/dev-seeds/0001_local_dev_seed.sql` when the curated bootstrap runs with `-WithDevSeed`.

First-boot bootstrap admin is separate and opt-in. Do not rely on it for the normal fresh local login path.

## API startup

Port: `8081`. Binary: `metaldocs-api.exe` (compiled from `./apps/api/cmd/metaldocs-api/...`).
Critical env vars that must be set explicitly via PowerShell (bash corrupts `PGPASSWORD` due to `<>` chars):

```
APP_PORT=8081
PGHOST=127.0.0.1
PGPORT=5433
PGDATABASE=metaldocs
PGUSER=metaldocs_app
PGPASSWORD=***REDACTED***   ← set via $env:PGPASSWORD in PowerShell, never via bash source .env
```

See `.env` for full var list. Use `scripts/start-api-ps.ps1` if it exists, otherwise set manually.

**CRITICAL:** `docgen-v2` (port 3001) must be running before starting the API. The approval/signoff transaction calls docgen-v2 synchronously — if it's down, signoffs fail and roll back. Start it with:
```powershell
cd apps/docgen-v2
$env:METALDOCS_STORAGE_PROVIDER="minio"; $env:MINIO_ENDPOINT="localhost:9000"; ...  # see .env.docgen-v2
npx tsx src/index.ts
```
Or use `scripts/dev-docgen.ps1` if available.

## DB access

```
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -c "<query>"
```

User tables: `metaldocs.auth_identities`, `metaldocs.iam_users`, `metaldocs.iam_user_roles`
Document tables: `public.documents`, `public.controlled_documents` (note: `public.documents_v2` was dropped by migration 0168)
Template tables: `public.templates`, `public.template_versions`, `public.templates_template`, `public.templates_template_version`

## Process-area roles (approval authz)

The approval authz system (`authz.Require`) resolves capabilities via:

```sql
SELECT rc.capability
FROM metaldocs.role_capabilities rc
JOIN metaldocs.user_process_areas upa ON upa.role = rc.role
WHERE upa.user_id = ? AND upa.area_code = ? AND upa.effective_to IS NULL
```

Dev seed: `admin` user has **qms_admin** role in `general` area (applied by migration 0158).

| role | key capabilities |
|---|---|
| `author` | doc.submit, doc.edit_draft |
| `reviewer` | doc.signoff, doc.submit |
| `signer` | doc.signoff |
| `area_admin` | doc.submit, doc.signoff, doc.publish, membership.grant |
| `qms_admin` | all of the above + doc.obsolete, doc.reconstruct, route.admin |

**Historical note:** `user_process_areas_role_check` originally only allowed `viewer/editor/reviewer/approver` (0125) while `role_capabilities` used a different set. Migration 0158 widens the constraint to align them. See `decisions/` for the ADR.
