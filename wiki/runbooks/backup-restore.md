# Backup & Restore Runbook (PostgreSQL)

> **Contract:** M8 F8.3 §3.4 (`docs/superpowers/milestones/global-maximum-remediation/milestone-8-ops-readiness/validation-contract.md`).
> **Last verified:** 2026-07-05
> **Operator:** run this cold, top-to-bottom. This runbook wraps EXISTING scripts —
> it does not reimplement backup/restore logic. Do not improvise flags; use the
> exact invocations below.
>
> Scripts wrapped: `scripts/backup-postgres.ps1`, `scripts/restore-postgres.ps1`,
> `scripts/validate-backup.ps1`, `scripts/run-backup-restore-gate.ps1`.

---

## 1. Purpose / when to use

Use this runbook for:

- **Disaster recovery** — the live database is lost, corrupted, or must be
  rolled back to a known-good point.
- **Pre-upgrade snapshot** — before applying a risky migration, schema change,
  or major version bump, take a backup you can restore from if the upgrade
  goes wrong.
- **Periodic backup** — routine scheduled backups per the operator's retention
  policy (cadence is an operator decision — see §7).

This runbook covers the **local / pg_dump class** of backup (single-host,
custom-format dump via `pg_dump`/`pg_restore`). It does not cover offsite/S3
automation (see §8).

---

## 2. Prerequisites

### 2.1 Stack must be reachable

Either the Docker Compose stack is up (`deploy/compose/docker-compose.yml`,
`postgres` service) or you have a reachable Postgres host. Local dev default
(from `.env.example`): host `127.0.0.1`, port `5433`, database `metaldocs`.

```powershell
# If using the local compose stack:
docker compose -f deploy/compose/docker-compose.yml up -d postgres
docker compose -f deploy/compose/docker-compose.yml ps postgres
# Expected: STATUS shows "healthy"
```

### 2.2 A DEDICATED backup DB role (NOT `metaldocs_app`)

`scripts/backup-postgres.ps1` **hard-refuses** `-PgUser metaldocs_app` (see
script, throws `"Backup deve usar usuario dedicado (nao metaldocs_app)."`).
This is intentional, not a bug to route around:

- `metaldocs_app` is the **app-runtime** credential — least-privilege
  principle says the runtime identity should never also be able to run
  filesystem-writing backup operations, and a backup credential should never
  carry the app's read/write/DDL surface used at request time.
- A dedicated backup role should be **read-only** (`SELECT` on all tables),
  narrowing blast radius if the credential leaks.

**The role MUST also have the `BYPASSRLS` attribute.** The M7 RLS-truth
sweep put several tables under `FORCE ROW LEVEL SECURITY` (e.g.
`metaldocs.audit_events`, `metaldocs.approval_signoffs`). A plain
`SELECT`-only role without `BYPASSRLS` fails a full-DB `pg_dump` against
those tables with:

```
ERROR: query would be affected by row-level security policy for table "audit_events"
```

`BYPASSRLS` grants read-past-RLS for `COPY`/`pg_dump` purposes while the role
stays **non-superuser and SELECT-only** — it is NOT equivalent to granting
superuser and does not add write/DDL privileges. This is the standard
Postgres attribute for backup roles over RLS-protected tables.

The role also needs `SELECT` on **both** the `metaldocs` and `public`
schemas, not just `metaldocs` — `backup-postgres.ps1` runs a full-database
`pg_dump` (no schema scope), and `public` holds real data too (River jobs +
migrations, ~30 tables). A role scoped to `metaldocs` alone would silently
miss `public` in the dump.

Create the role once per environment with the provided SQL, connecting as an
admin/superuser (never `metaldocs_app`):

```powershell
# Adjust password/database in scripts/sql/create-backup-role.sql first,
# then execute against the target Postgres as a superuser:
psql --host 127.0.0.1 --port 5433 --username postgres --dbname metaldocs `
  --file scripts/sql/create-backup-role.sql
