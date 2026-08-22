# T11 — MetalDocs Bidirectional Frontend ↔ Backend Trace

> **TEMPORARY T11 CANDIDATE WORK / BRANCH-ONLY.** F7 proves complete correspondence between accepted Product/backend authority and the F1→F6 frontend implementation-readiness pack. It does not redefine operation semantics.

## 1. Closure claim

The frontend pack is coherent only if both directions hold simultaneously:

```text
Product/backend → frontend
  every accepted human goal has a UI home
  every application operation has an implementation tranche + consumer surface + wireframe
  every user-relevant concurrency/idempotency/exact-content behavior has a material interaction home

frontend → Product/backend
  every material screen traces to an accepted human goal
  every displayed truth traces to an admitted read model
  every material write traces to exactly one owner operation
  every navigation identity traces to server-returned truth
  every state belongs to the accepted T8-F state classes
```

## 2. Accepted human-goal closure — 16 / 16

| # | Accepted human goal | Route / surface | Wireframe(s) | Status |
|---:|---|---|---|---|
| 1 | Establish/end session | APP-01/APP-02 | WF-00/WF-01 | COVERED |
| 2 | Discover official documents | LIB-01 | WF-02 | COVERED |
| 3 | Create Document | LIB-02 → Document Work | WF-03 → WF-06/07/08 | COVERED |
| 4 | Inspect official/current Document truth | OFF-01/OFF-02 | WF-04 | COVERED |
| 5 | Start or enter open Revision | OFF-04 → DW route | WF-04 → WF-06/07 | COVERED |
| 6 | Author DRAFT | DW-01 | WF-07/WF-09 | COVERED |
| 7 | Upload replacement source | DW-02 | WF-08/WF-09 | COVERED |
| 8 | Submit / withdraw / cancel | DW-03/DW-04 | WF-07/WF-10 | COVERED |
| 9 | See actor-relevant work | WRK-01/WRK-02 | WF-11 | COVERED |
| 10 | Participate in governance | GOV-01/02/03 | WF-12/WF-13 | COVERED |
| 11 | Inspect Document history | HIS-01 | WF-14 | COVERED |
| 12 | Initiate/manage obsolescence | OFF-05 | WF-05 | COVERED |
| 13 | Administer Organization | ORG-01..08 | WF-16..19 | COVERED |
| 14 | Administer access | ACC-01/02 | WF-20/WF-21 | COVERED |
| 15 | Administer document governance | DGV-01..06 | WF-22..25 | COVERED |
| 16 | Inspect Audit | AUD-01 | WF-15 | COVERED |

No human goal required a new Product route, semantic owner or application operation.

## 3. Exact 78-operation forward trace

`Primary surface` is the implementation home; supporting consumption in another lens does not transfer authority.

