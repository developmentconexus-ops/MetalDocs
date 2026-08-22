# T10 Fable independent review — Round 1

> **Evidence only — non-authoritative. This review branch must never merge.**

## Review identity

```text
Repository                developmentconexus-ops/MetalDocs
Gate                      T10 — Transition / Cutover
Candidate branch          arch/t10-transition-cutover
Exact candidate HEAD      0b90f26690b2b2bbf627f0c72283ff14c0ce9b84
Required candidate CI     #1153 SUCCESS
Candidate Draft PR        #158
Review branch             review/t10-fable
Round                     1 / ADVERSARIAL
```

## Read route

Read strictly:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/work/current/t10-transition-cutover.md
→ only the exact accepted owner needed to challenge a concrete claim
```

Do not recursively read history/legacy/closed PRs.

## Fixed envelope

```text
T1→T9                         CLOSED / OPERATOR-RATIFIED / INTEGRATED
T10                           OPEN / ACTIVE
T11→T12                       NOT OPEN
Product implementation         BLOCKED
legacy implementation live tree ABSENT
application operations         78
operation 79                   ABSENT
historical business corpus     NONE
required pre-R10 business corpus NONE
```

T7 remains binding: Launch has no historical business migration requirement. A review finding may not manufacture migration/compatibility machinery merely for symmetry.

## Candidate under attack

Selected posture:

```text
ONE-WAY GREENFIELD R10 ACTIVATION
+
PRIVATE PREPARATION
+
PROOF BEFORE AUTHORITATIVE BOOTSTRAP
+
FIRST-AUTHORITATIVE-MUTATION = POINT OF NO RETURN
+
SERVING ACTIVATION ONLY AFTER AUTHORITATIVE BASELINE EXISTS
+
ONE BUSINESS AUTHORITY AT A TIME
+
FAIL-CLOSED / R10 RECOVERY AFTER AUTHORITY BEGINS
```

Monotonic barriers:

```text
B0  source truth classified
B1  target privately prepared
B2  target proven while still non-authoritative
B3  first authoritative R10 Product mutation committed / point of no return
B4  canonical serving authority activated
```

## Adversarial questions

Do not optimize for agreement. Try to falsify the candidate.

Challenge especially:

1. **B0 source truth**
   - Can surviving DB/object/IdP/deploy state be incorrectly discarded even though it is actually authoritative?
   - Is the stop/reopen rule sufficient if a real pre-R10 business corpus appears?
   - Does the candidate silently assume an external estate that may not exist?

2. **B1/B2 private target and proof**
   - Can target proof require Product truth that would already cross B3?
   - Is there any accepted T8/T9 proof that cannot truthfully run while the target is still non-authoritative?
   - Does proof-before-B3 accidentally require a synthetic bootstrap or fake fixture that only proves itself?

3. **B3 point of no return**
   - Is “first authoritative Product mutation” the smallest mechanically identifiable boundary?
   - Can Company/User/ProviderSubjectBinding/configuration bootstrap commit partially or ambiguously so that the system crosses B3 without a coherent recoverable baseline?
   - Is any accepted Product truth established outside the definition of B3?
   - Is there a hidden need for an activation marker/table/state not already owned by Product? If yes, prove why rather than inventing it.

4. **B3→B4 serving activation**
   - After B3 but before B4, what failure modes exist while business truth exists but normal serving is still disabled?
   - Can external OIDC/DNS/ingress/config changes create competing authority or an unrecoverable half-cutover?
   - Is one business authority preserved even during retries/restarts?

5. **Rollback versus recovery**
   - Before B3, is reset/retry safe under every accepted target mechanism?
   - After B3, is destructive rollback truly forbidden in every path?
   - Can a binary/config rollback corrupt or misinterpret already-committed R10 state?
   - Are backup/restore/session invalidation/privacy/security readiness correctly inherited rather than weakened?

6. **Content / River / exact bytes**
   - Can content or River state become accidentally authoritative before B3?
   - Can pre-B3 content survive and later be mistaken for governed R10 content?
   - Can post-B3 cleanup remove required exact content or durable work evidence?

7. **Cleanup**
   - Are deletion preconditions strong enough to prevent removing recovery/provenance/security-critical resources?
   - Does cleanup need a waiting/observation barrier, or would that be ceremonial overengineering?

8. **Hidden migration/compatibility leakage**
   - Find any place where the candidate implicitly requires dual write, old/new reconciliation, legacy read fallback, schema translation, generic ETL/import, or historical reconstruction.
   - If such machinery is actually required, identify the exact accepted property that requires it.

9. **Stage leakage**
   - Identify any T11 implementation-decomposition or T12 readiness work smuggled into T10.
   - T10 may define required transition semantics/proof barriers, but must not implement or decompose Product code.

10. **Closed Product envelope**
    - 78 operations must remain exactly 78.
    - operation 79 must remain absent.
    - no new Permission, semantic owner, Product state, API operation, persistence authority or runtime capability may be invented solely for cutover.

## Finding standard

Classify every finding:

```text
MATERIAL
  candidate can violate accepted authority, lose/corrupt Product truth,
  cross an irreversible boundary ambiguously, or requires an upstream reopen

