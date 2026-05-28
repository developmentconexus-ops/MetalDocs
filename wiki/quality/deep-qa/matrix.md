# Documents + Approval Deep QA Matrix

Date: 2026-05-20
Status: active execution artifact
Canonical home: `wiki/quality/deep-qa/matrix.md`
Compatibility path: `wiki/references/documents-approval-deep-qa/matrix.md`
Proof policy: runtime + contract
Scope: scenario coverage artifact for canonical `/documents/:id`, `GET /api/v1/controlled-documents/{id}/active-document`, governed lineage in `documents`, approval submit/publish/supersede/scheduler behavior

This matrix remains scenario-shaped rather than procedural. It records what must be proved, which current boundary owns each scenario, and which evidence standard applies; `Status`, `Classification`, and `Artifact Links` stay blank until execution.

Each scenario must end in one of:
- `proved`
- `blocked`
- `deferred`

Each non-proved scenario must record:
- `classification`
- current owner boundary
- required next action
- artifact links

## Runtime truths this matrix assumes

- Governed history comes from `documents` lineage by `controlled_document_id`.
- `document_revisions` is technical autosave/artifact history only.
- `revision_title` belongs to `documents` and is born at finalize / submit-for-review.
- `GET /api/v1/controlled-documents/{id}/active-document` uses `documents.status` as the source of truth for `approvalState`.
- `approvalInstanceId` is enriched only for `under_review` and secondary lookup failures are explicit `500`.
- Published replacement lineage uses `documents.superseded_document_id` and scheduler cutover transitions old head to `superseded`.

## Scenario matrix

