# Plan 04 · Capability namespace collapse + IAM dual-surface consolidation

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:executing-plans` to implement task-by-task. Steps use checkbox (`- [ ]`) syntax.

**Goal:** Collapse the dual capability namespaces (`iam/domain/capabilities.go` strings vs `iam/domain/model.go` typed `Capability`) into one typed namespace; pick one area-membership write surface and delete the other; delete the unwired `AuthorizationService`. Closes iam T-001/T-002/T-003/T-009/T-012, documents T-008.

**Architecture:**
- Typed `iamdomain.Capability` wins (anchor decision in `wiki/backlog/roadmap.md:17`). All 16 string consts in `capabilities.go` are recreated as typed `Capability` consts in `model.go`; `capabilities.go` is deleted; the DB seed is reseeded so values match `document.*` instead of `doc.*` for the 5 overlapping consts.
- Application-service path (`iam/application/area_membership_service.go` + `UserAreaRepository.GrantAtomic`) keeps the area-membership write surface — composes with Plan 6 (`govLogger`). Delete the SECURITY DEFINER Go wrapper package `internal/modules/iam/area_membership/`. SQL functions `metaldocs.grant_area_membership` / `revoke_area_membership` STAY (still called by `internal/test/e2e_seed.go` and integration tests via raw SQL).
- `AuthorizationService` is deleted (Plan 5 will wire tier-2 `authz.Require` per module — third surface unnecessary). The orphan `domain.RoleCapabilities` map and `application.CheckRoleCapabilitiesVersion` (only consumer was `AuthorizationService` + integration test) go with it.
- `ErrCapabilityDenied` symbol collision resolved by renaming `authz.ErrCapabilityDenied` → `authz.ErrCapDenied` (struct kept, sentinel `iamapp.ErrCapabilityDenied` unchanged — used by docs handler).

**Tech Stack:** Go 1.22+ (`metaldocs/internal/modules/iam`), Postgres (numbered SQL migrations under `migrations/`), PowerShell start script for local API, `go test ./...`.

**Plan 3 dependency note (added 2026-05-11 after Plan 3 merge):**
- Plan 3 (commit `23b76a4b` + ancestors) migrated tenant resolution from `X-Tenant-ID` header to `tenant.FromContext(r.Context())` across `iam/delivery/http/{admin_handler,middleware,routes_memberships}.go`. Plan 4 does NOT touch tenant-resolution code — those files are unrelated to capability namespace work. Wiki anchors in `iam.md` for those files may have shifted line numbers; Task 10 must re-grep before editing.
- Plan 3 consumed migration number `0186` (`0185_revoke_ambiguous_sessions.sql`). Plan 4's typed-cap reseed therefore uses `0186_role_capabilities_typed.sql` (renumbered below).
- Plan 3 added `wiki/architecture/tenant-context.md` and refreshed `wiki/architecture/system-overview.md`. No conflict with Plan 4.
- `tenantIDFromRequest` now returns `(string, error)`. Not on Plan 4's edit path.
- Smoke test (Task 9): admin login still works via `POST /api/v1/auth/login`. Session now carries `tenant_id`; subsequent capability-gated requests pull tenant from session context, so the gating still exercises the renamed `metaldocs.role_capabilities` rows.

**Out of scope (push back if asked to expand):**
- Tier-2 `authz.Require` additions on IAM-owned tables → Plan 5.
- Tripwire trigger attach on IAM tables → Plan 5.
- RFC 9457 envelope migration in IAM handlers → Plan 7.
- Audit emission on `handleUserRoleUpsert` (T-005) → Plan 6.
- Renaming non-collision string literals (`doc.edit_draft`, `doc.publish`, `doc.obsolete`, `doc.supersede`, `doc.reconstruct`, `doc.view_published`, `workflow.instance.cancel`, `route.admin`) — these are single-source and not part of the dual-namespace conflict. Touch ONLY values that exist in `capabilities.go` (the 16 consts).
- Full ADR documents for new decisions — `// ADR-TODO:` one-liner stubs only; full ADRs belong to Plan 13.

---

## File Map

