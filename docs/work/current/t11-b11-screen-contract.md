# T11 — B11 Access Administration — P9 Screen Contract

> **Status:** COMPLETE / PASS / R3 POST-LOCK PROOF.  
> **Current P8 locked package:** `docs/work/current/t11-b11-access-administration-p8.html`.  
> **Exact R3 locked blobs:** HTML `ea20912e5259f4f3f51df7ce09ee3f2e5cfc7540`; CSS `9ce012007613777187ae70956c2bfa09e7066c16`; JavaScript `670ff9b905d94014ff27698e2a23c868316030a4`.  
> **Operator LOCK:** `docs/work/current/t11-b11-p8-r3-operator-relock.md`.  
> **Latest reopen Evidence:** `docs/work/current/t11-b11-final-challenge-r2.md`.  
> **Reopen Evidence:** `docs/work/current/t11-b11-final-challenge-r1.md`.  
> **Implementation:** BLOCKED.

## 1. Goal and proof boundary

Bind every material region and control in the operator-LOCKED clean B11 P8 to accepted Product, architecture, read and write authority in both directions.

P9 is proof, not redesign. The HTML/CSS/JavaScript package is a functional low-fidelity Evidence carrier. Its in-browser data and transport simulator makes the contract inspectable; it is not a proposed production client store, a hidden collection crawl, or Authorization authority.

No contradiction was found that invalidates the P8 LOCK. If one is later found, only the smallest affected P7/P8 or upstream owner may reopen.

## 2. Contract vocabulary

Every row below binds the P9 fields required by the Frontend Product Experience Planning Method:

```text
GOAL / FLOW                   → Goal / role
ROUTE / SURFACE               → Surface
INFORMATION ROLE              → Goal / role
OWNER + READ TRUTH            → Owner / read truth
WRITE CONTROL                 → Control / write
IDENTITY SOURCE               → Identity
CLIENT STATE CLASS            → State
WIRE MECHANICS                → Wire
MATERIAL FAILURES             → Failure
FAILURE MESSAGE INTENT        → Message intent
SUCCESS CONSEQUENCE           → Success
AUTHZ / DISCLOSURE            → AuthZ / disclosure
FORBIDDEN FRONTEND AUTHORITY  → Forbidden authority
BACKEND SUFFICIENCY           → Sufficiency
```

Client-state abbreviations used in the matrices:

```text
S = server state
N = navigation / presentation state
D = unaccepted form draft
E = ephemeral UI / Evidence-fixture state
```

## 3. Core region/control binding

