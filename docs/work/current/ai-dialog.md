# T9 Fable independent review

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t9-golden-flows`; this review branch must never merge.

## Review identity

```text
Repository                developmentconexus-ops/MetalDocs
Gate                      T9 — Golden Flows & Validation Baseline
Candidate branch          arch/t9-golden-flows
Exact candidate HEAD      2d5d127e95821eac355296e0a7f09c93aef6cef3
Required candidate CI     #1127 SUCCESS (verified live on that exact HEAD)
Candidate Draft PR        #154
Review branch             review/t9-fable
Review delta              docs/work/current/ai-dialog.md only
Round                     1
```

Fresh-actor route followed: `AGENTS.md` → `docs/index.md` → `docs/roadmap.md` → `docs/work/current/t9-golden-flows.md`, then only the owning T1→T8 authority needed for each concrete attack (`wire-contract.md`, T3 `authorization-and-audit.md`, T2 `lifecycle.md`, T4 `content-integrity.md`, T5 `async-and-search.md`, T8-G `runtime.md`, `api-operation-census.md`, `forward-obligations.md`, targeted sections of `interfaces.md` §11 and `persistence.md` §6). Exceeding the five-file default is the named material reason of this task: the candidate claims coverage of ten cross-cutting properties whose owners are distinct authorities; sufficiency cannot be attacked without reading each claimed owner.

Method: falsification-first. CI SUCCESS was treated as repository-envelope conformance only. The review goal is whether the candidate is the **smallest sufficient falsifiable composed-system validation contract** — both directions were attacked: missing failure classes (insufficient) and ceremonial/redundant elements (not smallest).

## Verified baseline (evidence, not trust)

```text
Golden Flows                     exactly 6 (GF1→GF6)
cross-cutting properties         exactly 10 (V1→V10)
evidence classes                 exactly 6 (E1→E6)
78-operation census              preserved verbatim in roadmap + candidate fixed input; V1 pins 78/0/0
operation 79                     absent; guarded twice (V1 + GF5 falsifier "cannot appear to repair a gap")
candidate delta vs main          docs/roadmap.md + docs/work/current/t9-golden-flows.md only; no durable
                                 architecture authority edited; no T10/T11/T12 content embedded
mock/fixture prohibition          §2 bans mock dependency success, self-proving fixtures, path-bypassing
                                 green tests; consistent with T8-G §25 "real mechanisms rather than mocks-only"
PDF→PDF                          GF4 variant + falsifier "cannot emit renderer intent merely for symmetry"
                                 + V7 same-PDF zero-transformation — exact match to T4-J/T5-D-E/wire §8.5
DOCX fidelity corpus             GF4 variant + falsifier + V7 + GF6 startup/conformance row — matches T8-G §9
malware binding                  GF4 "clean scan evidence must bind the exact bytes later admitted" — matches T4-G/T4-H
idempotency/replay               V5 = wire §2.5 exactly (scope, fingerprint, live AuthZ before replay disclosure,
                                 ReplaySnapshot, changed-request rejection, 24h semantic expiry)
ETag/concurrency                 V6 + GF2/GF3 falsifiers match wire §2.4 including DRAFT-specific 412
transaction+Audit atomicity      V4 + GF2 "required Audit append failure rolls back" + GF4 — matches T3 §12
River                            V8 + GF4 redelivery falsifier + GF6 terminal-visibility row — matches T5-M
runtime blast radius             GF6 fault matrix reproduces T8-G §14/§25 rows faithfully
backup/restore/privacy/security  V10 + GF6 restore rows match T4-L/M/N and T8-G §18
forward obligations              §6 dispositions match decisions/forward-obligations.md (AUTH-02/06, ASY-04,
                                 DB-01..07/10, SEC-01 PRESERVE; CNT-03/AUD-06/MIG-10 REOPEN; DEFERRED untouched)
T9 discipline                    no invented operation/state/Permission/owner/runtime capability found anywhere
                                 in the candidate; reopen law routes failures to smallest owner
