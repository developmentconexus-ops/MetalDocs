# Stage-2 Security & Tenant Isolation Evaluation

> **Theme:** security-secrets
> **Findings covered:** F-18, F-12, F-11
> **Author:** Stage-2 evaluation agent (claude-sonnet-4-6)
> **Date:** 2026-06-11
> **Standards:** OWASP ASVS 4.0, OWASP Top 10 2021, CWE-798, PostgreSQL Row Security Policies docs, ISO 27001:2022 A.8.3 / A.5.15

---

## Theme summary

This theme covers the three security-posture findings from the Stage-1 register that touch secrets hygiene, DB-layer tenant isolation, and the capability authorization registry. The benchmark is OWASP ASVS Level 2 (the minimum bar for a regulated QMS application) and the project's own REQ-SEC/REQ-TEN/REQ-AUTHZ requirements. F-18 is assessed as a P0 pre-requisite: it must be resolved before any other security work can be considered meaningful, because credentials committed to VCS are potentially already exfiltrated. F-12 and F-11 are genuine defects but scoped: the DB-layer shortfalls are real and require a migration plan, but the application layer already filters by `tenant_id` on the critical paths; the capability registry gap is a partial correctness risk that requires focused surgery, not a cross-system rewrite.

---

## F-18 — Hard-Coded Credentials and Stale Binaries in VCS

### Current state

Confirmed by direct file read. Four distinct hardcoded secrets/artifacts exist:

| Artifact | Location | Content |
|---|---|---|
| Plaintext Postgres DSN with password | `cmd/seed-test-document/main.go:25` | `password='***REDACTED***'` |
| Plaintext MinIO access key + secret key | `cmd/seed-test-document/main.go:27-28` | `minioadmin` / `minioadmin` |
| Compiled Windows binary of unknown provenance | `scripts/api-lint/api-lint.exe` | No build provenance |
| Stale compiled binary committed from initial commit | `bin/metaldocs-api.exe` | Commit `912879cba`; `.gitignore` covers root `.exe` but NOT `bin/` |

Additional lower-severity item confirmed:
- `scripts/start-spec1-api.ps1:2` hardcodes `C:\Users\leandro.theodoro.MN-NTB-LEANDROT\...` — a different machine's absolute path.

The credential-bearing file was last touched commit `c4a7d9a93` (2026-04). Git history permanently retains the secret unless the history is rewritten.

### Standard

**CWE-798: Use of Hard-coded Credentials** — MITRE CWE taxonomy entry. Definition: "The software contains hard-coded credentials, such as a password or cryptographic key, that it uses for its own inbound authentication or for authentication with external components."

**OWASP ASVS 4.0 §2.10** (Credential storage and service account requirements): V2.10.1 — "Verify that integration secrets do not rely on unchanging credentials such as passwords, API keys, or shared accounts with privileged access." V2.10.4 — "Verify that secrets, API keys, and passwords are not included in source code, or online source code repositories."

**OWASP Top 10 2021 A07: Security Misconfiguration** and **A02: Cryptographic Failures** — committing credentials to a version-controlled repository is a canonical instance of both.

**ISO 27001:2022 A.8.3 Information access restriction**: access credentials must be managed and revoked; committing them to a shared repository violates minimum access control principles.

**Git history note:** `git log --all` retains committed secrets indefinitely. BFG Repo-Cleaner or `git filter-repo` are the standard remediation tools (GitHub documented guidance: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository). Rotation alone without history scrub means the secret is still recoverable by any cloner.

### Verdict: REFACTOR — P0-prerequisite

**Rationale:** This is not a "fix before production" finding — it is a "fix before any further security analysis is meaningful" finding. A credential in VCS must be treated as compromised from the moment of the first clone. The MinIO credentials (`minioadmin`/`minioadmin`) are default values that are also likely present in the compose stack; the Postgres password `***REDACTED***` is non-default and specific, suggesting it may be a real credential used in a shared environment.

**Over-engineering check:** The fix is the minimum possible action. No abstraction or framework is required. This is a delete + rotate + gitignore + history-scrub operation.

**Smallest correct fix (ordered by urgency):**

