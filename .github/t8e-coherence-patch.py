from pathlib import Path

proposal = Path("docs/work/current/proposal.md")
s = proposal.read_text()

def repl(old, new, count=1):
    global s
    n = s.count(old)
    if n != count:
        raise SystemExit(f"proposal anchor mismatch expected={count} got={n}: {old[:140]!r}")
    s = s.replace(old, new, count)

# Email is erasable contact enrichment only; do not invent deliverability/canonicalization policy.
repl(
"""EmailAddress
  trim surrounding whitespace
  minLength=3; maxLength=254; OpenAPI format=email
  no case-folding/canonicalization, uniqueness or verification claim
  profile/contact metadata only; never authentication or Authorization identity
```

Except for scalars that explicitly define normalization above (`CodeInput`, `SearchQuery`, `EmailAddress`), accepted human text is **not** silently trimmed, case-folded or Unicode-normalized by convention.""",
"""EmailAddress
  nonblank human text
  no trim/case-fold/canonicalization, deliverability, uniqueness or verification claim
  no OpenAPI `format: email` Launch gate because no current email-delivery/identity consumer exists
  profile/contact enrichment only; never authentication or Authorization identity
```

Except for scalars that explicitly define normalization above (`CodeInput`, `SearchQuery`), accepted human text is **not** silently trimmed, case-folded or Unicode-normalized by convention."""
)

# A revision-created timeline row must not pretend current/final title was snapshotted at creation.
repl(
"""revision_created
  {kind,revision:RevisionIdentity,title:nonblank string,occurred_at:UtcInstant}""",
"""revision_created
  {kind,revision:RevisionIdentity,occurred_at:UtcInstant}"""
)
repl(
"""
The `revision_created.title` is Revision display metadata from the persisted Revision, not a claim that a separate title-at-created-time snapshot exists; exact historical submission titles come from immutable Submission snapshots.
""",
"""
Exact historical submitted titles come from immutable Submission snapshots; T8-E does not fabricate a title-at-revision-creation snapshot that T1→T8-D never persisted.
"""
)

# T8-C says actor/visibility meaning is owner-authored evidence; transport must not infer it from operation_code.
repl(
"""AuditActor {kind:user,user_id:Uuid} | {kind:system,system_actor_code:metaldocs}
AuditVisibility {kind:company} | {kind:area,area_id:Uuid}""",
"""AuditActor {kind:user,user_id:Uuid} | {kind:system}
AuditVisibility {kind:company} | {kind:area,area_id:Uuid}"""
)

start = "Typed wire facts exist only when operation/resource identity is insufficient:"
end = "`AuditEventPage={items:AuditEventView[],page:Page}` ordered `occurred_at DESC,event_id DESC`."
if s.count(start) != 1 or s.count(end) != 1:
    raise SystemExit("audit block anchors are not unique")
