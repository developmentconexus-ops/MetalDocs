# Stage-2 Evaluation — Performance & Query Patterns at Scale

> **Theme:** performance-scale
> **Produced:** 2026-06-11
> **Findings evaluated:** F-10 (N+1 and full-scan reads), F-20 (correlated/sequential SQL; in-memory rate-limit state)
> **Standards applied:** PostgreSQL indexing docs (GIN/GiST, index scan conditions), use-the-index-luke.com (sargability, N+1), keyset vs offset pagination (Markus Winand "SQL Performance Explained"), OWASP ASVS v4.4 §2.2 (anti-automation / rate limiting)
> **Size posture:** sane growth (hundreds of users, tens of thousands of documents, single Postgres instance). NOT hyperscale. KEEP/DEFER where current volume makes a fix premature.

---

## F-10 — N+1 and Full-Scan Read Patterns

### 1. Current state (code-confirmed)

Every claim in the register was re-confirmed against the anchored code:

| Sub-finding | File:line | Confirmed? |
|---|---|---|
| N+1 role queries in `ListUsers` (one `RolesByUserID` call per user) | `internal/modules/auth/application/service.go:432-451` (TODO comment at :432) | Yes |
| `RolesByUserID` makes two DB round trips per call (liveness check then roles query) | `internal/modules/iam/infrastructure/postgres/role_provider.go:20-57` | Yes |
| `PeopleService.ListFiltered` loads **entire** tenant user list then filters/paginates in Go | `internal/modules/iam/application/people_service.go:509-601` | Yes — calls `s.auth.ListUsers` which triggers N+1 chain above |
| `VerifyUserInTenant` calls `ListUsers` for a single membership check | `internal/modules/iam/application/people_service.go:585-601` | Yes — full list scan to find one user |
| `ListPendingForActor` uses SELECT DISTINCT IDs then `LoadInstance` per ID in a loop | `internal/modules/documents/approval/application/read_service.go:172-232` | Yes |
| No SQL-side pagination in search; offset hard-coded to 0 at call site | `internal/modules/search/application/service.go:53` (passes `offset=0` to reader) | Yes |
| `audit` ILIKE on `payload::text` (full-table sequential scan on JSONB) | `internal/modules/audit/infrastructure/postgres/writer.go:260-264` | Yes |
| `document_profiles` subquery duplicated identically in SELECT and WHERE | `internal/modules/search/infrastructure/v2documents/reader.go:34-41, 59-66` | Yes — same subquery verbatim in both positions |

Additional observation from code reading: `ListFiltered` calls `s.memberships.ListActive` per user inside the loop (line 530), compounding the N+1 beyond what the register noted. For N users each with M memberships this is O(N) IAM list calls + O(N) membership list calls before any filtering.

### 2. Standard

**N+1 antipattern:** Markus Winand, "SQL Performance Explained" ch. 2, and use-the-index-luke.com/sql/where-clause/slow-queries-do-not-use-indexes — a loop that issues one query per result row degrades as O(N) round trips; even sub-millisecond queries accumulate to seconds at hundreds of rows.

**Batch/JOIN pattern:** PostgreSQL docs §7.2 (subqueries and JOINs), §14.3 (explicit JOINs generally faster than correlated subqueries for same data). The correct fix is one query with a JOIN or an `IN (ids...)` batch.

**Keyset vs offset pagination:** "SQL Performance Explained" ch. 6; Markus Winand "Pagination Done the PostgreSQL Way" (use-the-index-luke.com/sql/partial-results/fetch-next-page). Offset pagination requires a full-table scan to the offset position; keyset (cursor-after-last-seen-ID) is O(1) seek via index. At ~10k rows the difference is measurable; at ~100k rows offset is impractical.

**JSONB full-text search:** PostgreSQL docs §8.14.4 (GIN indexes for JSONB) — `payload::text ILIKE '%needle%'` bypasses all indexes; a GIN index on the JSONB column or a `tsvector` stored column is the standard fix.