**Modified:**
- `internal/modules/iam/domain/model.go` — extend with the 11 additional typed `Capability` consts; rename three values (`doc.view`→`document.view`, `doc.create`→`document.create`, `doc.edit`→`document.edit`) so they match the existing `CapDocumentView/Create/Edit` consts; add `CapDocumentSubmit`/`CapDocumentSignoff` (replaces `CapDocSubmit`/`CapDocSignoff`).
- `apps/api/cmd/metaldocs-api/permissions.go` — rename consumer references: `CapDocView`→`CapDocumentView`, `CapDocCreate`→`CapDocumentCreate`, `CapDocEdit`→`CapDocumentEdit`, `CapDocSubmit`→`CapDocumentSubmit`, `CapDocSignoff`→`CapDocumentSignoff`. Other `CapTemplate*`/`CapRegistryCreate`/`CapTaxonomyManage`/`CapMembershipManage`/`CapUserManage` keep their names.
- `apps/api/cmd/metaldocs-api/permissions_test.go` — same renames.
- `internal/modules/documents/approval/application/submit_service.go:85` — `"doc.submit"` → `string(iamdomain.CapDocumentSubmit)` (add import).
- `internal/modules/iam/authz/authz.go` — rename type `ErrCapabilityDenied` → `ErrCapDenied`.
- `internal/modules/iam/authz/authz_test.go` — rename usages.
- `internal/modules/documents/approval/http/{submit,publish,obsolete,supersede,cancel,route_admin,errors}_handler*.go` and `*_test.go` — `authz.ErrCapabilityDenied` → `authz.ErrCapDenied` (mechanical).
- `internal/modules/documents/approval/application/{submit,publish,obsolete,supersede,cancel,decision,route_admin}_service_test.go` — same.
- `internal/modules/documents/http/{view,reconstruct,fillin,rbac}_handler*.go` and tests — same.
- `internal/modules/iam/integration_test.go` — same.
- `internal/modules/iam/application/area_membership_service.go:55` — replace `domain.RoleCapabilities[role]` known-role check with explicit role switch.
- `apps/api/cmd/metaldocs-api/main.go:217` — wiring stays; just uses the surviving `AreaMembershipService`. Remove any `CheckRoleCapabilitiesVersion` boot call if present (grep shows none today, but verify).
- `wiki/modules/iam-tech-debt.md` — flip T-001/T-002/T-003/T-009/T-012 status to `closed` with PR link; drop `RoleCapabilities` dual-source row.
- `wiki/backlog/iam-refactor.md` — flip R-001/R-002/R-003/R-009/R-012 to `merged` with PR.
- `wiki/modules/documents-tech-debt.md` — close T-008.
- `wiki/backlog/documents-refactor.md` — close R-008.
- `wiki/backlog/roadmap.md` — Plan 4 status → `done YYYY-MM-DD`, link PRs, bump `Last verified`.
- `wiki/modules/iam.md` — drop `area_membership/` row from §5 file table; bump `Last verified`.

**Deleted:**
- `internal/modules/iam/domain/capabilities.go`
- `internal/modules/iam/area_membership/area_membership.go` (and `area_membership_test.go`).
- `internal/modules/iam/application/authorization.go` (and `authorization_test.go`, `authorization_bench_test.go`).
- `internal/modules/iam/application/startup.go` (only contains `CheckRoleCapabilitiesVersion`).
- `internal/modules/iam/domain/role_capabilities.go` (the map + `RoleCapabilitiesVersion`).
- `internal/modules/iam/domain/role_capabilities_test.go`.
- `internal/modules/iam/domain/role_capabilities_integration_test.go`.

**Created:**
- `migrations/0186_role_capabilities_typed.sql` — `UPDATE metaldocs.role_capabilities SET capability = 'document.X' WHERE capability = 'doc.X'` for the five renamed values across both `0165` and `0169` seed sets.
- `migrations/0187_drop_iam_area_membership_wrapper.sql` — **NOT created**. Sentinel comment only — see Task 9 note. The DB-side SECURITY DEFINER funcs stay (still called by e2e seed + integration tests).

---

## Tasks

### Task 1: Add typed `Capability` consts for the 11 missing values + rename 5 overlapping ones

**Files:**
- Modify: `internal/modules/iam/domain/model.go`

- [ ] **Step 1: Edit `model.go` to the new canonical const set**

