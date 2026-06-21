# Feature F0.1 — Evidence

> **Milestone:** 0  ·  **Feature:** `f0.1-tracker-rewrite`  ·  **Closed:** 2026-06-21
> **Contract:** `spec.md` (per-screen row schema + Validation Gate this proves against).

## What was implemented

- Rewrote `wiki/implementation/screen-redesign-tracker.md`: replaced the stale 2026-05-08
  redesign-**block** `## Status` table with a verified **per-screen** table (operator-chosen schema +
  scope: per-screen, every routed screen).
- Each row carries `Screen | Route | Component (file) | Status | Milestone | Notes`, status ∈
  `done/partial/stub/not-started/cut`. 24 rows: every routed screen + 2 net-new (Obsoleto, Signoff) +
  1 unmounted dead export (Auth route) + 2 CUT slugs.
- Corrected the known-wrong legacy rows: Editor `🔲 Not started` → `done`; Documento Publicado
  `🔲 Not started` → `partial`; Dashboard now records its `MOCK_STATS`/`MOCK_ACTIVITY` debt (M1).
- Header: `Last updated` → 2026-06-21, added `Governing program:` pointer to the mission; redesign-spec
  lineage preserved (per spec non-goal). Legend replaced with the 5-term vocab. Key-Files +
  Design-System reference sections retained (still accurate).
- **Producer matches consumer contract:** the table shape is exactly the `spec.md` row schema; the
  consumer (operator + later milestones) reads status/route/milestone per the agreed contract.
- Not yet committed (M0 commits at milestone close / operator discretion).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| TDD — failing test first | n/a — markdown doc rewrite, no test object (honestly labeled; spec Validation Gate is a deterministic cross-check, not a unit test) | — | — |
| Every cited component exists | `ls` of all 20 page files under `features/**/pages` | **20/20 OK**, 0 MISS | real |
| No legacy status symbols left | `grep -cE "🔲\|🔄\|⏳\|✅" tracker.md` | **0** | real |
| Header stamped 2026-06-21 | `grep -c "2026-06-21"` | **3** | real |
| Mission pointer present | `grep -c "frontend-screen-completion/mission.md"` | **1** | real |
| Editor row corrected | `grep "^\| Editor "` | `… \| done \| done \|` (was "Not started") | real |
| Publicado row corrected | `grep "Documento Publicado"` | `… \| partial \| M4 / F4.1 \|` | real |
| Status vocab on every data row | `grep -cE "\| (done\|partial\|stub\|not-started\|cut) \|"` | **24** cells = 24 rows | real |

> No fixture involved — every check is a real grep/ls against the working tree this session. The router
> inventory itself was read from `AppRouter.tsx` + every `features/**/routes.tsx`.

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Every routed screen appears exactly once | yes | 24-row table; router-derived screen set reconciled 1:1 (router read + `ls`) |
| No row contradicts the implemented page set | yes | 20/20 component `ls` OK; statuses match discovery findings 1–16 |
| Status vocab is exactly the 5 terms | yes | 24 vocab cells; 0 legacy ✅/🔲/⏳ symbols |
| Known-wrong rows corrected | yes | Editor=done, Publicado=partial, Dashboard notes mock (M1) |
| Header stamped + mission pointer | yes | `2026-06-21` ×3 + mission path ×1 |

All 5 criteria **met**.

## Review disposition

- Spec-compliance review: self-review against `spec.md` contract — PASS. Schema + row-scope match the
  operator-gated decisions; non-goals respected (no code touched, lineage preserved, DoD left to F0.4).
- Code-quality review: n/a (no code). Independent judgement deferred to the M0 `milestone-validator`
  (separation of powers) — it re-greps a sample of rows vs implemented pages (milestone.md §4).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| `authRoutes` (`AuthRoutePage`) is an unmounted dead export | Out of M0 delete scope — D7 limits F0.3 deletion to Operations/Audit | Recorded in tracker as `not-started / out-of-scope`. Trigger: if a later milestone needs `/auth`, route or delete it then; else sweep at program close. Owner: operator |
| Content Builder wrapper status | Thin wrapper, not a target screen | Tracker row `partial / out-of-scope`; HS-6 if it proves a real gap |
