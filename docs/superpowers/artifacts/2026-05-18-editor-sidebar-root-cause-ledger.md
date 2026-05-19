# Editor Sidebar Root Cause Ledger

> Last updated: 2026-05-19
> Scope: Runtime evidence and mismatch classification for the editor right sidebar hardening plan.

## Runtime Snapshot

Docker Desktop was reinstalled during this session after the previous engine entered a broken WSL/Desktop state. Current runtime was rebuilt through the canonical MetalDocs path:

- `docker version`: client/server responding on `desktop-linux`
- `docker compose -f deploy/compose/docker-compose.yml --env-file .env up -d postgres redis minio`: passed
- `scripts/dev-bootstrap-baseline.ps1 -WithDevSeed`: passed
- `scripts/seed-system-blank-template.ps1`: passed
- `scripts/check-system-runnable.ps1 -TargetRoute /api/v1/controlled-documents`: passed

Current health evidence:

```text
metaldocs-api process: running
GET /api/v1/health/live: 200
GET /api/v1/auth/me: 200
GET /api/v1/controlled-documents: 200
```

Current authenticated user:

```json
{"userId":"admin","tenantId":"ffffffff-ffff-ffff-ffff-ffffffffffff","username":"admin","displayName":"Administrator","mustChangePassword":false,"roles":["system_admin"]}
```

Current database counts after clean bootstrap:

```text
users: 6
profiles: 0
areas: 0
controlled_documents: 0
documents: 0
editor_sessions: 0
```

Current controlled documents response:

```json
{"items":[]}
```

## Runtime Reset Impact

The pre-reset editor document IDs used earlier in this session no longer exist after the curated bootstrap reset. This invalidates direct reuse of payloads from `d79a1974-ecdd-444e-a5c4-740e46a9bc52` and similar documents.

That is a `runtime prerequisite` discovery, not a feature behavior. The next E2E pass must create taxonomy and documents through real runtime flows before validating editor sidebar states.

## Classification

| ID | Symptom | Classification | Truth layer used | Root cause status | Fix status |
|---|---|---|---|---|---|
| B1 | `POST /api/v1/taxonomy/*` returned `500 internal server error` from UI setup | runtime prerequisite | runtime + wiki + code | confirmed before reset; must revalidate on clean runtime | local patch needs professional review |
| B2 | New editor document opened readonly with "Another user is editing this document" | shared contract prerequisite | runtime + code | partially confirmed before reset; must reproduce with fresh document | not fixed |
| B3 | Editor sidebar draft metadata showed empty profile/area | module-local implementation | runtime + code + contract | confirmed before reset | local patch needs contract review |
| B4 | Revision timeline showed `REVNaN - undefined` | shared contract prerequisite | runtime + OpenAPI/generated types + code | confirmed before reset | local patch needs contract review |
| B5 | Draft sidebar profile label sometimes showed raw `pop` before hydration | screen-local implementation | browser + TanStack Query | unconfirmed on clean runtime | not fixed |
| B6 | Approval instance query was requested while document was draft in one runtime log | screen-local implementation | logs + frontend code | unconfirmed on clean runtime | not fixed |
| B7 | `scripts/start-api.ps1` held the shell until timeout while API was healthy | workflow/tooling gap | execution truth | confirmed on clean runtime | not fixed |

## Next Evidence Needed

- Create taxonomy data through the real UI/API path on the clean runtime.
- Create a document through the real document wizard.
- Re-check `documents.created_by`, `documents.active_session_id`, and `editor_sessions.user_id`.
- Validate editor sidebar draft behavior against runtime API payloads.
- Only then decide whether B2/B5/B6 require code changes.
