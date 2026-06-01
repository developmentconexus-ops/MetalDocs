# Architecture: Tech Stack

> **Last verified:** 2026-06-01
> **Status:** Stub. Bump versions + add rationale notes when each piece stabilizes.
> **Scope:** Languages, frameworks, infra components, third-party libs.

## Backend

- **Go** — primary backend language. Module-per-domain layout under `internal/modules/`.
- **Postgres** — single relational store. Migrations in `internal/platform/db/migrations/`.
- **MinIO** (S3-compatible) — object storage for frozen DOCX and rendered PDFs.
- **Gotenberg** — DOCX → PDF conversion service (called from `apps/docx-renderer`).
- **Outbox pattern** — `outbox_events` table + `internal/platform/worker/` runners for async work.

## Frontend

- **React + TypeScript + Vite** — `frontend/apps/web/`.
- **TanStack Query** — server state.
- **Material UI / styled-components** (verify) — UI primitives.
- **eigenpal docx-js-editor** — controlled package wrapping ProseMirror for DOCX editing. Vendored + patched (see `references/eigenpal-controlled-package.md`).

## Auxiliary apps

- `apps/docx-renderer/` — TypeScript Node service. Runs eigenpal headless for token substitution. Calls Gotenberg.

## Dev infra

- Docker Compose (Postgres, MinIO, Gotenberg, API, docx-renderer).
- pnpm workspaces (frontend monorepo).
- pnpm patch protocol for vendoring eigenpal.

## See also

- [architecture/system-overview.md](system-overview.md)
- [architecture/deployment.md](deployment.md)
- [references/local-dev-startup.md](../references/local-dev-startup.md)
- [references/eigenpal-controlled-package.md](../references/eigenpal-controlled-package.md)
