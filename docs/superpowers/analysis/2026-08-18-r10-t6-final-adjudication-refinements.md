# R10-T6 — Final Global-Maximum Adjudication Refinements

> **Status:** ACTIVE STAGING / NON-AUTHORITATIVE — **FINAL PRECEDENCE OVER T6 CANDIDATE/PACKET WHERE NAMED**  
> **Date:** 2026-08-18  
> **Repository:** `developmentconexus-ops/MetalDocs`  
> **Branch / PR:** `docs/a8-authz-approval-redesign-ledger` / PR #131  
> **Implementation:** BLOCKED

A second adversarial pass applied the already-selected T6 principle **natural HTTP idempotency first; add durable replay machinery only where the semantics genuinely need it**.

Only the four refinements below change the corrected adjudication packet. Everything else remains as proposed there.

---

## FR-1 — User eligibility is one current singleton resource

### Replaces

```text
POST /api/v1/users/{user_id}/offboarding
POST /api/v1/users/{user_id}/reenablement
```

### Final candidate

```text
GET /api/v1/users/{user_id}/eligibility
PUT /api/v1/users/{user_id}/eligibility
If-Match: <current eligibility ETag>

body:
  state = ENABLED | DISABLED
```

Semantics:

```text
ENABLED → DISABLED
  = T3 offboarding transaction
  = disable User
  + revoke ApplicationSessions
  + remove current GroupMemberships
  + remove direct User RoleAssignments
  + required Audit

DISABLED → ENABLED
  = T3 re-enable transition
  = eligibility only
  = no old Session/membership/grant resurrection

same desired state
  = idempotent no-op
  = no duplicate semantic Audit transition
```

Why Global Maximum: eligibility is already canonical **current state**. Modeling the desired singleton state as PUT gives natural retry safety and removes two command-specific replay records without weakening T3 semantics.

Stale eligibility ETag = `412 precondition.resource_changed`.

---

## FR-2 — One active Step has at most one immutable Decision resource

### Replaces

```text
POST /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decisions
```

### Final candidate

```text
GET /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
PUT /api/v1/governance-attempts/{attempt_id}/steps/{step_id}/decision
```

PUT body:

```text
outcome = ACCEPT | RETURN_FOR_CHANGES
reason  = required for RETURN_FOR_CHANGES; optional bounded note for ACCEPT
```

Semantics:

```text
no Decision yet + exact active eligible Step
→ create immutable Decision under T2/T3 serialization

exact repeat of already-created Decision
→ return existing result / idempotent success

different attempted Decision after one exists
→ 409 state.governance_step_already_decided
```

No `Idempotency-Key` is required for this operation because resource identity itself closes duplicate-creation uncertainty.

This does not make the Decision mutable: PUT is the transport expression for establishing the singleton fact once.

---

## FR-3 — DRAFT edit is explicit PATCH + strong If-Match

The corrected packet's `PUT /revisions/{id}/draft` could be misread as a full replacement while also allowing optional fields. Remove that ambiguity.

### Final candidate

```text
GET   /api/v1/revisions/{revision_id}/draft
PATCH /api/v1/revisions/{revision_id}/draft
If-Match: "draft-<expected_generation>"
```

PATCH accepts only the closed mutable DRAFT fields:

```text
title             optional set-to value
source_upload_id  optional newly READY upload reference
```

At least one field is required. Omitted field = unchanged. Null does not mean delete unless a future explicitly promoted semantic says so.

Server behavior:

```text
revalidate If-Match generation
→ if source_upload_id present, revalidate T4 READY + live admission binding and load server-owned descriptor
→ atomically apply title/source changes
→ increment the single WorkingContent generation exactly once
→ return updated DocumentWorkView + new strong ETag
```

Stale = `412 precondition.draft_changed` with zero mutation.

No durable Idempotency-Key row for DRAFT PATCH. If the network loses a successful response, a retry with the old ETag fails 412; the client reloads current DRAFT and can prove whether its desired state landed. Correctness stays with T2 OCC.

---

## FR-4 — Idempotency-Key set and retention after natural-idempotency subtraction

Remove from Idempotency-Key-required set:

```text
User offboarding / reenablement  → FR-1 PUT eligibility
Governance decision              → FR-2 PUT singleton Decision
```

Still require `Idempotency-Key` for non-idempotent semantic POST creation where a lost response could otherwise create a second fact/resource, including:

```text
POST /api/v1/users
POST /api/v1/areas
POST /api/v1/groups
POST /api/v1/role-assignments
POST /api/v1/document-types
POST /api/v1/documents
POST /api/v1/documents/{document_id}/revisions
POST /api/v1/revisions/{revision_id}/submissions
POST /api/v1/governance-attempts/{attempt_id}/feedback
POST /api/v1/documents/{document_id}/obsolescence-requests
```

The fingerprint/replay/in-flight/key-reuse semantics from the corrected packet remain unchanged.

### Replay retention

T6 architecture freezes only:

```text
retention is bounded
retention must safely exceed ordinary browser/network retry windows
expiry is transport-mechanism cleanup and never business dedupe/disposition
post-expiry correctness remains with T1/T2/T3 semantic uniqueness/eligibility
```

**24 hours remains the first implementation-default candidate, not a durable architecture invariant.** The implementation plan may select it without reopening T6 if operational evidence does not reveal a longer retry consumer; any materially longer/offline client contract must explicitly justify the retention change.

This removes an arbitrary time constant from semantic architecture while keeping implementation deterministic at planning time.

---

# Final precedence

For operator adjudication:

```text
base T6 candidate
→ corrected Global-Maximum adjudication packet
→ THIS final-refinement file for FR-1..FR-4 only
```

No other T6 decision changes.

Current gate remains:

```text
operator material adjudication NEXT
platform-facing T6 summary NOT YET
T6 durable promotion NOT YET
T7 NOT OPEN
implementation BLOCKED
```