| ID | Surface | Goal / information role | Owner / current read truth | Material control / write | Identity | State |
|---|---|---|---|---|---|---|
| B11-01 | `/admin/access` | Stable Access Administration home | T8-F accepted route | Route entry; no mutation | Route authority | N |
| B11-02 | Global shell and Quick Inbox | Preserve accepted MetalDocs navigation and notification utility | B01 + B01N LOCKs | Inherited navigation only | Accepted shell identities | N/S |
| B11-03 | `Por Área / Grupos / Funções` | Switch among the three human inspection jobs | B11 operator LOCK | Local lens selection only | Closed local lens key | N |
| B11-04 | Boundary note | Explain canonical configuration versus complete effective access | T3 Authorization + B11-F1 | None | Accepted access-model terms | N |
| B11-05 | Area selector | Select a disclosed Area or the real Company scope presentation | Organization op16 `listAreas` | Selection only | `area_id`; Company is `scope_kind=company`, never a synthetic Area | S/N |
| B11-06 | Area-scoped assignments | Inspect grants scoped exactly to selected Area | B11-F1 op31 filtered by `scope_kind=area&area_id=A` | Exact revoke; contextual grant entry | `assignment_id`, enriched subject, `AreaReference` | S/N |
| B11-07 | Company-wide assignments beside Area | Show canonical Company grants separately because they also apply across the Area | B11-F1 op31 filtered by `scope_kind=company` | Exact revoke; contextual grant entry | `assignment_id`, enriched subject, real Company scope | S/N |
| B11-08 | Area page traversal | Expose every returned page in the selected Area slice | op31 seek-page law | Next; Previous returns to a prior simulated window in P8 and to retained returned state in production | P8 fixture page index; production server cursor plus bounded visited-window state | S/N |
| B11-09 | Company page traversal | Expose every returned page in the Company slice | op31 seek-page law | Next; Previous returns to a prior simulated window in P8 and to retained returned state in production | P8 fixture page index; production server cursor plus bounded visited-window state | S/N |
| B11-10 | Area contextual grant | Reduce wrong-scope recomposition | P7 accepted composition + op32 | Open composer with selected real scope preselected | Exact `area_id` or Company scope | N/D/E |
| B11-11 | Group selector | Select one existing disclosed Group | Organization op22 `listGroups` raw page | Selection only | `group_id` | S/N |
| B11-12 | Group direct access footprint | Inspect every direct RoleAssignment to selected Group across Company and Area scopes | B11-F1 op31 filtered by `group_id=G` | Exact revoke; contextual grant | `group_id`, `assignment_id`, enriched scope | S/N |
| B11-13 | Group footprint traversal | Traverse the canonical filtered Group slice | op31 seek-page law | Next; P8 prior-window / production retained-page Previous | Cursor bound by server to original Group filter | S/N |
| B11-14 | Group member list | Inspect current direct memberships | Organization op27 `listGroupMembers(G)` | Remove entry | `group_id + user_id` | S/N |
| B11-15 | Group member traversal | Traverse direct membership pages in `user_id ASC` order | op27 seek-page law | Next; P8 prior-window / production retained-page Previous | Server cursor and selected `group_id` | S/N |
| B11-16 | Add-member User picker | Inspect a raw op6 `UserPage` and choose one enabled User | Organization op6 `listUsers`, `user_id ASC`, with no eligibility/membership filter | Selection only | `user_id`; server-returned `eligibility` | S/D |
| B11-17 | Add-member review | Restate exact User, Group and security-bearing consequence | Raw selected User + selected Group + current footprint | Confirm/cancel | Exact `group_id + user_id` | D/E |
| B11-18 | First membership relation | Add one exact User→Group relation | op28 `addGroupMember` | PUT relation | `group_id + user_id` | D/S/E |
| B11-19 | Existing membership reconciliation | Reconcile an enabled User that is already related without requiring complete member knowledge | op28 `addGroupMember` | Same PUT relation | Same `group_id + user_id` | D/S/E |
| B11-20 | Remove-member review and command | Remove one exact direct relation while preserving residual-access caveat | op29 `removeGroupMember` | Confirm then DELETE relation | Exact `group_id + user_id` | D/S/E |
| B11-21 | Group contextual grant | Reduce wrong-subject recomposition | P7 accepted composition + op32 | Open composer with exact Group preselected | `group_id` | N/D/E |
| B11-22 | Role selector | Inspect fixed Product Role vocabulary | op30 `listRoles`, fixed T3 order | Selection only | `RoleCode` | S/N |
| B11-23 | Role detail | Explain fixed meaning and allowed scopes | Server `RoleView.permissions + allowed_scope_kinds` | None | `RoleCode`, `PermissionCode` | S/N |
| B11-24 | Assignments by Role | Inspect canonical use of selected Role | B11-F1 op31 filtered by `role=R` | Exact revoke | `assignment_id`, `RoleCode` | S/N |
| B11-25 | Role-assignment traversal | Traverse the canonical filtered Role slice | op31 seek-page law | Next; P8 prior-window / production retained-page Previous | Cursor bound by server to original Role filter | S/N |
| B11-26 | General grant entry | Start one deliberate additive grant | op32 plus supporting reads | Open composer; no command on open/cancel | New local draft identity only | D/E |
| B11-27 | Grant User subject | Choose an existing User without changing the op6 page | Raw op6 `UserPage`, no eligibility filter | Select enabled User; disabled User remains visible and unavailable | `user_id`, server-returned `eligibility` | S/D |
| B11-28 | Grant Group subject | Choose an existing disclosed Group | Raw op22 `GroupPage` | Selection only | `group_id` | S/D |
| B11-29 | Grant Role | Choose one fixed Role with inspectable meaning | op30 `RoleListView` | Selection only | `RoleCode` | S/D |
| B11-30 | Grant scope | Choose Company or exact Area admitted by the Role | op30 allowed scopes + op16 Areas | Selection only | `scope_kind` + optional `area_id` | S/D |
| B11-31 | Grant final review | Restate exact Subject × Role × Scope before mutation | Accepted command model | Confirm/cancel | Normalized command fingerprint | D/E |
| B11-32 | Initial create success | Establish one additive canonical grant | op32 `createRoleAssignment` | POST with fresh UUID `Idempotency-Key` | Key + fingerprint; returned `assignment_id` | D/S/E |
| B11-33 | Completed create replay | Prove same successful command is stable | Global completed-replay law + op32 | Replay same key + same fingerprint | Same key, fingerprint, success status and `assignment_id` | S/E |
| B11-34 | Ambiguous post-commit retry | Recover an unknown transport outcome without a duplicate semantic mutation | Global Idempotency-Key law + op32 | Freeze composition/review/confirm/close; retry only the same logical command with the same key | Same key + same normalized fingerprint | D/S/E |
| B11-35 | Exact revoke | Revoke only one canonical assignment | op33 `deleteRoleAssignment` | DELETE exact assignment | `assignment_id` | S/E |
| B11-36 | Failure/recovery fixtures | Make materially distinct safe next actions inspectable | Global Problem contract + operation-local errors | Retry, reload or correct only where safe | Affected cursor/resource/command identity | E |
| B11-37 | Responsive and keyboard structure | Preserve semantics and deliberate security actions on narrow/keyboard use | B01/B01N + Frontend Method structural constraints | Tabs, drawers, dialogs, list/detail, focus return | Same domain identities as wide layout | N/D/E |
| B11-38 | P8 fixture transport simulator | Exercise returned pages, mutations and replay consequences in-browser | This locked Evidence carrier only | Fixture controls and deterministic simulated responses | Fixture IDs mirroring accepted shapes | E |

