# Backend Architecture Re-Audit Report

**Project:** MetalDocs
**HEAD:** ad8e6fc8
**Date:** 2026-06-19
**Method:** 10-dimension multi-agent audit with adversarial-skeptic verdict per Critical/Major finding; fresh read at HEAD for every cited line. Skeptic verdicts are binding — findings downgraded or refuted by the skeptic are excluded from the pass-bar count.

---

## 1. Header and Method

Each dimension was audited independently against HEAD ad8e6fc8. Critical and Major findings were independently reviewed by an adversarial skeptic who re-read every cited file and line before rendering a verdict. Only skeptic-confirmed findings are counted toward the pass-bar. H-D and H-G class counts come from grep commands run at HEAD (results reproduced below). The F5.1 mission re-audits the post-M5 program state to determine whether the grade-a terminal acceptance gate is reached.

---

## 2. Scorecard

| # | Dimension | 2026-06-16 Re-Audit | This Re-Audit Grade | Delta |
|---|-----------|---------------------|---------------------|-------|
| 1 | Authz / capability model | B+ | A- | +1 |
| 2 | Security / tenant isolation | B+ | B+ | = |
| 3 | Sessions / auth lifecycle | A- | A- | = |
| 4 | Middleware / HTTP kernel | A- | A- | = |
| 5 | Persistence / transactions | A- | A- | = |
| 6 | Code quality / Go idioms | B+ | B+ | = |
| 7 | Legacy / dead-code | A- | A- | = |
| 8 | Module boundaries / DDD | B+ | A- | +1 |
| 9 | Contract / API layer | B+ | B- | -1 |
| 10 | Composition / observability | A- | A- | = |

Prior grades (2026-06-16): authz B+, security B+, sessions A-, middleware A-, persistence A-, code-quality B+, legacy A-, module-boundaries B+, contract B+, composition A-.

---

## 3. §6 Pass-Bar Verdict

**(1) All 3 formerly-C dimensions (module-boundaries, contract-api, composition) ALL ≥ A-?**

- module-boundaries / DDD: A- — PASS (was B+, now A-; H-G fixes verified at auth/infrastructure/postgres/repository.go:117 via UserTenantReader port and templates/infrastructure/template_version_reader.go:45 via typed VersionStatusPublished constant)
- contract / API layer: B- — FAIL (regressed from B+ to B-; the two prior cited Majors closed, but a broader audit at HEAD finds 24 surviving map[string]any writeJSON sites across public delivery routes plus tri-source drift with OpenAPI on the templates lifecycle endpoints)
- composition / observability: A- — PASS (held at A-)

**Check 1: FAIL** (contract / API layer below A-)

**(2) Zero skeptic-confirmed Critical/Major findings?**

Skeptic-confirmed Major findings: 6 (all in Contract / API layer)
- templates lifecycle routes emit map[string]any envelopes; OpenAPI 200 declares no body schema (`routes_lifecycle.go:46,100,164,196,239`)
- templates query routes emit map[string]any instead of generated typed responses (`routes_query.go:73,104,145,211,260`)
- IAM admin role upsert and replace endpoints emit untyped map[string]any (`admin_handler.go:341,378`)
- IAM sessions list emits nested map[string]any with no typed schema (`sessions_handler.go:132,138,158`)
- IAM observability KPI and usage helpers return map[string]any (`observability_handler.go:81,109`)
- (one auditor-raised Major refuted by the skeptic — see §5)

**Check 2: FAIL (5 confirmed Majors)**

**(3) H-D count = 0?**

Re-measurement at HEAD ad8e6fc8 (see §6): writeJSON-map[string]any grep returns 24 hits across templates lifecycle/query/catalog/schema, IAM admin/memberships/sessions, security, and taxonomy. The class is wider than the 2026-06-16 spot-checks captured.

**Check 3: FAIL (H-D = 24)**

**(4) H-G count = 0?**

Re-measurement at HEAD ad8e6fc8 (see §6): cross-module iam_users reads = 0, cross-module iam_user_roles reads = 0, hardcoded `"published"` SQL literals = 0 (the only `"published"` string hits are doc-comment occurrences in approval-service files describing transitions, not SQL or status comparisons).

**Check 4: PASS (H-G = 0)**

**OVERALL PASS-BAR: FAIL (1 of 4 checks passed — H-G only)**

---

## 4. Confirmed Findings

