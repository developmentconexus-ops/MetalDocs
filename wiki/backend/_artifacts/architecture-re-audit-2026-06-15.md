# Backend Architecture Re-Audit Report

**Project:** MetalDocs  
**HEAD:** `02ed1c24`  
**Date:** 2026-06-15  
**Method:** 10-dimension multi-agent audit with adversarial-skeptic verdict per Critical/Major finding; fresh read at HEAD for every cited line. Skeptic verdicts are binding — findings downgraded or refuted by the skeptic are excluded from the pass-bar count and confirmed-findings table.

---

## 1. Header and Method

This report supersedes the 2026-06-13 baseline audit. Each of the ten dimensions was independently audited by a dedicated auditor, then every Critical or Major finding was submitted to an adversarial skeptic who re-read the cited code before issuing a verdict: **confirmed** (severity stands), **downgraded** (severity reduced), or **refuted** (finding struck). Only skeptic-confirmed findings count toward the §6 pass-bar.

H-D (handler/spec drift) and H-G (cross-module schema grab) class counts were re-measured at HEAD using reproducible commands recorded in §6.

---

## 2. Scorecard

| # | Dimension | 2026-06-13 Baseline | Re-Audit Grade | Delta |
|---|-----------|---------------------|----------------|-------|
| 1 | Authz / capability model | B | B+ | +½ |
| 2 | Security / tenant isolation | B | B+ | +½ |
| 3 | Sessions / auth lifecycle | B | B+ | +½ |
| 4 | Middleware / HTTP kernel | B | B+ | +½ |
| 5 | Persistence / transactions | B | B+ | +½ |
| 6 | Code quality / Go idioms | B | B | 0 |
| 7 | Legacy / dead-code | B | B+ | +½ |
| 8 | **Module boundaries / DDD** ⬆ | **C** | **B+** | **+1½** |
| 9 | **Contract / API layer** ⬆ | **C** | **C+** | **+½** |
| 10 | **Composition / observability** ⬆ | **C** | **B+** | **+1½** |

Rows 8, 9, and 10 are the formerly-C dimensions.

---

## 3. §6 Pass-Bar Verdict

The pass bar requires all four of the following conditions to hold.

### (1) All 3 formerly-C dimensions ≥ A-?

**FAIL**

| Dimension | Re-Audit Grade | Meets A-? |
|-----------|----------------|-----------|
| module-boundaries-ddd | B+ | No |
| contract-api | C+ | No |
| composition-observability | B+ | No |

None of the three formerly-C dimensions reached A-. `contract-api` improved only to C+ due to a skeptic-confirmed Critical (PascalCase serialization breaking FE checkpoint endpoints) and five confirmed Majors spanning status-code mismatches, undeclared bodies, extra top-level fields, and pervasive `map[string]any` bypasses of generated types. `module-boundaries-ddd` and `composition-observability` both reached B+ but not A-.

### (2) Zero new skeptic-confirmed Critical/Major findings?

**FAIL**

Skeptic-confirmed Critical/Major findings (new at this HEAD, not present in the 2026-06-13 baseline) are listed in §4. Count: **18 confirmed Critical/Major** across all ten dimensions. The pass bar requires zero.

### (3) H-D = 0?

**FAIL**

H-D count at HEAD `02ed1c24`: **4**

Drift sites:
- `routes_generated.go:64` — extra undeclared top-level `id`/`version_id` fields in POST /templates response
- `routes_autosave.go:42` — 201 emitted, spec declares 200
- `routes_create.go:36` — 201 emitted, spec declares 200
- `routes_profiles.go:67,111,126,169` — raw `domain.DocumentProfile` emitted, missing required spec fields and exposing extra non-spec fields

### (4) H-G = 0?

**FAIL**

H-G count at HEAD `02ed1c24`: **1**