```go
package domain

type Role string

const (
	RoleApprover    Role = "approver"
	RoleAuthor      Role = "author"
	RoleEditor      Role = "editor"
	RoleSystemAdmin Role = "system_admin"
	RoleViewer      Role = "viewer"
)

type Capability string

const (
	CapDocumentView    Capability = "document.view"
	CapDocumentCreate  Capability = "document.create"
	CapDocumentEdit    Capability = "document.edit"
	CapDocumentSubmit  Capability = "document.submit"
	CapDocumentSignoff Capability = "document.signoff"
	CapWorkflowReview  Capability = "workflow.review"
	CapWorkflowApprove Capability = "workflow.approve"

	CapTemplateView    Capability = "template.view"
	CapTemplateCreate  Capability = "template.create"
	CapTemplateEdit    Capability = "template.edit"
	CapTemplateSubmit  Capability = "template.submit"
	CapTemplateApprove Capability = "template.approve"
	CapTemplatePublish Capability = "template.publish"

	CapRegistryCreate   Capability = "registry.create"
	CapTaxonomyManage   Capability = "taxonomy.manage"
	CapMembershipManage Capability = "membership.manage"
	CapRouteManage      Capability = "route.manage"
	CapUserManage       Capability = "user.manage"
)
```

- [ ] **Step 2: Delete `internal/modules/iam/domain/capabilities.go`**

```bash
rm internal/modules/iam/domain/capabilities.go
```

- [ ] **Step 3: Compile-only check**

Run: `go build ./internal/modules/iam/...`
Expected: FAIL — consumers of removed names (`CapDocView`, `CapDocCreate`, `CapDocEdit`, `CapDocSubmit`, `CapDocSignoff`) cannot resolve. Other names (`CapTemplate*`, `CapRegistryCreate`, etc.) compile because they kept their names.

- [ ] **Step 4: Commit (broken build expected, but the namespace skeleton is in place)**

```bash
git add internal/modules/iam/domain/model.go internal/modules/iam/domain/capabilities.go
git commit -m "refactor(iam): collapse capability consts into typed Capability namespace"
```

---

### Task 2: Rename consumer call sites for the 5 renamed Doc* consts

**Files:**
- Modify: `apps/api/cmd/metaldocs-api/permissions.go`
- Modify: `apps/api/cmd/metaldocs-api/permissions_test.go`
- Modify: `internal/modules/documents/approval/application/submit_service.go:85`

- [ ] **Step 1: `permissions.go` — replace 5 names**

In `apps/api/cmd/metaldocs-api/permissions.go`, replace every occurrence:

| Old | New |
|---|---|
| `iamdomain.CapDocView` | `iamdomain.CapDocumentView` |
| `iamdomain.CapDocCreate` | `iamdomain.CapDocumentCreate` |
| `iamdomain.CapDocEdit` | `iamdomain.CapDocumentEdit` |
| `iamdomain.CapDocSubmit` | `iamdomain.CapDocumentSubmit` |
| `iamdomain.CapDocSignoff` | `iamdomain.CapDocumentSignoff` |

Use `replace_all: true` in the Edit tool with each pair to avoid uniqueness errors. Lines hit per the grep result: `:28, :31, :34, :40, :43, :72, :80, :88, :123, :125, :127, :129, :133, :135, :137, :139, :141, :143, :145, :147, :149, :151, :153, :155, :161, :169, :177, :185, :189, :191, :193, :205, :207`.

- [ ] **Step 2: `permissions_test.go` — same 5 renames**

Same `replace_all` mapping. Test rows hit: `:27, :39, :40, :44`.

- [ ] **Step 3: `submit_service.go:85` — replace string literal with typed const**

Read context first:

```go
if err := authz.Require(ctx, tx, "doc.submit", areaCode); err != nil {
```

becomes:

```go
if err := authz.Require(ctx, tx, string(iamdomain.CapDocumentSubmit), areaCode); err != nil {
```

Add the import `iamdomain "metaldocs/internal/modules/iam/domain"` to the file's import block if not present.

- [ ] **Step 4: Build check**