## 4. Wire, safety and sufficiency binding

| IDs | Wire mechanics | Material failure + message intent | Success consequence | AuthZ / disclosure | Forbidden frontend authority | Backend sufficiency |
|---|---|---|---|---|---|---|
| 01–04 | Safe route/read composition; B11 adds no shell operation | 403 is denial, never successful empty data; explain bounded configuration truth | Operator reaches one coherent administration home | Visibility is not authorization; `access.manage` remains server-enforced | New shell, route family, Notification truth, effective-access engine | Existing shell, route and accepted access model are sufficient |
| 05 | First op16 page may carry filter inputs; continuation is `cursor` + optional `limit` only | 400 means request composition is invalid; 403 denies; 404/current disappearance reconciles owner truth | Exact Area or Company lens becomes context | Supporting Organization read does not transfer identity ownership | Synthetic Company Area; hidden seek to an unreturned Area | op16 plus Company scope kind are sufficient |
| 06–07 | Two independent op31 first-page filters; filters run server-side before pagination; regions stay separate | Failure in one slice preserves the other loaded slice and names which truth is unavailable | Canonical Area and Company rows retain their actual scope | Disclosure-safe enriched references are recognition data, not mutable Authorization identity | Merged fake scope, client post-filter, inferred effective permissions | Refined op31 is sufficient; no global matrix/read needed |
| 08–09, 13, 25 | Production Next sends only returned `cursor` and optional `limit`; production Previous may revisit only retained returned page state. P8 proves the interaction with numeric fixture windows, not cursor-token storage | Failed continuation keeps the currently loaded canonical page visible and offers same-page retry | One more canonical server page replaces the current window; an already visited window may be revisited | Cursor authentication preserves original filter truth | Treating P8 fixture indexes as production offset authority; repeating filters; total; hidden crawl | Global cursor law + filtered op31 are sufficient |
| 10, 21, 26 | Opening/canceling composer is local; supporting reads occur only as needed | Missing contextual identity blocks unsafe preselection and reconciles current owner truth | A deliberate draft opens with only known real identity preselected | Preselection is convenience, never authorization | Context becoming immutable authority or an implicit mutation | Existing supporting reads + op32 are sufficient |
| 11 | op22 raw seek page, `group_id ASC`; continuation is cursor + optional limit | Failed continuation preserves page and names Group-directory failure | Selected disclosed Group drives exact reads | Absent/non-disclosable identity is not distinguished beyond contract | Authorization-owned Group registry or hidden global search | op22 is sufficient |
| 12 | `GET role-assignments?group_id=G`; server filters before pagination | 403 denies; 404/current identity loss reconciles Group truth; cursor failure preserves page | Every traversed row is one direct canonical Group grant | No expansion through members into per-User access | `Group.area_id`, effective-permission expansion, incomplete post-filter | Refined op31 is sufficient |
| 14–15 | op27 first page under exact Group; `user_id ASC`; continuation is cursor + optional limit | Parent absent/non-disclosable → 404; continuation failure preserves page | Current direct member window is inspectable | Member list is direct relation truth only | Calling one loaded page the complete membership set | op27 is sufficient |
| 16 | op6 returns raw `UserPage`, `user_id ASC`, no eligibility filter; picker preserves page boundaries | Disabled User remains visible with eligibility guidance and unavailable selection; read failure preserves current page | Enabled User may be selected whether membership is known, unknown or already present | Browser does not decide current membership or mutation eligibility | Eligibility-prefiltered directory, membership-based disablement, hidden crawl | Raw op6 + server-side op28 reconciliation are sufficient |
| 17–19 | Review precedes PUT op28. `201` means first relation; `204` means relation already existed | 409/current eligibility or offboarding conflict asks reload/review; 404 is disclosure-safe owner loss; no “already a member” client guess | 201 adds one semantic relation; 204 confirms canonical relation with zero new semantic mutation | Requires `access.manage`; server serializes security-bearing races | Complete membership cache, browser eligibility decision, duplicate semantic add | op28's exact 201/204 contract is sufficient |
| 20 | Review precedes DELETE op29; 204 includes already-absent relation when parent exists | 404 means absent/non-disclosable Group; message never promises hidden existence or total access removal | Exact relation is absent; residual direct/other-Group grants may remain | Requires `access.manage`; disclosure remains server-owned | “Remove all access”, effective-access recomputation | op29 is sufficient |
| 22–23 | op30 safe read returns fixed role order, permissions and allowed scopes | Failure blocks local fallback/custom role fiction | Operator can understand one fixed Role and admissible scopes | Role explanation is not a User authorization decision | Client Role registry/editor, editable permission matrix | op30 is sufficient |
| 24 | `GET role-assignments?role=R`; server filters before pagination | Denial is not empty; cursor failure preserves returned page | Canonical assignments for selected fixed Role are inspectable | Direct assignment configuration only | Role-derived membership semantics or effective matrix | Refined op31 is sufficient |
| 27 | Raw op6 first/continuation pages; no eligibility filter | Disabled User visible/unavailable; failures preserve page and draft | One enabled `user_id` enters the draft | Final server mutation rechecks current truth | Client eligibility registry or filtered “eligible users” endpoint | Raw op6 is sufficient |
| 28 | Raw op22 pages | Failure preserves page and draft; absent identity reconciles owner truth | One exact `group_id` enters the draft | Supporting read is not Authorization ownership | Generic principal registry | op22 is sufficient |
| 29–30 | op30 roles + op16 areas; allowed scope kinds guide composition | Missing read truth blocks fabrication; op32 still validates | Exact Role and scope enter normalized draft | Guidance is not authorization | Custom role/scope compatibility engine | op30 + op16 are sufficient |
| 31 | No mutation until explicit final confirmation | Correctable validation preserves draft; cancel produces zero mutation | Exact normalized command becomes intentional | Consequence copy does not claim final effective permissions | Implicit edit/replace or silent draft mutation | Accepted command model is sufficient |
| 32 | POST op32 uses session + CSRF + fresh UUID key and normalized semantic fingerprint; success is `201 CreateRoleAssignmentResult` | 409 current-state conflict; 422 validation/key reuse; safe draft retained for deliberate correction | Exactly one RoleAssignment and required same-commit Audit; terminal success remains inspectable in dialog | Requires `access.manage`; server rechecks current Authorization before a new mutation | Browser grant authority, auto-new key after unknown outcome | op32 is sufficient |
| 33 | Same key + same fingerprint within replay window returns exact stored success status/body; historical lifecycle/preconditions are not rerun | Different fingerprint with same key → `422 validation.idempotency_key_reused` and explicit correction path | Same `201` result and same `assignment_id`; semantic mutation/Audit count remains 1→1 | Current denial may mask completed replay only according to accepted replay/AuthZ ordering; fixture does not invent disclosure | Treating replay as a second success mutation or silently changing key | Global replay law + op32 are sufficient |
| 34 | Unknown post-commit outcome freezes Subject/Role/Scope, review, confirm and dialog close; the only resolution is same key + same fingerprint retry | Ambiguity message says outcome is unknown, preserves command identity, blocks fresh review/close, and offers same-command retry | Stable stored success recovered; same `assignment_id`; mutation count remains 1→1; close unlocks only after resolution | No duplicate grant is inferred from transport uncertainty | Fresh key/new review on ambiguous outcome, duplicate POST command, discarding unresolved identity | Global replay law + op32 are sufficient |
| 35 | DELETE op33 exact `assignment_id`; first revoke 204, absent 404 | 404 reconciles current disclosed truth; copy preserves residual-access caveat | Only selected assignment is absent | Requires `access.manage`; exact identity prevents broad revoke | Delete+create disguised as atomic edit; “all access removed” | op33 is sufficient |
| 36 | Deterministic fixtures operate `/admin/access` 403, selected-Group 404, op28/op32 409, continuation failure and ambiguous transport. 400/422 remain P9 contract mappings, not claimed operable P8 scenarios | Message names whether to retry, reload, correct, or stop; generic toast cannot erase decision state | Safe recovery is inspectable without changing authority | Error disclosure follows operation contract | Fixture output becoming Product truth | Existing Problem and operation contracts are sufficient |
| 37 | Semantic HTML, roving/arrow tabs, labeled controls, focus-managed dialog/drawer, narrow stacking | No hover-only material action; errors remain associated and visible | Same information and action consequence survive keyboard and narrow viewport | Layout never changes AuthZ/disclosure semantics | Mobile simplification that hides scope/consequence | Existing frontend structural authority is sufficient |
| 38 | Simulator applies accepted filters before fixture pagination and exposes deterministic response/mutation counters | Fixture failure is visibly labeled as Evidence; it cannot be mistaken for production persistence | Browser proof demonstrates contracts without implementation | Simulated server boundary remains explicit | Shipping fixture arrays as production cache or claiming browser filtering is canonical | Sufficient for P8/P9 Evidence only; production implementation remains blocked |

