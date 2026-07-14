# Module #11 Review — `internal/modules/search`

**Date:** 2026-05-22
**Reviewers:** ecc:go-reviewer
**Severity totals:** 4 Critical / 4 High / 6 Medium / 3 Low
**Files reviewed:**
- `domain/model.go`, `domain/port.go`
- `application/service.go`
- `delivery/http/handler.go`
- `infrastructure/v2documents/reader.go`

---

## Critical

### C1 — `delivery/http/handler.go:47` — no authentication gate → unauthenticated access to full document corpus

`handleSearchDocuments` has no session check. Any unauthenticated caller reaches `SearchDocuments`. The `shouldBypassPolicy` path in the service then skips all access-policy enforcement because no IAM context is installed, making the entire document corpus readable without credentials.

**Recommend:** add an authn middleware (or explicit check at handler entry) that rejects requests with no valid session. Return HTTP 401 before delegating to the service.

**Fix branch:** `fix/search-11-authz-c1-c2-c3`

---

### C2 — `application/service.go:168` — absence of IAM context bypasses policy enforcement

`shouldBypassPolicy` returns `true` when no IAM context is present, treating misconfiguration as a signal to bypass all enforcement rather than to deny. A misconfigured mux, route alias, or test wiring that omits the authn middleware silently grants full access to every document.

**Recommend:** invert the default — absence of IAM context should return `false` in `canView` (deny). Reserve bypass only for an explicit, trusted internal capability marker in context.

**Fix branch:** `fix/search-11-authz-c1-c2-c3`

---

### C3 — `infrastructure/v2documents/reader.go:20` — `ListDocuments` no `tenant_id` filter → cross-tenant IDOR

```sql
SELECT ... FROM public.documents d ... -- no WHERE tenant_id = $N
```

`ListDocuments` fetches all rows across all tenants. The service layer adds no tenant scope either. Any authenticated user in any tenant can read documents belonging to other tenants.

**Recommend:** add `tenantID` parameter to `ListDocuments` (and to `domain.Reader` interface — see C4), filter with `WHERE d.tenant_id = $1 AND d.archived_at IS NULL`, extract `tenant_id` from the verified authn context in the handler.

**Fix branch:** `fix/search-11-authz-c1-c2-c3`

---

### C4 — `domain/port.go:6` — `Reader.ListDocuments` interface has no tenant parameter

The interface accepts no tenant identifier, making tenant scoping impossible at the infrastructure layer without a breaking change.

**Recommend:** `ListDocuments(ctx context.Context, tenantID string) ([]Document, error)` now, before implementations proliferate.

**Fix branch:** `fix/search-11-authz-c1-c2-c3`

---

### C5 — `infrastructure/v2documents/reader.go:20` — no `LIMIT` clause → unbounded full-table scan per request

Every search request loads the entire `documents` table into memory, then filters and paginates in Go. Memory exhaustion and latency DoS vector.

**Recommend:** push filtering and `LIMIT` into SQL. Pass the caller's effective limit (capped at `maxLimit`) as a query parameter.

**Fix branch:** `fix/search-11-unbounded-c5`

---

## High

### H1 — `application/service.go:176` — empty policy slice → unconditional allow (policy layer provides zero enforcement for v2 documents)

`decidePolicies` is deny-only with a default-allow fallback (line 198). `ListAccessPolicies` in the v2 reader always returns an empty slice → unconditional allow for every document, for every user. The access-policy layer provides no enforcement for v2 documents.

**Recommend:** document this as an explicit architectural decision with a TODO, or implement a minimum capability-grant check so an empty policy set does not silently open everything.

---

### H2 — `application/service.go:140` — `policiesForDocument` issues up to 3 DB round-trips per document → O(3N) queries per search

`policiesForDocument` calls `ListAccessPolicies` up to 3 times per document inside the per-document loop. For N documents: O(3N) round-trips per request.

