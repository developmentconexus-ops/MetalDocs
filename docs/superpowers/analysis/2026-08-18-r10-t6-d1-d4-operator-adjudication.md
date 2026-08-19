# R10-T6 — D1→D4 Operator Adjudication

> **Status:** OPERATOR-APPROVED / BOUNDED PRECISION AMENDMENTS  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED  
> **T7:** NOT OPEN

This record captures the operator's explicit approval of the four bounded findings produced by `2026-08-18-r10-t6-bounded-coherence-delta-review.md`.

The approval does not reopen the product model, ownership topology, T1, T2, T4 or T5. It authorizes only the exact precision changes below.

## D1 — Idempotency replay disclosure remains subject to current access authorization

Accepted.

Before returning a completed `Idempotency-Key` replay response, MetalDocs must authenticate the current ApplicationSession and re-evaluate the current T3 permission/scope plus the minimum current resource-visibility predicate required to disclose the stored response.

The already-completed mutation's historical lifecycle preconditions are **not** re-executed. Replay is not a new business transition.

Therefore:

```text
current caller no longer authorized to receive the resource/result
→ no replay disclosure
→ normal current authorization failure semantics

current caller still authorized
+ same scoped key/fingerprint
→ exact completed replay may be returned
```

The replay store never becomes access authority.

## D2 — GroupMembership current-read surface

Accepted.

Access administration receives one current read surface:

```text
GET /api/v1/groups/{group_id}/members?cursor=...&limit=...
```

Authorization:

```text
access.manage
+ current Company scope
```

It returns a cursor-paginated bounded `UserReference`/membership representation sufficient for access administration. It is not a general User-profile directory and does not expose unnecessary PII.

No new owner, permission or GroupMembership history model is introduced.

## D3 — least-privilege document-creation/options projection

Accepted.

The create-document and owner-selection journeys receive a purpose-built read model rather than consuming administrative User/Area/DocumentType APIs.

Conceptual application surface:

```text
GET /api/v1/document-creation/options
```

The projection is computed from current canonical truth and current T3 authorization for the requesting actor.

It may return only currently usable references such as:

```text
DocumentTypeReference[]
AreaReference[]
eligible TemplateReference[]
responsible_owner_candidates[]   # only when actor has document.owner.manage in relevant scope
```

Laws:

```text
no administrative configuration fields
no general UserProfile directory
no permission bypass
no provider role/group inference
no generic reference-data platform
```

Template options require the exact current T3 effective-read permission for the selected Template plus current Template-role/eligibility/current-EFFECTIVE predicates. Owner candidates follow D4.

## D4 — responsible-owner target eligibility

Accepted as the smallest bounded T3 clarification.

For create-time deliberate owner selection or later owner replacement, an eligible target is exactly:

```text
existing MetalDocs User
+ same Company
+ current eligibility = ENABLED
```

Assignment as responsible owner:

```text
does NOT grant Role
does NOT grant Permission
does NOT import provider roles/groups
does NOT by itself grant document.read_working/document.edit/document.submit
```

The responsible-owner relationship remains a Controlled Documents relationship predicate consumed by T3; actual actions still require the normal T3 grant + scope + domain-predicate equation.

An offboarded/disabled User remains truthful historical attribution where already recorded but is not eligible for a new current owner assignment.

## Exact reopen result

```text
T1 = EMPTY
T2 = EMPTY
T3 = §9 responsible-owner target eligibility phrase only — APPROVED CLARIFICATION
T4 = EMPTY
T5 = EMPTY
T6 = D1→D3 contract/read-surface precision — APPROVED
```

Everything else remains frozen.

## Next gate

```text
apply D1→D3 to T6 platform summary
→ apply D4 to T3 durable authority
→ reconcile Decision Registry
→ run exact D1→D4 bounded delta review
→ if clean: return platform-facing T6 summary to explicit operator ratification
```

No implementation plan or product code is authorized.