```

The current `scripts/sql/create-backup-role.sql` provisions all of this: it
creates (or, if the role predates this update, idempotently alters)
`metaldocs_backup` with `CONNECT` + `BYPASSRLS`, and `USAGE` + `SELECT` on
all tables/sequences in **both** `metaldocs` and `public` (plus default
privileges for future tables in both schemas). Do not grant it write
privileges — `BYPASSRLS` bypasses row-level security checks only, it does
not grant `INSERT`/`UPDATE`/`DELETE`/DDL.

For the **restore** side, use a role with `CREATEDB`/owner-level privileges
on the target database (e.g. the admin/superuser `postgres` role, or
`metaldocs_app` if it owns the target scratch database) — `restore-postgres.ps1`
has no user restriction, but restoring is a destructive operation on whatever
database you point it at (see §4 warning).

### 2.3 Required environment variables

`backup-postgres.ps1` requires these in the environment (not as flags):

```powershell
$env:PGHOST = "127.0.0.1"
$env:PGPORT = "5433"
$env:PGDATABASE = "metaldocs"
```

`PGUSER`/`PGPASSWORD` may be supplied either as environment variables or via
the `-PgUser`/`-PgPassword` script parameters (parameters take precedence if
both are set).

`restore-postgres.ps1` requires `PGHOST`/`PGPORT` in the environment;
`PGDATABASE` is only used as a fallback default if `-TargetDatabase` is not
passed (see §4 — always pass `-TargetDatabase` explicitly).

### 2.4 Required binaries

`pg_dump`, `pg_restore`, and `psql` must be on `PATH` (or pass full paths via
`-PgDumpPath` / `-PgRestorePath` / `-PsqlPath`). Version must be compatible
with the target Postgres major version (Postgres 16 per
`deploy/compose/docker-compose.yml`).

---

## 3. Backup procedure

Run from the repo root, with the environment variables from §2.3 set and the
dedicated `metaldocs_backup` role's password at hand:

```powershell
$env:PGHOST = "127.0.0.1"
$env:PGPORT = "5433"
$env:PGDATABASE = "metaldocs"

.\scripts\backup-postgres.ps1 `
  -BackupDir "backups" `
  -EnvironmentName "local" `
  -PgUser "metaldocs_backup" `
  -PgPassword "<metaldocs_backup password>"
```

Parameters (all optional except the effective user/password, which must come
from either the flags above or `PGUSER`/`PGPASSWORD` env vars):

| Param | Default | Notes |
|---|---|---|
| `-BackupDir` | `backups` | Created if missing |
| `-PgDumpPath` | `pg_dump` | Override if not on PATH |
| `-EnvironmentName` | `local` | Tagged into filename + result object |
| `-PgUser` | (falls back to `$env:PGUSER`) | Must NOT be `metaldocs_app` |
| `-PgPassword` | (falls back to `$env:PGPASSWORD`) | — |

**Expected output:**

- A `.dump` file (pg_dump custom format, `--no-owner --no-privileges`) at
  `backups/metaldocs_<environment>_<database>_<UTC-timestamp>.dump`, e.g.
  `backups/metaldocs_local_metaldocs_20260705T120000Z.dump`.
- A SHA-256 checksum computed over that file.
- A PowerShell object (`[PSCustomObject]`) printed to the pipeline with shape:

```powershell
@{
  status           = "success"
  operation        = "backup"
  started_utc      = "2026-07-05T12:00:00.0000000Z"
  finished_utc     = "2026-07-05T12:00:03.1230000Z"
  duration_seconds = 3.123
  operator         = "<Windows username>"
  environment      = "local"
  database         = "metaldocs"
  pg_host          = "127.0.0.1"
  pg_port          = "5433"
  pg_user          = "metaldocs_backup"
  backup_file      = "backups\metaldocs_local_metaldocs_20260705T120000Z.dump"
  checksum_sha256  = "<64-hex-char sha256>"
}
```

Backups land under `backups/` (relative to wherever you invoke the script
from — invoke from repo root for a stable location). This directory is
gitignored; do not commit dump files.

---

## 4. Restore procedure

> **CRITICAL WARNING — never restore over the live/dev database.** Always pass
> `-TargetDatabase` pointing at a **scratch** database created for this
> purpose (e.g. `metaldocs_restore_test`). `restore-postgres.ps1` runs
> `pg_restore --clean --if-exists --exit-on-error`, which **drops and
> recreates objects in the target database** — if the target is your live
> dev/prod database, this destroys current data. A live-restore drill against
> the actual live database is only ever done during a real disaster-recovery
> event, deliberately, with the operator's explicit sign-off — never as a
> routine drill.

Create the scratch target database first (as an admin/owner role):

```powershell
$env:PGPASSWORD = "<admin password>"
psql --host 127.0.0.1 --port 5433 --username postgres --dbname postgres `
  --command "CREATE DATABASE metaldocs_restore_test;"
```

