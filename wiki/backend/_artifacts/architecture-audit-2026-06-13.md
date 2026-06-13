# Independent Architecture Audit — 2026-06-13

> **Method:** 10-dimension multi-agent audit (39 agents, ~2.27M tokens) reading the **actual code** (not wiki claims), grading each subsystem A–F against industry standards. Every Critical/Major finding was sent to an independent skeptic that confirmed, downgraded, or **refuted** it against the code. This is a *fresh* read independent of the Wave 0–Z program: it surfaced ~23 verified defects that were never in the legacy register (mostly structural/architectural).
> **Branch:** qa/iam-area-membership. **Trigger:** operator asked whether the backend architecture (as designed AND implemented) is genuinely industry-grade before moving to screens.
> **Headline:** Solid **B / "good senior Go."** The craft is real; debt is concentrated in three *structure* dimensions (module boundaries, contract integration, composition root). NOT a rewrite — a bounded, fixable list.

## Scorecard

| Subsystem | Grade |
|---|---|
| Authz / capability model | B |
| Security / tenant isolation | B |
| Sessions / auth lifecycle | B |
| Middleware / HTTP kernel | B |
| Persistence / transactions | B |
| Code quality / Go idioms | B |
| Legacy / dead-code | B |
| **Module boundaries / DDD** | **C** |
| **Contract / API layer** | **C** |
| **Composition / observability** | **C** |

**The capability model — the operator's #1 fear ("rules buried in roles") — is the STRONGEST area.** Capabilities are a typed registry, role→cap mapping is DB-table-driven (`metaldocs.role_capabilities`), two-tier PEP/PDP applied at ~80 sites, fail-closed, system-admin bypass audited in-tx. Only crack: one role-name string check (downgraded to Minor, restrictive-only, no escalation).

## MUST-FIX before screens (small, high-value)

| # | Sev | Defect | File |
|---|-----|--------|------|
| A1 | 🔴 CRITICAL | `statusWriter` lacks `Unwrap()`/Hijacker → every WebSocket upgrade returns **501 in prod**; `/iam/presence/stream` dead. Tests pass because they bypass the obs middleware. **One-line fix** (`func (w *statusWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }`). | internal/platform/observability/http.go:293-321 |
| A2 | 🔒 Major | LIKE wildcard injection — `opts.Q` wrapped `%…%` unescaped; `sqlescape.LikeEscape()` exists, unused. Info-inference + CPU exhaustion. CWE-943. | internal/modules/documents/repository/repository.go:463-465 |
| A3 | 🔒 Major | Self-service password change does NOT revoke other sessions (stolen cookie survives). CWE-613. `AdminResetPassword` already does it right; copy the one line. | internal/modules/auth/application/service.go:395-434 |
| A4 | 🔒 Major | Sliding idle timeout defaults 0 (disabled) + absent from every deploy artifact. 12h absolute TTL only. Matters for ISO 27001 path. | internal/platform/authn/config.go:70-76 |
| A5 | 📑 Major | `PATCH /iam/users/{id}` returns `{user_id,updated,changes}` — spec declares `ManagedUserCore`. **Breaks FE codegen types.** | internal/modules/iam/delivery/http/people_handler.go:228-232 |
| A6 | 📑 Major | `GET /documents` + `GET /documents/{id}` return domain structs / `map[string]any`, not generated response types (omitempty + type drift). **Breaks FE codegen types.** | internal/modules/documents/delivery/http/handler.go:215-222,339-399 |

A5/A6 bite the next screens directly if they touch users-admin or documents.

## Architecture debt (real; does NOT block screens) — candidate "Wave H"

**Module boundaries (C):**
- `documents` has **3 parallel delivery subtrees** (`delivery/http`, `http`, `approval/http`) — every other module has one. — internal/modules/documents/
- `controlleddocuments/infrastructure` reimplements taxonomy readers (`TaxonomyProfileReader`/`AreaReader`) — **data-ownership break + stale schema column + skips authz GUC/cap check.** — controlleddocuments/infrastructure/repository.go:720-799
- `*sql.DB` passed into application-layer service methods (approval, CD, templates) — hexagonal violation; no `BeginTx` on the `db` port. — documents/approval/application/decision_service.go:152
- `setAuthzGUC` copy-pasted ×4; canonical `authz.SeedTxIdentity` exists with an empty-string guard the copies skip. — documents/approval/application/authz_guc.go:9
- delivery imports infrastructure (approval `http` → approval `infrastructure`). — documents/approval/http/handler.go:13

**Composition / observability (C):**
- `main.go` 1099 lines, `main()` ~593, **13 adapter types inline**; `apps/api/internal/wiring/` underused. — apps/api/cmd/metaldocs-api/main.go:89-1099
- 9 direct `os.Getenv` outside config layer. — main.go:141,171,204,205,420,424,644,694,945
- No `slog.SetDefault` → 3 inconsistent log formats; startup/shutdown/scheduler logs unstructured. — main.go
- Worker has no graceful drain; jobs `log.Fatalf` skips cleanup; OTel `service.name` hardcoded `metaldocs-api` (worker/jobs get none). — minor

