# T10 Fable independent review — Round 2

> **Evidence only — non-authoritative. This review branch must never merge.**

## Review identity

```text
Repository                developmentconexus-ops/MetalDocs
Gate                      T10 — Transition / Cutover
Candidate branch          arch/t10-transition-cutover
Exact corrected HEAD      c1afc292bc94f48bfd2146c3b4374342ff5c2701
Required candidate CI     #1157 SUCCESS
Candidate Draft PR        #158
Round-1 Evidence PR       #159 CLOSED / UNMERGED
Round-1 verdict           NOT CONVERGED / MATERIAL=3
Review branch             review/t10-fable-r2
Round                     2 / BOUNDED CONFIRMATION
```

## Scope

This is **not** a fresh unconstrained redesign. Confirm the Round-1 corrections and detect regressions only.

Read strictly:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/work/current/t10-transition-cutover.md
→ only exact owner authority needed to verify a challenged correction
```

## Confirm F1 closure — B2 clean seal / B3 authority edge

Verify that:

```text
B2 proof binds to the exact deployed production candidate/profile
proof-fixture Product truth is removed before authority begins
a mechanical clean-baseline verification closes B2
clean-seal evidence is operations/provenance only, never Product state
affected B2 evidence is re-armed after reset/rebuild
proof mutation paths are fenced after the clean seal
any unexpected post-seal Product mutation blocks destructive reset pending classification
B3 = first post-seal authoritative Product bootstrap commit
no Product activation marker/table/endpoint/Permission/owner is invented
```

Attack both old failure directions:
1. proof fixture silently promoted into authority;
2. committed authority destroyed by a supposedly pre-B3 reset.

The Lead deliberately **did not** adopt a Product activation marker. Challenge whether the corrected clean-seal + first-post-seal Product commit law is mechanically decidable without one.

## Confirm F2 closure — authoritative recovery point before B4

Verify that B4 cannot expose ordinary serving unless:

```text
B3 authoritative baseline exists
target remains ready
at least one complete authoritative R10 recovery point covers the current B3 baseline
complete-set/manifest/exact-content integrity checks pass
```

T8-G restore capability / isolated restore drill belongs B2. T10 only sequences an actual authoritative recovery point after B3 and before B4.

Also challenge the Lead refinement:

```text
loss of canonical R10 authority + every coherent authoritative recovery point
→ catastrophic authority loss
→ remain fail-closed
→ no automatic re-bootstrap / disposable-state promotion
→ explicit operator/business recovery decision + smallest architecture reopen required
```

Confirm this is more truthful than treating re-bootstrap as ordinary recovery.

## Confirm F3 closure — serving fence / cleanup

Verify that B4 requires every inventoried user-reachable disposable DEV/test serving path to be unable to accept ordinary Product mutations, including stale DNS/cache/direct-origin/bookmark routes.

DNS switch alone must not count as fencing.

Verify cleanup now requires:

```text
contains no business truth requiring bounded reopen
```

with no temporal `pre-R10` loophole. Any unexpected business truth in a supposedly disposable estate must stop cleanup.

## Confirm bounded MINOR handling

- F4: exact production candidate/profile wording; reset-dependent proof re-armed.
- F5: T3 non-serving bootstrap concern is the semantic anchor. **Do not** reinterpret T8-D `bootstrap/provisioner` trust class as Product bootstrap; it is provisioning/DDL-only. If later implementation proves accepted T8-G surfaces cannot realize semantic bootstrap, only then is bounded T8-G reopen triggered.
- F6: external estate is unclassified until B0 proof; expected DEV/test status is never disposal permission.

## Regression envelope

Must remain:

```text
barriers                         exactly B0→B4 / 5 total
historical business migration   absent
business authority              singular
application operations          78
operation 79                    absent
new Permission                  none
new semantic owner              none
new Product state               none
new runtime capability          none in current T10 candidate
T11/T12                         not open
Product implementation          blocked
```

A possible future T8-G reopen named as a conditional implementation falsifier is not a current runtime-capability addition.

## Finding standard

Classify findings:

```text
MATERIAL
MINOR
NOTE
```

A new MATERIAL finding must identify:
1. exact corrected candidate claim;
2. exact accepted authority/property violated or still uncovered;
3. reachable failure sequence;
4. smallest correction/upstream reopen;
5. why Round 1 did not already cover it.

## Required verdict

End exactly:

```text
VERDICT = CONVERGED | NOT CONVERGED
MATERIAL findings = N
Round 3 justified = YES | NO
```

Write review Evidence below this line only.

---

## Fable Round 2 response

```text
Reviewer                  Fable (independent, isolated)
Basis                     corrected candidate c1afc292bc94f48bfd2146c3b4374342ff5c2701 (CI #1157 SUCCESS, verified)
Correction commit         7c5bb3e0 (CI #1156 SUCCESS, verified)
Corrected delta vs main   docs/roadmap.md + docs/work/current/t10-transition-cutover.md only (verified)
Review delta              docs/work/current/ai-dialog.md only
Mode                      bounded confirmation — no fresh unconstrained redesign performed
```

Route followed: `AGENTS.md` → `docs/index.md` → `docs/roadmap.md` → `docs/work/current/t10-transition-cutover.md`, plus only re-verification greps of the owners already cited in Round 1 (`architecture/runtime.md` §4/§18, `authorization-and-audit.md` non-serving bootstrap concern, `persistence.md` §20). The roadmap's recorded evidence chain was independently verified live: candidate CI #1153 on `0b90f266`, correction CI #1156 on `7c5bb3e0`, corrected-candidate CI #1157 on `c1afc292`, Round-1 final review CI #1155 on `0f47dfc2` — all SUCCESS on the exact stated HEADs.

### F1 closure — CONFIRMED

The corrected B2/B3/§5 text closes both Round-1 failure directions:

- **Fixture promotion (direction 1).** B2 now closes only on a verified clean seal whose conditions include removal of all proof-fixture Product truth, session invalidation, fixture-Audit absence, content/River-intent non-authority, recorded exact artifact/config/profile identity, mechanical clean-baseline verification, and fencing of proof mutation paths before bootstrap. Bootstrap may begin only after a valid seal, so fixture truth cannot lawfully survive into the authority baseline; the strengthened falsifier ("proof-fixture Product truth survives the B2 clean seal into the authority baseline") declares the contract false if a defective realization lets it happen, and §13 hands the realization attack ("B2 fixture cleanup + clean-seal evidence") to T12 where it belongs.
- **Authority destroyed by a supposed pre-B3 reset (direction 2).** The reset permission now has a decidable guard. The seal verifies the baseline empty of Product truth; therefore "any Product mutation after the seal" is mechanically checkable (any Product/Audit truth present now = post-seal mutation). §5 permits destructive reset only while no post-seal Product mutation with unresolved authority status exists; B2's incident clause blocks reset pending classification of any unexpected post-seal mutation; the committed bootstrap case is separately forbidden by the B3 law. The dangerous default is fail-closed in every branch. The Round-1 undecidable predicate is gone.
- **Mechanical decidability without a Product activation marker (direction 3).** Confirmed decidable. The decision procedure needs exactly two observables: (a) the operations/provenance clean-seal evidence (existence, identity, validity), and (b) whether any Product truth exists in the sealed baseline. Because the seal certifies verified emptiness, the first post-seal Product commit — flowing only through the accepted T3 non-serving bootstrap concern, with proof paths fenced — is itself the observable authority edge; its canonical Product facts plus required Audit (where T3 requires it) are durable system evidence. No Product marker/table/endpoint/Permission is needed, and the explicit-absent list plus falsifier now forbid inventing one. The Lead's rejection of a Product activation marker is the correct smaller shape: a marker would be new Product state whose only consumer is the transition — exactly what the regression envelope forbids.

### F2 closure — CONFIRMED

B4 now requires, before ordinary serving: the B3 baseline, continued readiness against the accepted production profile, at least one **complete authoritative R10 recovery point covering the current B3 baseline**, and complete-set/manifest/ExactContentDescriptor integrity checks on that point. Restore capability and the isolated restore drill correctly remain B2/T8-G proof (`runtime.md` §18: "Backup success alone is not restore proof" — the corrected split proves capability at B2 and requires an actual captured point at B4, which is precisely the sequencing Round 1 found missing). §8 binds the required recovery point to the full T8-G recovery set including required exact content and transaction-coupled River intents. The B3→B4 window is now explicitly an R10 recovery domain with B4 preconditions re-established after any recovery.

The catastrophic-authority-loss refinement is confirmed as **more truthful** than the Round-1 sketch of re-bootstrap-as-recovery: loss of canonical authority plus every coherent recovery point means the contract's own fail-closed premises failed, so remaining fail-closed with no automatic re-bootstrap, no disposable-state promotion, and an explicit operator/business decision plus smallest implicated architecture reopen is the only honest disposition — an automatic re-bootstrap path would silently hide the falsified premise and manufacture replacement business truth. The zero-recovery-point case Round 1 found lawless now has a named lawful terminal state.

### F3 closure — CONFIRMED

- **Fencing.** B4 requires every disposable DEV/test serving estate found by B0 to be stopped or fenced from ordinary user requests, and — estate-general, not inventory-bounded — that "no previously published user-reachable DEV/test origin can still accept Product mutations." DNS change alone is explicitly insufficient (stale resolver caches, direct old origins, bookmarked endpoints named). The B0 inventory gained the matching class ("previously published user-reachable origins/endpoints"), and the falsifier list carries both the stale-route clause and the estate-general post-B4 mutation clause.
- **Cleanup predicate.** The temporal qualifier is gone: "contains no business truth requiring bounded reopen," reinforced by the explicit sentence that business truth discovered in a supposedly disposable resource — written before or after B4 — stops cleanup and routes adjudication, and by the falsifier "cleanup deletes a resource containing any business truth merely because the truth is post-cutover." The Round-1 lawful-deletion path is closed.
- The declined waiting/observation barrier (§9) is the correct smaller shape — Round 1 asked for ordering, not ceremony, and the ordering now lives in B4.

### Bounded MINOR/NOTE handling — CONFIRMED

- **F4.** B2 binds to the exact deployed production candidate/profile; a drifted twin is explicitly insufficient for candidate-bound properties; any reset/rebuild invalidates the seal and dependent evidence, with survival only for structurally independent evidence whose exact subject is unchanged. Closed as specified.
- **F5.** §6 now anchors semantic bootstrap to the accepted T3 non-serving operations concern, explicitly refuses to reinterpret the T8-D `bootstrap/provisioner` trust class as semantic bootstrap (provisioning/DDL-only — matching `persistence.md` §20), and routes the only remaining uncertainty correctly: if T11 evidence proves no accepted T8-G surface can realize bootstrap truthfully, that is a bounded T8-G reopen before implementation, never silent shell expansion or a serving-wire workaround. This conditional reopen is a falsifier route, not a current runtime-capability addition. Closed.
- **F6.** §2 and the roadmap now hold later-discovered estate **unclassified until B0 proof**, with the DEV/test expectation never granting disposal permission; the B0 barrier-failure clause dropped its own "pre-R10" qualifier to match. Closed.
- **F7.** §10 MIG dispositions unchanged and still exact. Confirmed.

### Round-2 adversarial pass on the corrected text — no MATERIAL finding survived

- **Seal-condition ordering.** The seven seal conditions are listed unordered; a pedantic realization could run mechanical clean-baseline verification before fencing proof paths, leaving a mutation window between verification and fence. Attack not sustained as MATERIAL: the seal is valid only when **all** conditions hold, an unexpected post-seal mutation is an authority-boundary incident that blocks destructive reset, and the fixture-survival falsifier declares the contract false if such a mutation reaches the authority baseline. Residue recorded as **NOTE R2-N1**: the Lead may state that clean-baseline verification must hold at seal completion (i.e., be current after fencing); this is realization precision T12 would catch, not a contract gap.
- **§5 reset-guard wording.** "…has not been followed by any Product mutation whose authority status is unresolved" does not by itself forbid reset after a *resolved-authoritative* mutation — but that case is B3 by definition, outside the "Before B3" section, and §4/B3 forbids destructive reset absolutely. Composition is airtight; recorded as **NOTE R2-N2** (wording relies on section scoping) with no correction required.
- **Read-serving from unfenced estate.** B4's mutation-fencing floor plus "stopped or fenced from ordinary user requests" leaves at most transient stale *reads* of disposable estate during propagation; no business truth can be created or lost, one authority is preserved, and cleanup remains truth-gated. Not a transition-contract defect. NOT SUSTAINED.
- **Recovery-point staleness.** "Covers the **current** B3 baseline" correctly forces re-capture if bootstrap extends the baseline before B4. NOT SUSTAINED.
- **Integrity-check vs no-fresh-drill tension.** B4's manifest/descriptor integrity checks without a per-point full restore drill match `runtime.md` §18 ("the chosen backup profile must prove the complete recovery set and descriptor integrity") with the drill proven at B2. NOT SUSTAINED.
- **Regression envelope.** Verified on the corrected delta: barriers remain exactly five (B0→B4, relabeled not added); historical migration absent; business authority singular; 78 operations preserved and operation 79 absent in both files; no new Permission, semantic owner, or Product state (clean seal explicitly operations/provenance-only; activation marker explicitly in the absent list); no new runtime capability in the current candidate (fencing and sealing are operational acts; the T8-G reopen is conditional and correctly routed); T11/T12 remain NOT OPEN; Product implementation remains BLOCKED. Section renumbering (§12 adjudication record, §13/§14) is work-doc-internal with no authority effect. The roadmap's review-state block, adjudication table, and next-action sequence are consistent with the candidate and with the live CI/PR record. NO REGRESSION FOUND.

### Verdict

All three Round-1 MATERIAL findings are closed by bounded candidate-only corrections that match or improve on the Round-1 smallest-correction proposals; both Lead refinements (operations/provenance clean seal instead of any Product marker; catastrophic authority loss instead of re-bootstrap-as-recovery) are accepted as the more truthful smaller shapes. The two residual notes (R2-N1 seal-completion ordering, R2-N2 §5 wording scope) are non-blocking precision at the Lead's discretion and justify no further round.

```text
VERDICT = CONVERGED
MATERIAL findings = 0
Round 3 justified = NO
```
