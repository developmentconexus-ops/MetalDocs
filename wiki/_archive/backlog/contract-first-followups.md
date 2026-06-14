# Contract-First API Codegen — Followups

Migration plan: `docs/superpowers/plans/2026-05-08-contract-first-api-codegen.md`

## Out of scope this rollout

`api/openapi/v1/openapi.yaml` currently tags only 3 modules: `registry`, `documents`, `templates`. The following backend modules still serve HTTP via hand-written request structs and have NO spec coverage:

| Module | Path prefix | Status | Action |
|---|---|---|---|
| `approval` | `/api/v2/approvals/...` | No spec coverage | Author spec ops first → then run Phase 5 codegen migration |
| `taxonomy` | `/api/v2/taxonomy/...` | No spec coverage | Author spec ops → Phase 5 |
| `iam` | `/api/v2/iam/...` | No spec coverage | Author spec ops → Phase 5 |
| `platform` (auth, feature-flags) | `/api/v1/auth/...`, `/api/v1/feature-flags` | No spec coverage | Author spec ops → Phase 5 |

## Why deferred

Phase 5 of the contract-first rollout assumed all modules had spec coverage. They don't. Authoring spec for these modules is non-trivial (each has request/response shape audits, error envelopes, query params) and was not part of the bug-fix-driven contract-first rollout.

## When to action

Pick up when:
- A bug surfaces in one of these modules where hand-written struct drift is suspected (same class as the `documents.name` bug).
- A new endpoint is added to one of these modules — author the spec op as part of the change, not after.

## Migration template per module

1. Author spec ops in `api/openapi/v1/openapi.yaml` with `tags: [<module>]`.
2. Lint via `npx @redocly/cli lint` (config in `redocly.yaml` already excludes pre-existing rule categories).
3. Bootstrap codegen: `internal/modules/<x>/api/{cfg.yaml,gen.go}`.
4. Migrate handlers per registry pattern (commits `c968b8e0`, `9fccd8e7`).
5. Update `apps/api/cmd/metaldocs-api/main.go` to register `<x>api.HandlerWithOptions`.

## Documents module — handler ↔ spec gaps

> See also: [`wiki/modules/documents-tech-debt.md T-002`](../../modules/documents-tech-debt.md#t-002--openapi-spec-drift-on-apiv2documents-routes) — canonical tech-debt record with evidence anchors for all spec/handler drift items below.

Bootstrap landed (commit `81e7ec23`) — codegen wired, `internal/modules/documents/api/api.gen.go` produced. Handler migration NOT done because of pre-existing drift between spec and handlers:

**Handlers with no spec op (need spec authoring before migration):**
- `renameDocument` — PATCH or PUT on document name
- `duplicateDocument` — clone-from-existing
- `listComments`, `createComment`, `updateComment`, `deleteComment` — `/api/v1/documents/{id}/comments` CRUD

**Spec ops with no handler (need impl OR spec removal):**
- `createDocument` — POST /api/v1/documents — **RESOLVED (Phase C, 2026-06-07):** the POST op had already been removed from `openapi.yaml`; Phase C removed the leftover phantom `permissions.go` row + the orphan `CreateDocumentRequest`/`DocumentCreatedResponse`/`ListDocumentsResponse` schemas. Spec `/documents` is now GET-only, matching the handler.
- `renderDocumentPDF` — POST/GET on documents render — the served render path `POST /api/v1/documents/{id}/export/pdf` is spec'd + handler-backed; no separate `renderDocumentPDF` op exists in the live spec (no action needed).

> **Phase C reconciliation (2026-06-07):** the dead/unserved *spec→no-handler* class is now fully closed (see [`api-contract-hardening.md`](api-contract-hardening.md) Phase C). The *handler→no-spec* class for the **un-migrated modules below** (approval, full taxonomy CRUD beyond the registry subset, iam, platform) is unchanged — those modules still serve via hand-written structs and remain Phase E/F work. Phase C only added the two audit-export sub-route ops and annotated three intentional omissions; it did **not** author spec for the no-coverage modules.

## Central wiring TODO

`apps/api/cmd/metaldocs-api/main.go` may still reference the pre-codegen handler registration for `registry`, `documents`, `templates`. After Phase 5 lands:
- Confirm each of the 3 codegen'd modules registers via `<x>api.HandlerFromMux` or `RegisterHandlersWithBaseURL`.
- Smoke each module's primary endpoint.