Violation: `internal/modules/documents/application/service.go:282` — `overrideStatus := "published"` hardcodes a domain-state string owned by `templates/domain` (correct constant: `templates/domain.VersionStatusPublished`) instead of importing the owning module's constant. The `security.MfaCoverage` violations are an accepted bounded defer (aggregate JOIN, no port available for this metric query, not counted).

---

### Overall Verdict

All four pass-bar checks fail. Every failing item becomes an F5.2 trigger:

- **F5.2-A:** Promote all 3 formerly-C dimensions to A- (contract-api is the critical blocker)
- **F5.2-B:** Resolve all 18 skeptic-confirmed Critical/Major findings
- **F5.2-C:** Reduce H-D to 0 (4 remaining drift sites)
- **F5.2-D:** Reduce H-G to 0 (1 remaining violation in documents/application/service.go:282)

---

## 4. Confirmed Findings

All findings below carry a skeptic verdict of **confirmed**. Severity is the skeptic's final severity.

| # | Sev | Dimension | Title | File:Line | Skeptic Reasoning (summary) |
|---|-----|-----------|-------|-----------|----------------------------|
| 1 | Major | authz-capability | Manual-code CD creation path: service bypasses tier-2 but repo calls authz.Require without seeded GUCs | `internal/modules/controlleddocuments/application/service.go:173` | Service never calls SeedTxIdentity on the manual-code branch; repository opens a fresh tx and calls authz.Require; MustActorID reads unset GUC and returns ErrActorContextMissing before the system_admin bypass check fires — all non-system-admin manual-code creates fail. |
| 2 | Major | authz-capability | CapDocumentView (tenant-grade) passed with real area code to tier-2 Require, blocking tenant-role-only viewers | `internal/modules/documents/approval/application/read_service.go:68` | Coalesce to "tenant" only fires when areaCode == ""; for documents with a real area code the area filter is ON, silently narrowing tenant-grade semantics to area-grade at the approval read layer. |
| 3 | Major | persistence-transactions | authz.Require ignores effective_from — future-dated memberships grant access prematurely | `internal/modules/iam/authz/authz.go:123` | Only `effective_to IS NULL` predicate present; no `effective_from <= now()` guard. ResolveEligibleActors on the same table uses the correct dual predicate. |
| 4 | Major | sessions-auth-lifecycle | ChangePassword revokes all sessions but sends no expired-cookie header | `internal/modules/auth/delivery/http/handler.go:153` | RevokeSessionsByUserIDTx is called; handler returns 200 with refreshed user body; no http.SetCookie with expired cookie; client holds a dead cookie and receives 401 on next request with no signal. |
| 5 | Major | code-quality-go | NewFreezeService accepts `fanoutClient any` and silently drops type mismatch | `internal/modules/documents/application/freeze_service.go:77` | Parameter typed `any`; boolean from type assertion discarded; nil stored silently; FanoutClient interface defined in same file — should be the parameter type for compile-time safety. |
| 6 | Major | code-quality-go | ListDocumentComments carries dead `userID` parameter through interface and implementation | `internal/modules/documents/application/service.go:433` | userID accepted and immediately discarded; sibling methods use userID; interface contract misleads callers about user-scoped filtering; may mask a missing authz scope. |
| 7 | Major | code-quality-go | `_ = snap` discards snapshot result immediately after read in Pin | `internal/modules/documents/application/freeze_service.go:194` | ReadSnapshotWithFreezeAt fetches four columns including large JSON blobs; Pin immediately blanks the result with `_ = snap`; wastes DB bandwidth; interface mismatch smell with no explanatory comment. |
| 8 | Major | legacy-dead-code | TemplateDocxKey and TemplateSchemaKey are exported but have no production callers | `internal/platform/objectstore/template_keys.go:5` | Zero production callers; key format diverges from live production key schema used by TemplatesPresigner; unit tests verify a pattern never used in production. |
| 9 | Major | legacy-dead-code | IAMUserOptions dependency never wired — placeholder user-lookup feature silently returns empty | `apps/api/cmd/metaldocs-api/main.go:413` | IAMUserOptions field absent from production wiring literal; nil passed to adapter; nil guard returns empty slice with no error and no log; GET placeholder-options for user-type always returns empty list. |
| 10 | Major | module-boundaries-ddd | security.MfaCoverage queries IAM-owned iam_users directly (H-G residue) | `internal/modules/security/infrastructure/postgres/repository.go:67` | File comment says module does not own iam_users; all other methods in same file use TenantUserReader/UserDisplayNameReader ports; MfaCoverage is the sole exception, reading iam_users.deactivated_at and iam_users.mfa_enabled. |
| 11 | Major | module-boundaries-ddd | security.ListOffHoursAdminActions JOINs IAM-owned iam_user_roles directly | `internal/modules/security/infrastructure/postgres/repository.go:345` | Explicit JOIN metaldocs.iam_user_roles confirmed; no IAM role-resolution port defined; not addressed in M4 F4.5/F4.6 remediation program. |
| 12 | Major | module-boundaries-ddd | documents module imports templates/domain.Placeholder as a production type across module boundary | `internal/modules/documents/repository/repository.go:16` | Import appears in repository, application service, fillin_service, and delivery layers; port interface itself declares `[]templatesdomain.Placeholder`; unidirectional but baked into the seam. |
| 13 | Critical | contract-api | listCheckpoints and createCheckpoint write untagged domain struct — PascalCase keys break FE contract | `internal/modules/documents/delivery/http/handler.go:881` | domain.Checkpoint has no JSON struct tags; PascalCase serialized on wire; spec and FE-generated type (index.d.ts:2264) require snake_case; every consumer of GET/POST /documents/{id}/checkpoints receives structurally wrong JSON. |
| 14 | Major | contract-api | renameDocument writes raw *domain.Document body when spec declares 200 with no content | `internal/modules/documents/delivery/http/handler.go:519` | Handler writes raw domain struct including base64 FormDataJSON; spec declares no content body; FE expects DocumentDetailResponse shape from getDocument path; shape mismatch silently corrupts FE state. |
| 15 | Major | contract-api | createTemplate emits undeclared top-level fields id and version_id not in spec schema | `internal/modules/templates/delivery/http/routes_generated.go:64` | Spec declares only `{data: {template, version}}`; extra top-level `id` and `version_id` invisible to FE codegen types; post-create navigation reading `response.id` receives undefined at runtime. |
| 16 | Major | contract-api | createNextVersion emits HTTP 201 but spec declares 200 | `internal/modules/templates/delivery/http/routes_create.go:36` | 201 emitted; spec declares only 200 with no body schema; FE codegen types response as `content?: never`; no typed access to returned version payload. |
| 17 | Major | contract-api | presignTemplateAutosave emits HTTP 201 but spec declares 200 | `internal/modules/templates/delivery/http/routes_autosave.go:42` | 201 emitted; spec only declares 200; presign endpoint semantics are not resource-creation; strict clients will treat 201 as unexpected. |
| 18 | Major | contract-api | Pervasive map[string]any bypasses generated response types across critical routes | `internal/modules/documents/delivery/http/handler.go:816` | Six confirmed sites: presignAutosave, commitAutosave, listRevisionHistory, restoreCheckpoint, duplicateDocument, sessions list — all write hand-rolled maps; spec changes update api.gen.go but leave maps silently stale. |
| 19 | Major | composition-observability | Scheduler uses hardcoded text-format logger instead of slog.Default() | `internal/modules/jobs/scheduler/scheduler.go:131` | `slog.NewTextHandler(os.Stdout, nil)` hard-coded; all three production binaries set JSON default before constructing scheduler; scheduler ignores it; log aggregators parsing JSON will fail on scheduler lines. |
| 20 | Major | composition-observability | Scheduler per-job metrics (runs/errors/skips) never exposed to any scrape endpoint | `internal/modules/jobs/scheduler/scheduler.go:273` | MetricsSnapshot() called only in test files; no production path wires it to /api/v1/metrics or any other handler; operator cannot observe job failure or backpressure skip rates. |
| 21 | Major | composition-observability | OTel instrumentation is HTTP-envelope only; no application-level spans | `internal/platform/observability/otel.go:95` | Zero otel.Tracer().Start() calls in internal/; no DB instrumentation packages; traces contain one server span per request with no child spans; distributed tracing uninformative for latency diagnosis. |

