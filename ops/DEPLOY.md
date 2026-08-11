# Deploy Runbook — v1 (Docker Compose)

> **v1 deployment target: Docker Compose** (`deploy/compose/docker-compose.yml`).
> Kubernetes is a **post-v1 decision** — not built, not the current path. Any future move to
> Kubernetes must land as a named ADR trigger against ADR 0071 (multi-replica rate-limit store)
> before it is adopted, since ADR 0071 already assumes a shared Redis-backed rate limiter for
> when the API scales beyond a single Compose replica.
>
> Historical: the Approval-v2 K8s canary procedure is archived at
> [`ops/archive/approval-v2-k8s-canary.md`](archive/approval-v2-k8s-canary.md) — retained for
> reference, not runnable against the current stack.

**SRE sign-off required before production deploy.**

---

## Stack Overview

`deploy/compose/docker-compose.yml` defines the full v1 stack:

| Service | Role | Notes |
|---|---|---|
| `postgres` | Primary DB | healthcheck-gated (`pg_isready`) |
| `redis` | Rate-limit store, shared across API replicas (ADR 0071) | healthcheck-gated |
| `minio` | S3-compatible object storage (attachments, docx templates) | + `minio-init` one-shot bucket bootstrap |
| `gotenberg` | PDF rendering backend for docx-renderer | healthcheck-gated |
| `docx-renderer` | Internal-only render/fanout service | healthcheck-gated; depends on minio + gotenberg |
| `api` | `metaldocs-api` — sync + authz, stateless; also hosts the 4 leader-elected janitors | healthcheck-gated (`/api/v1/health/live`) |
| `worker` | `metaldocs-worker` — async outbox consumers | depends on healthy api + docx-renderer |
| `jobs` | `metaldocs-jobs` — River-scheduled publish + notifications fanout | depends on healthy api |
| `web` | Frontend static/preview server | healthcheck-gated |
| `gateway` | nginx — public entrypoint (port 80) | depends on healthy api + web |

Three Go images (`api`, `worker`, `jobs`) build non-root per F8.1.

---

## Prerequisites

