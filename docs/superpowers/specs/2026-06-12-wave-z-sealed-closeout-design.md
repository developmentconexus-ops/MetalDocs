# Wave Z — Sealed-Scope Backend Close-Out (Design Spec)

> **Date:** 2026-06-12
> **Status:** APPROVED (brainstorm complete, user sign-off in session #5)
> **Owners:** Leandro (final sign-off, v1 release act) · Claude (execution, fresh session)
> **Context:** The backend-professionalization program (Waves 0–2 + Wave F) is COMPLETE and verified — see `wiki/backend/roadmap.md` (frozen) and the Wave F closing report in `wiki/references/current-agent-handoff.md`. Wave Z is the FINAL backend wave: it executes everything that program deferred, so the backend never needs another standardization pass. After Wave Z, the only backend work left in this repository is feature work.
> **Driver:** v1 prototype release target Sunday 2026-06-14. Backend must be DONE — the operator moves to frontend/QA workstreams and performs the release re-baseline personally (runbook delivered by this wave).

---

## 1. Goal and the anti-circle rule

**Goal:** register-zero + ops-ready. Every finding in `wiki/backend/legacy-register.md` is fixed or closed-by-KEEP-verdict; the two consciously-deferred architecture items (OTel, full-table RLS) are EXECUTED (operator overrode D-1/D-3 scope limits on 2026-06-12); module tech-debt registers have zero `next-touch` rows owned by the backend; the ADR registry is trustworthy; `wiki/architecture/backend-blueprint.md` ends all-✅ (sole exception: one documented "at-release" line for the F-18 history re-baseline).

**The anti-circle rule (BINDING — this is the contract that ends the treadmill):**

1. Scope = the sealed manifest in §3. Nothing else. The manifest does not grow.
2. Findings discovered mid-wave are triaged by exactly one rule:
   - **Regression caused by Wave Z work** → fix in-wave (own commit, cites the manifest ID that caused it).
   - **Anything pre-existing** → appended to `wiki/backend/post-v1-backlog.md` (created in Phase 0) with a one-line description + pointer. It CANNOT block DONE, by definition.
3. The final review prompt is restricted to: *"did Wave Z break what Wave F verified?"* Reviews verify; they do not generate scope.
4. DONE is machine-checkable (§2). When the gate passes, the backend is closed. No reviewer, agent, or audit may reopen it; new ideas go to the post-v1 backlog.

## 2. Completion definition (the DONE gate, machine-checkable)

ALL of the following, evidenced with command + output per the CLAUDE.md evidence rule:

| # | Gate | Command / proof |
|---|------|-----------------|
| G1 | Build + vet clean | `go build ./...` exit 0 · `go vet ./...` exit 0 |
| G2 | Full test suite green | `go test -p 2 ./...` exit 0, zero FAIL |
| G3 | api-lint zero | `go run ./scripts/api-lint/ -strict api/openapi/v1/openapi.yaml .` → 0 violations |
| G4 | cilint zero, **zero allow-dualmode directives** | `go run ./tools/cilint/... ./...` exit 0 AND `git grep -c "cilint:allow-dualmode" -- internal/` = 0 |
| G5 | Manifest empty | every §3 row has a commit hash or a written KEEP/at-release rationale |
| G6 | Register zero | `legacy-register.md`: no entry without resolution-commit or KEEP/at-release closure note (defer-with-trigger no longer counts) |
| G7 | Blueprint all-green | `backend-blueprint.md` scoreboard: every concern ✅ (one documented at-release line for F-18) |
| G8 | ADR index trustworthy | every ADR has canonical `Status:` header; `wiki/decisions/index.md` rebuilt with status column; zero stubs/no-status files |
| G9 | Runtime smoke | login 200 · authed GET 200 · RLS NOSUPERUSER probe on ≥2 newly-covered tables · OTel inert-when-unconfigured boot + traceparent propagation proof · in-tx governance row · panic→500 · 429 |
| G10 | Regression-only review clean | review scoped per anti-circle rule #3 returns no Wave-Z-caused defects |
| G11 | Deliverables exist | `post-v1-backlog.md` · re-baseline runbook · updated handoff close-out block |

## 3. Sealed manifest

IDs are stable; the execution plan references them. Sizes: XS <30min · S <2h · M half-day (agent time).
Stage-2 verdicts and file anchors live in `wiki/backend/legacy-register.md` and `wiki/backend/stage2-evaluation.md` — executors read the finding's anchors first.

### 3.A Big two (operator override of D-1/D-3, runtime-verified)

| ID | Item | Source | Size | Verify |
|----|------|--------|------|--------|
| Z-1 | **Minimal OTel**: `otelhttp.NewHandler` wrap in the chain (position: with httpObs, order test updated) · W3C `traceparent` propagation (bridge existing `X-Trace-Id`: accept both inbound, emit traceparent outbound; requesttrace keeps working) · OTLP export via `go.opentelemetry.io/contrib/exporters/autoexport` (env-gated: NO env → no-op tracer, zero overhead, boots identical) · `normalizeRoute` replaced/extended via Go 1.22 `r.Pattern` · propagate trace context into outbox rows → worker (carry traceparent in `outbox_events.trace_id`, already a column). Consult context7/web for current otel-go API before writing. | F-17, RF-1, REQ-OBS-1/2/3 | M | env-unset boot byte-identical behavior (smoke) · env-set with console/OTLP exporter shows spans for one request · chain-order test green · traceparent header observed in response/log |
| Z-2 | **Full-table RLS**: migration `0237` extending the 0234 pattern (ENABLE+FORCE + NULL-permissive `tenant_isolation` policy, GUC `metaldocs.tenant_id`) to ALL remaining tenant-scoped tables. Census (2026-06-12, 28 tables = 30 tenant_id columns minus the 2 already covered): metaldocs.{audit_export_jobs, auth_sessions, document_process_areas, document_profiles, iam_group_members, iam_groups, iam_user_roles, iam_users, idempotency_keys, materialize_dispatch_outbox, pdf_dispatch_outbox, tenant_plans, user_process_areas} + public.{approval_instances, approval_routes, cd_sequence_counters, controlled_document_area_grants, controlled_document_user_grants, document_comments, document_placeholder_values, documents, editor_sessions, governance_events, template_audit_log, templates, templates_audit_log, templates_template, user_process_areas}. Re-run the census query in-session before writing the migration (tables may have moved). System paths that legitimately run GUC-less stay working via NULL-permissive policy (ADR 0027 altitude — tripwire, not boundary). ADR 0027 amended: sequencing section marked EXECUTED-IN-FULL by Z-2. | F-12 tail, RF-6, REQ-TEN-1 | M | migration applies + ledgered · full test suite green · NOSUPERUSER live probe on ≥2 new tables (iam_users + documents): GUC-unset→all, GUC=A→only A · worker+jobs still process (GUC-less system path proof) |
| Z-3 | `idempotency_keys.tenant_id` FK constraint (column exists; add FK + decide ON DELETE) — rides the Z-2 migration | F-09d | XS | migration applies; idempotency tests green |

### 3.B Dual-mode elimination (kills every `//cilint:allow-dualmode`)

| ID | Item | Source | Size | Verify |
|----|------|--------|------|--------|
| Z-4 | Audit `Service.writer` hard-required: refactor `WithExports` so the export feature gate lives at construction/wiring, writer never nil inside service; remove `service.go:183` allow-dualmode | T-012, F.6-C tail | S | grep 0 directives in audit; audit tests green |
| Z-5 | `freeze_service` in-tx-only collapse: remove the ADR-0015 optional-tx-enlistment dual paths (3 directives at freeze_service.go:175,308,370); both branches already write — keep the tx-required path only; amend ADR 0015 (status: amended, optional-enlistment retired) | T-013 | M | grep 0 directives; documents/freeze tests green; Pin/Materialize integration probe |
| Z-6 | IAM membership governance in-tx: `area_membership_service` best-effort post-commit → `RecordTx` in the mutation tx (closes T-007 fully; pattern = Wave 2.2) | T-007 | S | atomicity test (rollback drops event); iam tests green |

### 3.C Boundary + duplication tail

| ID | Item | Source | Size | Verify |
|----|------|--------|------|--------|
| Z-7 | Taxonomy `TemplateVersionChecker` → `TemplateVersionPort` in templates/domain, impl in templates/infra, taxonomy consumes port | F-06e | S | grep: no templates-table SQL under taxonomy; tests green |
| Z-8 | Security module structural JOIN to `iam_users` → narrow read port on iam/domain (same pattern as LoginContextPort) OR a written accepted-boundary note in the module doc if the JOIN is the tenancy mechanism itself (executor decides from anchors, documents choice) | F-06 residual | S | grep + module doc note; security tests green |
| Z-9 | Standalone `PostgresControlledDocumentRepository` constructed in main.go → expose via CD module Dependencies | F-06 residual | S | main.go constructs via module; build green |
| Z-10 | Outbox generics: `StagingOutboxWorker[R]`/`StagingOutboxRepository[R]` collapse pdf/materialize clones · delete dead restart loop in `startOutboxWorker` · dedup approval signoff/route-admin idemp-store ReplayHandle code | F-04, F-04-dead-loop | M | fanout+approval tests green; Wave-F worker relay probe re-run live |
| Z-11 | MinIO clients 3→2: `miniostore.NewStore` accepts `*minio.Client` from `buildMinioClients` | D-02 | S | build green; objectstore tests; live presign smoke |
| Z-12 | `parseBoolEnv` consolidation: exported `config.ParseBoolEnv` (4-value POSIX), both callers migrate; `splitCSV` stays (KEEP) | F-15 | XS | config+authn tests green |
| Z-13 | `DocumentStatus` enum completion: add 6 missing constants, convert string-literal callsites in documents module | D-07 | S | build+tests; grep: no raw status literals in documents/domain consumers |

### 3.D Contract tail

| ID | Item | Source | Size | Verify |
|----|------|--------|------|--------|
| Z-14 | CD handler typed problem codes (`INTERNAL_ERROR`/`IDEMPOTENCY_KEY_REQUIRED`/`VALIDATION_ERROR` literals → `problem.Code` consts) + enroll package in the F-09 catalog guard (widen `httpresponse.WriteError` to `problem.Code` if needed — follow the Wave 1.4 pattern) | F.6 finding D | S | guard test covers package; CD tests green |
| Z-15 | Idempotency TTL string → single const in postgres_store.go | F-09c | XS | idempotency tests green |
| Z-16 | Regen `CreateManagedUserRequestRoles` enum (add signer, area_admin, qms_admin — spec first, then `oapi-codegen` + FE regen per metaldocs-backend-api skill) | F-11 residual | S | codegen idempotent; api-lint 0; FE types compile |
| Z-17 | Partials disposition: confirm the 3 `api/openapi/v1/partials/*.yaml` are consumed by nothing (grep codegen configs + scripts) → delete; if consumed, reconcile casing to snake_case | F-13c | XS | grep proof; api-lint 0 |
| Z-18 | `deprecated: true` on `createManagedUser` (POST /iam/users) + regen | F-13d | XS | api-lint 0; codegen idempotent |
| Z-19 | iamapi codegen split: `iam/api/cfg.yaml` → three configs (iam, audit, security), three generated packages, imports updated | F-13e | M | build green; codegen idempotent; no behavior change (generated-only diff outside imports) |
| Z-20 | A3 promotion by verification: confirm api-contract-hardening backlog is CLOSED (it is — Phase F done 2026-06-08, re-audit 0 CRITICAL/0 HIGH, api-lint 0 blocking/0 reported); update blueprint A3 → ✅ citing the backlog's own closing evidence | A3, stale handoff note | XS | doc-only; blueprint row flipped with citation |

### 3.E Hygiene + ops

| ID | Item | Source | Size | Verify |
|----|------|--------|------|--------|
| Z-21 | `log.Printf` → `slog` sweep: 56 sites / 12 files (auth handler+middleware, approval doc+signoff handlers, documents handler, documents jobs ×2, iam dev_role_provider+admin+people+memberships handlers, templates lifecycle) + `platform/objectstore/document_presigner.go`. Preserve message content; add structured keys where trivially available (err, id) — no new context plumbing | F-02, REQ-OBS-1 | S | grep: 0 `log.Printf` non-test under internal/+apps/; touched-module tests green |
| Z-22 | WS presence drain: hub connection tracking + `server.RegisterOnShutdown` close signal | F-16B, RF-9 | M | unit test: shutdown closes tracked conns; presence tests green |
| Z-23 | Readiness probe concurrency: sequential 2s-each checks → errgroup with shared budget (`observability/runtime.go:286-323`) | F-16C, RF-9 | S | runtime tests green; /readyz live smoke |
| Z-24 | CD search trigram index: `pg_trgm` extension (verify available in image) + GIN index + rewrite leading-wildcard ILIKE (`controlleddocuments/infrastructure/repository.go:128` TODO(phase11)) | F-20b | S | migration applies; CD search tests green; EXPLAIN uses index |
| Z-25 | Drop orphan `metaldocs.documents.subject_code` column + index (migration; re-verify zero refs first — grep + information_schema) | CD T-010, 0236 residual | XS | migration applies; build+tests green; post-drop smoke |
| Z-26 | gitleaks full-history mode: `fetch-depth: 0`, drop `--no-git` · triage the ~15 historical findings (test fixtures/dummy tokens) into the singular `[allowlist]` (v8.24.3 ignores plural form — known) or pin a newer gitleaks supporting rule-level allowlists · workflow YAML parses · full-history run exit 0 | 1.11, F-18 round-2 | M | local full-history gitleaks run exit 0 with final config; YAML parses |

### 3.F Documentation + governance

| ID | Item | Source | Size | Verify |
|----|------|--------|------|--------|
| Z-27 | **ADR registry lifecycle audit** (operator's explicit ask — legacy ADRs mislead): every file in `wiki/decisions/` gets a canonical header `Status: Accepted | Accepted (amended YYYY-MM-DD) | Superseded by ADR-XXXX | Deprecated | Historical` · per-ADR content check vs current code — where the body describes dead reality, add a one-line **Current reality (2026-06):** note (do NOT rewrite bodies; ADRs are records) · 0003 stub → write the 10-line real record or mark Historical-stub-superseded · 0012/0013 get status lines · stray `2026-06-03-audit-events-cursor-shape.md` → renumber into sequence or fold into index as a dated decision record · rebuild `index.md` with columns # / Title / Status / Superseded-by / One-line current relevance · README updated with the status vocabulary | operator ask | M (parallel ×5 agents, 5 ADRs each) | every ADR greppable `^> \*\*Status:\*\*` with canonical value; index complete; zero NO-STATUS files |
| Z-28 | ADR 0022 Phase 6: the deliberately-last wiki sync — dispatch wiki-curator over the authz/capability surfaces; mark Phase 6 complete in ADR 0022; flip ADR status to `Accepted (fully executed)` | ADR 0022 | M | ADR 0022 phase table all-complete; authz wiki pages stamped |
| Z-29 | RF-3 authz-cache invalidation contract: write the `// CacheContract:` block on `CachedRoleProvider` (TTL, invalidation triggers, staleness bound, failure mode) + verify the invalidation paths actually exist for role grant/revoke (read the code; if a mutation path does NOT invalidate, that is a Wave-Z regression-class defect → fix: targeted invalidate call) + same contract doc for the Redis capability cache · covers D-05 | RF-3, D-05, REQ-CACHE-1 | S/M | doc block exists; invalidation path test or live probe (grant role → cache reflects within contract bound) |
| Z-30 | RF-8 feature-flag lifecycle doc (naming, ramp, cleanup-date rule) — one wiki page, referenced from blueprint D7 → ✅ | RF-8, D7 | XS | page exists; blueprint flipped |
| Z-31 | RF-7 messaging fence: `platform/messaging` noop/outbox stay; `servicebus` (gotenberg client) — document as the sync-render adapter it actually is (not a broker); blueprint D8 → ✅ with the fence note | RF-7, D8 | XS | doc note; blueprint flipped |
| Z-32 | **v1 release re-baseline runbook**: exact D-4b steps (init fresh repo, single v1 commit from current tree minus gitignored, what to verify before/after, old-repo archive note, secret-history closure statement, CI re-point) — operator executes Sunday | D-4b, F-18 | S | runbook file exists in wiki/runbooks/; dry-readable end-to-end |
| Z-33 | Final close-out: register final sweep (defer notes → resolution notes citing Z-* commits) · blueprint all-✅ re-score · `post-v1-backlog.md` finalized · roadmap gets a frozen Wave Z addendum table · handoff gets the Wave Z closing block | G5–G8, G11 | S | the DONE gate §2 passes end-to-end |

### 3.G Excluded forever (written rationale — NOT defers)

| Item | Why out |
|------|---------|
| F-09a finalize inline idempotency | Stage-2 verdict **KEEP**: structurally correct; middleware wrapping = over-engineering |
| F-20f sequential security queries | Stage-2 verdict **KEEP**: low-frequency admin surface; errgroup adds complexity for nothing |
| F-18 history residual | Closes physically at the Sunday re-baseline (Z-32 runbook); not executable before it |
| `splitCSV` duplication | Stage-2 verdict **KEEP** (Go proverb; 5 lines) |

## 4. Execution protocol

- **Fresh session.** Reads: `CLAUDE.md` → `wiki/references/current-agent-handoff.md` → THIS SPEC → the implementation plan (`docs/superpowers/plans/2026-06-12-wave-z-sealed-closeout.md`).
- **Branch:** `qa/iam-area-membership` (same integration branch; user merges after sign-off).
- **Workflows (ultracode), HARD LIMITS:** worker agents are **sonnet** for implement/review and **haiku** for mechanical verification — **NEVER fable** for execution agents. **Max 15 concurrent agents** at any moment; size `parallel()`/`pipeline()` batches accordingly. Prefer fewer, better-briefed agents over wide fan-out.
- **One commit per manifest ID** (tightly-related XS items may pair), message cites `Z-N` + finding ID + REQ IDs.
- **Phases:**
  - **P0 preflight:** baseline gates green (G1–G4) · Docker stack up · create `post-v1-backlog.md` · re-run RLS census.
  - **P1 parallel small code** (Z-4..Z-15, Z-17/18, Z-21, Z-23, Z-25): grouped by module ownership so no two agents touch one file; ≤15 concurrent.
  - **P2 big two** (Z-1 OTel, Z-2/Z-3 RLS): sequential, each with live runtime proof. Z-1 consults context7/web for current otel-go idiom first.
  - **P3 contract/codegen** (Z-16, Z-19, Z-24, Z-26): codegen items serialized (regen conflicts), metaldocs-backend-api skill governs.
  - **P4 structural M-items** (Z-5, Z-10, Z-22): one focused agent each.
  - **P5 docs/governance** (Z-27..Z-32): ADR audit fans out ×5; wiki-curator for Z-28.
  - **P6 DONE gate** (Z-33): run §2 G1–G11 in order, regression-only review (G10), close-out docs, evidence block.
- **Gates between phases:** G1–G4 rerun after every phase; a red gate stops the next phase.
- **GitNexus:** refresh the index at P0 (`node .gitnexus/run.cjs analyze`); use `impact` before structural edits (Z-5, Z-10, Z-19, Z-22); where the index lags, grep is truth.
- **Tooling:** web search + context7 available for library truth (otel-go, river, gitleaks versions) — consult before writing unfamiliar API code, never from memory.
- **Hard-stop rule stands:** any item that explodes into cross-module redesign beyond its row → STOP that item, record in post-v1-backlog with the boundary, continue the rest. (Expected for: none; the manifest was sized from Stage-2 verdicts.)
- **Environment:** `.gitnexus/` breaks `git add -A` → stage explicitly. C: SSD degraded writes → `go test -p 2`. API currently runs with `METALDOCS_E2E=1` — restart without it for final smoke.

## 5. What success means (one paragraph, for the fresh session)

Success is NOT "no one can find anything wrong" — that definition is unfalsifiable and caused a week of circles. Success IS: the §3 manifest has a commit or a written rationale on every row, the §2 gates G1–G11 all pass with fresh evidence, and everything else that exists or will be discovered lives in `post-v1-backlog.md` as post-v1 feature-era work. At that point the MetalDocs backend conforms to the documented target architecture (`backend-target-architecture.md` REQ-* set), is enforced by 7+ CI analyzers + chain test + full-history secret scan, is observable to OTel standard, is tenant-isolated at the database layer across every tenant-scoped table, has an honest ADR registry, and is frozen for feature work. The operator releases v1 Sunday with the Z-32 runbook.
