# T9 Fable independent review — Round 2

> **Evidence only — non-authoritative. This review branch must never merge.**

## Review identity

```text
Repository                developmentconexus-ops/MetalDocs
Gate                      T9 — Golden Flows & Validation Baseline
Candidate branch          arch/t9-golden-flows
Exact corrected HEAD      eb7e0147cf575fe69290c231ea360af229917eeb
Required candidate CI     #1130 SUCCESS
Candidate Draft PR        #154
Round-1 Evidence PR       #155 CLOSED / UNMERGED
Round-1 verdict           NOT CONVERGED / MATERIAL=2
Round                     2 / BOUNDED
```

## Scope

This is a bounded confirmation round, not a fresh unconstrained redesign.

Read fresh:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ docs/work/current/t9-golden-flows.md
```

Then read only the exact owner needed to verify a challenged correction.

Confirm:

```text
F1 closure
  V1 now owns runtime execution of T8-E §9.4 wire-conformance fixture classes
  against the real composed E3 HTTP path, in addition to static census/schema/generated proof.

F2 closure
  GF1 now has causal negatives at the AuthN boundary:
  forged/replayed/tampered callback cannot issue a session;
  verified-but-unbound provider subject cannot obtain a session or auto-create a User.

F3–F6 boundedness
  session lifecycle negatives, GC/backup-pin race, T3 §15 Audit-census binding,
  and concurrent distinct Document-code allocation are precise accepted obligations,
  not new Product/runtime authority.
```

Also attack for regression:

```text
Golden Flows                         must remain 6
cross-cutting properties             must remain 10
evidence classes                     must remain 6
application operations               must remain 78
operation 79                          must remain absent
new Permission/semantic owner         none
T10/T11/T12 work                      none
Product implementation                none
```

Do not optimize for agreement. If a new MATERIAL contradiction is found, identify the exact accepted authority/property it falsifies and why Round-1 corrections did not already cover it.

Classify findings:

```text
MATERIAL
MINOR
NOTE
```

End with exactly:

```text
VERDICT = CONVERGED | NOT CONVERGED
MATERIAL findings = N
Round 3 justified = YES | NO
```

Write review Evidence below this line only.

---

## Fable Round 2 response

### 1. Identity facts re-verified this round

```text
candidate HEAD                     eb7e0147cf575fe69290c231ea360af229917eeb  (origin/arch/t9-golden-flows)
candidate required CI              run #1130 SUCCESS on that exact SHA (actions run id 32541821576)
review HEAD at round open          9fe42af3d8b19caed6cbe3c4dd9282a03ff0a6cf
review CI                          run #1131 SUCCESS on that exact SHA
Round-1 Evidence PR #155           CLOSED / UNMERGED / head 47483960 / base 2d5d127e
Round-2 Evidence PR #156           OPEN / DRAFT / head review/t9-fable-r2 / base arch/t9-golden-flows
review delta                       docs/work/current/ai-dialog.md only (1 file, additive)
candidate correction delta         2d5d127e → eb7e0147 touches exactly docs/roadmap.md
                                   + docs/work/current/t9-golden-flows.md; nothing else
