# Mission: Global Maximum Remediation

> **Status:** Drafting (operator gate pending)
> **Date:** 2026-07-03  ·  **Branch of record:** main
> **Type:** remediation
> **Slug:** `global-maximum-remediation`  ·  **Owner / operator:** Leandro
> **Evidence base:** `./discovery-brief.md`  ·  **Program index:** `./README.md`
> **Governs:** Milestones M0..M9 below. Each milestone gets its own plan via the `milestone` skill,
> executed in a fresh session. This file is the **stable governing contract** — it says *what* the
> mission is and *what proves it done*; it contains **no execution detail**.

---

## 1. Problem / why now

The 2026-07-03 final architecture review (`docs/superpowers/analysis/2026-07-03-final-architecture-global-maximum-review.md`, commit 778f494a) found no Red dimension and confirmed the load-bearing architecture decisions — but identified one systemic defect class: **hand-synced enumerations and unenforced conventions** (tripwire arms, manual GUC seeding, advisory-only contract sync, unenforced FE boundaries, aspirational REQ-IDs). Both shipped incident classes of the last month (tripwire P0001→500s, nullable-not-required wizard bug) came from exactly this class. Additionally: the versioning lifecycle is not a single state machine, three parallel job infrastructures coexist, two ISO-core eQMS product capabilities are missing, and three named ops blockers stand between today and the first paying customer. Why now: pre-v1 is the only window where breaking cutovers are cheap (F-18 re-baseline plan) and where enforcement gates can be installed before external consumers exist.

## 2. Goals / Non-Goals

**Goals**
- Convert every discipline-dependent invariant named in the review into a structurally-enforced one (generator or blocking CI gate) — cross-cutting findings 1–7 of the report.
- Close the kernel correctness risks: unified 9-status state machine; publish-race proven safe or choked.
- Consolidate async onto one job infrastructure (River), retiring the custom lease scheduler and the staging poller.
- Land the two ISO-core eQMS product gaps (periodic review/expiry, structured reason-for-change) and the tenant lifecycle kernel (onboarding/export/erasure design).
- Clear the three pre-customer ops blockers (rate limiter, Dockerfiles, metrics/backup posture).
- Close governance debt (ADR hygiene, REQ-ID traceability gate, doc corrections incl. CLAUDE.md).
- Terminal bar: an independent re-run of the 10-dimension review reaches **CONFIRMED on every in-scope dimension** with all new CI gates green from clean state.

**Non-Goals** (excluded by decision; triggers recorded in discovery-brief findings 26–28)
- Audit hash-chain full-history validation (tied to retention/partitioning T-013 — separate design).
- Schemathesis-class contract fuzzing (post-v1; M1 gates cover the shipped bug classes).
- Threat model, SLO/capacity doc, consolidated C4 (named backlog; acceptable pre-v1 per review).
- Frontend visual/UX completion (owned by the frontend screen-completion mission).
- Any eigenpal/third_party packaging work (deferred post-v1 per standing decision).
- No push to origin at any point without explicit operator permission.

## 3. Locked decisions (operator-approved 2026-07-03)

