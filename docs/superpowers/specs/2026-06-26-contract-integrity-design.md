# Design Spec — API Contract Integrity (Workstream A)

- **Date:** 2026-06-26
- **Mission:** Document Contract & Storage Integrity (Grade-A)
- **Workstream:** A of 2 (sequenced first; unblocks the template editor and E2E)
- **Sibling spec:** [Storage Integrity](2026-06-26-storage-integrity-design.md)
- **Status:** Approved design — pending implementation plan

---

## 1. Problem

The HTTP contract is **asserted, not enforced**. A live defect proves it: the
template editor crashes with `Cannot read properties of undefined (reading 'version')`
because `getTemplateSchemas` reads a `{ data: { version } }` envelope from an
endpoint that ADR-0035 already migrated to a **flat** body.

Root facts (all pre-existing on `main`, unrelated to the eigenpal ACL work):

1. **The spec is half-migrated.** ADR-0035 ("flat typed bodies") was applied to
   *some* endpoints only. Within a single module, same resource family:
   - `GET /templates/{id}` returns **enveloped** (`resp.Data.Template`) —
     `internal/modules/templates/delivery/http/routes_query.go:118`
   - `GET /templates/{id}/versions/{n}` returns **flat** (`writeJSON(dto)`) —
     `routes_query.go:158`
2. **Two coexisting frontend client styles.** A generated typed client
   (`createClient<paths>`, `frontend/apps/web/src/lib/api/client.ts:130`) exists
   alongside legacy hand-rolled `apiFetch<InlineLiteral>` calls. The crash lives
   in the legacy style — `getTemplateSchemas` hard-codes the wrong shape
   (`frontend/apps/web/src/features/templates/api/templates.ts:344-356`).
3. **The client blind-casts.** `apiFetch` does `return (await res.json()) as T`
   (`client.ts:107`) — zero runtime decode, so drift is invisible until a deref
   explodes deep in render.
4. **Generated types are present but bypassed.** `templates.ts:3` imports the
   generated `paths`/`components`, then ignores them via inline generics 8 lines
   later. Availability of types ≠ enforcement.

The same file mixes conventions throughout — `getTemplate` enveloped, `getVersion`
flat, `submit`/`review`/`approve`/`getDocxURL` enveloped, `listTemplates`
defensively probing `data?.templates ?? data?.items` (code that does not trust its
own contract), plus thin legacy shims (`saveDraft`, `presignDocxUpload`,
`presignSchemaUpload`).

`getTemplateSchemas` specifically is **legacy redundancy**: it fetches the same
endpoint `getVersion` already covers, and the field it extracts
(`placeholder_schema`) is already a member of `VersionDTO` (`templates.ts:50`). It
survived only because it was written against the pre-migration enveloped shape and
nothing forced it to match the contract when `getVersion` went flat.

## 2. Goal

One response-shape law, enforced at **compile time**, with no hand-rolled shape
assertions anywhere. A wrong shape must fail the build, not crash at runtime.

## 3. Decisions (with rationale)

| # | Decision | Rationale |
|---|----------|-----------|
| A1 | **Envelope policy:** single-resource → flat body; collection/paginated → `{ data:[...], meta }`. | Industry-standard REST convention; completes ADR-0035's intent; gives lists a structured home for pagination meta. |
| A2 | **Enforcement: compile-time typed client only.** Route every call through the generated `paths` types; lint-ban inline-generic `apiFetch` shape literals. | Converts "shouldn't happen" (discipline) into "can't compile" (structure). Near-zero cost: types already generated, `createClient<paths>` already present, CI drift guard already exists. |
| A3 | **No runtime decode (zod).** | YAGNI. Backend already gates spec-conformance (cilint `noresponsemap`, contract tests, oapi-codegen drift guard). Runtime decode adds a dep + codegen + bundle weight to catch a class the backend already guards. |
| A4 | **Delete legacy duplicates/shims** rather than fix their shapes. | The bug was duplication + drift. Removing the duplicate is the root-cause fix; keeping and "decoding" it preserves the smell. |

