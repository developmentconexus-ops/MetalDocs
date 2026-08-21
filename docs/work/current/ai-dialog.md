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

> Independent Fable challenge of candidate `arch/t8f-frontend-realization @ a32ba8b58f5574336f825f46bd552dd96246de7f`. Authority reconstructed fresh: `AGENTS.md → docs/index.md → docs/roadmap.md → docs/architecture/frontend.md` + `docs/product/journeys.md` + `docs/architecture/wire-contract.md` + `docs/decisions/api-operation-census.md`, with `docs/product/contract.md` §J and T3 role bundles consulted for two specific findings. Evidence only; the Lead adjudicates.

### Mechanically re-executed closure claims

```text
coverage partition            §5 buckets extracted and diffed against seq 1..78
                              -> exactly 78 rows, exact set 1..78, zero missing, zero duplicate
operationId parity            §5 (number, operationId) pairs diffed against the ratified
                              T8-E ledger rows -> 78/78 exact match
operation 79                  absent; every "79" occurrence in the candidate is a prohibition
stable route set              candidate §4 ten routes == journeys.md §5 lines 209-232 verbatim;
                              zero added, zero removed, meanings preserved
family arithmetic             2+24+7+12+3+8+5+10+5+1+1 = 78 verified
permission vocabulary         §6.8-§6.11 use only organization.manage / access.manage /
                              document_type.manage / template_use.manage / audit.read
                              -> all inside the accepted 15-code T3 set; none invented
package topology              §13 tree ⊆ journeys.md §28 target vocabulary + admin split by the
                              three admin permission domains; no entities/services/stores layer
runtime/T8-G selection        §17 defers binaries/host/CDN/CORS/secrets/observability; React SPA,
                              TanStack Query, OpenAPI-generated TS are already ratified upstream
                              in journeys.md §28 and the T8-E probe, so naming them is not leakage
```

The coverage proof itself is arithmetically and nominally sound. The findings below are trace/realization breaks, not count breaks.

### Findings

All three MATERIAL findings are instances of one class: **a journey the candidate assigns to a lens requires an identifying read datum that no admitted operation or read model supplies purpose-built.** The candidate's §5 coverage proof holds in the inverse direction (every interaction → admitted operation), but the forward direction (journey → UX home with the data to render it) breaks at three points. Each break is exactly the class the candidate's own §15 names as a legitimate falsifier ("an admitted operation lacks the read data required to present its accepted decision safely").

---

**T8F-F1**

```text
severity: MATERIAL
claim: /documents/:document_id/work cannot be realized from its own URL key —
       no admitted read resolves document_id -> current open Revision.
owning authority implicated: T8-E wire read model (DocumentOfficialView, §3.5) as smallest owner;
       T6 route key choice (journeys.md line 215) is the origin of the obligation.
route/screen: /documents/:document_id/work (Document Work); /documents/:document_id (affordance)
operationId(s): 47 getDocument, 56 getRevision, 57 getRevisionDraft, 54 listAuthoringWork,
       52 createDocumentRevision, 53 getDocumentHistory
read model / state class: DocumentOfficialView / server state
```

Concrete counterexample. An authorized author bookmarks or refreshes `/documents/X/work`. The URL supplies `document_id` only; every §6.4 operation is keyed by `revision_id`. `DocumentOfficialView` is `{document, document_type, area, responsible_owner, status, official?}` — after first Release, `status` remains `effective` while an open successor DRAFT exists, so the view carries **no signal that open work exists and no revision reference**. The admitted resolution paths are all scan-based:

- `listAuthoringWork` (54) has no per-document filter (PAGED, `document.code, document_id` order) and its "actor-relevant" population is undefined upstream — an `area_manager` with `document.read_working` who is neither author nor owner may not be listed;
- `getDocumentHistory` (53) is ordered `occurred_at ASC`, so finding the newest `revision_created` requires walking every page of a long-lived document's history, then confirming via `getRevision`.

The same gap breaks the Document Official lens's own affordance decision: it cannot decide "Start next Revision" vs "Go to open work" from admitted reads; the only current-state signal is issuing `createDocumentRevision` and receiving `409 state.conflict`. The candidate's §6.4 vertical contract silently presupposes `revision_id` and §4 declares the route stable, so implementation would have to invent the resolver — precisely what T8-F exists to prevent.

Smallest correction (for Lead adjudication, ordered smallest-first):
(a) T8-F documents the concrete resolution mechanism it accepts (e.g. `listAuthoringWork` scan with its population semantics pinned upstream, or history-index + `getRevision` confirmation) together with its scale bound — no reopen, but the population ambiguity of `listAuthoringWork` must then be closed somewhere;
(b) bounded T8-E read-model precision: optional `open_revision?: RevisionIdentity (+ OpenRevisionState)` member on `DocumentOfficialView` with a closed presence law — **no new operation, census stays 78**; precedent is the operator-approved bounded read-symmetry precision already recorded in `api-operation-census.md`.