| ID | Category | Scenario | Owner Boundary | Preconditions | Action | Expected result | Evidence Standard | Status | Classification | Artifact Links |
|---|---|---|---|---|---|---|---|---|---|---|
| R1 | Route truth | Publish-only controlled document | `documents active-document route` | CD has latest `published`; no active sibling | GET `active-document` | 200 with `publishedDocumentId`; no `documentId`; no `approvalState`; no `approvalInstanceId` | runtime |  |  |  |
| R2 | Route truth | Draft sibling active | `documents active-document route` | CD has `draft` + prior `published` | GET `active-document` | 200 with active `documentId`, `approvalState=draft`, `publishedDocumentId`; no approval lookup side effects | runtime |  |  |  |
| R3 | Route truth | Under-review sibling active | `documents active-document route` | CD has `under_review` and single `in_progress` approval instance | GET `active-document` | 200 with `approvalState=under_review` and `approvalInstanceId` | runtime |  |  |  |
| R4 | Route truth | Approved sibling active before publish | `documents active-document route` | CD has `approved` + prior `published` | GET `active-document` | 200 with `approvalState=approved`; no `approvalInstanceId` | runtime |  |  |  |
| R5 | Route truth | Scheduled sibling active | `documents active-document route` | CD has `scheduled` + prior `published` | GET `active-document` | 200 with `approvalState=scheduled`; no `approvalInstanceId` | runtime |  |  |  |
| R6 | Route truth | Rejected sibling active | `documents active-document route` | CD has `rejected` + prior `published` | GET `active-document` | 200 with `approvalState=rejected`; no `approvalInstanceId` | runtime |  |  |  |
| R7 | Route failure | Under-review enrichment query fails | `documents active-document route` | Active row is `under_review`; DB error on approval lookup | GET `active-document` | 500 `INTERNAL_ERROR`; no silent omission | runtime |  |  |  |
| R8 | Route failure | No active or published lineage row | `documents active-document route` | CD has no documents | GET `active-document` | 404 `NO_ACTIVE_INSTANCE` | runtime |  |  |  |
| C1 | Canonical screen | Published doc starts revision | `canonical /documents/:id screen` | Viewing `/documents/:id` for `published` doc, no active sibling | Click `Iniciar revisao` | Composer opens with `Nome do documento` prefilled from current name | runtime+api |  |  |  |
| C2 | Canonical screen | Draft creation wording stays non-governed | `canonical /documents/:id screen` | Same as C1 | Inspect composer copy | UX explains working name only; no claim that governed title is being set | runtime+api |  |  |  |
| C3 | Canonical screen | Approved doc shows publish CTA | `canonical /documents/:id screen` | Viewing approved current revision | Open `/documents/:id` | `Publicar / Agendar` visible; no create-revision CTA | runtime+api |  |  |  |
| C4 | Canonical screen | Active sibling blocks duplicate revision create | `canonical /documents/:id screen` | Published doc with sibling in `draft` | Open `/documents/:id` | CTA becomes `Continuar rascunho`; no new revision creation path | runtime+api |  |  |  |
| C5 | Canonical screen | Active sibling under review | `canonical /documents/:id screen` | Published doc with sibling in `under_review` | Open `/documents/:id`; click CTA | CTA becomes `Acompanhar revisao`; navigates to sibling detail | runtime+api |  |  |  |
| C6 | Canonical screen | Active sibling approved | `canonical /documents/:id screen` | Published doc with sibling in `approved` | Open `/documents/:id`; click CTA | CTA becomes `Publicar revisao aprovada`; navigates to sibling detail | runtime+api |  |  |  |
| C7 | Canonical screen | Active sibling scheduled | `canonical /documents/:id screen` | Published doc with sibling in `scheduled` | Open `/documents/:id`; click CTA | CTA becomes `Ver publicacao agendada`; navigates to sibling detail | runtime+api |  |  |  |
| C8 | Canonical screen | Active sibling rejected | `canonical /documents/:id screen` | Published doc with sibling in `rejected` | Open `/documents/:id`; click CTA | CTA becomes `Retomar revisao rejeitada`; navigates to sibling editor | runtime+api |  |  |  |
| T1 | Revision-title lifecycle | First governed revision default | `documents finalize flow` | Draft has `revision_number=0`; user submits without title | POST finalize | Transition succeeds; persisted `revision_title='Criacao do documento'` | runtime |  |  |  |
| T2 | Revision-title lifecycle | Later governed revision requires title | `documents finalize flow` | Draft has `revision_number>=1`; blank title | POST finalize | 4xx domain error; document stays editable; no approval instance created | runtime |  |  |  |
| T3 | Revision-title lifecycle | Later governed revision with title | `documents finalize flow` | Draft has `revision_number>=1`; non-empty title | POST finalize | `revision_title` persisted on `documents`; status `under_review`; approval instance created | runtime |  |  |  |
| T4 | Revision-title lifecycle | Draft creation does not pre-write governed title | `documents create-revision flow` | Create revision from published screen | POST create revision | New draft row created without premature governed-title semantics | runtime |  |  |  |
| L1 | Lineage invariant | Single active sibling invariant | `documents lineage invariant` | Published head exists | Create revision twice concurrently | At most one active sibling survives; loser gets conflict/error | focused automated proof |  |  |  |
| L2 | Lineage invariant | Governed history source | `documents lineage read model` | Multiple revisions exist | GET revision history | Ordered lineage comes from `documents`; no autosave rows leak in | runtime |  |  |  |
| L3 | Lineage invariant | Supersede pointer persistence | `publish transition service` | Publish-now or scheduled replacement over existing head | Execute publish path | New replacement records previous head in `superseded_document_id` before cutover completes | contract/integration |  |  |  |
| L4 | Lineage invariant | Scheduler cutover final state | `scheduled publish jobs runtime` | Scheduled replacement reaches effective time | Run scheduler | Old head becomes `superseded`; scheduled row becomes `published`; pointer cleared or no longer actionable per runtime contract | runtime+api |  |  |  |
| O1 | OCC/concurrency | Double finalize same revision | `finalize OCC guard` | Two clients submit same draft with same revision version | Race finalize | One succeeds; other gets stale/conflict; no duplicate approval instances | focused automated proof |  |  |  |
| O2 | OCC/concurrency | Finalize vs autosave race | `editor autosave vs finalize invariant` | Draft open in editor while finalize submitted | Autosave after finalize | Server rejects unsafe write or keeps governed state stable; no status regression to `draft` | focused automated proof |  |  |  |
| O3 | OCC/concurrency | Publish-now vs second publish-now | `publish OCC guard` | Approved revision published twice concurrently | Race publish | One winner; no duplicated supersede/cutover transitions | focused automated proof |  |  |  |
| O4 | OCC/concurrency | Scheduler vs manual publish | `publish cutover concurrency boundary` | Scheduled doc is manually changed near cutover | Trigger scheduler around same time | Only one publish transition wins; lineage remains consistent | focused automated proof |  |  |  |
| P1 | Preconditions | Missing snapshot at finalize | `documents finalize preconditions` | Draft lacks required placeholder snapshot | POST finalize | Blocked before under-review transition with contract-appropriate error | runtime |  |  |  |
| P2 | Preconditions | Missing active publish context | `publish dialog preconditions` | Approved document but active-document lookup missing content hash | Open publish dialog | Publish action blocked with explicit UI explanation | runtime+api |  |  |  |
| E1 | Error UX | Create revision server error | `revision composer UX boundary` | API returns validation/conflict/internal error | Submit revision composer | Inline error shown; composer stays open; no false navigation | runtime+api |  |  |  |
| E2 | Error UX | Active-document route failure on canonical screen | `canonical screen plus active-document route` | `active-document` returns 500 | Open `/documents/:id` | UI degrades visibly and does not offer unsafe action based on missing truth | runtime+api |  |  |  |
| E3 | Error UX | Finalize blank revision title for REV01+ | `editor finalize UX boundary` | User omits title in editor modal | Submit | Error maps to actionable inline UX; no generic toast-only failure | runtime+api |  |  |  |
| A1 | Auth/authz | Unauthorized revision create | `revision create authz boundary` | User lacks create capability | Attempt create revision | 403/blocked UI; no draft created | runtime+api |  |  |  |
| A2 | Auth/authz | Unauthorized finalize | `finalize authz boundary` | User lacks `document.submit` | POST finalize | 403; no approval instance, no status transition | runtime |  |  |  |
| A3 | Auth/authz | Unauthorized publish/schedule | `publish and schedule authz boundary` | User lacks publish capability | Publish attempt | 403; approved row remains unchanged | runtime |  |  |  |
| A4 | Auth/authz | Cross-tenant route access | `tenant isolation contract boundary` | Wrong tenant context | GET `active-document`, finalize, publish | No data leakage; 404/403 per contract | contract/integration |  |  |  |
| S1 | Scheduler timing | Schedule in the past | `schedule-publish validation route` | Approved doc, invalid effective time | Schedule publish | Validation failure; no state transition | runtime |  |  |  |
| S2 | Scheduler timing | Minimum future time boundary | `schedule-publish contract boundary` | Approved doc, effective time exactly at minimum accepted threshold | Schedule publish | Deterministic accept/reject per contract; document state matches result | contract/integration |  |  |  |
| S3 | Scheduler timing | Multiple due rows | `scheduled publish jobs runtime` | Several scheduled revisions due simultaneously | Run scheduler | Each row processed independently; one failure does not corrupt others | contract/integration |  |  |  |
| H1 | Historical UX | Published page lineage display | `canonical /documents/:id screen` | Document has REV00..REV03 | Open `/documents/:id` | Timeline displays governed revisions and titles, not placeholder/mock versions | runtime+api |  |  |  |
| H2 | Historical UX | Current revision marker | `canonical /documents/:id screen` | Current row is published or approved | Open lineage timeline | Exactly one entry marked current | runtime+api |  |  |  |

