# T8-F Fable independent review

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t8f-frontend-realization`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Candidate branch: `arch/t8f-frontend-realization`

Exact candidate HEAD under review: `a32ba8b58f5574336f825f46bd552dd96246de7f`

Review branch: `review/t8f-fable`

Gate: **T8-F — Frontend Realization**

Canonical Method: `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0

Repository Standard: `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` v1.0.0

### Fresh-actor route

Reconstruct authority independently:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ only the smallest owning authority pack needed for a finding
```

For this gate the expected bounded starting pack is the current T8-F candidate plus the already-ratified Product/frontend journeys and T8-E wire authority routed by the repository. Read any additional accepted T1→T8-D authority only when a concrete finding requires it.

Do not use this handoff as Product or architecture authority. Repository current authority wins.

Do not read removed implementation, legacy frontend code, closed PR chronology, or broad historical docs unless a specific material falsifier requires exact provenance.

### Candidate target

Adversarially challenge whether the candidate is the **smallest sustainable frontend realization contract** that makes the already-accepted MetalDocs Product/API human-operable without inventing architecture during React implementation.

The target must preserve the accepted closed application surface:

```text
78 application operations
operation 79 absent
stable T6 frontend route meanings
same-origin browser application baseline
implementation blocked
T8-G not open
```

The candidate deliberately uses a coverage-first Architecture-to-UX derivation:

```text
accepted Product/API authority
→ human-interaction coverage
→ actor goals / flows
→ screen + route realization
→ vertical traces
→ state / query / transport / read-model behavior
→ editor/viewer boundaries
→ package topology last
→ subtraction attack
```

The inverse proof is equally binding:

```text
material frontend interaction
→ admitted API operation
→ semantic owner
→ accepted Product capability/journey
```

A break in either direction is a finding. It must not be silently repaired with React business truth, a generic BFF, an ad hoc endpoint, a second DTO authority, or operation 79.

### Lead evidence already claimed by the candidate

Do not trust these claims; re-execute or falsify them where material:

```text
78 / 78 accepted application operations have concrete frontend consumers
0 orphaned accepted operations
0 invented application operations
operation 79 absent

stable SPA route meanings remain the exact accepted T6 set
route/screen count derives from journeys rather than API nouns

frontend is not a semantic owner
Admin co-location does not merge Organization / Access / Document Governance ownership
My Work is an actor-relevant projection, not an Operational Work owner
Governance remains current Product semantics; no Approval peer product is introduced

OpenAPI-generated TypeScript shapes feed one thin application transport boundary
no handwritten parallel DTO/schema/route/problem authority

TanStack Query is the server-state baseline
URL/navigation state, form draft, and ephemeral UI state stay separate
no Redux/Zustand/global server-truth mirror baseline

ETag stays bound to its exact representation/concurrency domain
stale DRAFT preserves local edits and requires explicit conflict handling
Idempotency-Key is one UUID per logical command and reused only for retries of that command
frontend branches on canonical Problem.code, not detail text

read models are consumed as purpose-built Product projections rather than normalized frontend entity truth
allowed_actions remain UX hints and never a client Authorization engine

exact-byte resources preserve Release / Submission / WorkingContent / OfficialRendition authority
one interactive DOCX adapter boundary supports edit/read-only modes
PDF remains read-only at Launch
editor output re-enters Product truth only through the accepted upload/admission + DRAFT If-Match path
no EditorSession/lease/provider callback truth is introduced

