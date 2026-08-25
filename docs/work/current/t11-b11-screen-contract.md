# T11 — B11 Access Administration — P9 Screen Contract

> **Status:** COMPLETE / POST-LOCK PROOF.  
> **Locked P8:** `docs/work/current/t11-b11-access-administration-p8-r5.html`  
> **Locked blob:** `96094773435a88c357e308779639415d9853b327`  
> **Operator LOCK:** `docs/work/current/t11-b11-operator-lock.md`  
> **Implementation:** BLOCKED.

## 1. Goal

Bind every material B11 region/control in the operator-LOCKED P8 to accepted Product/architecture/read/write authority in both directions.

P9 is proof, not redesign. A contradiction here reopens the smallest affected B11 P7/P8 or upstream owner; it is not permission to weaken the user need or invent a screen-shaped API.

## 2. Material region/control contract

| ID | Surface / region | Goal / information role | Owner + current read truth | Material control / write | Identity source | Material failure / safe UX | Forbidden frontend authority | Status |
|---|---|---|---|---|---|---|---|---|
| B11-01 | `/admin/access` | Stable Access Administration home | T8-F / accepted route | none | route authority | 403 is denial, not empty data | new stable route family | READY |
| B11-02 | Global shell | Preserve accepted MetalDocs navigation/context | B01 LOCK | inherited navigation only | accepted shell | no B11-specific fallback | B11 redesigning shell/IA | READY |
| B11-03 | Notification chrome / Quick Inbox | Preserve accepted global notification utility | B01N LOCK | inherited only | Notification authority | inherited B01N behavior | B11 owning Notification truth | READY |
| B11-04 | `Por Área / Grupos / Funções` | Switch among human access-inspection jobs | B11 operator LOCK | local presentation/navigation state | local lens key | selected lens may reload its canonical reads | backend modules as navigation authority | READY |
| B11-05 | Access boundary note | Explain configuration vs complete effective access | T3 Authorization + B11-F1 | none | current authority | unknown/effective access is never fabricated | client Authorization/effective-access engine | READY |
| B11-06 | Area selector | Select exact disclosed Area or Company presentation lens | Organization `listAreas` op16 | none | `area_id`; Company has no synthetic id | 404/retirement handled by owner read/current truth | synthetic Area business entity | READY |
| B11-07 | `Toda a empresa` lens | Inspect canonical Company scope | B11-F1 op31 `scope_kind=company` | contextual grant may preselect Company | scope kind `company` | empty page is valid; denial is not empty | treating Company as Area | READY |
| B11-08 | Area-specific grants | Inspect assignments scoped exactly to selected Area | Authorization op31 `scope_kind=area&area_id=A` | exact revoke; contextual grant entry | `assignment_id`, `AreaReference`, subject ref | cursor/400/403/404/current-state handling | mixing Company grants into Area truth | READY |
| B11-09 | Company-wide grants shown beside Area | Explain grants that also apply while preserving actual scope | Authorization op31 `scope_kind=company` | exact revoke | `assignment_id`, subject ref | independent page/failure from Area slice | cloning/relabeling Company rows as Area | READY |
| B11-10 | Area assignment pagination | Traverse complete filtered Area slice | B11-F1 + global seek cursor law | next/previous presentation traversal | server cursor | failed continuation preserves loaded canonical rows | offset/total/client all-page crawl | READY |
| B11-11 | Company assignment pagination | Traverse complete Company slice | B11-F1 + global seek cursor law | next/previous | server cursor | failed continuation preserves loaded rows | fabricated total/global matrix | READY |
| B11-12 | Area contextual grant action | Reduce wrong-scope recomposition | accepted P7 R2 + op32 | open grant composer with selected real scope prefilled | exact selected `area_id` or Company scope | user may still cancel/change admissible dimensions | preselection becoming authorization | READY |
| B11-13 | Group selector | Select existing disclosed Group identity | Organization `listGroups` op22 | none | `group_id` | absent/non-disclosable identity reconciles via Organization read | Authorization owning Group identity | READY |
| B11-14 | Group access footprint | Inspect every direct grant to selected Group across scopes | Authorization op31 `group_id=G` | exact revoke; contextual grant | `group_id`, `assignment_id`, scope refs | paginated truth; no inferred missing pages | `Group.area_id`, effective permissions | READY |
| B11-15 | Group footprint pagination | Traverse selected Group's canonical direct assignments | B11-F1 filtered seek pagination | next/previous | cursor bound to `group_id` filter | failed continuation preserves loaded rows | hidden global crawl | READY |
| B11-16 | Group member list | Inspect current direct memberships | Organization state exposed by op27 `listGroupMembers` | remove entry | `group_id`, `user_id` | parent absent/non-disclosable -> 404 | derived effective access as membership state | READY |
| B11-17 | Add-member User picker | Choose an existing User deliberately | supporting Organization `listUsers` op6 | selection only | `user_id`, eligibility projection | disabled/current-member option may be unavailable as UX guidance; server rechecks mutation | User directory ownership transfer | READY |
| B11-18 | Add-member review | Make exact security-bearing target/consequence explicit | current Group footprint + exact User/Group identity | confirm/cancel | exact `user_id + group_id` | consequence remains bounded; no claim of final permission set | inferred effective permission result | READY |
| B11-19 | Add membership | Add exact User to exact Group | `access.manage`; op28 `addGroupMember` | PUT member relation | `group_id + user_id` | 201 first / 204 existing; conflict/offboarding truth remains server-owned | browser deciding eligibility race | READY |
| B11-20 | Remove-member review | Make exact removal and residual-access caveat explicit | current membership + T3 additive grants | confirm/cancel | exact `user_id + group_id` | user may retain direct/other-group access | promise “all access removed” | READY |
| B11-21 | Remove membership | Remove exact relation | `access.manage`; op29 `removeGroupMember` | DELETE relation | `group_id + user_id` | 204 including absent relation when parent exists; parent absent/non-disclosable -> 404 | frontend recomputing effective authority | READY |
| B11-22 | Group contextual grant action | Reduce wrong-subject recomposition | accepted P7 R2 + op32 | open grant composer with exact Group preselected | `group_id` | remaining Role/Scope stay deliberate | Group selection becoming owner of AuthZ | READY |
| B11-23 | Role selector | Inspect fixed Product Role vocabulary | op30 `listRoles` / `RoleListView` fixed T3 order | selection only | `RoleCode` | failure is unavailable role truth, not local fallback | client-maintained Role registry | READY |
| B11-24 | Role detail | Explain fixed Role meaning/scopes | server `RoleView.permissions + allowed_scope_kinds` | none | `RoleCode`, `PermissionCode` | unavailable truth blocks explanation | editable permission matrix/custom role | READY |
| B11-25 | Assignments by Role | Inspect canonical use of selected Role | op31 `role=<RoleCode>` | exact revoke | `assignment_id`, subject/scope refs | filtered pagination/failure | Role-owned membership semantics | READY |
| B11-26 | Role-assignment pagination | Traverse selected Role's canonical slice | B11-F1 filtered cursor | next/previous | cursor bound to role filter | continuation failure preserves loaded rows | total/global matrix inference | READY |
| B11-27 | General grant entry | Start deliberate access grant without lens preselection | op32 + supporting reads | open composer | no command until review | no mutation on open/cancel | generic IAM wizard authority | READY |
| B11-28 | Grant subject selection | Choose User or Group | supporting op6/op22 | local draft selection | exact `user_id` or `group_id` | unavailable identity remains unavailable | mixed “principal” registry becoming authority | READY |
| B11-29 | Grant Role selection | Choose one fixed Role and understand meaning | op30 `RoleView` | local draft | `RoleCode` | no fallback/custom role if read absent | client Role/Permission editing | READY |
| B11-30 | Grant scope selection | Choose Company or exact Area compatible with Role | `RoleView.allowed_scope_kinds` + op16 Areas | local draft | scope kind + optional `area_id` | invalid combinations are prevented as UX and revalidated server-side | scope compatibility as browser authorization | READY |
| B11-31 | Grant final review | Re-state exact command before security mutation | accepted Subject × Role × Scope model | confirm/cancel | exact command fingerprint | preserves draft on correctable failure | implicit edit/replace semantics | READY |
| B11-32 | Create RoleAssignment | Create additive canonical grant | op32 `createRoleAssignment`; `access.manage` | POST `CreateRoleAssignmentRequest` | subject ids + RoleCode + scope ids | 409 current-state conflict; 422 validation/key-reuse; server rechecks truth | client grant authority | READY |
| B11-33 | Ambiguous create retry | Recover unknown transport outcome without duplicate grant | global Idempotency-Key law + op32 | retry same logical command with same key | same command fingerprint + same key | do not create a second logical command silently | new key on ambiguous retry | READY |
| B11-34 | Exact revoke | Revoke only one existing assignment | op33 `deleteRoleAssignment` | DELETE exact assignment | `assignment_id` | 204 first revoke; absent -> 404; residual access may remain | delete+create disguised as atomic edit | READY |
| B11-35 | Failure / recovery evidence | Make materially different safe next actions inspectable | global problem contract + B11 operations | retry/reload/correct where semantically safe | affected resource/command identity | 400 invalid query, 403 denial, 404 absent/non-disclosable, 409 conflict, 422 validation/key-reuse, ambiguous transport | generic toast erasing decision state | READY |
| B11-36 | Responsive / accessibility structure | Preserve meaning and deliberate security actions on narrow/keyboard use | B01/B01N + FRONTEND-METHOD structural requirements | tabs, list/detail, dialogs, focus/labels | same identities as desktop | no hover-only material action; reading order remains Subject/Role/Scope and consequence-before-confirm | responsive visual rewrite of semantics | READY |

