# Backend Architecture Re-Audit Report

**Project:** MetalDocs
**HEAD:** 9a2a2f8d746b1f9e6820367134c6578f3ddc814c
**Date:** 2026-06-16
**Method:** 10-dimension multi-agent audit with adversarial-skeptic verdict per Critical/Major finding; fresh read at HEAD for every cited line. Skeptic verdicts are binding — findings downgraded or refuted by the skeptic are excluded from the pass-bar count.

---

## 1. Header and Method

Each dimension was audited independently against HEAD. Critical and Major findings were independently reviewed by an adversarial skeptic who re-read every cited file and line before rendering a verdict. Only skeptic-confirmed findings are counted toward the pass-bar. H-D and H-G class counts come from grep commands run at HEAD (results reproduced below).

---

## 2. Scorecard

| # | Dimension | 2026-06-15 Re-Audit | This Re-Audit Grade | Delta |
|---|-----------|---------------------|---------------------|-------|
| 1 | Authz / capability model | B+ | B+ | = |
| 2 | Security / tenant isolation | B+ | B+ | = |
| 3 | Sessions / auth lifecycle | B+ | A- | +1 |
| 4 | Middleware / HTTP kernel | B+ | A- | +1 |
| 5 | Persistence / transactions | B+ | A- | +1 |
| 6 | Code quality / Go idioms | B | B+ | +1 |
| 7 | Legacy / dead-code | B+ | A- | +1 |
| 8 | Module boundaries / DDD | B+ | B+ | = |
| 9 | Contract / API layer | C+ | B+ | +1 |
| 10 | Composition / observability | B+ | A- | +1 |

Prior grades (2026-06-15): authz B+, security B+, sessions B+, middleware B+, persistence B+, code-quality B, legacy B+, module-boundaries B+, contract C+, composition B+.

---

## 3. §6 Pass-Bar Verdict

**(1) All 3 formerly-C dimensions ≥ A-?**

- module-boundaries/DDD: B+ — FAIL (was not a formerly-C dimension; prior grade was B+)
- contract/API layer: B+ — FAIL (was C+, must reach A-)
- composition/observability: A- — PASS (was B+, now A-)

The contract/API layer dimension moved from C+ to B+ but did not reach the A- threshold required by the pass-bar. **Check 1: FAIL**

**(2) Zero skeptic-confirmed Critical/Major findings?**

Skeptic-confirmed Major findings: 4
- authz.Require denies time-bounded active memberships (authz.go:124)
- iam_users upsert omits tenant_id (role_admin_repository.go:61-68, 109-116)
- templates: toTemplateResponse/toVersionResponse map[string]any wired to all query and lifecycle routes (routes_create.go:44,67)
- iam/admin_handler.go: admin overview emits map[string]any (admin_handler.go:262)

**Check 2: FAIL (4 confirmed Majors)**

**(3) H-D count = 0?**

Confirmed H-D sites: 2
- `internal/modules/templates/delivery/http/routes_generated.go:128` — writeJSON with map[string]any on public route
- `internal/modules/templates/delivery/http/routes_generated.go:238` — writeJSON with map[string]any on public route

**Check 3: FAIL (H-D = 2)**

**(4) H-G count = 0?**

Confirmed H-G sites: 2
- `internal/modules/auth/infrastructure/postgres/repository.go:104` — direct read of `metaldocs.iam_user_roles` from auth module
- `internal/modules/templates/infrastructure/template_version_reader.go:44` — hardcoded `"published"` string literal instead of typed domain constant

**Check 4: FAIL (H-G = 2)**

**OVERALL PASS-BAR: FAIL (0 of 4 checks passed)**

---

## 4. Confirmed Findings

| # | Sev | Dimension | Title | File:Line |
|---|-----|-----------|-------|-----------|
| 1 | Major | Authz / capability model | authz.Require denies time-bounded active memberships (effective_to predicate too strict) | `internal/modules/iam/authz/authz.go:124` |
| 2 | Major | Security / tenant isolation | iam_users upsert omits tenant_id — always writes to system/default tenant | `internal/modules/iam/infrastructure/postgres/role_admin_repository.go:61-68, 109-116` |
| 3 | Major | Contract / API layer | templates: toTemplateResponse and toVersionResponse are map[string]any helpers wired to all query and lifecycle routes | `internal/modules/templates/delivery/http/routes_create.go:44,67` |
| 4 | Major | Contract / API layer | iam/admin_handler.go: admin overview emits map[string]any with nested []map[string]any | `internal/modules/iam/delivery/http/admin_handler.go:262` |