Then restore into it:

```powershell
$env:PGHOST = "127.0.0.1"
$env:PGPORT = "5433"

.\scripts\restore-postgres.ps1 `
  -BackupFile "backups\metaldocs_local_metaldocs_20260705T120000Z.dump" `
  -TargetDatabase "metaldocs_restore_test" `
  -EnvironmentName "local" `
  -PgUser "postgres" `
  -PgPassword "<admin password>"
```

Parameters:

| Param | Default | Notes |
|---|---|---|
| `-BackupFile` | (mandatory) | Path to the `.dump` from §3 |
| `-TargetDatabase` | falls back to `$env:PGDATABASE` | **Always pass this explicitly** — never rely on the fallback pointing at a scratch DB by accident |
| `-PgRestorePath` | `pg_restore` | Override if not on PATH |
| `-EnvironmentName` | `local` | Tagged into result object |
| `-PgUser` / `-PgPassword` | env fallback | Needs owner-level privileges on the target DB |

**Expected output:** a `[PSCustomObject]` with `status = "success"`,
`operation = "restore"`, `target_database`, `backup_file`, timing fields, and
`pg_user`/`pg_host`/`pg_port` — mirroring the backup result shape (see §3).

---

## 5. Validation

### 5.1 Standalone dump validation

`scripts/validate-backup.ps1` checks the dump's logical integrity without
touching any database (`pg_restore --list` against the file):

```powershell
.\scripts\validate-backup.ps1 `
  -BackupFile "backups\metaldocs_local_metaldocs_20260705T120000Z.dump" `
  -EnvironmentName "local"
```

Expected: `Dump valido para restore.` printed, and a result object with
`status = "success"`, `operation = "validate"`, `validation_passed = $true`.

### 5.2 Full gate (backup → validate → restore → smoke)

`scripts/run-backup-restore-gate.ps1` is the end-to-end orchestration — see
§6.

---

## 6. End-to-end recovery drill

`scripts/run-backup-restore-gate.ps1` chains: backup → dump validation →
scratch-database restore → row-count smoke checks, and writes an evidence
JSON file. This is the recommended way to prove backup/restore actually
works, rather than running the three scripts by hand.

```powershell
$env:PGHOST = "127.0.0.1"
$env:PGPORT = "5433"
$env:PGDATABASE = "metaldocs"

.\scripts\run-backup-restore-gate.ps1 `
  -EnvironmentName "local" `
  -BackupDir "backups" `
  -EvidenceDir "backups/evidence" `
  -RestoreValidationDatabase "metaldocs_restore_test" `
  -BackupUser "metaldocs_backup" `
  -BackupPassword "<metaldocs_backup password>" `
  -RestoreUser "postgres" `
  -RestorePassword "<admin password>"
