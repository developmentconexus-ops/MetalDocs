# Program: Backend Module-Boundary Hardening

> **Governing spec:** `./mission.md`
> **Status:** Executing — M0 **passed** + M1 **passed** + M2 **passed** + M3 **passed** (all validator PASS + operator HS-1; M0/M1 2026-06-20, M2/M3 2026-06-21). M3 HS-1 approved 2026-06-21 (not merged, not pushed). **M4 validator PASS 2026-06-21 — awaiting HS-1 operator acceptance** (not merged, not pushed). On M4 HS-1, the H-G debt ledger is empty → mission terminal acceptance (mission.md §8) is ready for the mission-validator.
> **Owner / operator:** leandrotca

Eliminate the pre-existing cross-module raw-SQL debt (~20 sites in 3 categories — see
`./discovery-brief.md`) that the parent `grade-a-completion` post-M9 re-audit surfaced but never measured, so
the module-boundaries / DDD dimension reaches **A** and **H-G = 0 under both readings** (canonical §6 greps
AND the broad "any cross-module owned-base-table read", reconciled by ADR-0039's published-contract
exemption). Terminal acceptance = a fresh re-run of the F5.1 10-dimension architecture re-audit hitting that
bar, judged by an independent `mission-validator`. The parent's **Grade-A sign-off is HELD** until this
program's terminal PASS.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-adr-and-census` | ADR-0039 locks the H-G definition + exemption list; binding re-census; cilint H-G guard | passed | [PASS](milestone-0-adr-and-census/qa/milestone-qa.md) + HS-1 approved 2026-06-20 |
| 1 | `milestone-1-category-a-constants` | Typed status constants in `controlleddocuments/domain/resolution.go` (no SQL) | passed | [PASS](milestone-1-category-a-constants/qa/milestone-qa.md) + HS-1 approved 2026-06-20 |
| 2 | `milestone-2-category-b-read-ports` | 9 clean foreign point-reads → owner-published read-ports (parity-before-delete) | passed | [PASS](milestone-2-category-b-read-ports/qa/milestone-qa.md) (validator 2026-06-21; F2.1–F2.4 + F2.5 HS-4 fix) + HS-1 approved 2026-06-21 |
| 3 | `milestone-3-category-c-membership-view` | iam publishes active-membership view; CD list/CanRead + approval (H-PRE-1) consume it | passed | [PASS](milestone-3-category-c-membership-view/qa/milestone-qa.md) (validator 2026-06-21; F3.1 view + F3.2 CD C1+C2 + F3.3 approval C3) + HS-1 approved 2026-06-21 |
| 4 | `milestone-4-search-visibility-contract` | search consumes a CD-published visibility contract (risk-isolated, last) | validator PASS — awaiting HS-1 | F4.1/F4.2/F4.3 closed; views 0243+0244; reader.go consumes 3 views; H-G ledger drained EMPTY; validator PASS 2026-06-21 (`qa/milestone-qa.md`) |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open).
The **Gate result** column links the milestone-validator's verdict (`qa/milestone-qa.md`); `passed` requires
a validator **PASS** *and* operator HS-1 approval.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| 2026-06-20 (M0/F0.2) | HS-6 | Binding census widen surfaced sites beyond the §5 inventory: **N1** (`documents/application/fillin_service.go:225` → `templates_template_version`) and **X1–X8** (auth/audit/platform reads the §2 Non-Goal called "already ported" but which are still raw SQL). | **Operator-ruled 2026-06-20:** 1a — fold **N1 into M2** (one templates read-port); 2a — resolve **X1–X8** via principled **ADR-0039 D3(d)** platform append-sink / **D3(e)** parent-ADR-dispositioned auth / **D3(f)** worker-layer exemptions, enumerated in the F0.3 guard allowlist. ADR-0039 amended; mission §5 row 16 + §2 updated. See `milestone-0-adr-and-census/f0.2-binding-census/hs-6-scope-decision.md`. |

## Program close-out / reconciliation

Fill in only when the last milestone has passed:

- [ ] Every planned feature has a complete evidence row.
- [ ] Zero unplanned scope (anything added is recorded with rationale).
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] **Terminal acceptance:** main session re-runs the F5.1 10-dimension re-audit from clean state →
      captures `wiki/backend/_artifacts/architecture-re-audit-<date>-post-boundary-hardening.md` → **dispatch
      `mission-validator`** to judge it against `mission.md` §8 and write `qa/mission-validation.md`. (This
      line is the durable trigger for the terminal gate — do not skip it once this planning session is gone.)
- [ ] Terminal acceptance passed — link `qa/mission-validation.md`.
- [ ] Parent `grade-a-completion` Grade-A sign-off unblocked and presented to operator.
- [ ] Operator sign-off: <date / name>
