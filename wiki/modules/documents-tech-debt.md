# Tech Debt Register — documents

> Companion to `wiki/modules/documents.md`. Debt only — fixes belong in `wiki/backlog/documents-refactor.md`.

**Last verified:** 2026-05-18 (governed sidebar runtime + contract sync)

## Severity scale

See `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md` for the canonical trigger list. Highest matching tier wins.

## Items

### T-001 · RFC 9457 Problem Details envelope not adopted on documents routes — CLOSED 2026-05-12 (Plan 7)
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/documents/delivery/http/handler.go:974` (`mapErr`) · `:1027-1029` (`httpErr`) — `httpErr` now delegates to `problem.Write(w, problem.New(code, msg, msg))`. All call sites that used `httpErr` inherit `application/problem+json` responses. `mapErr` continues to map domain errors to `(int, string)` tuples consumed by `httpErr`.
- **Observation (original):** Handlers emitted legacy `{error:{code,message,details,trace_id}}` shape. Codegen bootstrap was in place but handler migration was deferred per ADR 0012.
- **Evidence:** `_artifacts/02-flow-finalizeDocument.md` §5; `_artifacts/05-industry.md` IP-001.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-001 (merged Plan 7 2026-05-11, commit `5b792150`)
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`; `wiki/architecture/api-design-system.md`