| # | Sev | Dimension | Title | File:Line |
|---|-----|-----------|-------|-----------|
| 1 | Major | Contract / API layer | Templates lifecycle routes emit map[string]any envelopes on 5 public routes; OpenAPI 200 declares no body schema (tri-source drift) | `internal/modules/templates/delivery/http/routes_lifecycle.go:46,100,164,196,239` |
| 2 | Major | Contract / API layer | Templates query routes (list/get/getVersion) emit map[string]any instead of generated typed responses | `internal/modules/templates/delivery/http/routes_query.go:73,104,145,211,260` |
| 3 | Major | Contract / API layer | IAM admin role upsert and replace endpoints emit untyped map[string]any | `internal/modules/iam/delivery/http/admin_handler.go:341,378` |
| 4 | Major | Contract / API layer | IAM sessions list emits nested map[string]any with no typed schema | `internal/modules/iam/delivery/http/sessions_handler.go:132,138,158` |
| 5 | Major | Contract / API layer | IAM observability KPI and usage helpers return map[string]any consumed by public /iam/kpi and /iam/usage routes | `internal/modules/iam/delivery/http/observability_handler.go:81,109` |

---

## 5. Refuted / Downgraded Findings

| # | Auditor Severity | Dimension | Title | Skeptic Verdict | Reasoning |
|---|------------------|-----------|-------|-----------------|-----------|
| 1 | Major | Contract / API layer | IAM area-membership list and grant emit untyped envelopes (`routes_memberships.go:168,235`) | REFUTED | Documented model under ADR 0012 partial rollout. File header at `routes_memberships.go:4-6` explicitly cites "Hand-rolled rather than codegen-served — IAM is still pre-codegen on the BE side per ADR 0012 partial rollout." ADR 0012 lists IAM in the deferred scope (lines 11, 57, 71); StrictServerInterface adoption was explicitly deferred for future per-module passes. Cross-listed in §7 because the underlying untyped-envelope pattern is real but is a tracked Minor per ADR 0012 backlog, not a fresh contract gap. |

No findings were severity-downgraded; one Major was fully refuted on documented-model grounds and the affected sites move to §7 as Minors.

---

## 6. Class Re-Measurement

### H-D (untyped map[string]any on public delivery/http routes)

**Grep A — `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go'`:**

```
internal/modules/iam/delivery/http/admin_handler.go:341:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/iam/delivery/http/admin_handler.go:378:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/iam/delivery/http/routes_memberships.go:168:	writeJSON(w, http.StatusOK, map[string]any{"items": dtos})
internal/modules/iam/delivery/http/routes_memberships.go:235:	writeJSON(w, http.StatusCreated, map[string]any{
internal/modules/iam/delivery/http/sessions_handler.go:158:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/security/delivery/http/handler.go:67:		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
internal/modules/security/delivery/http/handler.go:94:	writeJSON(w, http.StatusOK, map[string]any{"items": out})
internal/modules/security/delivery/http/handler.go:107:		writeJSON(w, http.StatusOK, map[string]any{"items": []any{}})
internal/modules/security/delivery/http/handler.go:130:	writeJSON(w, http.StatusOK, map[string]any{"items": out})
internal/modules/security/delivery/http/handler.go:173:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/taxonomy/delivery/http/routes_areas.go:40:	writeJSON(w, http.StatusOK, map[string]any{"items": items})
internal/modules/taxonomy/delivery/http/routes_families.go:31:	writeJSON(w, http.StatusOK, map[string]any{"items": items})
internal/modules/templates/delivery/http/routes_catalog.go:35:	writeJSON(w, http.StatusOK, map[string]any{"items": placeholderCatalog})
internal/modules/templates/delivery/http/routes_lifecycle.go:46:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_lifecycle.go:100:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_lifecycle.go:164:	writeJSON(w, http.StatusOK, map[string]any{"data": data})
internal/modules/templates/delivery/http/routes_lifecycle.go:196:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_lifecycle.go:239:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_query.go:73:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_query.go:104:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_query.go:145:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_query.go:211:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_query.go:260:	writeJSON(w, http.StatusOK, map[string]any{
internal/modules/templates/delivery/http/routes_schema.go:68:	writeJSON(w, http.StatusOK, map[string]any{
```

**Grep B — `grep -rEl 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' | grep -v _test.go | sort` (files):**

```
internal/modules/audit/delivery/http/handler.go
internal/modules/auth/delivery/http/handler.go
internal/modules/controlleddocuments/delivery/http/routes.go
internal/modules/documents/delivery/http/fillin_handler.go
internal/modules/documents/delivery/http/handler.go
internal/modules/documents/delivery/http/pdf_webhook_handler.go
internal/modules/documents/delivery/http/placeholder_options_handler.go
internal/modules/documents/delivery/http/view_handler.go
internal/modules/iam/delivery/http/admin_handler.go
internal/modules/iam/delivery/http/observability_handler.go
internal/modules/iam/delivery/http/people_handler.go
internal/modules/iam/delivery/http/routes_memberships.go
internal/modules/iam/delivery/http/sessions_handler.go
internal/modules/search/delivery/http/handler.go
internal/modules/security/delivery/http/handler.go
internal/modules/taxonomy/delivery/http/routes_areas.go
internal/modules/taxonomy/delivery/http/routes_families.go
internal/modules/templates/delivery/http/routes_catalog.go
internal/modules/templates/delivery/http/routes_lifecycle.go
internal/modules/templates/delivery/http/routes_query.go
internal/modules/templates/delivery/http/routes_schema.go
```