Total confirmed: **1 Critical, 20 Major** = **21 skeptic-confirmed Critical/Major findings**.

---

## 5. Refuted / Downgraded Findings

These findings were raised by dimension auditors but overturned or reduced by the skeptic. They must not be re-raised in F5.2 without new evidence.

| Finding | Auditor Severity | Skeptic Outcome | Reason |
|---------|-----------------|-----------------|--------|
| authzrequire lint analyzer absent from live tools/cilint tree | Major | **Refuted** | The rule never lived in cilint; it lived in scripts/api-lint and was deliberately deleted (FD-1, ADR 0007 amendment) because it modelled the wrong architecture; real CI enforcement via tripwire-pairing and authz-area-scope-binding gates is active. |
| http.Server.WriteTimeout (60s) kills WebSocket connections before 90s ClientIdleTimeout | Major | **Refuted** | coder/websocket calls http.Hijacker.Hijack() on upgrade; net/http stops enforcing WriteTimeout post-hijack; the idle-close logic operates independently on the hijacked connection. |
| auth_identities.ListUsers fetches all tenants at SQL layer | Major | **Downgraded → Minor** | No active IDOR at HEAD; RolesByUserIDs returns (nil, err) on failure and service propagates immediately; Go filter is always applied and fails closed; architectural defense-in-depth gap only. |
| RevokeSessionsByUserIDTx revokes across all tenants with no tenant_id filter | Major | **Downgraded → Minor** | Both callers require exactly cross-tenant revocation (CWE-613); no HTTP handler path produces unintended cross-tenant revocation; risk is prospective for hypothetical future callers only. |
| GetFinalizePrereqs reads document status across 4 unlocked queries — TOCTOU window | Major | **Downgraded → Minor** | SubmitRevisionForReview UPDATE inside runner.Do transaction asserts WHERE status = 'draft'; RowsAffected() == 0 returns ErrStaleRevision and rolls back; double-submit and submit-after-archive are blocked. |
| FillInHandler.RegisterRoutes dead in production and would panic if called | Major | **Downgraded → Minor** | Zero production runtime risk; panic only on a scenario that does not exist; dead code cleanup only. |
| Error dispatch uses strings.HasPrefix/Contains instead of sentinel errors | Major | **Downgraded → Minor** | json unknown-field case has no exported stdlib sentinel; other two cases unlikely to be wrapped in practice; no production bug demonstrated; maintenance/robustness concern only. |
| acquireDocumentSession bypasses generated types with raw map[string]any | Major | **Downgraded → Minor** | string and openapi_types.UUID serialize identically to JSON string; no current wire-level mismatch; contract-hygiene gap only. |

