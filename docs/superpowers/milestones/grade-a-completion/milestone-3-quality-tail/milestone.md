# Milestone 3 — Code-quality & dead-code tail

> **Program:** grade-a-completion  ·  **Governing spec:** `../mission.md`
> **Status:** Spec (drafting) — awaiting operator approval before F3.1 starts
> **Authored:** 2026-06-16 — *before any feature in this milestone began.*

> This file is a **spec**, authored up front. It says **what** this milestone is,
> **which features** it contains, **what each feature implements**, and **what gets
> validated**. It contains **no execution steps** — the "how" of each feature lives in
> that feature's `plan.md`. The end-of-milestone QA (`qa/milestone-qa.md`) validates
> the milestone against *this* document.

## Objective

Close the 6 skeptic-confirmed code-quality / dead-code findings in mission §5 (E1–E6) and lift the
**code-quality** and **legacy/dead-code** dimensions of the F5.1 re-audit toward their terminal A−
bar at mission §8. One finding (**E1 — `IAMUserOptions` never wired**) is **functional**: after this
milestone, placeholder user-type lookups return real IAM user data instead of an empty list. The
remaining five remove type-safety holes (E2 `fanoutClient any`, E3 dead `userID` param), a heavy-read
discard (E4 `_ = snap`), dead exported keys (E5), and the Minor tail (E6 — duplicate constructors,
hardcoded PT string, 8× `tenantIDFromRequest` duplication, SHA-1 dedup, etc.).

The consumer of this work is the next developer reading these packages: every changed seam expresses
its real dependency in its real type, no dead parameter lies about a contract, no exported key
pretends to be used. The consumer of **E1** specifically is the document-fillin flow that asks the
placeholder catalog for a user-type option list — it should receive non-empty data backed by IAM.

**Bar (re-measured at close):** the F5.1 §6 code-quality + legacy/dead-code pass-bar — the six
findings are **fixed at their root**, not symptom-patched, with runtime proof for E1 (placeholder
options list non-empty against a real IAM read) and build/compile + test proof for E2–E5 (typed
parameter compiles, dead param removed without losing authz scope, `Pin` no longer fetches the
snapshot blob, dead keys gone with zero production refs). For E6, every Minor enumerated in report
§7 is either **closed** in scope or carries a **written defer trigger** owned by name — silence
counts as a fail.

## Appetite + rabbit holes

**Appetite:** 5 features (E1–E6 → F3.1–F3.5), one focused session each; pragmatic A− code-quality
bar — drive-by repair within the touched seams, not a repo-wide quality sweep. F3.5 (Minor sweep)
is **boxed**: the report §7 list is the universe; nothing outside it is in scope here.

**Rabbit holes (do NOT chase in this milestone):**
- **Repo-wide static-analysis cleanup** (golangci-lint full pass, every `any`, every dead param,
  every unused export). Out of scope — this milestone closes the six confirmed findings + the §7
  Minor list. *Reason:* mission Non-Goals; HS-6 trigger.
- **Refactoring `Pin` / freeze pipeline beyond the snapshot-discard fix.** Out of scope — F3.3
  removes the `_ = snap` discard at the cited line; nothing else changes in the freeze flow.
  *Reason:* mission Non-Goals; preserves M2 close-state.
- **IAM port redesign / new IAM-owned port.** Out of scope — F3.1 wires the **existing**
  `IAMUserOptions` dependency at the composition root; no new port, no new IAM API. *Reason:* HS-2
  boundary; the IAM port work is M4 / F4.2 + F4.3, locked-order-last.
- **Touching M4 module-boundary sites (C1–C4).** Out of scope — those have their own milestone and
  sequencing isolates port work last. *Reason:* mission §3 D3 sequencing.
- **Re-running M0/M1/M2 gate work** — only regression against those, not redo.

## Features

