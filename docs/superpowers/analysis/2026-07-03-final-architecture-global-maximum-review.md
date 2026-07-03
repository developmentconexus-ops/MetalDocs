# Final architecture review — Global Maximum alignment

**Date:** 2026-07-03
**Scope:** Whole-system adversarial review of MetalDocs architecture, structure, ADRs, and contracts against industry standards (Google AIP, Zalando, AWS SaaS Lens, Crunchy Data, OWASP, ISO 9001 §7.5.3 / 21 CFR Part 11, River/Temporal/microservices.io, bulletproof-react, MADR/adr.github.io, Google Testing Blog).
**Method:** 10 dimensions, one adversarial sonnet reviewer each (code evidence + external comparison with cited sources), synthesized by the main session. Verdict vocabulary per finding: **CONFIRMED** (global maximum), **DEBT** (design right, adoption/enforcement incomplete), **RE-LITIGATE** (design questionable; alternative named with trade-off).
**Question answered:** the system was built incrementally with big mid-course changes — is what stands today the global maximum, or a set of local maxima locked in by history?

---

## Executive scorecard

| # | Dimension | Verdict | One-line |
|---|-----------|---------|----------|
| 1 | Module structure & boundaries | **CONFIRMED** | Seams proven real by ADR 0039's 20-violation remediation + two CI guards; 2 debts, 1 re-litigate |
| 2 | Authorization (ADR 0022) | **CONFIRMED** | Bespoke capability×area PDP right-sized vs Cedar/OPA/SpiceDB; one open residual (tripwire hand-sync) |
| 3 | Contract-first API | **DEBT** | Stack correct (Zalando-aligned); governance tooling missing (oasdiff, shape lint, CI-wired sync check) |
| 4 | Multi-tenancy | **DEBT** | Sync path is textbook pooled+RLS; async fleet has ZERO RLS backstop (NULL-permissive policy) |
| 5 | Async architecture | **DEBT** | Correct semantics; three parallel job infrastructures where River could be one |
| 6 | Versioning kernel | **DEBT** | Core mechanics CONFIRMED; 9-status lifecycle is not actually one state machine; 2 ISO product gaps |
| 7 | DB as invariant enforcer | **CONFIRMED**/DEBT | Trigger bet vindicated for regulated data; tripwire arm hand-sync is the standing trap |
| 8 | Testing & QA | **CONFIRMED** | testdb/IntegreSQL architecture is global max; traceability and contract fuzzing are debts |
| 9 | Observability & ops | Yellow | Right-sized pre-v1; three named blockers before first paying customer |
| 10 | Decision governance | **CONFIRMED**/DEBT | MADR+stamps+curator loop demonstrably fires; ADR 0022 is a mega-ADR anti-pattern |

**Bottom line:** no dimension is Red. The load-bearing architecture decisions — modular monolith with enforced seams, capability authz, contract-first, pooled tenancy with RLS, transactional outbox, snapshot-by-value kernel, DB-enforced invariants, integration-heavy testing — are all independently CONFIRMED against industry evidence. The system's one systemic weakness is not a design choice but a **meta-pattern: hand-synced enumerations and unenforced conventions** (see Cross-cutting finding). The true re-litigation list is short and bounded (5 items, §Priorities).

---

## Cross-cutting finding: the one repeated defect class

Seven dimensions independently surfaced the same shape of defect — **a rule that is true by convention and kept true by human discipline, with no generator or CI gate**:

1. Tripwire trigger cap-arms hand-synced to the Go capability registry (2 shipped 500s: 0269, 0270; pinned reactively, not structurally).
2. `authz.SeedTxIdentity` called manually at ~85 application-service sites; forgetting one is silently absorbed by the NULL-permissive RLS policy.
3. `check-module-contract-sync.ps1` finds live DRIFT but is not CI-wired; no oasdiff breaking-change gate; redocly `struct` rule off.
4. Nullable-not-required spec fields (the 9f86828b bug class) have no schema-shape lint — the exact class AIP-134 field-behavior annotations exist to prevent.
5. FE hard rule #8 (no cross-feature imports) unenforced — no ESLint boundary config; violations widespread.
6. REQ-ID traceability: 3 citations in last 100 commits vs 16 ADR citations — convention promised, not practiced.
7. Wiki file:line anchors and Last-verified stamps checked by narration/curator, no CI anchor checker.