Reopens accepted authority: option (a) no (plus one upstream semantic pin); option (b) yes — bounded T8-E schema reopen, not a census/Product reopen. Operation 79 is **not** required and is not proposed.

---

**T8F-F2**

```text
severity: MATERIAL
claim: "render authorized navigation" (candidate §3 row 1; journeys §5 "Audit when authorized /
       Administration when authorized") has no admitted data source.
owning authority implicated: T8-F itself (realization statement missing); possibly journeys §5
       wording if the Lead reads "when authorized" as mandatory nav filtering.
route/screen: application shell navigation
operationId(s): 1 getSession
read model / state class: SessionView / server state
```

Concrete counterexample. `SessionView` is exactly `{user, csrf_token}`. No admitted operation exposes the current actor's effective permissions, roles, or lens usability: `listRoles` returns role definitions, `listRoleAssignments` is itself an `access.manage` admin surface, and `allowed_actions` exists only on `GovernanceCaseView`. So the shell cannot compute whether Audit or any Admin section is "authorized" for the actor. The candidate simultaneously (correctly) forbids a browser-maintained permission matrix (§14). The only realizations available are: always render nav and let route-level 403/404 Problems answer, or probe endpoints and use 403 as a permission oracle. Both are architecture decisions; the candidate names neither, while its own flow table promises "render authorized navigation". Implementation would have to rediscover this — the vagueness class the review contract flags.

Smallest correction: T8-F states the chosen realization explicitly — the smallest coherent one is *nav presence is not permission-filtered at Launch; entering an unauthorized lens yields the route-level `permission.denied`/`notfound` Problem state* — and rewords §3 row 1 accordingly. If the Lead instead reads journeys §5 "when authorized" as binding nav filtering, that is a bounded upstream contradiction (T6 promise vs T8-E census, which deliberately subtracted permission snapshots) and needs a targeted reopen decision, not silent repair.

Reopens accepted authority: no for the documentary correction; a journeys-§5 reading dispute would be a bounded T6 wording adjudication.

---

**T8F-F3**

```text
severity: MATERIAL (same class as F1)
claim: active ObsolescenceRequest discovery — "inspect/withdraw when authorized" at Document
       Official (candidate §3) cannot be rendered from admitted reads.
owning authority implicated: T8-E wire read model (DocumentOfficialView) as smallest owner;
       listGovernanceWork population semantics (undefined for initiators) secondarily.
route/screen: /documents/:document_id (Document Official); /work (My Work)
operationId(s): 75 createObsolescenceRequest, 76 getObsolescenceRequest,
       77 withdrawObsolescenceRequest, 53 getDocumentHistory, 55 listGovernanceWork
read model / state class: DocumentOfficialView, ObsolescenceRequestView, WorkGovernanceItem
```

Concrete counterexample. Human-governed obsolescence keeps the target EFFECTIVE during the attempt (contract.md §J; only human-governed requests have a withdrawal window). `DocumentOfficialView` carries no obsolescence reference and `DocumentOfficialStatus` remains `effective`, so the lens cannot show "obsolescence pending" or a withdraw affordance, and `withdrawObsolescenceRequest` needs a `request_id` no read at this lens supplies. The initiator returning in a fresh session discovers the active request only by: (i) attempting a second create and receiving `409` (no competing request law), or (ii) walking `getDocumentHistory` ASC pages for an `obsolescence_requested` item and confirming state via `getObsolescenceRequest` — read-model archaeology, not a purpose-built projection. Whether the initiator appears in `listGovernanceWork` ("actor-relevant") is undefined upstream, so My Work cannot be claimed as the home either. The candidate's §6.3 phrase "getObsolescenceRequest when referenced/currently relevant" is unsupported: nothing admitted *references* it.

Smallest correction: adjudicate jointly with F1 — the same bounded `DocumentOfficialView` precision can carry `active_obsolescence_request_id?: Uuid` under a closed presence law (no new operation), or T8-F documents the history-index realization and pins `listGovernanceWork` initiator semantics upstream.

Reopens accepted authority: same disposition as F1.

---

**T8F-F4**

```text
severity: MINOR
claim: §10.3 Problem map omits the 410 state.upload_expired family, and §6.4/§12 define no
       expiry-recovery law for the draft save chain.
owning authority implicated: T8-F only.
route/screen: /documents/:document_id/work
operationId(s): 58 updateRevisionDraft, 60 completeRevisionDraftUpload
read model / state class: form draft / editor boundary
```

`state.upload_expired` (410) is declared on operations 58 and 60; the shared 15-minute allocation TTL makes it reachable in ordinary stalled-network/suspended-tab saves. It changes the user's safe next action (restart at `startRevisionDraftUpload`; local bytes are preserved), which is exactly the property §10.3's own closing law protects. "At minimum" phrasing does not cover it because no listed row maps 410. Smallest correction: add the 410 row and one sentence in the §12 draft save law (re-allocate, re-upload same bytes, re-complete; never re-use the expired allocation). No reopen.