| # | operationId | Tranche | Primary surface(s) | Wireframe / interaction home |
|---:|---|---|---|---|
| 1 | `getSession` | S1 | APP-01/APP-02 | WF-00/WF-01 |
| 2 | `endSession` | S1 | APP-02 | WF-01 |
| 3 | `searchProviderSubjects` | S1 | ORG-02/ORG-04 | WF-17/WF-18 |
| 4 | `getCompany` | S1 | ORG-01 | WF-16 |
| 5 | `replaceCompany` | S1 | ORG-01 | WF-16/WF-X4 |
| 6 | `listUsers` | S1 | ORG-02; admitted Admin selectors | WF-17/WF-20/WF-21/WF-23 |
| 7 | `createUser` | S1 | ORG-02 | WF-17/WF-X5 |
| 8 | `getUser` | S1 | ORG-02→user drawer | WF-18 |
| 9 | `getUserProfile` | S1 | ORG-03 | WF-18 |
| 10 | `replaceUserProfile` | S1 | ORG-03 | WF-18/WF-X4 |
| 11 | `deleteUserProfile` | S1 | ORG-03 | WF-18 |
| 12 | `getUserProviderBinding` | S1 | ORG-04 | WF-18 |
| 13 | `replaceUserProviderBinding` | S1 | ORG-04 | WF-18/WF-X4 |
| 14 | `getUserEligibility` | S1 | ORG-05 | WF-18 |
| 15 | `replaceUserEligibility` | S1 | ORG-05 | WF-18/WF-X4 |
| 16 | `listAreas` | S1 | ORG-06/07; admitted Admin selectors | WF-19/WF-21/WF-23 |
| 17 | `createArea` | S1 | ORG-06 | WF-19/WF-X5 |
| 18 | `getArea` | S1 | ORG-06 | WF-19 |
| 19 | `replaceArea` | S1 | ORG-06 | WF-19/WF-X4 |
| 20 | `getAreaLifecycle` | S1 | ORG-07 | WF-19 |
| 21 | `replaceAreaLifecycle` | S1 | ORG-07 | WF-19/WF-X4 |
| 22 | `listGroups` | S1 | ORG-08; ACC/DGV selectors | WF-19/WF-20/WF-21/WF-23 |
| 23 | `createGroup` | S1 | ORG-08 | WF-19/WF-X5 |
| 24 | `getGroup` | S1 | ORG-08/ACC-01 | WF-19/WF-20 |
| 25 | `replaceGroup` | S1 | ORG-08 | WF-19/WF-X4 |
| 26 | `deleteGroup` | S1 | ORG-08 | WF-19 |
| 27 | `listGroupMembers` | S1 | ACC-01 | WF-20 |
| 28 | `addGroupMember` | S1 | ACC-01 | WF-20 |
| 29 | `removeGroupMember` | S1 | ACC-01 | WF-20 |
| 30 | `listRoles` | S1 | ACC-02 | WF-21 |
| 31 | `listRoleAssignments` | S1 | ACC-02 | WF-21 |
| 32 | `createRoleAssignment` | S1 | ACC-02 | WF-21/WF-X5 |
| 33 | `deleteRoleAssignment` | S1 | ACC-02 | WF-21 |
| 34 | `listDocumentTypes` | S2 | DGV-01 | WF-22 |
| 35 | `createDocumentType` | S2 | DGV-01 | WF-22/WF-X5 |
| 36 | `getDocumentType` | S2 | DGV-02 | WF-22 |
| 37 | `replaceDocumentType` | S2 | DGV-02 | WF-22/WF-X4 |
| 38 | `getDocumentTypeGovernance` | S2 | DGV-03 | WF-23 |
| 39 | `replaceDocumentTypeGovernance` | S2 | DGV-03 | WF-23/WF-X4 |
| 40 | `getDocumentTypeEligibleTemplates` | S2 | DGV-04 | WF-24 |
| 41 | `replaceDocumentTypeEligibleTemplates` | S2 | DGV-04 | WF-24/WF-X4 |
| 42 | `getDocumentTypeNumberingPreview` | S2 | DGV-05; LIB-02 support | WF-22/WF-03 |
| 43 | `listTemplateConfigurations` | S2 | DGV-06/DGV-04 | WF-25/WF-24 |
| 44 | `getDocumentCreationOptions` | S3 | LIB-02 | WF-03 |
| 45 | `listDocuments` | S3 | LIB-01 | WF-02 |
| 46 | `createDocument` | S3 | LIB-02 | WF-03 → Work |
| 47 | `getDocument` | S3 base; S4-S6 enrichment | OFF-01 + Work resolver + T8-E-RO | WF-04/WF-05/WF-06 |
| 48 | `getDocumentResponsibleOwner` | S3 | OFF-03 | WF-05 |
| 49 | `replaceDocumentResponsibleOwner` | S3 | OFF-03 | WF-05/WF-X4 |
| 50 | `getDocumentTemplateRole` | S3 | DGV-06 concrete Document role | WF-25 |
| 51 | `replaceDocumentTemplateRole` | S3 | DGV-06 | WF-25/WF-X4 |
| 52 | `createDocumentRevision` | S4 | OFF-04 → Work | WF-04 → WF-06/07 |
| 53 | `getDocumentHistory` | S3 | HIS-01 | WF-14 |
| 54 | `listAuthoringWork` | S4 | WRK-01 | WF-11 |
| 55 | `listGovernanceWork` | S5 | WRK-02 | WF-11 |
| 56 | `getRevision` | S4 | DW-01/DW-04; History support | WF-07/WF-10/WF-14 |
| 57 | `getRevisionDraft` | S4 | DW-01 | WF-07/WF-09 |
| 58 | `updateRevisionDraft` | S4 | DW-01/DW-02 | WF-07/WF-08/WF-09 |
| 59 | `startRevisionDraftUpload` | S4 | DW-02 | WF-08 |
| 60 | `completeRevisionDraftUpload` | S4 | DW-02 | WF-08 |
| 61 | `getRevisionDraftSource` | S4 | DW-01 | WF-07/WF-08 |
| 62 | `createSubmission` | S4 | DW-03 | WF-07/WF-10/WF-X5 |
| 63 | `getSubmission` | S4 | DW-04; History/Governance support | WF-10/WF-14 |
| 64 | `getSubmissionSource` | S4 | DW-04/GOV subject/History | WF-10/WF-12/WF-14 |
| 65 | `withdrawSubmission` | S4 | DW-04 | WF-10 |
| 66 | `cancelRevision` | S4 | DW-01/DW-04 | WF-07/WF-10 |
| 67 | `getGovernanceAttempt` | S5 | GOV-01 | WF-12/WF-13 |
| 68 | `listGovernanceFeedback` | S5 | GOV-01 | WF-12/WF-13 |
| 69 | `createGovernanceFeedback` | S5 | GOV-02 | WF-12/WF-13/WF-X5 |
| 70 | `getGovernanceStepDecision` | S5 | GOV-03 | WF-12/WF-13 |
| 71 | `recordGovernanceStepDecision` | S5 | GOV-03 | WF-12/WF-13 |
| 72 | `getRelease` | S5 | OFF-02; History support | WF-04/WF-14 |
| 73 | `getReleaseSource` | S5 | OFF-02; History support | WF-04/WF-14 |
| 74 | `getOfficialRenditionContent` | S5 | OFF-02; History support | WF-04/WF-14 |
| 75 | `createObsolescenceRequest` | S6 | OFF-05 | WF-05/WF-X5 |
| 76 | `getObsolescenceRequest` | S6 | OFF-05; History support | WF-05/WF-14 |
| 77 | `withdrawObsolescenceRequest` | S6 | OFF-05 | WF-05 |
| 78 | `listAuditEvents` | S6 | AUD-01 | WF-15 |

