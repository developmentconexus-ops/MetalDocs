from pathlib import Path

proposal = Path("docs/work/current/proposal.md")
s = proposal.read_text()

def repl(old, new, count=1):
    global s
    n = s.count(old)
    if n != count:
        raise SystemExit(f"proposal anchor mismatch expected={count} got={n}: {old[:120]!r}")
    s = s.replace(old, new, count)

repl(
"""This applies only to whole-replacement current truths. Stale DRAFT PATCH **always** -> `412 precondition.draft_changed`, even if requested values happen to equal current values.

An ETag-protected representation contains only fields governed by that concurrency token plus immutable identifiers; independently mutable display enrichment is excluded.""",
"""This applies only to whole-replacement current truths. Stale DRAFT PATCH **always** -> `412 precondition.draft_changed`, even if requested values happen to equal current values.

Exact-current comparison is over the normalized **semantic value**, not raw JSON bytes: JSON object member order is irrelevant; set-valued replacements such as eligible-template ids compare as sets after uniqueness/canonicalization; governance-route Step array order remains semantic and is never sorted away.

An ETag-protected representation contains only fields governed by that concurrency token plus immutable identifiers; independently mutable display enrichment is excluded."""
)

repl(
"""An explicitly requested absent/non-disclosable `area_id` or `document_type_id` ->404. Every pair of individually usable Area + DocumentType is semantically admissible at this read surface; inapplicable templates/candidates yield empty/absent sublists rather than an invented 422 mode. If real scale makes these complete arrays unsustainable, that evidence reopens T6; T8-E does not truncate or add operation79.

## 2.8 Header profiles""",
"""An explicitly requested absent/non-disclosable `area_id` or `document_type_id` ->404. Every pair of individually usable Area + DocumentType is semantically admissible at this read surface; inapplicable templates/candidates yield empty/absent sublists rather than an invented 422 mode. If real scale makes these complete arrays unsustainable, that evidence reopens T6; T8-E does not truncate or add operation79.

Deterministic unpaginated response-array order is closed rather than implementation-defined:

```text
ProviderSubjectSearchView.items                 provider relevance/enumeration order
ProviderSubjectOption.display_hints             provider order, max 3
RoleListView.items                              fixed T3 role order
RoleView.permissions                            bundle order written in §3.3
RoleView.allowed_scope_kinds                    company then area when both are legal
EligibleTemplatesView.templates                 document.code ASC, document_id ASC
TemplateConfigurationItem.eligible_document_type_ids  UUID ASC
DocumentCreationOptionsView.areas               area.code ASC, area_id ASC
DocumentCreationOptionsView.document_types      document_type.code ASC, document_type_id ASC
DocumentCreationOptionsView.templates           document.code ASC, document_id ASC
DocumentCreationOptionsView.responsible_owner_candidates user_id ASC
GovernancePolicy.steps                          configured route order (semantic)
GovernanceCaseView.allowed_actions              accept, return_for_changes, add_feedback filtered in that order
```

Set-valued request arrays (`template_document_ids`) are order-insensitive and reject duplicates; route Step arrays are order-sensitive. Map member order is never semantic.

## 2.8 Header profiles"""
)

repl(
"""Capability/admission law:

```text
create-only (`If-None-Match:*` or provider-equivalent)
+ exact HTTP body length == expected_size_bytes
+ one shared expires_at for provider capability and unconsumed admission claim
```

For browser PUT, `Content-Length` is user-agent generated from the exact Blob/body and is part of the signed/provider-enforced request constraint; it is not placed in `required_headers` as a script-settable header. `required_headers` contains only exact browser-settable provider headers.

Completion independently `Stat/OpenExact`s the object, requires actual size == expected size, derives SHA-256/actual format/actual size, performs structural validation, and only then establishes READY. Client-declared type/hash/name is non-authoritative.""",
"""Capability/admission law:

```text
create-only (`If-None-Match:*` or provider-equivalent)
+ provider-enforced body length <= expected_size_bytes
+ one shared expires_at for provider capability and unconsumed admission claim
```

`expected_size_bytes` is a **capability/resource bound only**, never semantic content identity. The AWS S3 reference profile is deliberately stricter: its presigned PUT binds the browser-generated exact `Content-Length` to that value. `Content-Length` is not placed in `required_headers` because browser script cannot set it; `required_headers` contains only exact browser-settable provider headers.

Completion independently `Stat/OpenExact`s the object, derives SHA-256/actual format/actual size, rechecks the global raw-byte ceiling and structural admission rules, and only then establishes READY. It does **not** turn the client-declared expected size into a second descriptor/equality authority. Client-declared type/hash/name/expected length remain non-authoritative; the actual server-derived descriptor is semantic truth."""
)