### T-002 · OpenAPI spec drift on `/api/v1/documents/*` routes
- **Severity:** critical
- **Surface:** `internal/modules/documents/delivery/http/handler.go:86, :111, :115, :116` (registrations) · `api/openapi/v1/openapi.yaml:3156, :3251`
- **Observation:** `renameDocument`, `duplicateDocument`, `comments` CRUD, and `archiveDocument` have handlers but no spec operations. `finalizeDocument` has a spec path (`openapi.yaml:3251`) with **no `operationId`** set. `listDocuments` handler exists; spec exposes `listDocumentsV2` (`openapi.yaml:3156`) — drift between handler name and spec id. Contract violation surface on regulated paths — clients have no typed binding for these ops.
- **Evidence:** `_artifacts/01-surface.md` (HTTP ops table); `_artifacts/02-flow-listDocuments.md` §1; `_artifacts/02-flow-renameDocument.md` §1; `wiki/backlog/contract-first-followups.md`.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-002
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`

### T-003 · `documents` table mutations lack tier-2 + tripwire defense-in-depth — CLOSED 2026-05-11 (Plan 5)
- **Severity:** major (closed)
- **Surface:** `internal/modules/documents/repository/repository.go` — `CreateDocumentTx`, `UpdateDocumentName`, `UpdateDocumentStatus`, `MarkArchived`, `Unarchive` — all 5 mutations now call `authz.Require` with `CapDocumentCreate` or `CapDocumentEdit` as appropriate. Migration `0188_tripwire_extend.sql:196-199` attaches `trg_require_cap_asserted` (`BEFORE INSERT OR UPDATE`) on `public.documents`.
- **Observation (original):** Approval-instance writes were layered (tier-1 role gate + tier-2 `authz.Require` + Postgres tripwire trigger). The `documents` table itself was single-layer: tier-1 role/owner check only; no `authz.Require` and no `enforce_capability_asserted` trigger attached. Defense-in-depth gap on a regulated mutation surface.
- **Evidence:** `_artifacts/04-persistence.md` (tripwire-pairing audit); `_artifacts/05-industry.md` IP-004.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-003
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

### T-004 · Duplicate route registration for `PATCH /api/v1/documents/{id}`
- **Severity:** minor
- **Surface:** `internal/modules/documents/delivery/http/handler.go:86` · `internal/modules/documents/delivery/http/handler.go:115`
- **Observation:** Both lines register `mux.HandleFunc("PATCH /api/v1/documents/{id}", h.renameDocument)`. stdlib mux is last-wins. Same handler reference → no behavioral change today, but a future edit could swap the second binding without anyone noticing the first.
- **Evidence:** `_artifacts/02-flow-renameDocument.md` §1.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-004
- **Linked ADR:** missing-ADR

### T-005 · `renameDocument` audit write happens outside the SQL UPDATE transaction — CLOSED 2026-05-11 (Plan 6a)
- **Severity:** major (closed)
- **Surface:** `internal/modules/documents/application/service.go:575` (UpdateDocumentName) · `internal/modules/documents/application/service.go:579` (s.audit.Write)
- **Observation:** `Service.RenameDocument` issues a plain `ExecContext` UPDATE and then calls `s.audit.Write` as an independent operation — there is no transaction boundary wrapping both. A crash between the two leaves the table mutated with no governance event. Defense-in-depth gap; audit is the only sink for QMS rename trail.
- **Evidence:** `_artifacts/02-flow-renameDocument.md` §2 ("Transaction boundary: NONE").
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-005
- **Linked ADR:** missing-ADR

### T-006 · `POST /api/v1/documents/{id}/finalize` HTTP idempotency contract — CLOSED 2026-05-18
- **Severity:** major (closed)
- **Surface (resolved):** `internal/modules/documents/delivery/http/handler.go:316` now enforces `Idempotency-Key`, records replay entries through `internal/platform/idempotency`, and returns `201 { instanceId }` with `Idempotent-Replay: true` on replay; `api/openapi/v1/openapi.yaml` and `frontend/apps/web/src/lib/api-types/index.d.ts` were aligned to the same response shape; `frontend/apps/web/src/features/documents/api/documents.ts` now sends the header.
- **Observation (original):** The runtime path already enforced HTTP idempotency, but the shared contract and frontend wrapper drifted: OpenAPI still documented a bare `200`, generated frontend types inherited that drift, and the handwritten wrapper did not send `Idempotency-Key`.
- **Evidence:** `internal/modules/documents/delivery/http/handler_test.go` (`MissingIdempotencyKey`, `InvalidIdempotencyKey`, `ReplayReturnsCreatedAndHeader`); `frontend/apps/web/src/features/documents/__tests__/documents.test.ts`; contract/codegen regeneration on 2026-05-18.
- **Linked backlog row:** `wiki/backlog/editor.md` integration audit item `Submit for review CTA` (closed prerequisite)
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

### T-009 · `document_placeholder_values.revision_id` FK targets `documents(id)` instead of `document_revisions(id)`
- **Severity:** major
- **Surface:** `migrations/0152_placeholder_fillin_columns.sql:51`
- **Observation:** Column `revision_id` is documented and used as a pointer into the revision lineage, but the migration declares `REFERENCES documents(id)`. Either the FK target is wrong, or the column semantics differ from the name. Latent data-integrity surface: inserts succeed only if `revision_id` happens to match a `documents.id` value (UUID collision near-zero, so the surface is effectively "FK never enforces anything useful").
- **Evidence:** `_artifacts/04-persistence.md` (table schema audit, document_placeholder_values row).
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-009
- **Linked ADR:** missing-ADR

### T-010 · Legacy `mux.HandleFunc` routing coexists with codegen bootstrap (`api.gen.go`)
- **Severity:** minor
- **Surface:** `internal/modules/documents/delivery/http/handler.go:82..116` (direct mux registrations) · `internal/modules/documents/api/api.gen.go:1` (generated, unused at runtime)
- **Observation:** `api.gen.go` is generated and committed but no route is mounted via the generated `ServerInterface`. The intent (ADR 0012) is to migrate; spec drift (T-002) blocks it. Carrying both shapes is fine short-term but rots if either side drifts further.
- **Evidence:** `_artifacts/01-surface.md`; ADR 0012 deferral note.
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
- **Surface (resolved):** `frontend/apps/web/src/features/documents/components/EditorMetaSidebar.tsx`, `frontend/apps/web/src/features/documents/pages/DocumentEditorPage.tsx`, `frontend/apps/web/src/features/documents/api/documents.ts`, `frontend/apps/web/src/features/registry/api/controlledDocuments.ts`
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