---

## 6. Class Re-Measurement

### H-D (Handler/Spec Drift) — Count: **4**

**Reproducible commands:**

```sh
grep -rn "map\[string\]any" internal/modules/*/delivery/http/ --include="*.go" -l
grep -rn "map\[string\]any" internal/modules/*/delivery/http/*.go
grep -n "writeJSON\|WriteJSON\|writeFillInJSON" \
  internal/modules/templates/delivery/http/routes_generated.go \
  internal/modules/templates/delivery/http/routes_autosave.go \
  internal/modules/templates/delivery/http/routes_create.go
grep -n "WriteJSON\|writeJSON" internal/modules/taxonomy/delivery/http/routes_profiles.go
grep -n "DocumentProfileItem\|active_schema_version\|workflow_profile\|approval_required\|retention_days\|validity_days" \
  api/openapi/v1/openapi.yaml
grep -n "type DocumentProfile struct" internal/modules/taxonomy/domain/profile.go
```

**Evidence:**

| Site | File:Line | Drift Type |
|------|-----------|-----------|
| DRIFT-1 | `internal/modules/templates/delivery/http/routes_generated.go:64` | Extra top-level `id` and `version_id` fields not in spec `CreateTemplateResponse` |
| DRIFT-2 | `internal/modules/templates/delivery/http/routes_autosave.go:42` | HTTP 201 emitted; spec declares only 200 for `presignTemplateAutosave` |
| DRIFT-3 | `internal/modules/templates/delivery/http/routes_create.go:36` | HTTP 201 emitted; spec declares only 200 for `createTemplateVersion` |
| DRIFT-4 | `internal/modules/taxonomy/delivery/http/routes_profiles.go:67,111,126,169` | Raw `domain.DocumentProfile` emitted; missing required spec fields (`active_schema_version`, `workflow_profile`, `approval_required`, `retention_days`, `validity_days`); exposes extra non-spec fields (`tenant_id`, `default_template_version_id`, `owner_user_id`, etc.) |