MINOR
  bounded precision/traceability issue that changes no accepted architecture

NOTE
  non-blocking observation
```

For each MATERIAL finding include:

```text
exact candidate claim
exact accepted authority/property contradicted or uncovered
reachable failure sequence
smallest correction or owning-authority reopen
why the finding is not preference/framework fashion/hypothetical scale
```

## Regression envelope

A correction must preserve unless a MATERIAL falsifier proves otherwise:

```text
barriers                         exactly B0→B4 / monotonic
historical business migration   absent
business authority              singular
application operations          78
operation 79                    absent
new Permission                  none
new semantic owner              none
T11/T12                         not open
Product implementation          blocked
```

## Required verdict

End exactly:

```text
VERDICT = CONVERGED | NOT CONVERGED
MATERIAL findings = N
Round 2 justified = YES | NO
```

Write review Evidence below this line only.

---

# Round 1 review Evidence

## Reviewer identity and method

```text
Reviewer                  Fable (independent, isolated)
Review HEAD basis         candidate 0b90f26690b2b2bbf627f0c72283ff14c0ce9b84
Candidate delta vs main   docs/roadmap.md + docs/work/current/t10-transition-cutover.md only (verified)
Review delta              docs/work/current/ai-dialog.md only
```

Fresh-actor route followed: `AGENTS.md` → `docs/index.md` → `docs/roadmap.md` → `docs/work/current/t10-transition-cutover.md`, then only the exact accepted owner needed per concrete attack: `decisions/repository-reset.md`-routed T7 authority at `architecture/transition.md`, `architecture/validation-baseline.md` (T9 — B2 proof-content attack), `architecture/runtime.md` §4/§18 (T8-G — runtime shells, backup/restore profile), `decisions/forward-obligations.md` MIG-05/06/10 rows, and targeted grep-verified lines of `architecture/authorization-and-audit.md` (bootstrap/ops concern, SYSTEM actor) and `architecture/persistence.md` §20 (bootstrap/provisioner trust class) for the B3-bootstrap attack. Exceeding the five-file default is the named material reason of this task: B0→B4 crosses source-truth (T7), proof (T9), runtime/recovery (T8-G) and obligation (forward-obligations) owners, and each barrier claim can only be attacked against its owner.

Method: falsification-first, both directions — missing failure classes (insufficient) and ceremonial machinery (not smallest). CI #1153 SUCCESS was treated as repository-envelope conformance only.

## Verified baseline (evidence, not trust)

```text
candidate delta                  roadmap + T10 work doc only; no durable architecture authority edited;
                                 no implementation, schema, OpenAPI, dependency or runtime content added
T7 conformance                   §2 source-truth block matches transition.md §1/§7 verbatim in substance;
                                 stop/bounded-reopen rule matches transition.md §6 reopen triggers