i = s.index(start)
j = s.index(end, i)
new_audit = """Typed wire facts exist only when operation/resource identity is insufficient:

```text
GroupMembershipAuditFacts { user_id:Uuid }
RoleAssignmentAuditFacts { subject:RoleAssignmentSubject, role:RoleCode, scope:RoleAssignmentScope }
GovernanceDecisionAuditFacts { governance_attempt_id:Uuid, step_id:Uuid, subject_kind:GovernanceSubjectKind, subject_id:Uuid, outcome:GovernanceDecisionOutcome }
ReleaseAuditFacts { document_id:Uuid, revision_id:Uuid, submission_id:Uuid, predecessor_revision_id?:Uuid }
RevisionCancellationAuditFacts { document_id:Uuid }
ObsolescenceAuditFacts { document_id:Uuid, target_revision_id:Uuid }
```

`resource_id` supplies the stable event/resource evidence identity; duplicate ids are not repeated inside facts. T3-required bounded DocumentType/configuration facts remain internal owner-authored Audit evidence at Launch: the wire exposes the closed operation code + resource identity rather than inventing a generic configuration-diff bag without a named UI consumer.

The closed wire projection constrains only `operation_code -> resource_kind -> exposed facts`:

| operation code(s) | resource_kind / resource_id | exposed wire facts |
|---|---|---|
| `provider_binding.accepted`, `provider_binding.replaced` | `provider_binding` / binding id | none |
| `user.created`, `user.offboarded`, `user.reenabled` | `user` / user id | none |
| `user_profile.erased` | `user_profile` / user id | none |
| `area.created`, `area.renamed`, `area.retired`, `area.reenabled` | `area` / area id | none |
| `group.created`, `group.renamed`, `group.deleted` | `group` / group id | none |
| `group_membership.added`, `group_membership.removed` | `group` / group id | `GroupMembershipAuditFacts` |
| `role_assignment.granted`, `role_assignment.revoked` | `role_assignment` / assignment id | `RoleAssignmentAuditFacts` |
| `document_type.created`, `document_type.reconfigured`, `document_type.activated`, `document_type.inactivated`, `document_governance.changed`, `template_eligibility.changed` | `document_type` / document_type id | none |
| `document.responsible_owner_changed`, `document.template_role_changed`, `document.created` | `document` / document id | none |
| `revision.created` | `revision` / revision id | none |
| `submission.created`, `submission.withdrawn` | `submission` / submission id | none |
| `governance.accepted`, `governance.returned_for_changes` | `governance_decision` / decision id | `GovernanceDecisionAuditFacts` |
| `revision.cancelled` | `revision` / revision id (also cancellation evidence identity) | `RevisionCancellationAuditFacts` |
| `official_rendition.completed` | `official_rendition` / rendition id | none |
| `release.completed` | `release` / release id | `ReleaseAuditFacts` |
| `obsolescence.requested`, `obsolescence.withdrawn`, `obsolescence.completed` | `obsolescence_request` / request id | `ObsolescenceAuditFacts` |

`actor` and historical `visibility` are serialized **verbatim from the owner-authored evidence required by T8-C §4**. T8-E never derives USER/SYSTEM or COMPANY/AREA from an operation code, current RoleAssignment, current Area, queue identity, or transport context. The internal product-owned `system_actor_code` is likewise not exposed until a named client needs to distinguish multiple system principals.

`AuditEventView` is a closed `operation_code`-discriminated union with common `{event_id:Uuid,occurred_at:UtcInstant,actor:AuditActor,operation_code:AuditOperationCode,resource_kind,resource_id:Uuid,visibility:AuditVisibility}`. Simple wire branches forbid `facts`; typed branches require exactly the matching exposed facts schema. No free-form feedback/reason/profile/provider/config-diff payload.

"""
s = s[:i] + new_audit + s[j:]

# Problem detail is server-authored and sanitized, not a provider-error passthrough.
repl(
"""`pointer` is a valid RFC6901 pointer rooted at `/path`, `/query`, `/header`, or `/body`. `ProblemError` has no rejected-value/meta/code bag and never echoes sensitive rejected values; machine branching remains on the top-level `Problem.code`.""",
"""`pointer` is a valid RFC6901 pointer rooted at `/path`, `/query`, `/header`, or `/body`. `ProblemError` has no rejected-value/meta/code bag and never echoes sensitive rejected values; machine branching remains on the top-level `Problem.code`. Top-level `detail` is server-authored/sanitized human text and never carries raw provider/database/scanner errors, tokens, headers, SQL, stack traces or rejected secrets."""
)

# Presence/discriminator laws are executable contract obligations, not prose-only wishes.
repl(
"""`RevisionView.current_submission_id` is present **iff** `state=submitted`; every other RevisionState forbids it. `DocumentCreationOptionsView.default_responsible_owner` is the current actor; candidate-list presence remains exactly the §2.7 owner-manage rule. Raw WorkingContent generation is never public; ETag is wire OCC authority.""",
"""`RevisionView.current_submission_id` is present **iff** `state=submitted`; every other RevisionState forbids it. `DocumentCreationOptionsView.default_responsible_owner` is the current actor; candidate-list presence remains exactly the §2.7 owner-manage rule. Required/forbidden-member and discriminator laws in this registry are encoded as closed OAS branches where OAS 3.0.3 can represent them; value-relational laws that would require distortion remain explicit contract-fixture assertions. Raw WorkingContent generation is never public; ETag is wire OCC authority."""
)

repl("# 8. Bounded upstream findings exposed by T8-E — RESOLVED", "# 8. Bounded upstream findings exposed by T8-E")

# Replace the final-open finding with the consolidated evidence-triggered package.
start = "## 8.4 Final Lead cross-layer finding"
end = "\n---\n\n# 9. Generation / provider feasibility and runtime conformance proof"
if s.count(start) != 1 or s.count(end) != 1:
    raise SystemExit("section 8 open-finding anchors are not unique")