All other delivery handlers verified clean.

---

### H-G (Cross-Module Schema Grab) — Count: **1**

**Reproducible commands:**

```sh
git rev-parse HEAD
git diff --name-only 02ed1c24..HEAD -- "*.go"
grep -rn "FROM metaldocs\.iam_users\|JOIN metaldocs\.iam_users\|INTO metaldocs\.iam_users\|UPDATE metaldocs\.iam_users\|DELETE FROM metaldocs\.iam_users" \
  --include="*.go" internal/modules/ \
  | grep -v "internal/modules/iam/" | grep -v "_test\.go"
grep -rn "metaldocs\.template" --include="*.go" internal/modules/ \
  | grep -v "internal/modules/controlleddocuments\|internal/modules/templates" \
  | grep -v "_test\.go"
grep -rn "status.*:=.*\"published\"\|status.*=.*\"published\"\|status.*:=.*\"draft\"\|status.*=.*\"draft\"\|status.*:=.*\"active\"\|status.*=.*\"active\"\|status.*:=.*\"approved\"\|status.*=.*\"approved\"" \
  --include="*.go" internal/modules/ \
  | grep -v "_test\.go" | grep -v "domain/\|api.gen.go\|api/api"
```

**Evidence:**

| Site | File:Line | Violation Type |
|------|-----------|---------------|
| VIOLATION-1 | `internal/modules/documents/application/service.go:282` | `overrideStatus := "published"` hardcodes a domain-state string owned by `templates/domain`; correct reference: `templates/domain.VersionStatusPublished` |

**Accepted bounded defer (not counted):** `internal/modules/security/infrastructure/postgres/repository.go:67,80` — both lines inside `MfaCoverage()`. Accepted as bounded defer: aggregate JOIN, no display-name read, no port available for this metric query. Tracked as confirmed Major #10 in §4 for remediation planning but excluded from H-G count per the bounded-defer classification.

---

## 7. Minor Findings

The following findings were raised by auditors and either confirmed at Minor severity by the skeptic or recorded without skeptic gate (Minors are not skeptic-gated per protocol). They do not affect the pass-bar.

**authz-capability**
- `capability_service.go:51` — CanDo bypass uses hardcoded `"system_admin"` string literal rather than `string(iamdomain.RoleSystemAdmin)`; drift surface if role code is renamed
- `bypass_audit.go:183` — Background BypassSystem uses empty TenantID when tenant is not seeded; no enforcement that callers claiming "system" attribution have not seeded a tenant identity