```

No redundant or ceremonial Golden Flow was found: GF1 (identity/access), GF2 (config→atomic create), GF3 (authoring/upload/OCC), GF4 (governance→release/rendition), GF5 (discovery/disclosure/obsolescence), GF6 (runtime/failure/recovery) protect disjoint accepted property clusters; overlaps with V-properties are flow-composition realization versus property law, which the coverage law (§7) makes proportional rather than duplicated. The census/composition split (§4: flows are a composition basis, not an operation acceptance suite) is the correct smaller-stronger shape — subject to F1 below.

---

## Findings

### F1 — MATERIAL — "full 78-operation surface closure" is assigned to V1, but V1 as written is a static census/schema proof; runtime wire-conformance ownership is a gap between the contract proof and the composed-flow proof

**Claim.** The candidate's coverage pillar "Operation-level contract/census conformance plus representative composed flows" (§7) is unfalsifiable as written, because no flow or property owns **runtime execution** of the accepted per-operation wire behavior for operations the six flows never exercise.

**Evidence.**

1. Candidate §4: "Full 78-operation surface closure is proved separately by V1." Candidate §7: "Operation-level contract/census conformance plus representative composed flows is the smaller and stronger architecture proof."
2. V1's must-prove list is: census counts (78/0/0/79-absent), "OpenAPI paths/operationIds/schemas/headers/Problems **match accepted wire law**", and generated Go/TypeScript projections compile/cannot widen. Its falsifier is census/schema mutation only ("a deliberately added/removed/renamed operation or incompatible schema"). Every subject in V1 is mechanically inspectable **without a running composed application**: the OpenAPI document and generated projections.
3. `docs/architecture/wire-contract.md` §9.4 defines the required negative/edge **runtime** fixture classes (strict decoding, duplicate members, bodyless-body rejection, 405 exact `Allow`, media/coding/65,536-byte ceiling, PROFILE_REPLACE full conditional matrix, cursor tamper/filter replay, complete options arrays, Content-Digest == body SHA-256, Range rejection, corrupt-bytes 500 with zero success bytes, ReplaySnapshot ≤ 2048, etc.) and then explicitly hands them off: "Actual runtime execution of these fixtures **belongs to the later validation/implementation program** once a runtime exists." T9 is that program.
4. The six flows behaviorally exercise a subset of the 78 operations. Operations/behavior with **no** T9 flow or property owner include, at minimum: `deleteUserProfile` (absent→404 law), `replaceUserProviderBinding` (revokes all sessions — see F2/F3), `deleteGroup` dependency-409 fail-closed law (T3 §8), the `PROFILE_REPLACE` If-Match/If-None-Match matrix, 405 exact `Allow`, cursor tamper + per-page AuthZ recheck, `listDocuments status=obsolete|cancelled` requiring `document.read_history` else 403, and the complete non-truncated creation-options arrays.

**Why it falsifies sufficiency.** Candidate §2 itself rules "presence of a file, dependency or config key without executing its property" insufficient. Under the current text a T9 prover can close the census pillar with a document-diff plus compile proof — exactly the static-inspection class §2 bans for behavioral claims — while the wire's own conformance contract (§9.4), which T8-E explicitly deferred to this stage, has no T9 owner. A runtime that violates the wire on any non-flow operation would leave the composed-system validation contract green.

**Smallest correction (candidate file only; no upstream reopen).** Precise V1 (or split its second half) to own, in addition to the census/projection proof: *runtime wire-conformance execution of the T8-E §9.4 fixture classes against the real composed E3 path (transport → application → owners), with positive + causal negative cases per class*. Tool identity stays non-authoritative (Schemathesis-class property attack may realize the lane). This keeps the property count at 10 (precision of V1's scope, not a new property) and keeps flows at 6.

### F2 — MATERIAL — GF1 has zero falsifiers at the authentication boundary itself; the first link of the identity chain is closable by a permissive adapter

**Claim.** GF1's success lane begins "external OIDC authentication → ProviderSubjectBinding resolves exact User → HttpOnly ApplicationSession issued", but all five of GF1's required falsifiers attack **post-authentication** properties (claims→Permission, CSRF, revocation staleness, OIDC outage, disclosure). No causal negative attacks the AuthN entry boundary — the only internet-facing trust boundary in the accepted topology.

**Evidence.**

1. `interfaces.md` §11 (T8-C): `ExchangeAuthorizationCode(...) -> verified issuer string, verified subject string`; "Session issuance requires protected current enabled-User truth from Organization"; provider claim bags never cross as truth. `runtime.md` §8 (T8-G): OIDC "does not establish MetalDocs User eligibility / Permission / …"; the app "performs the **required** callback/exchange/**validation**".
2. GF1 required falsifiers (candidate lines 99–107): none of them makes the flow fail when the callback/exchange itself is attacked. Missing causal negatives, both grounded in accepted invariants:
   - a forged/tampered/replayed authorization callback (invalid or replayed code, broken state/nonce binding, wrong issuer) **must not** produce an ApplicationSession — this is the executable meaning of "verified issuer/subject" and "required validation";
   - a provider-verified subject with **no current ProviderSubjectBinding** must fail closed: no session, no auto-provisioned User — this is the executable meaning of "ProviderSubjectBinding resolves exact User" and T8-G §8.
3. Candidate §2 bans "mock/fake success for an external dependency", but E5 evidence for GF1 as specified can be satisfied by a happy-path real-provider login. A permissive identity adapter that accepts any exchange result would pass GF1's success lane **and all five listed falsifiers** (they all execute after a session exists).

**Why it falsifies sufficiency.** The candidate's own global law (§2, clause 3) demands a causal negative per claim; GF1's operative falsifier list omits the negative for the chain's first two claims. Session issuance to the wrong or unbound principal is the highest-consequence composed failure the flow exists to disprove, and it is currently unfalsified.

**Smallest correction (candidate file only; no upstream reopen).** Add two required falsifiers to GF1: (a) forged/replayed/tampered OIDC callback (state/nonce/code/issuer) cannot produce an ApplicationSession; (b) a verified provider subject without a current ProviderSubjectBinding cannot obtain a session or create a User. Flow count stays 6. No new Product state or operation is implied — both negatives attack the existing `/auth/callback` boundary.

### F3 — MINOR — session-lifecycle revocation falsifiers (expiry, endSession, binding replacement) are not named

**Claim.** GF1's revocation falsifier covers access-administration staleness but not the accepted session-lifecycle terminations.

**Evidence.** `persistence.md` §6 `authn.application_sessions`: `expires_at … CHECK(expires_at > created_at)`; "logout/revoke DELETE; offboard DELETE all for User; **binding replacement DELETE all for User**". Wire `endSession` clears the cookie via `SESSION_END`. No T9 element requires proving: an expired session fails 401; a session cookie replayed after `endSession` fails; `replaceUserProviderBinding` terminates all live sessions for that User.

**Why MINOR not MATERIAL.** GF1's falsifier 3 ("revoked/disabled current access cannot remain usable through stale browser/server cache") plus V3's "current access is revocable" plausibly subsume these under a rigorous prover; the gap is enumeration precision, not an unowned class. Bounded correction: add the three lifecycle negatives to GF1 or V3.

### F4 — MINOR — the destructive managed-content reclamation class is covered only by classification inference; the backup-pin/GC race has no named falsifier

**Claim.** GC is the only accepted code path that physically deletes provider content, and the candidate never names it.

**Evidence.** T4-K/T5-J require re-prove-before-delete; T4 §19: "GC cannot delete current WorkingContent or immutable governed content", "stale cleanup intent cannot delete after eligibility/reference/claim changes", "**backup/GC race cannot lose selected DRAFT content before capture**" (the T4-L backup pin/GC-exclusion property); T5 §19: "GC recheck prevents current/governed/claim-protected/backup-protected content deletion". In the candidate, V8 covers "each activated named durable-effect class" with "current-state revalidation before effect/finalization" — GC is in the T5-B durable-effect census, so the recheck law is reachable **by inference** — but neither GF6's fault matrix nor §4's negative-control exemplars nor V10 mention GC, and the backup-capture race is required by no falsifier (V10 as written proves restore-time completeness, not capture-time protection).

**Why MINOR not MATERIAL.** A faithful reading of V8 against the T5-B census does bind GC revalidation, and a V10 completeness proof executed adversarially would surface a raced backup; the defect is that §7 requires an *identified* subject and falsifier per property and the identification is missing. Bounded correction: name GC explicitly as a V8-covered effect class and add "backup pin/GC-exclusion race cannot lose selected required content before capture" to V10 (or GF6's backup row).

### F5 — MINOR — V4's "every representative transition class" should bind to the closed T3 §15 census

**Claim.** V4's enumeration authority is ambiguous. T3 §15 is a **closed, finite** same-local-commit Audit census; "representative transition class" invites selection that could omit a required class (e.g. `user.reenabled`, `user_profile.erased`, offboarding multi-event teardown reconstructibility), whereas the census makes "all classes, representative instances per class" both smaller and fully falsifiable. Bounded correction: reference T3 §15 (as reconciled by wire §8.4) as the class enumeration source in V4.

### F6 — MINOR — concurrent duplicate Document-code allocation is not a named falsifier

**Claim.** T2 §16's first proof obligation — "concurrent code allocation cannot create duplicate committed Document codes" — is distinct from GF2's covered falsifier (idempotent duplicate of the *same* logical create). Two *different* concurrent logical creates in one numbering scope must not commit the same DocumentCode. V6 lists "one-open/effective uniqueness" but not code uniqueness under real concurrent transactions. Bounded correction: add the cross-command numbering-uniqueness negative to GF2 or V6.

### F7 — NOTE — frontend read-symmetry validation is claimed by the roadmap but unnamed in the candidate

The roadmap states "The bounded T8-E-FR read-symmetry meaning … T9 validates it." The candidate carries no named element for it. Coverage exists implicitly (the two precision GETs are V1 census rows; their ETag domains fall under V6; E4 lanes consume them), but a one-line traceability note in V1 or GF1/GF2 would close the roadmap claim explicitly. No sufficiency defect.

### F8 — NOTE — rate limiting is correctly absent

`ratelimit.exceeded` 429 appears in every operation's Problem set, but no accepted T1→T8 authority mandates a concrete limiter mechanism/invariant. The candidate rightly adds no rate-limit property; recorded so its absence is not mistaken for a coverage gap in later rounds.

---

## Attacked and NOT sustained (recorded so the Lead need not re-derive)

- **Ceremonial-flow attack.** Tried to collapse GF2 into GF3 (both exercise ETag+idempotency) and GF5 into GF4 (both touch obsolescence): fails — GF2 uniquely owns configuration-truth→creation atomicity and preview non-authority; GF5 uniquely owns disclosure-safe derived routing and History/Audit non-authority. Tried to demand a seventh flow (migration/cutover): correctly out of scope — T10 owns transition; candidate §8 routes transition-only problems there. NOT SUSTAINED.
- **PDF→PDF renderer-intent leak.** GF4 falsifier + V7 + T9's own falsifier line exactly enforce wire §8.5/T4-J/T5-D-E: no renderer intent/copy/duplicate bytes for already-PDF. NOT SUSTAINED.
- **DOCX fidelity corpus.** GF4 mandatory variant + falsifier ("representative corpus fidelity failure blocks production eligibility for that renderer profile") + GF6 envelope row match T8-G §9's proof-gated reference exactly, including re-run on renderer/font/config change being an eligibility (not per-job) property. NOT SUSTAINED.
- **Idempotency post-expiry race.** Wire §8.4(9)'s one-winner post-expiry reuse path is reachable through V5's "24h semantic expiry" + V6's narrow-serialization law. Enumeration could name it, but coverage exists. NOT SUSTAINED as a finding beyond F5's general precision.
- **Late-rendition no-op for terminated candidates.** Covered by V8 "current-state revalidation before effect/finalization" + GF4's redelivery falsifier, matching T5 §6/T2 §12. NOT SUSTAINED.
- **Offboarding vs governance-action serialization (T3 §11).** Covered by V6 "protected current eligibility during correctness-critical transitions" + GF1 revocation falsifier. NOT SUSTAINED.
- **Restore barriers.** V10 + GF6 rows cover exact-content verification, session invalidation before serving, privacy erasure and security-teardown reconciliation, matching T4-M/N and T8-G §18 including fail-closed with no bypass. NOT SUSTAINED.
- **Search second-authority attack.** GF5 falsifier ("Search cannot acquire a second materialized/external current-truth authority") matches T5-F/H including the canonical-baseline posture; no T9 element forces materialization into existence. NOT SUSTAINED.
- **Evidence-class attack.** E1–E6 partition is minimal and each class carries an explicit cannot-be-replaced-by column consistent with §2; "no class is mandatory when the protected property does not depend on it" prevents ceremonial E2E duplication. NOT SUSTAINED.
- **Forward-obligation drift.** §6 dispositions cross-checked against `decisions/forward-obligations.md`: no obligation is silently activated, no DEFERRED counterexample is instantiated, and the three stage-relevant REOPEN rows are consciously consumed. NOT SUSTAINED.
- **Scope-creep attack.** Searched the candidate for T10/T11/T12 content, implementation permission leakage, operation 79 pressure, new Permissions/owners/runtime capability: none present; §1/§8/§9 fence them explicitly. NOT SUSTAINED.

## Verdict

```text
VERDICT = NOT CONVERGED
MATERIAL findings = 2
Round 2 justified = YES
```

F1 and F2 are bounded precisions of `docs/work/current/t9-golden-flows.md` only (V1 runtime-conformance ownership; GF1 AuthN-boundary falsifiers). Neither requires reopening any T1→T8 authority, changing the 78-operation census, adding a seventh flow or an eleventh property. F3–F6 are enumeration precisions at the Lead's discretion; F7–F8 are notes. Round 2 should be a bounded confirmation of the F1/F2 uptake (plus any F3–F6 uptake), not a full re-attack.
