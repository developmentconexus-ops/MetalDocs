# Feature F0.1 — Spec

> **Milestone:** 0 — Truth reset & structural cleanup  ·  **Folder:** `f0.1-tracker-rewrite`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-21 / leandrotca — *schema + row-scope decisions answered via operator gate; contract below derived from them.*

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

`superpowers:brainstorming` not invoked as a separate engine — the two load-bearing contract
decisions were genuinely open and were resolved directly with the operator via an `AskUserQuestion`
gate (one batch, two questions); the rest of the contract is explicit in the consumer (mission §5
inventory + verified router read). Q&A persisted below.

| # | Question | Answer |
|---|----------|--------|
| 1 | Tracker schema — keep the existing redesign-**block** table (Block \| Description \| Plan \| Status) or **restructure** to a per-screen route table using the mission status vocab (done / partial / stub / not-started / cut)? | **Restructure per-screen.** New table keyed by screen/route + status vocab + in-scope milestone ref. (operator, 2026-06-21) |
| 2 | Which screens get a row — every routed screen (incl. already-DONE Login/Library/Editor/etc.) or only in-scope-to-complete screens? | **Every routed screen.** Full honest inventory; in-scope screens flagged with their milestone. (operator, 2026-06-21) |
| 3 | Governance header (current: Branch `feature/screen-redesign`, Spec `2026-05-05-screen-redesign-design.md`, Last updated 2026-05-08) — rewrite lineage or keep? | **Default applied (not a blocker):** keep the original redesign-spec lineage; update **Last updated → 2026-06-21**; add a pointer to `frontend-screen-completion/mission.md` as the current governing program. Row *truth* changes; lineage does not (M0 rabbit-hole rule). |

## Consumer contract (FIRST — before any producer)

The tracker is a **document**, not code; its "consumers" read it as the durable resume doc.

- **Consumer(s):** the operator (resume/status reference) and every later milestone (M1–M5) of this
  mission, which cite the tracker to know each screen's starting truth.
- **Contract (row schema):** a single per-screen status table. Each row =
  `| Screen | Route | Component (file) | Status | Milestone | Notes |` where:
  - **Status** ∈ `done` / `partial` / `stub` / `not-started` / `cut` (mission F0.1 vocab) — and the
    status MUST match the verified reality of the implementing page file (or its absence).
  - **Route** = the mounted path from the router (or `—` / `unmounted` when the component exists but
    is not mounted; `cut` slugs have no route).
  - **Milestone** = the mission milestone that completes the screen, or `done`/`out-of-scope`/`cut`.
  - Every **routed** screen in `src/app/AppRouter.tsx` + `features/**/routes.tsx` has exactly one row;
    net-new (M4 Obsoleto, M5 Signoff) and the two CUT slugs also each get a row for completeness.
- **Source of truth for the contract:** the verified router read this session (`AppRouter.tsx` + all
  `features/**/routes.tsx`), the implemented page files under `features/**/pages`, mission §5 work
  inventory, and `discovery-brief.md` findings 1–16.

## What this feature implements

Rewrite `wiki/implementation/screen-redesign-tracker.md` so its status section is the per-screen
table above, every row stating verified 2026-06-21 status. The known-wrong legacy rows are corrected:
Editor and Documento Publicado (claimed "Not started" / 🔲) now read their real status (Editor ships;
Publicado `partial`). The redesign-block "Status" table is replaced by the per-screen table; the
"Key Files Reference" + "Design System Reference" sections may stay (still-true reference), with the
header `Last updated` stamped 2026-06-21 and a governing-mission pointer added.

## Non-goals (mandatory)

- **Not** fixing any screen's code — no mock removal, no route fix, no stub deletion (those are
  F0.2/F0.3 and M1+). F0.1 only records truth.
- **Not** deleting the stale tracker's reference sections (Key Files / Design System) — they remain
  if still accurate; this is a row-truth rewrite, not a doc teardown.
- **Not** rewriting the tracker's spec/branch lineage to a new program (header pointer added, lineage
  preserved).
- **Not** authoring the per-screen Definition-of-Done — that is F0.4.

## Validation Gate (concrete — approved before code)

The tracker is a doc; proof is a **deterministic cross-check**, not a unit test (labeled honestly —
no TDD test object exists for a markdown rewrite).

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Every routed screen appears exactly once | `grep -c` rows vs router-derived route count; manual 1:1 reconcile in evidence | real |
| No row contradicts the implemented page set | For each row, the cited component file exists (or is absent for `not-started`/`cut`): `ls`/`grep` of `features/**/pages` | real |
| Status vocab is exactly the 5 terms | `grep -niE "done\|partial\|stub\|not-started\|cut"` covers every status cell; no legacy ✅/🔲/⏳ left in the status table | real |
| Known-wrong rows corrected | Editor row ≠ "not started"; Publicado row = `partial`; Dashboard row notes mock-data (M1) | real |
| Header stamped + mission pointer | `grep "2026-06-21"` and the mission path present in the header | real |

> No fixture/mock involved — every check is a real grep/ls against the working tree. The
> milestone-validator (C1/§4) re-greps a sample of rows against the implemented pages.

## ADR needed?

- [x] No durable decision — skip. (Documentation truth-sync; no architectural decision.)