**H-D count: 24** (writeJSON-map[string]any sites on public delivery routes, per Grep A)

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

All 7 hits are doc-comment / code-comment occurrences within the documents/approval module describing the published→superseded/obsolete state transitions intrinsic to that module's own workflow. None are SQL literals or cross-module status comparisons. H-G hardcoded-status-literal count: 0.

**Grep C — `grep -rEn 'FROM[[:space:]]+(metaldocs\.)?iam_user_roles' internal/modules/ --include='*.go' | grep -v 'internal/modules/iam/' | grep -v _test.go`:**

```
(no output)
```

**H-G count: 0**

The two H-G sites flagged at the 2026-06-16 re-audit are both verified fixed at HEAD ad8e6fc8: `auth/infrastructure/postgres/repository.go:117` now routes GetUserTenants through the UserTenantReader IAM port (F5.2), and `templates/infrastructure/template_version_reader.go:45` now compares against the typed domain constant `templatesdomain.VersionStatusPublished` (F5.1).

---

## 7. Minor Findings

| # | Sev | Dimension | Title | File:Line |
|---|-----|-----------|-------|-----------|
| 1 | Minor | Authz / capability model | authz.Require does not validate capability against iamdomain.IsValidCapability before querying | `internal/modules/iam/authz/authz.go:76-138` |
| 2 | Minor | Authz / capability model | BypassSystem returns nil when bypass audit sink is unset (composition-root forgetfulness footgun) | `internal/modules/iam/authz/authz.go:185-193` |
| 3 | Minor | Security / tenant isolation | document_revisions content-hash lookup not tenant-scoped in GetFinalizePrereqs | `internal/modules/documents/repository/repository.go:1731-1736` |
| 4 | Minor | Security / tenant isolation | document_exports table lacks tenant_id; reads/writes keyed only on document_id | `internal/modules/documents/repository/export_repository.go:11-52` |
| 5 | Minor | Security / tenant isolation | document_families is a global (tenant-less) table | `internal/modules/taxonomy/infrastructure/family_repository.go:40-160` |
| 6 | Minor | Sessions / auth lifecycle | TouchSession existence re-check can silently succeed against a concurrently-revoked session | `internal/modules/auth/infrastructure/postgres/repository.go:137-163` |
| 7 | Minor | Sessions / auth lifecycle | No session token rotation on resolution / on privilege change | `internal/modules/auth/application/service.go:368-399` |
| 8 | Minor | Sessions / auth lifecycle | Login response emits untyped map[string]any instead of generated response type | `internal/modules/auth/delivery/http/handler.go:90-93` |
| 9 | Minor | Sessions / auth lifecycle | RevokeSessionsByUserIDTx has no tenant_id predicate (latent cross-tenant blast-radius) | `internal/modules/auth/infrastructure/postgres/repository.go:200-212` |
| 10 | Minor | Sessions / auth lifecycle | Service.UpdateUser can change password without revoking active sessions (port footgun) | `internal/modules/auth/application/service.go:618-631` |
| 11 | Minor | Middleware / HTTP kernel | audit handleExport 405 missing Allow header | `internal/modules/audit/delivery/http/handler.go:134-136` |
| 12 | Minor | Middleware / HTTP kernel | audit handleExportSubresource 405 missing Allow header | `internal/modules/audit/delivery/http/handler.go:225-227` |
| 13 | Minor | Middleware / HTTP kernel | audit handleEvents 405 uses raw header.Set instead of canonical WriteMethodNotAllowed | `internal/modules/audit/delivery/http/handler.go:80-83` |
| 14 | Minor | Middleware / HTTP kernel | Recovery middleware best-effort write can produce corrupt body on already-committed response | `internal/platform/middleware/recovery.go:43-46` |
| 15 | Minor | Persistence / transactions | CodeExists off-tx inside lock-holding CD-create transaction (H-PRE-1 variant) | `internal/modules/controlleddocuments/application/service.go:373` |
| 16 | Minor | Persistence / transactions | Standalone ControlledDocumentRepository.Create calls authz.Require without SeedTxIdentity | `internal/modules/controlleddocuments/infrastructure/repository.go:335-352` |
| 17 | Minor | Persistence / transactions | No idle_in_transaction_session_timeout configured at the DB connection layer | `internal/platform/db/postgres/connect.go:39-56` |
| 18 | Minor | Persistence / transactions | Inconsistent defer rollback patterns in documents repository (bare vs `_ =`) | `internal/modules/documents/repository/repository.go:620,703,736,839,876` |
| 19 | Minor | Code quality / Go idioms | Anonymous interface in struct field and constructor (valuesRead) | `internal/modules/documents/application/freeze_service.go:49-51, 75-77` |
| 20 | Minor | Code quality / Go idioms | Pin recomputes values_hash already computed inside pinValidateAndHash | `internal/modules/documents/application/freeze_service.go:200-212` |
| 21 | Minor | Legacy / dead-code | DocgenVer / GrammarVer fields in documents.Dependencies are never set by the composition root | `internal/modules/documents/module.go:44-45` |
| 22 | Minor | Legacy / dead-code | Deprecated documents/application.New still wired by 14 test call-sites | `internal/modules/documents/application/service.go:127-136` |
| 23 | Minor | Contract / API layer | Audit list/export still emit map[string]any envelope | `internal/modules/audit/delivery/http/handler.go:120,127,216,268` |
| 24 | Minor | Contract / API layer | Auth login and refresh emit map[string]any | `internal/modules/auth/delivery/http/handler.go:90,161` |
| 25 | Minor | Contract / API layer | Documents fillin/view/pdf-webhook/placeholder-options emit map[string]any | `internal/modules/documents/delivery/http/fillin_handler.go:58,116` |
| 26 | Minor | Contract / API layer | Security signals/risk/MFA-coverage endpoints emit map[string]any | `internal/modules/security/delivery/http/handler.go:67,94,107,130,154,173` |
| 27 | Minor | Contract / API layer | Taxonomy areas/families list endpoints untyped | `internal/modules/taxonomy/delivery/http/routes_areas.go:40` |
| 28 | Minor | Contract / API layer | Templates placeholder-catalog and schema GET emit map[string]any | `internal/modules/templates/delivery/http/routes_catalog.go:35` |
| 29 | Minor | Contract / API layer | Search endpoint emits map[string]any | `internal/modules/search/delivery/http/handler.go:134` |
| 30 | Minor | Contract / API layer | IAM area-membership list and grant emit untyped envelopes (downgraded from §5 — ADR 0012 deferred scope) | `internal/modules/iam/delivery/http/routes_memberships.go:168,235` |
| 31 | Minor | Composition / observability | App-level OTel spans not extended to scheduler job invocations or background sweepers | `internal/modules/jobs/scheduler/scheduler.go:249` |
| 32 | Minor | Composition / observability | Scheduler job duration is logged but not exported as a metric in /api/v1/metrics | `internal/modules/jobs/scheduler/scheduler.go:256, 286-319` |
| 33 | Minor | Composition / observability | /api/v1/metrics payload composed entirely of map[string]any (runtime, scheduler, db_pool) | `internal/platform/observability/http.go:183-195` |