Run: `go build ./...`
Expected: Likely FAIL on remaining issues (Workstreams C/D not yet done). Document errors — they should be limited to: (a) `RoleCapabilities` map references in `area_membership_service.go:55`, `authorization.go:127`, `role_capabilities*test.go`; (b) `authz.ErrCapabilityDenied` consumer mismatches if any (none expected since type still exists). Anything else means a missed rename.

- [ ] **Step 5: Commit**

```bash
git add apps/api/cmd/metaldocs-api/permissions.go apps/api/cmd/metaldocs-api/permissions_test.go internal/modules/documents/approval/application/submit_service.go
git commit -m "refactor(iam,documents): retarget consumers to typed CapDocument* consts"
```

---

### Task 3: Reseed `metaldocs.role_capabilities` to use renamed values

**Files:**
- Create: `migrations/0186_role_capabilities_typed.sql`

- [ ] **Step 1: Author the migration**

```sql
-- migrations/0186_role_capabilities_typed.sql
-- Plan 4 (2026-05-11): collapse capability namespace.
-- The five doc.* values created by 0165 + 0169 are renamed to document.*
-- so they match the typed iamdomain.Capability consts in
-- internal/modules/iam/domain/model.go. Idempotent: ON CONFLICT-safe via
-- the existing UNIQUE (role, capability).

BEGIN;

UPDATE metaldocs.role_capabilities SET capability = 'document.view'    WHERE capability = 'doc.view';
UPDATE metaldocs.role_capabilities SET capability = 'document.create'  WHERE capability = 'doc.create';
UPDATE metaldocs.role_capabilities SET capability = 'document.edit'    WHERE capability = 'doc.edit';
UPDATE metaldocs.role_capabilities SET capability = 'document.submit'  WHERE capability = 'doc.submit';
UPDATE metaldocs.role_capabilities SET capability = 'document.signoff' WHERE capability = 'doc.signoff';

INSERT INTO public.schema_migrations (version, description)
VALUES ('0186', 'Plan 4: rename doc.* role_capabilities rows to document.*')
ON CONFLICT (version) DO NOTHING;

COMMIT;
```

- [ ] **Step 2: Re-run migrations against local DB**

Run: `.\scripts\start-api.ps1` (the script applies pending migrations on boot).
Expected: log line shows `0186_role_capabilities_typed.sql applied`.

If start-api.ps1 fails on Workstream C (Task 10) deletions still pending, defer this verification step until Task 14.

- [ ] **Step 3: Manual SQL spot check (in another terminal while API is up)**

```sql
SELECT capability, COUNT(*) FROM metaldocs.role_capabilities
WHERE capability LIKE 'doc.%' GROUP BY capability;
-- expect 0 rows
SELECT capability, COUNT(*) FROM metaldocs.role_capabilities
WHERE capability LIKE 'document.%' GROUP BY capability;
-- expect 5 rows: document.view, .create, .edit, .submit, .signoff
```

- [ ] **Step 4: Commit**

```bash
git add migrations/0186_role_capabilities_typed.sql
git commit -m "feat(migration): 0186 rename doc.* role_capabilities to document.*"
```

---

### Task 4: Delete `AuthorizationService` (Workstream C)

**Files:**
- Delete: `internal/modules/iam/application/authorization.go`
- Delete: `internal/modules/iam/application/authorization_test.go`
- Delete: `internal/modules/iam/application/authorization_bench_test.go`

- [ ] **Step 1: Confirm zero production wiring**

Run: `Grep "NewAuthorizationService" --type go --glob "!*_test.go" --glob "!authorization*.go"`
Expected: NO matches (already verified — only the constructor itself + test files reference it).

- [ ] **Step 2: Delete the three files**

```bash
rm internal/modules/iam/application/authorization.go internal/modules/iam/application/authorization_test.go internal/modules/iam/application/authorization_bench_test.go
```

- [ ] **Step 3: Build check**

Run: `go build ./internal/modules/iam/...`
Expected: still FAILS on `domain.RoleCapabilities` consumers in `area_membership_service.go:55` and (now-deleted-?) `role_capabilities*test.go`. AuthorizationService gone — no errors related to its symbols.

- [ ] **Step 4: Commit**

```bash
git add -A internal/modules/iam/application
git commit -m "refactor(iam): delete unwired AuthorizationService (third authz surface)

// ADR-TODO: rationale — Plan 5 wires tier-2 authz.Require per module; a third
// resource-aware surface adds zero coverage and was never composed."
```

