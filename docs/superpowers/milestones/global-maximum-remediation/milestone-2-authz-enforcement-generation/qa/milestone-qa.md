# Milestone 2 — Validation Verdict (C1–C7) — RE-VALIDATION after live-QA correction

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4, binding; HS-7) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-03 (RE-VALIDATION) · **Verdict:** see C7 → **PASS**.
> This supersedes the prior PASS. That verdict accepted a **bounded defer** (F2.1 integration drives
> compile-verified / live-run deferred). The operator required an actual live QA; running it caught a
> real defect (the deferred drives were **non-functional** — drove the full application stack and died
> at fixture setup before reaching the tripwire arm) and corrected it (rewrite to the proven sibling
> DB-arm idiom + live RED→GREEN). Only two artifacts changed since the prior PASS: the rewritten test
> and its evidence (commit `5e5b50e3`). The tripwire generation core is **byte-identical** to the prior
> PASS and re-spot-checked green. Validator judged from a clean tree, **independently reproduced the
> live RED→GREEN** against the running container, and left the tree + dev DB clean.

## Inputs loaded (none missing)

`milestone.md`, `validation-contract.md` (D4), `f2.1/{spec,plan,evidence}.md`, `f2.2/{spec,plan,evidence}.md`,
program `README.md`, governing `mission.md` (linked). Aggregate diff since prior milestone close
`git diff 1a86f419..HEAD`. **Focused re-review** on the changed artifact: `git show 5e5b50e3`
(2 files: `tests/integration/documents/tripwire_documents_test.go` −185/+55 net rewrite,
`f2.1/evidence.md` reworked "Integration drives" section). Core diff `git diff f170e4e6..5e5b50e3 --
internal/platform/tripwire/ scripts/api-lint/tripwire_arm_rules.go db/migrations/` = **empty** (core
untouched by the correction).

## C1 — Spec & plan conformance (per feature)

Both feature `spec.md`/`plan.md`/`evidence.md` exist and are execution-shaped. Approval lines filled
(2026-07-03); binding D4 contract committed (`4a815be9`) before first code (`70c2803b`). Consumer
contract honored — verified from source.