## 5. Exact operation homes

B11 introduces no application operation.

### Supporting Organization reads

```text
op6   listUsers
op16  listAreas
op22  listGroups
```

These supply raw identity and selection truth only. They do not transfer User, Area or Group ownership to Authorization or B11.

### Primary B11 operations

```text
op27  listGroupMembers
op28  addGroupMember
op29  removeGroupMember
op30  listRoles
op31  listRoleAssignments
op32  createRoleAssignment
op33  deleteRoleAssignment
```

op31 consumes the operator-ratified B11-F1 refinement in `docs/decisions/access-assignment-read.md`. Operation count remains 89; operation 90+ consumption is zero.

## 6. Bidirectional trace

### Product/backend → frontend

```text
UserPage                     → raw add-member picker + raw grant User picker
AreaPage                     → Por Área selector + grant Area scope
GroupPage                    → Grupos selector + grant Group subject
GroupMemberPage              → selected Group direct-member pages
RoleListView / RoleView      → Funções + fixed meaning + scope guidance
filtered RoleAssignmentPage  → Area / Company / Group / Role canonical slices
CreateRoleAssignmentResult   → terminal initial success + stable replay result
op28 201 / 204               → first relation / existing-relation reconciliation
op29 204                     → exact membership absence reconciliation
op33 204                     → exact assignment revoke
```