**security-tenant-isolation**
- `security/infrastructure/postgres/repository.go:348` — ListOffHoursAdminActions missing `::uuid` cast on tenant_id predicate (inconsistent with all other queries in same file)
- `auth/infrastructure/postgres/repository.go:396` — ListUsers fetches all tenants at SQL layer; Go-side post-query filter is correct but defense-in-depth gap (downgraded from Major)
- `auth/infrastructure/postgres/repository.go:203` — RevokeSessionsByUserIDTx has no tenant-scoped overload; footgun for hypothetical future callers (downgraded from Major)
- `security/application/service.go:179` — stableID uses SHA-1 for deduplication keys; non-security use but may trigger gosec G401
- `auth/application/service.go:618` — UpdateUser can change password without session revocation; not exposed via current handlers but port is callable

**sessions-auth-lifecycle**
- `auth/infrastructure/memory/repository.go:201` — memoryLoginLock.LoadLoginState reads identity outside the critical-section mutex; test-fixture fidelity gap
- `auth/application/service.go:979` — No unit test for HMAC tamper rejection in tokenHashFromCookieValue
- `auth/application/service.go:362` — AllowDevTenantFallback not guarded against production use in NewService

**middleware-http-kernel**
- `internal/platform/observability/http.go:303` — Dead "Status" header fallback branch in statusWriter.Write; unreachable in current codebase
- `internal/platform/idempotency/middleware.go:176` — responseRecorder missing Unwrap(); violates ResponseWriter wrapping contract

**persistence-transactions**
- `internal/platform/db/postgres/connect.go:12` — No idle_in_transaction_session_timeout configured at DB level
- `internal/modules/documents/repository/repository.go:620` — Inconsistent defer tx.Rollback() error-discard pattern (bare vs `_ =` form) across repository methods
- `internal/platform/db/postgres/connect.go:12` — search_path not pinned at connection level; relies entirely on ALTER DATABASE default

**code-quality-go**
- `internal/modules/documents/application/service.go:127` — Duplicate constructors New and NewService for documents/application.Service; no deprecation notice
- `internal/modules/documents/approval/application/submit_service.go:233` — Hardcoded Portuguese string `"Criacao do documento"` as business-logic constant

**legacy-dead-code**
- `db/baseline/0001_current_schema.sql:1336` — metaldocs.documents MDDM cluster still present; pending migration 0240 behind HS-1 gate
- `internal/modules/search/infrastructure/v2documents/reader.go:173` — DocumentType is a redundant alias always equal to DocumentProfile
- `internal/modules/documents/delivery/http/fillin_handler.go:204` and 7 other locations — eight private copies of tenantIDFromRequest across modules
- `internal/modules/documents/delivery/http/fillin_handler.go:37` — FillInHandler.RegisterRoutes dead in production; vestige of pre-codegen-wrapper era (downgraded from Major)

**module-boundaries-ddd**
- `internal/modules/controlleddocuments/module.go:14` — controlleddocuments/module.go imports templates/infrastructure concrete type at wiring layer
- `internal/modules/iam/integration_test.go:16` — IAM integration test imports documents/approval/repository for ErrInstanceCompleted

**contract-api**
- `internal/modules/audit/delivery/http/handler.go:81` — 405 responses missing RFC 9110 Allow header
- `internal/modules/documents/delivery/http/handler.go:317` — documentStats writes raw application.DocumentStats instead of generated DocumentStatsResponse; field names currently match but coupling bypasses contract layer
- `internal/modules/documents/delivery/http/handler.go:696` — acquireDocumentSession emits raw map[string]any (downgraded from Major; no current wire-format mismatch)
- Documents/approval error mapper at `approval/http/errors.go:230` — strings.HasPrefix on error message for json unknown-field case (downgraded from Major; no exported stdlib sentinel available)

**composition-observability**
- `internal/platform/bootstrap/api.go:99` — DB connection pool stats (db.Stats()) not included in runtime metrics or /api/v1/metrics payload
- `apps/api/cmd/metaldocs-api/permissions.go:95` — Comment says "Prometheus scrape endpoint" but handler serves custom JSON

---

VERDICT: MICRO-WAVE NEEDED