```

Parameters:

| Param | Default | Notes |
|---|---|---|
| `-EnvironmentName` | `local` | Tagged into filenames + evidence |
| `-BackupDir` | `backups` | Where the dump lands |
| `-EvidenceDir` | `backups/evidence` | Where the evidence JSON lands |
| `-RestoreValidationDatabase` | `metaldocs_restore_test` | Scratch DB; created automatically if missing |
| `-PgDumpPath` / `-PgRestorePath` / `-PsqlPath` | `pg_dump` / `pg_restore` / `psql` | Override if not on PATH |
| `-BackupUser` (mandatory) | — | Must NOT be `metaldocs_app` (hard-checked) |
| `-BackupPassword` (mandatory) | — | — |
| `-RestoreUser` (mandatory) | — | Needs `CREATEDB`-capable/owner role to auto-create the scratch DB and restore into it |
| `-RestorePassword` (mandatory) | — | — |

**What it does internally** (for orientation, not to be reimplemented):

1. Calls `backup-postgres.ps1` with `-BackupUser`/`-BackupPassword`.
2. Calls `validate-backup.ps1` against the resulting dump file.
3. Checks whether `-RestoreValidationDatabase` exists (`pg_database` lookup)
   and creates it if not.
4. Calls `restore-postgres.ps1` targeting that scratch database.
5. Runs row-count smoke queries against the restored scratch database:
   `metaldocs.documents`, `metaldocs.document_versions`,
   `metaldocs.iam_users`, `metaldocs.iam_user_roles`,
   `metaldocs.audit_events`, `metaldocs.outbox_events`.
6. Writes an evidence JSON file to
   `backups/evidence/backup_restore_gate_<environment>_<database>_<UTC-timestamp>.json`.

**Expected evidence JSON shape** (`status = "approved"` on success):

```json
{
  "status": "approved",
  "operation": "backup_restore_gate",
  "started_utc": "2026-07-05T12:00:00.0000000Z",
  "finished_utc": "2026-07-05T12:00:10.0000000Z",
  "duration_seconds": 10.0,
  "operator": "<Windows username>",
  "environment": "local",
  "source_database": "metaldocs",
  "restore_validation_database": "metaldocs_restore_test",
  "backup": { "status": "success", "operation": "backup", "...": "..." },
  "validation": { "status": "success", "operation": "validate", "...": "..." },
  "restore": { "status": "success", "operation": "restore", "...": "..." },
  "smoke": {
    "status": "success",
    "checks": {
      "documents_count": 0,
      "document_versions_count": 0,
      "iam_users_count": 0,
      "iam_user_roles_count": 0,
      "audit_events_count": 0,
      "outbox_events_count": 0
    }
  },
  "error": null
}
```

If any step throws, `status` becomes `"rejected"`, `error` carries the
exception message, and the evidence file is still written (`finally` block)
before the script re-throws. Counts of `0` are expected/valid on a
freshly-seeded or empty dev database — the smoke check proves the tables
exist and are queryable post-restore, not that they are non-empty.

Treat `status: "approved"` in the evidence JSON as the PASS signal for this
drill.

---

## 7. Recovery decision guidance

**Restore vs. repair-in-place:**

- **Repair-in-place** when the damage is scoped and understood (e.g. a single
  bad migration, a known-bad write from a specific incident) and a targeted
  SQL fix or migration rollback is lower-risk than a full restore (which
  loses all writes since the backup).
- **Restore from backup** when: the database is unreachable/corrupted at the
  storage layer, the scope of damage is unknown or wide, or a repair attempt
  has already failed/made things worse. Restore is the last-resort, highest-
  blast-radius option — it discards all changes since the backup point.

**RPO/RTO:** backup cadence (how often backups are taken) and therefore the
Recovery Point Objective is an **operator policy decision**, not something
this runbook prescribes — it depends on write volume, tenant SLAs, and
storage budget. Name the chosen cadence (e.g. "nightly", "every 6 hours") in
the operator's runbook/schedule when adopted; do not invent a number here.
Recovery Time Objective depends on database size and restore-host I/O and
should be measured empirically (time the §6 drill against a realistically-
sized dataset) rather than assumed.

---

## 8. Scale-out note (out of scope for v1)

Offsite/S3 (or equivalent object-storage) backup automation, retention-policy
enforcement, and multi-region replication are **out of scope for v1**. This
runbook covers the local/pg_dump class only: a single dump file on local
disk, restored by hand or via the gate script. The named trigger to revisit
this is **the first production hosting decision** (i.e. once MetalDocs has a
concrete production host/cloud target, backup automation and offsite
retention become part of that hosting design, not a standalone v1 addition).

---

## See also

- `scripts/backup-postgres.ps1`, `scripts/restore-postgres.ps1`,
  `scripts/validate-backup.ps1`, `scripts/run-backup-restore-gate.ps1` — the
  scripts this runbook wraps.
- `scripts/sql/create-backup-role.sql` — creates the dedicated
  `metaldocs_backup` read-only role.
- `docs/superpowers/milestones/global-maximum-remediation/milestone-8-ops-readiness/validation-contract.md`
  §3.4 — the binding contract this runbook satisfies.
