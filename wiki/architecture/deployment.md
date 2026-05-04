# Architecture: Deployment

> **Last verified:** 2026-05-01
> **Status:** Stub. Expand with prod topology, env-var matrix, and SRE runbooks when prod target is locked.
> **Scope:** How MetalDocs runs locally and (eventually) in prod.
> **Key files:**
> - `docker-compose.yml` — local stack
> - `scripts/start-api.ps1` — Windows dev launcher
> - `internal/platform/config/` — env-driven config

## Local dev

Single command path:

```powershell
.\scripts\start-api.ps1        # starts API on :8081
.\scripts\start-api.ps1 -Build # rebuild binary first
docker compose up -d           # Postgres, MinIO, Gotenberg, docgen-v2
```

Frontend:

```powershell
cd frontend\apps\web
npm.cmd run dev
```

See [references/local-dev-startup.md](../references/local-dev-startup.md) for the canonical procedure.

## Services + ports

| Service     | Port (default) | Purpose                              |
|-------------|----------------|--------------------------------------|
| API (Go)    | 8081           | Main backend                         |
| Frontend    | 5173           | Vite dev server                      |
| Postgres    | 5432           | Relational store                     |
| MinIO       | 9000 / 9001    | Object store + console               |
| Gotenberg   | 3000           | DOCX → PDF                           |
| docgen-v2   | (internal)     | Token substitution + fanout dispatch |

## Env vars

TBD — extract from `internal/platform/config/` and `docker-compose.yml`.

## Prod (TBD)

- Container orchestration target: TBD.
- Secret management: TBD.
- Observability: TBD.

## See also

- [architecture/tech-stack.md](tech-stack.md)
- [references/environment-setup.md](../references/environment-setup.md)