---

## 5. Refuted / Downgraded Findings

No auditor-raised findings were refuted or downgraded by the skeptic in this cycle. All four Critical/Major findings brought to skeptic review were confirmed at the stated severity.

---

## 6. Class Re-Measurement

### H-D (untyped map[string]any on public delivery/http routes)

**Grep 1 — files containing map[string]any in delivery/http (non-test):**
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
internal/modules/templates/delivery/http/routes_create.go
internal/modules/templates/delivery/http/routes_generated.go
internal/modules/templates/delivery/http/routes_lifecycle.go
internal/modules/templates/delivery/http/routes_query.go
internal/modules/templates/delivery/http/routes_schema.go
```

**Grep 2 — writeJSON calls in templates routes_generated.go:**
```
routes_generated.go:77:   writeJSON(w, http.StatusCreated, resp)
routes_generated.go:128:  writeJSON(w, http.StatusOK, map[string]any{...})
routes_generated.go:238:  writeJSON(w, http.StatusOK, map[string]any{...})
```

**H-D count: 2**

**H-D sites:**
1. `internal/modules/templates/delivery/http/routes_generated.go:128` — public route returns untyped map[string]any response body instead of typed OpenAPI response struct
2. `internal/modules/templates/delivery/http/routes_generated.go:238` — public route returns untyped map[string]any response body instead of typed OpenAPI response struct

Note: The broader file list above reflects the full prevalence of map[string]any across delivery handlers (feeding into the Major findings in §4 and §7), but the H-D class counter is scoped to the two confirmed sites in routes_generated.go per the grep methodology applied.

---

### H-G (cross-module direct table reads and hardcoded domain-state literals)

**Grep 1 — cross-module iam_users reads (excluding iam/ and _test.go):**
```
(no output)
```

**Grep 2 — hardcoded status literals (excluding _test.go, domain/, api.gen.go, api/api):**
```
internal/modules/templates/infrastructure/template_version_reader.go:44:
    if !status.Valid || status.String != "published" {
```

**Grep 3 — cross-module iam_user_roles reads (excluding iam/ and _test.go):**
```
internal/modules/auth/infrastructure/postgres/repository.go:104:
    FROM metaldocs.iam_user_roles
```

**H-G count: 2**

**H-G sites:**
1. `internal/modules/auth/infrastructure/postgres/repository.go:104` — auth module reads iam_user_roles directly; iam_user_roles is IAM-owned
2. `internal/modules/templates/infrastructure/template_version_reader.go:44` — hardcoded `"published"` string literal; should reference a typed domain constant

---

## 7. Minor Findings

| # | Sev | Dimension | Title | File:Line |
|---|-----|-----------|-------|-----------|
| 1 | Minor | Authz / capability model | MembershipDirectoryScope has_managed_areas subquery inherits effective_to IS NULL flaw | `internal/modules/iam/infrastructure/postgres/user_area_repository.go:155, 190` |
| 2 | Minor | Security / tenant isolation | document_revisions queried without tenant join in GetFinalizePrereqs content-hash step | `internal/modules/documents/repository/repository.go:1731-1736` |
| 3 | Minor | Security / tenant isolation | document_exports table has no tenant_id — InsertExport and GetExportByHash scope by document_id only | `internal/modules/documents/repository/export_repository.go:11-52` |
| 4 | Minor | Security / tenant isolation | document_families is a global table with no tenant_id — reads and writes cross all tenants | `internal/modules/taxonomy/infrastructure/family_repository.go:40-112` |
| 5 | Minor | Sessions / auth lifecycle | TouchSession existence re-check can silently succeed on concurrently-revoked session | `internal/modules/auth/infrastructure/postgres/repository.go:156-163` |
| 6 | Minor | Sessions / auth lifecycle | No session token rotation on resolution | `internal/modules/auth/application/service.go:368-399` |
| 7 | Minor | Middleware / HTTP kernel | audit handleExport 405 missing Allow header | `internal/modules/audit/delivery/http/handler.go:134-136` |
| 8 | Minor | Middleware / HTTP kernel | audit handleExportSubresource 405 missing Allow header | `internal/modules/audit/delivery/http/handler.go:225-227` |
| 9 | Minor | Middleware / HTTP kernel | audit handleEvents 405 uses raw header.Set instead of canonical WriteMethodNotAllowed helper | `internal/modules/audit/delivery/http/handler.go:80-83` |
| 10 | Minor | Persistence / transactions | CodeExists off-tx inside lock-holding CD-create transaction | `internal/modules/controlleddocuments/application/service.go:373-378` |
| 11 | Minor | Persistence / transactions | Standalone Create on ControlledDocumentRepository has no SeedTxIdentity — never called by service | `internal/modules/controlleddocuments/infrastructure/repository.go:333-351` |
| 12 | Minor | Code quality / Go idioms | Anonymous interface in struct field and constructor (valuesRead) | `internal/modules/documents/application/freeze_service.go:49-51, 75-77` |
| 13 | Minor | Code quality / Go idioms | Pin recomputes values hash already computed inside pinValidateAndHash | `internal/modules/documents/application/freeze_service.go:200-212` |
| 14 | Minor | Legacy / dead-code | DocgenVer / GrammarVer fields in documents.Dependencies are never set by the composition root | `internal/modules/documents/module.go:44-45` |
| 15 | Minor | Legacy / dead-code | Deprecated documents/application.New still used across ~15 test files | `internal/modules/documents/application/service.go:127-136` |
| 16 | Minor | Module boundaries / DDD | auth module reads iam_user_roles directly in GetUserTenants (H-G class residual) | `internal/modules/auth/infrastructure/postgres/repository.go:104` |
| 17 | Minor | Contract / API layer | audit/handler.go: handleEvents emits map[string]any for primary list response | `internal/modules/audit/delivery/http/handler.go:120, 127` |
| 18 | Minor | Contract / API layer | audit/handler.go: handleExport and handleExportSubresource emit untyped maps | `internal/modules/audit/delivery/http/handler.go:216, 268` |
| 19 | Minor | Contract / API layer | auth/handler.go: handleLogin emits map[string]any for the login response | `internal/modules/auth/delivery/http/handler.go:90` |
| 20 | Minor | Contract / API layer | documents/fillin_handler.go: fillin schema and placeholder update emit map[string]any | `internal/modules/documents/delivery/http/fillin_handler.go:58, 116` |
| 21 | Minor | Contract / API layer | documents/view_handler.go: view URL response emits map[string]any | `internal/modules/documents/delivery/http/view_handler.go:46` |
| 22 | Minor | Contract / API layer | templates/routes_create.go: toTemplateResponse/toVersionResponse helpers not cleaned up post-M1 | `internal/modules/templates/delivery/http/routes_create.go:44` |
| 23 | Minor | Composition / observability | App-level OTel spans limited to 2 of ~5 critical async flows | `internal/modules/controlleddocuments/application/service.go:152; internal/modules/documents/approval/application/decision_service.go:155` |

---

## 8. Terminal Acceptance Verdict

**VERDICT: FAIL**

The M0–M4 program delivered measurable progress — seven of ten dimensions improved or held at B+ or better, with sessions, middleware, persistence, code quality, legacy, and composition all reaching A-. However, the §6 pass-bar requires all four checks to pass simultaneously, and all four fail at this audit: the contract/API layer remains at B+ (not A-) because map[string]any response emission persists across templates query/lifecycle, auth login, IAM admin overview, audit events, and document fillin/view routes; H-D stands at 2 confirmed sites in routes_generated.go; H-G stands at 2 confirmed sites (auth→iam_user_roles direct read; hardcoded "published" literal in templates infrastructure); and 4 skeptic-confirmed Major defects survive — the authz.Require effective_to predicate incorrectly denies time-bounded active memberships, the iam_users upsert silently discards tenant_id, and two Major contract violations in templates and IAM admin. The next milestone must close all four confirmed Majors, eliminate both H-D sites in routes_generated.go, port the auth GetUserTenants query behind a UserTenantReader IAM port, replace the hardcoded "published" literal in template_version_reader.go, and drive the contract/API layer to A- before re-audit.
