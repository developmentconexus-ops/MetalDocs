# Feature F4.2 — Spec

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Folder:** `f4.2-template-version-state-reader`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-06-15 / operator (delegated the engineering call — "study it, reach the
> best solution by industry standards"; approach recorded in the interview record below + milestone.md
> HS-6 reconciliation).

> This is the feature's **contract**, written and approved **before any code**. It defines what the
> feature must do and how it will be proven — not how it will be built (that is `plan.md`). The
> milestone-validator judges the feature against *this* file (C1).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | F4.2 said "**introduce** `TemplateVersionStateReader`", but templates already owns `TemplateVersionPort.IsPublished` (Wave Z Z-7, consumed by taxonomy) running the identical `templates_*` JOIN. Its `(bool, docTypeCode)` shape is insufficient (CD needs raw status). Extend the existing port, or introduce a parallel new one? | Operator: *"we need to study it then so we reach the best solution based on standards, professional backend and how big companies operate and industry standard"* — i.e. delegated the engineering call. **Decision: extend the existing owning port.** Single-owning-adapter (one type owns the `templates_*` SQL + tenant-scoping); consumer-driven contracts (raw `GetTemplateVersionState` is the primitive, `IsPublished` a derived predicate kept for taxonomy). A second reader over the same tables would be the duplication anti-pattern this milestone kills. Recorded in `milestone.md` HS-6 reconciliation. |
| 2 | How should CD stop reaching `templates_*` (close `repository.go:702`)? | Operator chose **"Delete CD checker, wire templates reader"** — remove `controlleddocuments` `PostgresTemplateVersionChecker` entirely; wire the templates-owned reader into CD's module so `service.go:209/308` consume the port. CD's `application.TemplateVersionChecker` consumer interface (the `(*string,string,error)` shape) preserved exactly. |
| 3 | Does CD's `Resolve` need the raw status string, or is `published bool` enough? | Verified: raw status. `controlleddocuments/domain/resolution.go:42,55,58` distinguish `"published"` / `"obsolete"` / draft — `IsPublished` (bool) cannot express this. Confirms the existing port shape is insufficient and the raw-state primitive is required. |
| 4 | Does the `status := "published"` hardcode matter, or is it cosmetic? | It **masks a real bug.** `documents_adapters.go:113` feeds `DefaultTemplate.Status = "published"` into `resolveDefaultTemplate`, which rejects `"obsolete"` and non-`"published"`. With the hardcode, an *obsolete* (or any-status) profile-default template version wrongly resolves as publishable. Reading the real status via the port fixes this. (Behavior change is intended + recorded as a non-cosmetic correctness fix, not a contract break.) |

## Consumer contract (FIRST — before any producer)

Two existing consumers already define the exact shape; the producer (templates port) is built to match
them. **Read from the consumers; nothing invented.**

### Consumer 1 — CD override-template validation
- **Consumer:** `controlleddocuments/application.TemplateVersionChecker` (interface declared by the
  consumer), called at `service.go:209` (manual path, off-tx) and `service.go:308` (auto path).
- **Contract (unchanged):**
  `GetTemplateVersionState(ctx, tenantID, templateVersionID string) (status *string, docTypeCode string, err error)`
  - `status`: pointer to the raw template-version status string; `nil` when the version is not found
    **or** its status column is NULL (current `PostgresTemplateVersionChecker` semantics — preserve).
  - `docTypeCode`: the owning template's `doc_type_code` (used as `ProfileCode` in `Resolve`'s
    `TemplateVersionCandidate`); `""` when not found.
  - `err`: non-nil only on a real query error; **not-found is `(nil, "", nil)`** (preserve).
- **Source of truth:** the consumer interface `internal/modules/controlleddocuments/application/service.go:24-26`
  and its only producer today `internal/modules/controlleddocuments/infrastructure/repository.go:702-723`.

### Consumer 2 — documents-create profile-default resolver
- **Consumer:** `wiring.profileDefaultsAdapter` implementing `docapp.ProfileDefaultTemplateReader.GetDefaultTemplateVersionID(ctx, tenantID, profileCode) (*string, *string, error)`
  (`apps/api/internal/wiring/documents_adapters.go:105-115`).
- **Contract it must satisfy unchanged (toward *its* consumer):** returns
  `(defaultTemplateVersionID *string, status *string, err error)` — but the `status` it returns must be
  the **real** status of that template version, no longer the literal `"published"`.
- **What it needs from the F4.2 port:** the same `GetTemplateVersionState(ctx, tenantID, versionID)`
  raw-status read (it uses only the `status` return; `docTypeCode` ignored).
- **Source of truth:** `docapp.ProfileDefaultTemplateReader` interface + the hardcode at line 113.