| Feature id | Slug / folder | What to implement | What to validate (acceptance) |
|------------|---------------|-------------------|-------------------------------|
| F3.1 | `f3.1-wire-iam-user-options` | Wire the **existing** `IAMUserOptions` dependency at the composition root (`apps/api/cmd/metaldocs-api/main.go` — path drifted from mission §5 `cmd/metaldocs-api/main.go:413`; re-anchor at feature start) so the documents-module placeholder adapter at `internal/modules/documents/module.go:96` (`newPlaceholderOptionsIAMAdapter`) is fed a non-nil reader backed by the IAM module's user-options query. Consumer: the document-fillin placeholder catalog asks for a user-type option list. | Runtime proof: integration test (real IAM read or labeled in-process IAM fake against the real adapter contract per mission §8) — placeholder-options for the user-type returns a **non-empty** list when the tenant has ≥1 IAM user; empty-tenant case returns empty without nil-deref. `grep -RIn 'IAMUserOptions' apps/api/cmd/metaldocs-api/` shows the wiring line; whole-repo `go test ./...` green. |
| F3.2 | `f3.2-type-safety-deadparam` | Give `NewFreezeService` a typed `FanoutClient` parameter at `internal/modules/documents/application/freeze_service.go:77` (replace `fanoutClient any` + the `fc, _ := fanoutClient.(FanoutClient)` assertion at :79). Remove the dead `userID` from `ListDocumentComments` at `internal/modules/documents/application/service.go:433` — but **first verify** no authz scope was being computed from that parameter (would be an HS-2 boundary signal). Consumer: every caller of `NewFreezeService` (gets a real type, no runtime cast) + every caller of `ListDocumentComments` (sheds a lying param). | Build compiles with the typed param (no `any`); the type assertion at :79 is gone; `ListDocumentComments` signature has no `userID`; **authz-scope check recorded in `evidence.md`** — confirmation that no authz path consumed the removed `userID`; whole-repo `go test ./...` green; no caller compiles against the old shape. |
| F3.3 | `f3.3-snapshot-read` | At `internal/modules/documents/application/freeze_service.go:194` (re-anchor at feature start — `_ = snap` confirmed at HEAD), stop fetching-then-discarding the snapshot blob in `Pin`. Either drop the read entirely (if unused) or rewrite the call to read only what is used. Consumer: the `Pin` path — pays only for the I/O it consumes. | `Pin` reads only what it uses (`_ = snap` discard absent); `grep -n '_ = snap' internal/modules/documents/application/freeze_service.go` returns 0; whole-repo `go test ./...` green; runtime proof: a pin operation runs without the discarded read (test or trace evidence). |
| F3.4 | `f3.4-dead-keys` | Remove the unused `TemplateDocxKey` / `TemplateSchemaKey` exports at `internal/platform/objectstore/template_keys.go:5,9` **or** align to the live key schema if a real consumer is discovered during the read-pass (recorded in `evidence.md`). Consumer: the objectstore key catalog — the exports declare a contract; if no caller exists, the contract is a lie. | `grep -RIn 'TemplateDocxKey\|TemplateSchemaKey' --include='*.go'` returns 0 production refs after the change (test/fixture refs explicitly excluded and named); build + `go test ./...` green; if aligned-not-removed, the live key schema is cited. |
| F3.5 | `f3.5-minor-sweep` | Close report §7 Minors in scope **or** record each as a bounded defer with a named trigger and owner. Boxed list (the §7 universe): duplicate constructors, hardcoded `"pt"` / PT string, 8× `tenantIDFromRequest` duplication, SHA-1 dedup site. Consumer: future readers of the touched seams — every Minor is either gone or has a written reason it survives. | A close-out table in `evidence.md` rows = report §7 Minors; each row is **closed** (cite + commit) or **deferred** (trigger + owner). Nothing on the §7 list is silently skipped. Whole-repo `go test ./...` green. |

For each feature, "what to validate" is objectively checkable — a named command, a grep that returns
0, a test that passes, a runtime artifact. No "works" / "looks right".

## Milestone validation definition

The close gate is run by the **`milestone-validator` subagent** (separation of powers — it judges
and writes `qa/milestone-qa.md`; the main session flips status only on its PASS), following the
binding C1–C7 checklist in `.claude/skills/milestone/references/milestone-end-validation.md`. What
that gate enforces for this milestone:

1. **Per-feature acceptance** — every F3.x above meets its declared "what to validate", and each
   feature's consumer contract (`spec.md`) was honored. **F3.1 runtime proof must be real (or
   explicitly labeled in-process IAM fake against the real adapter contract per mission §8)** — a
   pure mock that bypasses the adapter is a FAIL.
2. **Workflow-class QA checklist** — [`wiki/quality/backend-api-qa-checklist.md`](../../../../wiki/quality/backend-api-qa-checklist.md)
   with a **code-quality lens**: typed dependencies at constructor seams, no `any` smuggling at
   public application boundaries, no dead parameters, no dead exports, no fetch-then-discard reads.
3. **Regression** — whole-repo `go test ./...` green; M0 / M1 / M2 gates still pass (re-run
   `go test ./...` and re-grep the H-D class from report §6 — must remain 0; re-run M2
   observability runtime smoke — composition root injection unchanged).
4. **Quality-bar / root-cause check** — every E1–E5 fix lands at the **right seam** (E1 at the
   composition root, not via a per-call default; E2 at the constructor signature, not via a wrapper
   cast at every call site; E3 at the method signature, not via passing a dummy through; E4 at the
   read site, not via swallowing the error; E5 by removing the export, not by adding a "deprecated"
   comment). F3.5 has a complete §7 row table — no silent skips.
5. **No unplanned scope** — anything implemented beyond F3.1–F3.5 (especially anything touching M4
   module-boundary sites C1–C4) is recorded with rationale or rolled back. Rabbit-hole list above
   is the scope-drift baseline.

## Dependencies & constraints

