# R10-T6 — Bounded Coherence Delta Review After C1→C8

> **Status:** ACTIVE STAGING / BOUNDED DELTA REVIEW — **NEW MATERIAL PRECISION FINDINGS / SUMMARY RATIFICATION HELD**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Reviewed target:** corrected `docs/superpowers/analysis/2026-08-18-r10-t6-platform-facing-summary.md`  
> **Implementation:** BLOCKED  
> **T7:** NOT OPEN

This review is intentionally bounded. It verifies only the operator-approved C1→C8 + L1→L5 corrections, checks for contradictions created by those corrections, and tests whether the resulting closed operation census can actually serve the already-approved Launch journeys without bypassing T3 least privilege.

It does not restart T6 or re-review frozen T1→T5 decisions wholesale.

---

# 1. Delta verdict

```text
C1 = CLOSED
C2 = CLOSED
C3 = CLOSED
C4 = CLOSED
C5 = CLOSED
C6 = CLOSED
C7 = CLOSED
C8 = CLOSED

L1 = CLOSED
L2 = CLOSED
L3 = CLOSED
L4 = CLOSED
L5 = CLOSED

NEW MATERIAL FINDINGS = 4
DELTA VERDICT = MATERIAL PRECISION DELTA
T6 SUMMARY RATIFICATION = HELD
T7 = BLOCKED
IMPLEMENTATION = BLOCKED
```

The corrected T6 core remains valid. Three findings are T6 contract/read-surface corrections. One exposes an earlier undefined T3 phrase and needs the smallest possible T3 precision reopen.

---

# 2. Closure verification C1→C8

## C1 — CLOSED

Corrected summary defines lens-scoped derived status, defaults Library to `effective`, permits bounded history-authorized obsolete/cancelled catalog modes, and explicitly rejects persisted `Document.current_status`.

This satisfies Product Contract status discovery while preserving T1/Registry current-state authority.

## C2 — CLOSED

Corrected summary couples scoped Idempotency-Key lock/fingerprint, semantic transition, required Audit/durable intents and completed replay result in the same local PostgreSQL transaction. It removes the public/durable `IN_PROGRESS` baseline and serializes concurrent same-key requests.

Crash window from semantic commit without replay proof is closed.

## C3 — CLOSED

Corrected summary distinguishes:

```text
/api/v1 application API
/auth browser OIDC integration
operations/readiness/metrics surface
```

OpenAPI is now correctly the `/api/v1` application-contract SSOT, not every HTTP endpoint.

## C4 — CLOSED

Corrected summary restates the complete Launch lifecycle outcomes: create, no-human synchronous Release, human governance, return, withdrawal, cancellation, rendition-pending truth, first/replacement Release, human/no-human obsolescence and reader consequences.

## C5 — CLOSED

Create-next-Revision now performs managed source copy outside the semantic tx and then revalidates exact current EFFECTIVE source + lifecycle eligibility under Document serialization before creating the new Revision. Failed revalidation creates no Revision and leaves reclaimable mechanism bytes.

## C6 — CLOSED

Provider binding, responsible owner and Template role are explicit strong-ETag singleton resources with `If-Match`, `412` stale failure and zero mutation.

## C7 — CLOSED

Template administration uses a bounded `template_use.manage` metadata projection and explicitly denies implicit source/history access to governance_admin.

## C8 — CLOSED

Normalized DocumentType and Area codes are Company-unique; committed Document codes remain Company-unique/no-reuse. Exact code-length maximum is correctly deferred as non-semantic implementation-contract detail.

## L1→L5 — CLOSED

```text
interactive DOCX adapter != OfficialRendition renderer
Range read optional
allowed_actions shares canonical policy components
User/Profile/Binding/Audit local creation atomic after provider preflight
all named future seams PASS without dormant implementation
```

---

# 3. D1 — Idempotency replay must never become authorization authority

## Severity: MAJOR / T6 ONLY

The corrected C2 law makes replay crash-consistent, but the summary does not yet say **which current authorization is required before a stored response may be disclosed**.

Counterexample:

```text
User successfully POSTs a semantic command with Idempotency-Key
→ response is stored
→ later direct/Group grants are revoked but User remains enabled
→ User opens a fresh valid ApplicationSession
→ sends same key/fingerprint
→ transport returns stored response before current T3 authorization
```

That would make the idempotency mechanism a stale access authority, contradicting T3:

```text
Authorization = current live grants + scope + domain relationship predicates
mechanism/cache never grants authority
```

### Required correction

Idempotency replay pipeline must be:

```text
validate current ApplicationSession + CSRF
→ resolve current User
→ recheck current T3 permission/scope and current disclosure/access predicate sufficient to reveal the stored result
→ only then lock/read scoped Idempotency-Key
→ same fingerprint completed result may replay without re-executing the already-completed T2 mutation eligibility predicate
```

Important distinction:

```text
current access authorization is rechecked
completed business command is NOT re-executed
```

A lifecycle transition caused by the original success must not make an otherwise authorized network retry fail merely because the resource is no longer in the pre-command state. But a later access revocation must prevent the replay record from disclosing protected response data.

If current access is denied, return the normal current authorization failure and do not reveal the stored response.

Idempotency remains transport mechanism, never access grant.

---

# 4. D2 — Access administration lacks a read surface for GroupMembership

## Severity: MAJOR / T6 ONLY

T3 says:

```text
GroupMembership is current Organization truth
mutation requires access.manage
```

T6 exposes:

```text
PUT    /groups/{group_id}/members/{user_id}
DELETE /groups/{group_id}/members/{user_id}
```

but the closed application operation census has no way for Access Admin to inspect current GroupMemberships.

Embedding arbitrary membership lists in `GET /groups/{id}` would conflict with T6's potentially-unbounded-list cursor law and would blur Group identity with access configuration.

### Required correction

Add one bounded access-management query:

```text
GET /api/v1/groups/{group_id}/members?cursor=...&limit=...
```

Authorization/read model:

```text
requires access.manage
returns bounded UserReference metadata needed to administer membership
uses cursor pagination when unbounded
never grants authority itself
```

Group identity remains Organization-owned; the read projection lives in the Access admin lens because current membership is security-bearing configuration.

No new permission/owner is introduced.

---

# 5. D3 — Create/owner journeys need least-privilege reference data instead of admin APIs

## Severity: MAJOR / T6 ONLY

The approved create journey requires:

```text
choose DocumentType
choose Area
optionally choose eligible current-EFFECTIVE Template
responsible owner defaults actor; document.owner.manage may choose another eligible User
```

But the closed operation census currently exposes the obvious reference collections under Organization/Document-Governance administration. Reusing full admin User/Area/DocumentType reads for Authors/Area Managers would either:

```text
require admin permissions the actor does not have
OR
leak administrative/PII/config data merely to populate a selector
```

C7 already proved the correct pattern for Templates: purpose-bounded read models rather than implicit admin/content access.

### Required correction — one purpose-built creation/options query

Add a semantic create-journey read model, recommended shape:

```text
GET /api/v1/document-creation/options
  ?document_type_id=...
  &area_id=...
```

Server derives only currently usable bounded references:

```text
document_types
  active types for which actor currently has document.create in at least one eligible scope

areas
  active Areas inside actor's current document.create scopes

eligible_templates    # when document_type_id is supplied
  current EFFECTIVE Documents
  + current Template role
  + current target-DocumentType eligibility
  + actor currently has document.read_effective for the Template

responsible_owner_candidates   # only when actor has document.owner.manage for selected target scope
  bounded UserReference values satisfying the canonical T3 responsible-owner eligibility rule
```

No admin-only fields, provider identifiers, grants, emails or unrelated UserProfile data are exposed merely for selection.

Ordinary author needs no owner-candidate list because owner defaults to actor.

This purpose-built query is a semantic read lens, not a new owner or generic reference-data platform.

---

# 6. D4 — T3 phrase `eligible target User` is undefined

## Severity: MAJOR / MINIMAL T3 PRECISION REOPEN

T3 currently authorizes responsible-owner change with:

```text
document.owner.manage
+ matching scope
+ eligible target User
→ change current responsibility
```

but `eligible target User` is not defined in Product Contract, T1, T2, T3 or the Decision Registry.

T6 now needs that definition for both:

```text
create another responsible owner
change responsible owner
```

Leaving it to implementation would allow incompatible policies, for example:

```text
any User
only enabled User
only User with author role
only User with document.edit in target Area
provider-group member
```

Some of these would silently couple Organization ownership to current AuthZ grants or resurrect provider-derived eligibility.

### Recommended Global-Maximum clarification

Define the smallest current domain eligibility:

```text
responsible-owner target eligibility
=
current MetalDocs User exists
+ User is ENABLED at the assignment transition
+ User belongs to the same single Company deployment
```

Explicitly:

```text
responsible-owner assignment grants NO Role/Permission
responsible-owner eligibility does NOT depend on provider roles/groups
responsible-owner eligibility does NOT require current document.edit grant
```

Why not require `document.edit`?

Ownership/responsibility and Authorization are intentionally separate authorities. Requiring an edit grant as identity eligibility would make current access configuration part of the validity of the responsibility relation and create hidden coupling. A responsible owner without an authoring grant remains a valid accountable owner but cannot perform protected authoring commands until access is separately granted.

Offboarding may later leave a historical/current Document pointing to a disabled responsible owner; that does not rewrite the relation automatically. A subsequent authorized owner-management action may assign a new **enabled** User.

This is a one-phrase T3 precision clarification only. No role, permission, bundle, scope or authorization equation changes.

### Minimal reopen set

```text
T3 §9 Change responsible owner — define eligible target User only
```

Everything else in T3 remains frozen.

---

# 7. New-delta disposition

```text
D1  current authorization before idempotency replay disclosure
D2  GET GroupMembership access-admin list
D3  purpose-built document-creation/options read model
D4  define responsible-owner target eligibility = current enabled Company User; no implicit grants
```

Recommended disposition:

```text
ACCEPT D1→D4
```

Formal reopen set:

```text
T1 = EMPTY
T2 = EMPTY
T3 = { §9 eligible target User definition only }
T4 = EMPTY
T5 = EMPTY
T6 = { D1, D2, D3 contract precision }
```

Everything else remains frozen.

---

# 8. Gate

```text
C1→C8 + L1→L5                  CLOSED
new bounded delta D1→D4        OPERATOR ADJUDICATION NEXT
platform-facing summary         RATIFICATION HELD
T6 durable promotion            NOT YET
T7                              NOT OPEN
implementation                  BLOCKED
```

If D1→D4 are approved:

```text
apply D1→D3 to T6 summary
→ apply D4 bounded clarification to T3 + Registry
→ rerun exact bounded delta only over D1→D4
→ if clean, operator platform-summary ratification NEXT
```
