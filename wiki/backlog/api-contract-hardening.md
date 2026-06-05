# API Contract Hardening Program

> **Last verified:** 2026-06-05 (Phase A shipped — spec base-path normalization + double-prefix kill + CI gate)
> **Owner discipline:** same as [`roadmap.md`](roadmap.md) — each Phase = one fresh implementation session = one PR series. Ship each Phase without compromise; if a Phase bloats mid-flight, split it (e.g. C → C1 + C2) and update this doc.
> **Why this exists:** Plan 8 ("OpenAPI / contract-first completion", roadmap.md, marked done 2026-05-13) landed the codegen plumbing but left the *contract surface itself* incoherent. A 4-subagent audit (2026-06-05) found systemic drift between the OpenAPI spec, the runtime router, `permissions.go`, the generated FE types, and the live handlers — plus two confirmed authz-enforcement gaps. This program closes that gap and brings the API surface to the same industry-standard bar the auth/middleware/authz refactor already reached.
> **Supersedes:** [`contract-first-followups.md`](contract-first-followups.md) — its "no spec coverage" module list is folded into Phase C/E here.

## How to use this doc

1. Before starting a Phase, read this file + the audit ledger at the bottom. Know what the Phase owns and what decision (if any) must be made first.
2. After finishing a Phase: flip its **Status** to `done YYYY-MM-DD`, link the PR(s), record the **Evidence** (commands run / gates passed), and bump `Last verified`.
3. If scope shifts mid-Phase, edit the Phase's `Scope` / `Out of scope` rows here so the next session sees reality. Do **not** silently grow scope.
4. Every finding in the audit ledger maps to exactly one Phase. When a finding is closed, mark its ledger row `closed`. Nothing gets dropped silently — if a finding is intentionally *not* fixed, record it as `wont-fix` with a one-line reason.

## Anchor decisions (locked)

- **AD-1 — One base path.** `servers.url: /api/v1`; **all** OpenAPI path keys are relative (`/documents`, not `/api/v1/documents`). This matches the existing roadmap anchor "every route lives under `/api/v1/*`" and is the root fix for the double-prefix class of bugs. Settled in Phase A.
- **AD-2 — One error envelope.** RFC 9457 `Problem` (`application/problem+json`) is the only error shape. `ApiErrorEnvelope` is retired. (Continues Plan 7.) Settled in Phase D.
- **AD-3 — Unified authz is the only enforcement model.** Per-resource visibility = `controlled_document_{area,user}_grants` + role/area capabilities. The pre-unification `AccessPolicy` / `document_access_policies` ABAC concept is dead (table dropped in migration 0232) and must not be an enforcement path anywhere. Settled in Phase B.

## Open decisions (must resolve before the named Phase)

- **OD-1 (Phase B):** Is approval-signoff **password re-authentication a real control** (21 CFR Part 11-style e-signature for a regulated QMS)? If **yes** → wire the existing bcrypt `PasswordReauthProvider`. If **no** → delete the dead provider and drop `password_token` from the contract. *Recommendation: yes — a controlled-document QMS almost always requires signed-record reauthentication.*
- **OD-2 (Phase C):** For each unserved spec path, **remove (superseded)** vs **implement (genuinely planned)**. Default = remove, per the access-policies precedent. Candidates that may be "planned not dead": `/notifications`, `/workflow/documents/{id}/transitions|approvals`. Decide per-path with the same forensic check used for access-policies.
- **OD-3 (Phase E):** **One payload casing.** Design-system mandates `snake_case`; the live spec mixes PascalCase (Go-struct leakage in the documents module), camelCase, and snake_case. Measure live FE consumption before locking, because flipping casing is a breaking change to every consumer. Decide: migrate-to-snake_case (design-system-true) vs ratify-camelCase (least FE churn). *Recommendation: snake_case for new/changed ops; PascalCase leakage is a bug to kill regardless.*

---

## Execution order