repl(
"""GovernancePolicy
  {mode:no_human_approval}
  {mode:use_governance_route,steps:[GovernanceRouteStep,...]}
GovernanceRouteStep""",
"""GovernancePolicy
  {mode:no_human_approval}
  {mode:use_governance_route,steps:[GovernanceRouteStep,...]} // minItems=1
GovernanceRouteStep"""
)

repl(
"""RepresentationPolicy
  {kind:source_only}
  {kind:require_official_rendition,format:pdf}
```

Pagination:""",
"""RepresentationPolicy
  {kind:source_only}
  {kind:require_official_rendition,format:pdf}
```

`no_human_approval` forbids a `steps` member; `use_governance_route` requires at least one Step. Step array order is the configured sequential route order.

Pagination:"""
)

repl(
"`PermissionCode` is exactly the accepted **15-value** T3 dot-spelled vocabulary.",
"""`PermissionCode` is exactly this accepted 15-value T3 vocabulary:

```text
organization.manage
access.manage
document_type.manage
template_use.manage
document.read_effective
document.read_history
document.read_working
document.create
document.edit
document.submit
document.cancel_revision
document.obsolete
document.owner.manage
governance.act
audit.read
```"""
)

repl(
"""TemplateConfigurationPage { items:TemplateConfigurationItem[], page:Page }
```

Create explicitly supplies initial governance/representation""",
"""TemplateConfigurationPage { items:TemplateConfigurationItem[], page:Page }
```

`TemplateConfigurationItem.current_effective_title` is present iff `has_effective_revision=true`; `eligible_document_type_ids` is unique and UUID-ascending.

Create explicitly supplies initial governance/representation"""
)

repl(
"""`official` is present iff at least one Release exists; obsolete retains last released official. Newer cancelled/open work never replaces older EFFECTIVE official truth. Before first Release, status may be draft/submitted/cancelled and `official` is absent.""",
"""Closed presence laws prevent the official/catalog lenses from drifting into a second lifecycle authority:

```text
DocumentSummary.status=effective|obsolete  -> official_revision required
DocumentSummary.status=cancelled            -> official_revision absent

DocumentOfficialView.status=effective|obsolete -> official required
DocumentOfficialView.status=draft|submitted|cancelled -> official absent
```

`official` is therefore present iff at least one Release exists; obsolete retains the last released official. Newer cancelled/open work never replaces older EFFECTIVE official truth. Before first Release, status may be draft/submitted/cancelled and `official` is absent."""
)

repl(
"""Raw WorkingContent generation is never public; ETag is wire OCC authority.""",
"""`RevisionView.current_submission_id` is present **iff** `state=submitted`; every other RevisionState forbids it. `DocumentCreationOptionsView.default_responsible_owner` is the current actor; candidate-list presence remains exactly the §2.7 owner-manage rule. Raw WorkingContent generation is never public; ETag is wire OCC authority."""
)

