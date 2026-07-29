# Incident Runbook — MetalDocs (v1, Docker Compose)

**Deploy target:** Docker Compose on a single host (`deploy/compose/docker-compose.yml`), nginx gateway on `:80`. See `ops/DEPLOY.md`.
**Operator model:** solo operator. No PagerDuty, no on-call rotation, no Kubernetes — the previous K8s-era runbook is archived (`ops/archive/approval-v2-k8s-canary.md`).
**Last verified:** 2026-07-13 against the compose stack and current schema (`db/baseline/0001_current_schema.sql`).

---

## Quick triage (start here for any incident)

```bash
cd deploy/compose

# What is running / healthy?
docker compose --env-file ../../.env ps

# Logs per service (api | worker | jobs | web | gateway | postgres | redis | minio | docx-renderer)
docker compose --env-file ../../.env logs --tail=200 api

# Health surface (through the gateway)
curl -s http://localhost/api/v1/health/live
curl -s http://localhost/api/v1/health/ready

# Scripted end-to-end check (login + auth/me + target route + blank-template object)
pwsh ../../scripts/check-system-runnable.ps1
```

SQL access:

```bash
docker compose --env-file ../../.env exec postgres psql -U metaldocs_app -d metaldocs
```

**Emergency stop (replaces the retired `freeze.sh`):** there is no read-only-mode flag in the v1 API. To stop all traffic: `docker compose stop gateway`. To stop mutations from async work too: `docker compose stop gateway api worker jobs`. Reads/writes resume with `docker compose start ...`.

**Stale-image check (known false-green trap):** before trusting any live diagnosis, confirm the running images were built from current source:

```bash
docker image inspect metaldocs-api:dev --format '{{.Created}}'
git log -1 --format=%cI
```

If the image predates the commit you expect to be running, rebuild before diagnosing further:
`docker compose --env-file ../../.env build --progress plain 2>&1 | tee ../../logs/compose-build.log && docker compose --env-file ../../.env up -d`

---

## Scenario 1: API container down or crash-looping

**Detection:** `docker compose ps` shows `api` restarting/exited; gateway returns 502; `health/live` unreachable.

**Diagnosis:**
1. `docker compose logs --tail=300 api`
2. The API applies DB migrations on boot and exits (`os.Exit(1)`) on any config or migration failure (`internal/platform/migrate/migrate.go`, `apps/api/cmd/metaldocs-api/main.go`). Look for a migration filename or a config-validation error in the last lines.

**Mitigation:**
- Config error → fix `.env`, `docker compose up -d api`.
- Migration failure → do **not** retry in a loop. Take stock: `SELECT * FROM public.schema_migrations ORDER BY version DESC LIMIT 5;`. Fix forward with a corrected migration, or restore from backup (`wiki/runbooks/backup-restore.md`) if data was mutated.
- Crash after clean boot → check Postgres health (`docker compose ps postgres`, `pg_isready` inside the container) and MinIO/Redis dependencies.

---

## Scenario 2: Async work not processing (worker / jobs)

The old Postgres-lease scheduler and `job_leases` flow are **retired** — periodic and scheduled work runs on River inside the `jobs` binary; outbox consumers run in `worker`.

**Detection:** documents stuck in `scheduled`, notifications not fanning out, maintenance jobs silent. Note: `worker`/`jobs` have **no compose healthcheck** — `ps` showing "running" proves nothing; read logs.

**Diagnosis:**
1. `docker compose logs --tail=200 jobs` and `... worker`
2. River queue state:
   ```sql
   SELECT state, count(*) FROM river_job GROUP BY state;
   SELECT id, kind, state, attempt, errors FROM river_job
   WHERE state IN ('retryable','running') ORDER BY id DESC LIMIT 20;
   ```
3. Outbox backlog and dead-letter queue (`processed_at` does not exist on this table — the delivery columns are `published_at` / `dead_lettered_at`):
   ```sql
   SELECT count(*) FILTER (WHERE published_at IS NULL AND dead_lettered_at IS NULL) AS backlog,
          count(*) FILTER (WHERE dead_lettered_at IS NOT NULL)                      AS dead_lettered
     FROM metaldocs.outbox_events;
   ```

