# Feature F0.3 — Spec

> **Milestone:** 0 — Auth / authz / session correctness  ·  **Folder:** `f0.3-tenant-grade-view`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / leandrotca.work (root-cause family fix — both sibling sites; align to declared `ScopeTenant` for `CapDocumentView`) — *implementation may begin.*

> This is the feature's **contract**, written and approved **before any code**.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Mission §7 F0.3 cites only `read_service.go:68` (LoadInstance), but the identical bug exists at `read_service.go:114` (LoadActiveInstanceByDocument). Fix one or both? | **Both.** Root-cause family fix, matches F0.1 "shared predicate, not per-caller" discipline. The acceptance criterion as written ("tenant-role-only viewer can read a document carrying a real area code") fails for the document-keyed consumer too if line 114 is left. Fixing only the cited line would be a symptom patch (HS-2 risk). Operator confirmed: "Both with root-cause fix, professional SaaS / industry standards." |
| 2 | Authoritative source for the "tenant-grade" classification — interpretation, or a declared property? | Declared. `iam/domain/capability_scope.go:51` maps `CapDocumentView → ScopeTenant` (ADR 0022 Phase 2). The fix realigns the two divergent call sites with the declared scope; it does **not** redefine scope. |
| 3 | Canonical idiom for a tenant-grade `Require` in this codebase? | `internal/modules/documents/application/view_service.go:71` — `authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant")` unconditionally, with the "tenant-grade read" comment. The two approval-side sites diverged from this idiom (they resolved a real area and only coalesced empty → "tenant"). Fix aligns them byte-for-byte with the canonical sibling. |
| 4 | Should `loadInstanceAreaCode` / the `docapp.LoadDocumentAreaCode` calls be removed? | `loadInstanceAreaCode` stays — its `found` flag is the existence probe before `repo.LoadInstance`. Its returned `areaCode` becomes unused for authz; mark with a one-line comment so the next reader knows the discard is deliberate. `docapp.LoadDocumentAreaCode` in `LoadActiveInstanceByDocument` is **removed** — it has no other consumer at that site, and repo lookup already short-circuits a missing doc/instance. Net: no resolver renames, no helper churn outside this file. |
| 5 | Does this redesign the shared authz model (HS-2)? | No. `authz.Require`'s `areaCode == "tenant"` sentinel branch (`authz.go:126`) is the existing, documented "area filter OFF" path. We use the public contract as-is. |
| 6 | H-PRE-1 advisory-lock hazard? | Not applicable — these are read paths with no authz-recording write moved across the tx boundary. The diff strictly **removes** SQL from the in-tx prologue of `LoadActiveInstanceByDocument`; nothing is added. |

## Consumer contract (FIRST — before any producer)

- **Consumer(s):**
  - `ReadService.LoadInstance` — called from approval HTTP handlers that render the per-instance approval detail screen (instance-keyed inbox detail).
  - `ReadService.LoadActiveInstanceByDocument` — called from document-detail HTTP handlers that surface the active approval instance on a document page.
  - Both consumers carry an authenticated actor (tier-1 already enforced) whose intended capability is `document.view` (tenant-grade per `iam/domain/capability_scope.go:51`).
- **Contract:** when an actor holds `document.view` **anywhere in the tenant** (any role in `role_capabilities` mapping to `document.view`, in any `user_process_areas` row that is current — i.e. `effective_from <= now()` and `effective_to IS NULL`, per F0.1) the call **succeeds** regardless of the target document's `process_area_code`. A system_admin always succeeds (existing bypass, unchanged). An actor with **no** `document.view` grant anywhere in the tenant is **denied** with `authz.ErrCapDenied{AreaCode: "tenant"}`. Behavior is byte-identical to the canonical `documents/application/view_service.go:71` site.
- **Source of truth for the contract:**
  - Declared scope: [`internal/modules/iam/domain/capability_scope.go:51`](../../../../../internal/modules/iam/domain/capability_scope.go) (`CapDocumentView: ScopeTenant`).
  - Canonical idiom: [`internal/modules/documents/application/view_service.go:70-71`](../../../../../internal/modules/documents/application/view_service.go).
  - `Require` tenant-sentinel branch: [`internal/modules/iam/authz/authz.go:74-76, 126`](../../../../../internal/modules/iam/authz/authz.go).

## What this feature implements

In `internal/modules/documents/approval/application/read_service.go`:

1. **`LoadInstance` (lines 45-90).** Keep `loadInstanceAreaCode` for its `found` existence probe. Drop the `if areaCode == "" { areaCode = "tenant" }` coalesce. Replace the `authz.Require(ctx, tx, CapDocumentView, areaCode)` call with `authz.Require(ctx, tx, string(iamdomain.CapDocumentView), "tenant")`. Add a one-line comment matching the `view_service.go:70` idiom; drop the now-stale "COALESCE / Phase 11 F7" comment block.
2. **`LoadActiveInstanceByDocument` (lines 93-136).** Delete the `docapp.LoadDocumentAreaCode` call and the `areaCode`/`found` locals — nothing else at this site consumes them. Drop the coalesce. Replace `authz.Require(... areaCode)` with `authz.Require(... "tenant")`. Repo lookup (`repo.LoadActiveInstanceByDocument`) already returns nil for a missing instance; tenant-grade authz is the only gate needed here.
3. **No other change.** `LoadActiveInstanceByDocumentForMutation`, the `*Inbox`/`*Pending` list/count methods, the `loadInstanceAreaCode` helper definition, and every other file under `internal/modules/documents/approval/` stay byte-identical.