---

### Task 5: Delete `RoleCapabilities` map + `RoleCapabilitiesVersion` + `CheckRoleCapabilitiesVersion`

**Files:**
- Delete: `internal/modules/iam/domain/role_capabilities.go`
- Delete: `internal/modules/iam/domain/role_capabilities_test.go`
- Delete: `internal/modules/iam/domain/role_capabilities_integration_test.go`
- Delete: `internal/modules/iam/application/startup.go`
- Modify: `internal/modules/iam/application/area_membership_service.go:55`

- [ ] **Step 1: Replace the `RoleCapabilities` known-role check in `area_membership_service.go`**

Current (`area_membership_service.go:54-57`):

```go
func (s *AreaMembershipService) Grant(...) error {
	if _, ok := domain.RoleCapabilities[role]; !ok {
		return ErrUnknownRole
	}
```

Replace with:

```go
func (s *AreaMembershipService) Grant(...) error {
	switch role {
	case domain.RoleApprover, domain.RoleAuthor, domain.RoleEditor, domain.RoleSystemAdmin, domain.RoleViewer:
		// known
	default:
		return ErrUnknownRole
	}
```

- [ ] **Step 2: Delete the four files**

```bash
rm internal/modules/iam/domain/role_capabilities.go \
   internal/modules/iam/domain/role_capabilities_test.go \
   internal/modules/iam/domain/role_capabilities_integration_test.go \
   internal/modules/iam/application/startup.go
```

- [ ] **Step 3: Confirm `CheckRoleCapabilitiesVersion` unwired**

Run: `Grep "CheckRoleCapabilitiesVersion" --type go`
Expected: NO matches (verified — not called from `apps/api/cmd/metaldocs-api/main.go`).

- [ ] **Step 4: Build check**

Run: `go build ./...`
Expected: PASS for `internal/modules/iam/...`. Remaining failures (if any) should only be in `authz.ErrCapabilityDenied` consumers — addressed in Task 6.

- [ ] **Step 5: Commit**

```bash
git add -A internal/modules/iam
git commit -m "refactor(iam): delete RoleCapabilities map + version check (orphaned by AuthorizationService deletion)

// ADR-TODO: rationale — DB role_capabilities table is the single source of
// truth (CapabilityService.CanDo reads it). The in-process map only fed
// AuthorizationService (deleted in prior commit). Known-role validation
// moves to an explicit Role switch in AreaMembershipService.Grant."
```

---

### Task 6: Rename `authz.ErrCapabilityDenied` → `authz.ErrCapDenied`

**Files:**
- Modify: `internal/modules/iam/authz/authz.go` (`ErrCapabilityDenied` struct + constructor)
- Modify: `internal/modules/iam/authz/authz_test.go`
- Modify: `internal/modules/iam/integration_test.go`
- Modify (mechanical replace): every test/handler file matching the grep result for `authz.ErrCapabilityDenied` — see file list in the File Map.

- [ ] **Step 1: Rename in `authz.go`**

Edit `internal/modules/iam/authz/authz.go`: replace `ErrCapabilityDenied` with `ErrCapDenied` (3 occurrences: type decl, method receiver, return-value construction).

- [ ] **Step 2: Bulk-rename consumers**

For each file in this list, replace `authz.ErrCapabilityDenied` → `authz.ErrCapDenied` (use `replace_all: true`):

```
internal/modules/iam/authz/authz_test.go
internal/modules/iam/integration_test.go
internal/modules/documents/approval/http/supersede_handler_test.go
internal/modules/documents/approval/http/submit_handler_test.go
internal/modules/documents/approval/http/route_admin_handler_test.go
internal/modules/documents/approval/http/publish_handler_test.go
internal/modules/documents/approval/http/obsolete_handler_test.go
internal/modules/documents/approval/http/cancel_handler_test.go
internal/modules/documents/approval/http/errors_test.go
internal/modules/documents/approval/http/errors.go
internal/modules/documents/approval/application/supersede_service_test.go
internal/modules/documents/approval/application/decision_service_test.go
internal/modules/documents/approval/application/submit_service_test.go
internal/modules/documents/approval/application/route_admin_service_test.go
internal/modules/documents/approval/application/cancel_service_test.go
internal/modules/documents/approval/application/publish_service_test.go
internal/modules/documents/approval/application/obsolete_service_test.go
internal/modules/documents/http/view_handler_test.go
internal/modules/documents/http/view_handler.go
internal/modules/documents/http/reconstruct_handler_test.go
internal/modules/documents/http/reconstruct_handler.go
internal/modules/documents/http/rbac_test.go
internal/modules/documents/http/fillin_handler_test.go
internal/modules/documents/http/fillin_handler.go
```

