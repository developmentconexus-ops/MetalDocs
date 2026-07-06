# Feature F8.1 — Spec (production Dockerfiles + deploy-target truth)

> **Milestone:** 8 — Ops Readiness  ·  **Folder:** `f8.1-dockerfiles`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-05 — operator delegated approval via the M8 `/goal` brief; the
> binding shape is `../validation-contract.md` §1 (committed `218e2d12` before any code).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | none needed — the contract is fully explicit in the operator-locked M8 `/goal` brief (F8.1 acceptance) + committed validation-contract §1; remaining unknowns were resolved by runtime verification, below | — |
| 2 | (verified) Is the `db/migrations` COPY in api.Dockerfile load-bearing? | YES — api applies startup migrations (`internal/platform/config/migration.go:21` default `db/migrations`; `main.go:237`). Keep the COPY. |
| 3 | (verified) Is Docker available on the box? | YES — engine 29.5.2, compose v5.1.4. |
| 4 | (verified) Does a root `.dockerignore` exist and is it sufficient? | EXISTS but gaps: `.env` itself NOT excluded (only `.env.local`/`.env.*.local`); `docs/` and `third_party/` not excluded (`third_party` verified NOT a Go build input); `**/*.exe` not excluded (stray `apps/api/cmd/metaldocs-api/metaldocs-api.exe` in tree). |
| 5 | (verified) Deploy-doc state? | `ops/DEPLOY.md` is K8s-targeted and Approval-v2-scoped; contradicts compose. No general deploy doc. |

## Consumer contract (FIRST)

- **Consumers:** (a) operator deploy flow — `docker compose build` / `docker compose up` from
  `deploy/compose/docker-compose.yml`, which references `deploy/docker/{api,worker,jobs}.Dockerfile`
  at fixed paths (compose lines 115/184/219 — paths must not move); (b)
  `scripts/check-system-runnable.ps1`, which probes `http://127.0.0.1:8081` (login → `/auth/me` →
  target route → blank-template seed) and must pass unchanged against the containerized api;
  (c) the F8.2/F8.3 live drives, which run on this containerized stack.
- **Contract:** images build from clean tree, run as non-root, boot to healthy under compose with the
  same runtime behavior the PowerShell dev path produces (same env contract, port 8081 for api);
  build context contains no secrets (`.env*`) and no multi-hundred-MB dev trees.
- **Source of truth:** `deploy/compose/docker-compose.yml` service definitions + validation-contract §1.

## What this feature implements

Harden the three existing Go Dockerfiles to production standard (non-root `USER`, pinned bases,
no cargo-cult layers) at their current paths; close the `.dockerignore` gaps (`.env` class, `docs/`,
`third_party/`, `**/*.exe`); prove build + boot + runnable-check against containers; declare Docker
Compose the v1 deployment target in a rewritten deploy doc, re-homing (not deleting) the Approval-v2
K8s content.

## Non-goals (mandatory)

- No K8s manifests/Helm; no CI docker-build gate; no web/docx-renderer image changes.
- No compose topology changes beyond what boot proof requires (F8.2 wires Redis env — not here).
- No change to local dev path (PowerShell scripts stay canonical for dev).
- No digest-pinning ceremony if tag-pinning is recorded as the chosen level.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| 3 images build clean-tree | `docker compose build api worker jobs` exit 0, transcript tail | real |
| Containerized stack boots; runnable check passes | `docker compose up -d` (postgres, redis, docx-renderer, api) then `.\scripts\check-system-runnable.ps1` all checkpoints PASS | real |
| Non-root runtime | `docker compose exec api id -u` ≠ 0; `docker run --rm --entrypoint id <worker/jobs image>` ≠ 0 | real |
| No secrets/dev-bloat in context | `.dockerignore` contains `.env`, `.env.*`, `docs/`, `third_party/`, `**/*.exe`; `docker history` of api image shows no `.env` layer | real |
| Deploy-target truth | rewritten deploy doc names Compose as v1 target, zero live kubectl instructions; K8s Approval-v2 content re-homed with pointer | real (doc review) |

## ADR needed?

- [x] No durable decision — deploy-target declaration is documentation of the already-shipped compose
  stack; the K8s trigger is recorded in ADR 0071 (F8.2) alongside the scale-out trigger.