**Global maximum for this class is uniform: single source of truth + generation or a blocking gate.** The repo already proves it knows how — `TestCapabilityRegistrySize`, the 5 authz CI lints, `check-module-boundaries.ps1`, `check-test-discipline.sh`, and the e2e-coverage-gate all made drift structurally impossible in their domains, and those domains have stopped producing incidents. The seven items above are the domains that haven't received the same treatment yet.

---

## Dimension detail

### 1. Module structure & boundaries — CONFIRMED
- 14 modules real; **CLAUDE.md inventory stale**: lists a `docs` module that does not exist; `tokens` exists and is unlisted. Fix doc.
- Seams are real, not ritual: ADR 0039 census found and remediated ~20 cross-module base-table reads into published views + read-ports; enforced since by TWO independent guards (`check-module-boundaries.ps1` Go-import layer guard; `hgcrossmodule` cilint SQL guard). Above industry norm (most modular monoliths rely on review discipline — Fowler, Spring Modulith).
- 14 modules for one operator: defensible on domain grounds; the enforcement machinery has demonstrably earned its cost (caught real defects), so not ceremony.
- DEBT: layer naming split — `documents`/`templates` use `repository/`, other 12 use `infrastructure/`. Mechanical rename.
- RE-LITIGATE: `documents/approval` is structurally a 15th module (full domain/application/http/repository tree) hiding inside `documents/` — the boundary CI guard under-scopes it. Promote to `internal/modules/approval` or ADR the exception.
- Top risks: stale CLAUDE.md inventory; layer-name ambiguity; approval sub-module invisibility to guards.

### 2. Authorization — CONFIRMED (strongest dimension)
- Two-tier PDP + tripwire, 29-capability registry pinned by `TestCapabilityRegistrySize` (internal/modules/iam/domain/model_test.go:95), 5 CI lints bind all truth sources (no-inline-capability, no-rawstring-capability, authz-area-scope-binding, seed-registry-parity, wiki-capability-parity).
- **"Dual capability sources" suspicion overturned**: tier-1 (iam_user_roles⋈role_capabilities, coarse route gate) vs tier-2 (role_capabilities⋈user_process_areas, fine in-tx) is a deliberate coarse/fine split per OWASP microservices pattern; tier-2 strictly narrower; system_admin inheritance proven consistent. Collapsing tier-1 into a cache of tier-2 would add a DB round-trip per route for no correctness gain. The historical dev-seed confusion was tripwire hand-sync lag (dimension 7), misdiagnosed as dual-truth.
- Bespoke PDP vs Cedar/OPA/SpiceDB: ADR 0022's own amendment ran this comparison with revisit triggers (ReBAC hierarchy → SpiceDB/OpenFGA; runtime policy → Cedar). At 29 capabilities, no relationship graphs, and in-tx bypass auditing (`BypassAuditSink` writes to audit_events), bespoke is correct — a policy engine would have to reimplement the audit story. CONFIRMED global-maximum-for-scale.
- Open residuals: (a) tripwire arms have no generative guard (see dim 7); (b) hand-rolled pre-codegen IAM handlers outside the authz-call-present lint (documented exception); (c) two live tier-1↔tier-2 cap-name divergences reported in ADR 0022 Phase 7 and left open (forceReleaseDocumentSession: tier-1 `membership.manage` vs tier-2 `document.edit`; approval-route-management tier-1 `document.submit`).

### 3. Contract-first API — DEBT (stack right, governance incomplete)
- Single 6,546-line openapi.yaml + oapi-codegen strict-server + openapi-typescript: matches Zalando's recommended shape (single self-contained file); TypeSpec/Fern/buf would be a rewrite for marginal gain at this scale. Stack CONFIRMED.
- Governance gaps (all RE-LITIGATE-the-tooling): no oasdiff-class breaking-change CI gate; redocly disables `struct`/`security-defined`/`operation-summary` outright + 133 suppressed errors with no burn-down ticket; no AIP-134-style "nullable ⇒ required" shape lint (the 9f86828b bug class will recur without it); `check-module-contract-sync.ps1` advisory-only, currently reporting 4 live DRIFT items for templates.
- FE hand-written type overrides (templates.ts Omit<> re-typing + 9 hand-rolled types) are the drift vector that let the wizard bug hide. ADR 0065 (VersionRef) referenced in plans but not yet written — unlanded plan, not a completed decision.
- "Spec is truth" enforced for codegen sync (api-contract.yml CI) but NOT for route-presence/handler-binding truth.

