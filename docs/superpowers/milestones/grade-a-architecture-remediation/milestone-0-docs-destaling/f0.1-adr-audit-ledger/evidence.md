# Feature F0.1 — Evidence

> **Milestone:** 0 (Docs De-Staling)  ·  **Feature:** `f0.1-adr-audit-ledger`  ·  **Closed:** 2026-06-14
> A feature is closed only when every row below is filled with real output — not "done" / "green" / "looks good".

## What was implemented

- **Decision-vs-code re-audit** of all **25** `wiki/decisions/` ADRs (0001, 0002, 0003, 0007, 0008–0028; gaps 0004–0006 permanent). Result captured in [`drift-ledger.md`](drift-ledger.md): 23 MATCH, 2 STATUS-DRIFT (0001, 0027), 1 LEDGER-DRIFT (`index.md:34`), **0 DECISION-DRIFT** (no decision contradicted by code → no HS-6).
- **ADR 0027** (`wiki/decisions/0027-rls-adoption-sequencing.md`): reworded the ambiguous "27 tables" status clause → "27 remaining (29 total including the 2 already enabled in migration 0234)". Decision body unchanged.
- **ADR 0001** (`wiki/decisions/0001-eigenpal-adoption.md`): corrected stale vendor-path references (`vendor/eigenpal/` → real `apps/docx-renderer/vendor/eigenpal/`); added a `Path note (2026-06-14)` documenting the broken FE `file:` reference and its HS-2 defer. `Accepted` status unchanged. **No code touched.**
- **`wiki/decisions/index.md`**: replaced the stale legacy-ADR-tree sentence (referenced the deleted tree) with post-deletion reality (no navigable link); reconciled the 0001/0027 rows with the reworded statuses; bumped `Last verified:` → `2026-06-14 (M0/F0.1 …)`.

Commits: `060f536d` (ledger) · `0f8c8348` (status/vendor-path 0001+0027) · `31739e3b` (index refresh) · `dba08168` (reword to drop literal `docs/adr` token).

## Verification

| Check | Command / action | Result (evidence) |
|-------|------------------|-------------------|
| Status-presence gate | `foreach ($f in Get-ChildItem wiki/decisions/00*.md){ if(-not(Select-String $f -Pattern '^> \*\*Status:\*\*' -Quiet)){"NO STATUS: $f"} }` | **empty** (all 25 ADRs carry a `> **Status:**` line) |
| Vocabulary-valid status | grep each status line against `Accepted\|Historical\|Superseded by ADR\|Deprecated\|Proposed` | **empty** (every status drawn from the canonical vocabulary) |
| No stale `docs/adr` ref in ledger | `Select-String wiki/decisions/*.md -Pattern 'docs/adr'` | **empty** (zero matches) |
| ADR file count | `(Get-ChildItem wiki/decisions/00*.md).Count` | **25** |
| Decision-vs-code match | per-ADR checks in `drift-ledger.md` (live-tree Grep/Glob/Read, independently spot-checked) | 23 MATCH + 2 STATUS-DRIFT + 0 DECISION-DRIFT |
| No code modified | `git show --stat` on the 4 F0.1 commits | all touch only `wiki/decisions/**` + the feature folder |

> Docs-only feature — "runtime proof" is the gate command output above. No app behavior observable.

## Acceptance vs milestone spec

From `../milestone.md` F0.1: *"Every `wiki/decisions/` ADR has a `> **Status:**` line and that status matches code; drift is marked Historical/Superseded/amended; `decisions/index.md` ledger refreshed to match."*

| Acceptance criterion (from milestone.md) | Met? | Evidence |
|------------------------------------------|------|----------|
| Every ADR has a `> **Status:**` line | yes | status-presence gate empty (25/25) |
| Status matches code | yes | drift-ledger 23 MATCH + 2 STATUS-DRIFT corrected; 0 DECISION-DRIFT |
| Drift flagged/corrected | yes | 0001 + 0027 status/path corrected; index:34 stale ref replaced |
| `index.md` ledger refreshed to match | yes | rows reconciled + stamp bumped (commit `31739e3b`/`dba08168`) |

## Review disposition

- **Spec-compliance review (Task 1 ledger):** ✅ compliant. Independently re-ran 5+ checks (0001 vendor path, 0027 table count, 0009/0019/0025 MATCH); confirmed 25 grounded rows, no DECISION-DRIFT concealed as MATCH. Surfaced that the directory holds **25** ADRs, not the "27" my plan headline miscounted — plan corrected.
- **Spec-compliance review (Tasks 2+3):** ✅ compliant. Confirmed only status lines / body path refs / index rows / stamp changed; **no** decision section altered; **no** code file touched (`git show --stat` on both commits = `wiki/decisions/**` only); both acceptance greps empty.
- **Code-quality review:** docs-only markdown evidence artifact — handled inline by the controller (fixed 2 cosmetic summary miscounts in the ledger; reworded index prose to keep the acceptance grep clean). No separate quality subagent (no code surface to assess).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| **HS-2 — FE eigenpal `file:` path broken.** `frontend/apps/web/package.json` + `pnpm-lock.yaml` reference `file:../../../vendor/eigenpal/eigenpal-docx-js-editor-0.2.0.tgz`, resolving to a non-existent repo-root `vendor/eigenpal/`. Real tgz is at `apps/docx-renderer/vendor/eigenpal/`. Fresh `pnpm install` would fail. | Pre-existing code defect, discovered during docs audit; **out of scope** for a docs-only milestone (M0). Fixing it = a code-boundary change (copy/symlink the tgz to root, or repoint `package.json` + lockfile). Not in the Grade-A defect inventory. | **Trigger:** fix before any frontend `pnpm install` / FE feature work (incl. M2's `gen:api`/UsageGauges) — a clean install on this branch currently fails. **Owner:** operator to schedule a separate fix session (spawned as a tracked task). Documented in ADR 0001 `Path note (2026-06-14)`. |
