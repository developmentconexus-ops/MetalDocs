---
id: repository-status
kind: authority
owner: architecture
summary: Sole current MetalDocs stage, gate, and next-action authority.
---

# Current status

```text
REPOSITORY MODE                              CLEAN-SLATE / ARCHITECTURE-FIRST
PRODUCT CONTRACT                             OPERATOR-APPROVED
WHOLE-PRODUCT ALIGNMENT / OWNERSHIP          OPERATOR-APPROVED
T1 → T8-D                                    CLOSED / OPERATOR-RATIFIED
T8-E EXECUTABLE WIRE CONTRACT                ACTIVE
T8-F → T12                                   NOT OPEN
IMPLEMENTATION                               BLOCKED
LEGACY IMPLEMENTATION                        REMOVED FROM LIVE TREE
```

## Current authority rule

The removed implementation is Git provenance, not a codebase to maintain. It may be inspected only for a concrete proof-backed reuse question.

Do not create application code, schemas, deployment, workers, frontend, or dormant capability while implementation is blocked.

## Active gate

T8-E must finish the exact executable OpenAPI 3.0.3 application contract.

Approved T8-E design layers already cover:

```text
single application OpenAPI SSOT and generated Go/TypeScript boundaries
HTTP-native ETag / If-Match / idempotency / CSRF / pagination
closed RFC 9457 Problem Details profile
errors.conexus.fun/{product_namespace}/{code}
create-only upload and server-authoritative completion
exact-byte response contract
Submission / Governance / Obsolescence / Release / OfficialRendition views
```

The active proposal is `docs/work/current/proposal.md`.

## Exact next action

Close the remaining T8-E executable schema ledger:

```text
78-operation census
→ exact request/success field + enum + nullability matrix
→ operation-specific problem-code matrix
→ list filters/order/allowed_actions contract
→ upload/content size measurements
→ generated Go + TypeScript conformance proof design
→ subtractive/global-coherence pass
→ final Fable review
→ operator ratification
```

Only then may T8-F open.