Note: the `iamapp.ErrCapabilityDenied` sentinel in `documents/delivery/http/handler.go:980` and `service_caps_test.go:23,51` is the SEPARATE one in `iam/application/capability_service.go` — leave it untouched.

- [ ] **Step 3: Build + vet**

Run: `go build ./...; go vet ./...`
Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "refactor(iam,documents): rename authz.ErrCapabilityDenied to ErrCapDenied (resolves dual-symbol collision with iamapp.ErrCapabilityDenied)"
```

---

### Task 7: Run unit tests; fix any breakage caused by Workstreams A+C+D

- [ ] **Step 1: Run the IAM + documents unit suites**

Run: `go test ./internal/modules/iam/... ./internal/modules/documents/... ./apps/api/...`
Expected: PASS.

If a test fails because it references a deleted symbol (`RoleCapabilities`, `AuthorizationService`, `CapDocView`, etc.), fix the reference. Do NOT write new tests for deleted code paths.

- [ ] **Step 2: Run full unit suite**

Run: `go test ./...`
Expected: PASS (modulo pre-existing flakes unrelated to this plan).

- [ ] **Step 3: Commit any test fixes**

```bash
git add -A
git commit -m "test(iam): repair tests broken by capability namespace collapse"
```

(Skip if Step 1+2 produced no further changes.)

---

### Task 8: Delete `iam/area_membership/` Go package (Workstream B)

**Files:**
- Delete: `internal/modules/iam/area_membership/area_membership.go`
- Delete: `internal/modules/iam/area_membership/area_membership_test.go`
- Modify: `tools/cilint/internal/analyzers/analyzers.go:25` — remove the `"iam/area_membership"` arch-boundary entry.

- [ ] **Step 1: Confirm zero non-test Go consumers**

Run: `Grep "iam/area_membership" --type go --glob "!**/area_membership*"`
Expected: only `tools/cilint/internal/analyzers/analyzers.go:25` (boundary list entry — gets deleted with the package).

- [ ] **Step 2: Delete the package**

```bash
rm -r internal/modules/iam/area_membership
```

- [ ] **Step 3: Remove the cilint boundary entry**

Edit `tools/cilint/internal/analyzers/analyzers.go`: delete the line `"iam/area_membership",` from the slice.

- [ ] **Step 4: Build + test check**

Run: `go build ./...; go test ./tools/cilint/...`
Expected: PASS.

The SECURITY DEFINER SQL functions `metaldocs.grant_area_membership` / `metaldocs.revoke_area_membership` STAY (still called via raw SQL by `internal/test/e2e_seed.go:489` and `tests/integration/scenarios/membership_fn_test.go`). No DB migration in this task.

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "refactor(iam): delete area_membership/ Go wrapper (zero non-test consumers)

// ADR-TODO: rationale — application-service path
// (iam/application/area_membership_service.go) is the canonical write surface.
// SECURITY DEFINER SQL funcs stay; only the Go wrapper that competed with the
// service is removed. Plan 6 will add MembershipGovernanceLogger to make the
// service emit governance_events at parity with the SQL function path."
```

---

### Task 9: Local API smoke check — capability-gated route still works

**Files:** none modified.

- [ ] **Step 1: Start API**

Run: `.\scripts\start-api.ps1 -Build`
Expected: server up on `:8081`, all migrations including `0186` applied, no panic at boot.

- [ ] **Step 2: Login**

```bash
curl -sS -X POST http://localhost:8081/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"identifier":"admin","password":"AdminMetalDocs123!"}' \
  -c cookies.txt
```
Expected: HTTP 200 with session cookie.

- [ ] **Step 3: Exercise a `CapUserManage`-gated route (positive case)**

