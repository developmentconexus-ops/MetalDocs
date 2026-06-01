# Local Dev Startup

**Last verified:** 2026-05-20

## TL;DR

```powershell
# From repo root — PowerShell only, never bash
.\scripts\start-api.ps1
```

Frontend (separate terminal):
```bash
cd frontend/apps/web && pnpm dev
# → http://localhost:4173
```

---

## Why PowerShell, not bash

`.env` contains `PGPASSWORD=Lepa12<>!`. The `<>` characters are I/O redirect operators in bash. Running `source .env` or `set -o allexport; source .env` silently corrupts the value — Postgres connection fails with auth error that looks unrelated.

PowerShell string assignment is literal — `<>` is safe.

**Never use `scripts/start-api.sh` or `bash source .env`.**

---

## What the script does

1. Loads all vars from `.env` (split on first `=` — safe for `<>`)
2. Forces `APP_PORT=8081` (binary defaults to 8080 if this var is missing)
3. Rebuilds on `-Build`, or auto-rebuilds when timestamp checks show a stale API, jobs, or worker binary
4. Replaces the current workspace API on `:8081` only when ownership is proven; otherwise fails loudly on an occupied port
5. Starts `metaldocs-jobs` after timestamp-based freshness checks unless `-NoJobs` is passed
6. Starts the API after timestamp-based freshness checks
7. Starts the worker too unless `-NoWorker` is passed

Pass `-Build` to force rebuild:
```powershell
.\scripts\start-api.ps1 -Build
```

Skip the dedicated jobs runtime when you intentionally want API-only startup:
```powershell
.\scripts\start-api.ps1 -NoJobs
```

## Fast API-only loop

For backend debugging, route testing, or quick iteration where you do not want
worker/jobs startup or broad binary freshness scans, use:

```powershell
.\scripts\dev-api.ps1
```

What it does:

1. Loads `.env`
2. Forces `APP_PORT=8081`
3. Reuses the current healthy `metaldocs-api.exe` on `:8081` when possible
4. Builds `metaldocs-api.exe` only when it is missing or when `-Build` is passed
5. Starts only the API
6. Waits only for `GET /api/v1/health/ready`

Useful flags:

```powershell
.\scripts\dev-api.ps1 -Build
.\scripts\dev-api.ps1 -ForceRestart
```

Use this for the inner dev loop. Keep `.\scripts\start-api.ps1` as the
canonical heavier startup path when you intentionally need worker/jobs parity.

## Script-truth policy

Local startup follows `script-truth` policy:

- startup scripts are authoritative
- ad hoc startup commands are not authoritative
- stale binaries are not trusted
- if startup evidence conflicts with the script or current source, trust the script and rerun from there

If you see a route, auth, or startup result that does not match current code expectations, assume the first suspect is stale runtime truth until the canonical script rebuilds or proves freshness.

## Runnable preflight before screen work

Runnable preflight is required before screen work. Do not start a screen task just because the frontend renders.

Run the API from the canonical script, then run the runnable preflight:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\scripts\check-system-runnable.ps1 -TargetRoute /api/v1/templates
```

This checkpoint verifies:

1. login succeeds from the current runtime
2. session validation works after login
3. the target route responds from the same runtime boundary

When `-StartApi` is passed, `check-system-runnable.ps1` now uses the fast
API-only launcher (`scripts/dev-api.ps1`) before it performs the auth and route
checks. That keeps the verification gate intact without forcing worker/jobs
startup for every backend smoke.

If any checkpoint fails, classify it as a prerequisite first. Do not absorb startup, auth, or shared contract repair silently into a screen implementation task.

---

## docgen-v2 (token substitution service)

Required for document approval flows that fan out to DOCX/PDF generation. If `METALDOCS_FANOUT_URL` is unset, approval can complete locally with freeze skipped. If fanout is configured but docgen-v2 is unavailable, approval can fail during fanout.

**Setup (first time):**
```powershell
# 1. Copy env template
Copy-Item .env.docgen-v2.example .env.docgen-v2
# 2. Fill in MinIO creds (dev: minioadmin/minioadmin) and set DOCGEN_V2_SERVICE_TOKEN
# 3. Install dependencies
cd apps/docgen-v2 && npm install
```

**Start:**
```powershell
.\scripts\dev-docgen.ps1   # runs on port 3001
```

**Wire to API:** In `.env`, set:
```
METALDOCS_FANOUT_URL=http://localhost:3001
METALDOCS_DOCGEN_V2_SERVICE_TOKEN=<same value as DOCGEN_V2_SERVICE_TOKEN in .env.docgen-v2>
```

**Health check:** `GET http://localhost:3001/health` → `{"status":"ok"}`

---

## Credentials

| field | admin | approver | author-test | approver-test |
|---|---|---|---|---|
| identifier | `admin` | `approver` | `author-test` | `approver-test` |
| password | `AdminMetalDocs123!` | `ApproverMetalDocs123!` | _(curated seed)_ | `ApproverTest123!` |
| roles | `system_admin` | `admin` (full caps) | `author` | `approver` (limited caps) |
| use case | full admin / wizard / list QA | template approve + publish QA (satisfies SoD vs admin author) — preferred for editor lifecycle drive | draft authoring | capability-gate QA: role `approver` holds `template.approve` + `template.review` + `template.view`, lacks `template.submit` / `template.publish` |

