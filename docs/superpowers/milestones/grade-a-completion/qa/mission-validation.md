# Mission Validation — grade-a-completion

> Written by: mission-validator subagent (separation of powers). Validates against: ../mission.md §8.
> Re-audit artifact judged: ../../../../wiki/backend/_artifacts/architecture-re-audit-2026-06-16.md
> Date: 2026-06-16 · Verdict: see bottom.

## Inputs loaded
- `mission.md` §8 Terminal acceptance — the binding rubric (four checks, all must hold simultaneously).
- `README.md` — all of M0..M4 show `passed` (validator PASS + HS-1 operator approval). Mission is
  procedurally ready for terminal validation; the gate is the re-audit result, not milestone closure.
- Re-audit report `wiki/backend/_artifacts/architecture-re-audit-2026-06-16.md` at HEAD `9a2a2f8d`
  — the fan-out artifact handed to me by the main session. It self-reports VERDICT: FAIL.

I did not trust the report's transcripts: I independently re-ran the §6 grep commands, spot-checked
fix sites with my own `Read`, and re-ran the full test suite.

## §8 Pass-bar (quoted)
> (1) module-boundaries, contract-api, and composition all ≥ A−; (2) 0 skeptic-confirmed new
> Critical/Major; (3) H-D = 0; (4) H-G = 0. All four must hold simultaneously.

## Per-criterion results

| # | §8 criterion | Method run | Real evidence | Pass? |
|---|--------------|-----------|---------------|-------|
| 1 | 3 formerly-C dims ≥ A− | Read report §2 scorecard | module-boundaries **B+**, contract-api **B+**, composition **A−**. Two of three below A−. | ❌ |
| 2 | 0 confirmed Critical/Major | Read report §3/§4; spot-checked authz.go:124 | **4** skeptic-confirmed Majors survive (authz effective_to predicate, iam_users upsert drops tenant_id, templates map[string]any, IAM admin map[string]any) | ❌ |
| 3 | H-D = 0 | Re-ran §6 grep on routes_generated.go myself | `128: writeJSON(...map[string]any{` and `238: writeJSON(...map[string]any{` — **H-D = 2** | ❌ |
| 4 | H-G = 0 | Re-ran §6 grep 1+2+3 myself | iam_user_roles read at auth/.../repository.go:104 **AND** hardcoded `"published"` at template_version_reader.go:44 — **H-G = 2** | ❌ |

## §8 Check 1 — Formerly-C dimensions ≥ A-
Report §2 scorecard: contract-api lifted C+ → **B+** (not A−); module-boundaries held **B+**; composition
lifted B+ → **A−** (only this one meets the bar). Pass-bar requires all three ≥ A−. **FAIL.**

## §8 Check 2 — Zero confirmed Critical/Major
Report §4 lists **4** skeptic-confirmed Majors; §5 records none were refuted/downgraded.
Spot-check of Major #1 (authz): `internal/modules/iam/authz/authz.go:124` reads
`upa.effective_to IS NULL`, which excludes memberships with a future `effective_to` — independently
confirming the "denies time-bounded active memberships" Major. The M0/F0.1 fix addressed
`effective_from <= now()` (line 123 present) but left the `effective_to` half of the predicate wrong.
**FAIL.**

## §8 Check 3 — H-D = 0
My own re-run of the §6 grep:
`grep -rn "map\[string\]any" internal/modules/templates/delivery/http/routes_generated.go`
→ lines 128 and 238 both emit `map[string]any` on public routes. **H-D = 2. FAIL.**

## §8 Check 4 — H-G = 0
My own re-run of all three §6 H-G greps:
- iam_user_roles cross-module: `auth/infrastructure/postgres/repository.go:104` (confirmed via Read —
  `GetUserTenants` issues raw `FROM metaldocs.iam_user_roles`, IAM-owned table, no port).
- hardcoded status literal: `templates/infrastructure/template_version_reader.go:44`
  (`status.String != "published"`, confirmed via Read — not the typed domain constant).
**H-G = 2. FAIL.**

## Regression gate
`go test -count=1 ./...` → all packages `ok`; no `FAIL`/`panic`/build error lines. Suite is green.
(Green tests do not rescue the verdict — the §8 bar fails on checks 1–4 regardless.)

## Spot-checks
- `iam/authz/authz.go:124` (M0/F0.1) — effective_from fixed; **effective_to predicate still wrong** →
  confirms surviving Major #1.
- `templates/infrastructure/template_version_reader.go:44` — hardcoded `"published"` present → H-G site real.
- `auth/infrastructure/postgres/repository.go:104` — raw `FROM metaldocs.iam_user_roles` → H-G site real.
- `routes_generated.go:128,238` — map[string]any on public routes → H-D sites real.

All spot-checks corroborate the report. No contradiction between the artifact and my independent reads.

## Forbidden-list (any hit = FAIL)
- [ ] Fixture/mock passed off as real-provider proof — n/a (grep/Read are real-code evidence)
- [ ] A criterion marked pass without a command actually run — none (all four marked FAIL on real evidence)
- [ ] Split-brain / guessed contract surfaced — the iam_user_roles direct read is a live boundary leak
- [ ] Self-judged / validator edited or fixed code — no; validator only re-grepped, read, tested, and wrote this file

## Verdict
- **VERDICT: FAIL**
- Failed criteria: **1, 2, 3, 4** (all four §8 checks). The artifact's own §8 verdict (FAIL) is
  independently corroborated by my re-greps and spot-checks.
- Bounded remediation micro-milestone needed to clear them (HS-5) — must, before re-audit:
  1. Fix the `authz.Require` `effective_to` predicate to honor time-bounded active memberships
     (root-cause at the shared predicate, mirroring `ResolveEligibleActors`; not per-caller).
  2. Fix the `iam_users` upsert to persist `tenant_id` (no silent system/default-tenant write).
  3. Eliminate both `map[string]any` emits in `templates/.../routes_generated.go:128,238` →
     typed OpenAPI response structs (drives H-D → 0 and lifts contract-api toward A−).
  4. Replace the hardcoded `"published"` in `template_version_reader.go:44` with the typed domain
     constant, and port `auth.GetUserTenants` behind an IAM-owned reader port to remove the
     `iam_user_roles` direct read (drives H-G → 0; keep read off any lock-holding tx, H-PRE-1).
  5. Sweep the remaining contract map[string]any emits (IAM admin overview, audit events, auth login,
     documents fillin/view) to reach contract-api ≥ A−.
  Then the main session re-runs the F5.1 re-audit fan-out and re-dispatches this validator. The operator
  decides continue vs replan. Per §8, any single dimension missing twice is an HS-2 design-boundary signal.
- The mission stays **open**. The main session does not flip status, does not declare Grade A; no §12 close-out.
