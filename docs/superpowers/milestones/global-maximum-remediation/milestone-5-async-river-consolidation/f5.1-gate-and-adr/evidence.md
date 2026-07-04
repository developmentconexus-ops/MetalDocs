# F5.1 — evidence (gate + ADR)

> **Feature:** developing-new-work gate + consolidation ADR · **Status:** CLOSED · **Type:** decision/gate (no runtime code).

## Deliverables (committed)

| Artifact | Path | Commit |
|---|---|---|
| System-impact analysis (gate) | `docs/superpowers/analysis/2026-07-04-async-river-consolidation-system-impact.md` | `cd2bceb3` |
| ADR 0067 (Accepted) | `wiki/decisions/0067-async-job-infrastructure-consolidated-onto-river.md` | `5eb270c3` |
| Decisions index rows (0066 + 0067) | `wiki/decisions/index.md` | `5eb270c3` |

## Validation Gate — met

- ✅ Gate artifact committed; verdict **🟡 Yellow** (not Red) → HS-8 does not fire, design unblocked.
  10 locked constraints handed to design (analysis §9/§10).
- ✅ No AS-1 (no invariant violated — transactional outbox preserved/strengthened), no AS-2 stop
  (foundation named local-maximum, global-max structure = River-as-single-primitive proposed and
  taken), no AS-3 (owning modules unambiguous: jobs + render.fanout + notifications).
- ✅ ADR 0067 **Accepted**, under `wiki/decisions/`, present in `wiki/decisions/index.md`.
- ✅ Committed **before** any F5.2+ implementation (mission D7 order honoured).
- ✅ River native-capability premise re-proven against **River v0.37.1** (periodic jobs via
  `river.NewPeriodicJob`/`PeriodicInterval`/`RunOnStart`; leader election via `river_leader` +
  client `ID`; retention via `CompletedJobRetentionPeriod`/`CancelledJobRetentionPeriod`/
  `DiscardedJobRetentionPeriod`; transactional enqueue via `InsertTx`; uniqueness via
  `UniqueSkippedAsDuplicate`) — Context7-verified, recorded in analysis §0.

## Decisions the ADR settled (consumed by F5.2–F5.5)

1. **Topology** — janitors + staging dispatch hosted in `metaldocs-jobs` (already a required binary;
   no new binary). `metaldocs-api` becomes stateless sync + authz (drops the 4 hosted janitors).
2. **H-PRE-1 (advisory-lock deadlock constraint) — stays LIVE** *(corrected 2026-07-04 per ADR 0067
   §H-PRE-1 HS-7 erratum; this feature's original wording claimed "RETIRED" on a false premise)*. What
   M5 removes is the stuck-instance-watchdog's **unrelated** `pg_try_advisory_lock` single-runner guard
   (River's elector + single-queue claim subsume it) — **not** H-PRE-1. H-PRE-1 ("never call an
   authz-recording read inside a lock-holding atomic tx") is motivated by the **audit hash-chain
   writer's** `pg_advisory_xact_lock`, which M5 does not touch; it remains a live invariant. The
   watchdog-lock removal is contingent on a **singleton integration-proof** (only one instance runs a
   periodic job per tick) — carried as an F5.2 acceptance obligation — but that proof justifies the lock
   removal; it is **not** an H-PRE-1 retirement proof.
3. **Retention** — native River cleaner for River's own job rows + a River **periodic purge job** for
   the app-owned `staging_outbox` dispatched rows (F5.4). Values locked in `validation-contract.md` §4.
4. **Fanout ordering (F5.5)** — proven **idempotent-commutative**, not order-dependent: fanout writes
   additive per-recipient-per-event rows with `ON CONFLICT` de-dup and no shared mutable status row,
   so delivery order cannot change the terminal state. Race test required as the proof.

## TDD / runtime proof

N/A for this feature — it produces decisions and rails, not runtime code. TDD + runtime proof are the
obligation of F5.2–F5.5, whose evidence files carry them. The gate's own "test" is the
River-capability re-proof (above) and the AS-1/2/3 hard-stop screen (all clear).

## Review / QA disposition

Self-reviewed against the `developing-new-work` 10-section template and CLAUDE.md meta-rules
(Orientation + Global-Maximum). ADR self-reviewed against `documentation-governance.md` (status,
REQ-ID citation `backend-target-architecture.md:250-254`, index entry). No separate reviewer subagent —
a decision artifact's judge is the milestone-validator (C1) + the operator HS-1 gate.

## Bounded defers

None. All decisions the gate raised were settled in ADR 0067; the concrete per-job values were pushed
one layer down to `validation-contract.md` (by design — that is the D4 contract's job), not deferred.