**Subquery deduplication:** PostgreSQL docs §7.8 (CTEs) and §14.3.4 (subquery factoring) — identical subqueries in SELECT and WHERE cannot be memoized by the planner unless expressed as a CTE or `LATERAL` join; they are evaluated independently on each row.

### 3. Verdict and rationale

**F-10 verdict: REFACTOR (P1) — but sub-findings are not equal. Triage below.**

The N+1/full-scan family spans six distinct sub-problems across four modules. Lumping them into one ticket produces a sprawling unfocused change. The correct decomposition:

#### F-10a — `RolesByUserID` 2-round-trip pattern (REFACTOR, S, contained)

**REFACTOR.** The two-query pattern (liveness check then roles query) in `role_provider.go:20-57` exists because the liveness check was likely added after the roles query. A single query can return `(roles[], is_active)` in one round trip using a `LEFT JOIN` and then return `ErrUserNotFound` if no rows come back. This is a direct improvement with near-zero blast radius: one infrastructure file, one interface method signature unchanged.

Evidence: PostgreSQL docs §7.2.1 — a `LEFT JOIN` on `iam_users` + `iam_user_roles` returns both existence and roles atomically; `sql.ErrNoRows` signals the user is absent without a separate round trip.

#### F-10b — `ListUsers` N+1 role queries in auth service (REFACTOR, M, module)

**REFACTOR.** `service.go:432-451` iterates users and calls `RolesByUserID` per user. With F-10a fixed, a batch variant `RolesByUserIDs(ctx, []userID, tenantID) (map[userID][]Role, error)` eliminates the N round trips to 1. The TODO comment at line 432 acknowledges this debt explicitly. Blast radius: auth service + IAM role provider interface — two files, no handler changes.

Volume context: at current size (tens to low hundreds of users per tenant) the N+1 produces tens of round trips, each ~1-3ms = ~100-300ms total on a cold cache. That is already observable latency for an admin list endpoint. Not a crisis but real.

#### F-10c — `PeopleService.ListFiltered` full load + in-Go pagination (REFACTOR, M, module)

**REFACTOR.** This calls `ListUsers` (which itself has the N+1) and then paginates in Go. The comment at `people_service.go:504-508` explicitly defers this to PR-5/PR-11. With F-10b landed, the fix is a SQL projection that applies filters and returns only the requested page. The in-Go pagination also has a correctness bug: `VerifyUserInTenant` calls `ListUsers` as a membership guard — this triggers the full N+1 chain for every password-reset or unlock action on a single user. That particular path should be a point lookup (`EXISTS (SELECT 1 FROM iam_users WHERE user_id=$1 AND tenant_id=$2)`) not a full scan.

#### F-10d — `VerifyUserInTenant` full-list scan (REFACTOR, S, contained)

**REFACTOR.** Replace `ListUsers` scan with a direct `EXISTS` query or a dedicated `IsUserInTenant(ctx, tenantID, userID) (bool, error)` method on the repository. One-line change at the call site; one new repository method. Blast radius: one file each in application and infrastructure layers.

#### F-10e — `ListPendingForActor` SELECT DISTINCT then LoadInstance loop (REFACTOR, M, module)

**REFACTOR.** The pattern is: fetch N IDs, then call `LoadInstance` per ID in a loop. `LoadInstance` is itself a multi-query function (see approval module; it loads the instance with its stages and signoffs). This is the highest-volume N+1 in the set — it fires on every "inbox" page load by any approver. The correct fix is a single SQL JOIN that returns the instance rows directly, or a batch `LoadInstancesByIDs` that issues one query per table (instances, stages, signoffs) and assembles in Go. Medium effort because `LoadInstance` builds a complex aggregate; this is the one sub-finding that genuinely needs design thought before implementation.

#### F-10f — Audit ILIKE on `payload::text` (REFACTOR, M, module)