Count proof:

```text
S1  operations  1..33 except none omitted          33
S2  operations 34..43                              10
S3  operations 44..51 + 53                          9
S4  operations 52 + 54 + 56..66                    13
S5  operations 55 + 67..74                           9
S6  operations 75..78                                4
------------------------------------------------------
TOTAL                                                78
```

```text
unassigned operations     0
multiply-owned operations 0
invented operations       0
operation 79              absent
```

## 4. Reverse trace — screen truth and controls

F2/F5 reconciliation:

```text
material surfaces                 36 / 36 → wireframe home
material write/navigation controls F6-traced
screen data blocks                F3 operation/read-model traced
navigation identities             F4 server-returned/refetched
client state classes              only T8-F server/URL/form/ephemeral classes
```

Reverse-trace prohibitions verified:

```text
screen data block with no admitted read truth                 0
material write with no owner operation                        0
material navigation target requiring guessed identity         0
History used as current-resource resolver                     0
Audit used as current-business resolver                       0
provider location/claim used as Product identity/authority    0
frontend lifecycle state machine competing with server        0
frontend Authorization evaluator                              0
parallel handwritten DTO/wire authority                       0
parallel global server-truth store                            0
```

## 5. Cross-cutting census closure

### Idempotent creations

```text
accepted          10
F6 controls       10
missing            0
extra              0
```

Exact set remains:

```text
createUser
createArea
createGroup
createRoleAssignment
createDocumentType
createDocument
createDocumentRevision
createSubmission
createGovernanceFeedback
createObsolescenceRequest
```

### ETag domains

```text
accepted read domains      13
accepted mutation domains  13
frontend contracts         13 / 13
```

No candidate/read enrichment is added to an ETag domain accidentally; T8-E-RO candidates remain outside ResponsibleOwner OCC.

### Exact-byte resources

```text
accepted exact-byte resources 4
frontend material consumers   4 / 4
```

```text
getRevisionDraftSource          → editable/read-only DRAFT context
getSubmissionSource             → submitted/governed/history read-only context
getReleaseSource                → official/history exact source
getOfficialRenditionContent     → official/history exact PDF
```

## 6. T9 Golden Flow implementation linkage

| Golden Flow | T11 implementation nodes | Frontend proof surfaces |
|---|---|---|
| GF1 Identity/session/access/revocation | P2/P3 + S1 | WF-00/01, WF-16..21, permission/session overlays |
| GF2 Governance config → atomic Document creation | S1/S2/S3 | WF-22..24 + WF-03 → Work |
| GF3 Revision authoring/upload/concurrency | S4 | WF-06..10 incl. upload-expiry + OCC reconciliation |
| GF4 Governance → Release/OfficialRendition | S4/S5 | WF-10/WF-12/13/WF-04 |
| GF5 Official discovery/obsolescence/disclosure | S3/S6 | WF-02/WF-04/05/WF-14 |
| GF6 Runtime failure/shutdown/recovery | P3/P4 | frontend only where failure reaches a user-safe-action state; primary proof E2/E5/E6 runtime subjects |