### 4. Multi-tenancy — DEBT (one sharp asymmetry)
- Sync path CONFIRMED: pooled + per-tx GUC + hand predicates primary + FORCE RLS on all 29 tenant tables (0237) as backstop — the exact pattern AWS SaaS Lens / Crunchy Data recommend for pooled Postgres. Blob keys tenant-namespaced (`tenants/{tenantID}/...`, documents/application/keys.go:12,19). Cross-tenant 404 invariant integration-tested (tenant_isolation_test.go, 6 tests).
- **Sharpest finding of the review:** the RLS policy is deliberately NULL-permissive (`GUC unset → rows visible`, migration 0237 comment) to let GUC-less system paths work — which means **worker/jobs binaries run with zero RLS backstop**. Async isolation rests entirely on 229 hand-written `tenant_id` predicates. One bad join in a future worker query = silent cross-tenant leak, no gate. This asymmetry lives only in a migration comment, not in ADR 0027 or the wiki.
- RE-LITIGATE (bounded, cheap): (a) auto-seed tenant GUC at the TxRunner chokepoint (or a tenant-scoped TxRunner variant) instead of 85 manual `SeedTxIdentity` sites — Crunchy's guidance names this exact failure mode; (b) seed GUCs in worker/jobs per message so RLS backstop covers async too, or ADR-amend the accepted gap with compensating lint.
- Kernel gap (DEBT, tracked not deferred silently): no tenant onboarding API, offboarding, export, or erasure path; audit-immutability vs GDPR-erasure tension unresolved (standard answer: crypto-shredding of PII fields, immutable skeleton). Wiki already self-flags (T-003, backend-canon.md:181).