1. **Rotate the Postgres password** (`***REDACTED***`) immediately if it matches any non-dev environment. Assess which environments use this password. Document the rotation outcome.
2. **Delete `cmd/seed-test-document/main.go`** — confirmed dead binary (no canonical script, no CI reference, `D-06` topology flag). The seed function is covered by `apps/api/cmd/metaldocs-e2e-seed/` which reads credentials from env. Delete the whole `cmd/` directory; it has no other inhabitants.
3. **Rewrite git history** using `git filter-repo --invert-paths --path cmd/seed-test-document/main.go` (or BFG) to remove the secret from all commits, then force-push. Add `cmd/` to `.gitignore` or confirm it is gone.
4. **Remove `bin/metaldocs-api.exe`** with `git rm bin/metaldocs-api.exe` and add `bin/*.exe` to `.gitignore`. The file is a stale build artifact from the initial commit.
5. **Remove or replace `scripts/api-lint/api-lint.exe`**: a pre-built Windows binary of unverifiable provenance violates supply-chain hygiene (OWASP Top 10 2021 A08: Software and Data Integrity Failures — "software and data integrity failures relate to code and infrastructure that does not protect against integrity violations"). Replace with `go run ./scripts/api-lint/` or a `go build`-produced artifact tracked by a reproducible build. At minimum, add a SHA-256 manifest and build instructions.
6. **Fix `scripts/start-spec1-api.ps1`** — delete (confirmed dead by repo-topology flag `DEAD-SCRIPT-1`).
7. **Audit `.gitignore`** for full coverage of all binary and compiled artifact paths produced by the build.

**Effort:** S (hours, not days — delete + rotate + history rewrite + gitignore patch).

**Blast radius:** `cmd/` directory only. No module code affected. History rewrite requires force-push to all branches and notifying all active cloners.

**ADR needed:** No. Rotation and deletion do not require a design decision. If the team decides to retain a dev-seed tool, document it in a new `apps/` binary reading from env — that is a separate implementation task.

---

## F-12 — Tenant Isolation Gaps

### Current state

The register identifies five distinct gap patterns. Code verification confirms all five:

**Gap 1 — `auth_identities` has no `tenant_id` column.**
`archive/migrations/0021_init_auth_identities_and_sessions.sql` confirms: `auth_identities` has `user_id TEXT PRIMARY KEY REFERENCES metaldocs.iam_users(user_id)`. No `tenant_id` column. The table is intentionally tenant-global: identity (username/password) is global, while role assignments are tenant-scoped in `iam_user_roles`. This is a deliberate design, commented at `internal/modules/security/infrastructure/postgres/repository.go:83-86`: "JOIN iam_users.tenant_id binds the lockout to the caller's tenant — auth_identities is a global-PK table (no tenant_id column), so the JOIN is how we get tenant scoping."

**Gap 2 — `controlled_documents` and `cd_sequence_counters` have no GUC + RLS backstop.**
`internal/modules/controlleddocuments/infrastructure/repository.go` shows all queries filter by `tenant_id` parameter (verified: lines 36, 57, 93, etc.), but there is no `SET LOCAL metaldocs.tenant_id` before queries, no RLS policy, and no trigger tripwire on the `controlled_documents` table for tenant isolation. However, the tripwire trigger at `db/migrations/0231` covers `controlled_documents` for authz (capability assertion), NOT for tenant isolation. REQ-TEN-1 requires both: every query must filter by tenant_id AND there must be a DB-layer backstop.

**Gap 3 — Security module cross-tenant isolation via JOIN only.**
`internal/modules/security/infrastructure/postgres/repository.go:83-100` confirmed: uses `JOIN metaldocs.iam_users u ON u.user_id = i.user_id WHERE u.tenant_id = $1::uuid`. The JOIN is correct in isolation but fragile — if a future query path omits the JOIN, cross-tenant data becomes accessible.

**Gap 4 — `audit_export_jobs.tenant_id` type mismatch: UUID in Postgres, `string` in Go.**
`internal/modules/audit/domain/port.go:146-158` confirms `TenantID string` in the Go struct. This works today because pgx/lib-pq will cast a valid UUID string to `uuid` in the SQL parameterized query. It is not a runtime bug for valid UUIDs, but it means type safety is not enforced at compile time and a malformed tenant_id would produce a Postgres error, not a Go compile error.

**Gap 5 — `iam_users` INSERT has no DB tripwire for tenant isolation.**
`internal/modules/iam/infrastructure/postgres/role_admin_repository.go:51-58` shows the `iam_users` upsert is inside an `authz.Require(CapUserManage)` transaction, which provides capability assertion via the tripwire. But the tripwire at migration 0231 does NOT cover `iam_users` itself — it covers `iam_user_roles` and `user_process_areas`. The `iam_users` table itself has no trigger for tenant isolation; it relies on the `tenant_id` column being set correctly by the application.