old_history = """`DocumentHistoryItem` closed kinds:

```text
revision_created             revision,title,occurred_at
submission_created           submission_id,revision,title,submitter,occurred_at,?governance_attempt_id
governance_decision          decision_id,governance_attempt_id,step_id,actor,outcome,occurred_at,?reason
feedback_added               feedback_id,governance_attempt_id,actor,message,occurred_at
submission_withdrawn         submission_id,actor,occurred_at
revision_cancelled           revision_id,actor,reason,occurred_at
release_completed            release_id,revision_id,submission_id,occurred_at,?predecessor_revision_id
official_rendition_completed official_rendition_id,submission_id,occurred_at
obsolescence_requested       request_id,target_revision_id,initiator,reason,occurred_at,?governance_attempt_id
obsolescence_withdrawn       request_id,actor,occurred_at
obsolescence_completed       request_id,target_revision_id,occurred_at
```

```text
DocumentHistoryPage {items:DocumentHistoryItem[],page:Page}
WorkAuthoringItem {document,revision,title,state:OpenRevisionState,responsible_owner,updated_at}
WorkAuthoringPage {items:WorkAuthoringItem[],page:Page}
WorkGovernanceItem {governance_attempt_id,subject_kind,document,created_at}
WorkGovernancePage {items:WorkGovernanceItem[],page:Page}
```"""
new_history = """`DocumentHistoryItem` is a closed `kind`-discriminated union:

```text
revision_created
  {kind,revision:RevisionIdentity,title:nonblank string,occurred_at:UtcInstant}
submission_created
  {kind,submission_id:Uuid,revision:RevisionIdentity,title:nonblank string,submitter:UserReference,occurred_at:UtcInstant,governance_attempt_id?:Uuid}
governance_decision
  {kind,decision_id:Uuid,governance_attempt_id:Uuid,step_id:Uuid,actor:UserReference,outcome:GovernanceDecisionOutcome,occurred_at:UtcInstant,reason?:nonblank string}
  reason present iff outcome=return_for_changes
feedback_added
  {kind,feedback_id:Uuid,governance_attempt_id:Uuid,actor:UserReference,message:nonblank string,occurred_at:UtcInstant}
submission_withdrawn
  {kind,submission_id:Uuid,actor:UserReference,occurred_at:UtcInstant}
revision_cancelled
  {kind,revision_id:Uuid,actor:UserReference,reason:nonblank string,occurred_at:UtcInstant}
release_completed
  {kind,release_id:Uuid,revision_id:Uuid,submission_id:Uuid,occurred_at:UtcInstant,predecessor_revision_id?:Uuid}
official_rendition_completed
  {kind,official_rendition_id:Uuid,submission_id:Uuid,occurred_at:UtcInstant}
obsolescence_requested
  {kind,request_id:Uuid,target_revision_id:Uuid,initiator:UserReference,reason:nonblank string,occurred_at:UtcInstant,governance_attempt_id?:Uuid}
obsolescence_withdrawn
  {kind,request_id:Uuid,actor:UserReference,occurred_at:UtcInstant}
obsolescence_completed
  {kind,request_id:Uuid,target_revision_id:Uuid,occurred_at:UtcInstant}
```

The `revision_created.title` is Revision display metadata from the persisted Revision, not a claim that a separate title-at-created-time snapshot exists; exact historical submission titles come from immutable Submission snapshots.

```text
DocumentHistoryPage {items:DocumentHistoryItem[],page:Page}
WorkAuthoringItem {document:DocumentReference,revision:RevisionIdentity,title:nonblank string,state:OpenRevisionState,responsible_owner:UserReference,updated_at:UtcInstant}
WorkAuthoringPage {items:WorkAuthoringItem[],page:Page}
WorkGovernanceItem {governance_attempt_id:Uuid,subject_kind:GovernanceSubjectKind,document:DocumentReference,created_at:UtcInstant}
WorkGovernancePage {items:WorkGovernanceItem[],page:Page}
```"""
repl(old_history, new_history)