**Deviation (documented, sound, not a dodge).** Contract §1.5.b named the drive
`TestTripwire_ForceReleaseWritesDocumentRow` (application-stack-driven). The live QA proved that
construction **non-functional** (tier-2 `document.create` denial + `process_area_code_snapshot`
NULL-scan at fixture setup — died before the arm). The drives were renamed/rewritten to
`TestTripwire_DocumentsUpdate_{MembershipManageArm,DocumentObsoleteArm}`, driving the guarded
`documents` UPDATE **directly** (proven sibling idiom, `tests/integration/templates/tripwire_caps_test.go`).
The deviation carries a full written rationale in `evidence.md` (the "Correction" callout) → satisfies
C1.7. This is a correctness upgrade, not scope drift: the test now tests the **DB arm** (the thing
under contract) instead of the application authz grant machinery (not the thing under contract).

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1 | ✅ `TripwireArms` == contract §1.2 (entry #6 == `{document.edit, document.obsolete, membership.manage}`, verified line-by-line `arms.go:82-88` vs §1.2/§1.7 row #6); 0271 generated from the map; both lint rules blocking; **live drives now genuinely exercise the arm** | ✅ §1.7 exit criteria met, **defer #1 DISCHARGED** (validator reproduced live RED→GREEN, below) | ✅ additive arm widen only; no tightening (`template.submit` superset retained → M9); no cross-layer graph | `arms.go`, `render.go`, `0271_*.sql`, `tripwire_arm_rules.go`, `tripwire_documents_test.go` |
| F2.2 | ✅ tier-1==tier-2==`membership.manage`; pin targeted to the two routes | ✅ §2.3 (verify + pin RED/GREEN + doc-truth) | ✅ coarse/fine `/approval/` split recorded intentional | `permissions_test.go` pin, ADR 0022/0018 annotations |

**HS-7 clean.** `arms.go` `TripwireArms` == contract §1.2 exactly (entry #6 verified against
`validation-contract.md:30` and `:93`). 0271-vs-0270 delta is only the documents(UPDATE)
`v_required_caps` literal (confirmed via the byte-pinned golden test + PARITY rule, both green). Pinned
in code by `TestTripwireArms_MatchesContractTable`. No divergence → no HS-7.

## C2 — Gates re-run, isolated (validator-run; live half INDEPENDENTLY REPRODUCED, not trusted)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F2.1 | `go vet -tags integration ./tests/integration/documents/` | exit 0 (rewritten drives compile) | ✅ |
| F2.1 (**live GREEN vs 0271**) | ephemeral role `qa_val`, `METALDOCS_DATABASE_URL=… go test -tags integration -run TestTripwire_DocumentsUpdate_ -v -count=1 ./tests/integration/documents/` | `--- PASS: …MembershipManageArm (78.69s)`, `--- PASS: …DocumentObsoleteArm (2.16s)`, `ok … 83.721s` | ✅ |
| F2.1 (**live RED vs 0270**) | `mv 0271_*.sql /tmp/`, re-run same | `--- FAIL: …MembershipManageArm: ErrCapabilityNotAsserted: none of {document.edit} present in asserted_caps on documents (SQLSTATE P0001)` + identical FAIL on `…DocumentObsoleteArm`; then `mv` back → tree byte-clean | ✅ (fails for the right reason: the cap arm, not a fixture crash) |
| F2.1 | `go test ./internal/platform/tripwire/...` | `ok` (golden `RenderMigration()==0271` bytes; parity; determinism) | ✅ |
| F2.1 | `go test ./scripts/api-lint/... -run TripwireArm` | `ok 1.686s` (tripwire-arm rule tests) | ✅ |
| F2.1 | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` (both new rules blocking, clean tree) | ✅ |

**The RED half is the load-bearing new proof.** The *prior* broken drives died at fixture setup with a
different error class (tier-2 denial / NULL-scan). This RED reproduces the exact contract failure mode
— `SQLSTATE P0001 … none of {document.edit}` raised by `enforce_capability_asserted()` on the live
`trg_require_cap_asserted` trigger — proving the test reaches and exercises the cap arm. GREEN vs 0271
proves the widened arm clears it. Byte-identical to the evidence transcript (lines 88-101).

## C3 — Senior review of the aggregate milestone diff (focus: the changed test)

- **The rewrite genuinely exercises the DB arm, not the application stack.** The UPDATE is
  `SET active_session_id = NULL, updated_at = now()` with `status` unchanged. I verified the live BEFORE
  trigger order on `public.documents` (alphabetical): `enforce_snapshot_on_submit_trg` →
  `trg_documents_legal_transition` → `trg_documents_revision_version_monotonic` → `trg_require_cap_asserted`
  (cap fires **last**). Inspecting the three pre-cap function bodies from the live DB:
  `enforce_document_transition` gates on `OLD.status IS DISTINCT FROM NEW.status` (skipped — status
  unchanged); `enforce_revision_version_monotonic` only raises on `NEW.revision_version < OLD` (unchanged
  → passes); `enforce_snapshot_on_submit` only fires for `status IN (under_review,approved,scheduled,published)`
  (seeded doc is `draft`, unchanged → skipped). So a neutral-column update cleanly reaches the cap arm.
  Isolating from the legal-transition machine by holding `status` is **sound**, not a dodge — those
  triggers are out of scope for a cap-arm test, and driving them would reintroduce the exact
  application-fixture fragility that broke the first attempt.
- **The `RowsAffected==1` assertion closes the RLS-false-green hole.** A row hidden by RLS yields a
  0-row UPDATE; a BEFORE trigger never fires on 0 rows, so without this assert an RLS-hidden row would
  masquerade as a pass. `driveGuardedDocumentUpdate` returns an error (→ `t.Fatalf` via `SeedWithCaps`)
  on `n != 1`. Verified against `SeedWithCaps`/`seedWithCaps` in `testdb/fixtures.go` — caps asserted
  tx-locally (`is_local=true`), pool-safe, discarded on commit, and any P0001 surfaces as the test
  failure. `InsertDraftDocument` seeds a real `draft` doc through the canonical template family.
- **No split-brain, no dead code, no feature-cross-break.** Core unchanged since prior PASS
  (`git diff f170e4e6..5e5b50e3` on core = empty); SQL still generated from the one Go map, byte-pinned.
  The sibling `tripwire_caps_test.go` is untouched. The rewrite deletes the non-functional
  application-driver scaffolding (net −130 lines) — dead-code removal, not accumulation.
- **Findings:** none blocking.
- Staff-engineer bar met? ✅ (the correction is exactly what a senior reviewer would demand:
  compile ≠ works; test the layer under contract; fail for the right reason.)

## C4 — Workflow-class QA + regression (backend-api authz / DB-invariant class)

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api authz/DB-invariant) | **pass (no remaining defer)** | Deterministic gates green from clean tree; **live integration now discharged** — RED→GREEN reproduced by the validator against the running `metaldocs-postgres` container. The one bounded defer from the prior verdict is closed. |
| Regression vs M0 (VersionRef) + M1 (contract/FE gates) | all still pass | `go run ./scripts/api-lint -strict … .` = `0 violation(s)` (the 5 authz lints + M1's nullable/contract-sync blocking rules all green in the same run); no route/contract shape changed (correction is test+doc only). Tripwire golden/parity/determinism green. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Hand-sync tripwire arm defect class (0269/0270) | DEBT — arms hand-typed per migration | **CONFIRMED closed** | Arms generated from `TripwireArms`; PARITY byte-pins SQL↔map; DRIFT catches a new unarmed gated write. Both blocking, unchanged and re-green. Root cause (no single source of truth) removed. |
| Two latent P0001 incidents (force-release, obsolete) | Live fail-closed for every actor | **Fixed by the generation AND now proven live** | Validator reproduced: 0270 arm → `P0001 none of {document.edit}` for both `membership.manage` and `document.obsolete`; 0271 arm → both PASS. No longer inferred from a static drift proof — demonstrated on the real trigger. |
| **Live-QA integrity of the defer itself** | Prior PASS trusted a compile-only defer | **Corrected** | The defer was the weak point; running it exposed non-functional drives (application-stack death) and forced the DB-arm rewrite. The proof is now end-to-end, not fixture-shaped. |

- **Could it be built better?** The `render.go` note from the prior verdict (inlines 0270 SQL body as
  Go string concatenation rather than reading the prior migration as a template) still stands — fully
  protected by golden + PARITY today, recorded as M9 arm-hygiene input, not a FAIL. No new construction
  concern from the correction; the rewrite is the better construction. No blocking retrospective.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass — *clean; each acceptance mapped to a re-run command, live half reproduced.*
- [ ] **Fixture/mock passed off as real-provider proof** — *clean, and materially stronger than the prior verdict: the earlier compile-only defer was the risk here; it is now discharged with a real-trigger live RED→GREEN the validator reproduced. This is the exact class the re-validation was ordered to close, and it is closed.*
- [ ] Consumer contract guessed — *clean; map/migration/pin/arm verified against source.*
- [ ] Split-brain — *clean; SQL generated from the one Go map, byte-pinned; core unchanged.*
- [ ] Self-judged close / validator edited or fixed code — *clean; validator only judged, ran the drives via an ephemeral role, reverted the temporary 0271 move, dropped `qa_val` + both `metaldocs_test*` clones, left tree and real `metaldocs` DB clean, writes only this file.*
- [ ] Scope drift — *clean; the test rename/rewrite is a documented correctness correction with written rationale (C1.7), not new scope.*
- [ ] Symptom-patch — *clean; root cause (hand-sync) structurally removed; the correction fixed a non-functional test (real defect) rather than masking it.*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (senior-level, single source of truth, no dead code, no guessed
  contract; core byte-unchanged and re-green) and **function-wise** (arms generated + both incident-class
  P0001s now proven fixed on the **live trigger**, RED→GREEN independently reproduced by the validator;
  the rewritten drives genuinely exercise the documents(UPDATE) cap arm with `RowsAffected==1` closing
  the RLS-false-green hole; status-held isolation from the legal-transition triggers verified sound).
  HS-7 clean (arms == contract §1.2; 0271 diff-minimal vs 0270). The prior bounded defer is **DISCHARGED**.
- Handed back to the main session to flip M2 status and present the **HS-1** operator gate. Validator
  does not flip status.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only after main session acts on this PASS