**Gap 6 — `MembershipGovernanceLogger` wired nil in production.**
`apps/api/cmd/metaldocs-api/main.go:325` (noted T-007 open). Grant/revoke produce no governance log. Not a tenant isolation gap per se, but an asymmetric audit trail in a QMS system.

### Standard

**REQ-TEN-1** (project requirement): "Pooled model: every tenant-owned table carries `tenant_id`; every query path filters by it; tenant resolved once at the edge and carried in context."

**OWASP ASVS 4.0 §4.1 General Access Control Design** — V4.1.3: "Verify that the application enforces the principle of least privilege; users should only be able to access functions, data files, URLs, controllers, services, and other resources, for which they possess specific authorization." V4.1.5: "Verify that access control failures are logged and administrators are alerted when appropriate."

**OWASP Top 10 2021 A01: Broken Access Control** — "Access control enforces policy such that users cannot act outside of their intended permissions... Failures typically lead to unauthorized information disclosure, modification, or destruction of all data."

**PostgreSQL Row Security (RLS)** — official documentation at https://www.postgresql.org/docs/current/ddl-rowsecurity.html. RLS provides a DB-layer backstop that survives application bugs: `CREATE POLICY tenant_isolation ON table USING (tenant_id = current_setting('app.tenant_id')::uuid)`. For a pooled-connection SaaS, the pattern is: set `app.tenant_id` as a transaction-local GUC (`SET LOCAL`), then RLS policies enforce isolation even if the application forgets the WHERE clause. This is the standard defense-in-depth layer for multi-tenant Postgres applications.

**ISO 27001:2022 A.8.3** — access to information must be restricted based on access control policy. Multi-tenant isolation is a concrete instantiation of this control.

### Verdict: REFACTOR — P1

**Rationale:** The application-layer predicate filtering is functionally correct on all confirmed hot paths. The risk is not a current data leak — it is that the absence of a DB-layer backstop means a single missing WHERE clause (a developer error, a refactor regression, or an N+1 optimization that rewrites a query) produces a cross-tenant data exposure with no backstop. For a regulated QMS product, OWASP ASVS Level 2 requires defense-in-depth; predicate-only is defense-in-one.

The `auth_identities` gap (Gap 1) is intentional and KEEP: identity is deliberately tenant-global; the security module's JOIN-based scoping is the correct pattern for this table. Calling this a "gap" overstates it — the design is correct and documented in comments.

The `audit_export_jobs` type mismatch (Gap 4) is a SIMPLIFY: change `TenantID string` to `TenantID uuid.UUID` in the Go domain type and update the insert to use the typed value. This is a 5-line change with no behavioral effect for valid inputs.

**Over-engineering check:** Full RLS on every table is expensive to implement and maintain. The ROI is highest on the tables with the most direct user-facing query paths and the most sensitive data. A pragmatic sequencing is: (1) fix the type mismatch, (2) add RLS to the highest-risk tables (`documents`, `controlled_documents`, `audit_events`) using the existing GUC infrastructure, (3) defer lower-risk tables to a migration program. Do not attempt blanket RLS in one PR — it will break tests and require substantial query review.

**Smallest correct fix (sequenced):**

1. **Gap 4 fix (S effort):** Change `TenantID string` to `TenantID uuid.UUID` in `audit/domain/port.go:ExportJob` and update the insert/scan sites. No migration needed.
2. **T-008 / Gap 1 acknowledgement (no fix needed):** Add a comment + wiki note confirming `auth_identities` is intentionally tenant-global; the security module JOIN is the correct isolation pattern. Close T-008 as "by design."
3. **Gap 2 and 3 (M effort, needs migration):** Add RLS policies to `controlled_documents` and `audit_events`. This requires: (a) a migration that enables RLS and creates policies using `current_setting('metaldocs.tenant_id', true)`, (b) updating repository methods to `SET LOCAL metaldocs.tenant_id = $tenantID` before queries (or via a shared query context helper), (c) test coverage confirming cross-tenant queries are rejected. This is a standalone migration PR, not a module refactor.
4. **Gap 5 (DEFER):** The `iam_users` upsert is already inside an `authz.Require` transaction and the tenant_id is always passed from the session context. The risk is low and the fix (adding a trigger to `iam_users`) conflicts with the authz tripwire's table list management. Defer until the authz tripwire program (RF-6) runs.
5. **Gap 6 / T-007 (DEFER):** `MembershipGovernanceLogger` nil wiring. This is an audit gap, not an isolation gap. Address in the audit-atomicity program (F-07 family).