old_audit = """Typed wire facts only when operation/resource identity is insufficient:

```text
GroupMembershipAuditFacts { user_id:Uuid }
RoleAssignmentAuditFacts { subject:RoleAssignmentSubject, role:RoleCode, scope:RoleAssignmentScope }
GovernanceDecisionAuditFacts { governance_attempt_id:Uuid, step_id:Uuid, subject_kind:GovernanceSubjectKind, subject_id:Uuid, outcome:GovernanceDecisionOutcome }
ReleaseAuditFacts { document_id:Uuid, revision_id:Uuid, submission_id:Uuid, predecessor_revision_id?:Uuid }
RevisionCancellationAuditFacts { document_id:Uuid, revision_id:Uuid }
ObsolescenceAuditFacts { document_id:Uuid, target_revision_id:Uuid }
```

`resource_id` supplies group/assignment/decision/release/obsolescence evidence identity; duplicate IDs are not repeated inside facts.

`AuditEventView` is a closed `operation_code`-discriminated union with common `event_id,occurred_at,actor,operation_code,resource_kind,resource_id,visibility`. Simple branches have no `facts`; typed branches require the matching facts. No free-form feedback/reason/profile/provider payload.
"""
new_audit = """Typed wire facts exist only when operation/resource identity is insufficient:

```text
GroupMembershipAuditFacts { user_id:Uuid }
RoleAssignmentAuditFacts { subject:RoleAssignmentSubject, role:RoleCode, scope:RoleAssignmentScope }
DocumentTypeConfigurationAuditFacts { document_type_code:CodeToken } // resulting canonical code at event commit
GovernanceDecisionAuditFacts { governance_attempt_id:Uuid, step_id:Uuid, subject_kind:GovernanceSubjectKind, subject_id:Uuid, outcome:GovernanceDecisionOutcome }
ReleaseAuditFacts { document_id:Uuid, revision_id:Uuid, submission_id:Uuid, predecessor_revision_id?:Uuid }
RevisionCancellationAuditFacts { document_id:Uuid }
ObsolescenceAuditFacts { document_id:Uuid, target_revision_id:Uuid }
```

`resource_id` supplies the stable event resource/evidence identity; duplicate IDs are not repeated inside facts. Configuration facts carry one bounded code rather than arbitrary before/after JSON.

The closed `operation_code -> resource_kind -> actor -> visibility -> facts` matrix is:

| operation code(s) | resource_kind / resource_id | actor | visibility snapshot | facts |
|---|---|---|---|---|
| `provider_binding.accepted`, `provider_binding.replaced` | `provider_binding` / binding id | USER | COMPANY | none |
| `user.created`, `user.offboarded`, `user.reenabled` | `user` / user id | USER | COMPANY | none |
| `user_profile.erased` | `user_profile` / user id | USER | COMPANY | none |
| `area.created`, `area.renamed`, `area.retired`, `area.reenabled` | `area` / area id | USER | AREA = same area id | none |
| `group.created`, `group.renamed`, `group.deleted` | `group` / group id | USER | COMPANY | none |
| `group_membership.added`, `group_membership.removed` | `group` / group id | USER | COMPANY | `GroupMembershipAuditFacts` |
| `role_assignment.granted`, `role_assignment.revoked` | `role_assignment` / assignment id | USER | Company scope => COMPANY; Area scope => that AREA | `RoleAssignmentAuditFacts` |
| `document_type.created`, `document_type.reconfigured`, `document_type.activated`, `document_type.inactivated`, `document_governance.changed`, `template_eligibility.changed` | `document_type` / document_type id | USER | COMPANY | `DocumentTypeConfigurationAuditFacts` |
| `document.responsible_owner_changed`, `document.template_role_changed`, `document.created` | `document` / document id | USER | document Area snapshot | none |
| `revision.created` | `revision` / revision id | USER | document Area snapshot | none |
| `submission.created`, `submission.withdrawn` | `submission` / submission id | USER | document Area snapshot | none |
| `governance.accepted`, `governance.returned_for_changes` | `governance_decision` / decision id | USER | governed Document Area snapshot | `GovernanceDecisionAuditFacts` |
| `revision.cancelled` | `revision` / revision id (also cancellation evidence identity) | USER | document Area snapshot | `RevisionCancellationAuditFacts` |
| `official_rendition.completed` | `official_rendition` / rendition id | SYSTEM=`metaldocs` | document Area snapshot | none |
| `release.completed` | `release` / release id | SYSTEM=`metaldocs` | document Area snapshot | `ReleaseAuditFacts` |
| `obsolescence.requested`, `obsolescence.withdrawn` | `obsolescence_request` / request id | USER | document Area snapshot | `ObsolescenceAuditFacts` |
| `obsolescence.completed` | `obsolescence_request` / request id | SYSTEM=`metaldocs` | document Area snapshot | `ObsolescenceAuditFacts` |

SYSTEM attribution on rendition/release/obsolescence completion records the system-owned effect after its gates, never a fake human approver. RoleAssignment visibility must match the scope repeated in `RoleAssignmentAuditFacts`. Area/document attribution is frozen at event commit and never recomputed from later current state.

`AuditEventView` is a closed `operation_code`-discriminated union with common `{event_id:Uuid,occurred_at:UtcInstant,actor:AuditActor,operation_code:AuditOperationCode,resource_kind,resource_id:Uuid,visibility:AuditVisibility}`. Simple branches forbid `facts`; typed branches require exactly the matching facts schema. No free-form feedback/reason/profile/provider payload.
"""
repl(old_audit, new_audit)