Golden Flow mapping does not turn the six flows into the complete 78-operation acceptance suite.

## 7. T9 cross-cutting linkage

| Validation property | Implementation owner | Frontend/readiness contribution |
|---|---|---|
| V1 wire census/generated/runtime conformance | P1/P5 | F7 exact 78/78 consumer trace; generated TS only |
| V2 closed-world dependency graph | P1/P5 | frontend package law + backend import graph remain accepted; no new owner |
| V3 AuthN/AuthZ/disclosure/CSRF | S1 + all protected slices | WF-00/01 + denied/notfound/CSRF laws; no frontend evaluator |
| V4 transaction + required Audit atomicity | P2 + semantic mutations | F6 identifies every material mutation path; browser proof never substitutes DB proof |
| V5 idempotency/replay | P2 + owning slices | exact 10 interaction rows + same-key ambiguous retry UX |
| V6 concurrency/ETag/serialization | P2 + owning slices | exact 13 OCC domains + WF-09/WF-X4 explicit reconciliation |
| V7 exact content/malware/rendition | S4/S5 | exact four byte resources + upload/admission/viewer flows |
| V8 durable work/River | S5 + P4 | rendition pending/Release visibility only; queue state absent from Product UI |
| V9 runtime readiness/failure/resource/observability | P3/P4 | dependency/integrity user-state overlays only where claim reaches browser |
| V10 backup/restore/privacy/security readiness | P4 | no serving frontend until runtime readiness; no Product restore controls invented |

## 8. Finding ledger / F8 classification

Frontend-readiness findings to date:

```text
F1 graph/coverage findings
  5 found
  5 corrected inside T11
  0 upstream semantic reopen

F3-F01 responsible-owner candidate discovery
  MATERIAL accepted journey/read asymmetry
  operator-approved bounded T6/T8-E/T8-F precision T8-E-RO
  operations added 0
  operation 79 absent

F4-F01 create successor target
  T11-local navigation mismatch
  corrected to createDocument → Document Work

F4-P01 arbitrary Library filter directory
  no accepted independent consumer
  deliberate YAGNI absence; use already-disclosed references
  reopen trigger recorded if real UX evidence proves insufficiency

F5
  new material findings 0

F6
  new material findings 0
```

```text
unresolved MATERIAL frontend-readiness finding  0
```

## 9. F9 frontend implementation-readiness closure

All required closure predicates now hold at candidate level:

```text
F0 authority baseline                         COMPLETE
F1 Coverage Matrix                            COMPLETE
F2 material surface inventory                 COMPLETE / 36 surfaces
F3 Screen Contracts                           COMPLETE / 36 READY after adjudication
F4 Navigation/Data Graph                      COMPLETE
F5 functional wireframes                      COMPLETE CANDIDATE / 36 surfaces covered
F6 Material Interaction Ledger                COMPLETE CANDIDATE
F7 bidirectional reconciliation               COMPLETE
F8 finding classification                     COMPLETE / 0 unresolved MATERIAL

human goals                                   16 / 16
application operations                        78 / 78
orphaned operations                            0
invented operations                            0
operation 79                                  absent
Idempotency-Key creations                     exact 10 / 10
ETag read / mutation domains                  13 / 13
exact-byte resources                           4 / 4
stable SPA Product routes                     exact accepted 10
frontend semantic owner                       none
frontend Authorization engine                 absent
parallel global server-truth store            absent
screen-shaped API                             absent
```

**Frontend Implementation Readiness = COMPLETE CANDIDATE.**

This does not authorize Product implementation and does not open T12.

## 10. Remaining pre-review obligations

Before the exact T11 candidate is eligible for operator approval to begin independent Fable review:

```text
1. feed F1→F9 results into the T11 Node Completion Contracts / implementation-program closure;
2. consolidate operator-approved T8-E-RO precision into the effective T6/T8-E/T8-F owning documents so no temporary second authority remains;
3. self-review the complete T11 pack for placeholders, contradictory counts, dead routes/controls and unowned proof obligations;
4. run required CI on the exact candidate HEAD;
5. present the exact candidate for explicit operator approval before creating review/t11-fable.
```

T12 remains NOT OPEN. Product implementation remains BLOCKED.