**Effort:** M (the fix is a migration + 2-3 repository files; the ADR write-up is needed to close T-008 as by-design).

**Blast radius:** `module` — contained to audit domain type, controlled-documents repository, and a DB migration. No cross-module changes needed.

**ADR needed:** Yes, one ADR: "auth_identities tenant-global by design; RLS adoption plan for pooled tables." This closes T-008 and documents the RLS rollout sequence so future reviewers understand why some tables have RLS and others do not.

---

## F-11 — Capability String Literals and Undeclared Capability Constants

### Current state

Three distinct violations confirmed by code reading:

**Violation 1 — `"template.admin"` literal at `routes_lifecycle.go:192`, no registry entry.**

`internal/modules/templates/delivery/http/routes_lifecycle.go:192` passes `"template.admin"` as a raw string to `h.authz(r, tenantID, "*", "template.admin")`.

The `AuthzFunc` type is `func(r *http.Request, tenantID, area string, action string) error`. In production, this is wired to `CapabilityService.CanDo`.

`capability_service.go:41-43` confirms: `CanDo` calls `iamdomain.IsValidCapability(iamdomain.Capability(capability))` first. `"template.admin"` is NOT in `validCapabilities` (confirmed: `model.go` lists 8 `CapTemplate*` constants — `View`, `Create`, `Edit`, `Submit`, `Review`, `Approve`, `Publish`, `Archive`; no `CapTemplateAdmin`). Therefore `IsValidCapability` returns false, and `CanDo` returns `ErrCapabilityDenied` immediately.

**Runtime effect confirmed:** `PUT /approval-config` (the `upsertApprovalConfig` handler) permanently returns 403 for all non-`system_admin` users, regardless of their actual roles. The route is effectively locked. This is a correctness defect, not a security weakening — but it means the template approval config is only editable by system admins, which is likely not the intended permission boundary.

**Violation 2 — Route admin event type string literals in `route_admin_service.go`.**

`internal/modules/documents/approval/application/route_admin_service.go:223,375,520` use string literals `"route.config.created"`, `"route.config.updated"`, `"route.config.deactivated"` assigned to `EventType` (a typed `string`-based Go type defined in `events.go:10`). The typed constants in `events.go` cover approval-instance and document-published events but NOT the three route admin event types. These are unregistered strings: a typo would produce a governance log with a malformed event type, undetectable at compile time.

**Violation 3 — Tier-1 / tier-2 capability name mismatch on publish route.**

`internal/modules/templates/delivery/http/routes_generated.go:203` passes `"template.approve"` (via `h.authz`) at tier-1. The tier-2 call on the same route uses `CapTemplatePublish` (`"template.publish"`). This means: the HTTP edge checks "can user approve templates?" while the service layer enforces "can user publish templates?". The two are different capabilities with potentially different grant sets. For routes that have this mismatch, a user with `template.approve` but NOT `template.publish` would be admitted by the middleware and then rejected inside the transaction — silent partial denial, no external indicator. Conversely, a user with `template.publish` but NOT `template.approve` is rejected at the edge even though the tier-2 enforcement would allow them. This is an authz coherence defect.

### Standard

**REQ-AUTHZ-2** (project requirement): "Every capability referenced anywhere is a typed registry const; raw strings and inline `Capability('...')` are CI-rejected (`no-inline-capability`, raw-string guard). (MUST)"

**OWASP ASVS 4.0 §4.1** — V4.1.1: "Verify that the application enforces access control rules on a trusted service layer, especially if client-side access control is present and could be bypassed." A locked route due to an unregistered capability string is an access control misconfiguration with the same ASVS relevance as an open one.

**OWASP Top 10 2021 A01: Broken Access Control** — "Bypassing access control checks by modifying the URL, internal application state, or the HTML page" — in this case the inversion: a correct user is denied because the capability name is wrong.

**CWE-284: Improper Access Control** — "The software does not restrict or incorrectly restricts access to a resource from an unauthorized actor." An unregistered capability permanently denying a valid user is an instance of improper access control.

**ADR 0022 (project):** The project's own ADR 0022 explicitly bans raw capability strings from delivery and application packages. These three violations are direct regressions against that ADR.

### Verdict: REFACTOR — P1

**Rationale:** Violation 1 is a concrete correctness defect — a production route is permanently locked to system_admin only. Violation 3 is an authz coherence defect that may silently deny or silently grant (depending on grant set overlap). Both require targeted fixes. Violation 2 (event type strings) is a lower-risk hygiene issue that should be fixed in the same pass.

