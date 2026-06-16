# Milestone 4 — Module boundaries / systemic ports  *(LAST — risk-isolation)*

> **Program:** grade-a-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Spec (drafting) — awaiting operator approval before F4.1 starts
> **Authored:** 2026-06-16 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

Close the 4 skeptic-confirmed module-boundary findings in mission §5 (C1–C4) and drive the
**module-boundaries/DDD** dimension of the F5.1 re-audit to **≥ A−**, with the **H-G class re-grepped
to 0**. After this milestone: no module issues raw SQL against another module's owned table, and no
module hardcodes another module's domain-state literal. The two surviving cross-module SQL reaches in
`security` (`iam_user_roles`, `iam_users`) are routed through **IAM-owned ports** — extending the
port surface that already exists in `iamdomain` (`UserDisplayNameReader`, `TenantUserReader`, off-tx,
H-PRE-1) rather than inventing a new pattern. This **retires the M4 accepted MfaCoverage defer**
recorded in the prior wave.

The consumer of this work is the **architecture re-audit's module-boundaries auditor** and the next
developer reading `security` and `documents`: every cross-module data need is expressed as a typed
port call into the owning module, not a raw JOIN that silently couples to another module's schema;
every domain-state comparison uses the owning module's exported constant, not a string literal that
drifts. Sequenced **last** (mission §3 D3) so port work cannot regress an already-lifted grade —
H-G is re-grepped only after all four features land.

**Bar (re-measured at close):** the F5.1 §6 module-boundaries pass-bar —
**(1)** the **H-G grep commands from the re-audit report §6 return 0** (no hardcoded domain-state
literal; no cross-module owned-table reach without a port); **(2)** module-boundaries indicatively
**≥ A−** with cited evidence; **(3)** each port read proven at parity with the SQL it replaces by a
**live test** (real DB), not a fixture-only assertion; **(4)** whole-repo `go test ./...` green and
M0/M1/M2/M3 gates non-regressed. Fixture-only parity for F4.2/F4.3 is a **FAIL** (these replace live
SQL — parity must be proven against a live read).

## Appetite + rabbit holes

**Appetite:** 4 features (C1–C4 → F4.1–F4.4), one focused session each; pragmatic A− module-boundaries
bar — close the four cited boundary sites + re-grep the H-G class to 0. **Extend** the existing
`iamdomain` port pattern; do **not** redesign IAM's API. Port reads stay **off** any lock-holding
atomic tx (H-PRE-1). This is the **last** milestone — its close (after HS-1) triggers the mission's
terminal re-audit + `mission-validator` gate.

**Rabbit holes (do NOT chase in this milestone):**
- **A general IAM-API / IAM-port redesign.** Out of scope — F4.2/F4.3 add **two narrow reader
  ports** (role-membership reader; mfa/user reader) mirroring the existing `TenantUserReader` /
  `UserDisplayNameReader` shape with Noop defaults. Reshaping IAM's public surface, consolidating the
  port family, or generalizing a "query gateway" is an **HS-2** boundary. *Reason:* mission §3 D2 +
  HS-2; the appetite is two reader ports, not an IAM redesign.
- **Migrating every remaining cross-module reference in the repo.** Out of scope — the closed
  universe is exactly C1–C4 + whatever the report §6 H-G grep flags. A repo-wide DDD-purity sweep is
  not in this milestone. *Reason:* mission Non-Goals; HS-6 trigger.
- **Re-grading the already-lifted dimensions (contract-api, composition, code-quality).** Out of
  scope — those are M1/M2/M3 closed-state; M4 only regresses-tests them, never reopens. *Reason:*
  ports-last sequencing exists precisely to not disturb them.
- **Schema / migration changes for the new ports.** Out of scope — the ports read **existing**
  IAM-owned tables via IAM's own repository; no new table, column, or migration. *Reason:* mission
  Non-Goals (no schema redesign); the legacy teardown is already done (M4b `071931c9`).
- **Moving the `templates/domain.Placeholder` type or restructuring the templates↔documents seam**
  beyond F4.4's decision. Out of scope — F4.4 **decides** whether the cross-package type is a
  legitimate port-typed dependency or a leak, and routes via a port **only if** it is a leak; it does
  not pre-emptively re-layer the modules. *Reason:* avoid a redesign masquerading as a boundary fix
  (HS-2).
