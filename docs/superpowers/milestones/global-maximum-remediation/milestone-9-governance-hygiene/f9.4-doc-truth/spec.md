# F9.4 — doc-truth (feature spec)

> **Milestone:** M9 governance-hygiene · **Contract:** `../validation-contract.md` §4 (binding)
> **Approved:** 2026-07-06 — approved against mission.md M9 row + contract §4 (operator-locked
> sources; autonomous session per mission D2). Code may start (docs feature).
> **Sequencing:** initial pass may run after F9.3; the FINAL verification pass runs after F9.5 so all
> layout text describes the post-rename tree (milestone.md ordering note).

## Consumer contract (first)

**Consumer 1 — every future session bootstrapped from CLAUDE.md:** reads a module inventory that
matches `internal/modules/` on disk exactly; a middleware-chain description matching `chain.go`; a
janitor/scheduler description matching the post-M5 River runtime.

**Consumer 2 — wiki readers:** mission-touched docs carry current `Last verified` stamps and valid
anchors (wiki-curator clean pass over the enumerated list).

**Consumer 3 — developing-new-work skill:** its `references/invariant-checklist.md` repeats the
stale idempotency-chain-link claim (line ~56) — references are governed like the wiki (skill's own
rule); the same correction lands there.

## Interview record (B1.5 — resolved from runtime, verified 2026-07-06)

| Q | A | Source (runtime truth) |
|---|---|--------|
| Module inventory? | 14 dirs: audit auth controlleddocuments distribution documents iam jobs notifications render search security taxonomy templates **tokens**. CLAUDE.md wrongly lists `docs`, omits `tokens`. Count stays 14. F9.5 keeps approval nested (ADR exception) so inventory unchanged by F9.5; CLAUDE.md should footnote the approval exception ADR. | `ls internal/modules/`; mini-gate artifact |
| Idempotency wording? | `chain.go:25-36` chain = panic_recovery → otel → http_obs → cors → origin → pre_auth_login_rate_limit → authn → iam_authz → presence → rate_limit → method_not_allowed. **No idempotency link.** Idempotency is per-handler/per-service (e.g. `approval/application/signoff_idemp.go`, idemp stores in `approval/infrastructure`). CLAUDE.md's "…tier-1 authz→idempotency→handler" chain claim is false; also its chain omits presence/origin — rewrite the lifecycle line to match `chain.go` (or reference it) rather than hand-listing wrongly again. | `apps/api/cmd/metaldocs-api/chain.go` |
| Janitor/scheduler wording? | Post-M5: janitors run as **River periodic jobs** (`maintenance.PeriodicJobs()` + retention) with River leader election; api binary joins the same election (`main.go:645-672`); the old lease scheduler is retired (M5), watchdog is alert-only (ADR 0068). CLAUDE.md's "4 leader-elected janitors — …, lease-reaper" is stale: implementer verifies the exact current janitor set from `maintenance.PeriodicJobs()` and writes that. | `main.go:645+`; M5 evidence; ADR 0068 |
| Which wiki docs stamped? | Enumerated by implementer from M0–M9 feature evidence files (`docs/superpowers/milestones/global-maximum-remediation/*/f*/evidence.md` "files touched" + wiki diffs in mission commits). The enumerated list is written into evidence.md; wiki-curator pass runs over that list. | Contract §4.2 |
| invariant-checklist reference fix? | Yes — same idempotency correction in `.claude/skills/developing-new-work/references/invariant-checklist.md` + bump its Last verified. | Skill's own governance rule |

## Non-goals (mandatory)

- No CLAUDE.md restructure — corrections only (inventory, lifecycle line, janitor wording, approval
  footnote); no new sections/rules.
- No wiki content rewrites — stamps, anchors, and factual one-liners the curator flags; nothing else.
- No edits to normative REQ text (F9.2's doc untouched unless a stamp).
- No memory-file edits (session memory is not repo truth).

## Validation Gate

1. `ls internal/modules/` output pasted next to the corrected CLAUDE.md inventory — exact match.
2. Corrected lifecycle line cross-checked against `chain.go` link names (evidence shows both).
3. Janitor wording cross-checked against `maintenance.PeriodicJobs()` source + `main.go` wiring
   (evidence shows the extracted job list).
4. Wiki-curator pass over the enumerated doc list — clean (no stale stamps/broken anchors in list).
5. invariant-checklist.md corrected + stamp bumped.
6. Diff scope: CLAUDE.md, wiki/** (stamps/anchors), skill references file, feature folder.