repl(
"""Required: `type,title,status,detail,instance,code,trace_id`. `instance=urn:uuid:<fresh UUID>` per occurrence. `trace_id` opaque/nonblank. Optional non-empty `errors[]` is allowed only on `request.invalid` and validation-family variants; each error uses an RFC6901 pointer rooted at `/path`, `/query`, `/header`, or `/body` and never echoes sensitive rejected values.""",
"""Required: `type,title,status,detail,instance,code,trace_id`. `instance=urn:uuid:<fresh UUID>` per occurrence. `trace_id` opaque/nonblank.

Optional non-empty `errors[]` is allowed only on `request.invalid` and validation-family variants. Its sole item shape is closed:

```text
ProblemError { pointer:Rfc6901Pointer, detail:nonblank string }
```

`pointer` is a valid RFC6901 pointer rooted at `/path`, `/query`, `/header`, or `/body`. `ProblemError` has no rejected-value/meta/code bag and never echoes sensitive rejected values; machine branching remains on the top-level `Problem.code`."""
)

repl(
"""actual provider bytes != expected_size_bytes
  -> completion rejected; READY not established

actual DOCX expanded bytes > DOCX_EXPANDED_MAX_BYTES""",
"""actual raw bytes > DOC_RAW_MAX_BYTES
  -> 422 validation.content_invalid; READY not established (defense-in-depth against a broken provider bound)

actual DOCX expanded bytes > DOCX_EXPANDED_MAX_BYTES"""
)

repl(
"""`StartDraftUploadRequest.expected_size_bytes <= DOC_RAW_MAX_BYTES`. The allocation does not echo a redundant global `max_bytes`; the client already supplied the exact intended length and the schema owns the Launch ceiling.""",
"""`StartDraftUploadRequest.expected_size_bytes <= DOC_RAW_MAX_BYTES`. The allocation does not echo a redundant global `max_bytes`; the client already supplied the requested capability bound and the schema owns the Launch ceiling. Completion derives the actual length independently; it never persists/promotes the client bound as semantic content truth."""
)

repl(
"""current seam: PresignCreate(handle,maxBytes,ttl)
consumer call: maxBytes = expected_size_bytes
provider PUT profile: exact Content-Length = maxBytes
```

A capability that allows exactly `maxBytes` bytes is a valid stricter realization of an at-most-`maxBytes` contract. No new parameter or durable authority is needed.""",
"""current seam: PresignCreate(handle,maxBytes,ttl)
consumer call: maxBytes = expected_size_bytes
portable property: provider forbids body > maxBytes
S3 reference: stricter signed exact Content-Length = maxBytes
```

No new parameter, AdmissionClaim field, or durable authority is needed. The client bound protects ingress resource use; completion still derives the real descriptor and does not compare against a persisted client expectation."""
)