| Prio | Phase | Title | Risk | Status |
|------|-------|-------|------|--------|
| P0 | A | Spec base-path normalization + double-prefix kill + lint gate | Low (mechanical) | done 2026-06-05 |
| P0 | B | Authz enforcement gaps (signoff reauth + search visibility) | Med (security) | pending |
| P1 | C | Dead / unserved surface prune (OWASP API9) | Med (contract shape) | pending |
| P1 | D | Error-envelope unification (finish RFC 9457) | Med (contract shape) | pending |
| P2 | E | Spec hygiene + standards conformance | Med | pending |
| P2 | F | FE transport unification + Go dead-code purge | Low | pending |

**Sequencing rationale:** A first — it is mechanical, fixes a live 404, and unblocks every consumer (you should not edit the contract while it is self-inconsistent). B second — security-grade, but cleaner to wire once the contract is coherent. C and D are the contract-shape redesigns (hard-stop territory → each opens with a written decision). E and F are cleanup that depends on the shape being settled.

---

## Phase A · Spec base-path normalization + double-prefix kill

- **Goal:** The OpenAPI spec is self-consistent and the `/api/v1/api/v1/...` double-prefix bug class is impossible to reintroduce.
- **Why:** `openapi.yaml:7` sets `servers.url: /api/v1`, yet ~half the path keys re-include `/api/v1/` (templates, documents, taxonomy, controlled-documents, approval) while the rest are relative (auth, iam, audit, search). Every spec consumer/SDK/mock builds wrong URLs for half the API; the FE openapi-fetch client doubles the prefix (live 404 on the documents list and `useBlankTemplateQuery.ts:10`).
- **Scope:**
  - Strip the `/api/v1` prefix from every absolute path key in `api/openapi/v1/openapi.yaml` (AD-1).
  - Regenerate FE types (`npm run gen:api`) → `frontend/apps/web/src/lib/api-types/index.d.ts`.
  - Fix every FE caller that hand-prepends `/api/v1` to an openapi-fetch path (confirmed: `features/documents/queries/useBlankTemplateQuery.ts:10`; sweep for siblings).
  - Add a CI lint gate (`npx @redocly/cli lint`, config already present) that fails on a path key starting with `/api/`.