### Producer (built to match)
- **Owning module:** templates. **Port (extended):** `templates/domain.TemplateVersionPort` gains
  `GetTemplateVersionState(ctx, tenantID, versionID string) (*string, string, error)`; `IsPublished`
  retained unchanged (taxonomy's contract).
- **Impl:** `templates/infrastructure.TemplateVersionReader.GetTemplateVersionState`, reusing the single
  existing `templateVersionQuery` (`SELECT v.status, t.doc_type_code FROM templates_template_version v
  JOIN templates_template t ON t.id = v.template_id WHERE v.id = $1 AND t.tenant_id = $2::uuid`), with
  the same NullString / not-found semantics CD's checker had. `tenantID` is passed **explicitly** (not
  `tenant.FromContext`) to match Consumer 1's contract and the off-tx CD-create path.

## What this feature implements

1. **Extend the templates-owned port** with the raw-state primitive (domain interface + infra method),
   reusing the existing query. One owning adapter over `templates_*`.
2. **Delete** `controlleddocuments/infrastructure.PostgresTemplateVersionChecker` (struct, constructor,
   `GetTemplateVersionState`) — the cross-module reach. CD's `module.go` wires
   `templatesinfra.NewTemplateVersionReader(deps.DB)` as `tplCheck`; it satisfies CD's
   `application.TemplateVersionChecker` directly (same method name + signature). `service.go:209/308`
   unchanged.
3. **Replace the `status := "published"` hardcode** in `wiring/documents_adapters.go`:
   `profileDefaultsAdapter` gets the templates reader injected and reads the real status via
   `GetTemplateVersionState(ctx, tenantID, *defaultTemplateVersionID)`.
4. Reads stay **live** (no snapshot/migration); reads stay **off** the lock-holding tx — call sites are
   not moved (auto-path read at `service.go:308` already runs on the reader's own pool conn and is
   non-authz; H-PRE-1 not in play and not regressed).

## Non-goals (mandatory)

- **No** migration of taxonomy's `IsPublished` consumer to the new method, and **no** signature change
  to `IsPublished` (keep ctx-tenant). Migrating it to explicit-tenant for consistency is a **bounded
  defer** (trigger: next structural touch of the templates port), not F4.2.
- **No** change to CD's `application.TemplateVersionChecker` interface shape, to OpenAPI, or to any HTTP
  route/contract (internal module port only).
- **No** change to CD-create tx/lock structure; **no** moving the override-validation read sites; **no**
  snapshot/denormalization of template status (D4/Approach-3 — reads stay live).
- **No** adjacent refactor or opportunistic cleanup beyond the named sites + the port files
  (CLAUDE.md §5.3).
- **No** ADR authoring here — the two port ADRs are **F4.3** (this spec only links them once written).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| Templates port returns raw status present / NULL→nil / not-found→nil, tenant-scoped, with `doc_type_code` | new `TestTemplateVersionReader_GetTemplateVersionState_Live` (`-tags integration` against live PG) — present_returns_status+doctype, absent_returns_nil_nil, other_tenant_returns_nil | **real (live PG)** |
| `IsPublished` still works (no taxonomy regression) | `go test ./internal/modules/taxonomy/... ./internal/modules/templates/...` green | fixture + real |
| CD override validation behavior-identical via the port (published passes, draft/obsolete rejected, profile-mismatch rejected) | existing `controlleddocuments/application/service_test.go` override cases green (reader injected as `tplCheck`) | fixture |
| **0** `templates_template`/`templates_template_version` SQL under `controlleddocuments/` | `grep -rn "templates_template" internal/modules/controlleddocuments/` → only comments, no SQL | real |
| **0** `status := "published"` in `wiring/` | `grep -rn 'status := "published"' apps/api/internal/wiring/` → none | real |
| Adapter returns the **real** status (obsolete default no longer falsely resolves) | new/updated unit test on `profileDefaultsAdapter.GetDefaultTemplateVersionID` asserting status comes from an injected port (obsolete fixture → status "obsolete", not "published") | fixture |
| Status read stays off the CD-create lock-holding tx (H-PRE-1 intact) | CD-create runtime path + `pg_locks` evidence (read on reader pool conn, non-authz; call sites unmoved) | **real (runtime)** |
| `go build ./...` + `go vet` (incl. `-tags integration`) clean | `go build ./...`; `go vet ./internal/modules/templates/... ./internal/modules/controlleddocuments/... ./apps/api/internal/wiring/...` | — |
| backend-api-qa-checklist + workflow-async-qa-checklist (CD-create lock-bearing) green | checklists run at feature close | — |

> TDD: write the failing test first (port raw-state read; adapter real-status), then implement to green.

## ADR needed?

- [x] Durable decision made → recorded in **F4.3** as the `TemplateVersionStateReader` boundary ADR
  under `wiki/decisions/` (owning-module port; reads live, no snapshot; alternatives rejected incl.
  Approach 2 and the "introduce a parallel port" option this feature rejected). Link added here once
  F4.3 writes it: `<f4.3 adr link>`.