**REFACTOR.** `payload::text ILIKE '%needle%'` is a full-table sequential scan. The standard fix for PostgreSQL is either:
- A GIN index on the JSONB column itself: `CREATE INDEX ON audit_events USING GIN (payload)` enables `payload @> '{"key":"val"}'::jsonb` lookups but NOT arbitrary substring search.
- A `tsvector` stored column populated by a trigger or migration: enables `to_tsvector` + `@@` operator with a GIN index, supporting full-text search.
- For simple substring on the text cast: `CREATE INDEX ON audit_events ((payload::text) text_pattern_ops)` — only helps anchored LIKE (not leading-wildcard `%needle%`).

The ILIKE with leading wildcard is **not indexable** under any of these options. The correct path is to define the search fields explicitly (action, actor_id, resource_id are already separate columns and are indexed) and drop the `payload::text` catch-all, or accept its cost for infrequent audit admin searches. Audit log queries are admin/compliance-tier, not user-hot-path. At current size (<100k events) a sequential scan completes in <100ms. At 1M events it becomes seconds. This is a **medium-term** cliff.

#### F-10g — `document_profiles` subquery duplication in search reader (SIMPLIFY, S, contained)

**SIMPLIFY.** The identical correlated subquery at `reader.go:34-41` (in SELECT) and `reader.go:59-66` (in WHERE) is a maintenance hazard — a change to one copy must be replicated to the other. PostgreSQL cannot memoize correlated subqueries; the planner evaluates both independently on each row. The fix is a `LATERAL` subquery or CTE factored out once:

```sql
WITH dp AS (
  SELECT d.id AS doc_id,
    (SELECT family_code FROM metaldocs.document_profiles dp2
     WHERE dp2.code = COALESCE(d.profile_code_snapshot, cd.profile_code)
       AND dp2.tenant_id IN (d.tenant_id, 'ffffffff-...'::uuid)
     ORDER BY CASE WHEN dp2.tenant_id = d.tenant_id THEN 0 ELSE 1 END
     LIMIT 1) AS family_code
  FROM public.documents d ...
)
```

Or equivalently, promote the subquery to a `LEFT JOIN LATERAL`. This eliminates the duplication, makes the intent clearer, and may allow the planner to evaluate it once per document row. Blast radius: one SQL constant in one file.

### 4. Over-engineering check

No sub-finding warrants a distributed cache, a read-replica, a materialized view, or an external search service. All fixes are standard PostgreSQL + Go patterns. The search module's visibility predicate (the complex EXISTS chains) is intentionally correct and properly avoids post-fetch filtering per REQ-AUTHZ-6 — do not simplify it.

### 5. ADR needed?

No. None of these changes alter the module boundary, the authz model, or the contract surface. They are internal implementation improvements that do not change observable API behavior.

### 6. Effort / blast radius summary

| Sub-finding | Verdict | Effort | Blast radius |
|---|---|---|---|
| F-10a: 2-round-trip RolesByUserID | REFACTOR | S | contained (1 infra file) |
| F-10b: ListUsers N+1 role queries | REFACTOR | M | module (auth service + IAM infra) |
| F-10c: ListFiltered full-load + in-Go pagination | REFACTOR | M | module (IAM application layer) |
| F-10d: VerifyUserInTenant full scan | REFACTOR | S | contained (1 applic + 1 infra method) |
| F-10e: ListPendingForActor load-in-loop | REFACTOR | M | module (approval read path) |
| F-10f: Audit ILIKE on payload::text | REFACTOR | M | module (audit infra + migration) |
| F-10g: document_profiles subquery duplication | SIMPLIFY | S | contained (1 SQL constant) |

---

## F-20 — Correlated / Sequential SQL Performance Patterns

### 1. Current state (code-confirmed)