---

## 8. Terminal Acceptance Verdict

**VERDICT: FAIL**

The F5.x mission delivered the targeted M5 remediations cleanly: H-G is reduced to 0 (cross-module iam_users and iam_user_roles reads eliminated, the hardcoded `"published"` SQL literal in templates infrastructure replaced with a typed domain constant), authz climbed to A- with the iam_users tenant_id upsert fix verified at `role_admin_repository.go:65-71, 117-123`, and module-boundaries / DDD crossed into A-. Sessions, middleware, persistence, legacy, and composition all held at A-. However, the §6 pass-bar requires all four checks to pass simultaneously and only Check 4 (H-G = 0) passes. Check 1 fails because the contract / API layer regressed from B+ to B- when a broader audit at HEAD exposed 24 surviving map[string]any writeJSON sites on public delivery routes plus tri-source drift between handlers, OpenAPI, and generated bindings on the templates lifecycle endpoints. Check 2 fails on 5 skeptic-confirmed Majors, all in the contract dimension (templates lifecycle, templates query, IAM admin role upsert/replace, IAM sessions list, IAM observability KPI/usage). Check 3 fails with H-D = 24 (was 2 at 2026-06-16; the prior audit had spot-counted only routes_generated.go and missed the wider class). The next mission must convert all writeJSON map[string]any sites to generated typed *JSONResponse structs, complete the OpenAPI 200 schema declarations for the templates lifecycle endpoints (submit / review / archive / upsertApprovalConfig and align approveTemplateVersion to its declared ApproveTemplateVersionResponse), and drive the contract / API layer to A- before re-audit.
