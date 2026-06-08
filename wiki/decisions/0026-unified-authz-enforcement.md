# ADR 0026 — Unified authz is the only enforcement model (no dead ABAC path)

> **Status:** Accepted 2026-06-08 (formalises anchor decision AD-3 of the API Contract Hardening Program; extends ADR [`0022`](0022-authz-capability-coherence.md). Shipped in Phase B 2026-06-05). Retroactive ADR.
> **Last verified:** 2026-06-08
> **Scope:** The single per-resource authorization model for documents/search; confirmation that the pre-unification `AccessPolicy`/`document_access_policies` ABAC concept is dead and is not an enforcement path anywhere.
> **Out of scope:** The two-tier capability/area model itself (ADR 0007); authz-area spec markers (ADR 0023).
> **Key files:**
> - `internal/modules/iam/authz/authz.go` — tier-2 `Require` (capability + area, system_admin bypass)
> - `controlled_document_{area,user}_grants` tables — the live per-resource visibility mechanism
> - `internal/modules/search/infrastructure/v2documents/reader.go` — search visibility predicate (verbatim port of the controlled-documents list predicate)
> - migration `0232` — `document_access_policies` table dropped

## Context

A 4-subagent audit (2026-06-05) found two enforcement gaps, both tails of a dead ABAC slice predating the authz unification: (B1) approval sign-off accepted a `password_token` but never verified it; (B2) v2 search default-allowed when no `AccessPolicy` rows existed, so any authenticated user could find every tenant document. The `AccessPolicy` / `document_access_policies` concept had zero runtime, no handler, no tripwire, and was paradigmatically alien to the role+area model — yet code paths still referenced it.

## Decision (AD-3)

**Per-resource authorization flows through exactly one model:** capability (tier-1) + area (tier-2, `authz.Require`) + the per-resource `controlled_document_{area,user}_grants`, all backstopped by the Postgres capability tripwire. The pre-unification `AccessPolicy` / ABAC concept is dead: its table was dropped (migration 0232) and it must not be an enforcement path anywhere. Search enforces visibility with a **verbatim port** of the controlled-documents list predicate (one source of truth — no second authz system). E-signature sign-off verifies the credential (bcrypt) inside the tx before recording, through the existing auth port (no accepted-but-unverified state).

## Consequences

- One visibility predicate, reused by the controlled-documents list and search — drift between them is impossible by construction.
- The dead `AccessPolicy` search slice (`decidePolicies`/`ListAccessPolicies`/`shouldBypassPolicy` + the `AccessPolicy` domain type) was removed in Phase B; no fail-open default-allow remains.
- Sign-off is a real 21 CFR Part 11-style control: bad/blank credential → rejected, verified flag server-set (unforgeable), raw token never stored.
- New per-resource access features extend the unified grants model, not a revived ACL table.

## References
- `wiki/backlog/api-contract-hardening.md` Phase B (F-SEC-REAUTH, F-SEC-SEARCH, F-DEAD-SEARCHPOLICY) — Evidence 2026-06-05
- ADR [`0022-authz-capability-coherence.md`](0022-authz-capability-coherence.md) (the two-tier coherence program this extends), ADR [`0007-two-tier-authz.md`](0007-two-tier-authz.md)
- ADR [`0024-openapi-single-base-path.md`](0024-openapi-single-base-path.md), ADR [`0025-error-envelope-rfc9457.md`](0025-error-envelope-rfc9457.md)
