# Tech Debt Register — documents

> Companion to `wiki/modules/documents.md`. Debt only — fixes belong in `wiki/backlog/documents-refactor.md`.

**Last verified:** 2026-06-11 (Stage-1 adversarial-verification pass round 2: T-001 httpErr anchor corrected :1027-1029 → :1202-1204; T-002 OpenAPI line refs corrected :1952/:2050 → :2292/:2405; T-003 migration reference corrected to live path 0231:64-69; T-009 non-actionable status updated — FK bug already resolved in curated baseline db/baseline/0001_current_schema.sql:4333; T-010 evidence ranges tightened to exact call-site lines; prior: T-004 anchor corrected :86/:115 → :117/:147; T-005 surface anchors and observation corrected to match post-fix code; T-006 backlog sync discrepancy noted; prior: T-010 stale "not mounted" claim corrected — routes ARE mounted via documentsapi.HandlerWithOptions at module.go:120/134; prior: 2026-06-08 Phase F F8: handler.go finalizeDocument anchor :316 → :435)

## Severity scale

See `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md` for the canonical trigger list. Highest matching tier wins.

## Items

### T-001 · RFC 9457 Problem Details envelope not adopted on documents routes — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/documents/delivery/http/handler.go:974` (`mapErr`) · `:1202-1204` (`httpErr`) — `httpErr` now delegates to `problem.Write(w, problem.New(code, msg, msg))`. All call sites that used `httpErr` inherit `application/problem+json` responses. `mapErr` continues to map domain errors to `(int, string)` tuples consumed by `httpErr`.
- **Observation (original):** Handlers emitted legacy `{error:{code,message,details,trace_id}}` shape. Codegen bootstrap was in place but handler migration was deferred per ADR 0012.
- **Evidence:** `_artifacts/02-flow-finalizeDocument.md` §5; `_artifacts/05-industry.md` IP-001.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-001 (merged Plan 7 2026-05-11, commit `5b792150`)
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`; `wiki/architecture/api-design-system.md`

### T-002 · OpenAPI spec drift on `/api/v1/documents/*` routes
- **Severity:** critical
- **Surface:** `internal/modules/documents/delivery/http/handler.go:86, :111, :115, :116` (registrations) · `api/openapi/v1/openapi.yaml:2292, :2405`
- **Observation:** `renameDocument`, `duplicateDocument`, `comments` CRUD, and `archiveDocument` have handlers but no spec operations. `finalizeDocument` has a spec path at `openapi.yaml:2405` with `operationId: finalizeDocument` (now set). `listDocuments` handler exists; spec at `openapi.yaml:2292` (`operationId: listDocuments`). Contract violation surface on remaining unspec'd routes — clients have no typed binding for them.
- **Evidence:** `_artifacts/01-surface.md` (HTTP ops table); `_artifacts/02-flow-listDocuments.md` §1; `_artifacts/02-flow-renameDocument.md` §1; `wiki/backlog/contract-first-followups.md`.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-002
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`

### T-003 · `documents` table mutations lack tier-2 + tripwire defense-in-depth — CLOSED 2026-05-11 (Plan 5)
- **Severity:** major (closed)
- **Surface:** `internal/modules/documents/repository/repository.go` — `CreateDocumentTx`, `UpdateDocumentName`, `UpdateDocumentStatus`, `MarkArchived`, `Unarchive` — all 5 mutations now call `authz.Require` with `CapDocumentCreate` or `CapDocumentEdit` as appropriate. `trg_require_cap_asserted` (`BEFORE INSERT OR UPDATE`) is attached on `public.documents` via live migration `db/migrations/0231_db_hardening_tripwire_and_dead_schema.sql:64-69`; confirmed in curated baseline `db/baseline/0001_current_schema.sql:3790-3793`.
- **Observation (original):** Approval-instance writes were layered (tier-1 role gate + tier-2 `authz.Require` + Postgres tripwire trigger). The `documents` table itself was single-layer: tier-1 role/owner check only; no `authz.Require` and no `enforce_capability_asserted` trigger attached. Defense-in-depth gap on a regulated mutation surface.
- **Evidence:** `_artifacts/04-persistence.md` (tripwire-pairing audit); `_artifacts/05-industry.md` IP-004.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-003
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

### T-004 · Duplicate route registration for `PATCH /api/v1/documents/{id}`
- **Severity:** minor
- **Surface:** `internal/modules/documents/delivery/http/handler.go:117` (`RegisterRoutes`) · `internal/modules/documents/delivery/http/handler.go:147` (`RegisterRoutesWithRateLimit`)
- **Observation:** Both `RegisterRoutes` (line 117) and `RegisterRoutesWithRateLimit` (line 147) call `mux.HandleFunc("PATCH /api/v1/documents/{id}", h.renameDocument)`. The two functions are alternative entrypoints used at module wiring time — only one is called per startup path — so there is no live double-registration at runtime. However, the binding is duplicated in source: a future edit that changes the handler in one function without the other would diverge silently. `wiki/backlog/documents-refactor.md` R-004 references the stale anchor `:86`; the correct lines are `:117` and `:147`.
- **Evidence:** `_artifacts/02-flow-renameDocument.md` §1.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-004 (note: backlog row cites stale anchor `:86` — see correction above)
- **Linked ADR:** missing-ADR

### T-005 · `renameDocument` audit write happens outside the SQL UPDATE transaction — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/documents/application/service.go:575-586` — `RenameDocument` now uses `db.BeginTx` (line 575), calls `repo.UpdateDocumentNameTx` inside the transaction (line 580), calls `audit.WriteTx` inside the same transaction (line 583), and commits (line 586). Both mutations are atomic.
- **Observation (original):** `Service.RenameDocument` issued a plain `UpdateDocumentName` (`ExecContext`) and then called `s.audit.Write` as an independent operation outside any transaction boundary. A crash between the two left the table mutated with no governance event. Defense-in-depth gap; audit is the only QMS rename trail.
- **Evidence:** `_artifacts/02-flow-renameDocument.md` §2 (pre-fix flow, "Transaction boundary: NONE"); commit `0e106ed9`.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-005 (merged Plan 6a 2026-05-11)
- **Linked ADR:** missing-ADR

### T-006 · `POST /api/v1/documents/{id}/finalize` HTTP idempotency contract — CLOSED 2026-05-18
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/documents/delivery/http/handler.go:435` now enforces `Idempotency-Key`, records replay entries through `internal/platform/idempotency`, and returns `201 { instance_id }` with `Idempotent-Replay: true` on replay; `api/openapi/v1/openapi.yaml` and `frontend/apps/web/src/lib/api-types/index.d.ts` were aligned to the same response shape; `frontend/apps/web/src/features/documents/api/documents.ts` now sends the header.
- **Observation (original):** The runtime path already enforced HTTP idempotency, but the shared contract and frontend wrapper drifted: OpenAPI still documented a bare `200`, generated frontend types inherited that drift, and the handwritten wrapper did not send `Idempotency-Key`.
- **Evidence:** `internal/modules/documents/delivery/http/handler_test.go` (`MissingIdempotencyKey`, `InvalidIdempotencyKey`, `ReplayReturnsCreatedAndHeader`); `frontend/apps/web/src/features/documents/__tests__/documents.test.ts`; contract/codegen regeneration on 2026-05-18.
- **Linked backlog row:** `wiki/backlog/editor.md` integration audit item `Submit for review CTA` (closed prerequisite). **Note:** `wiki/backlog/documents-refactor.md` R-006 is listed as `open` — that is a stale backlog entry; code (`handler.go:440-448`) confirms idempotency enforcement is implemented and this item is closed.
- **Linked ADR:** `wiki/decisions/0011-cd-atomic-create.md` (sibling idempotency pattern); `wiki/decisions/0012-contract-first-api.md`

### T-007 · Audit emission relies on `Audit` interface with no `audit/domain` import in module graph
- **Severity:** minor
- **Surface:** `internal/modules/documents/application/service.go:81` (Audit interface) · `_artifacts/03-deps.md` (OUT list — no `audit/domain` edge)
- **Observation:** Documents emits audit events via a consumer-port interface; the concrete adapter is wired in `apps/api/cmd/metaldocs-api/main.go`. Module compiles without ever importing the audit package, which is intentional decoupling — but it leaves the audit contract un-codified at module level (no shared types, no shared error sentinels). Latent: any breaking change to audit's domain types lands at wiring time, not compile time of this module.
- **Evidence:** `_artifacts/03-deps.md` OUT-edge list.
- **Linked backlog row:** none yet — latent
- **Linked ADR:** missing-ADR

### T-008 · Capability namespace straddle: typed `iamdomain.Capability` vs. string `"doc.submit"` — CLOSED 2026-05-11 (Plan 4)
- **Severity:** minor (closed)
- **Surface:** `internal/modules/documents/application/fillin_authz.go:9` (typed import) · `internal/modules/documents/approval/application/submit_service.go:85` (`authz.Require(ctx, tx, "doc.submit", areaCode)`)
- **Observation:** Documents consumes both capability surfaces — typed `iamdomain.Capability` consts in fillin authz and string `"doc.*"` capabilities through the tripwire path. Cross-references `iam` T-001 (dual capability namespaces).
- **Evidence:** `_artifacts/00-context.md`; `wiki/modules/iam-tech-debt.md` T-001.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-008
- **Linked ADR:** missing-ADR

### T-009 · `document_placeholder_values.revision_id` FK targets `documents(id)` instead of `document_revisions(id)` — RESOLVED IN BASELINE (non-actionable)
- **Severity:** major (non-actionable — resolved in curated baseline)
- **Surface:** `archive/migrations/0152_placeholder_fillin_columns.sql:51` (original wrong declaration, archive-only); `archive/migrations/0191_document_placeholder_values_revision_fk.sql` (fix migration, archive-only)
- **Observation:** The original wrong FK (`REFERENCES documents(id)`) and the fix migration (0191) both exist only in `archive/migrations/`; neither is in `db/migrations/`. However, the curated baseline (`db/baseline/0001_current_schema.sql`) was built with the fix already incorporated: line 1926 creates the `document_placeholder_values` table, and line 4333 declares `FOREIGN KEY (revision_id) REFERENCES public.document_revisions(id) ON DELETE CASCADE` — the correct constraint. The table IS present in the live schema path via the baseline. **Item is non-actionable:** the FK bug was folded into the curated baseline and the correct constraint is in effect. No further migration is required.
- **Evidence:** `db/baseline/0001_current_schema.sql:1926` (table creation); `db/baseline/0001_current_schema.sql:4329-4333` (correct FK constraint); `archive/migrations/0191_document_placeholder_values_revision_fk.sql` (fix migration, applied at baseline-cut time).
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-009 (should be closed — baseline evidence above)
- **Linked ADR:** missing-ADR

### T-010 · Generated `ServerInterface` wrapper mounted but all typed path/query params discarded by the adapter (FLAG-03)
- **Severity:** minor
- **Surface:** `internal/modules/documents/module.go:120` and `:134` (two `documentsapi.HandlerWithOptions` call sites) · `internal/modules/documents/delivery/http/generated_adapter.go` (`NewGeneratedServerAdapter`) · `internal/modules/documents/api/api.gen.go:1`
- **Observation:** `api.gen.go` IS mounted at runtime — `RegisterRoutes` and `RegisterRoutesWithRateLimit` both call `documentsapi.HandlerWithOptions(dhttp.NewGeneratedServerAdapter(legacyMux), ...)` at `module.go:120` and `module.go:134`. The "not mounted" claim was stale. However, the legacy-delegating adapter (`NewGeneratedServerAdapter`) passes each request straight to the underlying legacy mux without consuming any of the typed parameters decoded by the generated `ServerInterfaceWrapper`, so the full value of codegen (typed params, request validation, binding) is not realized. This is the active FLAG-03 capture. The intent (ADR 0012) is to migrate handlers to consume generated types natively; spec drift (T-002) is the primary blocker.
- **Evidence:** `module.go:120` (`RegisterRoutes` call site); `module.go:134` (`RegisterRoutesWithRateLimit` call site); `_artifacts/01-surface.md`; ADR 0012; Stage-1 audit `wiki/backend/_artifacts/stage1/module-documents-core.md`.
- **Linked backlog row:** none yet — gated on T-002 closure
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`

### T-011 · Approval can complete without verified unresolved-comment server gate — CLOSED 2026-05-18
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/documents/approval/application/decision_service.go` now checks unresolved `document_comments` before final approval completion/freeze/document status transition; `internal/modules/documents/approval/http/errors.go` maps the guard to 409 `approval.unresolved_comments`; frontend signoff/editor flows consume the mapped conflict and persistent comment-load error state.
- **Observation (original):** Documents approval could complete without a verified server-side unresolved-comment gate. Review comments are active review feedback; release should be clean. Until the approval/signoff path checked unresolved comments, frontend screens could not claim enforcement beyond local UI hints.
- **Evidence:** `wiki/backlog/editor.md` (`Integration Audit (2026-05-17)` + `approval-blocking-unresolved-comments`), `decision_service_test.go`, `decision_service_freeze_test.go`, `errors_test.go`, `SignoffDialog.test.tsx`, `useDocumentComments.load.test.tsx`, `DocumentEditorPage.test.tsx`.
- **Linked backlog row:** `wiki/backlog/editor.md#approval-blocking-unresolved-comments` (closed by implementation sync 2026-05-18)
- **Linked ADR:** missing-ADR

---

### T-012 · Editor sidebar depended on mock governed revision / visibility / approver data — CLOSED 2026-05-18
- **Severity:** major (closed)
- **Surface (resolved):** `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`, `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`, `frontend/apps/web/src/features/documents/api/documents.ts`, `frontend/apps/web/src/features/controlled-documents/api/controlledDocuments.ts`
- **Observation (original):** The editor sidebar shipped mock metadata, revision history, and approver rows. That violated runtime truth and blurred the boundary between governed `documents` lineage and technical `document_revisions`.
- **Evidence:** `GET /api/v1/documents/{id}/revision-history`, `GET /api/v1/documents/{id}/approval-instance`, `GET /api/v1/controlled-documents/{id}`; focused frontend/backend tests added on 2026-05-18. Runtime debug on 2026-05-19 added migration `0205` so persisted `documents.revision_number` is zero-based and matches `REV00`/`REV01` labels directly, fixed approval-instance read transactions, and prevented readonly `under_review` editor loads from acquiring writer sessions.
- **Linked backlog row:** `wiki/backlog/editor.md`
- **Linked ADR:** missing-ADR

2026-05-19 follow-up: sidebar revision rows were tightened to code/title/date and `REV00` now defaults to `Criacao do documento`; no new debt opened. Rich history search/filtering remains deferred until governed lineages are long enough to need more than collapse/expand.

## Coverage stats (computed at compose time)

- Public symbols undocumented: ~480 / 517 (domain DTOs, repository internal helpers, generated stubs in `api/api.gen.go` not enumerated in module doc; the ~37 named symbols cover the public surface that crosses module/handler/service boundaries — see `_artifacts/01-surface.md` for the full list). Recorded as undocumented-on-purpose, not debt.
- Operations missing C4 placement: 0 / 22 (all routes in `_artifacts/01-surface.md` appear in §5.1 Container view or §5.3 HTTP table).
- Cross-deps missing in §5/§8: 0 / 25 (all 14 OUT + 11 IN edges from `_artifacts/03-deps.md` appear in §3.2 or §5).
- State transitions missing in §6: 0 / 5 (draft → under_review → approved → published → superseded/obsolete + rejected branch all tabled).
- Decisions without ADR link: 4 / 9 (T-004, T-005, T-007, T-009 flagged missing-ADR; T-008 closed by Plan 4).