| Sub-finding | File:line | Confirmed? |
|---|---|---|
| ILIKE on `payload::text` (full-table sequential scan on JSONB) | `internal/modules/audit/infrastructure/postgres/writer.go:260-264` | Yes — same finding as F-10f; already evaluated above |
| TODO(phase11) leading-wildcard ILIKE for CD full-text search | `internal/modules/controlleddocuments/infrastructure/repository.go:128` | Yes — `code ILIKE $N OR title ILIKE $N` with `%prefix%` pattern; TODO comment present |
| `listRoutesQuery` correlated `SELECT COUNT(*) FROM approval_routes WHERE tenant_id=$1` re-evaluated per row | `internal/modules/documents/approval/repository/postgres_approval_repository.go:437-444` | Yes — count subquery in the SELECT list of a JOIN query; evaluated once per row returned |
| `document_profiles` subquery duplication in SELECT and WHERE | `internal/modules/search/infrastructure/v2documents/reader.go:34-41, 59-66` | Yes — same as F-10g above |
| `listRoutesQuery` aliases `created_at AS updated_at`; no `updated_at` column | `postgres_approval_repository.go:437` | Yes — confirmed in the SELECT list |
| `InMemoryAuthFailureRateLimiter` wired in production | `internal/modules/documents/approval/infrastructure/signature/password_reauth.go:118-127`; `apps/api/cmd/metaldocs-api/reauth.go:49` | Yes — `NewInMemoryAuthFailureRateLimiter()` called at wiring; comment says "tests/dev only" |
| Sequential security rule evaluation (4 independent SQL queries in series) | `internal/modules/security/application/service.go:93-149` | Yes — four separate `s.repo.*` calls, each its own DB query, no early exit or parallelism |

### 2. Standard

**Correlated subqueries:** PostgreSQL docs §14.3.4 and use-the-index-luke.com/sql/where-clause/obfuscation — a correlated `COUNT(*)` in a SELECT list is re-evaluated for every output row. For a query returning R rows it fires R+1 queries (1 outer + R correlated). The standard fix is a `window function` (`COUNT(*) OVER ()`) or a pre-aggregated CTE, both of which compute the count once.

**ILIKE with leading wildcard:** use-the-index-luke.com/sql/where-clause/searching-for-ranges/like-performance — `LIKE '%pattern%'` and `ILIKE '%pattern%'` suppress index use regardless of the index type on the column itself. Only `LIKE 'prefix%'` (no leading wildcard) can use a btree index with `text_pattern_ops`. Full-text: PostgreSQL docs §12 (Full-Text Search) — `tsvector`/`tsquery` with a GIN index is the correct tool; `pg_trgm` with a GIN index supports arbitrary substring at moderate cost.

**In-memory rate limiter in production:** OWASP ASVS v4.4 §2.2.1 — "Verify that anti-automation controls are effective at mitigating breached credential testing, brute force, and account lockout attacks." A process-local counter is reset on every restart and is invisible to sibling replicas. If the API scales to 2 replicas an attacker can halve their lockout exposure by routing alternate attempts to each replica. The code's own comment (`// InMemoryAuthFailureRateLimiter is process-local and intended for tests/dev only`) acknowledges this gap.

**Sequential fan-out queries:** For independent sub-queries (security signals) that do not depend on each other's results, Go's `sync/errgroup` (golang.org/x/sync/errgroup) allows concurrent DB calls, reducing wall-clock latency from O(N×RTT) to O(max(RTT)) under stable connection pool capacity.

### 3. Verdict and rationale

**F-20 verdict: mixed — individual sub-findings below.**

#### F-20a — Audit ILIKE on `payload::text` (same as F-10f)

Already evaluated under F-10f. **REFACTOR, M, module.** Duplicate finding — combine with F-10f in the backlog.

#### F-20b — CD repository leading-wildcard ILIKE (DEFER, P2)

**DEFER.** The TODO(phase11) comment at `repository.go:128` explicitly marks this as a known deferral. At current data volumes (<10k controlled documents per tenant in a small QMS deployment) a sequential scan on `code` and `title` columns completes well under 50ms. The fix (add a `pg_trgm` GIN index and rewrite to `ILIKE` with the trigram operator, or switch to full-text) is a migration + query change. Trigger: when p95 latency on the CD list endpoint exceeds 200ms in production profiling, or when the controlled documents table exceeds ~50k rows. At that point the same TODO comment already names the fix. No action needed now. The finding is real but the TODO is appropriate.

