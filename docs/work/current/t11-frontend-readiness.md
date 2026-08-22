# T11 — Frontend Implementation Readiness

> **TEMPORARY T11 CANDIDATE COMPANION / BRANCH-ONLY WORK.** This file derives implementation-readiness detail from accepted Product/T6/T8-E/T8-F authority. It is not durable authority and must be absorbed or removed before T11 integration. Current stage/status/implementation permission remains exclusively in `docs/roadmap.md`.

## 1. Purpose

T8-F already closes frontend semantic realization. T11 closes the next question:

```text
Can an implementer build every material MetalDocs screen/interaction
without inventing Product meaning, guessing a backend relationship,
creating screen-shaped API, or deciding material UX behavior while coding?
```

This is implementation-readiness precision, not permission to silently reopen accepted authority. Material evidence routes explicitly to the smallest implicated owner.

## 2. Fixed envelope

```text
stable SPA route meanings             exact accepted T6/T8-F set
application operations                78
orphaned operations                   0
invented operations                   0
operation 79                          absent
frontend semantic owner               none
frontend Authorization engine         absent
parallel DTO/API authority             absent
parallel global server-truth store     absent
React SPA                              accepted
TanStack Query                         accepted server-state mechanism
generated TypeScript wire projection  required
interactive DOCX boundary              one adapter boundary
Product implementation                 BLOCKED
```

Stable Product routes remain exactly:

```text
/documents
/documents/:document_id
/documents/:document_id/work
/documents/:document_id/history
/work
/work/governance/:attempt_id
/audit
/admin/organization
/admin/access
/admin/document-governance
```

`/auth/login` and `/auth/callback` remain browser AuthN integration routes outside the 78-operation application census.

## 3. Method

```text
F0 accepted authority recovery
→ F1 Frontend Coverage Matrix
→ F2 material interaction-surface inventory
→ F3 Screen Contracts
→ F4 Navigation/Data Graph
→ F5 functional low/mid-fidelity wireframes
→ F6 Material Interaction Ledger
→ F7 bidirectional frontend↔backend trace
→ F8 finding classification / smallest justified reopen
→ F9 frontend readiness closure
```

Reusable idea from Marketplace Central: **coverage before drawing**. Marketplace-specific Product ontology is not imported into MetalDocs.

## 4. F0 — Authority baseline — COMPLETE

Planning input remains accepted human goal/journey, owner, route/lens, exact reads/writes, read models, Authorization/disclosure, concurrency/idempotency/exact-content behavior and T9 proof obligations.

## 5. F1 — Frontend Coverage Matrix — COMPLETE

Artifact:

```text
docs/work/current/t11-frontend-coverage.md
```

Result:

```text
accepted human goals mapped                 16 / 16
stable route/lens homes                     covered
T8-F frontend consumer reconciliation       78 / 78
T11 implementation assignment               78 / 78 exactly once
operation 79                                absent
material T11 findings                       5 found / 5 adjudicated
unresolved MATERIAL finding                 0 at F1 closure
```

F1 corrected only the open T11 implementation graph.

## 6. F2 — Material interaction-surface inventory — COMPLETE

Artifact:

```text
docs/work/current/t11-frontend-surfaces.md
```

Result:

```text
stable SPA Product routes               10
material frontend surfaces              36
accepted human goals with surface       16 / 16
new Product routes                       0
application operations changed          0
operation 79                             absent
new material gap at F2                  0
```

A surface is a coherent user decision context, not a URL/component. Cosmetic loading, responsive variants, success toasts, ordinary validation copy, page number and component boundaries do not create surfaces by themselves.

## 7. F3 — Screen Contracts — BLOCKED ON ONE MATERIAL FINDING

Artifact:

```text
docs/work/current/t11-screen-contracts.md
```

Progress:

```text
F2 surfaces                              36
Screen Contracts derived                 36 / 36
READY from current accepted authority    35
BLOCKED                                  1 — OFF-03 Responsible Owner / F3-F01
new application operation required       0
operation 79                             absent
```

Each contract closes, as claim-relevant:

```text
human goal + stable route/lens
semantic owner + exact reads/writes
generated schemas/read models
source of every target identity
TanStack/URL/form/ephemeral state class
CSRF / If-Match / Idempotency-Key / cursor / exact-byte mechanics
material lifecycle and Problem branches
mutation cache/refetch/navigation consequence
current server Authorization/disclosure
browser/composed proof
explicit non-authority
backend sufficiency
```

### F3-F01 — responsible-owner candidate discovery

Later responsible-owner replacement is an accepted user journey requiring a selectable target satisfying:

```text
existing User + same Company + ENABLED
```

The normal `document.owner.manage` holder is `area_manager`, which does not have `organization.manage`. Current `ResponsibleOwnerView` supplies only the current owner + ETag. Creation options expose candidates only inside active creation-eligible context, while later replacement has no active Area/DocumentType precondition. Therefore current wire cannot provide a complete least-privilege candidate selector for every valid owner-replacement state.

Rejected repairs:

```text
Admin listUsers                         privilege/semantic coupling
manual opaque UUID                     unusable human interaction
creation options as universal truth    wrong lifecycle/eligibility contract
candidate list inside ETag Owner view  cross-domain concurrency pollution
operation 79                            unnecessary larger change
```

Recommended smallest correction:

```text
existing operation 47 getDocument
→ non-ETag DocumentOfficialView gains optional derived
   responsible_owner_candidates?: UserReference[]
```

Presence/meaning:

```text
present only when caller may receive the Document context
AND current document.owner.manage is ALLOW for that Document

when present:
  complete current D4 eligible set = existing ENABLED same-Company Users
  order = user_id ASC
  server revalidates target eligibility on replace

absent:
  no inference about authorization or candidate existence
```

The wire already accepts complete non-truncated responsible-owner candidate arrays in `DocumentCreationOptionsView` and specifies `user_id ASC`; if measured scale later makes complete arrays unsustainable, that is the existing T6 scale reopen trigger rather than reason to add pagination/op79 now.

This is the same class of bounded read precision already used for `DocumentOfficialView.open_revision` and `active_obsolescence_request_id`: derived browser-necessary read truth, no new Product capability/owner/route/operation.

Smallest implicated accepted authority appears to be:

```text
T6 Product journeys  → clarify later owner-management candidate presentation
T8-E wire            → optional field + presence/order/disclosure fixtures
T8-F frontend        → OFF-03 consumes the field
```

Existing T8-C owner/application composition can gather Organization-owned User truth and map it after canonical Authorization; no new semantic owner or cross-owner SQL is required. T9's existing proof classes can exercise disclosure/eligibility without adding a GF/V class.

**Accepted authority is not edited without explicit operator approval of this bounded correction.**

## 8. F4 — Navigation / Data Graph — NOT OPEN WHILE F3-F01 UNRESOLVED

When opened, every transition must prove:

```text
source surface
→ accepted source of target identity/reference
→ target stable route
→ target initial read operation
→ absent/non-disclosable/unauthorized behavior
```

Navigation may not use History, Audit, provider identity, DOM state or client-only inference as current-resource truth.

## 9. F5 — Wireframe law — NOT OPEN

Wireframes begin only after F3 and F4 are coherent. They are functional low/mid-fidelity implementation contracts, not visual-brand authority.

They freeze interaction meaning: hierarchy, material data, actions, forms/dialogs, navigation, lifecycle/denied states, OCC reconciliation, idempotent retry where user-visible, upload/admission/recovery and exact-content editor/viewer mode. They do not freeze ornamental pixels without material reason.

## 10. F6 — Material Interaction Ledger — NOT OPEN

Every material wireframe control later receives one row covering owner, operationId, input source, wire mechanics, success truth, material failures, client consequence, retry law and forbidden inference.

## 11. F7 — Bidirectional trace

Final closure must prove:

```text
Product/backend → every accepted goal/op/read model/material behavior has frontend home
frontend → every surface/data block/write/navigation/state traces to accepted backend authority
application operations = 78/78
orphaned = 0
invented = 0
operation 79 = absent
```

## 12. F8 — Findings / reopen law

```text
frontend placement/presentation only
  → T11 correction

missing trace precision but backend truth exists
  → Screen Contract/trace correction

accepted frontend mapping contradictory
  → smallest T8-F reopen

accepted human goal not safely representable from wire/read models
  → smallest Product/T6/T8-E owner reopen

convenience endpoint without Product need
  → reject

operation 79 proposal
  → hard STOP unless material Product/T6/T8-E reopen proves necessity
```

**Hard no screen-shaped API.** Conversely, accepted user behavior is never forced into frontend guessing merely to preserve a count by convenience.

## 13. F9 — Frontend closure

Requires completed F1→F7, all findings adjudicated, 78/78 reconciliation, no invented authority, operation79 absent unless an explicit upstream material reopen changes the canonical census, and no unresolved MATERIAL finding.

## 14. Implementation-node linkage

Future semantic tranches implement the completed T11 frontend pack against real generated client/server paths. No S node closes as “backend complete, frontend follows later”.

P5 later proves no drift from reviewed T11 Screen Contracts/wireframes.

## 15. Current status

```text
F0 authority baseline                  COMPLETE
F1 Coverage Matrix                     COMPLETE
F2 material surface inventory          COMPLETE / 36 surfaces
F3 Screen Contracts                    36/36 DERIVED / 35 READY / 1 BLOCKED
F3-F01 responsible-owner candidates    MATERIAL / OPERATOR ADJUDICATION REQUIRED
F4 Navigation/Data Graph               BLOCKED by F3-F01
F5 wireframes                          NOT OPEN
F6 Material Interaction Ledger         NOT OPEN
F7 bidirectional reconciliation        PARTIAL via F1→F3 / final not executed
F8 finding classification              ACTIVE
F9 frontend readiness closure          NOT REACHED
```

The method is working as intended: material backend/frontend ambiguity is surfaced before drawing or Product code.

## 16. Promotion law

This file is temporary review structure. If T11 converges, its binding implementation-readiness outcome is absorbed into durable T11 authority and the reviewed screen/wireframe pack is retained only in the minimum repository location required by T12/future implementation. Temporary method/work files are removed before integration.