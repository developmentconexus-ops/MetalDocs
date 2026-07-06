# M9 validation contract (D4 — binding, authored before implementation)

> **Milestone:** milestone-9-governance-hygiene · **Authored:** 2026-07-06, before any F9.x work.
> **Binding:** the milestone-validator checks implementation against THIS file section-by-section.
> Divergence = **HS-7**: fix the work to the contract, or re-open the contract with operator approval
> and a dated erratum — never silently edit this file to match the work.

## 0. Runtime-truth basis (facts this contract is built on, verified 2026-07-06)

### 0.1 Modules on disk
`internal/modules/` contains exactly **14** directories:
`audit auth controlleddocuments distribution documents iam jobs notifications render search security taxonomy templates tokens`.
There is **no `docs` module**; `tokens` exists. CLAUDE.md currently lists
`audit · auth · controlleddocuments · distribution · docs · documents · iam · …` (has `docs`, lacks
`tokens`) — the F9.4 defect, confirmed.

### 0.2 Layer naming
- `internal/modules/documents/` has `repository/` (no `infrastructure/`) + a full-layer `approval/`
  subtree (`api application domain http infrastructure jobs repository`) — the hidden 15th module.
- `internal/modules/templates/` has **both** `infrastructure/` **and** `repository/`.
- `scripts/check-module-boundaries.ps1` contains **zero** mention of `approval`, `repository`, or
  `infrastructure` — approval is not specifically guarded today.

### 0.3 ADRs
- `wiki/decisions/` has **70** ADR files.
- Status-field length sweep (single-line `> **Status:**` char count): **0022 = 2757 chars** (the
  mega-status), **0027 = 665 chars**; all others under ~400. The ≤3-line audit in F9.1 must re-run
  this sweep with the line-based rule and treat its own output as the binding inventory.