```bash
curl -sS -X GET http://localhost:8081/api/v1/iam/users -b cookies.txt -i | head -1
```
Expected: `HTTP/1.1 200 OK` (admin holds `user.manage` via `system_admin`).

- [ ] **Step 4: Exercise a `CapDocumentSubmit`-gated route (positive case)**

Pick any draft document the seed user owns and POST a submit; OR (lighter) confirm the new value reaches the DB:

```sql
psql ... -c "SELECT 1 FROM metaldocs.role_capabilities WHERE role='author' AND capability='document.submit';"
```
Expected: 1 row.

- [ ] **Step 5: Stop API. No commit unless smoke surfaced a regression.**

---

### Task 10: Update wiki — close debt rows, refresh anchors

**Files:**
- Modify: `wiki/modules/iam-tech-debt.md` — set Status to `closed YYYY-MM-DD` on T-001, T-002, T-003, T-009, T-012; bump `Last verified`.
- Modify: `wiki/backlog/iam-refactor.md` — set R-001/R-002/R-003/R-009/R-012 Status to `merged` with PR link; bump `Last verified`.
- Modify: `wiki/modules/documents-tech-debt.md` — set T-008 Status to `closed YYYY-MM-DD`.
- Modify: `wiki/backlog/documents-refactor.md` — set R-008 Status to `merged`.
- Modify: `wiki/backlog/roadmap.md` — Plan 4 row Status to `done 2026-05-NN`; add `**PRs:**` line; bump top `Last verified`.
- Modify: `wiki/modules/iam.md` — drop `area_membership/` row from §5 file table; drop §11 T-001 surface line referencing `area_membership.go:53`; bump `Last verified`.

- [ ] **Step 1: Apply the doc edits one file at a time using Edit tool**

For each file above, locate the relevant row by its `T-NNN` / `R-NNN` / `## Plan 4` heading and update Status + `Last verified` stamp. Append `**PRs:** <links>` under the Plan 4 body in `roadmap.md`.

- [ ] **Step 2: Dispatch `wiki-curator` agent** (per CLAUDE.md guidance)

Use the Agent tool with `subagent_type: wiki-curator`. Prompt:

> Plan 4 (capability namespace collapse + IAM dual-surface consolidation) just merged. Refresh `Last verified` stamps and Key files anchors on every wiki doc that references `internal/modules/iam/domain/capabilities.go`, `internal/modules/iam/domain/model.go`, `internal/modules/iam/area_membership/`, `internal/modules/iam/application/authorization.go`, `internal/modules/iam/application/startup.go`, or `internal/modules/iam/domain/role_capabilities.go`. Update `wiki/README.md` index entries if file paths changed. Do not author new ADRs (Plan 13).

- [ ] **Step 3: Commit doc + wiki-curator changes**

```bash
git add wiki/
git commit -m "docs(wiki): mark Plan 4 debt rows closed + refresh anchors"
```

---

## Self-Review Checklist (run before declaring complete)

1. **Spec coverage:**
   - T-001 (dual namespaces) → Tasks 1-3, 7. ✓
   - T-002 (dual area-membership surfaces) → Task 8. ✓
   - T-003 (unwired AuthorizationService) → Task 4. ✓
   - T-009 (dual `ErrCapabilityDenied`) → Task 6. ✓
   - T-012 (`RoleCapabilities` map duplication) → Task 5. ✓
   - documents T-008 (capability namespace straddle) → Task 2 step 3. ✓

2. **Anchor decision compliance (`roadmap.md:17`):** Typed `iamdomain.Capability` wins. ✓ DB seed regenerated (Task 3). ✓ Consumers fanned out (Task 2). ✓

3. **Scope locks honored:**
   - No tier-2 `authz.Require` additions (Plan 5). ✓
   - No tripwire trigger attach (Plan 5). ✓
   - No RFC 9457 envelope (Plan 7). ✓
   - No audit emission additions (Plan 6). ✓
   - Non-collision string literals (`doc.edit_draft`, etc.) untouched. ✓

4. **No placeholders:** every step has the actual code/command. ✓

5. **Type consistency:** `CapDocumentSubmit` named identically in Tasks 1, 2, 3 (column value `document.submit`), 9 (smoke check). ✓
