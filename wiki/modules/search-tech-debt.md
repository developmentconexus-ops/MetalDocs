# Tech Debt Register - search

> Companion to `wiki/modules/search.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-06-05 (T-003 closed: AccessPolicy ABAC path removed in Phase B API contract hardening)

## Items

### T-001 · Route still emits legacy `{error:{code,message}}` envelope
- **Severity:** major
- **Surface:** `internal/modules/search/delivery/http/handler.go:141`
- **Observation:** search handler writes custom API error envelope instead of RFC 9457 Problem+JSON.
- **Evidence:** `writeAPIError` helper in handler.
- **Linked backlog row:** `R-001`
- **Linked ADR:** missing-ADR

### T-002 · Reader currently covers documents only; template search path is deferred
- **Severity:** minor
- **Surface:** `internal/modules/search/infrastructure/v2documents/reader.go:19`
- **Observation:** indexing scope is narrower than module stub ambition.
- **Evidence:** only `ListDocuments` path is implemented in infrastructure.
- **Linked backlog row:** `R-002`
- **Linked ADR:** missing-ADR

### ~~T-003 · Access policy resolver is in-memory policy composition without DB-backed ACL joins~~ — CLOSED 2026-06-05
- **Severity:** ~~minor~~ closed
- **Resolution:** Dead `AccessPolicy` ABAC path (`decidePolicies`, `shouldBypassPolicy`, `matchesPolicySubject`, `policiesForDocument`, `canView`) removed entirely in Phase B API contract hardening. The `document_access_policies` table was dropped in migration 0232. Per-document visibility is now enforced at the data layer in `internal/modules/search/infrastructure/v2documents/reader.go` using the unified model (AD-3 / ADR 0022). `domain/port.go` no longer declares `ListAccessPolicies`; `domain/model.go` no longer carries `AccessPolicy`/`SubjectType`/`Effect` types.

## Coverage stats

- Public symbols undocumented: n/a (not fully audited)
- Operations missing C4 placement: n/a (stub-level doc)
- Cross-deps missing in section map: n/a (stub-level doc)
- State transitions missing: n/a (read-oriented module)
- Decisions without ADR link: 2 (T-001, T-002)
