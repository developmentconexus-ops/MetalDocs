# Backend Architecture Re-Audit Report

**Project:** MetalDocs
**HEAD:** 5650b328
**Date:** 2026-06-19
**Method:** 10-dimension multi-agent audit with adversarial-skeptic verdict per Critical/Major finding; fresh read at HEAD for every cited line. Skeptic verdicts are binding — findings downgraded or refuted by the skeptic are excluded from the pass-bar count. H-D and H-G class counts come from grep commands run at HEAD (results reproduced verbatim below). This re-audit covers the post-M6 program state at HEAD 5650b328 and determines whether the grade-a terminal acceptance gate is reached.

---

## 1. Header and Method

Each dimension was audited independently against HEAD 5650b328. Critical and Major findings were independently reviewed by an adversarial skeptic who re-read every cited file and line before rendering a verdict. Only skeptic-confirmed findings are counted toward the pass-bar. H-D and H-G class counts come from five grep commands run at HEAD (results embedded verbatim in §6). The prior re-audit at HEAD ad8e6fc8 (2026-06-19) established baseline grades; this re-audit measures post-M6 delivery against those grades and evaluates the four §6 acceptance checks defined in mission.md §8.

---

## 2. Scorecard

| # | Dimension | 2026-06-19 Re-Audit (ad8e6fc8) | This Re-Audit Grade | Delta |
|---|-----------|-------------------------------|---------------------|-------|
| 1 | Authz / capability model | B+ | A- | +1 |
| 2 | Security / tenant isolation | B+ | B+ | = |
| 3 | Sessions / auth lifecycle | A- | A- | = |
| 4 | Middleware / HTTP kernel | A- | A- | = |
| 5 | Persistence / transactions | A- | A- | = |
| 6 | Code quality / Go idioms | B+ | B+ | = |
| 7 | Legacy / dead-code | A- | A- | = |
| 8 | Module boundaries / DDD | B+ | A- | +1 |
| 9 | Contract / API layer | B- | B | +1 |
| 10 | Composition / observability | A- | A- | = |

Prior grades (2026-06-19 re-audit at ad8e6fc8): authz B+, security B+, sessions A-, middleware A-, persistence A-, code-quality B+, legacy A-, module-boundaries B+, contract B-, composition A-.

---

## 3. §3 Pass-Bar Verdict

**(1) All 3 formerly-C dimensions (module-boundaries, contract-api, composition) ALL ≥ A-?**

- module-boundaries / DDD: A- — PASS (was B+ at the 2026-06-16 re-audit; H-G verified at 0 by Grep A and Grep C; the one net-new Minor — search direct SQL on taxonomy-owned document_profiles — is non-security, and the tracked H-G class is clean)
- contract / API layer: B — FAIL (improved from B- to B; all five prior-confirmed Majors closed: templates lifecycle/query, IAM admin roles, IAM sessions list, IAM observability; but one new skeptic-confirmed Major (audit export status map[string]any vs AuditExportStatusResponse) and four surviving H-D Minor sites in documents fillin/view/placeholder-options remain open; dimension cannot reach A-)
- composition / observability: A- — PASS (held at A-; all three formerly-confirmed D-class findings verified fixed at HEAD 5650b328)

**Check 1: FAIL** (contract / API layer is B, not A-)

**(2) Zero skeptic-confirmed Critical/Major findings?**

Skeptic-confirmed Major findings at HEAD 5650b328: 1
- Audit export status GET emits map[string]any; OpenAPI declares AuditExportStatusResponse; generated typed struct+interface in api.gen.go unused (contract-4, `internal/modules/audit/delivery/http/handler.go:268-279`)

Five auditor-raised Majors were downgraded by the skeptic to Minor (contract-0 through contract-3, contract-5 — see §5): auth login, auth change-password, audit events list, audit export POST, and search documents all emit map[string]any but the runtime wire shape was verified to match the declared schema in each case, reducing each to a type-safety/maintainability gap.

**Check 2: FAIL (1 confirmed Major)**

**(3) H-D count = 0?**