- **Re-running M0/M1/M2/M3 gate work** — only regression against those, not redo.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F4.1 | `f4.1-published-constant` | Replace the hardcoded `overrideStatus := "published"` at `internal/modules/documents/application/service.go:283` (re-anchor by symbol/pattern at feature start) with the **templates-module exported constant** `templatesdomain.VersionStatusPublished` (confirm the exact exported identifier in `internal/modules/templates/domain/` at feature start; if no such constant exists, **add it in the templates module** and import it — do not invent a `documents`-local copy). Consumer: the document version-status comparison — reads the *owning* module's published-state constant, not a literal that drifts if templates renames the state. | **H-G grep for hardcoded status literals (report §6 command) returns 0** at this site; `grep -n '"published"' internal/modules/documents/application/service.go` returns 0 (or only non-status occurrences, named in `evidence.md`); the constant resolves to the same wire value (`"published"`) — proven by a test asserting `templatesdomain.VersionStatusPublished == "published"` and the documents path still behaves identically; whole-repo `go test ./...` green. |
| F4.2 | `f4.2-iam-role-port` | Add a narrow **IAM-owned role-membership reader port** (mirroring `iamdomain.TenantUserReader` — interface in `iamdomain`, Postgres impl in `iam`, `Noop` default, **off-tx** read, H-PRE-1) that returns the admin-role user set for a tenant; wire it into `security.Repository` and rewrite `ListOffHoursAdminActions` (`internal/modules/security/infrastructure/postgres/repository.go:328`, JOIN at `:345` — re-anchor at feature start) to resolve admin-role membership via the port instead of `JOIN metaldocs.iam_user_roles`. Consumer: `ListOffHoursAdminActions` — gets its admin-role user set from IAM's port, never JOINs IAM's owned table. | `grep -RIn 'iam_user_roles' internal/modules/security/ --include='*.go'` (excluding `_test.go`, named) returns 0; the port interface lives in `iamdomain`, the impl in `iam`, with a `Noop` default in `NewRepository`; **live test** (real DB, `testdb` framework) proves the off-hours-admin-actions result is at **parity** with the prior JOIN on a seeded fixture (same OffHoursAction set); the port read is **not** inside any lock-holding tx (H-PRE-1 confirmed in `evidence.md`); whole-repo `go test ./...` green. |
| F4.3 | `f4.3-mfa-coverage-port` | Add a narrow **IAM-owned reader port** serving the per-tenant user set `MfaCoverage` needs (mirroring the existing `iamdomain` reader pattern; Noop default; off-tx) and rewrite `MfaCoverage` (`internal/modules/security/infrastructure/postgres/repository.go:63`, `FROM metaldocs.iam_users` at `:67` — re-anchor at feature start) to read via the port instead of `FROM metaldocs.iam_users`. This **retires the M4 accepted MfaCoverage defer** (the last `iam_users` reach in `security`). Consumer: `MfaCoverage` — gets its user set from IAM's port; `security` reads **no** `iam_users`. | `grep -RIn 'iam_users' internal/modules/security/ --include='*.go'` (excluding `_test.go` + the boundary-explaining comment, named) returns 0 SQL reaches; **live test** (real DB, `testdb`) proves the `MfaCoverage` metric value is at **parity** with the prior `FROM iam_users` read on a seeded fixture; off-tx confirmed (H-PRE-1); the prior accepted defer is cited as retired in `evidence.md`; whole-repo `go test ./...` green. |
| F4.4 | `f4.4-placeholder-seam` | Resolve the cross-module `templatesdomain.Placeholder` type used in the `documents` repository seam (`internal/modules/documents/repository/repository.go:112` — `CreateDocumentTx(..., requiredPlaceholders []templatesdomain.Placeholder)`; re-anchor at feature start). **Decide and document** whether this is a **legitimate port-typed dependency** (templates is the owning module; documents depends on its published domain type by design — analogous to a shared value object) or a **leak** (documents reaches into templates' internals). If legitimate → record the boundary decision (an ADR under `wiki/decisions/` if durable) and leave it. If a leak → route the dependency through a port / a documents-owned input type. Consumer: the `documents` repository's create path — depends on templates only through a sanctioned seam. | A written **boundary decision** in `evidence.md` (+ ADR link if durable) stating legitimate-or-leak with the rule applied; if leak, the dependency goes via a port/owned type and `grep` shows no direct `templatesdomain` internal reach on the changed seam; if legitimate, the decision cites *why* (owning module, published type, no internals reached) so the re-audit auditor can confirm it is not an H-G site; whole-repo `go test ./...` green. |

For each feature, "what to validate" is objectively checkable — a named command, a grep that returns
0, a live parity test, a written decision. No "works" / "looks right". F4.2 and F4.3 parity proof
**must be a live DB read** (they replace live SQL); fixture-only is a FAIL.

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges
and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the
binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What
that gate enforces for this milestone:

1. **Per-feature acceptance** — every F4.x above meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored (the consumer — `ListOffHoursAdminActions`,
   `MfaCoverage`, the version-status comparison, the documents create seam — drives the port/constant
   shape, not the reverse). **F4.2 / F4.3 parity proof must be a live DB read** (real `testdb`
   integration), not a fixture-only mock — they replace live SQL; a mock-only parity claim is a FAIL.
2. **Workflow-class QA checklist** — [`wiki/quality/backend-api-qa-checklist.md`](../../../../wiki/quality/backend-api-qa-checklist.md)
   with a **module-boundaries lens**: no cross-module raw SQL against another module's owned table;
   no hardcoded foreign domain-state literal; cross-module dependencies expressed as typed ports
   owned by the producing module; new port reads off any lock-holding tx (H-PRE-1).
3. **Regression** — whole-repo `go test ./...` green; **M0 / M1 / M2 / M3 gates still pass** — re-run
   `go test ./...`; re-grep the **H-D class** from report §6 (must remain **0** — M1 close-state);
   re-run the M2 observability runtime smoke (composition root injection unchanged); confirm M3's
   E1–E6 sites are untouched (no quality regression). Port wiring must not disturb the composition
   root's M2/M3 state beyond adding the two reader dependencies.
4. **Quality-bar / root-cause check** — the **H-G grep commands from report §6 return 0 after all
   four features** (the bar is re-measured, not asserted); each fix lands at the **right seam**
   (F4.1 at the comparison via the owning module's constant — not a `documents`-local duplicate
   constant; F4.2/F4.3 via an **IAM-owned** port interface in `iamdomain` with the impl in `iam` —
   not a `security`-local helper that still JOINs, and not by smuggling the JOIN into a view or a
   raw-SQL string built elsewhere; F4.4 by a **documented boundary decision**, not by silently
   leaving the reach). Module-boundaries indicatively **≥ A−** with cited evidence.
5. **No unplanned scope** — anything implemented beyond F4.1–F4.4 (especially an IAM-API redesign, a
   repo-wide DDD sweep, or reopening M1/M2/M3 dimensions) is recorded with rationale or rolled back.
   The rabbit-hole list above is the scope-drift baseline.

## Dependencies & constraints

- **Depends on:** M0 passed (HS-1 approved 2026-06-15), M1 passed (HS-1 approved 2026-06-15), M2
  passed (HS-1 approved 2026-06-16), M3 passed (HS-1 approved 2026-06-16). HEAD includes all four
  close commits. Mission §5 `file:line` anchors **drifted** — root is `internal/modules/...` (not
  `apps/api/internal/...`); re-verified at authoring time:
  - C1 → `internal/modules/documents/application/service.go:283` (`overrideStatus := "published"`).
  - C2 → `internal/modules/security/infrastructure/postgres/repository.go:328` (`ListOffHoursAdminActions`), JOIN `iam_user_roles` at `:345`.
  - C3 → `internal/modules/security/infrastructure/postgres/repository.go:63` (`MfaCoverage`), `FROM iam_users` at `:67`.
  - C4 → `internal/modules/documents/repository/repository.go:112` (`CreateDocumentTx(..., []templatesdomain.Placeholder)`).
  - **Re-anchor by symbol/pattern at each feature start** (lines shift after intervening edits).
- **Existing port surface to extend (not redesign):** `security.Repository` already holds
  `iamdomain.UserDisplayNameReader` + `iamdomain.TenantUserReader` (off-tx, Noop defaults,
  `NewRepository` injection). F4.2 adds a **role-membership reader**; F4.3 adds a **user/mfa reader**.
  Mirror the existing interface+Noop+injection shape exactly.
- **Quality goals (top 3, ranked):**
  1. **Correct module ownership** (cross-module need served by an **IAM-owned** typed port; foreign
     domain-state via the **owning module's** constant) > 2. **Parity, proven live** (the port read
     returns exactly what the replaced SQL returned, proven by a real-DB test) > 3. **Surgical scope**
     (only C1–C4 + the report §6 H-G grep universe; no IAM redesign, no DDD sweep).
- **Architectural constraints:**
  - **No IAM-API redesign.** Two narrow reader ports mirroring the existing pattern only. A broader
    IAM surface change is an **HS-2** boundary (esp. F4.2/F4.3 growing into a shared IAM-API redesign,
    per mission HS-2).
  - **H-PRE-1 advisory-lock hazard:** new port reads stay **off** any lock-holding atomic tx
    (`memory/advisory-lock-deadlock-constraint.md`). Confirmed per feature in `evidence.md`.
  - **Ports-last sequencing (mission §3 D3):** M4 must not regress an already-lifted grade — F4.x
    touch only the boundary seams + the composition-root wiring for the two new ports; no change to
    M1 contract surface, M2 observability wiring, or M3 quality sites.
  - **No schema / migration changes.** Ports read existing IAM-owned tables via IAM's own repository.
  - **F4.1 uses the owning module's constant**, not a `documents`-local duplicate (a duplicate would
    be a split-brain, not a boundary fix).
  - **Drive-by repair, not big-bang.** Per CLAUDE.md §4 test-framework gate + §5.3 surgical-change
    rule — touched test files migrate to the canonical framework (`testdb` for the new live parity
    tests); adjacent untouched tests stay.
  - **Skill routing:** backend module/port wiring → `metaldocs-backend-api`; new live DB parity tests
    → `testdb` framework per CLAUDE.md §4 hard gate; DB read only if a port needs a new query →
    `metaldocs-database`; module-wiki sync after the structural port change →
    `metaldocs-module-doc-sync`; prereq repair → `runtime-contract-prereq`.
  - **Root-cause-over-symptom-patch policy** (`memory/authz-root-cause-over-symptom.md`) binds
    C5/C6 — no masking the H-G grep (e.g. moving the JOIN into a view or a string-built query to
    dodge the grep) counts as a pass.
  - **No FE work.** Mission Non-Goals — no codegen, no frontend.