**Recommend:** pre-load all policies outside the per-document loop, or push access filtering into SQL via a join/subquery.

---

### H3 — `delivery/http/handler.go:91` — service error mapped to generic 500 with no logging

Failed search requests produce a generic 500 with no server-side log. Operators have no visibility.

**Recommend:** log the error with trace ID and request context before writing the 500. Distinguish `context.Canceled`/`context.DeadlineExceeded` (408/499) from true 500s.

---

### H4 — `domain/port.go:5` — `Reader` bundles `ListDocuments` + `ListAccessPolicies` in one interface; v2 stub silently returns nil for policies

Two semantically distinct capabilities share an interface. The v2 infrastructure stubs `ListAccessPolicies` with nil, meaning every consumer of `Reader` implicitly depends on a method it cannot meaningfully use.

**Recommend:** split into `DocumentReader` and `PolicyReader`. Inject separately into the service so the v2 stub is explicit, type-safe, and impossible to silently misuse.

---

## Medium

### M1 — `domain/model.go:7` — `Classification`, `Status`, `Effect`, `SubjectType`, `Capability` are bare `string`

Invalid values compile silently; no exhaustiveness protection.

**Recommend:** `type Classification string; const ClassificationPublic Classification = "PUBLIC"` etc. for at least `Classification`, `Status`, `AccessPolicy.Effect`.

---

### M2 — `domain/model.go:5` — `Document`, `Query`, `AccessPolicy` exported structs with no constructors

Zero-value instances with empty required fields (e.g. `Document.ID == ""`) pass through silently.

**Recommend:** `NewQuery(...)` and `NewDocument(...)` constructors, or `Validate() error` called at the service boundary.

---

### M3 — `application/service.go:52` — all filtering done in Go after full-table fetch

Every filter field (document type, profile, family, area, subject, owner, classification, status, tag, expiry) is applied in memory. Wasteful at any non-trivial scale.

**Recommend:** pass `Query` fields as SQL `WHERE` predicates with parameterized placeholders. Retain in-memory filtering only as a fallback for unindexed fields.

---

### M4 — `delivery/http/handler.go:14` — handler holds concrete `*searchapp.Service`

Couples HTTP layer to implementation. Unit tests require a full service graph.

**Recommend:** define local `Searcher interface { SearchDocuments(ctx, Query) ([]Document, error) }` in the handler package.

---

### M5 — `application/service.go:227` — `cloneOptionalUTC` defined but never called (dead code)

**Recommend:** remove. Flag via `staticcheck` if not already.

---

### M6 — `delivery/http/handler.go:134` — `X-Trace-Id` header echoed from client request → log poisoning

An arbitrary client-supplied string is used as trace ID in structured logs.

**Recommend:** generate trace ID server-side. Sanitize (length cap, alphanumeric) if client header is accepted.

---

## Low

### L1 — `infrastructure/v2documents/reader.go:43` — `rows.Next()` loop does not check context cancellation between iterations

Cancelled context does not abort the scan loop until next DB I/O (driver-dependent).

**Recommend:** check `ctx.Err()` at top of loop, or confirm driver propagates context cancellation through cursors.

---

### L2 — `infrastructure/v2documents/reader.go:43` — `doc.Tags` nil slice serializes as `null` not `[]`

**Recommend:** initialize as `doc.Tags = []string{}` in scan loop for consistent JSON shape.

---

### L3 — `infrastructure/v2documents/reader.go:58` — `doc.DocumentType = doc.DocumentProfile` silent aliasing undocumented

**Recommend:** add explanatory comment, or fix query to select the real document type column.

---

## Fix Branch Index

| Branch | Covers | Land order |
|--------|--------|-----------|
| `fix/search-11-authz-c1-c2-c3` | C1 no authn gate + C2 bypass-on-no-context + C3/C4 no tenant_id | 1st (all one authz chain) |
| `fix/search-11-unbounded-c5` | C5 no LIMIT → full-table scan | 2nd |