Evidence: PostgreSQL pg_trgm docs — `CREATE EXTENSION pg_trgm; CREATE INDEX ON controlled_documents USING GIN (code gin_trgm_ops, title gin_trgm_ops)` enables ILIKE without leading-wildcard restriction; but this is premature at current volume.

**Over-engineering check:** adding pg_trgm today for a dataset that fits in memory is pure overhead. DEFER is the correct verdict.

#### F-20c — `listRoutesQuery` correlated COUNT(*) per row (SIMPLIFY, S, contained)

**SIMPLIFY.** The query at `postgres_approval_repository.go:437-444` places `(SELECT COUNT(*) FROM approval_routes WHERE tenant_id = $1::uuid) AS total_count` in the SELECT list of a JOIN against `approval_route_stages`. PostgreSQL re-evaluates this subquery once per row of the outer result. For a tenant with R routes and S average stages per route, the count fires R×S times. The fix is trivial: replace with `COUNT(*) OVER ()` as a window function, which is computed once over the full result set. Window functions are evaluated in one pass; PostgreSQL docs §3.5 (Window Functions).

Secondary defect: `r.created_at AS updated_at` — the column alias produces a misleading `updated_at` field. This is a data accuracy issue independent of performance; the field should be renamed or a real `updated_at` column added to `approval_routes`. Low blast radius: one SQL constant + any struct field consuming `updated_at`.

**ADR needed?** No. **Effort:** S. **Blast radius:** contained (one query constant, one struct field).

#### F-20d — `document_profiles` subquery duplication (same as F-10g)

Already evaluated under F-10g. **SIMPLIFY, S, contained.** Duplicate finding.

#### F-20e — `InMemoryAuthFailureRateLimiter` wired in production (REFACTOR, M, module)

**REFACTOR.** This is the most operationally significant item in F-20. The code's own comment says it is "intended for tests/dev only." Wiring it in production (`reauth.go:49`) means:

1. Rate-limit state is lost on every API restart (deploy, crash, OOM). An attacker can reset the counter by triggering a restart or waiting for a rolling deploy.
2. In a multi-replica deployment (even 2 replicas) the limiter is per-process. Total allowed attempts = `limit × replica_count`.

The correct production replacement is a Postgres-backed limiter (already partially implemented — `platform/security/ratelimit.go` has a DB-backed fixed-window implementation, though it has the domain-import problem flagged in F-05) or a Redis-backed limiter. Given the target architecture lists Redis for rate limiting (target architecture diagram, state tier), a Redis-backed counter with `INCR + EXPIRE` (Redis documentation: redis.io/commands/incr, standard sliding-window pattern) is the target. Until Redis is available, a Postgres-backed implementation using a small `reauth_failures` table (INSERT + COUNT within a time window) is acceptable and eliminates the restart-reset problem.

**OWASP ASVS v4.4 §2.2.1 and §2.2.6** require that brute-force protections be effective across all authentication paths. The signoff re-authentication path is a credential verification path; it requires a shared, persistent counter.

**ADR needed?** No new ADR — the rate limiting architecture is already scoped in RF-9 and the target architecture. This is an implementation fix, not a design decision.

**Over-engineering check:** Do not implement Redis here if Redis is not already deployed. Use a Postgres table first; it is two SQL statements and a 5-line migration. Redis is the target but "good enough now, better later" applies.

#### F-20f — Sequential security rule evaluation (KEEP, P3)

**KEEP.** The four queries in `service.go:93-149` are:
1. `CountRecentFailedLoginsByUser`
2. `CountRecentLockouts`
3. `ListNewDeviceLogins`
4. `ListOffHoursAdminActions`