## Non-goals (mandatory)

- **Not** changing `authz.Require`, the `"tenant"` sentinel semantics, or `capability_scope.go` (HS-2: shared authz redesign is out of scope; the public contract suffices).
- **Not** modifying `loadInstanceAreaCode` body, signature, or its `found` semantics (single-purpose probe stays; renaming is churn outside this fix).
- **Not** touching area-grade approval call sites (`decision_service.go`, `publish_service.go`, `submit_service.go`, `supersede_service.go`, `cancel_service.go`) — those correctly resolve a real area for area-grade caps.
- **Not** touching the canonical `documents/application/view_service.go` — it is already correct and serves as the reference.
- **Not** changing repo signatures, HTTP handlers, or any route.
- **No** schema / migration change.
- **Not** introducing a `Require`-side scope-vs-passed-area assert (worth doing program-wide, but outside this milestone — flagged as a defer below).

## Validation Gate (concrete — approved before code)

| # | Acceptance criterion | Named test / proof command | Real vs fixture |
|---|----------------------|----------------------------|-----------------|
| 1 | `LoadInstance` — actor holds `document.view` (any tenant role) but **no `user_process_areas` row in the document's area** → call **succeeds** when the instance's resolved area code is a real area (e.g. `qa`). | `TestLoadInstance_TenantGradeViewer_DocWithAreaCode_Granted` (new, `//go:build integration`, real Postgres) — RED on pre-fix code (denied with `ErrCapDenied{AreaCode:"qa"}`), GREEN on post-fix (nil error, instance returned). | **real** |
| 2 | `LoadActiveInstanceByDocument` — same actor, same doc with area `qa` → call **succeeds**. | `TestLoadActiveInstanceByDocument_TenantGradeViewer_DocWithAreaCode_Granted` (new, same integration file) — RED pre-fix, GREEN post-fix. | **real** |
| 3 | `LoadInstance` — actor with **no `document.view` grant at all** → call **denied** with `ErrCapDenied{AreaCode:"tenant"}` (deny still works; deny envelope changes from `qa` → `tenant` per the fix). | `TestLoadInstance_NoViewGrant_Denied` (new, same integration file). | **real** |
| 4 | `LoadActiveInstanceByDocument` — same no-grant actor → denied with `ErrCapDenied{AreaCode:"tenant"}`. | `TestLoadActiveInstanceByDocument_NoViewGrant_Denied` (new, same integration file). | **real** |
| 5 | system_admin bypass still works on both methods (no regression). | `TestLoadInstance_SystemAdmin_Granted` + `TestLoadActiveInstanceByDocument_SystemAdmin_Granted` (new, same file). | **real** |
| 6 | Existing `read_service_test.go` cases stay green — including the `RequiresDocumentViewBeforeRepoLoad` sqlmock case, retargeted to the new `"tenant"` sentinel (the test currently asserts deny via `qa` area filter; post-fix it asserts deny via `"tenant"` sentinel — same security property, new envelope). | `go test ./internal/modules/documents/approval/application/...` no `FAIL`. | fixture |
| 7 | No authz regression across module. | `go test -tags integration -count=1 ./internal/modules/documents/approval/...` no `FAIL` (modulo the pre-existing `TestSequenceAllocatorNextAndIncrement_Concurrent` defer from F0.2 — re-verified unrelated, reproduced on pre-F0.3 HEAD). | **real** |
| 8 | Whole-repo green. | `go build ./...` clean; `go test ./...` no `FAIL`. | mixed |

> TDD: integration tests written first. Criteria 1–2 MUST fail on pre-fix code (proving B3 reproduces with `ErrCapDenied{AreaCode:"<real>"}`). Then apply the alignment patch; all six new integration tests must pass.

## ADR needed?

- [x] No durable decision — skip. ADR 0022 Phase 2 (`capability_scope.go`) already declares `CapDocumentView → ScopeTenant`; this feature aligns two divergent runtime call sites with that declared scope. Recorded in this feature's spec/evidence and in the commit message.

## Bounded defers (recorded for milestone-validator visibility)

| Defer | Why bounded here | Trigger |
|-------|------------------|---------|
| `authz.Require` does not assert `ScopeOf(cap) == ScopeTenant ⇒ areaCode == "tenant"` (would catch this class of bug at the shared layer, statically/dynamically). | Shared authz-API surface change → HS-2; deserves its own ADR and a program-wide audit of every `Require` call site. Out of M0 scope (M0 is the 4 named correctness defects). | Trigger: M1 (contract-tail) or a dedicated authz-hardening micro-milestone. Owner: grade-a-completion operator. |
| Rename `loadInstanceAreaCode` → `instanceExists` (its only remaining use is the `found` probe). | Cosmetic refactor — touches one helper and one caller, no behavior change, no security property. CLAUDE.md §5.3 ("touch only what you must"). | Trigger: next planned edit of `read_service.go`. |
