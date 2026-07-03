# Milestone 2 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4, binding; HS-7) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-03 · **Verdict:** see C7 → **PASS**.
> Validator judged from a clean tree (only untracked `docs/release/`); re-ran every gate itself; reproduced both negative proofs independently and left the tree clean.

## Inputs loaded (none missing)

`milestone.md`, `validation-contract.md` (D4), `f2.1/{spec,plan,evidence}.md`, `f2.2/{spec,plan,evidence}.md`, program `README.md`, governing `mission.md` (linked), aggregate diff `git diff 1a86f419..HEAD` (22 files, +2579/−2 — matches declared scope: tripwire source+gen+migration, api-lint rules, integration drives, F2.2 pin + 3 doc-truth files, milestone artifacts).

## C1 — Spec & plan conformance (per feature)

Both feature `spec.md`/`plan.md`/`evidence.md` exist and are execution-shaped (plans carry task lists, files, ordering, test strategy — not re-spec). Approval line filled (2026-07-03) on both; the binding D4 contract was committed (`4a815be9`) **before** the first code commit (`70c2803b`). Consumer contract (CI + DB trigger for F2.1; CI + future maintainers for F2.2) declared and honored — verified from source, not the reverse.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F2.1 | ✅ `TripwireArms` == contract §1.2 (18 entries, verified line-by-line); 0271 generated from the map; both lint rules blocking | ✅ all §1.7 exit criteria met (re-run below) | ✅ no arm tightening (`template.submit` retained, §1.3 defer); no cross-layer call-graph; function-body swap only | `arms.go`, `render.go`, `0271_*.sql`, `tripwire_arm_rules.go` |
| F2.2 | ✅ tier-1 `permissions.go:157`==tier-2 `repository.go:798,828`==`membership.manage`; `/approval/routes`==`route.manage`; pin targeted to the two routes only | ✅ §2.3 (verify + pin RED/GREEN + doc-truth) | ✅ deliberate coarse/fine `/approval/` split recorded intentional, not "aligned"; no behavior change | `permissions_test.go` pin, ADR 0022/0018 + module-iam.md annotations |