Authorization equations, offboarding serialization, same-commit Audit and disclosure remain server/domain authority. The frontend presents accepted commands and current disclosed projections only.

### Frontend → Product/backend

```text
Por Área
  listAreas
  + listRoleAssignments(scope_kind=area, area_id=A)
  + listRoleAssignments(scope_kind=company)

Grupos
  listGroups
  + listRoleAssignments(group_id=G)
  + listGroupMembers(G)

Funções
  listRoles
  + listRoleAssignments(role=R)

Adicionar membro
  raw listUsers pages
  → addGroupMember(G,U) → 201 or 204 reconciliation

Remover membro
  removeGroupMember(G,U)

Conceder acesso
  raw listUsers/listGroups/listAreas + listRoles
  → createRoleAssignment(subject, role, scope)
  → initial 201 / completed replay / same-key ambiguous recovery

Revogar
  deleteRoleAssignment(assignment_id)
```

No material control requires operation 90+, a global matrix/search endpoint, a per-User effective-access endpoint, custom Role/Permission mutation, or an eligibility-filtered User endpoint.

## 7. Client state and pagination law

```text
SERVER STATE
  raw User/Area/Group pages
  GroupMemberPage
  RoleListView
  filtered RoleAssignmentPage slices

NAVIGATION / PRESENTATION
  selected lens and selected Area / Group / Role
  current returned page/window
  production: bounded retained state for already visited returned pages/cursors
  P8 Evidence only: numeric prior-window index over simulated server fixtures

FORM DRAFT
  unaccepted membership User selection
  unaccepted Subject × Role × Scope composition
  one logical create command identity while outcome is unresolved

EPHEMERAL UI / EVIDENCE
  open dialog/drawer, focus return, fixture selector
  deterministic simulated response and mutation counter
```

