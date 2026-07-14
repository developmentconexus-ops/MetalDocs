# Reference: Environment Setup

> **Last verified:** 2026-06-01
> **Status:** Stub. See `references/local-dev-startup.md` for the canonical Windows path.
> **Scope:** Local dev environment bootstrap, secrets, seed data.
> **Out of scope:** Prod deployment (see `architecture/deployment.md`), credentials (see `references/local-dev-credentials.md`).

## Canonical procedure

See [references/local-dev-startup.md](local-dev-startup.md) — single source of truth for the Windows PowerShell flow.

## Prereqs

- Docker Desktop (Postgres, MinIO, Gotenberg, docx-renderer containers).
- Go 1.22+ (verify version against `go.mod`).
- Node 20+ + pnpm.
- PowerShell 7+ on Windows; bash/zsh on macOS/Linux (use equivalent commands).

## First-time setup

```powershell
# clone + cd into repo
git clone <url> MetalDocs
cd MetalDocs

# bring up infra
docker compose up -d

# run curated DB bootstrap with local dev seed
.\scripts\dev-bootstrap-baseline.ps1 -WithDevSeed

# start API
.\scripts\start-api.ps1 -Build

# start frontend
cd frontend\apps\web
pnpm install
pnpm dev
```

## Seed data

The local dev seed creates `admin` / `AdminMetalDocs123!`. Product reference data and local dev seed data stay separate.

## See also

- [references/local-dev-startup.md](local-dev-startup.md)
- [references/local-dev-credentials.md](local-dev-credentials.md)
- [references/how-to-run-tests.md](how-to-run-tests.md)
- [architecture/deployment.md](../architecture/deployment.md)