## 3. Exact operation homes

B11 introduces no new application operation.

### Supporting Organization reads

```text
op6   listUsers
op16  listAreas
op22  listGroups
```

These supply identity/selection truth only. They do not transfer User/Area/Group ownership to Authorization or B11.

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

`op31` consumes the operator-ratified B11-F1 refinement in `docs/decisions/access-assignment-read.md`. Operation count remains 89; operation 90+ consumption is zero.

## 4. Product/backend → frontend trace

```text
UserPage                     → membership picker / grant User subject
AreaPage                     → Por Área selector / grant Area scope
GroupPage                    → Grupos selector / grant Group subject
GroupMemberPage              → selected Group members
RoleListView / RoleView      → Funções + grant Role meaning/scope compatibility
RoleAssignmentPage + filters → Area / Company / Group / Role canonical slices
CreateRoleAssignmentResult   → created assignment identity / command success
GroupMembership mutations    → selected Group member truth after reconciliation
```

Authorization equations, offboarding serialization, same-commit Audit and disclosure remain server/domain authority; the frontend only presents commands and current disclosed projections.

## 5. Frontend → backend trace

```text
Por Área
  listAreas
  + listRoleAssignments(scope_kind=area, area_id=A)
  + listRoleAssignments(scope_kind=company)

Toda a empresa
  listRoleAssignments(scope_kind=company)

Grupos
  listGroups
  + listRoleAssignments(group_id=G)
  + listGroupMembers(G)

Funções
  listRoles
  + listRoleAssignments(role=R)

Adicionar membro
  listUsers
  → addGroupMember(G,U)

Remover membro
  removeGroupMember(G,U)

Conceder acesso
  supporting listUsers/listGroups/listAreas + listRoles
  → createRoleAssignment(subject,role,scope)

Revogar
  deleteRoleAssignment(assignment_id)
```

