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

| field | value |
|---|---|
| Login endpoint | `POST /api/v1/auth/login` |
| Body field | `identifier` (NOT `username`) |
| identifier | `admin` |
| password | `AdminMetalDocs123!` |

The curated local bootstrap creates this user when `scripts/dev-bootstrap-baseline.ps1 -WithDevSeed` runs. First-boot bootstrap admin is separate and opt-in.

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