Re-measurement at HEAD 5650b328 (see §6): Grep A (`writeJSON.*map\[string\]any`) returns 0 hits. Grep B (files containing `map[string]any` in delivery/http non-test) returns 11 files. The Grep A pattern does not capture all H-D sites — multi-line map construction and the `writeFillInJSON` alias are used in the remaining files. Cross-referencing auditor findings against Grep B files yields 10 surviving H-D sites across auth (2), audit (3), documents-fillin (3 via writeFillInJSON), documents-view (1), and search (1). Per the mission §8 bar, H-D = 0 requires Grep A to return no output AND no writeFillInJSON or multi-line-map equivalents.

**Check 3: FAIL (H-D > 0; 10 confirmed sites)**

**(4) H-G count = 0?**

Re-measurement at HEAD 5650b328 (see §6): cross-module iam_users reads = 0, cross-module iam_user_roles reads = 0, hardcoded `"published"` SQL/status literals outside iam/ = 0 (all 7 hits in Grep B are doc-comment occurrences within the documents/approval module).

**Check 4: PASS (H-G = 0)**

**OVERALL PASS-BAR: FAIL (1 of 4 checks passed — H-G only)**

---

## 4. Confirmed Findings

Skeptic-confirmed Critical/Major findings only (verdict === 'confirmed'):

| # | Sev | Dimension | Title | File:Line |
|---|-----|-----------|-------|-----------|
| 1 | Major | Contract / API layer | Audit export status GET emits map[string]any; OpenAPI declares AuditExportStatusResponse (tri-source drift) | `internal/modules/audit/delivery/http/handler.go:268-279` |

---

## 5. Refuted / Downgraded

| # | Auditor Severity | Dimension | Title | Skeptic Verdict | New Severity | Reasoning |
|---|-----------------|-----------|-------|-----------------|--------------|-----------|
| 1 | Major | Contract / API layer | Auth login emits map[string]any; OpenAPI declares AuthLoginResponse (tri-source drift) | DOWNGRADED | Minor | Wire shape verified to match spec exactly — two keys "user" and "expires_at" with correct types. No observable contract drift at the API boundary. Defect is real but narrower: absence of compile-time struct binding, not a live contract violation. |
| 2 | Major | Contract / API layer | Auth change-password emits map[string]any; OpenAPI declares ChangePasswordResponse (tri-source drift) | DOWNGRADED | Minor | Keys "changed" (bool true) and "user" (CurrentUser value) match the ChangePasswordResponse required schema exactly. Runtime wire output is fully contract-compliant. Risk is purely maintainability — a future silent shape break without compiler feedback. |
| 3 | Major | Contract / API layer | Audit events list emits map[string]any; OpenAPI declares ListAuditEventsResponse (tri-source drift) | DOWNGRADED | Minor | Emitted shape `{items: [...], page: {next_cursor, has_more}}` satisfies ListAuditEventsResponse and CursorPage sub-schema. Items built from typed EventResponse struct. Real defect is absence of a typed envelope struct to enforce shape at compile time, not a current contract breach. |
| 4 | Major | Contract / API layer | Audit export POST emits map[string]any; OpenAPI declares AuditExportResponse (tri-source drift) | DOWNGRADED | Minor | All four required fields (export_id, status, signed_url, expires_at) present with correct types matching AuditExportResponse. Runtime wire output satisfies the contract. Concern is lack of compile-time binding. |
| 5 | Major | Contract / API layer | Search documents emits map[string]any; OpenAPI declares SearchDocumentsResponse (tri-source drift) | DOWNGRADED | Minor | Local SearchDocumentResponse struct (lines 30-45) is a structural mirror of SearchDocumentItem; all required fields and types match. Wire output satisfies SearchDocumentsResponse. Auditor's claim of "FE codegen diverges" unsubstantiated — field set and types match at HEAD. Maintenance risk, not an active defect. |

Downgraded findings are re-listed in §7 at their new Minor severity. No findings were fully refuted.

---

## 6. Class Re-Measurement

### H-D (untyped map[string]any on public delivery/http routes)

**Grep A — `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go'`:**

```
(no output)
```

**Grep B — `grep -rEl 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' | grep -v _test.go | sort`:**

