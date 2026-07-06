# Feature F8.1 — Production Dockerfiles + deploy-target truth

> **Milestone:** 8 — Ops Readiness  ·  **Folder:** `f8.1-dockerfiles`
> **Status:** Planning

## Source

- Milestone spec row F8.1 (`../milestone.md`) + validation-contract §1 (binding).

## Plan

**Task A — image hardening + context hygiene (implementer subagent, sonnet)**
Files: `deploy/docker/api.Dockerfile`, `worker.Dockerfile`, `jobs.Dockerfile`, `.dockerignore`.
- Each Dockerfile: keep multi-stage `golang:1.25-alpine` builder; runtime stage `alpine:3.21` adds
  `ca-certificates`, creates non-login user/group (fixed uid/gid, e.g. 10001 `metaldocs`), `USER`
  directive before ENTRYPOINT; api keeps `COPY db/migrations` (load-bearing — startup migrations) with
  correct ownership; worker/jobs stay migration-free.
- `.dockerignore`: add `.env`, `.env.*` (keep existing `.env.local` lines harmless), `docs/`,
  `third_party/`, `**/*.exe`. Do not remove existing exclusions.
- No tests to write (infra files); proof is Task B's build/run. Self-check: `docker build -f ... .`
  for one image locally if fast enough, else defer to Task B.

**Task B — build + boot + runnable proof (main session, live)**
- `docker compose build api worker jobs` from clean tree; capture tail.
- `docker compose up -d postgres redis docx-renderer api`; wait healthy; run
  `.\scripts\check-system-runnable.ps1`; capture transcript.
- Non-root proofs (`id -u` in each image). `docker history` spot-check for `.env`.
- Main session runs this (live-drive class work; docker on the host, long-running).

**Task C — deploy doc truth (implementer subagent, sonnet)**
Files: `ops/DEPLOY.md` (rewrite), `ops/archive/approval-v2-k8s-deploy.md` (re-home), wiki stamp pass.
- New DEPLOY.md: Compose = v1 deployment target; build/up/down/upgrade flow as proven in Task B;
  backup pointer to F8.3 runbook (placeholder link acceptable, filled by F8.3); pg_dump checklist line
  updated to point at runbook.
- Old K8s content moved verbatim to archive file with header note (superseded, why, date).
- wiki-curator-style index/stamp updates if any wiki doc references DEPLOY.md.

Order: A → B → C (C documents the flow B proved). Spec + quality review after A and C; B is evidence.

## Execution notes

(filled during execution)
