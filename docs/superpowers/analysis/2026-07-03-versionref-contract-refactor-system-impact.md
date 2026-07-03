# System-impact analysis — VersionRef value-object contract refactor

**Date:** 2026-07-03
**Intent (one line):** Replace flat coupled version/revision scalars in list/detail contracts with nested version-reference value objects (templates: `latest_version` / `published_version`; documents: `current_revision`), split read models from aggregates, and unify the version-pointer *pattern* across templates and documents.
**Work type:** feature (contract reshaping across two existing modules; no new module, no new capability)
**Author:** developing-new-work skill
**Verdict:** 🟡 Yellow — proceed; ADR required (contract-design policy), breaking pre-v1 wire change flagged.

---

## 1. Classify & own

- **Work type:** feature.
- **Owning module(s):**
  - `templates` — owns `TemplateDTO`, `VersionDTO`, `getTemplate` envelope, list projection (`internal/modules/templates`).
  - `documents` — owns `DocumentSummary`, `DocumentDetailResponse` and the `current_revision_id`/`revision_version`/`revision_number` triple (`internal/modules/documents/api/api.gen.go` is the only generated owner of `DocumentSummary` — verified).
- **Explicitly NOT owning:**
  - `controlleddocuments` — consumes documents via published interfaces; its own DTOs are not in scope (revision history item shape `DocumentRevisionHistoryItem` is a history row, not a pointer — different concept, untouched).
  - `taxonomy` — FE `ProfileEditDialog` consumes the templates list DTO; that's a frontend consumer update, not module ownership.
  - `render`, `distribution`, `search` — snapshot/reference template-version IDs by value; they do not consume the reshaped DTOs.
- **Cross-module edges:** none change. Each module regenerates its own `api.gen.go` from the shared spec; no Go type is shared across module boundaries (oapi-codegen `include-tags` generates independent per-package types). Backend-internal ports unchanged.
- **Ambiguity?** None. AS-3 not triggered.

## 2. Foundation verdict

- **Base:** DB is correctly normalized (`templates_template` stores only `latest_version` int + `published_version_id` FK; revisions live on version rows; list query LEFT-JOIN-projects the rest — `internal/modules/templates/repository/postgres.go:92`). Domain axes (version counter vs regulated revision, ADR 0013) are sound. The **wire contract shape** is the local-maximum artifact: 5 parallel scalars with an implicit tri-field null-coupling invariant, plus a `latest_version` key that is `integer` in `TemplateDTO` but a full `VersionDTO` object in the `getTemplate` envelope (openapi.yaml:5790 vs :5956).
- **Sound or patch?** Foundation sound; contract shape is the defect. The 2026-07-03 HIGH bug (unpublished template selectable in wizard, fixed in 9f86828b) was this shape's first casualty — the fix hardened serialization but left the shape.
- **Global maximum:** nested version-reference value objects (Google AIP-style nested messages, never parallel coupled scalars). This work IS the global-maximum correction; AS-2 not triggered (we are replacing the local maximum, not optimizing inside it).
- **Unification finding (verified against spec):** documents' triple (`current_revision_id`, `revision_version`, `revision_number` — openapi.yaml:5052/5110) is coupled but all **required non-nullable**, and documents have a single current-revision pointer, not templates' latest/published duality. Therefore: **unify the pattern (one nested ref object per pointer), NOT the schema**. Separate `TemplateVersionRef` and `DocumentRevisionRef` component schemas — forcing one shared schema across two bounded contexts would be false unification (DDD: same-looking concept in two contexts stays duplicated unless truly identical; templates' ref needs `status`, documents' does not).

## 3. Invariant alignment

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | No | No route, capability, or authz path changes — response shape only | n/a |
| Contract-first (OpenAPI + oapi-codegen) | **Yes — core of the work** | All shape changes land in `api/openapi/v1/openapi.yaml` first; `go generate ./internal/modules/{templates,documents}/api/...` + FE `pnpm run gen:api`; zero hand-edits to generated code | per-module `cfg.yaml` + `gen.go` |
| Multi-tenant pooled | No | Projection reshaping only; queries and tenant predicates unchanged | n/a |
| Async = transactional outbox | No | No side effects touched | n/a |
| DB enforces invariants | No | No schema/migration change — DB already normalized; the null-coupling invariant becomes structurally unrepresentable at the wire instead | n/a |
| Cross-module via published interface only | Yes (guard) | Each module regenerates its own types; no shared Go DTO package is introduced | oapi-codegen include-tags |

AS-1 not triggered.

## 4. Capability wiring

**N/A** — no capability added or changed.

## 5. Module wiring

**N/A** — no new module.

## 6. Frameworks to reuse, not reinvent

- `testdb` factory (`tests/integration/testdb/`) for any new integration assertions; `//go:build integration`.
- Existing marshal-shape pin-test idiom (`routes_typed_response_f53_test.go`, `template_dto_nullable_fields_test.go`) for the new contract pins — extend/replace, don't invent a new harness.
- `oapi-codegen` + `openapi-typescript` generation paths (existing).
- No tx/error/outbox/tenant primitives touched — read-side reshaping only.
- **New read-model structs** (`TemplateListItem` + `VersionRef` in templates; ref struct in documents) live inside each owning module (application or repository layer), not in a shared platform package — a value object per bounded context is not a cross-cutting framework.

