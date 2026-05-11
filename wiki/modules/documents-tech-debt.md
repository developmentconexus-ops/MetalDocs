# Tech Debt Register — documents

> Companion to `wiki/modules/documents.md`. Debt only — fixes belong in `wiki/backlog/documents-refactor.md`.

**Last verified:** 2026-05-11

## Severity scale

See `.claude/skills/metaldocs-module-doc/templates/tech-debt-register.md` for the canonical trigger list. Highest matching tier wins.

## Items

### T-001 · RFC 9457 Problem Details envelope not adopted on documents routes
- **Severity:** major
- **Surface:** `internal/modules/documents/delivery/http/handler.go:958` (mapErr) · `internal/modules/documents/delivery/http/handler.go:1013` (httpErr)
- **Observation:** Handlers emit legacy `{error:{code,message,details,trace_id}}` shape via `httpErr` + `mapErr`. Codegen bootstrap is in place (`internal/modules/documents/api/api.gen.go`) but handler migration is deferred per ADR 0012. Documented contract not followed yet — measurable impact for tooling that depends on Problem+JSON.
- **Evidence:** `_artifacts/02-flow-finalizeDocument.md` §5 (error response anchors); `_artifacts/05-industry.md` IP-001.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-001
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`

### T-002 · OpenAPI spec drift on `/api/v2/documents/*` routes
- **Severity:** critical
- **Surface:** `internal/modules/documents/delivery/http/handler.go:86, :111, :115, :116` (registrations) · `api/openapi/v1/openapi.yaml:3156, :3251`
- **Observation:** `renameDocument`, `duplicateDocument`, `comments` CRUD, and `archiveDocument` have handlers but no spec operations. `finalizeDocument` has a spec path (`openapi.yaml:3251`) with **no `operationId`** set. `listDocuments` handler exists; spec exposes `listDocumentsV2` (`openapi.yaml:3156`) — drift between handler name and spec id. Contract violation surface on regulated paths — clients have no typed binding for these ops.
- **Evidence:** `_artifacts/01-surface.md` (HTTP ops table); `_artifacts/02-flow-listDocuments.md` §1; `_artifacts/02-flow-renameDocument.md` §1; `wiki/backlog/contract-first-followups.md`.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-002
- **Linked ADR:** `wiki/decisions/0012-contract-first-api.md`

### T-003 · `documents` table mutations lack tier-2 + tripwire defense-in-depth
- **Severity:** major
- **Surface:** `internal/modules/documents/repository/repository.go:73` (CreateDocumentTx) · `:216` (UpdateDocumentName) · `:428` (UpdateDocumentStatus) · `:1071` (MarkArchived) · `:1082` (Unarchive)
- **Observation:** Approval-instance writes are layered (tier-1 role gate + tier-2 `authz.Require` + Postgres tripwire trigger `trg_require_cap_asserted_instances` at `migrations/0142b_role_capabilities_v2_enforce.sql:201`). The `documents` table itself is single-layer: tier-1 role/owner check only; no `authz.Require` and no `enforce_capability_asserted` trigger attached. Defense-in-depth gap on a regulated mutation surface.
- **Evidence:** `_artifacts/04-persistence.md` (tripwire-pairing audit); `_artifacts/05-industry.md` IP-004.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-003
- **Linked ADR:** `wiki/decisions/0007-two-tier-authz.md`

### T-004 · Duplicate route registration for `PATCH /api/v2/documents/{id}`
- **Severity:** minor
- **Surface:** `internal/modules/documents/delivery/http/handler.go:86` · `internal/modules/documents/delivery/http/handler.go:115`
- **Observation:** Both lines register `mux.HandleFunc("PATCH /api/v2/documents/{id}", h.renameDocument)`. stdlib mux is last-wins. Same handler reference → no behavioral change today, but a future edit could swap the second binding without anyone noticing the first.
- **Evidence:** `_artifacts/02-flow-renameDocument.md` §1.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-004
- **Linked ADR:** missing-ADR

### T-005 · `renameDocument` audit write happens outside the SQL UPDATE transaction
- **Severity:** major
- **Surface:** `internal/modules/documents/application/service.go:575` (UpdateDocumentName) · `internal/modules/documents/application/service.go:579` (s.audit.Write)
- **Observation:** `Service.RenameDocument` issues a plain `ExecContext` UPDATE and then calls `s.audit.Write` as an independent operation — there is no transaction boundary wrapping both. A crash between the two leaves the table mutated with no governance event. Defense-in-depth gap; audit is the only sink for QMS rename trail.
- **Evidence:** `_artifacts/02-flow-renameDocument.md` §2 ("Transaction boundary: NONE").
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-005
- **Linked ADR:** missing-ADR

### T-006 · `POST /api/v2/documents/{id}/finalize` lacks HTTP `Idempotency-Key` integration
- **Severity:** major
- **Surface:** `internal/modules/documents/delivery/http/handler.go:316` (handler entry) · `internal/modules/documents/approval/application/submit_service.go:61` (internal key compute) · `internal/modules/documents/approval/application/idempotency.go:20` (ComputeIdempotencyKey)
- **Observation:** Submit path computes a deterministic internal key but does not read the `Idempotency-Key` header nor write `metaldocs.idempotency_keys`. Replay safety relies entirely on the partial unique index `ux_approval_instances_active` (`migrations/0135_*.sql:33`) — a retry returns 409, not a replayed 201. Contract drift vs. Stripe-style idempotency (`internal/platform/idempotency/`).
- **Evidence:** `_artifacts/02-flow-finalizeDocument.md` §6 ("Idempotency: no"); `_artifacts/05-industry.md` IP-002.
- **Linked backlog row:** `wiki/backlog/documents-refactor.md` R-006
- **Linked ADR:** `wiki/decisions/0011-cd-atomic-create.md` (sibling pattern on CD create)

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

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: ~480 / 517 (domain DTOs, repository internal helpers, generated stubs in `api/api.gen.go` not enumerated in module doc; the ~37 named symbols cover the public surface that crosses module/handler/service boundaries — see `_artifacts/01-surface.md` for the full list). Recorded as undocumented-on-purpose, not debt.
- Operations missing C4 placement: 0 / 22 (all routes in `_artifacts/01-surface.md` appear in §5.1 Container view or §5.3 HTTP table).
- Cross-deps missing in §5/§8: 0 / 25 (all 14 OUT + 11 IN edges from `_artifacts/03-deps.md` appear in §3.2 or §5).
- State transitions missing in §6: 0 / 5 (draft → under_review → approved → published → superseded/obsolete + rejected branch all tabled).
- Decisions without ADR link: 4 / 9 (T-004, T-005, T-007, T-009 flagged missing-ADR; T-008 closed by Plan 4).