- ADR 0013 status: `Accepted` (template REV labels). Its superseding decision must be **researched**
  (candidates: later template-versioning/naming ADRs); if research shows 0013 is NOT superseded,
  that contradicts the mission finding → HS-6 (surface, don't force a fake stamp).
- `wiki/standards/documentation-governance.md` currently contains **no ADR status-field rule**
  (grep for "status" over it: no rule text).

### 0.4 REQ traceability
- `wiki/architecture/backend-target-architecture.md`: **67 unique REQ IDs** (`REQ-<AREA>-<n>`), 86
  total mentions; RFC 2119 classification is **inline in each REQ line** (`(MUST)`, `(SHOULD)`,
  `(MAY)`, sometimes with satisfied-by annotations).
- Citations in tests today: **~10** `REQ-*` matches across `internal/`+`apps/` `*_test.go`.
- Existing CI gate patterns to mirror: `.github/workflows/e2e-coverage-gate.yml`,
  `module-boundaries.yml`, `governance-check.yml`.

### 0.5 Tests
- **386** `*_test.go` files under `internal/`; **12** call `t.Parallel()`.
- Full integration suite is NOT runnable locally in bounded time (mission §10 box constraint):
  wall-clock measurement uses a **named representative package set**, same command before/after.
- Test-framework hard gate + legacy-test memory rule are in force until F9.3 codifies the policy.

## 1. F9.1 — adr-hygiene (binding acceptance)

### 1.1 Status rule + sweep
- Rule text added to `wiki/standards/documentation-governance.md`: an ADR `> **Status:**` block is
  ≤3 lines; allowed vocabulary (Proposed / Accepted / Superseded by NNNN / Rejected / Deprecated,
  plus one optional date-and-scope line); execution history lives OUTSIDE the status field.
- A repeatable check (script or documented one-liner) sweeps `wiki/decisions/*.md`; **on the final
  tree it reports 0 violations**. The sweep command + output are in the feature evidence.

### 1.2 ADR 0022 split
- 0022 status reduced to ≤3 lines preserving: current state (Accepted, fully executed), date,
  pointer to the relocated history.
- Companion history doc (e.g. `wiki/decisions/0022-execution-history.md` or equivalent under
  `wiki/decisions/`) contains the relocated phase-history **verbatim-or-clearly-restructured, with
  zero information loss** — validator spot-checks ≥3 phase entries against git history of 0022.
- Any other ADR failing the ≤3-line rule (0027 confirmed; sweep may find more) gets the same
  split treatment. No ADR **decision content** (Context/Decision/Consequences) is altered.

### 1.3 ADR 0013 supersession
- 0013 status = `Superseded by <NNNN>` (or the researched true state surfaced via HS-6).
- The superseding ADR contains a back-reference (`Supersedes 0013`).
- Evidence records the research trail (which ADR/commit supersedes it and why).

## 2. F9.2 — traceability-gate (binding acceptance)

### 2.1 Gate semantics
- Scope: **MUST-classified** REQ IDs from `backend-target-architecture.md` (extraction is scripted,
  not hand-listed).
- A MUST REQ is *covered* iff the evidence map links it to ≥1 of: (a) a citing test file
  (`REQ-…` literal in a `*_test.go`), (b) a citing commit hash, (c) a satisfied-by annotation
  already present in the doc line itself. The map is a committed artifact, regenerable by one command.
- Gate exit code ≠0 when any MUST REQ has zero evidence. SHOULD/MAY are reported, non-blocking.

### 2.2 Proof pair (both mandatory in evidence)
- **POSITIVE:** gate run on the clean final tree → exit 0, output captured.
  **[AMENDED — see Erratum E1, 2026-07-06]** First-population positive proof = gate run on the clean
  final tree whose output (a) is anti-rot clean (`stale=false`, matches the committed
  `req-traceability.md` exactly) and (b) reports an uncovered-MUST set equal to **exactly** the four
  ledgered defers of E1 — no more, no fewer. The tool itself stays strict (exit ≠0 on any uncovered
  MUST); the erratum redefines the *acceptance check*, not the gate.
- **NEGATIVE:** a planted uncovered MUST REQ (or a temporarily removed citation) → exit ≠0, output
  captured, plant reverted. The negative run drives the REAL gate entrypoint (script or CI job step),
  not an internal function.

### 2.3 Anti-gaming clause
- Initial green must come from real evidence links. Bulk-annotating REQs as "satisfied" without a
  verifiable pointer (test/commit/doc-recorded wave reference) = FAIL. The validator samples ≥5
  MUST REQs from the map and verifies their evidence pointers resolve.

### 2.4 CI wiring
- A workflow (new or extending an existing governance job) runs the gate; job present in
  `.github/workflows/` and referenced in the feature evidence. Local execution command documented.

## 3. F9.3 — test-policy (binding acceptance)

### 3.1 Policy doc
- New doc under `wiki/quality/` (linked from `test-discipline.md` and `wiki/quality/index.md`):
  repair-class (guards a REQ ID, tripwire, contract shape, or DB invariant) vs delete-class
  (one-off task scaffolding), a decision procedure, and ≥2 concrete worked examples from this repo.
- Policy explicitly supersedes the informal memory rule and cites the test-framework hard gate.

### 3.2 t.Parallel expansion
- Expansion applied only where isolation is verified (testdb-factory or stateless); files left
  serial get a one-line reason in the plan/evidence (not in code noise).
- **Measured:** before/after wall-clock for the named package set, same command, same box, both
  outputs captured. A reduction is expected; if measurement shows no gain, the honest number is
  recorded and the reason analyzed (no fabricated win).
- All touched suites green after expansion (`go test <named set> -count=1`).

## 4. F9.4 — doc-truth (binding acceptance)

### 4.1 CLAUDE.md corrections (each paired with runtime evidence)
- Module inventory: exactly the 14 dirs of §0.1 (−`docs`, +`tokens`); the "14 bounded-context
  modules" count statement re-verified (stays 14).
- Idempotency wording: corrected to where idempotency actually executes (verify against the real
  middleware chain / handler wiring; if runtime shows it IS a chain link, the mission-brief claim is
  falsified → HS-6 surface, don't blind-edit).
- Janitor/scheduler wording: matched to post-M5 runtime (api-hosted leader-elected janitors vs
  River-scheduled jobs binary) — verified against the binaries' wiring (`apps/api/cmd/…`,
  `apps/…/metaldocs-jobs`).
- **Post-F9.5 consistency:** any CLAUDE.md/wiki text describing module layout reflects the final
  (post-rename, post-approval-decision) tree.

### 4.2 Wiki stamps
- The contract binds the **list** of mission-touched docs to stamp: every `wiki/` doc edited by
  M0–M9 feature work (enumerated in the F9.4 spec from the milestones' evidence files) gets a
  current `Last verified` stamp via wiki-curator; curator pass output in evidence (clean: no broken
  anchors in the touched set).

## 5. F9.5 — structure-hygiene (binding acceptance)

### 5.1 Rename
- After F9.5: `internal/modules/documents/repository/` and `internal/modules/templates/repository/`
  no longer exist; their contents live under the respective `infrastructure/` (templates: folded
  into the EXISTING `infrastructure/` coherently — no `infrastructure/repository/` nesting;
  documents/approval's own `repository/` dir is resolved per the same convention).
- Mechanical only: package moves/renames + import updates; **zero** exported-signature or behavior
  changes (validator checks the diff class).

### 5.2 Approval decision (mini gate)
- A mini `developing-new-work` gate is run BEFORE deciding; its written verdict is in the feature
  folder. Outcomes: (a) promote `documents/approval` → `internal/modules/approval/` (then CLAUDE.md
  says 15 modules — F9.4 consistency), or (b) ADR-recorded exception (new ADR under
  `wiki/decisions/`, status per the F9.1 rule) keeping it nested WITH boundary guards extended.
  Red verdict on promotion = take path (b) or escalate; never promote through a Red.

### 5.3 Boundary guard coverage
- `scripts/check-module-boundaries.ps1` (or its successor) demonstrably covers approval's boundary
  per the decided model. **Negative proof:** a planted cross-boundary import violation involving
  approval is caught (RED), then reverted; both outputs captured.

### 5.4 Gates
- `go build ./...` clean; `scripts/check-module-boundaries.ps1` green; targeted `go test` over
  documents/templates/approval packages green; api-lint blocking gate stays `0 blocking`.

## 6. Milestone close gates (validator re-runs from clean state)

1. `go build ./...` — exit 0.
2. `scripts/check-module-boundaries.ps1` — OK.
3. `go run ./scripts/api-lint …` blocking gate — 0 blocking (same invocation as CI `lint.yml`).
4. F9.2 gate — positive run per §2.2 as amended by Erratum E1: anti-rot clean AND uncovered set ==
   exactly the four E1 defers (its negative proof is in F9.2 evidence, not re-planted at close).
5. ADR status sweep (§1.1) — 0 violations.
6. Targeted test suites: the F9.3 named package set + documents/templates/approval packages —
   green. Full suite explicitly NOT required locally (box constraint, mission §10); the bounded
   defer is CI's `test-full.yml` on next push.
7. Regression spot: M2 tripwire parity test, M3/M7 RLS suite invocation named in their evidence,
   M8 ratelimit unit tests — green (M9 must not have touched them; a diff-scope check that M9's
   commits touch only docs/wiki/scripts/CI/test files + the F9.5 package moves).

## 7. Forbidden (validator fails on sight)

- Any `api/openapi/` change; any `migrations/` or `db/` schema change; any capability/authz edit.
- Deleting or summarizing-away ADR history (relocation must preserve content).
- Weakening an existing gate/lint/test to make an M9 check pass.
- Commits containing `docs/release/` or `docs/superpowers/plans/` content, or any `.env` material.
- Marking 0013 Superseded without a researched superseding reference.
- F9.2 map entries whose evidence pointer does not resolve (anti-gaming §2.3).
- Push to any remote.

## Errata

### E1 — 2026-07-06 — §2.2 POSITIVE proof & §6.4 close gate (operator-approved, HS-7 path)

**Trigger:** F9.2's gate, run honestly on the clean tree, found **4 MUST REQs with zero real
evidence** — genuine doc-vs-runtime falsity and coverage gaps of exactly the mission's meta-defect
class. Satisfying the literal "exit 0" of §2.2 would require inventing evidence (§2.3 forbidden) or
weakening the gate (§7 forbidden).

**Operator disposition:** operator was presented the three options (accept-RED erratum / fix-in-M9
F9.6 / defer-to-HS-1) and responded: *"What do you recommend and why? to be global maximum"* —
delegating to the session's recommendation with rationale recorded. Recommendation adopted:
accept documented RED; all 4 gaps become **bounded defers with named triggers**; no normative-doc
edits inside M9; gate strictness unchanged.

**Ledgered defers (the exact allowed uncovered set):**

| REQ | Finding | Trigger |
|-----|---------|---------|
| REQ-AUTHN-1 | Doc mandates Argon2id; runtime is bcrypt cost-12 (`internal/modules/auth/domain/model.go:12`). Constant-time verify + uniform failure real (b5f00d73). | Post-mission operator security-posture ADR: accept bcrypt-12 (re-grade doc line) **or** plan Argon2id migration. Not decidable inside a hygiene milestone. |
| REQ-AUTHN-3 | Doc assumes RFC 8725 JWT handling; runtime is opaque server-side session cookies — no JWT exists anywhere in `internal/modules/auth/`. | Post-mission ADR recording token-format truth (opaque sessions; RFC 8725 inapplicable) + doc-line amendment. |
| REQ-SEARCH-1 | No reindex/rebuild procedure or test; search is SQL-backed read projections. | Backlog feature: rebuild procedure + reindex test, or ADR re-grade if projections proven trivially rebuildable. |
| REQ-SEC-3 | OWASP ASVS referenced descriptively, never operationalized as a review checklist/gate. | Backlog feature: ASVS checklist wired into review gate (api-lint/governance-check family). |

**Amendment:** §2.2 POSITIVE and §6.4 now accept the first-population state defined above. Any
uncovered MUST **outside** this table remains a hard FAIL. The req-trace tool's own exit semantics
are unchanged (strict). These 4 defers are presented at HS-1 and carried into the mission close-out
defers ledger.

## 8. Feature → evidence map (shape fixed now, filled at close)

| Feature | Evidence file | Key proofs required |
|---------|---------------|---------------------|
| F9.1 | `f9.1-adr-hygiene/evidence.md` | sweep cmd+output (0 violations); 0022 split diff summary; 0013 chain research; governance-doc rule anchor |
| F9.2 | `f9.2-traceability-gate/evidence.md` | positive run output; negative run output (plant+revert); map sample verification; CI job ref |
| F9.3 | `f9.3-test-policy/evidence.md` | policy doc path+links; before/after wall-clock outputs; touched-suite green runs |
| F9.4 | `f9.4-doc-truth/evidence.md` | per-claim runtime evidence (file:line/cmd); wiki-curator pass output; touched-doc stamp list |
| F9.5 | `f9.5-structure-hygiene/evidence.md` | mini-gate verdict; rename diff class; boundary negative proof; §5.4 gate outputs |