```

The correction delta is bounded to the mutable status authority plus the temporary T9 work file. No durable T1→T8 architecture authority was edited by the adjudication commits. The roadmap keeps `T1 → T8-H CLOSED / OPERATOR-RATIFIED / INTEGRATED`, `IMPLEMENTATION BLOCKED`, `T10 → T12 NOT OPEN`; the T8-H detail it dropped remains owned by the immutable `docs/decisions/t8h-ratification.md`, still routed from the decision register.

### 2. F1 closure — CONFIRMED

Round-1 F1: runtime execution of the T8-E §9.4 wire-conformance fixture classes had no T9 owner.

Verified against the owner, `docs/architecture/wire-contract.md` §9.4:

- §9.4 exists, is titled *Runtime conformance contract*, closes the validator-split and fixture-class census, and states verbatim that *"Actual runtime execution of these fixtures belongs to the later validation/implementation program once a runtime exists."* That unowned obligation is exactly what V1 now owns.
- V1 now declares two mandatory lanes. The runtime lane requires execution of *"the accepted T8-E §9.4 wire-conformance fixture classes"* against the real composed path `transport → application → owners/mechanisms` and explicitly disqualifies direct owner calls — closing the exact bypass Round-1 named.
- The 13-item enumeration in V1 was cross-checked item-by-item against §9.4's fixture-class list; every item corresponds to an accepted §9.4 class (no invented class, no weakened class). The list is introduced as *"including at least"*, so the binding census remains §9.4 itself.
- §7 coverage law adds the matching closure line, and each fixture class requires a positive case plus a causal negative on the same production path — satisfying the §2 three-part validation law.

No new authority, stage, operation or runtime capability is created; V1 references §9.4 as the class source rather than duplicating it.

### 3. F2 closure — CONFIRMED

Round-1 F2: GF1 had zero falsifiers at the authentication boundary itself.

Verified against the owning authorities:

- GF1's success lane now inserts `callback/exchange validates the external protocol result → verified issuer+subject resolves one current ProviderSubjectBinding → exact enabled User` before session issuance, matching `docs/architecture/interfaces.md` §11 (the seam yields only *verified issuer string + verified subject string*) and `docs/architecture/backend.md` §7.3 / `runtime.md` (anti-corruption mapping; provider claims never escape as Product truth).
- Falsifier *forged/replayed/tampered callback — including invalid state/nonce/code/issuer — cannot create an ApplicationSession* attacks the protocol-verification subject the accepted seam already requires; it invents no protocol authority.
- Falsifier *verified provider subject with no current ProviderSubjectBinding cannot obtain a session or auto-create a User* is exactly the accepted admission model: the only User-creation path in the 78-operation census is the explicit `createUser` operation (*"CreateUser establishes enabled User + required profile + binding atomically"*, wire-contract §3.2). Login-time auto-provisioning exists nowhere in accepted authority, so the boundary fails closed.
- V3 is consistently extended with *"OIDC callback results cannot bypass binding/session admission"*, keeping the property singular rather than GF1-only.

### 4. F3–F6 boundedness — CONFIRMED

**F3 (session lifecycle).** Expiry, endSession-replay and binding-replacement revocation are accepted semantics, not new ones: `docs/product/journeys.md` §4 — the replacement transaction *"revalidates the expected current binding, replaces it, invalidates existing ApplicationSessions as required and appends Audit"*; `docs/architecture/persistence.md` — *"binding replacement DELETE all for User"*, which grounds GF1's absolute *"terminates all existing ApplicationSessions for that User"*.

**F4 (GC + backup pin).** V8's stale-cleanup-intent falsifier classes (current WorkingContent / immutable governed / claim-protected / backup-protected) map one-to-one onto `docs/architecture/persistence.md` §15 (AdmissionClaim and backup pins; GC lock-order law: no protective pin can appear behind GC's `FOR UPDATE` root lock) and `docs/architecture/content-integrity.md` (bounded backup pin/lease or equivalent GC exclusion during capture). V10's causal capture race and GF6's backup line exercise the same accepted mechanism. Nothing new is introduced.

**F5 (V4 census binding).** `docs/architecture/authorization-and-audit.md` §15 is the closed *Required same-local-commit Audit census*; V4 now names it as the sole enumeration source, and *"reconciled by later accepted T8-E precision"* matches wire-contract §8.2 (unreachable `provider_binding.disabled` removed — §15 already reads *accepted / replaced* only) and §8.4. The offboarding multi-event reconstructibility sentence restates §15's own law.

**F6 (concurrent code allocation).** GF2 now correctly separates same-logical-command retry (idempotent replay, no duplicate) from two distinct concurrent creates in one numbering scope (no shared code). `docs/architecture/persistence.md` §9 owns `numbering_scope = DOCUMENT_TYPE | DOCUMENT_TYPE_AREA` with the numbering counter taken under the DocumentType lock inside the create transaction, plus Document-code non-reuse; V6 adds the matching cross-cutting clause. Precise accepted obligation, no new Product/runtime authority.

### 5. Regression sweep

```text
Golden Flows                       6        (### GF count = 6)
cross-cutting properties           10       (### V count = 10)
evidence classes                   6        (E1–E6 table rows = 6)
application operations             78       (wire-contract §6 ledger 1→43 / 44→77 / Audit—78)
operation 79                       ABSENT   (asserted + attacked in GF5/V1; census closed)
new Permission                     NONE     (correction delta adds no Permission token)
new semantic owner                 NONE     (no owner/package/stage added)
T10/T11/T12 work                   NONE     (T10 appears only in the pre-existing reopen law)
Product implementation             NONE     (docs-only delta; roadmap keeps IMPLEMENTATION BLOCKED)
```

### 6. Findings

**F1-R2 — MINOR — §7 closure-line phrased in execution mode among plan-mode siblings.**
§7 requires for T9 closure: *"T8-E §9.4 runtime wire-conformance classes execute against the real composed HTTP path"*, while sibling lines require *"executable future proof plans"* and *"an identified production subject and falsifier"*. Since implementation is BLOCKED and §9.4 assigns actual execution to the later validation/implementation program once a runtime exists, a strictly literal reading would make T9 unclosable now. This falsifies no accepted authority and fails in the safe direction (over-strict closure, operator-visible at adjudication), so it is not MATERIAL and justifies no round; aligning the line to plan mode (e.g. *"have an executable proof plan targeting the real composed HTTP path"*) at the next natural edit suffices.

**F2-R2 — NOTE — V1 runtime-lane list is an illustrative floor, not the census.**
Several §9.4 classes (role bundles/scope matrix, stale-DRAFT always-412, upload/admission-boundary ceilings, package rejection, governance label snapshot, unknown-path 404) appear only through the binding sentence *"the accepted T8-E §9.4 wire-conformance fixture classes … including at least"*. The floor phrasing keeps §9.4 authoritative and creates no gap; traceability observation only.

**F3-R2 — NOTE — absolute revocation wording confirmed, no action.**
GF1's *"terminates all existing ApplicationSessions"* is stronger surface wording than journeys.md's *"invalidates existing ApplicationSessions as required"*, but persistence's *"binding replacement DELETE all for User"* fixes the accepted meaning as the absolute reading. Consistent.

No new MATERIAL issue exists: every Round-1 MATERIAL/MINOR correction is mechanically present, each grounds in the exact accepted owner named above, and the bounded regression envelope (6 / 10 / 6 / 78 / no-79 / no new authority / no T10+ / no implementation) holds on the exact corrected HEAD with required CI green.

### 7. Verdict

```text
VERDICT = CONVERGED
MATERIAL findings = 0
Round 3 justified = NO
```