Login endpoint: `POST /api/v1/auth/login`, body field `identifier` (NOT `username`).

The curated local bootstrap creates these users when `scripts/dev-bootstrap-baseline.ps1 -WithDevSeed` runs. First-boot bootstrap admin is separate and opt-in.

**SoD note:** templates lifecycle requires the approver's `userId` to differ from the author's `userId`. `admin` authors → `approver` approves (or `approver-test` for capability-gate-only checks). Two distinct seeded admin-role users (`admin` + `approver`) exist solely to satisfy this segregation locally; see [`migrations/0159_seed_dev_approver_user.sql`](../../migrations/0159_seed_dev_approver_user.sql).

`approver-test` password was reset to `ApproverTest123!` on 2026-05-31 (template-editor QA Preview drive) via `PATCH /api/v1/iam/users/approver-test` `{"newPassword":"ApproverTest123!","mustChangePassword":false}` — required `Origin: http://localhost:8081` to bypass CSRF check.

**To reset local credentials:** rerun `scripts/dev-bootstrap-baseline.ps1 -WithDevSeed`. This rebuilds the curated local database and reapplies the local dev seed.

---

## DB access

```powershell
docker exec metaldocs-postgres psql -U metaldocs_app -d metaldocs -c "SELECT 1;"
```

Port: `5433` (host) → `5432` (container). DB: `metaldocs`. Schema split:
- `metaldocs.*` — users, auth, IAM
- `public.*` — documents, templates, approvals

---

## Worker (PDF generation)

Required for PDF generation after document approval. Polls `messaging_outbox` using the configured worker interval, calls docgen-v2 `/convert/pdf`, and writes `final_pdf_s3_key` to DB.

**Start (separate terminal, after API + docgen-v2 are up):**
```powershell
.\scripts\start-worker.ps1        # uses existing metaldocs-worker.exe
.\scripts\start-worker.ps1 -Build # rebuild binary first
```

**Verify running:** startup logs `MetalDocs Worker running (poll_interval_s=...)` on boot. The exact value comes from `METALDOCS_WORKER_POLL_INTERVAL_SECONDS` in `.env`, and ongoing `worker_batch result=completed ...` logs follow that configured interval.

**Env vars required (already in `.env`):**
- `METALDOCS_DOCGEN_V2_URL=http://localhost:3001`
- `METALDOCS_DOCGEN_V2_SERVICE_TOKEN=dev-local-service-token-32chars!!`

**If PDF not generated after signoff:** check worker log for `event_type=docgen_v2_pdf result=published`. If missing, first confirm whether `METALDOCS_FANOUT_URL` is configured at all; when it is unset, freeze is skipped. If it is configured, inspect worker/docgen-v2 connectivity because fanout failures can block approval.

---

## Jobs runtime (scheduled publish cutover)

Required for scheduled publish execution after the River cutover. The API now owns only transactional enqueue; `metaldocs-jobs` owns the temporal queue worker that performs the publish at effective time.

**Default startup:** `.\scripts\start-api.ps1` now starts jobs automatically unless `-NoJobs` is passed.

**Start manually (separate terminal):**
```powershell
.\scripts\start-jobs.ps1
.\scripts\start-jobs.ps1 -Build
```

**Verify running:** startup logs `MetalDocs Jobs running (queues=temporal)`.

**Env vars:**
- `METALDOCS_JOBS_ENABLED=true` by default; set false only when intentionally disabling the jobs host
- `METALDOCS_JOBS_RIVER_SCHEMA` optional River schema override
- `METALDOCS_JOBS_TEMPORAL_MAX_WORKERS` optional queue concurrency override

**Runtime truth:** scheduled publish should no longer be hosted by the API runtime. If a schedule request succeeds but no future publish happens, inspect the `metaldocs-jobs` process first.

---

## Deep QA companion

For deep QA on the modern `documents + approval` flow, use:

- `wiki/references/documents-approval-deep-qa/README.md`
- `wiki/references/documents-approval-deep-qa/runbook.md`
- `wiki/references/documents-approval-deep-qa/fixtures.md`
- `wiki/references/documents-approval-deep-qa/matrix.md`

This startup guide remains the startup truth. The deep QA runbook owns the session workflow, evidence recipes, and fixture guidance.

---

## Common mistakes

| Mistake | Symptom | Fix |
|---|---|---|
| Used bash to source .env | `pq: password authentication failed` | Use PS script |
| Missing `APP_PORT` | API starts on :8080 not :8081 | Script sets it explicitly |
| Occupied :8081 by another process | startup script fails before boot | Free the port or stop the conflicting process intentionally |
| Wrong login body field | `AUTH_INVALID_CREDENTIALS` | Use `identifier`, not `username` |
| Dev seed skipped | Can't login as admin | Rerun `scripts/dev-bootstrap-baseline.ps1 -WithDevSeed` |