## 4. Scope

### 4.1 Contract policy (codify the law)
- Write/extend an ADR (extends ADR-0035) stating the single=flat / collection=enveloped law explicitly.
- Update `wiki/architecture/api-contract.md` with the rule and the enforcement mechanism.

### 4.2 Backend
- **Inventory** every endpoint's current shape (flat vs enveloped) vs the new law.
- **Migrate stragglers** so each single-resource op is flat. Known targets:
  - `GET /templates/{id}` (single) → flat
  - `submit` / `review` / `approve` version ops (return a single version) → flat
  - audit/inventory of remaining modules (documents, controlled-documents, etc.)
- For each migrated endpoint: edit `api/openapi/v1/openapi.yaml` → run codegen
  (`go generate ./...`) → fix the handler to write the regenerated type → update
  contract tests.
- CI drift guard (`.github/workflows/api-contract.yml`) and `tools/cilint` already
  enforce; optionally add a check that single-resource ops do not envelope.

### 4.3 Frontend
- Route **all** calls through the generated typed client; replace every
  `apiFetch<InlineLiteral>` with a typed operation sourced from `paths`.
- **Delete** legacy: `getTemplateSchemas` (fold its placeholder + `lock_version`
  read into the single typed `getVersion`), `presignDocxUpload`,
  `presignSchemaUpload`, `saveDraft`, and `listTemplates` defensive probing.
- **Keep** the placeholder wire↔domain mapper (`placeholderFromWire` /
  `placeholderToWire`) — a legitimate ACL between snake_case wire and camelCase
  domain, not legacy.
- **Lint rule** banning inline-generic shape literals on the fetch transport.
- Investigate `lock_version`: if the OpenAPI `VersionDTO` does not declare it,
  add it to the spec (frontend currently reads an undeclared field — its own
  Disease-A instance).

## 5. Non-goals
- Runtime response validation / zod (decision A3).
- Rewriting the snake_case↔camelCase mapping layer.
- Any storage-key work (that is Workstream B).
- Touching the eigenpal ACL walls (sealed under ADR 0046).

## 6. Acceptance criteria
1. Template editor loads end-to-end (the `getTemplateSchemas` crash class is gone).
2. No `apiFetch<InlineLiteral>` shape assertions remain; lint enforces it.
3. Every single-resource endpoint returns flat; every collection returns enveloped; spec, generated types, and handlers agree (CI drift guard green).
4. `go build ./...`, `go test ./...`, `make test` / pnpm tests, and typecheck all green.
5. Legacy duplicates/shims deleted, not merely patched.

## 7. Risks
- **Backend blast radius:** A touches `openapi.yaml`, generated code, handlers, and
  contract tests — not frontend-only. Mitigate: migrate endpoint-by-endpoint, lean
  on the CI drift guards, keep each endpoint's change atomic.
- **Test fallout:** mocks that assert the old enveloped shape will break. Per
  CLAUDE.md test discipline: delete one-off scaffolding, repair only contract /
  invariant guards; new tests use the canonical framework.

## 8. Test plan
- Backend: per-endpoint contract test asserting the flat/enveloped shape; rely on
  oapi-codegen drift guard for spec↔code agreement.
- Frontend: typed-client smoke per migrated call; the editor-load path exercised in
  the E2E (deferred to post-A, where the editor finally mounts).

## 9. Evidence / references
- ADR 0035 — flat typed responses (`wiki/decisions/0035-flat-typed-responses-and-presign-status.md`)
- `internal/modules/templates/delivery/http/routes_query.go:118,158`
- `frontend/apps/web/src/features/templates/api/templates.ts` (mixed conventions throughout)
- `frontend/apps/web/src/lib/api/client.ts:107,130`
- CI contract guard: `.github/workflows/api-contract.yml`; `tools/cilint`