## 7. Contract & data

- **OpenAPI-first:**
  - New component schemas: `TemplateVersionRef` `{id, number, revision_number, status}` (all required); `DocumentRevisionRef` `{id, version, number}` (all required).
  - `TemplateDTO`: replace `latest_version`(int) + `latest_revision_number` with `latest_version: TemplateVersionRef` (required); replace `published_version_id` + `published_version_number` + `current_revision_number` with `published_version: TemplateVersionRef | nullable` (required, present-and-null).
  - `getTemplate` envelope: keeps `latest_version: VersionDTO` (full view). Same field name, compact-ref in list vs full object in detail = AIP view semantics; the int-vs-object collision dissolves.
  - `DocumentSummary` / `DocumentDetailResponse`: replace `current_revision_id` + `revision_version` + `revision_number` with `current_revision: DocumentRevisionRef` (required, non-nullable).
  - Regenerate: templates + documents `api.gen.go`, FE `api-types/index.d.ts`.
- **Migration:** none. No DB change.
- **Destructive change?** Yes — breaking wire change, deliberately WITHOUT expand/contract: no external API consumers exist pre-v1, FE ships from the same repo, and v1 is a clean re-baseline (F-18 plan). Single atomic cutover per module, backend + FE + tests in one reviewed unit. This exception is recorded in the ADR; post-v1 the same change would require versioned deprecation.

## 8. Test & QA plan

- **Backend per module:** rewrite marshal-shape pin tests to the new shape (templates: `published_version: null` key-present pin for never-published; nested ref field-set pin; documents: `current_revision` object pin). Existing contract tests that assert flat keys are updated in the same commit (contract/invariant guards → repair, not delete; one-off scaffolding tests that break → delete per legacy-test policy).
- **FE:** regenerate types; update consumers (templates: `templates.ts`, 2 adapters, `TemplatesListPage`, `StepTemplate` + `StepConfirm`, taxonomy `ProfileEditDialog` + `usePublishedTemplatesQuery`; documents: `documents.ts`, 2 adapters, `DocumentDetailRoute`, `DocumentEditorPage`, approval `DocumentApprovalExtras`, `SignoffDetailPage`). StepTemplate regression test keeps absent/null/published fixtures against the new single `published_version` gate.
- **QA gates that apply (feature subset):** contract gate, docs gate. Authz / multi-tenant / async / DB-invariant gates N/A (untouched).
- **Evidence shape:** `go build ./...`; targeted `go test` per touched package; `pnpm exec tsc --noEmit`; `pnpm exec vitest run src/features/{documents,templates,taxonomy,approval}`; `openapi-lint-local.ps1`; live drive of `GET /templates` + `GET /documents` + wizard Step 3 with proof.
- **Not run:** full integration suite (20+ min box constraint) — targeted `-run` filters only; bounded defer recorded in evidence.

## 9. Docs / ADR

- **Wiki:** update `wiki/modules/templates.md`, `wiki/modules/documents.md`, `wiki/architecture/api-contract.md` (+ `api-design-system.md` if it enumerates DTO conventions); refresh `Last verified` stamps. No new module doc.
- **REQ IDs:** contract-first REQs in `wiki/architecture/backend-target-architecture.md` (cite the contract/API-design REQ IDs in the PR); ADR 0013 (revision numbering — semantics unchanged, presentation reshaped); ADR 0035 (flat/envelope — this closes its optional-vs-null subclass structurally); ADR 0022 untouched.
- **ADR required?** **Yes — ADR 0065 "Version references are nested value objects in wire contracts"**: policy that coupled resource-pointer fields (id + counters + labels) are always one nested required ref object (nullable as a whole when the pointer may not exist), never parallel scalars; records the pre-v1 atomic-cutover exception to expand/contract; complements ADR 0035.

## 10. Verdict & locked constraints

- **Verdict:** 🟡 **Yellow** — clean fit, proceed; flagged: (a) ADR 0065 must land with or before the change, (b) breaking wire change justified only pre-v1 (window closes at v1 re-baseline), (c) documents-side migration is consistency-driven (its triple is required/non-nullable, so no correctness bug there — do templates first, documents second, but ship both or the pattern is a new inconsistency).
- **Open hard-stops:** none (AS-1/AS-2/AS-3 all clear).
- **Locked constraints handed to design/plan:**
  1. Contract-first: every shape change via openapi.yaml + regeneration; zero hand-edits to generated files.
  2. Two schemas, one pattern: `TemplateVersionRef` ≠ `DocumentRevisionRef`; no shared cross-context schema or Go package.
  3. `published_version` is required-and-nullable (present-and-null, never absent) — the 9f86828b guarantee carries forward; pin tests enforce.
  4. `TemplateVersionRef` includes `status` (unblocks wizard "why not selectable" UX — StepTemplate.tsx:80 TODO).
  5. Read model split: list projection gets its own struct(s) inside each module; `domain.Template` aggregate stops carrying join-projection fields.
  6. No DB migration; SQL projections unchanged.
  7. ADR 0065 written; ADR 0035 memory/doc annotated as structurally closed for this class.
  8. FE consumers gate on the single nullable object (`published_version == null`), never on inner fields.