- **Risks (named, with owner/mitigation):**
  - **R1 — F4.2/F4.3 port grows into an IAM-API redesign.** *Mitigation:* hard appetite of two
    narrow reader ports mirroring `TenantUserReader`; if the consumer needs more than a tenant-scoped
    id/role set (e.g. cross-tenant, write, or a new aggregate), **HS-2** stop and report the boundary.
    *Owner:* F4.2/F4.3 author at spec time.
  - **R2 — Port read returns a different set than the JOIN** (e.g. the JOIN dropped users lacking a
    membership row; the port includes them, or vice-versa — silent behavior change). *Mitigation:*
    mandatory **live parity test** on a seeded fixture asserting the exact result set equality before
    and after; the prior JOIN semantics (incl. INNER-JOIN drop behavior, see repo comment at
    `:269-270`) are reproduced or the difference is documented + operator-confirmed. *Owner:*
    F4.2/F4.3 author.
  - **R3 — H-PRE-1 deadlock:** a new port read placed inside a lock-holding tx. *Mitigation:* port
    reads are off-tx by construction (pool read, like the existing `members`/`displayNames` ports);
    `evidence.md` records the off-tx confirmation per feature. *Owner:* F4.2/F4.3 author.
  - **R4 — F4.1 introduces a `documents`-local constant** instead of importing templates' export
    (split-brain). *Mitigation:* F4.1 acceptance requires the constant resolve from
    `templatesdomain`; if templates lacks the export, it is **added in templates** and imported —
    never duplicated in documents. *Owner:* F4.1 author.
  - **R5 — F4.4 turns a legitimate dependency into an over-engineered port** (a redesign disguised
    as a boundary fix). *Mitigation:* F4.4 is **decide-first** — only route via a port if the read is
    a genuine leak; a legitimate published-type dependency is documented and left. Operator/validator
    can challenge the decision. *Owner:* F4.4 author.
  - **R6 — Anchor drift** (line numbers shifted since authoring). *Mitigation:* re-anchor by
    symbol/pattern at feature start; `spec.md` records the re-anchored line. *Owner:* feature author.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | This milestone's boundary — operator review gate after the validator PASS. M4 is the **last** milestone: HS-1 approval here unblocks the **terminal acceptance** (re-run F5.1 re-audit + `mission-validator`), not a next milestone. No merge without approval. |