open_finding = """
## 8.4 Final Lead cross-layer finding — T8-D transaction-census precision — OPEN

The final T3↔T8-D parity attack found one bounded owner-local correction package. T3 remains the Audit owner; T8-C remains the zero-or-one durable-intent owner. The persistence structures already support every required transition, so this finding adds no table, owner, state, API, permission, worker or capability.

Current T8-D transaction-census deltas:

```text
1. Company replacement currently says -> Audit
   T3 does not require semantic Company-display replacement Audit
   -> subtract that mandatory Audit

2. User DISABLED -> ENABLED is a ratified operation and T3 requires user.reenabled
   T8-D has enabled + eligibility_version but transaction census names only offboarding
   -> add explicit re-enable CAS/transition + required Audit; no grants/memberships/sessions resurrect

3. UserProfile replacement/erasure says "Audit when required"
   -> close it: ordinary replacement has no mandatory semantic Audit;
      lawful erasure emits user_profile.erased when T3 requires evidence

4. DRAFT PATCH says "Audit when required"
   T3 explicitly does not require autosave/WorkingContent semantic Audit
   -> no mandatory semantic Audit for Launch DRAFT PATCH

5. Feedback says "Audit/Replay as upstream requires"
   T3 explicitly says SubmissionFeedback is its own immutable evidence and needs no duplicate Audit
   -> append feedback + Replay only; no duplicate semantic Audit

6. SUBMIT currently says "required River intent" unconditionally
   T8-C owns zero-or-one named durable intent and T8-E proves SourceOnly / already-PDF paths need none
   -> insert a River intent iff the transition actually activates one

7. OfficialRendition finalization omits the T3-required rendition-completion Audit
   -> append official_rendition.completed; if that same transition establishes Release,
      also append release.completed in the same local commit

8. Multi-effect transactions currently use a generic singular "Audit" shorthand
   -> state once that each transaction appends ALL AND ONLY T3-required semantic events
      for the facts/effects it commits (e.g. User+Binding, teardown+offboarding,
      Submission+Release, Decision+Release, requested+completed obsolescence)
```

This is a **bounded precision/reduction**, not a semantic reopen: it removes two speculative/duplicate Audit paths and an unconditional job, while making three already-required evidence paths explicit. T8-E does not edit T8-D until explicit operator approval.
"""
repl("\n---\n\n# 9. Generation / provider feasibility and runtime conformance proof", open_finding + "\n---\n\n# 9. Generation / provider feasibility and runtime conformance proof")

repl(
"""Therefore the concrete reference provider can bind both exact body length and create-only precondition in the signed PUT request without POST-form/multipart.""",
"""Therefore the concrete S3 reference provider can strengthen the portable `<= expected_size_bytes` capability bound to an exact browser body length while also enforcing create-only semantics, without POST-form/multipart. This stronger provider mechanism does not make the client-declared length semantic content identity."""
)

repl(
"""unknown JSON/query member and duplicate scalar/member rejection
bodyless operation rejects a body""",
"""unknown JSON/query member and duplicate scalar/member rejection
bodyless operation rejects a body
query/bodyless rules derived from the matched OpenAPI operation, never a handwritten route allowlist"""
)

repl(
"""upload exact Content-Length/create-once/shared15min expiry + completion size re-proof""",
"""upload provider <= expected-size bound; S3 exact signed Content-Length; create-once/shared15min expiry; completion independently derives actual descriptor/global ceiling"""
)

repl(
"""Therefore MetalDocs does **not** add a second schema/validation framework. The envelope guard owns only raw/request-shape properties OAS/kin-openapi does not enforce; OpenAPI remains the schema authority.""",
"""Therefore MetalDocs does **not** add a second schema/validation framework. The envelope guard owns only raw/request-shape properties OAS/kin-openapi does not enforce; OpenAPI remains the schema authority. Unknown-query/bodyless checks derive their permitted shape from the matched OpenAPI operation, and duplicate-scalar handling derives from the current scalar query parameter definitions — no second handwritten route/parameter allowlist exists."""
)