- **Depends on:** M0 passed (HS-1 approved 2026-06-15), M1 passed (HS-1 approved 2026-06-15), M2
  passed (HS-1 approved 2026-06-16). HEAD includes M0+M1+M2 close commits; mission §5 `file:line`
  anchors **must be re-verified at feature start** — confirmed drift: `cmd/metaldocs-api/main.go`
  → `apps/api/cmd/metaldocs-api/main.go`. E1–E5 sites otherwise present at the cited symbol /
  pattern at HEAD `22a80208`.
- **Quality goals (top 3, ranked):**
  1. **Root-cause fixes at the right seam** (composition root, constructor, signature — not
     wrapper / default / per-call workaround) > 2. **Surgical scope** (only the cited sites and the
     §7 Minor list; no drive-by sweeps) > 3. **Type safety / contract honesty** (typed parameters,
     no dead params, no dead exports).
- **Architectural constraints:**
  - **No new IAM port and no IAM API change.** F3.1 wires the **existing** dependency; new port
    work is M4. Inventing a new port here is an HS-2 trigger.
  - **F3.2 authz-scope check is mandatory** before removing `userID` from `ListDocumentComments` —
    silently dropping a parameter that was scoping access would be a security regression. Recorded
    in `evidence.md`.
  - **Drive-by repair, not big-bang.** Per CLAUDE.md §4 test-framework gate + §5.3 surgical-change
    rule — touched test files migrate to the canonical framework; adjacent untouched tests stay.
  - **F3.5 §7 Minor list is closed-universe.** Nothing added; nothing silently dropped.
  - **No FE work.** Mission Non-Goals — no codegen, no frontend, no UI.
  - **Skill routing:** backend wiring/composition → `metaldocs-backend-api`; new tests must use
    canonical framework per CLAUDE.md §4 test-framework hard gate (testdb for DB integration,
    handler-test framework for HTTP, table-driven + in-memory fakes for app/domain); module-wiki
    sync after structural change → `metaldocs-module-doc-sync`.
  - **Root-cause-over-symptom-patch policy** (`memory/authz-root-cause-over-symptom.md`) extends
    here — bind C5/C6.
- **Risks (named, with owner/mitigation):**
  - **R1 — F3.2 silently drops a parameter that was scoping authz.** *Mitigation:* mandatory
    authz-scope read recorded in `evidence.md` before deletion; if the `userID` is consumed by any
    authz path (PEP/PDP, scope check, row filter), HS-2 stop and replan as an authz-boundary
    feature. *Owner:* F3.2 author at spec time.
  - **R2 — F3.4 removes an export that has an undiscovered consumer (e.g. via reflection / string
    key build).** *Mitigation:* `grep` includes string-form matches; build + test pass before
    commit; if a consumer surfaces post-merge, the fix is a re-add (small blast radius). *Owner:*
    F3.4 author.
  - **R3 — F3.1 wiring exposes a latent IAM contract mismatch** (e.g. tenant scoping mismatch,
    nil-tenant case). *Mitigation:* integration test exercises empty-tenant + populated-tenant; on
    a real IAM contract gap, HS-2 — IAM module change is its own boundary. *Owner:* F3.1 author.
  - **R4 — F3.5 scope creep into a repo-wide quality sweep.** *Mitigation:* the §7 Minor list is
    the closed universe; the close-out table is row-for-row §7; deviations are recorded as defers
    with triggers, not absorbed. *Owner:* F3.5 author.
  - **R5 — Anchor drift bites at F3.3 / F3.4** (line numbers shift after intervening edits).
    *Mitigation:* re-anchor by symbol/pattern (not line) at feature start; `spec.md` records the
    re-anchored line. *Owner:* feature author at spec time.

## Applicable hard-stops

| ID | What would trip it here |
|----|--------------------------|
| HS-1 | This milestone's boundary — operator review gate after the validator PASS; no next milestone (M4) / no merge without approval. |
| HS-2 | If F3.1 implies a new IAM port or an IAM API change (rather than wiring the existing dependency); or if F3.2's authz-scope check reveals `userID` was used for authz (parameter deletion would be a security regression — promote to an authz-boundary feature instead); or if F3.3 implies redesigning the freeze pipeline beyond the discard-fix; or if F3.5 grows into a repo-wide static-analysis cleanup — stop, report the boundary + minimum prerequisite plan, do not symptom-patch. |
| HS-3 | If a prerequisite boundary fails (build / runnable / auth-session / route / contract truth — e.g. composition-root reshuffle breaks startup; F3.4 removal triggers a build error from an undiscovered consumer) — repair via `runtime-contract-prereq`, rerun the failed checkpoint, resume the feature. |
| HS-4 | If `milestone-validator` returns FAIL (symptom-patch, unmet acceptance, root-cause check fails, F3.5 §7 table incomplete) — open the named fix feature, re-run its lifecycle, re-dispatch the validator. |
| HS-6 | If a fix uncovers a code-quality defect F5.1 missed, or scope drifts off these five features (e.g. an M4 boundary finding surfaces and an attempt is made to absorb it) — stop, surface the deviation, replan before continuing. |
