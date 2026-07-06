# Feature F8.1 — Evidence

> **Milestone:** 8 — Ops Readiness  ·  **Feature:** `f8.1-dockerfiles`  ·  **Closed:** 2026-07-06
> **Contract:** `spec.md` (production Dockerfiles + deploy-target truth; consumers = compose deploy
> flow, `check-system-runnable.ps1`, F8.2/F8.3 live drives).

## What was implemented

- **Hardened the three Go Dockerfiles** at their compose-referenced paths
  (`deploy/docker/{api,worker,jobs}.Dockerfile`): multi-stage `golang:1.25-alpine` →
  `alpine:3.21`, non-root `USER` (uid/gid 10001 `metaldocs`), pinned bases, api keeps the
  load-bearing `db/migrations` COPY (startup migrations, `config/migration.go:21`).
- **Closed `.dockerignore` gaps**: `.env` + `.env.*` (secret class — previously only `.local`
  variants), `docs/`, `third_party/` (verified not a Go build input), `**/*.exe` (stray dev binary).
- **Deploy-target truth**: `ops/DEPLOY.md` rewritten — Docker Compose declared the v1 deployment
  target, zero live kubectl instructions; Approval-v2 K8s content re-homed to
  `ops/archive/approval-v2-k8s-canary.md` with pointer (commit `74a08e61`).
- **Compose runtime fix surfaced by the live boot**: api `METALDOCS_GOTENBERG_URL` must be the
  Docker-DNS `http://gotenberg:3000`, not the dev `.env` localhost value (the `/health/ready`
  gotenberg probe runs inside the api container); + `depends_on: gotenberg: service_healthy`
  (commit `b15a4480`).
- Producer matches consumer contract: compose references the Dockerfiles at unchanged paths;
  the runnable check passes unchanged against the containerized api on 8081.
- Commits: `1bbb59be` (Dockerfiles hardening + .dockerignore within the M8 batch),
  `74a08e61` (DEPLOY.md rewrite + K8s re-home), `b15a4480` (compose gotenberg URL fix).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| 3 images build clean-tree | `docker compose build api` / `build worker jobs` | exit 0 both: `Image metaldocs-api:dev Built`, `Image metaldocs-worker:dev Built`, `Image compose-jobs Built` | real |
| Containerized stack boots + runnable check | `docker compose up -d` then `.\scripts\check-system-runnable.ps1` (unmodified, api container on host 8081) | all checkpoints PASS: `blank-template-object`, `login-endpoint` (200), `login-session` (1 cookie), `auth-me` (200), `target-route /api/v1/health/ready` (200) | **real** |
| Non-root runtime | `docker exec metaldocs-api id`; `docker run --rm --entrypoint id metaldocs-worker:dev` / `compose-jobs` | all three: `uid=10001(metaldocs) gid=10001(metaldocs)` | real |
| No secrets/dev-bloat in context | `.dockerignore` grep; `docker history --no-trunc metaldocs-api:dev \| grep .env` | entries present: `.env` (67), `.env.*` (68), `docs/` (72), `third_party/` (76), `**/*.exe` (80); history: **no `.env` layer** | real |
| Deploy-target truth | doc review `ops/DEPLOY.md` + `ops/archive/approval-v2-k8s-canary.md` | DEPLOY.md names Compose as v1 target, no live kubectl; K8s content re-homed with pointer | real (doc review) |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| 3 images build clean-tree | yes | build transcripts exit 0 |
| Containerized stack boots; runnable check passes | yes | check-system-runnable all PASS vs containers |
| Non-root runtime (api exec + worker/jobs entrypoint id) | yes | uid=10001 all three |
| No secrets/dev-bloat in context | yes | .dockerignore entries + clean docker history |
| Deploy-target truth (Compose v1, K8s re-homed) | yes | DEPLOY.md rewrite + ops/archive pointer |

## Review disposition

- Spec-compliance review: PASS — no K8s manifests/Helm, no CI gate, no web/docx-renderer image
  changes, dev PowerShell path untouched; tag-pinning recorded as the chosen level (no digest
  ceremony). One compose env fix (gotenberg URL) was boot-proof-required, within the non-goal
  boundary ("no topology changes beyond what boot proof requires").
- Code-quality review: PASS — multi-stage keeps runtime images minimal; single `USER` convention
  (10001) across all three; migrations COPY justified by verified runtime dependency.

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| CI docker-build gate | Non-goal per spec; local builds proven | Add when CI pipeline lands (program-level ops backlog) |
| Digest pinning | Tag-pinning chosen + recorded; reproducibility adequate for v1 | Revisit at first supply-chain hardening pass |