**Persistence (B):**
- Post-commit `audit.Record` in 5 IAM/auth delivery handlers → torn-write window the `postcommitaudit` linter misses (commit is in the service layer). — iam/delivery/http/sessions_handler.go:184-194 +4
- nil-tx autocommit fallback in `PostgresSequenceAllocator.NextAndIncrement` (violates `db.Tx` contract; `nodualmode` linter doesn't cover `infrastructure/`). Unreachable in prod today, latent trap. — controlleddocuments/infrastructure/repository.go:656-686

**Contract (C):**
- Raw error-code string literals bypass the typed catalog in auth/iam/audit/security/search; guard covers only 5 of ~11 packages; 2 codes off-catalog (`CURSOR_EXPIRED`, `NOT_IMPLEMENTED`). — multiple delivery handlers
- `GeneratedServerAdapter` discards typed params, forwards to legacy mux (double-parse, drift). Tracked debt per ADR 0012. — documents/delivery/http/generated_adapter.go

**Code quality (B):**
- `RecordSignoff` 407-line god function with raw SQL in the application layer. — documents/approval/application/decision_service.go:152-558
- documents `Handler.Service` interface = 28 methods (4 unused) — fat interface. — documents/delivery/http/handler.go:30-59
- `PeopleService.ListFiltered` loads all users + N+1 membership query + silent error swallow; in-process filter/pagination. — iam/application/people_service.go:511-581

**Legacy / dead-code (B):**
- `SnapshotFromTemplate` dead (zero prod callers, no backfill binary exists; deprecation rationale unsupported). — documents/application/snapshot_service.go:45-68
- Legacy non-atomic `CreateDocument`/`DuplicateDocument` chain live + interface bloat (TODO-keyed). — documents/application/service.go:226-229

## Skeptic dispositions (do NOT re-raise these)

- **REFUTED — oapi-codegen `NewStrictHandler` text/plain errors:** dead generated code; `NewStrictHandler` has zero callers. All live error paths use `problem.Write`/`WriteError` (problem+json). RFC 9457 contract intact.
- **REFUTED — `tenantIDFromContext` DevTenantID fallback (CD handler):** only feeds idempotency-key namespace, NOT data scoping. All data paths use `tenant.FromContext` → `ErrTenantMissing` (fail-closed). No cross-tenant leak.
- **DOWNGRADED → Minor — role-name check `approval_config.go`:** restrictive-only (no escalation; `qms_admin` still fails `authz.Require` for `template.edit`); lifecycle.go `containsRole` cases are legit document-workflow SoD bindings, not PDP bypasses.
- **DOWNGRADED → Minor — CD/taxonomy infra import `iam/authz`:** intentional per ADR 0007 (tier-2 enforcement at the tx/repo layer so area is DB-derived & un-spoofable). Cross-module `authz` is a shared platform primitive, not IAM-internal infra.
- **DOWNGRADED → Minor — `GeneratedServerAdapter` shim:** intentional, tracked (ADR 0012 + contract-first-followups backlog), functionally correct.
- **DOWNGRADED → Minor — `authn.Enabled()/CacheTTL()` live `os.Getenv`:** startup-only wiring, captured to typed values; not a per-request re-read.

## On the Wave Z "all-green"

No contradiction. Wave Z went green against its **own register + REQ set**. This audit was an **independent** read and found *new*, mostly *structural* defects the register never catalogued. The program fixed what it tracked; it didn't track these.

## Remediation dispositions (Wave H — 2026-06-13)

Per-finding closure. Tracker rows mirror these in [`../roadmap.md`](../roadmap.md) §Wave H. `RESOLVED` = bounded commit; `DEFERRED` = written trigger at a hard-stop boundary.

| # | Disposition | Commit | Evidence |
|---|-------------|--------|----------|
| A1 | ✅ RESOLVED | (this commit) | `Unwrap()` added to `statusWriter`; coder/websocket hijacker follows the Unwrap chain. Runtime proof at close. |
| A2 | ✅ RESOLVED | (this commit) | `sqlescape.LikeEscape()` + `ESCAPE '\'` clause; regression test added. |
| A3 | ✅ RESOLVED | (this commit) | `RevokeSessionsByUserID` on password change; regression test added. Revokes ALL sessions (matches AdminResetPassword) — FE re-auths after change. |
| A4 | ☐ open | | |
| A5 | ☐ open | | |
| A6 | ☐ open | | |
| H-1 boundaries | ☐ open | | |
| H-2 composition/obs | ☐ open | | |
| H-3 persistence | ☐ open | | |
| H-4 contract | ☐ open | | |
| H-5 code quality | ☐ open | | |
| H-6 dead-code | ☐ open | | |