- **Out of scope:** removing dead paths (Phase C), error/security/tag fixes (Phase D/E).
- **Closes (ledger):** F-DBL-1, F-DBL-2, F-SPEC-BASEPATH.
- **Verify:** redocly lint passes; `go build ./...`; FE `tsc`; Preview QA on documents-list + blank-template (the two confirmed double-prefix callers) return 200.
- **Status:** done 2026-06-05.
- **Evidence (2026-06-05):**
  - **Spec:** stripped `/api/v1` from all 76 absolute path keys in `api/openapi/v1/openapi.yaml`; `servers.url: /api/v1` retained; the other ~60 already-relative keys untouched. No duplicate keys after normalization (`grep '^  /' | sort | uniq -d` empty). Only the spec keys changed — no operations/schemas/tags/responses/security touched. Runtime router + `api.gen.go` NOT regenerated → served paths stay `/api/v1/*` (AD-1: spec was wrong, runtime was right).
  - **FE codegen:** `npm run gen:api` → `src/lib/api-types/index.d.ts` regenerated; 0 `/api/v1`-prefixed `paths` keys remain.
  - **FE callers fixed:** openapi-fetch (`api` client, baseUrl `/api/v1`) double-prefix offenders → relative: `useBlankTemplateQuery.ts:10` (`/templates/system/blank`), `library.ts:35,47` (`/documents`, `/documents/stats`). Type-key lookups `paths['/api/v1/...']` → `paths['/...']` in `documents.ts`, `templates.ts`, `templates/catalog.ts`. String `apiFetch('/api/v1/...')` callers left as-is (guard `apiUrl` passes `/api/`-prefixed paths through unchanged → no doubling; Phase F owns transport migration).
  - **CI gate (blocking):** new `PATH-BASE-PREFIX` rule in `scripts/api-lint/spec_rules.go` (`checkBasePrefix`) + `-only RULE` filter in `main.go`; wired as blocking job `spec-base-path-gate` in `.github/workflows/api-contract.yml`. Unit test `path_base_prefix` + testdata `path_base_prefix.openapi.yaml`. Proven both directions: `go run ./scripts/api-lint -only PATH-BASE-PREFIX api/openapi/v1/openapi.yaml` → `0 violation(s)` exit 0; same against the bad testdata → 1 violation exit 1.
  - **Gates:** `npx @redocly/cli lint api/openapi/v1/openapi.yaml` → valid; `go build ./...` clean; `go test ./scripts/api-lint/...` ok; `npx tsc --noEmit` (frontend/apps/web) clean.
  - **Preview QA (Vite dev :4173 → API :8081, admin login):** documents library `GET /api/v1/documents?page=1&pageSize=20 → 200` + `GET /api/v1/documents/stats → 200` (single prefix); new-document wizard blank-template `GET /api/v1/templates/system/blank → 500` (**single** prefix — no doubling; 500 is an unrelated pre-existing SQL defect, see defer); doubled `GET /api/v1/api/v1/documents → 404` confirms the bug class is dead. API-level: doubled path 404, both single paths reach their handlers.
  - **Bounded defer (out of scope, NOT a Phase A regression):** templates module 500s on all reads — `internal/modules/templates/repository/postgres.go` selects `lv.revision_number`, a column absent from `templates_template_version` (SQLSTATE 42703). Pre-existing on `main`; Phase A made no Go/SQL changes. Spawned as a separate task. Phase A's deliverable (route resolves single-prefix) is met; the 200 on blank-template is blocked only by this defect.

## Phase B · Authz enforcement gaps (SECURITY)

- **Goal:** Every authz decision flows through the unified model (AD-3); no accepted-but-unverified credential; no open-by-default resource read.
- **Why:** Two confirmed gaps, both tails of the dead ABAC slice.
  - **B1 — Signoff password reauth not enforced.** `approval/http/signoff_handler.go:90-91` stuffs `password_token` into a signature payload; `approval/application/decision_service.go RecordSignoff` only validates + marshals + stores it (lines 112/230/259) — **never verifies it**. The bcrypt `PasswordReauthProvider` + `Registry` are fully built but never wired.
  - **B2 — Search has no per-document visibility.** v2 search is wired in prod (`main.go:219`); `search/application/service.go decidePolicies` default-allows when no policies (line 143-144) and the v2 reader's `ListAccessPolicies` returns nil — so any authenticated user finds every tenant document, ignoring `controlled_document_{area,user}_grants`. The whole `AccessPolicy` search path is now a no-op after migration 0232.
- **Scope (gated on OD-1):**
  - B1: wire `PasswordReauthProvider` (bcrypt against `iam_users.password_hash`) into the signoff path via the existing `Registry`, OR delete the provider + drop `password_token` from the contract — per OD-1.
  - B2: replace the dead `AccessPolicy` search path with enforcement against the unified model (area capabilities + `controlled_document_{area,user}_grants`); remove `decidePolicies`/`ListAccessPolicies`/`shouldBypassPolicy` dead-ABAC code and the dead `AccessPolicy` domain type from search.
- **Out of scope:** non-search dead code (Phase F).
- **Closes (ledger):** F-SEC-REAUTH, F-SEC-SEARCH, F-DEAD-SEARCHPOLICY.
- **Verify:** new tests — signoff rejects a bad/blank `password_token`; search omits documents the caller has no grant for. `go test ./internal/modules/documents/approval/... ./internal/modules/search/... -race`.
- **Status:** pending. **Hard-stop note:** if B2 implies a cross-module authz-model change beyond search wiring, stop and report the boundary per CLAUDE.md.

## Phase C · Dead / unserved surface prune (OWASP API9)