| # | Decision | Value |
|---|----------|-------|
| D1 | Scope | Full: P1–P4 + hygiene from the review's Priorities section. Out-of-scope items limited to the three named in §2 Non-Goals with recorded triggers. |
| D2 | Execution | One mission.md governs. Each milestone runs in a **fresh session dispatched via `spawn_task`** with a self-contained context-brief handoff, opening with a **`/goal` command** stating the milestone's done-condition. Fresh session per milestone; operator HS-1 gate between milestones. |
| D3 | Proof of done | **Re-review + gates green:** independent re-run of the 10-dimension review (fan-out from main session) — every in-scope dimension must reach CONFIRMED — PLUS every CI gate this mission installs green from clean state. Judged by `mission-validator`. |
| D4 | Validation-contract-first | **Every milestone authors a detailed `validation-contract.md` BEFORE implementation starts** — expected behaviors, shapes, gate outputs, and observable end-states, spelled out without sparing detail. Implementation drift is compared against the contract, never rationalized after. Milestones with runtime-visible behavior end with a **live/preview QA drive** (API drives, functional verification) as contract evidence. |
| D5 | Sequencing | M0 first (ready plan; proves the loop). Enforcement gates (M1–M3) before correctness (M4) before consolidation (M5) before product features that depend on consolidated scheduling (M6–M7). Ops (M8) and governance/hygiene (M9) last — risk-isolating, cannot regress the bar. |
| D6 | CLAUDE.md edits | Approved: module inventory, idempotency wording, janitor-scheduler wording corrected in M9. |
| D7 | Pre-design gates | M5, M6, M7 are architecture/feature-class work: each MUST run the `developing-new-work` skill gate (system-impact analysis, Green/Yellow required) before its milestone plan is authored. M5 and M7 additionally require an ADR landed with the change. |

## 4. Discovery summary

Discovery = the committed 10-dimension adversarial review (verified: per-dimension code anchors + external source citations + one skeptic pass; belief corrections recorded). Confidence high on all findings feeding M0–M4 and M8–M9 (direct code evidence); M5's consolidation claim (River can subsume janitors + staging poller) is evidence-based but must be re-proven by M5's own developing-new-work gate before design; M6/M7 product gaps are standards-based (ISO 9001 §7.5.3, 21 CFR Part 11, GDPR) and scoped by their own gates. See `./discovery-brief.md` — every milestone below traces to a numbered finding.

## 5. Work / requirement inventory

Full finding→milestone mapping lives in `./discovery-brief.md` (findings table, 28 rows: 25 in-scope → M0–M9, 3 out-of-scope with reasons). That table is normative; this section does not duplicate it.

## 6. Program architecture (by reference)

This mission executes via the `milestone` skill. The per-feature close-out loop, the per-feature consumer-contract spec gate, and the per-milestone `milestone-validator` gate are defined there — not duplicated here. See `.claude/skills/milestone/SKILL.md`. Program-scale shape:

```
Mission: global-maximum-remediation
└── Milestone (M0..M9)           ── each: validation-contract.md (BEFORE impl, per D4)
    │                                → features → milestone-validator gate → HS-1 operator gate
    └── Feature (Fx.y)           ── each: spec(consumer-contract) → plan → TDD → evidence
Terminal acceptance (§8)         ── 10-dim re-review fan-out + CI-gate rerun, judged by mission-validator
```

Mission-specific addition (D4): each `milestone-<n>-<slug>/` folder contains `validation-contract.md`, authored and committed before the first feature's implementation begins. The milestone-validator's checklist includes verifying the implementation against that contract.

## 7. Milestones

