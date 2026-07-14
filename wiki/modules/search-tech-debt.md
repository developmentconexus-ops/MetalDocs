# Tech Debt Register - search

> Companion to `wiki/modules/search.md`. Debt only; no fix prescriptions.

**Last verified:** 2026-06-10 (Stage-1 backend audit drift patch: T-001 evidence corrected)

## Items

### ~~T-001 · 405 response body is empty (no RFC 9457 Problem+JSON envelope)~~ **DRIFT-CLOSED 2026-07-02 (no change — handler.go:66-69 calls `httpresponse.WriteMethodNotAllowed`, which emits `WriteError` RFC 9457 problem+json)**
- **Severity:** ~~minor~~ closed
- **Surface:** `internal/modules/search/delivery/http/handler.go:66-69`; `internal/platform/httpresponse/response.go:22-25`
- **Observation (original):** The method guard called `w.WriteHeader(http.StatusMethodNotAllowed)` and returned without a body.
- **Evidence:** `handler.go:66-69` now calls `httpresponse.WriteMethodNotAllowed(w, http.MethodGet)`; `response.go:22-25` shows `WriteMethodNotAllowed` sets the `Allow` header and delegates to `WriteError(w, http.StatusMethodNotAllowed, problem.CodeMethodNotAllowed, "Method not allowed")` — a full RFC 9457 problem+json body. No code change needed; row was already fixed, register was stale.
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