- **Goal:** Every spec path has a handler; every handler is in the spec; every `permissions.go` row maps to a real route. No published-but-unserved endpoints.
- **Why:** ~30 spec paths have no handler (entire legacy taxonomy bloc — `document-profiles`×11, `process-areas`×3, `document-subjects`×4, `document-types`, `document-families`, `document-departments`×4, `document-areas` — all superseded by the canonical `/taxonomy/*`; plus `notifications`×2, `workflow/documents/*`×2, `operations/stream`, `attachments/{id}/content`, `telemetry/mddm-shadow-diff`). Phantom `permissions.go` rows (`auth/refresh`, `POST /documents`, `documents/{id}/artifact-metadata`) classify routes that cannot be reached. ~22 orphaned component schemas. Same liability class as the `access-policies` slice already removed.
- **Scope (gated on OD-2, per-path):**
  - Remove superseded legacy-taxonomy paths + their schemas + their `permissions.go` rows (mirror the access-policies removal playbook: spec → FE regen → permissions → tests → baseline/dictionary if a table is involved).
  - Remove phantom `permissions.go` rows.
  - Remove orphaned `components/schemas`.
  - For each "planned not dead" path (OD-2): either implement + spec it, or move it to a tracked deferred list with `x-deferred` + a reason.
  - Document the handler-only endpoints that *should* be in the spec (`/iam/presence/stream`, `/healthz`, audit export sub-routes, `documents/{id}/pdf-complete`) — add them or annotate why omitted.
- **Closes (ledger):** F-DEAD-TAXPATHS, F-DEAD-MISCPATHS, F-PHANTOM-PERMS, F-ORPHAN-SCHEMAS, F-UNDOC-HANDLERS.
- **Verify:** a generated route truth-table shows zero spec-without-handler and zero handler-without-spec (excluding intentionally-annotated); redocly lint passes; full test suite green.
- **Status:** pending. **Hard-stop note:** contract-shape change — each removal opens with the per-path forensic check, not a blanket delete.

## Phase D · Error-envelope unification

- **Goal:** One error shape across the whole API (AD-2). A single FE error handler works everywhere.
- **Why:** `ApiErrorEnvelope` (legacy) and `Problem` (RFC 9457) coexist; older paths still emit the legacy shape; **no operation documents a `500`**. Plan 7 rolled out `Problem` for the handlers it touched but the contract still advertises both.
- **Scope:**
  - Migrate remaining `ApiErrorEnvelope` responses to `Problem` in both the spec and any handler still emitting the old shape.
  - Document the standard error set (400/401/403/404/409/422/500) once via a reusable `$ref`, applied consistently.
  - Retire the `ApiErrorEnvelope` schema + the hand-written FE `ApiErrorEnvelope` type once unreferenced.
- **Out of scope:** tags/security/pagination (Phase E).
- **Closes (ledger):** F-ENVELOPE-SPLIT, F-NO-500.
- **Verify:** zero `ApiErrorEnvelope` refs in spec + FE; sample error responses across modules return `application/problem+json`; FE single error parser handles all.
- **Status:** pending.

## Phase E · Spec hygiene + standards conformance

