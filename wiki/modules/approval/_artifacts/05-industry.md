# Phase 5 — Industry comparison (approval)

Patterns drawn ONLY from `.claude/skills/metaldocs-module-doc/references/industry-patterns-index.md`. No fresh web research this run.

## Applicable

### IP-004 · Defense-in-depth authz (NIST SP 800-95 §4.3)
- Source: NIST SP 800-95 (2007) — "Multiple layers of access control reduce single-point bypass risk."
- Approval is the canonical MetalDocs showcase: tier-1 capability gate at HTTP edge (verified upstream by `auth` middleware), tier-2 `authz.Require(ctx, tx, "doc.signoff", areaCode)` in-tx (`internal/modules/documents/approval/application/decision_service.go:141`) and `authz.Require("doc.submit", areaCode)` (`internal/modules/documents/approval/application/submit_service.go:85`), tier-3 Postgres tripwire `enforce_capability_asserted` `BEFORE INSERT ON public.approval_instances` + `public.approval_signoffs` (`migrations/0142b_role_capabilities_v2_enforce.sql:201,207`) reading GUC `metaldocs.asserted_caps` set by `authz.Require`.
- J1 eligibility adds a 4th layer specific to signoff: `enforce_signoff_eligibility_trg` `BEFORE INSERT ON approval_signoffs` (`migrations/0180_signoff_eligibility_trigger.sql:26`).
- Verdict: **applied — only module today with full 3-layer pairing on a regulated path.**

### IP-002 · Stripe-style idempotency
- Source: Stripe docs — "Keys are eligible to be removed from the system after they're at least 24 hours old."
- Approval signoff path: HTTP-layer `Idempotency-Key` via `PostgresSignoffIdempStore.CheckReplay`/`RecordReplay` (`internal/modules/documents/approval/infrastructure/postgres_signoff_idemp_store.go:25,42`) backed by `metaldocs.idempotency_keys` (24h `expires_at`, payload_hash, response_status+body — `migrations/0147_idempotency_keys.sql:6-17`).
- Repository layer adds field-compare replay: `INSERT … ON CONFLICT (approval_instance_id, actor_user_id) DO NOTHING` + `LoadSignoffByActor` (`postgres_approval_repository.go:127-173`) — replays return the prior `Signoff` row when stage/decision/content_hash match.
- Submit path derives a deterministic `idempotency_key` (`application/idempotency.go:25`) stored on `approval_instances` with UNIQUE `(document_v2_id, idempotency_key)`.
- Verdict: **applied — two-layer (HTTP store + DB ON CONFLICT) is stricter than baseline IP-002.**

### IP-006 · Forward-only migrations (Fowler 2016)
- Source: Fowler — "Each change to the database is described by a migration script."
- 18 migrations affect approval-owned tables, never edited post-merge: `0016` (init), `0134/0135` (schema), `0140` (revision/inbox), `0141` (governance dedupe), `0142a/0142b` (role caps + tripwire), `0144` (cancel state), `0145` (route immutability trigger), `0146` (active column), `0151` (seed), `0167` (documents bridge), `0173` (displayname snapshot), `0180` (eligibility trigger). `0142b_down.sql` exists as rollback companion (only such file among approval migrations).
- Verdict: **applied.**

### IP-008 · Row-level tenant id + scoped indexes (Crunchy Data)
- Source: Crunchy Data — "Add tenant_id to every multi-tenant table and index it first."
- Every owned table carries `tenant_id` first column except `approval_signoffs` which uses `actor_tenant_id` (`migrations/0135_approval_instances.sql:77`) — semantically the same, named distinctly.
- Inbox index `ix_approval_instances_inbox` leads with `tenant_id` (`migrations/0140_revision_version_and_inbox_index.sql:23-24`).
- Tenant scoping additionally GUC-enforced via `metaldocs.tenant_id` set in `authz_guc.go:12`.
- Verdict: **applied.**

### Transactional Outbox (informal — not in index)
- The PDF dispatch path uses an outbox enqueued in the same tx as instance approval: `pdfOutbox.Enqueue` at `decision_service.go:387` precedes `tx.Commit` at `:394`. Implementation `internal/modules/render/fanout/pdf_outbox_repository.go:25-35` writes `INSERT INTO metaldocs.pdf_dispatch_outbox … ON CONFLICT (tenant_id, revision_id) DO NOTHING`.
- Pattern not in index. Not citing as industry — described in module doc as a MetalDocs decision, sourced to `wiki/concepts/pdf-fanout.md` if/when that doc lands.

## Not applicable

### IP-001 · RFC 9457 Problem Details
- Approval HTTP handlers use **legacy** `contracts.ErrorResponse{Error: ErrorBody{Code, Message, Details, TraceID}}` envelope (`internal/modules/documents/approval/http/contracts/errors.go:3-12`), not RFC 9457.
- `looksLikeValidationError` (`http/errors.go:181`) does substring matching on `" must be "` to coerce 422s — fragile.
- Status: **drift from contract** — logged as Major debt (T-RFC9457-DRIFT) in tech-debt register.

### IP-003 · Cursor pagination
- Inbox uses `LIMIT/OFFSET + COUNT` two-query pattern (`read_service.go:153,224`), not cursor.
- Status: **drift from index recommendation** — logged as Minor debt (snapshot-skew between the two queries also flagged).

### IP-005 · OpenAPI codegen
- Approval HTTP routes wired via raw `*http.ServeMux` (`internal/modules/documents/approval/http/router.go:6`), not generated from spec. `signoffs` document-scoped path absent from `api/openapi/spec2.yaml` per Phase 2 trace.
- Submit path has partial OpenAPI surface via documents-module `finalizeDocument` entry but lacks operationId.
- Status: **drift from contract** — logged as Major debt (T-NO-CODEGEN) in tech-debt register.

### IP-007 · Observability
- Per index: "observability not yet wired in MetalDocs — flag as missing-ADR if a module assumes it." Approval does NOT assume it. Not applicable, not debt.

## New patterns considered but NOT added

None. All citations resolved against existing index rows.
