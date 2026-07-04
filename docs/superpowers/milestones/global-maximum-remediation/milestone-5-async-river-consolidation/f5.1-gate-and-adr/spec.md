# F5.1 — developing-new-work gate + consolidation ADR (spec)

> **Milestone:** M5 (async consolidation onto River) · **Status:** Done (pre-implementation gate feature)
> **Type:** decision/gate feature (D7-mandated) — no runtime code; produces the design rails the rest of M5 builds within.

## Consumer contract (who consumes this, and the shape they require)

- **Consumer 1 — the milestone planner (this session):** requires a committed system-impact analysis with
  a **Green/Yellow** verdict (Red ⇒ HS-8, design blocked) and an **accepted ADR** before authoring
  `milestone.md` / `validation-contract.md` (mission D7 mandatory order).
- **Consumer 2 — the M5 milestone-validator:** requires the gate artifact + ADR 0067 to exist, be
  committed, and be **cited by** the later F5.x commits, as the rails it judges conformance against.
- **Consumer 3 — every F5.2–F5.5 implementer subagent:** consumes ADR 0067's decisions (River as single
  primitive; janitors→periodic jobs in `metaldocs-jobs`; staging→transactional job; retention; fanout
  commutative; H-PRE-1 retired; migration ordering) as locked constraints.

## What to implement

1. Run the `developing-new-work` skill → a written system-impact analysis (10 sections) with a
   Green/Yellow/Red verdict, committed under `docs/superpowers/analysis/`.
2. Author **ADR 0067** "Async job infrastructure consolidated onto River" recording: River as the single
   primitive; deployment topology; the **H-PRE-1 re-verification/retirement decision under River
   semantics**; outbox retention policy; fanout ordering vs idempotent-commutative proof. Accepted, indexed.

## Non-goals

- No runtime code, no migration, no test — this feature is the **decision**, not the implementation.
- Does not decide the concrete per-job schedule/idempotency values — that is `validation-contract.md`'s job
  (this feature makes the ADR-level decisions the contract then enumerates).

## Validation Gate (acceptance)

- Gate artifact committed; verdict is **Green or Yellow** (not Red). ✅ **Yellow** (10 locked constraints).
- ADR 0067 exists under `wiki/decisions/`, **Accepted**, in `wiki/decisions/index.md`.
- Both committed **before** any F5.2+ implementation (mission D7 order).
- River native-capability premise (periodic jobs + leader election + retention + `InsertTx`) re-proven
  against River v0.37.1 — recorded in the gate §0.

## Interview record

Not an operator interview — this is a system-orientation gate driven by the `developing-new-work`
static checklist + runtime-truth verification (agent code-map of the three job infrastructures) + River
v0.37.1 doc verification (Context7). The "answers" are the code anchors and the River-capability proof
captured in the gate artifact §0.

| Question the gate forced | Answer (evidence) |
|---|---|
| Do three parallel job infrastructures actually exist? | Yes — anchored map (gate §0.1): River/jobs, lease-scheduler/api, staging-poller/api. |
| Can River subsume the other two? | Yes — River v0.37.1 ships periodic jobs + `river_leader` elector + retention + `InsertTx` (gate §0, Context7-verified). |
| Is the foundation sound or a patch? | Local maximum — hand-rolled scheduler/backoff beside a River deployment (gate §2). Moving to global max ⇒ no AS-2 stop. |
| Any invariant violated? | No — transactional outbox preserved/strengthened; no AS-1 (gate §3). |
| H-PRE-1 disposition? | Retire — River elector+queue subsume the advisory lock; contingent on a singleton proof (ADR 0067 §H-PRE-1). |
| Verdict? | 🟡 Yellow — ADR mandated (D7) + H-PRE-1 + topology are real design choices; not Red, HS-8 does not fire. |

## ADR

**ADR 0067** — `wiki/decisions/0067-async-job-infrastructure-consolidated-onto-river.md` (Accepted 2026-07-04).