i = s.index(start)
j = s.index(end, i)
open_findings = """## 8.4 T8-D — transaction-census + expired-idempotency precision — OPEN

The final T3↔T8-D parity attack found a bounded owner-local persistence/transaction correction. It adds no table, owner, state, API, permission, worker or capability.

```text
1. Company replacement currently says -> Audit
   T3 does not require semantic Company-display replacement Audit
   -> subtract that mandatory Audit

2. User DISABLED -> ENABLED is ratified and T3 requires user.reenabled
   -> add explicit re-enable CAS/transition + required Audit;
      old sessions/memberships/grants remain absent

3. UserProfile replacement/erasure says "Audit when required"
   -> ordinary replacement has no mandatory semantic Audit;
      lawful erasure emits user_profile.erased when T3 requires evidence

4. DRAFT PATCH says "Audit when required"
   -> no mandatory semantic Audit for Launch autosave/WorkingContent mutation

5. Feedback says "Audit/Replay as upstream requires"
   -> immutable SubmissionFeedback + Replay only; no duplicate semantic Audit

6. OfficialRendition finalization omits T3-required rendition-completion Audit
   -> append official_rendition.completed; when the same transaction establishes Release,
      append release.completed too

7. Generic singular "Audit" shorthand on multi-effect transactions is incomplete
   -> each transaction appends ALL AND ONLY T3-required semantic events for facts/effects committed
      (including User+Binding, teardown+offboarding, Submission+Release,
       Decision+Release, requested+completed obsolescence)

8. T8-E semantic idempotency expires at completed_at+24h, while the unique key row may outlive cleanup
   -> acquisition that finds now>=expires_at serializes on that key, removes the expired Replay then Key,
      and may establish the new claim in the same transaction; the janitor is cleanup only and can never
      extend replay authority. Concurrent post-expiry reuse still has one winner/loser path.
```

This package reduces accidental Audit/job-like behavior and makes already-ratified evidence/replay semantics executable.

## 8.5 T4/T5/T8-C/T8-D — required PDF rendition when source is already PDF — OPEN

A material contradiction survived Structural Inversion:

```text
T5 RV-1:
  PDF source -> direct PDF viewer; no duplicate generated PDF without a named need

T4-J / T5-B,C,D,E / T8-C §17.2:
  frozen policy requires OfficialRendition -> render / durable rendition intent
```

For `RequireOfficialRendition(PDF)` with an already-admitted submitted PDF, PDF→PDF rendering/copying has no consumer and can change bytes without adding a product property. The smallest coherent correction is:

```text
submitted source = PDF + required format = PDF
  -> create the required OfficialRendition semantic fact over the SAME admitted handle + descriptor
  -> no provider copy
  -> no renderer execution
  -> no River rendition intent
  -> T3 official_rendition.completed Audit
  -> representation gate satisfied synchronously

submitted source = DOCX + required format = PDF
  -> existing durable renderer-intent / T4 admission / OfficialRendition path remains unchanged
```

`controlled_docs.official_renditions` already permits the same managed-content handle/descriptor and no new persistence object is required. Bounded authority edits, if approved:

```text
T4-J       make rendering conditional on transformation being required
T5-B/C/D/E preserve the durable job only for DOCX->required PDF transformation
T8-C §17.2 enqueue intent iff renderer work is activated; preserve zero-or-one named intent law
T8-D SUBMIT realize same-PDF OfficialRendition synchronously; River intent only on transformation path
```

This is a reduction, not a new rendition mode.

## 8.6 T8-C/T8-D — session CSRF bootstrap reversibility — OPEN

T6 and the accepted T8-E checkpoint require:

```text
GET /api/v1/session -> session-bound csrf_token
unsafe request      -> X-CSRF-Token validated against that session
```

T8-D currently persists only `csrf_secret_digest`, and no other ratified state can reconstruct the token on a later `GET /session`. A one-way digest can validate a token that the caller already knows, but cannot bootstrap it after OIDC redirect/reload.

The smallest stateful synchronizer-token correction follows the normal server-side session pattern:

```text
authn.application_sessions
  csrf_secret BYTEA NOT NULL     // replaces csrf_secret_digest; random per session

session issue
  -> generate cryptographically random per-session CSRF secret
  -> persist it as server-side session state

session resolve
  -> authenticated session lookup returns User truth + opaque CSRF token material to Application
  -> GET /session emits CsrfToken

unsafe request
  -> session cookie still authenticates
  -> supplied X-CSRF-Token constant-time compares with the session CSRF secret/token
```

The CSRF secret is **not** an authentication bearer token and grants nothing without the valid HttpOnly session cookie. No second cookie, localStorage credential, global CSRF-HMAC rotation subsystem or new endpoint is introduced. T8-C needs only a bounded session-resolve result precision; T8-D changes one session field from non-reconstructible digest to server-held synchronizer secret. This matches OWASP's stateful Synchronizer Token Pattern, where a per-session token is stored in the server-side session and returned in response content for the client to echo in a custom header.
"""
s = s[:i] + open_findings + s[j:]