## Execution order recommendation

1. Route contract and failure scenarios: `R1-R8`
2. Canonical screen behavior: `C1-C8`, `E1-E2`, `H1-H2`
3. Revision-title lifecycle: `T1-T4`, `E3`
4. Lineage + publish/scheduler invariants: `L1-L4`, `S1-S3`
5. Concurrency and OCC: `O1-O4`
6. Auth/authz sweep: `A1-A4`

## Evidence to capture per scenario

- HTTP request/response payloads, status, and error code
- Relevant DB rows before/after in `documents` and `approval_instances`
- For lineage scenarios: `controlled_document_id`, `revision_number`, `revision_title`, `status`, `superseded_document_id`
- For scheduler scenarios: effective timestamp, runner logs, final lineage state
- For canonical UI scenarios: screenshot or DOM assertion of CTA, label, inline error, and navigation target

## Stop conditions during QA

- Any contradiction between `/documents/:id`, `active-document`, and `documents.status`
- Any path that writes or infers governed history from `document_revisions`
- Any path that sets or mutates governed `revision_title` before finalize
- Any silent fallback from route/query failure into misleading UI state
- Any lineage state with two active siblings for one `controlled_document_id`

## Evidence standard legend

- `runtime`: runtime proof on the owning surface, typically API, logs, or direct state observation without requiring a paired canonical UI assertion
- `runtime+api`: canonical runtime UX proof plus the backing API or state proof for the same scenario
- `contract/integration`: focused contract or integration proof is acceptable when runtime setup is disproportionate or synthetic by nature
- `focused automated proof`: targeted automated proof for concurrency, OCC, or similarly unstable manual-reproduction scenarios