approach selection               B (dual authority) / C (compatibility bridge) rejections trace to T7 §3/§4
                                 ("no compatibility structure may survive merely to preserve disposable
                                 DEV/test data"); no accepted continuity requirement contradicts them
explicit-absent list             §3 mirrors T7 §3 consequences; no dual-write / reconciliation / ETL /
                                 legacy-fallback / shadow-mutation machinery found anywhere in the candidate
MIG dispositions                 §10 matches decisions/forward-obligations.md exactly
                                 (MIG-05 PRESERVE not activated, MIG-06 PRESERVE no mode selected,
                                 MIG-10 REOPEN not triggered); no DEFERRED idea instantiated
B3 vocabulary                    Company / User / ProviderSubjectBinding / DocumentType / Release are all
                                 accepted domain concepts (domain-model.md, persistence.md); no new Product
                                 state, Permission or semantic owner appears in any barrier
identity law                     §7 matches preserved AUTH-06 (no cross-system atomic IdP transaction) and
                                 the accepted anti-corruption seam; provider claims never Product authorization
River law                        §8 matches T8-G §18 (River intents share PostgreSQL; recovery point restores
                                 facts + intents coherently); River state never Product truth
hard invariants                  78 operations preserved; operation 79 absent; T11/T12 NOT OPEN;
                                 Product implementation BLOCKED; no historical migration introduced
T11/T12 leakage                  none found — §12 hands off decomposition/readiness without performing them;
                                 no implementation plan, work graph or readiness attack exists in the candidate
```

---

## Findings

### F1 — MATERIAL — B3 is not mechanically identifiable as written: the B2→B3 boundary has no verified-clean-baseline event and no recorded authority declaration, so the pre-B3 destructive-reset permission is gated on a predicate that is undecidable from system state

**Exact candidate claim.** §4/B3: "B3 is the point of no return. It occurs on the first committed authoritative R10 Product mutation." §13 exit criterion: "B3 is unambiguously the first authoritative R10 Product mutation / point of no return." Roadmap (candidate edit): "Before B3, technical preparation may be reversed only while no authoritative R10 Product mutation has committed."

**Exact accepted authority/property uncovered.** Two, one per failure direction:

1. T7 (`transition.md` §1/§2): no DEV/test state may become business truth — and its reopen trigger 3 ("a production MetalDocs dataset created before cutover that becomes business-authoritative") names silent promotion as a reopen-class catastrophe. The candidate's own falsifier list concedes this class: "pre-R10 sessions or DEV/test Audit/history are silently promoted into R10 business truth."
2. The candidate's own §5 pre-B3 recovery law permits "reset/destroy the non-authoritative target Product database" — a destructive permission whose validity condition is "no authoritative mutation has committed."

**Why the boundary is undecidable as written.** B2 proof legitimately commits Product-kind mutations through the real production paths (T9 global law: "the real production subject/boundary is executed"; E2/E3 forbid mock substitution — GF1–GF6 create Users, sessions, config, Documents, Audit rows, content, River intents on the target). A B2 proof-fixture Company row and a B3 bootstrap Company row are identical in kind, table, and mechanism-visible properties. §6's channel rule ("authoritative bootstrap data may be created only through explicit accepted non-serving bootstrap/administrative mechanisms") gives a partial identity criterion — authoritative mutations travel a distinct channel — but that mechanism is explicitly "defined by later implementation planning," and nothing requires its first invocation to be preceded by verified baseline cleanliness or to be recorded as the authority transition. B2 states a cleanliness *requirement* ("the authoritative production Product database/content baseline is clean of DEV/test business state before B3") with no defined event, evidence, or actor that establishes it.

**Reachable failure sequences (both directions).**

- *Fixture promotion:* B2 proof rounds leave fixture rows/content on the target; the team proceeds to bootstrap on top of them without a mandated verified-clean checkpoint; fixture state silently becomes part of the authoritative R10 baseline. The candidate names this a falsifier but provides no barrier that detects or prevents it — the falsifier is true and nothing in B0→B4 makes it false.
- *Authority destruction:* bootstrap has committed; a later incident prompts an operator to invoke §5 pre-B3 recovery ("reset/destroy the non-authoritative target Product database") in the belief that the proof phase is still running. The predicate "no authoritative mutation has committed" cannot be evaluated from the database — it is operational history. Committed R10 business truth is destroyed under a lawful-looking reading of the contract. This is the exact event B3 exists to make impossible.

**Smallest correction (candidate files only; no upstream reopen).** Make the B2→B3 transition an explicit, evidence-backed barrier event: (a) a verified-clean check of the authoritative Product database/content baseline (fixture state destroyed, evidence recorded); (b) a recorded authority declaration — an operational/provenance record, **not** a new Product table/state/operation — after which the destructive-reset permission of §5 is dead; (c) B3 defined as the first Product mutation committed through the accepted non-serving bootstrap channel **after** that declaration, making "authoritative" decidable by construction. The five-barrier structure, monotonicity, and the first-authoritative-mutation law survive intact; the declaration merely gives the law an observable edge. The review brief's own question 3 warning ("hidden need for an activation marker … prove why rather than inventing it") is answered by the two failure sequences above: without a recorded transition event, the point of no return exists only in narrative, and both catastrophic directions remain reachable.

**Why not preference/fashion/hypothetical scale.** The candidate's own exit criterion claims unambiguity; the claim is falsified by two concrete, opposite-direction failure sequences on the accepted Launch topology. No framework, tooling, or scale assumption is involved — only the decidability of the single predicate on which the contract's only irreversible permission flips.

### F2 — MATERIAL — no barrier requires a verified authoritative R10 recovery point to exist before B4 serving activation; the zero-recovery-point loss case after B3 is lawless under the contract's own permitted-response classes

**Exact candidate claim.** §4/B4: "Only after B3 has established the required authoritative R10 baseline and the target remains ready may canonical production ingress/origin expose normal R10 serving." §5 post-B3 permitted response classes: "fail closed / stop normal serving; recover from a coherent R10 recovery point; restore R10 and satisfy all restore readiness barriers; apply a proven compatible R10 correction/deployment; redrive accepted durable work under canonical Product truth." Selected law: "FAIL-CLOSED / R10 RECOVERY AFTER AUTHORITY BEGINS."

**Exact accepted authority/property uncovered.** T8-G §18: "A repeatable isolated restore-drill path is required before production readiness is considered proven. Backup success alone is not restore proof"; "Every production recovery profile must state and measure its achieved/target RPO." T9 V10/GF6 prove backup/restore **capability**. Neither owner sequences *an actual captured recovery point covering the authoritative baseline* against *the beginning of authority/serving* — that ordering is precisely T10's property to own, and the candidate does not own it.

**Reachable failure sequence.** B3 bootstrap commits the authoritative baseline. B4 activates immediately — the contract permits it: the baseline exists and the target "remains ready." Before the first authoritative backup ever captures the baseline (nothing requires one to exist), the single PostgreSQL instance of the accepted Launch topology (replicas=1 baseline; one Product-state database) suffers storage loss. Now: destructive reset to DEV/test is forbidden (correctly); "recover from a coherent R10 recovery point" is impossible — no recovery point exists; "restore R10" has nothing to restore; "apply a compatible correction/deployment" does not recreate lost truth; "redrive durable work" has no canonical truth to redrive under. Every permitted response class is inapplicable. Committed R10 business history — possibly including real post-B4 user commits — is unrecoverable, and the contract offers no truthful path forward (re-bootstrap is not among the permitted classes, so the only real-world exit is extra-contractual).

**Smallest correction (candidate files only; no upstream reopen).** Two bounded additions: (a) B4 activation precondition — at least one verified authoritative R10 recovery point covering the B3 baseline exists before canonical serving is exposed (and recommended: captured immediately at/after B3); (b) one sentence in §5 naming the residual catastrophic class truthfully — if every authoritative recovery point is lost, re-establishing authority is a new authoritative bootstrap under B2/B3 discipline, never a promotion of disposable state. This adds no infrastructure, no new mechanism, and no RPO number; it sequences an already-accepted T8-G capability against the candidate's own barriers.

**Why not preference/fashion/hypothetical scale.** The failure needs no scale hypothesis — it is the accepted single-instance Launch topology on day one. The candidate's headline posture ("FAIL-CLOSED / R10 RECOVERY AFTER AUTHORITY BEGINS") is presently unsatisfiable in a reachable window that the barrier order itself creates; a fail-closed claim whose recovery leg can be vacuously empty is untruthful, and T10's stated goal is the smallest **truthful** contract.

### F3 — MATERIAL — legacy-estate serving is not fenced by any barrier relative to B4, and the §9 cleanup predicate's "pre-R10" qualifier makes deletion of real post-cutover business content contract-lawful

**Exact candidate claim.** §4/B4: "one canonical production ingress/origin → one R10 serving system → no ordinary fallback to a prior DEV/test implementation." §9: a resource may be removed only after proving, among others, "contains no pre-R10 business truth requiring bounded reopen." Candidate falsifier: "an ordinary request can fall back to DEV/test authority."

**Exact accepted authority/property uncovered.** T7 (`transition.md` §6, trigger 3): a production dataset that becomes business-authoritative is a reopen-class event. T10's own scope (transition.md §4) includes "DNS/ingress/origin configuration" and the "legacy technical deletion map" — the half-cutover ordering of exactly these resources is T10's property, and no barrier owns it: B0 only inventories DNS/ingress; §9 cleanup is evidence-driven and explicitly unordered relative to B4.

**Reachable failure sequence.** The B0 inventory finds a surviving DEV/test deployment serving at the current name/origin. B4 activates R10 on the canonical production ingress. DNS propagation, resolver caches, and bookmarked old origins keep the still-running DEV/test deployment reachable for hours or days — cleanup (§9) has no deadline and no ordering constraint relative to B4. Real users transact against the DEV/test estate believing it is production and commit real business content there. Then either (a) that content must become authoritative — the T7 reopen-class event — or (b) cleanup later evaluates §9 truthfully: the content is not current R10 authority, not needed for pre-B3 reversal or R10 recovery, and is **not "pre-R10 business truth"** (it was written after cutover) — all four preconditions pass and real business content is deleted lawfully. The candidate's falsifier ("an ordinary request can fall back to DEV/test authority") is true in this sequence, yet no barrier makes it false: stale-path arrival is not "fallback" by the R10 system, so even a faithful implementer of B4's activation law never touches it.

**Smallest correction (candidate files only; no upstream reopen).** (a) Add a B4 activation precondition: no disposable DEV/test estate remains reachable for ordinary serving on the canonical production name/origin path — legacy serving endpoints are stopped or fenced at or before B4 (fencing is reversible pre-B3-style technical work; it requires no new machinery). (b) In §9, drop the temporal qualifier: "contains no business truth requiring bounded reopen." Both are single-line edits that convert a named falsifier into a guaranteed property.

**Why not preference/fashion/hypothetical scale.** DNS/cache half-cutover is not hypothetical — it is the default behavior of the exact resources the candidate's own B0 inventory enumerates. The finding requires no new infrastructure, only ordering; and the §9 defect is textual: as written, the contract *authorizes* deleting real business truth, which contradicts its own evidence-driven fail-closed cleanup posture.

### F4 — MINOR — B2 evidence validity across pre-B3 resets and the "production-equivalent" phrase are unpinned

§5 pre-B3 recovery permits destroying the target Product database after B2 proof has passed; B4 then requires only that "the target remains ready," with no statement of which B2 evidence survives a reset and which must be re-established (schema/config/secret fail-closed evidence plausibly survives; instance-state-dependent evidence may not). Separately, B2 requires "T9 proof coverage on production-equivalent real mechanisms/boundaries" while T9's global law demands "the real production subject/boundary"; "production-equivalent" invites proof on a drifted twin. Bounded correction: state that B2 evidence binds to the deployed production candidate/profile (not an equivalent environment) and that any pre-B3 reset re-arms the affected B2 evidence before the F1 verified-clean/declaration event. Not MATERIAL: a rigorous prover reading B2's "production candidate/profile" phrase together with T9 reaches the right answer; the defect is precision, not an unowned class.

### F5 — MINOR — §6's future bootstrap mechanism should cite its accepted anchors so T11 cannot be forced into a silent runtime-census extension

§6 defers the bootstrap mechanism to "explicit accepted non-serving bootstrap/administrative mechanisms defined by later implementation planning." Accepted authority already reserves this exact concern: `authorization-and-audit.md` ("Bootstrap/recovery remains an explicit non-serving operations concern, never an ordinary RBAC bypass"), the SYSTEM actor attribution class (`system_actor_code`), and `persistence.md` §20's bootstrap/provisioner trust class. T8-G §4's runtime-shell list (`serve`/`migrate`/`jobs`/`recovery`) contains no bootstrap mode, so an unanchored T11 definition risks either an unrouted T8-G shell extension or operation-79 pressure. Bounded correction: one traceability sentence in §6 naming these anchors and stating that the mechanism's runtime realization belongs to the accepted non-serving operations concern class (T8-G private-surface family), never the serving wire. Not MATERIAL: the accepted concern class exists; only the citation is missing.

### F6 — NOTE — §2 "classified first as technical DEV/test estate" versus B0's proof-gated disposal

§2 presumes later-discovered estate is DEV/test "first"; B0 and §9 permit disposal only after proof of non-authority. These compose correctly — classification default is not disposal permission — but a careless reader could take §2's presumption as a disposal shortcut. Recorded so the Lead may align the wording; the operative disposal gates are fail-closed as written. Not a defect in the contract's behavior.

### F7 — NOTE — forward-obligation dispositions verified exact

§10 was cross-checked line-for-line against `decisions/forward-obligations.md`: MIG-05 PRESERVE (not activated — no migration exists), MIG-06 PRESERVE (no mode selected — no source corpus), MIG-10 REOPEN not triggered (no imported families). No DEFERRED platform idea is instantiated by transition convenience anywhere in the candidate. Recorded as conformance evidence for the operator.

---

## Attacked and NOT sustained (recorded so the Lead need not re-derive)

- **B0 wrongly discards real authoritative state.** B0's disposal rule is proof-gated ("marked disposable only after proving it is non-authoritative DEV/test state"), absence is a valid result, and concrete contrary evidence routes a bounded T7 reopen exactly matching `transition.md` §6 triggers. The presumption wording is F6 only. NOT SUSTAINED beyond F6.
- **B2 proof cannot occur before authority begins.** T9 proofs do commit Product-kind mutations, but the candidate correctly classifies them as disposable fixtures with a clean-baseline requirement and pre-B3 reset; the residual defect is the missing boundary event (F1), not an impossibility. Proof-before-authority is achievable. NOT SUSTAINED as an independent finding.
- **Hidden dual-write / reconciliation / ETL / compatibility machinery.** Searched every section: §3's explicitly-absent list is enforced by §6 (zero imported rows/history/content/sessions), §10 (MIG rows dormant), and the falsifier list; no accepted property requires any such machinery given T7. NOT SUSTAINED.
- **Blue/green or bridge rejection wrong.** No accepted continuity requirement exists that would justify a second Product-truth-committing system or a translation layer; T7 §3/§4 affirmatively reject the machinery both would need. Rejections are sound and cited. NOT SUSTAINED.
- **Binary/config rollback vs committed R10 state.** §5's compatibility-gated rollback law plus the candidate falsifier ("a post-B3 rollback deploy cannot safely consume current R10 state") cover the class; forward-only business truth is explicit. NOT SUSTAINED.
- **River/exact-content state crossing the authority boundary.** §8 matches T8-G §18 (shared PostgreSQL, coherent facts+intents recovery) and the accepted content-integrity/GC laws; unbound pre-B3 content stays disposable; queue state never Product truth. NOT SUSTAINED.
- **OIDC half-cutover creating competing authority.** §7 preserves AUTH-06 (no cross-system atomic transaction), provider config ≠ Product commit, no session import, no claim-derived authorization. Identity-side authority stays singular; the DNS/ingress residue is F3, not an identity defect. NOT SUSTAINED beyond F3.
- **Cleanup deletes recovery/provenance-critical resources.** §9's four preconditions include pre-B3-reversal and R10 recovery/audit/provenance retention; evidence-driven and monotonic. The one textual defect is the "pre-R10" qualifier (F3(b)). NOT SUSTAINED beyond F3.
- **T11/T12 leakage.** §12 hands off decomposition and readiness attack without performing either; no work graph, no implementation plan, no readiness evidence claims. NOT SUSTAINED.
- **Closed-envelope violation.** 78 operations, operation-79 absence, no new Permission/owner/Product state verified across both changed files; B3's examples use only accepted domain vocabulary. NOT SUSTAINED.
- **Ceremonial machinery (not-smallest direction).** Each barrier carries distinct load (source truth / preparation / proof / authority / serving); no barrier can be merged into a neighbor without losing an ordering property; the explicitly-absent lists subtract rather than add. Nothing ceremonial found. NOT SUSTAINED.

## Verdict

F1, F2 and F3 are bounded corrections of the two candidate files only (B2→B3 verified-clean + recorded authority-declaration event; B4 recovery-point precondition + zero-point truthfulness sentence; B4 legacy-fencing precondition + §9 qualifier removal). None requires reopening any T1→T9 authority, changing the 78-operation census, adding a sixth barrier, a Product state, an operation, a Permission or a semantic owner. F4–F5 are precision corrections at the Lead's discretion; F6–F7 are notes. Round 2 should be a bounded confirmation of F1–F3 uptake, not a full re-attack.

```text
VERDICT = NOT CONVERGED
MATERIAL findings = 3
Round 2 justified = YES
```