- **Goal:** The spec reads like a big-company public API: discoverable, secure-by-default in-doc, one convention per concern.
- **Why (audit scorecard FAILs):** no global `security:` block (125+ ops implicitly unauthenticated in-doc); ~30 untagged operations + no top-level `tags` declarations; three pagination patterns with `limit` max 200 vs design-system 100; payload casing triple-split; `DELETE` with required body + query-param `DELETE`; `bearerAuth` referenced in wiki but absent from spec; no `operationId`/`summary` on a block of ops; no examples.
- **Scope (gated on OD-3 for casing):**
  - Global `security:` block + explicit `security: []` on public ops.
  - Top-level `tags` with descriptions; tag every operation.
  - One pagination convention (cursor canonical; `x-pagination-exempt` for bounded lists); clamp `limit` max to 100 to match the design system.
  - Casing per OD-3; at minimum kill the PascalCase Go-struct leakage in the documents module schemas.
  - Fix non-standard methods: `DELETE /iam/area-memberships` (query-param identity), `DELETE /approval/routes/{id}` (required body).
  - `operationId` + `summary` on every operation; consistent verb naming.
  - Enforce the redocly lint gate in CI (built on Phase A's gate).
- **Closes (ledger):** F-NO-GLOBAL-SEC, F-UNTAGGED, F-PAGINATION, F-LIMIT-200, F-CASING, F-DELETE-SHAPE, F-OPID-SUMMARY, F-BEARERAUTH.
- **Verify:** redocly lint clean with hygiene rules enabled; FE regen + `tsc` green.
- **Status:** pending. **Hard-stop note:** casing migration is breaking — do not start until OD-3 is decided + FE consumption measured.

## Phase F · FE transport unification + Go dead-code purge

- **Goal:** One FE API transport; no hand-written type duplication; internal Go dead code removed.
- **Why:** `approval/*` + `templates/*` FE use raw `fetch()` bypassing `apiFetch` (no auth-expiry dispatch, no Problem parse, no credentials, no idempotency); `lib/types/index.ts` hand-duplicates generated types (`DocumentListItem` vs generated `DocumentSummary`, etc.); approval mutations skip TanStack cache invalidation. Backend carries unreachable deprecated branches.
- **Scope:**
  - Route all FE API calls through the shared `apiFetch` / openapi-fetch client; delete raw-`fetch` clients in `approval/` + `templates/`.
  - Delete hand-written `lib/types` interfaces that duplicate generated types; keep only genuinely-unspecced shapes (and spec those in Phase C/E).
  - Add cache invalidation to approval mutations; convert ad-hoc mutations to `useMutation`.
  - Backend dead-code purge: unreachable `Freeze()` + `pdfDispatcher` branches, `legacyFanout` variadic, `CheckReplay/RecordReplay`, `SnapshotFromTemplate`, `shouldBypassPolicy(){return false}`, `AllowDevTenantFallback` dead field, `GeneratedServerAdapter` shim (once handlers use strict-server), deprecated `authn` context wrappers (29 callers — migrate then delete). Resolve or re-ticket PR-7 / phase11 TODOs.
- **Closes (ledger):** F-FE-RAWFETCH, F-FE-DUPTYPES, F-FE-NOINVALIDATE, F-GO-DEADCODE.
- **Verify:** no raw `fetch(` in feature API layers; `lib/types` contains only unspecced shapes; `go vet` / staticcheck clean; full FE + BE suites green.
- **Status:** pending.

---

## Audit findings ledger

Source: 4-subagent audit, 2026-06-05. Severity from the audit. Every row maps to one Phase; status starts `open`.

| ID | Severity | Finding | Phase | Status |
|----|----------|---------|-------|--------|
| F-SPEC-BASEPATH | CRITICAL | `servers.url:/api/v1` + half the path keys re-prefix `/api/v1` → self-inconsistent spec | A | closed 2026-06-05 |
| F-DBL-1 | CRITICAL | `useBlankTemplateQuery.ts:10` doubles prefix via openapi-fetch client | A | closed 2026-06-05 |
| F-DBL-2 | CRITICAL | documents-list call resolves to `/api/v1/api/v1/documents` → 404 | A | closed 2026-06-05 |
| F-SEC-REAUTH | CRITICAL | signoff `password_token` accepted + stored, never verified; bcrypt provider dead | B | open |
| F-SEC-SEARCH | CRITICAL | v2 search open-by-default; ignores `controlled_document_*` grants | B | open |
| F-DEAD-SEARCHPOLICY | HIGH | dead `AccessPolicy` path in search (`decidePolicies`/`ListAccessPolicies` no-op) | B | open |
| F-DEAD-TAXPATHS | CRITICAL | ~22 legacy taxonomy spec paths, no handlers (superseded by `/taxonomy/*`) | C | open |
| F-DEAD-MISCPATHS | HIGH | `notifications`, `workflow/*`, `operations/stream`, `attachments/content`, `telemetry/*` spec paths, no handlers | C | open |
| F-PHANTOM-PERMS | HIGH | `permissions.go` rows with no route/spec (`auth/refresh`, `POST /documents`, `artifact-metadata`) | C | open |
| F-ORPHAN-SCHEMAS | HIGH | ~22 `components/schemas` never `$ref`'d | C | open |
| F-UNDOC-HANDLERS | MEDIUM | handlers absent from spec (`presence/stream`, `healthz`, audit export sub-routes, `pdf-complete`) | C | open |
| F-ENVELOPE-SPLIT | CRITICAL | `ApiErrorEnvelope` (legacy) vs `Problem` (RFC 9457) coexist | D | open |
| F-NO-500 | HIGH | no operation documents a `500` response | D | open |
| F-NO-GLOBAL-SEC | CRITICAL | no global `security:` block; 125+ ops implicitly unauthenticated in-doc | E | open |
| F-UNTAGGED | CRITICAL | ~30 ops untagged; no top-level `tags` declarations | E | open |
| F-PAGINATION | HIGH | three pagination patterns coexist | E | open |
| F-LIMIT-200 | HIGH | spec `limit` max 200 vs design-system 100 | E | open |
| F-CASING | HIGH | payload casing triple-split (PascalCase/camelCase/snake_case) | E | open |
| F-DELETE-SHAPE | HIGH | `DELETE /iam/area-memberships` query-param identity; `DELETE /approval/routes/{id}` required body | E | open |
| F-OPID-SUMMARY | MEDIUM | block of ops missing `operationId`/`summary`; inconsistent verb naming | E | open |
| F-BEARERAUTH | HIGH | `bearerAuth` in wiki/design-system but not declared in spec | E | open |
| F-FE-RAWFETCH | HIGH | `approval/*` + `templates/*` FE use raw `fetch()` bypassing `apiFetch` | F | open |
| F-FE-DUPTYPES | HIGH | `lib/types` hand-duplicates generated types | F | open |
| F-FE-NOINVALIDATE | MEDIUM | approval mutations skip TanStack cache invalidation | F | open |
| F-GO-DEADCODE | LOW-MED | unreachable deprecated branches / dead fields / migration shims (see Phase F scope) | F | open |

## Closing verification gate (mandatory — re-audit)

Do **not** call the program done on Phase F completion alone. Re-run the **same 4-dimension audit** that opened it, so coherence is proven, not assumed:

1. **Route truth-table & drift** — runtime router vs `api.gen.go` vs `openapi.yaml` vs `permissions.go`.
2. **OpenAPI spec quality** vs industry standard (operationId/tags/responses/error envelope/security/components/versioning/pagination).
3. **Dead/legacy/workaround/stub hunt** in the handler layer.
4. **FE consumption + double-prefix** root-cause sweep.

Run as parallel subagents (the original fan-out). **Pass bar:** zero CRITICAL, zero HIGH; every ledger row `closed` or an explicit `wont-fix` with reason; redocly lint clean; full BE + FE suites green; Preview QA on the touched surfaces. Record the re-audit result here as evidence. Any new CRITICAL/HIGH → a new ledger row + a new Phase, not a silent patch.

## On close of the program

- Pass the closing re-audit gate above and record its evidence.
- Author the ADRs the locked anchors imply: spec base-path rule (AD-1), error-envelope finalization (AD-2 closes Plan 7), unified-authz-only enforcement (AD-3 extends ADR 0022). Add to the Plan 13 ADR sweep or as standalone ADRs.
- Dispatch `wiki-curator` to refresh `Last verified` stamps on every wiki doc whose referenced code changed.
- Update [`roadmap.md`](roadmap.md): mark Plan 8 as superseded-by this program, note the partial-completion correction.