First-page filters execute server-side before seek pagination. Continuation uses cursor + optional limit only. Filters are cursor-authenticated and are not repeated on continuation.

The P8 Previous controls prove only return to a prior simulated page window; their numeric fixture indexes are not a production offset contract. Production binding may revisit only retained state from pages already returned by the server and never requests an invented reverse cursor, discovers an unvisited page, derives a total, or crawls every page.

The locked JavaScript's complete fixture arrays are test data behind a simulated server boundary. Production frontend state is constrained to returned pages and accepted client state classes above.

## 8. Failure-message intent

```text
400 request.invalid
  Correct request composition; never label it an empty collection.

403 permission.denied
  Deny the action/lens; never render a successful empty access state.

404 notfound.resource
  Reconcile owner truth; do not distinguish absent from non-disclosable.

409 state.conflict
  Explain that current eligibility/referential/security state changed;
  preserve a deliberate draft only where review/correction remains safe.

422 validation.failed / validation.idempotency_key_reused
  Preserve a safe draft and require deliberate correction;
  never mutate command identity silently.

ambiguous createRoleAssignment transport outcome
  Preserve the same logical command; freeze recomposition and closing;
  retry only with the same Idempotency-Key until the outcome resolves.

continuation failure
  Preserve the currently loaded canonical page and retry that continuation only.
```

Membership/offboarding and grant races remain server-serialized. The browser never predicts the winner.

## 9. Access, disclosure and negative contract

```text
visible control != Authorization
access.manage != organization.manage
Organization supporting read != Authorization ownership
RoleView explanation != User effective permission
Group RoleAssignment != Group organizational Area ownership
Company grant != Area grant
loaded page != complete universe
known member != complete membership knowledge
op28 204 != failed mutation
completed replay != second mutation
revoke one assignment != all access removed
remove one membership != all access removed
```

B11 frontend MUST NOT own or synthesize:

```text
Group.area_id
global access matrix or hidden full-collection crawl
client post-filter of incomplete pages presented as complete
eligibility-filtered User directory or complete membership cache
custom Role / Permission editor
per-User effective-access engine / troubleshooter
generic IAM framework or generic admin CRUD authority
provider role/group/claim authority
atomic “Edit grant” built from delete + create
fresh Idempotency-Key for an ambiguous retry
screen-shaped B11 convenience endpoint
```

## 10. Closure

```text
material regions / controls traced             38 / 38
required P9 contract fields bound              14 / 14
B11 primary operations bound                    7 / 7   (ops 27–33)
supporting existing reads                       3       (ops 6,16,22)
raw op6 page/eligibility fidelity               PASS
unknown membership + op28 201/204              PASS
initial op32 terminal success                  PASS
completed replay same result / 1→1             PASS
ambiguous same-key recovery / 1→1              PASS
ambiguous alternate fresh-review path blocked  PASS
continuation page preservation                 PASS
reverse-cursor/offset/total assumption            0
unbound material controls                         0
invented application operations                   0
operation 90+ consumed                             0
frontend Authorization/effective engine           0
unresolved P9 findings                             0
```

**P9 status: COMPLETE / PASS.** Against the exact operator-LOCKED R3 bytes, the unresolved op32 command freezes Subject × Role × Scope, review, confirm and close; only same-key recovery remains available until the terminal replay result. All material controls remain bound in both directions with no invented operation or frontend Authorization authority.