feature/package topology is semantic/lens-sliced and derived after flows
router library, design system, component library, deployment shape, provider choice and runtime mechanics remain deferred
```

Required CI on exact candidate HEAD `a32ba8b58f5574336f825f46bd552dd96246de7f` passed as run **#1035** before this review branch was opened. CI success is not evidence of architectural convergence.

### Review focus

Try to **falsify**, not confirm, the candidate. Attack at least these classes:

1. **Coverage truth**
   - Reconcile the candidate's 78-operation partition against the ratified T8-E ledger.
   - Look for an accepted operation with no legitimate human/application consumer, or a claimed consumer that exists only to make the count pass.
   - Look for a material accepted Product journey with no UX home even though all 78 operations are nominally assigned.
   - Look for any frontend interaction whose realization actually requires operation 79 or an unadmitted Product capability.

2. **Actor → goal → flow completeness**
   - Verify the candidate does not invent personas/roles as a second authority.
   - Look for machine/system transitions incorrectly surfaced as manual human actions.
   - Check whether create, author, submit, govern, withdraw/cancel, history, obsolescence, administration and Audit flows preserve their accepted business meaning.

3. **Screen / route derivation**
   - Attack whether the exact stable T6 route meanings are sufficient and non-duplicative.
   - Look for a materially different long-lived work context incorrectly collapsed into a panel/dialog merely to minimize route count.
   - Conversely, reject noun-driven routes or deep links with no accepted journey.
   - Verify a route never silently switches official vs DRAFT semantics by caller identity.

4. **Vertical trace / owner integrity**
   - For each material screen region/action, trace Product capability → semantic owner → read model/API → Permission/allowed_actions → frontend behavior.
   - Look for UI co-location that silently transfers ownership.
   - Verify Document History and Audit remain distinct truths.
   - Verify My Work remains a lens/projection and not a new work-domain owner.

5. **State authority**
   - Attack server-state / URL / form-draft / ephemeral-state separation.
   - Look for accepted semantics that require another durable/global client-state class and prove why the current four are insufficient before proposing one.
   - Look for cache behavior that could become lifecycle/current-state authority, especially after mutation, sign-out, offboarding, governance decision, release or obsolescence.
   - Look for forbidden optimistic semantic transitions or stale server truth presented as committed fact.

6. **Generated transport consumption**
   - Verify the generated OpenAPI TypeScript boundary can be consumed without a second handwritten DTO/schema/operation registry.
   - Attack whether the proposed thin transport has hidden semantic policy that belongs in Product/API owners.
   - Check session cookie/CSRF handling, ETag capture, Idempotency-Key reuse, exact bytes, pagination and Problem decoding.
   - Ensure the frontend does not parse provider capability URLs or provider identifiers as Product identity.

7. **Concurrency / idempotency / mutation behavior**
   - Verify ETags stay coupled to the exact representation and concurrency domain that produced them.
   - Attack stale DRAFT conflict UX for accidental silent overwrite/rebase/merge.
   - Verify whole-replacement retry semantics are not generalized to DRAFT PATCH.
   - Verify mutation retry rules do not create a second semantic action and do not generate a fresh Idempotency-Key for the same logical retry.
   - Look for over-broad cache invalidation or local patching that can manufacture truth not returned/revalidated by the server.

8. **Read-model consumption / Authorization**
   - Verify purpose-built read models remain projections, never write/current-state authority.
   - Verify `DocumentCreationOptionsView` is used instead of administrative directories for ordinary creation selectors.
   - Verify `allowed_actions` are hints derived from server truth, not a maintained client role matrix.
   - Verify hidden/disabled controls never substitute for server authorization.

9. **Editor / viewer boundary**
   - Attack the DOCX adapter boundary against exact-byte authority and DRAFT OCC.
   - Verify editable DOCX output is complete resulting bytes and can only become WorkingContent through allocate → provider PUT → complete → DRAFT PATCH + If-Match.
   - Verify PDF is not accidentally given an editor lifecycle.
   - Verify Governance displays exact immutable Submission content and cannot mutate WorkingContent.
   - Verify SourceOnly vs RequireOfficialRendition(PDF) presentation remains semantically distinct without duplicate frontend truth.
   - Identify any frontend mechanism that would force an EditorSession, lease, provider callback state or runtime decision; if truly required, route it to the smallest owning later stage rather than silently adding it.

10. **Method contamination / ontology leakage**
    - The planning method was imported from another product as a derivation/falsification method only.
    - Attack for leaked concepts such as generic Operational Work ownership, Approval as a peer product, external convergence state, provider-business readiness, or generic known/unknown/partial/stale ontology that current MetalDocs authority does not contain.
    - Conversely, do not remove a real MetalDocs semantic distinction merely because the generic method used different terminology.

11. **Package topology / Structural Inversion**
    - Verify the semantic/lens-sliced topology follows accepted frontend meaning rather than Go packages, tables, endpoints or legacy module taxonomy.
    - Attack whether `library`, `document-official`, `document-work`, `governance-work`, `history`, `audit`, and bounded Admin sections are the smallest coherent units.
    - Look for generic `entities`, `services`, `repositories`, `workflows`, global `domain`, or other familiar frontend architecture layers with no protected property/current consumer.

12. **T8-G leakage / hidden Writer choices**
    - T8-F must give T8-G concrete frontend runtime consumers without choosing T8-G decisions.
    - Flag router/framework version, deployment topology, provider CORS, renderer runtime, secrets/configuration, process topology, observability, recovery/runtime profiles or editor provider freeze if the candidate accidentally owns them.
    - Also flag frontend choices still left so vague that implementation would have to rediscover material architecture rather than realize it.

13. **YAGNI / Global Maximum**
    - Search for fields, abstractions, stores, route variants, caching rules, editor/session mechanics, generic state taxonomies or package layers with no named current consumer.
    - Prefer subtraction when all accepted properties survive.
    - Propose an additive correction only when a concrete Product/API/UX property cannot otherwise be realized.

14. **Targeted reopen discipline**
    - A legitimate frontend falsifier may reopen earlier authority; that is allowed.
    - Identify the *smallest owning authority actually implicated*.
    - Do not reopen T1→T8-E by preference, frontend convention, framework taste, legacy implementation shape or speculative future UX.
    - Operation 79 is material Product/T6 reopen and must never be inferred merely for convenience.

### Output contract

Write your independent review **only in this file** below `## Fable response`.

For each material finding, provide:

```text
ID
severity: MATERIAL | MINOR
claim
owning authority implicated
concrete counterexample/failure
smallest correction
whether it reopens an accepted authority
```

For a frontend-specific finding, also state the affected realization surface where useful:

```text
route/screen
operationId(s)
read model / state class / editor-viewer boundary
```

Separate non-material observations from findings.

Re-execute the strongest closure claims you can falsify mechanically, especially:

```text
coverage partition = exact 78 census rows, no missing/duplicate primary row
operation 79 absent
stable route set unchanged from accepted T6 meanings
no Product/API authority introduced by package/state organization
no runtime/T8-G mechanism selected
```

If no material finding survives, say so explicitly and list the strongest attacks attempted.

Do not edit the candidate branch or any other file. Reviewer output is Evidence; the Lead adjudicates every finding. Do not ratify T8-F and do not open T8-G.

## Fable response