HS-7 map/contract diff: `TripwireArms` matches contract §1.2 **exactly** (entry #6 widened to `{document.edit, document.obsolete, membership.manage}`; #13 retains `template.submit`). Pinned in code by `TestTripwireArms_MatchesContractTable`. 0271-vs-0270 diff shows the **only** CASE-branch delta is the documents(UPDATE) `v_required_caps` literal (plus header/ledger/comment). No divergence → no HS-7.

## C2 — Gates re-run, isolated (validator-run from clean tree, not trusted from evidence)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| both | `go build ./...` | exit 0 | ✅ |
| F2.1 | `go test ./internal/platform/tripwire/... -v` | `TestTripwireArms_CapsAreRegistryReal`, `TestTripwireArms_MatchesContractTable`, `TestRenderMigration_MatchesCommittedFile` (golden), `TestRenderMigration_Deterministic` — all PASS | ✅ |
| F2.1 | `go test ./scripts/api-lint/... -run TripwireArm -v` | 12 rule tests PASS: parity clean-green + non-registry-cap RED + mutated-migration RED; drift clean-green + synthetic-unarmed RED + match-one green + force-release-shape reproduction | ✅ |
| F2.1 | `go test ./internal/modules/iam/domain -run TestCapabilityRegistrySize -v` | `--- PASS` (`const want = 34`, confirmed in `model_test.go:96`) | ✅ |
| F2.1 | `go run ./scripts/api-lint -strict api/openapi/v1/openapi.yaml .` | `0 violation(s)` (both new rules registered in `RunCodeRules`, blocking — exit 1 on any hit) | ✅ |
| F2.1 (neg) | revert arm→`{CapDocumentEdit}`, `go run ./scripts/api-lint -only TRIPWIRE-ARM-DRIFT …` | `3 violation(s)` at `obsolete_service.go:93`, `repository.go:811`, `repository.go:841` — each: "the DB tripwire will reject this write with P0001 for every actor"; `git checkout` → `0 violation(s)`. Tree left clean. | ✅ |
| F2.1 | `go vet -tags integration ./tests/integration/documents/` | exit 0 (drives compile under the integration tag) | ✅ |
| F2.2 | `go test ./apps/api/cmd/metaldocs-api/ -run TestTier1Tier2CapabilityCoherence_F4Sites -v` | `--- PASS` | ✅ |
| F2.2 (neg) | flip `permissions.go:157` force-release cap→`CapDocumentView`, re-run pin | `--- FAIL … tier-1 cap="document.view" diverges from tier-2 cap="membership.manage" (ADR 0022 F4 regression)` (permissions_test.go:717); `git checkout` → PASS. Tree left clean. | ✅ |

## C3 — Senior review of the aggregate milestone diff

- **No split-brain.** `render.go` produces every arm literal from `TripwireArms` via `findArm()/renderArray()`; the committed `0271_*.sql` is byte-pinned to `RenderMigration()` by both the golden test and the `TRIPWIRE-ARM-PARITY` api-lint rule. The Go map is the sole source of truth; the SQL is generated output. This is a genuine structural closure of the hand-sync defect class (0269/0270), not a re-hand-sync.
- **No dead code / no feature-cross-break.** F2.1 (tripwire files) and F2.2 (permissions test + docs) are disjoint. F2.1's single behavior change is additive (arm widened) — no `document.edit` writer can regress (match-one). Pre-existing `tests/integration/templates/tripwire_caps_test.go` (0269/0270 guard) is **unchanged** since M1 close (verified `git diff 1a86f419..HEAD` empty for that path).
- **Drift rule is not a tautology.** Independently reproduced: it fires on exactly the two real latent write-paths (three call sites) when the arm is narrowed, and is silent on the clean tree. It catches the real incident class.
- **Findings:** none blocking.
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression (backend-api authz / DB-invariant class)

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api authz/DB-invariant) | pass | Deterministic gates green from clean tree: `go build`, api-lint strict `0 violation(s)`, tripwire golden/parity/determinism, drift RED/GREEN negative proof, registry-size 34. Live integration = bounded defer (below). |
| Regression vs M0 (VersionRef) + M1 (contract/FE gates) | all still pass | `go build ./...` green; `go run ./scripts/api-lint -strict … .` = `0 violation(s)` (the 5 authz lints + M1's nullable/contract-sync blocking rules all green under the same run); no route/contract shape changed (net code diff = additive arm + test + docs). `TestCapabilityRegistrySize`(34) unchanged. |

**Bounded defer — integration live run (legitimate, per contract §1.6/§4).** Confirmed `METALDOCS_DATABASE_URL` and `DATABASE_URL` both unset in the process env; `scripts/test.ps1` is a bare `go test ./...` with no integration tag and no env load; creds live only in `.env` (forbidden to read). Drives are **authored + compile-verified** (`go vet -tags integration` clean) with a recorded run-trigger. Compensating control: the `TRIPWIRE-ARM-DRIFT` static proof — which I reproduced independently — demonstrates those same two write paths were fail-closed P0001 pre-0271 and covered post-0271. Adequate per §1.6/§4; matches M1's env-risk precedent.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Hand-sync tripwire arm defect class (review Dim-2 finding #1; shipped 0269, 0270) | DEBT — arms hand-typed per migration; two prod P0001 incidents + two latent | Structurally closed toward CONFIRMED | Arms generated from `TripwireArms` (same registry the services read); `TRIPWIRE-ARM-PARITY` byte-pins SQL↔map; `TRIPWIRE-ARM-DRIFT` catches a new unarmed gated write. Both blocking. Root cause (no single source of truth) removed, not symptom-patched. |
| Two latent P0001 incidents (force-release, obsolete) | Live fail-closed for every actor | Fixed **by the generation** (arm widened via the map), not a one-off patch | 0271 diff-minimal vs 0270; drift rule fires on exactly those 3 sites pre-fix (reproduced). |
| F2.2 tier-1↔tier-2 coherence | Review claimed 2 open divergences | Confirmed already code-resolved (ADR 0022 Phase 11 F4); pinned + doc-truth restored | Pin GREEN on HEAD / RED on synthetic re-divergence (reproduced); ADR 0022/0018 + module-iam.md back-annotated RESOLVED. Not a hidden HS-2 — no redesign, verify-and-pin only. |

- **Could it be built better?** Minor construction note (not unsound, not a FAIL): `render.go` inlines 0270's SQL body as Go string concatenation rather than reading the prior migration file as a literal template, so the non-arm SQL text is transcribed into Go. It is fully protected today (golden test + PARITY rule byte-compare against the committed 0271, and the 0270↔0271 diff confirms byte-faithfulness), so there is no live drift risk — but a future regeneration could read the prior migration as the template. Record as input to M9 arm-hygiene (alongside the already-recorded superset-prune defer). Does not affect the verdict.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean; each acceptance mapped to a re-run command.*
- [ ] Fixture/mock passed off as real-provider proof — *clean; integration live run is an explicit, legitimate bounded defer with a static compensating control, labeled as such.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean; map/migration/pin verified against source consumer sites.*
- [ ] Split-brain — *clean; SQL is generated from the one Go map, byte-pinned.*
- [ ] Self-judged close / validator edited or fixed code — *clean; validator only judged, reverted its two temporary probes via `git checkout`, left tree clean, and writes only this file.*
- [ ] Scope drift — *clean; the two latent incidents + F2.2 reframing were surfaced up-front in `milestone.md`/contract as HS-6 discovered truth with rationale; aggregate diff matches declared scope.*
- [ ] Symptom-patch — *clean; root cause (hand-sync) structurally removed; drift power proven live, not gutted.*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (senior-level, contract-clean, single source of truth, no dead code, no guessed contract) and **function-wise** (arms generated + both incident-class incidents fixed by the generation; drift + parity rules blocking and proven; F2.2 coherence pinned and doc-truth restored). HS-7 clean (map == contract §1.2; 0271 diff-minimal vs 0270). Live integration is a legitimately-recorded bounded defer with an adequate reproduced static compensating control.
- Handed back to the main session to flip M2 status (README + milestone.md) and present the **HS-1** operator gate. Validator does not flip status.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only after main session acts on this PASS