---

**T8F-F5**

```text
severity: MINOR
claim: §10.3 maps all "403 permission.*" to denied-action UX, but permission.csrf_failed has a
       different safe next action (session/CSRF re-bootstrap + retry), not permission denial.
owning authority implicated: T8-F only.
route/screen: application shell / transport
operationId(s): all unsafe operations; 1 getSession for recovery
read model / state class: SessionView / thin transport
```

A stale in-memory CSRF token (session re-established in another tab, long-lived tab after re-login) yields `403 permission.csrf_failed` on a user who *is* authorized. Presenting denied-action UX misinforms; the recovery is `getSession` → fresh `csrf_token` → retry the identical logical command (same Idempotency-Key where applicable — which the candidate's own retry law already permits). Smallest correction: split `permission.csrf_failed` from `permission.denied` in §10.3. No reopen.

---

**T8F-F6**

```text
severity: MINOR
claim: candidate §3 Audit flow promises "filter/page meaningful action evidence", but the
       admitted census forbids it: listAuditEvents accepts only cursor/limit ("No inferred
       Audit filter", wire §4).
owning authority implicated: T8-F flow table wording; T6/T8-E deliberately closed the surface.
route/screen: /audit
operationId(s): 78 listAuditEvents
read model / state class: AuditEventPage
```

Journeys promises only "AuditEvent inspection". A frontend "filter" over cursor-paged partial data would manufacture a filtered view the server never produced. Smallest correction: subtract the word "filter" from §3 (the §6.11 vertical contract is already consistent with the wire). If unfiltered company-wide DESC paging is later shown unusable at real scale, that is a T6 evidence-based reopen, not a frontend patch. No reopen now.

### Non-material observations (not findings)

- §5.7 groups two lenses (My Work + Document History supporting reads) in one coverage bucket; arithmetic is unaffected.
- §10.3 notation "409 state./conflict." does not match any wire family literally (codes are `state.conflict`, `state.governance_step_already_decided`); cosmetic.
- 400/413/415/429 rows are absent from §10.3; acceptable under "at minimum" since generic failure UX is safe for them (no `Retry-After` exists by contract).
- PROFILE_REPLACE recreation (`If-None-Match:*` after profile delete, 201 path) is only implicitly covered by §9's conditional-carriage line; a one-line §6.8 note would help implementers but nothing is contradicted.
- `getDocumentHistory` ASC ordering makes "newest event" access O(all pages); on its own this is an accepted T6/T8-E choice — it is material only as the workaround cost inside F1/F3.

### Strongest attacks attempted that did NOT survive

- **Coverage/count fraud**: partition and operationId parity re-executed mechanically — exact.
- **Route drift**: candidate §4 vs journeys §5 — verbatim identical ten routes; no noun-driven or collapsed route found; creation/admin-detail-as-panel is consistent with journeys' route minimalism.
- **Ownership transfer via co-location**: Admin sections keep the three permission domains distinct (§6.8–§6.10); Group identity stays Organization-owned under Access membership mutation — matches journeys §17.
- **T8-G leakage**: React/TanStack/generated-TS are ratified upstream (journeys §28, T8-E probes); §17 hands only consumers/constraints to T8-G; no router/deploy/CORS/provider freeze found.
- **Method/ontology contamination**: no unknown/partial/stale/converged vocabulary, no Operational Work owner, no Approval peer; §7/§2 bar them explicitly.
- **State-class smuggling**: Idempotency-Key (in-memory mutation scope) and ETag (transport metadata beside the cached representation) both fit the four accepted classes; no fifth durable class needed anywhere I could construct.
- **Optimistic-truth / cache-authority attacks**: §8 mutation law (replace exact representation, semantic-lens invalidation, no optimistic lifecycle) survives sign-out/governance/release scenarios I ran.
- **Editor boundary**: DOCX complete-bytes → allocate → verbatim `required_headers` PUT → complete → PATCH+If-Match chain matches wire §2.10 exactly; PDF gets no editor lifecycle; Governance content read-only; no EditorSession/lease/callback truth.
- **Second DTO/schema authority**: none; the thin transport's listed responsibilities are all wire mechanics, no semantic policy found hidden in it.
- **Idempotency durability attack**: page-reload key loss cannot double-submit `createSubmission` (DRAFT If-Match fingerprint) and matches the wire's deliberate "new deliberate command" semantics elsewhere — upstream-consistent, not a T8-F defect.

### Verdict

Material findings survive: **F1, F2, F3** (one shared class: missing purpose-built identifying reads for assigned UX homes), plus three MINOR corrections (F4–F6). None requires operation 79; none reopens Product scope, the 78-census, routes, ownership, or lifecycle. F1/F3's smallest upstream option is a bounded `DocumentOfficialView` member precision inside existing operation 47. The candidate is otherwise the smallest sustainable frontend realization I could not falsify. Lead adjudication requested; do not treat this Evidence as authority.