### M0 — VersionRef contract refactor
**Objective:** Land the already-planned nested version-reference contract cutover (templates + documents), proving the plan→review→implement loop on a ready workpiece. (Findings 1.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F0.1 versionref-cutover | Execute `docs/superpowers/plans/2026-07-03-versionref-template-contract.md` (13 tasks): `TemplateVersionRef`/`DocumentRevisionRef` component schemas, TemplateDTO + DocumentSummary reshaped, regeneration BE+FE, consumer updates, ADR 0065 | Marshal-shape pin tests pass (published_version present-and-null; nested ref field-set); `go build ./...`; targeted go tests; `tsc --noEmit`; vitest for documents/templates/taxonomy/approval; openapi lint; live drive GET /templates + /documents + wizard Step 3 |
| F0.2 adr-0065 | ADR 0065 "Version references are nested value objects in wire contracts" incl. pre-v1 atomic-cutover exception | ADR exists, cited by F0.1 commits; ADR 0035 annotated structurally closed for this class |

**Milestone gate:** contract gate + docs gate (per gate artifact's QA subset); validation contract = the plan's per-task expected outputs + the gate artifact's 8 locked constraints. Live QA drive required (runtime-visible). Validated by `milestone-validator`.

### M1 — Contract & frontend governance gates
**Objective:** Make contract truth and FE boundaries machine-enforced — the exact 9f86828b bug class becomes structurally impossible. (Findings 2, 3.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F1.1 oasdiff-gate | Breaking-change CI gate (oasdiff or equivalent) on api/openapi/v1/openapi.yaml PRs | Gate fails a synthetic breaking change; passes a compatible one; wired in CI workflow |
| F1.2 shape-lint | Nullable⇒required lint (custom redocly rule or script) + redocly `struct` re-enabled with suppression burn-down plan | Lint fails a nullable-not-required field fixture; suppressed-error count recorded with owner/trigger |
| F1.3 contract-sync-ci | `check-module-contract-sync.ps1` promoted to blocking CI; the 4 live templates DRIFT items fixed | CI red on injected drift; green on clean tree; zero live DRIFT |
| F1.4 fe-boundaries | ESLint config with feature-boundary rule (no cross-feature imports) + explicit allowlist; remove remaining hand-written FE type overrides not already deleted by M0 | ESLint red on synthetic cross-feature import; zero Omit<>-style overrides of generated types |

**Milestone gate:** each gate demonstrated failing-then-passing from clean state (negative + positive proof per D4 contract). No live drive needed (build-time only).

### M2 — AuthZ/DB enforcement generation
**Objective:** Tripwire arms and capability naming become generated/gated, retiring the two-incident hand-sync class. (Findings 4, 5.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F2.1 tripwire-generation | Derive `enforce_capability_asserted()` arms from the Go capability registry (codegen of migration content or a `cap_table_requirements` lookup table) + CI drift check scanning `authz.Require` call sites | Drift check red when a new asserted cap lacks an arm; existing tripwire_caps_test.go pins still green; adding a synthetic capability without an arm fails CI, not runtime |
| F2.2 cap-name-divergence | Close the two tier-1↔tier-2 divergences (forceReleaseDocumentSession, approval-route-management) — align or ADR-record as intentional | Route→capability table and in-tx Require agree, or a written exception exists; authz CI lints green |

**Milestone gate:** authz gate + DB-invariant gate; integration tripwire drives green; validation contract enumerates every tripwire-gated table and its expected arm set.

### M3 — Tenancy enforcement chokepoint
**Objective:** RLS backstop covers the async fleet; GUC seeding stops being 85 manual acts of discipline. (Findings 6, 7.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F3.1 txrunner-autoseed | Auto-seed tenant/actor GUCs at the TxRunner chokepoint (or tenant-scoped TxRunner variant); collapse manual SeedTxIdentity sites | No hand-seeded site remains outside the chokepoint (grep census = 0 or allowlisted); tenant-isolation integration tests green |
| F3.2 async-rls-backstop | Seed GUCs per message in worker/jobs so FORCE RLS backstops async; or ADR-amend the accepted gap with a compensating lint | A worker query missing a tenant predicate is caught by RLS in an integration test (negative proof), or the ADR amendment + lint exist |
| F3.3 adr0027-amendment | Document the NULL-permissive design + async posture in ADR 0027/wiki (currently only a migration comment) | ADR/wiki updated with the asymmetry, its rationale, and the new enforcement |

**Milestone gate:** multi-tenant gate; cross-tenant 404 suite green; validation contract states the exact RLS behavior expected per binary (api/worker/jobs).

### M4 — Versioning kernel correctness
**Objective:** One exhaustive document state machine; publish race proven safe. (Findings 8, 9, 10.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F4.1 state-machine-unification | Single exhaustive transition function covering all 9 document statuses (pattern: templates' CanTransition); approval services route through it | Table-coverage test proves all 9 statuses × transitions handled; no scattered `if status !=` lifecycle guards remain in approval services |
| F4.2 publish-race | Concurrent test: scheduled vs manual publish racing on one revision; if unsafe, single PublishRevision choke point | Race test green (deterministically exercises both orders); at most one publish wins with correct terminal state |
| F4.3 concurrency-idiom | Unify optimistic-concurrency transport (decide If-Match vs body lock_version; migrate the minority) or ADR-record the split | One idiom across documents+templates, or written exception |

**Milestone gate:** DB-invariant + contract gates; kernel lifecycle integration suite green; validation contract = the full 9-status transition table written out before implementation.

### M5 — Async consolidation onto River
**Objective:** One job infrastructure: janitors + staging outbox onto River; retire the custom lease scheduler and poller; fix outbox growth and fanout ordering. (Findings 11, 12, 13.) **Requires developing-new-work gate (Green/Yellow) + ADR before planning (D7).**

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F5.1 gate-and-adr | developing-new-work system-impact analysis; ADR for job-infrastructure consolidation (incl. H-PRE-1 re-verification under River semantics) | Gate artifact committed, verdict Green/Yellow; ADR accepted |
| F5.2 janitors-on-river | 4 janitors as River periodic jobs with leader election; custom lease scheduler + job_leases retired | Janitor behaviors preserved (each has an integration proof); lease scheduler code deleted; H-PRE-1 constraint re-verified or retired with evidence |
| F5.3 staging-on-river | StagingOutboxWorker dispatch via River transactional enqueue; duplicated backoff code deleted | pdf/materialize dispatch integration proof; idempotency preserved |
| F5.4 outbox-retention | Purge/retention job for dispatched outbox rows | Retention job proof; growth bounded (dispatched rows older than policy removed) |
| F5.5 fanout-ordering | Ordering guarantee (or explicit idempotent-commutative proof) for lifecycle fanout published/superseded | Race test or formal argument in ADR; no lost/inverted terminal state |

**Milestone gate:** async gate; all binaries build + system-runnable check; live drive of scheduled publish + notification fanout. Validation contract enumerates every job, its schedule, its idempotency key, and its failure behavior BEFORE migration.

### M6 — eQMS product gaps: periodic review + reason-for-change
**Objective:** ISO 9001 §7.5.3 / Part 11 core capabilities: document periodic review/expiry and structured reason-for-change. (Finding 14.) **Requires developing-new-work gate before planning (D7).**

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F6.1 gate | developing-new-work system-impact analysis (owning modules, capabilities, contract, schedule dependency on M5's River base) | Gate artifact committed, Green/Yellow |
| F6.2 periodic-review | Review-due/expiry model + scheduled surfacing (River) + capability-gated review workflow, contract-first | Contract + generated code + FE consumer; scheduled job proof; live drive of a due-review cycle |
| F6.3 reason-for-change | Structured reason-for-change captured at revision creation (contract field(s), not free text only) | Contract + pin tests; revision-creation drive shows structured capture; audit trail carries it |

**Milestone gate:** full gate set (authz + contract + DB-invariant + docs); live/preview QA drive mandatory (user-facing feature). Validation contract defines the expected workflow states and API shapes up front.

### M7 — Tenant lifecycle kernel
**Objective:** Onboarding, export, and erasure become designed capabilities, resolving audit-immutability vs GDPR-erasure. (Finding 15.) **Requires developing-new-work gate + ADR (crypto-shredding decision) before planning (D7).**

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F7.1 gate-and-adr | System-impact analysis + ADR: tenant lifecycle model incl. erasure strategy (crypto-shredding of PII vs immutable audit skeleton) | Gate artifact Green/Yellow; ADR accepted |
| F7.2 onboarding | Tenant onboarding path (API or operator runbook per ADR scope) replacing manual seed-only | New tenant reachable end-to-end (login → capability-gated action) from the onboarding path |
| F7.3 export-erasure | Tenant data export + erasure per ADR design (implementation depth per gate verdict — design-complete minimum, implemented preferred) | Export produces complete tenant-scoped artifact; erasure demonstrably removes/shreds in-scope data while audit chain integrity validation stays green |

**Milestone gate:** multi-tenant + DB-invariant + docs gates; live drive of onboarding. Validation contract defines exact data inventory in/out of erasure scope before implementation.

### M8 — Ops readiness
**Objective:** Clear the three named pre-customer blockers and the ops-posture gaps. (Findings 16, 17, 18.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F8.1 dockerfiles | Dockerfiles for metaldocs-api/worker/jobs at the paths compose references; resolve DEPLOY.md vs compose target inconsistency | `docker compose build` succeeds; system-runnable check passes against containers; DEPLOY.md consistent |
| F8.2 rate-limiter | Distributed rate limiter (redis_rate class) or ADR-recorded single-replica constraint with scale-out trigger | Limiter correct across 2 instances in a test, or ADR + startup guard exists |
| F8.3 metrics-backup | Prometheus /metrics endpoint (or ADR keeping JSON with named trigger); log↔trace correlation verified; backup/restore runbook written | /metrics scrapes (or ADR); a request's trace id provably links logs↔trace; runbook executes against local stack |

**Milestone gate:** system-runnable + obs checks green; validation contract lists each blocker's observable exit criterion.

### M9 — Governance & hygiene close-out
**Objective:** Governance debt and structural hygiene cleared; docs truth restored. (Findings 19–25 + D6.)

| Feature | What to implement | What to validate (acceptance) |
|---------|-------------------|-------------------------------|
| F9.1 adr-hygiene | ADR 0022 split (status ≤3 lines, history to linked doc); ADR 0013 stamped Superseded; ADR status-field rule written into governance doc | All ADR statuses ≤3 lines; supersession chain complete; rule documented |
| F9.2 traceability-gate | REQ-ID→test traceability automation (grep-map gate, e2e-coverage-gate pattern) for backend REQ IDs | Gate red when a MUST REQ has zero citing test/commit; green on clean tree |
| F9.3 test-policy | Legacy-test deletion taxonomy written (guards REQ/tripwire/contract = repair; else delete); t.Parallel expansion across integration files | Policy in wiki/quality; integration wall-clock reduction measured |
| F9.4 doc-truth | CLAUDE.md corrections (module inventory: −docs +tokens; idempotency per-handler; janitor scheduler wording); wiki stamps refreshed for mission-touched docs | CLAUDE.md matches runtime truth; wiki-curator pass clean |
| F9.5 structure-hygiene | repository/→infrastructure/ rename (documents, templates); documents/approval promoted to top-level module OR ADR-recorded exception (decision via mini gate) | One layer-naming convention; boundary guards cover approval; build + module-boundary checks green |

**Milestone gate:** docs gate + module-boundary checks + full build; risk-isolating by design (cannot regress earlier milestones' bars).

## 8. ★ Terminal acceptance — definition of done (written up front)

- **Pass bar:** (a) an independent re-run of the 10-dimension architecture review returns **CONFIRMED on every in-scope dimension** (dimensions 1–10, with observability judged against M8's scope, product-gap findings against M6/M7's) and **no new DEBT/RE-LITIGATE introduced by mission work**; (b) **every CI gate installed by this mission is green from clean state** and each has a recorded negative proof (it fails when it should); (c) all three out-of-scope findings remain excluded-by-decision with their triggers intact.
- **What to validate:** the re-review verdict table; the gate inventory (F1.1–F1.4, F2.1, F3.2-lint-if-chosen, F9.2) each with positive+negative proof; per-milestone `validation-contract.md` compliance confirmed by each milestone's validator verdict; discovery findings 1–25 each traceable to a shipped feature evidence row.
- **How to validate:** fan-out validation — the **main session** re-runs the 10-dimension review (same dimension charters as the 2026-07-03 review, fresh adversarial agents, inline-only fallback on rate-limit) and captures the artifact; then dispatches `mission-validator` to judge that artifact + the gate evidence against this section. Deterministic parts (`go build ./...`, gate CI runs, system-runnable check) run from clean state.
- **Who validates:** the independent `mission-validator` subagent — judges and writes `qa/mission-validation.md` only; never edits code, never flips status. Invoked by whichever session closes M9 (the program README close-out checklist carries the dispatch line).
- **On miss (HS-5):** missed criteria become a bounded remediation micro-milestone via `milestone`, then re-dispatch. Operator decides continue vs replan at each loop.

## 9. Hard-stop catalog

| ID | Trigger | Action |
|----|---------|--------|
| HS-1 | Every milestone boundary | Operator review gate; no next milestone / no merge without approval |
| HS-2 | A fix implies redesign outside the assigned boundary | Stop; report boundary + minimum prerequisite plan; no symptom-patch |
| HS-3 | Prerequisite boundary fails (build/runnable/auth/route/contract truth) | Repair prerequisite first; rerun checkpoint; resume |
| HS-4 | `milestone-validator` returns FAIL | Open the named fix feature; re-run lifecycle; re-dispatch validator |
| HS-5 | Terminal `mission-validator` misses the §8 bar | Bounded remediation micro-milestone; re-dispatch; operator decides |
| HS-6 | Scope drift / off-plan discovery | Stop; surface deviation; replan before continuing |
| HS-7 (mission-specific) | A milestone's implementation deviates from its committed `validation-contract.md` | Stop; either fix the implementation to the contract or re-open the contract WITH operator approval — never silently adjust the contract to match the code |
| HS-8 (mission-specific) | M5/M6/M7 developing-new-work gate returns Red | Design blocked; surface the redesign gate to the operator before any planning |

## 10. Constraints respected

- All CLAUDE.md always-on rules: never read/print/commit `.env`; PowerShell startup scripts; runtime truth beats docs; evidence before closure; commit-without-push standing authorization; do not commit `docs/release/`.
- Non-negotiable invariants (capabilities-never-roles per ADR 0022; contract-first via openapi.yaml + regen; pooled tenancy; transactional outbox; DB-enforced invariants; RFC 9457) — every milestone's validation contract restates the ones it touches.
- H-PRE-1 advisory-lock constraint holds until M5 formally re-verifies or retires it.
- Test framework hard gate (testdb factory for DB integration); legacy-test policy per memory until F9.3 codifies it.
- Full integration suite is NOT run locally (20+ min box constraint) — targeted `-run` filters; bounded defers recorded in evidence.
- Model policy: sonnet implement/review, haiku mechanical, never fable workers, ≤15 concurrent.
- Plans dir (`docs/superpowers/plans/`) is gitignored — never force-add.

## 11. Execution model

One `mission.md` governs all milestones. Per D2: each milestone is dispatched to a **fresh session via `spawn_task`**, whose prompt is a self-contained context brief (mission path, milestone id, evidence pointers, constraints) opening with a **`/goal`** statement of the milestone's done-condition. Inside the session: `milestone` skill owns the flow — author `milestone.md` + **`validation-contract.md` first (D4)**, then per-feature spec→plan→TDD→evidence with subagent-driven implementation, then `milestone-validator`, then HS-1 operator gate. Preview/live QA drives close every runtime-visible milestone (M0, M5, M6, M7, M8). Token discipline: fan-out only where it pays (M5–M7 gates, terminal re-audit); direct tools elsewhere. Commit after verified work; never push.

## 12. End-state / reconciliation

Fill only when the last milestone has passed and the terminal gate is green:
- [ ] Every planned feature (M0..M9) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] Terminal acceptance (§8) passed — link `qa/mission-validation.md`.
- [ ] Operator sign-off: <date / name>
