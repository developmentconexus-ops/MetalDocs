# Feature F0.2 — Evidence — binding re-census against ADR-0039

> **Milestone:** 0 · **Feature:** `f0.2-binding-census` · **Closed:** 2026-06-20
> **Contract:** `spec.md` (consumer contract + Validation Gate). **Census of record:** `census.md`.

## What was implemented

- **`census.md`** — the binding cross-module owned-base-table read census, re-run against ADR-0039 and
  widened from the brief's 6 named tokens to the **full owned-base-table set** across `internal/modules/**`
  non-test. Contains: owner map (9 owning modules), Part 1 (the ~20 mission-scoped sites reproduced at cited
  lines, classed A/B/C/C4 with verdict + assigned milestone), Part 2 (NEW sites: N1 document-domain +
  X1–X8 auth/audit/platform), coverage statement (`unclassified: 0`), brief-delta.
- **`hs-6-scope-decision.md`** — the HS-6 surface: the two decisions (N1 shape change; X1–X8
  Non-Goal/terminal-bar contradiction), options, recommendation, and the **operator resolution** (1a + 2a).
- **ADR-0039 amendment** (HS-4 → F0.1, triggered by this census): D3(d) platform append-sink, D3(e)
  parent-ADR-dispositioned auth, D3(f) worker-layer; N1+X1–X8 classification table; honest "0 violations
  *outside the recorded allowlist*" scope note. (`wiki/decisions/0039-*.md`; F0.1 evidence carries the addendum.)
- **Mission replan** (HS-6-authorized): §5 row 16 (N1→M2), §2 Non-Goal + §4 updated to record the exemptions.

## Verification (real greps over the live tree, this session)

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| N1 reproduced | grep `templates_template_version` in `documents/application/fillin_service.go` | `225: FROM templates_template_version tv` — confirmed foreign read (templates-owned) by documents | real |
| `auth_failure_counters` is a false positive (same-module) | grep `auth_failure_counters` across `internal/modules` | writes live in **documents/approval**: `postgres_auth_failure_rate_limiter.go:64` `INSERT INTO`, `:86` `DELETE FROM`; its read `:32` `FROM` is **same-module** → compliant, dropped from in-scope | real |
| Brief B-sites reproduced | grep `FROM (public.)?controlled_documents|document_revisions|approval_instances` | B1 `documents/repository/repository.go:1701` `SELECT profile_code FROM controlled_documents` (→CD); B2 `controlleddocuments/infrastructure/repository.go:532` `FROM document_revisions` (→documents); B4 `…repository.go:593` `FROM approval_instances` (→approval) — all present at cited lines | real |
| X8 worker-layer reproduced | same grep | `jobs/stuck_instance_watchdog/job.go:147 FROM approval_instances ai` — confirmed cross-module worker read (exempt D3(f)) | real |
| Same-module reads correctly excluded | same grep | CD reading `controlled_documents` (`:35,:56,:74,:94,:476`), approval reading `approval_instances` (`postgres_approval_repository.go:260…`) are **own-table** — not flagged (diff filters reader==owner) | real |
| 0 unclassified | `census.md` coverage statement | explicit `Unclassified: 0`; every cross-module SQL read carries an ADR-0039 verdict; non-SQL A1–A3 out-of-D1-range; false positive recorded | real |

> Static-analysis feature — "real" = actual grep output over the live tree, recorded above and in `census.md`.
> No fixture. Rigor = per-site re-grep at cited `file:line`, not re-assertion of the brief.

## Acceptance vs spec Validation Gate

| Acceptance criterion (spec.md) | Met? | Evidence |
|--------------------------------|------|----------|
| Reproduces every brief site (A1–A3, B1–B8, C1–C4 ⇒ ~20), none dropped | yes | `census.md` Part 1 (all cited lines re-grepped); sample B1/B2/B4 + the resolution.go/area_catalog/search sites tabulated |
| Owned-table set widened beyond the 6 named tokens | yes | `census.md` owner map (9 modules) + Part 2 widen (iam_users, templates_template(_version), auth_*, audit_events, approval_* swept); N1 + X1–X8 are the widen's yield |
| **0 sites unclassified** | yes | `census.md` `Unclassified: 0`; every row carries a verdict |
| New in-scope sites routed/recorded; shape-changing ones → HS-6 | yes | N1 → HS-6 → M2; X1–X8 → HS-6 → ADR-0039 D3(d)–(f) exempt; `hs-6-scope-decision.md` resolution |
| Coverage statement present (swept set, residual, no silent caps) | yes | `census.md` coverage statement — swept paths, dynamic/aliased-SQL residual recorded, Docker-not-run noted (no false green) |

## Review / QA disposition

- Self-review against `spec.md` consumer contract: all contract elements present (table→owner map,
  authoritative in-scope list, brief-delta, coverage statement, `unclassified: 0`). HS-6 honored as designed
  (STOPPED, surfaced, applied ruling) rather than narrowing/expanding scope unilaterally.
- The independent `milestone-validator` (Phase 4) is the separation-of-powers reviewer of record for M0; it
  will spot-check the census's "0 remaining" claim by re-grepping a sample (spec C-consumer #3).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| Dynamic / aliased cross-module SQL (table name behind a Go var) invisible to a literal-token grep | Same residual as the H-D guard; no evidence of such a site found in the sweep; the F0.3 guard inherits and records the same limitation | **Trigger:** any future report of a runtime cross-module read the guard missed → extend the analyzer. **Owner:** mission (F0.3 + terminal re-audit). |
| Runtime/Docker (:5433) reproduction | M0 is static analysis by design (spec Q5); Docker may be down | **Trigger:** M1–M4 parity tests run against a live DB. **Owner:** later milestones. |
| `jobs` worker-layer boundary rule (X8) | Exempt-with-note under D3(f); whether jobs is a peer module is a separate question | **Trigger:** a dedicated jobs-boundary pass. **Owner:** future mission. |