# Mark the PDF-source no-job result as a candidate pending upstream reconciliation, not already-final authority.
repl(
"""PDF-source RequireOfficialRendition creates no duplicate bytes/job
```""",
"""PDF-source RequireOfficialRendition creates no duplicate bytes/job after §8.5 reconciliation
```"""
)

repl(
"""duplicate PDF rendition bytes for PDF source
persisted/client-authored expected-size descriptor truth""",
"""duplicate PDF rendition bytes/job for PDF source (candidate; §8.5 reconciliation pending)
persisted/client-authored expected-size descriptor truth"""
)

repl(
"""The two upstream material findings were resolved by the smallest owner-local corrections: two persistence fields for truthful Step-label history and one impossible Audit event removed.""",
"""The earlier Step-label and impossible-binding-Audit findings are resolved. The final evidence-triggered bounded coherence package is §§8.4–8.6; none changes Product scope, the 78-operation census, ownership topology or Launch lifecycle."""
)

old_gate = """The measurement, generated-boundary feasibility, provider presign feasibility, strict-request validator split and 78-row ledger-census fixture obligations are closed at candidate level. The final Lead attack is complete except for the bounded T8-D transaction-census package in §8.4, which crosses an already-ratified authority and therefore requires operator adjudication.

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
new_gate = """The measurement, generated-boundary feasibility, provider presign feasibility, strict-request validator split and 78-row ledger-census fixture obligations are closed at candidate level. The final Lead attack exposed the bounded cross-layer coherence package in §§8.4–8.6; because it touches already-ratified owners, it requires explicit operator adjudication before durable edits.

Remaining Lead gate:

```text
A. operator adjudication of §§8.4–8.6 bounded coherence package
B. if approved, reconcile only the implicated T4/T5/T8-C/T8-D lines
C. rerun whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence exact-delta check
D. revalidate main/base + exact candidate HEAD + intended authority/work diff + required CI
E. only if A→D converge, create review/t8e-fable from that exact candidate HEAD
F. independent Fable challenge
G. Lead adjudication of Fable evidence
H. explicit operator ratification
```

Until A→D converge:"""
repl(old_gate, new_gate)

repl(
"""External evidence checked during T8-E includes OpenAPI 3.0.3, RFC9110, RFC9457, RFC9530, Fetch forbidden-header behavior, OWASP archive/upload resource controls, current AWS S3 PutObject/presigning behavior, controlled-document upload limits, Stripe/Adyen idempotency practice, and current `oapi-codegen` / `openapi-typescript` behavior.""",
"""External evidence checked during T8-E includes OpenAPI 3.0.3, RFC9110, RFC9457, RFC9530, Fetch forbidden-header behavior, OWASP archive/upload controls and stateful Synchronizer Token Pattern guidance, current AWS S3 PutObject/presigning behavior, controlled-document upload limits, Stripe/Adyen idempotency practice, and current `oapi-codegen` / `openapi-typescript` behavior."""
)

proposal.write_text(s)

roadmap = Path("docs/roadmap.md")
r = roadmap.read_text()

def rrepl(old, new, count=1):
    global r
    n = r.count(old)
    if n != count:
        raise SystemExit(f"roadmap anchor mismatch expected={count} got={n}: {old[:140]!r}")
    r = r.replace(old, new, count)

old = """The final Lead coherence attack found one remaining bounded upstream precision package in T8-D's **transaction census only**: remove mandatory Audit where T3 does not require it, make required T3 evidence/multi-event paths explicit, make SUBMIT River intent conditional on a real activated effect, and name the already-ratified User re-enable transition. No table/state/API/permission/worker/capability is added.

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
new = """The final Lead coherence attack exposed one consolidated **bounded upstream coherence package**, recorded in `docs/work/current/proposal.md` §§8.4–8.6:

```text
T8-D               transaction/Audit + expired-idempotency precision
T4/T5/T8-C/T8-D    no renderer/job for already-PDF required PDF rendition
T8-C/T8-D           reconstructible server-side CSRF synchronizer secret for GET /session
```

The package is subtractive/precision-only: no Product operation, owner, lifecycle state, permission, table family, generic worker, second cookie or new API is added.

## Exact next action

```text
operator adjudicates the bounded coherence package in proposal §§8.4–8.6
→ if approved, reconcile only the implicated T4/T5/T8-C/T8-D lines
→ rerun exact whole-candidate Structural Inversion / YAGNI / overengineering / global-coherence delta
→ revalidate main/base + exact candidate HEAD + intended authority/work diff + required CI
→ only then create isolated review/t8e-fable from exact candidate HEAD
→ independent Fable challenge
→ Lead adjudication
→ explicit operator ratification
```"""
rrepl(old, new)
roadmap.write_text(r)