### 5. Async architecture — DEBT (semantics right, three engines)
- Semantics CONFIRMED: transactional enqueue everywhere (`river.Client.InsertTx` on the business tx — outbox pattern by construction); notification consumer idempotent (`ON CONFLICT (recipient_user_id, source_event_id) DO NOTHING`); polling over LISTEN/NOTIFY acceptable for seconds-scale QMS latencies; failed-status-as-DLQ satisfies REQ-ASYNC-3 inspectability.
- RE-LITIGATE (the review's biggest consolidation candidate): **three parallel job infrastructures** — River v0.37.1 (scheduled-publish + notifications fanout, jobs binary), custom Postgres-lease ticker scheduler (4 janitors, api binary, `metaldocs.acquire_lease`/job_leases), and the generic StagingOutboxWorker poller (pdf/materialize dispatch, own backoff math duplicated in two branches). River natively ships leader election, periodic jobs, transactional enqueue, and maintenance services — it can subsume the other two. The stuck-instance-watchdog's belt-and-suspenders (lease + advisory lock) and the H-PRE-1 advisory-lock deadlock constraint are both symptoms of not having one trusted primitive. Trade-off: migration cost + re-verifying H-PRE-1 under River semantics vs deleting ~3 duplicated retry/idempotency/election code paths permanently.
- DEBT: outbox tables never purged (`MarkDispatched` only flips status — unbounded growth, SKIP LOCKED scan cost rises); no ordering guarantee on lifecycle fanout (published/superseded race is plausible, unverified).

### 6. Versioning kernel — DEBT (mechanics confirmed, lifecycle fragmented, 2 product gaps)
- CONFIRMED: atomic CD create (ADR 0011, single tx, gap-free allocation — textbook); TemplateSnapshot full byte-copy + SHA-256 hashes with deliberately NO rebase path (correct for regulated docs; new revision is an explicit act per ADR 0052 — matches ISO change-control); ADR 0030 port closed a live correctness bug (hardcoded `status := "published"`); version_number/revision_number duality justified (lifecycle counter vs regulatory label).
- **RE-LITIGATE #1: the 9-status lifecycle is not one state machine.** `CanTransitionDocument` (documents/domain/state.go:7) covers 3 of 9 statuses; approve/publish/supersede/obsolete/schedule transitions live as scattered `if status != X` guards across 4 approval service files. Templates already has the right pattern (`TemplateVersion.CanTransition`). Fix: one exhaustive transition function + a test that the table covers all statuses.
- **RE-LITIGATE #2: scheduled-publish vs manual-publish race unverified.** Scheduler takes FOR UPDATE + `status == scheduled`; manual path checks `status == approved` independently; no shared choke point demonstrated. Needs a concrete two-goroutine race test, or route both through one PublishRevision method.
- DEBT (product, not code): vs Veeva/MasterControl/ISO 9001 §7.5.3 — no effective-date distinct from publish-date, **no periodic review/expiry**, **no structured reason-for-change** (free-text revision_title only). These two are core eQMS expectations, not YAGNI. Training acknowledgment legitimately out-of-module (distribution covers obligated readers).
- DEBT (minor): two concurrency transport idioms for one mechanism (documents If-Match `"vN"` vs templates body `expected_lock_version`).

### 7. DB as invariant enforcer — CONFIRMED bet, one standing trap
- The trigger bet is vindicated for this product class: ~19+ triggers all map to named regulated invariants (SoD signoff guard, immutability guards, tenant-consistency, capability tripwires) — exactly the case where external guidance (and FDA tamper-evidence expectations) favors DB enforcement. No generic business logic in triggers.
- Audit trail CONFIRMED: grant-level append-only (INSERT/SELECT only + explicit REVOKE in 0266), real hash chain (`audit_event_row_hash` over prev_hash+row fields, shape CHECKs), janitor recomputes chain via LAG. DEBT: validation window bounded to last 10,000 rows — tamper beyond the tail is invisible until retention/partitioning (T-013) lands.
- SequenceAllocator CONFIRMED: counter-row FOR UPDATE + UPDATE...RETURNING in one tx, nil-tx hard-rejected — canonical gap-free pattern.
- **RE-LITIGATE (the review's #1 structural fix): tripwire cap-arms must be generated, not hand-synced.** `enforce_capability_asserted()` hardcodes per-table TEXT[] cap lists; two production 500s (0269 template.review, 0270 template.archive) shipped because Go `authz.Require` call sites and trigger arms are linked by nothing. Current remediation (tripwire_caps_test.go) pins known incidents only — a new capability reproduces the class silently. Fix: derive arms (or a `cap_table_requirements` lookup table the trigger reads) from the same registry Go reads, CI-fail on drift — same AST approach as the existing authz-area-scope-binding lint. Third incident is a when, not an if, until then.
- DEBT: forward-only migrations (no down) — stated convention, acceptable, but corrective migrations under pressure are the cost.

### 8. Testing & QA — CONFIRMED
- Architecture is global max for a DB-invariant-centric system: 1,099 test files (260 integration-tagged), testdb factory (ADR 0034 — real root-cause narrative: GUC leakage/search_path), IntegreSQL template-DB-per-test cloning (better than testcontainers norms for Postgres GUC semantics), `check-test-discipline.sh` as blocking CI gate with shrink-only allowlist, 15 Playwright E2E specs WITH a coverage-map gate (e2e-coverage-gate.yml). Prior belief "flows proven only by manual curl drives" was **wrong** — E2E automation exists and is gated.
- RE-LITIGATE (small): only 52/260 integration files use t.Parallel — IntegreSQL cloning makes full parallelization safe by construction; CI config tweak, cuts the 20-min wall clock.
- DEBT: "delete legacy tests" policy is judgment-call, not written taxonomy — codify ("guards a REQ ID/tripwire/contract shape = repair; else delete"); no schemathesis-class generated contract fuzzing; **no coverage gate and no REQ-ID→test traceability automation** — for a regulated QMS, traceability being convention-only is the gap an auditor would find first. The e2e-coverage-gate grep pattern is the right template to replicate for backend REQ IDs.

### 9. Observability & ops — Yellow (right-sized pre-v1; 3 blockers before first customer)
- CONFIRMED: middleware chain pinned by chain_test.go (REQ-MW-7) — actual order includes origin_protection and TWO rate-limit layers (pre-auth login limiter before authn; general envelope limiter after authz), which is more precise than the CLAUDE.md doctrine line; panic recovery re-panics ErrAbortHandler and writes problem+json; graceful shutdown with WS drain + 15s budget; boot-time fail-fast config; OTel wired but env-gated inert (sane); slog JSON everywhere.
- **Doctrine corrections found:** (a) idempotency is NOT a chain link — it's per-handler `idempotency.Require(...)` on mutating routes (arguably the right design; CLAUDE.md wording overstates chain uniformity); (b) **T-001/T-005 legacy error envelopes are CLOSED since 2026-05-12** — no `{error:{code}}` shape remains in documents or templates; RFC 9457 holds system-wide. Session memory was stale on both.
- Idempotency store CONFIRMED Stripe-equivalent: two-phase BeginReplay/CompleteReplay, composite PK (tenant, actor, route, key), SHA-256 payload-conflict → 422, 24h TTL + janitor.
- Blockers before first paying customer (fine today): in-memory per-instance rate limiter (horizontal scale silently multiplies limits — redis_rate class fix); **no Dockerfiles for api/worker/jobs** (compose references deploy/docker/*.Dockerfile paths that don't exist); no Prometheus /metrics (custom JSON /api/v1/metrics only — wiki self-flags); log↔trace correlation unverified; no backup/restore doc (only a pg_dump line in a feature runbook); DEPLOY.md assumes K8s while compose assumes Docker — inconsistent target.

### 10. Decision governance — CONFIRMED loop, one anti-pattern
- The loop demonstrably fires: 63 ADRs, 100% status-stamped, MADR-consistent; templates.md drift self-healed by stamp+curator before this audit reached it; D-4a secret rule born from a real incident; developing-new-work gate operational. Not theater — right-sized for one operator + agents (Rust/HashiCorp RFC ceremony would be over-engineering here; Diátaxis formalization low-priority).
- **RE-LITIGATE: ADR 0022 is a mega-ADR** — multi-hundred-word phase changelog crammed into the status field, the exact anti-pattern adr.github.io/ozimmer.ch warn on. Rule to adopt: status ≤3 lines; execution history goes to a linked tracker doc. ADR 0013 never got a formal Superseded stamp despite 0052/0053 moving past it — supersession isn't mechanically closed.
- DEBT: 44% of wiki files carry no Last-verified stamp; ~19% of stamped ones stale >30d (mostly acceptable historical carve-outs); no CI anchor checker (file:line rot caught by narration); REQ-ID citation aspirational (see cross-cutting); C4 diagrams fragmentary; missing docs: threat model, SLO/capacity targets (name as backlog, acceptable pre-v1).

---

## Priorities (ranked by risk × cost)

**P1 — structural drift-proofing (the cross-cutting class; all bounded, high leverage):**
1. Generate tripwire cap-arms from the capability registry + CI drift check (dim 7 — two shipped incidents already).
2. Auto-seed tenant/actor GUCs at the TxRunner chokepoint; seed in worker/jobs too so RLS backstops async (dim 4).
3. Wire `check-module-contract-sync.ps1` + oasdiff into CI; add nullable⇒required shape lint; re-enable redocly `struct` with a burn-down (dim 3).
4. ESLint feature-boundary rule with explicit allowlist (dim FE).

**P2 — kernel correctness:**
5. Unify the 9-status document state machine into one exhaustive transition table + coverage test (dim 6).
6. Concurrent race test: scheduled vs manual publish; single PublishRevision choke point if it fails (dim 6).

**P3 — architecture consolidation (bigger, pre-v1 window):**
7. River consolidation study: janitors + staging outbox onto River periodic/transactional jobs; delete custom lease scheduler + poller; retires H-PRE-1 class (dim 5). Requires developing-new-work gate + ADR.
8. Promote `documents/approval` to a top-level module or ADR the nested exception (dim 1).

**P4 — product gaps (roadmap, not code debt):**
9. Periodic review/expiry + structured reason-for-change (ISO 9001 §7.5.3 core eQMS features) (dim 6).
10. Tenant lifecycle: onboarding API, export, erasure design (crypto-shredding vs audit immutability) (dim 4).

**Hygiene (cheap, this week):** fix CLAUDE.md module inventory (no `docs`, add `tokens`); rename `repository/`→`infrastructure/` in documents+templates; outbox retention job; ADR 0022 split + 0013 Superseded stamp; write the missing api/worker/jobs Dockerfiles; codify the legacy-test deletion rule.

**Corrections to prior session beliefs (memory updated):** "dual capability sources" is a sound coarse/fine split, not a wart; T-001/T-005 error-envelope debt closed 2026-05-12; Playwright E2E exists and is coverage-gated; CLAUDE.md's `docs` module doesn't exist.

---

## Verdict

The architecture is at or near global maximum in its decisions; it is below global maximum in **enforcement automation**. Nothing found warrants a redesign. The incremental history did NOT lock in local maxima at the structural level — the ADR supersession record (0008→0050, auto-spawn→0052, envelope→9457-complete) shows the system corrects course. What incremental history DID leave behind is a trail of hand-maintained parallel truths, and both shipped incident classes of the last month (tripwire arms, nullable-not-required) came from exactly that trail. Closing P1 converts the remaining discipline-dependent invariants into structurally-enforced ones — the same move that already made module boundaries and capability naming incident-free.