- Docker Engine + Docker Compose v2 (`docker compose` CLI, not the legacy `docker-compose` binary).
- A populated env file for compose (default `.env` at repo root). Start from
  [`.env.example`](../.env.example) — **never** inline secrets in this doc or in compose files.
  Required env groups (see `.env.example` for the full variable list and defaults):
  - **Postgres**: `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_HOST_PORT`
  - **MinIO**: `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, `METALDOCS_MINIO_BUCKET`, `METALDOCS_MINIO_*`
  - **Auth/session**: `METALDOCS_AUTH_ENABLED`, `METALDOCS_AUTH_SESSION_SECRET`,
    `METALDOCS_AUTH_SESSION_COOKIE_NAME`, `METALDOCS_AUTH_SESSION_TTL_HOURS`,
    `METALDOCS_AUTH_SESSION_IDLE_MINUTES`, `METALDOCS_AUTH_COOKIE_SECURE`,
    `METALDOCS_AUTH_PASSWORD_MIN_LENGTH`, `METALDOCS_AUTH_LOGIN_MAX_FAILED_ATTEMPTS`,
    `METALDOCS_AUTH_LOGIN_LOCK_MINUTES`
  - **Bootstrap admin**: `METALDOCS_BOOTSTRAP_ADMIN_ENABLED`, `METALDOCS_BOOTSTRAP_ADMIN_USER_ID`,
    `METALDOCS_BOOTSTRAP_ADMIN_USERNAME`, `METALDOCS_BOOTSTRAP_ADMIN_EMAIL`,
    `METALDOCS_BOOTSTRAP_ADMIN_DISPLAY_NAME`, `METALDOCS_BOOTSTRAP_ADMIN_PASSWORD`
  - **Rate-limit store**: `METALDOCS_RATELIMIT_STORE` (must be `redis` for multi-replica —
    ADR 0071), `METALDOCS_RATELIMIT_REDIS_ADDR`, `METALDOCS_MULTI_REPLICA`
  - **River schema**: `METALDOCS_JOBS_RIVER_SCHEMA` — must match exactly between `api` and `jobs`
    services (both default to the public schema when unset; set explicitly so they cannot diverge)
  - **docx-renderer**: `DOCX_RENDERER_SERVICE_TOKEN`, `DOCX_RENDERER_VERSION`
- Sufficient disk for the named volumes: `metaldocs_postgres_data`, `metaldocs_redis_data`,
  `metaldocs_minio_data`.

---

## Build

Build the three Go images (non-root, F8.1) plus the frontend image:

```bash
docker compose -f deploy/compose/docker-compose.yml build
```

To build a single service (e.g. after an API-only change):

```bash
docker compose -f deploy/compose/docker-compose.yml build api
```

---

## Up

```bash
docker compose -f deploy/compose/docker-compose.yml up -d
```

Startup order is enforced by healthcheck-gated `depends_on`:

1. `postgres`, `redis` become healthy; `minio` starts; `minio-init` bootstraps the bucket and exits.
2. `gotenberg` becomes healthy; `docx-renderer` waits on `minio` + `minio-init` + `gotenberg`, then
   becomes healthy.
3. `api` waits on healthy `postgres` + `redis` + started `minio`, then becomes healthy
   (`GET /api/v1/health/live` inside the container).
4. `worker` and `jobs` wait on healthy `api` (and `worker` also waits on healthy `docx-renderer`).
5. `web` waits on healthy `api`; `gateway` waits on healthy `api` + `web` and publishes port 80.

Check status:

```bash
docker compose -f deploy/compose/docker-compose.yml ps
```

## Down

```bash
docker compose -f deploy/compose/docker-compose.yml down
```

Add `-v` only if you intend to discard the named volumes (Postgres/Redis/MinIO data) — this is
destructive and not part of a normal stop/restart cycle.

---

## Verify

The API is published on host port `${APP_PORT}` mapped to the container's `8081`. With
`APP_PORT=8081` (the default `check-system-runnable.ps1` target), run the runnable check against
the containerized API:

```powershell
.\scripts\check-system-runnable.ps1
```

This checks: system blank-template object present in MinIO, `/api/v1/health/ready`,
login + session cookie, `/api/v1/auth/me`, and the target route (default
`/api/v1/health/ready`). Pass `-TargetRoute` to check a different route once the stack is up.

Do **not** pass `-StartApi` when verifying the Compose stack — that switch drives the local
non-container dev flow (`scripts/dev-api.ps1`), not compose.

### Post-bootstrap checklist (first bring-up of a new database volume)

1. **Rotate the `metaldocs_ci` password — required on every non-dev environment.**
   `db/grants/0001_role_grants.sql` creates the non-owner, `NOSUPERUSER` + `NOBYPASSRLS`
   `metaldocs_ci` role with the **non-secret dev fixture password** `metaldocs_ci_dev`, so a
   fresh bootstrap always has a known-password login role in the cluster. It is DML-only and
   RLS-bound, but a published password is still a published password. On any environment that
   is not a throwaway dev box, rotate it immediately after the first bootstrap:

   ```bash
   docker compose -f deploy/compose/docker-compose.yml exec postgres \
     psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
     -c "ALTER ROLE metaldocs_ci PASSWORD '<deployment-secret>'"
   ```

   Then point the integration suite at the rotated secret via `METALDOCS_CI_DB_PASSWORD`
   (`tests/integration/testdb/ci_role.go`). Re-running the grants stage never resets the
   password — the `CREATE ROLE` is guarded on the role being absent — so the rotation survives
   every subsequent API start. If the role is not needed at all on this environment, `DROP ROLE
   metaldocs_ci` instead; the grants stage will recreate it only if the app's DB user holds
   `CREATEROLE`, and skips cleanly with a `NOTICE` otherwise.

2. **Set `METALDOCS_RUNTIME_DB_PASSWORD` to a real per-environment secret — required on every
   non-dev environment, and required at all to bring the stack up.** Unlike `metaldocs_ci`
   above, `db/grants/0000_identity_roles.sql`'s `CREATE ROLE metaldocs_runtime` bakes in **no**
   password at all (`rolpassword` is `NULL`, fail-closed) — the role cannot authenticate until
   `db-provision` rotates it, which it does unconditionally on every run, using whatever
   `METALDOCS_RUNTIME_DB_PASSWORD` resolves to. `deploy/compose/docker-compose.yml` requires this
   variable (`${METALDOCS_RUNTIME_DB_PASSWORD:?...}`) — `docker compose up`/`config` aborts
   rather than substituting a default if it is unset — but that only proves the variable is
   *set*, not that it is *safe*. `.env.example` ships `METALDOCS_RUNTIME_DB_PASSWORD=
   metaldocs_runtime_dev`, a literal published in this repository (public since 2026-08-07). On
   any environment that is not a throwaway dev box, the deploy `.env` MUST set this to a real
   secret **before the first `docker compose up`** — do not copy the `.env.example` default.

   If the stack was ever brought up with this variable unset-or-defaulted to the published
   literal (check the deploy `.env` and deploy history — this is knowable without touching the
   live database), treat it as a credential rotation event, not a one-line fix:

   ```bash
   # 1. Rotate the LIVE password first, while api/worker/jobs still hold the old one.
   docker compose -f deploy/compose/docker-compose.yml exec postgres \
     psql -U "$POSTGRES_USER" -d "$POSTGRES_DB" \
     -c "ALTER ROLE metaldocs_runtime PASSWORD '<deployment-secret>'"

   # 2. Update METALDOCS_RUNTIME_DB_PASSWORD (and PGPASSWORD, and DATABASE_URL if you use
   #    it — see .env.example's comments on why all three are independent literals) in the
   #    deploy .env to the SAME <deployment-secret>.

   # 3. Restart api/worker/jobs so they pick up the new value. Restarting before step 1
   #    would authenticate with a password the role does not have yet; restarting api/worker/
   #    jobs is what actually invalidates the old, now-rotated-away credential.
   docker compose -f deploy/compose/docker-compose.yml up -d api worker jobs
   ```

---

## Observability — metrics listener isolation (F-R1, Dim-9)

The API process serves **two** listeners:

| Listener | Address (in container) | Serves | Host-published? |
|---|---|---|---|
| Public API | `:8081` (`APP_PORT`) | the versioned API + auth chain | **Yes** — `${APP_PORT}:8081` |
| Metrics | `:9090` (`METRICS_ADDR`) | `GET /metrics` (Prometheus, unauthenticated) | **No** — infra-network only |

`/metrics` is served **only** on the dedicated metrics listener. The public API server has no
`/metrics` route registered, so the scrape surface cannot be reached on `${APP_PORT}` — isolation is
a property of the process topology, not of ingress/firewall configuration. The scrape stays
credential-less by design; it is protected by **not being host-published**, not by auth.

- **Scrape target:** a Prometheus running on the compose (infra) network scrapes `http://api:9090/metrics`.
  Do **not** add a `ports:` mapping for `9090` on the `api` service — that would re-expose the scrape
  surface on the host and re-open the Dim-9 defect.