| HS-2 | If F4.2/F4.3 imply a shared **IAM-API redesign** rather than two narrow reader ports mirroring the existing pattern; or if F4.4 implies re-layering the templates↔documents seam beyond a single port/decision — stop, report the boundary + minimum prerequisite plan, do not symptom-patch. |
| HS-3 | If a prerequisite boundary fails (build / runnable / auth-session / route / contract truth — e.g. wiring the new ports at the composition root breaks startup; a port impl breaks a `security` query) — repair via `runtime-contract-prereq`, rerun the failed checkpoint, resume the feature. |
| HS-4 | If `milestone-validator` returns FAIL (H-G grep not 0, fixture-only parity for F4.2/F4.3, split-brain constant in F4.1, undocumented F4.4 decision, symptom-patch, scope drift) — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-5 | *(Program-terminal, raised after M4 HS-1.)* If the terminal re-audit / `mission-validator` misses the §8 bar (a dimension < A−, a surviving/new confirmed Critical/Major, H-D ≠ 0, or **H-G ≠ 0**) — bounded remediation micro-milestone, re-run re-audit, re-dispatch; operator decides continue vs replan. |
| HS-6 | If a fix uncovers a module-boundary defect F5.1 missed, or scope drifts off these four features (e.g. another cross-module JOIN surfaces and an attempt is made to absorb it beyond the report §6 H-G universe) — stop, surface the deviation, replan before continuing. |