**Over-engineering check:** The fix does NOT require a typed `EventType` registry for governance log events (that would be over-engineering the event model). The fix is: (a) add a `CapTemplateAdmin` constant OR replace `"template.admin"` with the intended capability (most likely `CapTemplateManage` or collapse into an existing cap), (b) add the three route-admin event types as `EventType` constants in `events.go`, (c) fix the tier-1/tier-2 mismatch on the publish route. None of these require cross-module changes.

**Smallest correct fix (ordered):**

1. **Violation 1 (S effort — 2 files):** Decide the intended capability for `upsertApprovalConfig`. Options: (a) `CapTemplateArchive` is the closest write-grade template cap; (b) introduce `CapTemplateManage` if the intent is a super-admin cap; (c) check if the route should use an existing cap. Whichever cap is chosen, replace the string literal with the typed constant and update `permissions.go` to declare the tier-1 rule for that route. Add a `TestPermissions` lock test covering this route. **Do not introduce a new `CapTemplateAdmin` constant without an ADR** — ADR 0016 established that new cap additions require documented rationale (precedent for View caps; same applies here).

2. **Violation 2 (S effort — 1 file):** Add `EventTypeRouteConfigCreated`, `EventTypeRouteConfigUpdated`, `EventTypeRouteConfigDeactivated` to `events.go` and replace the three string literals in `route_admin_service.go`.

3. **Violation 3 (M effort — cross-layer verification):** Audit the tier-1/tier-2 pairing on all template lifecycle routes. The canonical check: for each `h.authz(r, tenantID, area, capString)` call in the templates delivery layer, find the corresponding tier-2 `authz.Require` call in the service or repository layer and confirm they name the same capability. Fix any mismatch. Document the expected pairings in a lock test or in `wiki/concepts/authz-tiers.md §Modules with tier-2 coverage`.

4. **CI guard (no additional work if API lint rule is already in place):** `api-lint`'s `no-inline-capability` rule should catch Violations 1 and 2 if it scans delivery and application packages. Verify the rule's scope covers `internal/modules/templates/delivery/http/` and `internal/modules/documents/approval/application/`. If the current scope is narrower, expand it. The capability catalog SHA-256 gate (`ops/CAPABILITY_CATALOG.sha256`) is currently a no-op (placeholder hash, missing seed file — confirmed in F-03); fixing that gate is part of the REQ-AUTHZ-5 closure, not this finding.

**Effort:** M (Violation 1 requires a design call on the intended cap; Violations 2-3 are mechanical).

**Blast radius:** `module` — changes are contained to two delivery packages and the events.go file. The ADR for Violation 1 may touch `permissions.go` and one migration if a new cap is introduced.

**ADR needed:** Conditional — if Violation 1 is resolved by introducing `CapTemplateManage` (a new cap), yes, a micro-ADR is required per ADR 0016 precedent. If it is resolved by reusing an existing cap (e.g. `CapTemplateArchive`), no ADR is needed — just the code change and a lock test.

---

## Cross-finding observations

1. **F-18 and the dead binary convention (D-06):** Deleting `cmd/seed-test-document/` eliminates the only inhabitant of the root `cmd/` convention. This closes D-06 as a side effect.

2. **F-12 and the idempotency table:** `internal/platform/idempotency/postgres_store.go:54,87` (noted in F-09) — the `idempotency_keys` table is missing `tenant_id`. This is a related gap outside this finding's scope but should be scheduled alongside the F-12 RLS work.

3. **F-11 and the capability catalog gate:** The CAPABILITY_CATALOG.sha256 gate being a no-op (F-03/BROKEN-GATE) means Violation 1 was not caught by CI. Fixing F-11 and simultaneously fixing the capability catalog gate are complementary — the gate provides the ongoing CI enforcement.

---

## Sources (external citations)

- OWASP ASVS 4.0: https://owasp.org/www-project-application-security-verification-standard/
- OWASP Top 10 2021: https://owasp.org/Top10/
- CWE-798 Use of Hard-coded Credentials: https://cwe.mitre.org/data/definitions/798.html
- CWE-284 Improper Access Control: https://cwe.mitre.org/data/definitions/284.html
- PostgreSQL Row Security Policies: https://www.postgresql.org/docs/current/ddl-rowsecurity.html
- GitHub: Removing sensitive data from a repository: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository
- ISO 27001:2022 A.8.3 / A.5.15 (access restriction and control): https://www.iso.org/standard/82875.html