These are independent reads against the `auth_identities` / `audit_events` tables. They could be run concurrently with `errgroup`. However:

- The security signals endpoint is a low-frequency admin surface (security dashboard), not a hot data-plane path.
- Each query is bounded (recent time window, small result set).
- At current data volume all four complete in <50ms total.
- Parallelising with `errgroup` adds concurrency complexity to what is currently a simple sequential function; the overhead is not justified.

**Simplicity-first rule applies:** Do not add `errgroup` concurrency to a function that is already fast enough. If profiling ever shows this endpoint as a bottleneck (>500ms), revisit. Not now.

### 4. Over-engineering check (aggregate)

The finding register frames F-20 as low severity and correctly so. The items that warrant action are the correlated COUNT (trivial window-function fix), the in-memory rate limiter (a real production risk), and the payload ILIKE (medium-term cliff). Everything else is either already deferred with appropriate TODO comments or is fine at current volume. No external search engine, no read replicas, no materialized views, no caching layer beyond what already exists is warranted by any of these findings.

### 5. Effort / blast radius summary

| Sub-finding | Verdict | Effort | Blast radius |
|---|---|---|---|
| F-20a: Audit ILIKE payload::text | REFACTOR (dup F-10f) | M | module |
| F-20b: CD leading-wildcard ILIKE | DEFER (trigger: >50k rows / >200ms p95) | — | — |
| F-20c: listRoutesQuery correlated COUNT | SIMPLIFY | S | contained |
| F-20d: subquery duplication in search | SIMPLIFY (dup F-10g) | S | contained |
| F-20e: InMemoryAuthFailureRateLimiter in prod | REFACTOR | M | module (approval infra + reauth.go wiring) |
| F-20f: Sequential security queries | KEEP | — | — |

---

## Cross-finding observations

1. **F-10f and F-20a are the same finding** (audit ILIKE on payload::text). The register entered it twice from different vantage points (F-10 via module-audit, F-20 via pattern-family). One backlog item covers both.

2. **F-10g and F-20d are the same finding** (document_profiles subquery duplication in search reader). Same consolidation applies.

3. **Sequencing:** F-10a (RolesByUserID) is a prerequisite for F-10b (ListUsers N+1) and F-10c (ListFiltered full-load). Fix them in order: F-10a → F-10b → F-10c+F-10d. F-10e (approval inbox), F-10f/F-20a (audit ILIKE), F-20c (routes COUNT), F-20e (in-memory limiter) are independent and can be parallelised.

4. **No finding in F-10 or F-20 requires an ADR.** All are internal implementation improvements with no contract-surface change, no authz model change, and no module boundary change.

---

## Citations

- PostgreSQL 16 docs §8.14.4 (GIN indexes for JSONB): https://www.postgresql.org/docs/16/datatype-json.html#JSON-INDEXING
- PostgreSQL 16 docs §12 (Full Text Search): https://www.postgresql.org/docs/16/textsearch.html
- PostgreSQL 16 docs §3.5 (Window Functions): https://www.postgresql.org/docs/16/tutorial-window.html
- PostgreSQL 16 docs §14.3 (Using EXPLAIN): https://www.postgresql.org/docs/16/using-explain.html
- use-the-index-luke.com — LIKE Performance: https://use-the-index-luke.com/sql/where-clause/searching-for-ranges/like-performance
- use-the-index-luke.com — N+1 pattern: https://use-the-index-luke.com/sql/where-clause/slow-queries-do-not-use-indexes
- Markus Winand "SQL Performance Explained" ch. 6 (Pagination): https://sql-performance-explained.com/
- OWASP ASVS v4.4 §2.2.1, §2.2.6 (Anti-automation): https://owasp.org/www-project-application-security-verification-standard/
- Redis INCR + EXPIRE pattern (sliding window rate limiting): https://redis.io/docs/latest/commands/incr/
- golang.org/x/sync/errgroup: https://pkg.go.dev/golang.org/x/sync/errgroup