**Mitigation:**
- Restart the affected service: `docker compose restart jobs` (or `worker`). Consumers are idempotent by invariant; restart is safe.
- Poison job (high `attempt`, repeating error) → read the error, fix the cause; River discards after max attempts.
- The stuck-instance watchdog is **alert-only** (ADR 0068): it surfaces stuck approval instances, it never auto-cancels. Stuck instances are resolved by an operator decision in the product UI, not by SQL.
- Dead-lettered `outbox_events` rows are retained **90 days**, then purged by the `outbox-events-retention` job; published rows are purged after 7 days. Investigate the DLQ before that window closes. A row that is neither published nor dead-lettered is never purged at any age, so a backlog is never silently eaten by retention. Semantics: [`wiki/database/tables/outbox_events.md`](../../wiki/database/tables/outbox_events.md) §Retention.

---

## Scenario 3: Tripwire firings (P0001 capability errors)

**Detection:** API logs show Postgres `P0001` errors from `enforce_capability_asserted()`; users report 500s on writes.

**Diagnosis:**
1. Find the failing route + capability in api logs (correlate via `X-Trace-Id`).
2. A tripwire firing means a code path wrote a gated table without asserting the arm's capability — this is a **defect** (drift between `internal/platform/tripwire/arms.go` and a call site), not an operational condition.

**Mitigation:**
1. If a single route floods: block it at the gateway (nginx `deny`/location 503) rather than stopping the stack.
2. Fix = code change re-aligning `authz.Require` capability with the tripwire arm; the `TRIPWIRE-ARM-PARITY`/`TRIPWIRE-ARM-DRIFT` CI lints must pass on the fix.
3. Never widen a tripwire arm or disable the trigger to silence the flood.

---

## Scenario 4: Wrong document states (cascade bug)

**Detection:** multiple documents transitioning unexpectedly (e.g. published without signoff).

**Diagnosis:**
```sql
SELECT document_id, event_type, actor_user_id, created_at
FROM public.governance_events
WHERE created_at > now() - interval '1 hour'
ORDER BY created_at DESC LIMIT 50;
```
(Verify column names with `\d public.governance_events` first — schema is source of truth.)

**Mitigation:**
1. Stop the bleeding: `docker compose stop gateway` (whole app) or block the offending route at nginx.
2. Take a backup snapshot **before** any manual repair (`scripts/backup-postgres.ps1`).
3. Prefer restore-from-backup over hand-editing state. Direct `UPDATE public.documents SET status = ...` bypasses the DB transition trigger only if run as a role the trigger exempts — treat manual state surgery as last resort and record every statement.
4. Deploy the fix, rebuild images, `up -d`, re-run `check-system-runnable.ps1`.

---

## Scenario 5: Idempotency conflict spike (409s)

**Detection:** elevated 409 `problem+json` responses on mutating routes.

**Diagnosis:**
```sql
SELECT method, path, count(*) AS keys
FROM metaldocs.idempotency_keys
WHERE created_at > now() - interval '1 hour'
GROUP BY method, path ORDER BY keys DESC LIMIT 20;
```
(Check actual columns with `\d metaldocs.idempotency_keys`.)

**Mitigation:**
1. Usually a client bug (same key, different body). Identify tenant/user from api logs.
2. Abusive tenant → rate limits already apply per the Redis GCRA limiter; tighten env config if needed.
3. Do not delete idempotency rows to "unblock" a client — that re-opens replay windows.

---

## Rollback / restore

- Image rollback: rebuild the previous commit (`git checkout <sha>` in a clean worktree → `docker compose build`) and `up -d`. No registry/tag history exists yet — images are local `:dev` tags.
- Data restore: `wiki/runbooks/backup-restore.md` (`scripts/restore-postgres.ps1`, validated round-trip via `scripts/run-backup-restore-gate.ps1`).
- Post-rollback: `check-system-runnable.ps1` + gateway `:80` smoke before declaring recovery.
