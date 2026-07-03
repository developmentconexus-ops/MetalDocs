# Discovery Brief — Global Maximum Remediation

**Date:** 2026-07-03 · **Mission type:** remediation · **Slug:** `global-maximum-remediation`

## Evidence base

Discovery for this mission was executed as the **final architecture global-maximum review** — a 10-dimension adversarial fan-out (one sonnet reviewer per dimension, code evidence + industry comparison with cited sources), synthesized and committed:

> **`docs/superpowers/analysis/2026-07-03-final-architecture-global-maximum-review.md`** (commit 778f494a)

That report IS the discovery artifact. Every finding below cites its dimension section (§1–§10) and the report's Priorities section; the report carries the file:line anchors and external source citations. This brief adds only the mission-facing index and the coverage statement.

Supplementary evidence artifacts (pre-existing, referenced by milestones):
- `docs/superpowers/analysis/2026-07-03-versionref-contract-refactor-system-impact.md` — developing-new-work gate (Yellow, 8 locked constraints) for the M0 contract refactor.
- `docs/superpowers/plans/2026-07-03-versionref-template-contract.md` — 13-task implementation plan for M0 (untracked; plans dir gitignored).

## Findings index (finding → mission milestone)

| # | Finding (report section) | Class | Milestone |
|---|--------------------------|-------|-----------|
| 1 | Flat coupled version/revision scalars in TemplateDTO/DocumentSummary; FE hand-written type overrides (§3) | contract shape | M0 |
| 2 | No oasdiff breaking-change CI gate; no nullable⇒required shape lint; contract-sync script advisory-only w/ 4 live DRIFT; redocly struct off + 133 suppressed (§3) | enforcement gap | M1 |
| 3 | FE feature-boundary rule #8 unenforced — no ESLint boundary config (§ cross-cutting #5) | enforcement gap | M1 |
| 4 | Tripwire cap-arms hand-synced to Go capability registry; 2 shipped P0001→500 incidents (0269/0270); pins reactive only (§7) | enforcement gap | M2 |
| 5 | Tier-1↔tier-2 capability-name divergences open (forceReleaseDocumentSession, approval-route-management) (§2) | correctness | M2 |
| 6 | ~85 manual SeedTxIdentity sites; RLS NULL-permissive ⇒ worker/jobs fleet has zero RLS backstop (§4) | enforcement gap | M3 |
| 7 | RLS async asymmetry documented only in a migration comment, not ADR 0027/wiki (§4) | docs gap | M3 |
| 8 | CanTransitionDocument covers 3/9 statuses; lifecycle fragmented across 4 approval services (§6) | correctness | M4 |
| 9 | Scheduled-vs-manual publish race unverified; no shared choke point demonstrated (§6) | correctness risk | M4 |
| 10 | Two concurrency transport idioms (If-Match vs body lock_version) (§6) | consistency | M4 |
| 11 | Three parallel job infrastructures (River + custom lease scheduler + StagingOutboxWorker poller) (§5) | architecture consolidation | M5 |
| 12 | Outbox tables never purged — unbounded growth (§5) | operability | M5 |
| 13 | No ordering guarantee on lifecycle fanout (published/superseded race plausible) (§5) | correctness risk | M5 |
| 14 | No periodic review/expiry; no structured reason-for-change — ISO 9001 §7.5.3 / Part 11 core eQMS gaps (§6) | product gap | M6 |
| 15 | No tenant onboarding API / offboarding / export / erasure; audit-immutability vs GDPR-erasure unresolved (T-003) (§4) | product gap | M7 |
| 16 | In-memory per-replica rate limiter — breaks on scale-out (§9) | ops blocker | M8 |
| 17 | No Dockerfiles for api/worker/jobs; compose references missing paths; DEPLOY.md/compose target inconsistency (§9) | ops blocker | M8 |
| 18 | No Prometheus /metrics (custom JSON only); log↔trace correlation unverified; no backup/restore doc (§9) | ops gap | M8 |
| 19 | ADR 0022 mega-ADR (changelog in status field); ADR 0013 never stamped Superseded (§10) | governance | M9 |
| 20 | REQ-ID→test traceability convention-only; no coverage automation (§8, §10) | enforcement gap | M9 |
| 21 | Legacy-test deletion policy unwritten (§8) | governance | M9 |
| 22 | CLAUDE.md stale: lists nonexistent `docs` module (missing `tokens`); idempotency described as chain link (it's per-handler); janitor scheduler wording (§1, §9) | docs defect | M9 |
| 23 | repository/ vs infrastructure/ layer-naming split in documents+templates (§1) | consistency | M9 |
| 24 | documents/approval structurally a hidden 15th module — boundary guards under-scope it (§1) | boundary | M9 |
| 25 | t.Parallel used in 52/260 integration files — wall-clock waste (§8) | test perf | M9 |
| 26 | Audit hash-chain validation window last 10k rows only (§7) | bounded risk | out-of-scope: tied to retention/partitioning T-013, tracked in wiki; revisit at retention design |
| 27 | No schemathesis-class contract fuzzing (§8) | test depth | out-of-scope: deferred post-v1; M1's gates cover the shipped bug classes |
| 28 | Missing threat model, SLO/capacity doc, consolidated C4 (§10) | docs backlog | out-of-scope: named backlog pre-v1 acceptable per review verdict |

## Coverage statement

- The 10-dimension review swept: module boundaries, authz, contract, tenancy, async, versioning kernel, DB invariants, testing, observability/ops, governance. It did NOT sweep: frontend visual/UX quality (separate frontend-screen mission owns that), the eigenpal vendored tree (third_party/, deferred post-v1 per memory), or performance/load testing (no findings claimed there).
- Findings 26–28 are deliberately out-of-scope with named triggers — they are recorded here so the terminal validator can confirm they were excluded by decision, not omission.
- One skeptic pass ran during the review itself (adversarial verify per dimension); prior-session belief corrections it produced (T-001/T-005 closed, E2E exists, dual-capability framing overturned) are recorded in the report.