- **Override:** set `METRICS_ADDR` (e.g. `:9191`) to move the listener; a malformed value fails the
  API fast at boot (same discipline as `APP_PORT`).
- The nginx `gateway` proxies only `/api/` and `/` — it never proxied `/metrics` and still does not.

---

## Upgrade / Rollout

Compose has no native canary/rolling mechanism; upgrades are recreate-in-place:

1. **Back up first.** Follow the pre-upgrade backup step in the backup/restore runbook:
   [`wiki/runbooks/backup-restore.md`](../wiki/runbooks/backup-restore.md).
2. Pull or rebuild the new images:
   ```bash
   docker compose -f deploy/compose/docker-compose.yml build
   # or, if pulling pre-built images tagged via API_IMAGE / WORKER_IMAGE / WEB_IMAGE:
   docker compose -f deploy/compose/docker-compose.yml pull
   ```
3. Recreate the changed services:
   ```bash
   docker compose -f deploy/compose/docker-compose.yml up -d
   ```
   Compose recreates only containers whose image/config changed; healthcheck gating still applies
   to dependents.
4. Re-run `scripts/check-system-runnable.ps1` to confirm the upgraded stack is healthy end-to-end.
5. **Rollback:** re-tag/rebuild the previous image version and repeat step 3, or
   `docker compose down` + restore from the pre-upgrade backup per the backup/restore runbook.

---

## Historical

The Approval-v2 K8s canary procedure (feature-flagged rollout, `kubectl set image` /
`kubectl rollout` / `kubectl set env`) is archived verbatim at
[`ops/archive/approval-v2-k8s-canary.md`](archive/approval-v2-k8s-canary.md). It documents a
Kubernetes deployment path that was never the actual v1 stack and is retained for reference only.

**SRE sign-off:** _______________________ Date: _______