No UI control requires operation 90+, a global matrix/search endpoint, a per-User effective-access endpoint, or custom Role/Permission mutation.

## 6. Client state classes

```text
SERVER STATE
  User/Area/Group pages
  GroupMemberPage
  RoleListView
  filtered RoleAssignmentPage slices

NAVIGATION / PRESENTATION
  selected local lens
  selected Area / Group / Role
  page cursors/window state

FORM DRAFT
  unaccepted membership User selection
  unaccepted Subject × Role × Scope grant composition

EPHEMERAL UI
  open dialog
  focus return
  one-shot Evidence fixture
  ambiguous-command recovery presentation
```

No global normalized Authorization store, expanded permission cache, or browser-maintained effective access graph is admitted.

## 7. Pagination and identity proof

`listRoleAssignments` filters execute server-side before seek pagination. Continuation uses cursor + optional limit only; filters are cursor-authenticated and are not repeated with continuation.

```text
Area lens    = one exact filtered collection
Company lens = one exact filtered collection
Group lens   = one exact filtered collection
Role lens    = one exact filtered collection
```

A loaded page is never labeled as the entire global collection when `has_more=true`. B11 does not crawl every page to build a hidden matrix.

## 8. Failure-message intent

```text
400 request.invalid
  frontend must not intentionally generate invalid op31 filter combinations;
  if received, recover from the request composition rather than calling it empty.

403 permission.denied
  deny the action/lens; never render a successful empty access state.

404 notfound.resource
  selected Group/Area/assignment is absent or non-disclosable;
  reconcile current owner truth and do not infer hidden existence.

409 state.conflict
  current referential/eligibility/security state changed;
  preserve deliberate draft where correction/review remains meaningful.

422 validation.failed / validation.idempotency_key_reused
  preserve safe draft and require a deliberate correction;
  never mutate the rejected command silently.

ambiguous createRoleAssignment transport outcome
  keep the same logical command and retry with the same Idempotency-Key.

CSRF recovery
  re-bootstrap only as required and retry only the same logical unsafe command when safe.
```

Membership/offboarding races remain server-serialized. The frontend never predicts the winner.

## 9. Access / disclosure proof

```text
visible control != Authorization
access.manage != organization.manage
Organization supporting read != Authorization ownership
RoleView permission explanation != User effective permission
Group RoleAssignment != Group organizational Area ownership
Company grant != Area grant
loaded page != complete universe
revoke one assignment != all access removed
remove one membership != all access removed
```

The LOCKED copy preserves those distinctions before consequential actions.

## 10. Negative contract

B11 frontend MUST NOT own or synthesize:

```text
Group.area_id
global access matrix
global User/Group/Area search absent accepted authority
custom Role / Permission editor
per-User effective-access engine / troubleshooter
client post-filter of incomplete pages presented as complete
hidden full-collection crawl
generic IAM framework
provider role/group/claim authority
atomic “Edit grant” built from delete + create
screen-shaped B11 convenience endpoint
```

## 11. Closure

```text
material regions / controls traced        36 / 36
B11 primary operations bound               7 / 7   (ops 27–33)
supporting existing reads                  3       (ops 6,16,22)
unbound material controls                  0
invented application operations            0
operation 90+ consumed                      0
screen-shaped APIs                         0
frontend Authorization/effective engine    0
fake global search/post-filter              0
unresolved P9 findings                     0
```

**P9 verdict:** PASS. The operator-LOCKED B11 structure is realizable from current accepted authority without a frontend business/Authorization authority or new backend capability.