```
internal/modules/audit/delivery/http/handler.go
internal/modules/auth/delivery/http/handler.go
internal/modules/controlleddocuments/delivery/http/routes.go
internal/modules/documents/delivery/http/fillin_handler.go
internal/modules/documents/delivery/http/handler.go
internal/modules/documents/delivery/http/pdf_webhook_handler.go
internal/modules/documents/delivery/http/placeholder_options_handler.go
internal/modules/documents/delivery/http/view_handler.go
internal/modules/iam/delivery/http/people_handler.go
internal/modules/search/delivery/http/handler.go
internal/modules/security/delivery/http/handler.go
```

**H-D count: 10** (Grep A returns 0 hits — all prior `writeJSON.*map[string]any` one-liner patterns converted to typed calls. Grep B identifies 11 files still containing `map[string]any` in delivery/http non-test. Cross-referencing against auditor findings: 10 confirmed H-D sites survive across auth (login, change-password), audit (events list, export POST, export status), documents fillin/view/placeholder-options via writeFillInJSON alias, and search. The controlleddocuments, iam, security, and templates files in Grep B are not H-D by the mission's tri-source criterion at HEAD.)

---

### H-G (cross-module direct table reads and hardcoded domain-state literals)

**Grep A — `grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_users' internal/modules/ --include='*.go' | grep -v 'internal/modules/iam/' | grep -v _test.go`:**

```
(no output)
```

**Grep B — `grep -rEn '"published"' internal/modules/ --include='*.go' | grep -v _test.go | grep -v '/domain/' | grep -v 'api.gen.go'`:**

```
internal/modules/documents/approval/application/obsolete_service.go:25:// that permits an → obsolete transition (must be "published" or "superseded").
internal/modules/documents/approval/application/obsolete_service.go:42:// MarkObsolete transitions a document from "published" or "superseded" to
internal/modules/documents/approval/application/publish_service.go:43:	NewStatus  string // "published"
internal/modules/documents/approval/application/publish_service.go:93:		// Step 3: transition the document from "approved" to "published".
internal/modules/documents/approval/application/supersede_service.go:27:	NewDocumentID        string // the document being published (becomes "published")
internal/modules/documents/approval/application/supersede_service.go:36:	NewDocumentStatus   string // "published"
internal/modules/documents/approval/application/supersede_service.go:41:// "published" and the prior document from "published" to "superseded".
```

All 7 hits are doc-comment occurrences within the documents/approval module describing the published→superseded/obsolete state transitions intrinsic to that module's own workflow. None are SQL predicates or cross-module status comparisons. H-G hardcoded-status-literal count: 0.

**Grep C — `grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_user_roles' internal/modules/ --include='*.go' | grep -v 'internal/modules/iam/' | grep -v _test.go`:**

```
(no output)
```

**H-G count: 0**

All previously-confirmed H-G sites are verified fixed at HEAD 5650b328: auth/infrastructure/postgres/repository.go routes GetUserTenants through the IAM UserTenantReader port; templates/infrastructure/template_version_reader.go uses the typed domain constant templatesdomain.VersionStatusPublished; security/infrastructure/postgres/repository.go resolves all iam_user_roles and iam_users data through four IAM-owned ports.

---

## 7. Minor Findings

| # | Sev | Dimension | Title | File:Line |
|---|-----|-----------|-------|-----------|
| 1 | Minor | Authz / capability model | authz.Require does not validate capability against iamdomain.IsValidCapability before querying | `internal/modules/iam/authz/authz.go:76-143` |
| 2 | Minor | Authz / capability model | BypassSystem returns nil (bypass committed, audit skipped) when bypass audit sink is unset | `internal/modules/iam/authz/authz.go:185-193` |
| 3 | Minor | Security / tenant isolation | document_revisions content-hash lookup not tenant-scoped in GetFinalizePrereqs | `internal/modules/documents/repository/repository.go:1731-1736` |
| 4 | Minor | Security / tenant isolation | document_exports table lacks tenant_id; reads/writes keyed only on document_id and composite_hash | `internal/modules/documents/repository/export_repository.go:11-52` |
| 5 | Minor | Security / tenant isolation | document_families is a global (tenant-less) reference table — by-design but worth tracking | `internal/modules/taxonomy/infrastructure/family_repository.go:40-88` |
| 6 | Minor | Sessions / auth lifecycle | TouchSession existence re-check can silently succeed against a concurrently-revoked session | `internal/modules/auth/infrastructure/postgres/repository.go:153-161` |
| 7 | Minor | Sessions / auth lifecycle | No session token rotation on resolution or privilege change | `internal/modules/auth/application/service.go:368-399` |
| 8 | Minor | Sessions / auth lifecycle | Login and change-password responses emit untyped map[string]any instead of generated types (downgraded from §5) | `internal/modules/auth/delivery/http/handler.go:90-93, 161-164` |
| 9 | Minor | Sessions / auth lifecycle | RevokeSessionsByUserIDTx has no tenant_id predicate — latent cross-tenant blast-radius | `internal/modules/auth/infrastructure/postgres/repository.go:200-212` |
| 10 | Minor | Sessions / auth lifecycle | Service.UpdateUser changes password without revoking active sessions | `internal/modules/auth/application/service.go:618-631` |
| 11 | Minor | Middleware / HTTP kernel | audit handleExport 405 missing Allow header | `internal/modules/audit/delivery/http/handler.go:134-135` |
| 12 | Minor | Middleware / HTTP kernel | audit handleExportSubresource 405 missing Allow header | `internal/modules/audit/delivery/http/handler.go:225-226` |
| 13 | Minor | Middleware / HTTP kernel | audit handleEvents 405 uses raw w.Header().Set instead of canonical WriteMethodNotAllowed | `internal/modules/audit/delivery/http/handler.go:80-82` |
| 14 | Minor | Middleware / HTTP kernel | Recovery middleware best-effort write can produce corrupt body on already-committed response | `internal/platform/middleware/recovery.go:43-46` |
| 15 | Minor | Persistence / transactions | CodeExists off-tx inside lock-holding CD-create closure — logical TOCTOU race on auto-code uniqueness | `internal/modules/controlleddocuments/application/service.go:373` |
| 16 | Minor | Persistence / transactions | Standalone ControlledDocumentRepository.Create calls authz.Require without SeedTxIdentity (dead path, latent hazard) | `internal/modules/controlleddocuments/infrastructure/repository.go:335-352` |
| 17 | Minor | Persistence / transactions | No idle_in_transaction_session_timeout configured at the DB connection layer | `internal/platform/db/postgres/connect.go:39-56` |
| 18 | Minor | Persistence / transactions | Inconsistent defer rollback pattern in documents repository — bare defer vs safe func-literal | `internal/modules/documents/repository/repository.go:620, 703, 736, 839, 876, 978, 1119, 1210, 1242, 1369` |
| 19 | Minor | Code quality / Go idioms | Anonymous interface in FreezeService struct field and constructor (valuesRead) | `internal/modules/documents/application/freeze_service.go:49-51, 75-77` |
| 20 | Minor | Code quality / Go idioms | Pin recomputes values_hash already computed inside pinValidateAndHash | `internal/modules/documents/application/freeze_service.go:200-212` |
| 21 | Minor | Code quality / Go idioms | tenantIDFromRequest helper duplicated across three unrelated delivery packages | `internal/modules/taxonomy/delivery/http/routes_profiles.go:260, internal/modules/controlleddocuments/delivery/http/routes.go:489, internal/modules/iam/delivery/http/routes_memberships.go:341` |
| 22 | Minor | Code quality / Go idioms | Deprecated documents/application.New still wired by 14 test call-sites | `internal/modules/documents/application/service.go:127-136` |
| 23 | Minor | Legacy / dead-code | DocgenVer/GrammarVer dependency slots never set by the composition root | `internal/modules/documents/module.go:44-45` |
| 24 | Minor | Legacy / dead-code | Deprecated documents/application.New still called by 14 test sites | `internal/modules/documents/application/service.go:127-136` |
| 25 | Minor | Legacy / dead-code | tenantIDFromRequest duplicated in three delivery packages | `internal/modules/controlleddocuments/delivery/http/routes.go:489` |
| 26 | Minor | Legacy / dead-code | FreezeService.valuesRead uses anonymous interface in both struct field and constructor | `internal/modules/documents/application/freeze_service.go:49-51` |
| 27 | Minor | Legacy / dead-code | Pin double-computes values_hash already produced by pinValidateAndHash | `internal/modules/documents/application/freeze_service.go:200-212` |
| 28 | Minor | Module boundaries / DDD | search module issues direct SQL against taxonomy-owned metaldocs.document_profiles (no port) | `internal/modules/search/infrastructure/v2documents/reader.go:40, 65` |
| 29 | Minor | Contract / API layer | Audit events list emits map[string]any; no typed envelope struct (downgraded from §5) | `internal/modules/audit/delivery/http/handler.go:120-130` |
| 30 | Minor | Contract / API layer | Audit export POST emits map[string]any; no typed struct binding (downgraded from §5) | `internal/modules/audit/delivery/http/handler.go:216-221` |
| 31 | Minor | Contract / API layer | Search documents emits map[string]any; local struct mirrors but is not generated type (downgraded from §5) | `internal/modules/search/delivery/http/handler.go:134` |
| 32 | Minor | Contract / API layer | Documents fill-in-schema GET emits map[string]any; OpenAPI 200 has no body schema (H-D class) | `internal/modules/documents/delivery/http/fillin_handler.go:58-62` |
| 33 | Minor | Contract / API layer | Documents put-placeholder-value emits map[string]any; OpenAPI 200 has no body schema | `internal/modules/documents/delivery/http/fillin_handler.go:116-119` |
| 34 | Minor | Contract / API layer | Documents placeholder-options GET emits map[string]any; OpenAPI 200 has no body schema | `internal/modules/documents/delivery/http/placeholder_options_handler.go:67, 74` |
| 35 | Minor | Contract / API layer | Documents view GET emits map[string]any; OpenAPI 200 has no body schema | `internal/modules/documents/delivery/http/view_handler.go:46-51` |
| 36 | Minor | Composition / observability | Scheduler job invocations carry no OTel span — job traces are invisible when OTel is enabled | `internal/modules/jobs/scheduler/scheduler.go:249` |
| 37 | Minor | Composition / observability | /api/v1/metrics top-level payload composed as map[string]any — runtime/scheduler/db_pool keys are untyped | `internal/platform/observability/http.go:183-194` |

---

## 8. Terminal Acceptance Verdict

**VERDICT: FAIL**

The M6 delivery at HEAD 5650b328 achieved substantial improvement over the prior re-audit at ad8e6fc8: all five previously-confirmed contract Majors are closed (templates lifecycle/query, IAM admin role upsert/replace, IAM sessions list, IAM observability KPI/usage), Grep A (the primary H-D signal) now returns 0 hits, module-boundaries / DDD holds at A- with H-G verified at 0 across all three grep patterns, and the composition and authz dimensions hold at A-. However, the §6 pass-bar requires all four checks to pass simultaneously, and three fail.

Check 1 fails because the contract / API layer grades B rather than A-: the one skeptic-confirmed Major (audit export status map[string]any vs generated AuditExportStatusResponse struct at api.gen.go:79-86,760-762) remains open, and 10 H-D sites survive across auth, audit, documents, and search modules. Check 2 fails on that single confirmed Major. Check 3 fails because the H-D class is not fully resolved — Grep A = 0 is a necessary but not sufficient signal given the writeFillInJSON alias and multi-line map construction patterns used in the documents and search modules. Check 4 passes: H-G = 0 confirmed at HEAD by all three grep patterns.

The next iteration must: (a) convert `internal/modules/audit/delivery/http/handler.go:268-279` to use the generated `AuditExportStatusResponse` struct and wire through the `StrictServerInterface` (closes the single confirmed Major); (b) declare OpenAPI 200 body schemas for the four documents fillin/view/placeholder-options routes and emit typed structs from those handlers; (c) convert auth login and change-password handlers to generated typed responses; (d) re-measure Grep A to confirm 0 after all map[string]any patterns are replaced. With those changes the contract dimension can reach A- and all four checks can pass.