repl(
"""duplicate PDF rendition bytes for PDF source
dormant future capability""",
"""duplicate PDF rendition bytes for PDF source
persisted/client-authored expected-size descriptor truth
second request-schema/route-parameter validation authority
dormant future capability"""
)

old_gate = """The measurement, generated-boundary feasibility, provider presign feasibility and ledger-census fixture obligations are closed at candidate level.

Remaining Lead gate:

```text
A. run one final whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence attack
B. close every surviving Lead finding without speculative capability
C. revalidate exact candidate HEAD + intended 5-file durable/work diff + required CI
D. only if A→C converge, create review/t8e-fable from that exact candidate HEAD
E. independent Fable challenge
F. Lead adjudication of Fable evidence
G. explicit operator ratification
```

Until A→C converge:"""
new_gate = """The measurement, generated-boundary feasibility, provider presign feasibility, strict-request validator split and 78-row ledger-census fixture obligations are closed at candidate level. The final Lead attack is complete except for the bounded T8-D transaction-census package in §8.4, which crosses an already-ratified authority and therefore requires operator adjudication.

Remaining Lead gate:

```text
A. operator adjudication of §8.4 bounded T8-D transaction-census precision
B. if approved, apply only that owner-local correction package
C. rerun whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence exact-delta check
D. revalidate main/base + exact candidate HEAD + intended 5-file diff + required CI
E. only if A→D converge, create review/t8e-fable from that exact candidate HEAD
F. independent Fable challenge
G. Lead adjudication of Fable evidence
H. explicit operator ratification
```

Until A→D converge:"""
repl(old_gate, new_gate)

proposal.write_text(s)

roadmap = Path("docs/roadmap.md")
r = roadmap.read_text()

def rrepl(old, new, count=1):
    global r
    n = r.count(old)
    if n != count:
        raise SystemExit(f"roadmap anchor mismatch expected={count} got={n}: {old[:120]!r}")
    r = r.replace(old, new, count)

rrepl(
"""direct S3 PUT           signed exact Content-Length + If-None-Match:* probe PASS
```""",
"""direct S3 PUT           signed exact Content-Length + If-None-Match:* probe PASS
strict request split        kin-openapi + minimal envelope-guard probe PASS
```"""
)

rrepl(
"""The direct-PUT exact-length concern remains resolved subtractively without reopening T8-C: the existing `PresignCreate(handle,maxBytes,ttl)` seam is sufficient when T8-E supplies `maxBytes=expected_size_bytes` and the provider PUT profile binds that value as exact `Content-Length`.

## Exact next action

```text
prove exact ledger/census/profile fixtures across all 78 operation rows
→ run final whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence attack
→ if and only if no material contradiction survives, create isolated review/t8e-fable from exact candidate HEAD
→ independent Fable challenge
→ Lead adjudication
→ explicit operator ratification
```""",
"""The direct-PUT concern remains resolved subtractively without reopening T8-C: the existing `PresignCreate(handle,maxBytes,ttl)` seam is sufficient when T8-E supplies `maxBytes=expected_size_bytes`; the portable property is an at-most bound, while the reference S3 profile is stronger and signs exact `Content-Length`. Completion derives the actual descriptor independently, so no client-size truth is persisted.

The final Lead coherence attack found one remaining bounded upstream precision package in T8-D's **transaction census only**: remove mandatory Audit where T3 does not require it, make required T3 evidence/multi-event paths explicit, make SUBMIT River intent conditional on a real activated effect, and name the already-ratified User re-enable transition. No table/state/API/permission/worker/capability is added.

## Exact next action

```text
operator adjudicates the bounded T8-D transaction-census precision package recorded in docs/work/current/proposal.md §8.4
→ if approved, reconcile only docs/architecture/persistence.md
→ rerun exact whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence delta
→ revalidate main/base + exact candidate HEAD + intended 5-file diff + required CI
→ only then create isolated review/t8e-fable from exact candidate HEAD
→ independent Fable challenge
→ Lead adjudication
→ explicit operator ratification
```"""
)

roadmap.write_text(r)
