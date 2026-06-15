# ADR 0030 — Template-version state reads go through the templates-owned `TemplateVersionPort` (no hardcoded status)

> **Status:** Accepted 2026-06-15
> **Last verified:** 2026-06-15
> **Scope:** How modules other than templates obtain a template version's raw status + owning `doc_type_code`. The owning module (templates), the decision to **extend the existing** owned port rather than introduce a parallel one, the elimination of the `status := "published"` hardcode, and the reads-live / off-tx (H-PRE-1) constraint.
> **Out of scope:** taxonomy's `IsPublished` predicate (kept unchanged); CD-create tx/lock structure; OpenAPI/route shapes (unchanged — internal module port).
> **Key files:**
> - `internal/modules/templates/domain/template_version_port.go` — owned port; gained `GetTemplateVersionState(ctx, tenantID, versionID) (*string, string, error)`; retains `IsPublished`
> - `internal/modules/templates/infrastructure/template_version_reader.go` — single owning adapter over `templates_*`; reuses `templateVersionQuery`
> - `internal/modules/controlleddocuments/module.go` — wires the templates reader as `tplCheck` (satisfies CD's `application.TemplateVersionChecker`)
> - `internal/modules/controlleddocuments/application/service.go` — override-validation consumers (`:209` manual, `:308` auto path) — call sites unchanged
> - `apps/api/internal/wiring/documents_adapters.go` — profile-default resolver; reads real status via the port (was `status := "published"`)

## Context

`templates_template` / `templates_template_version` are owned by the templates module. Two cross-module
sites depended on that state without going through a templates port — both **H-G class** defects:

1. **Reach-without-a-port:** `controlleddocuments/infrastructure.PostgresTemplateVersionChecker` issued
   raw SQL against `templates_*` (`repository.go:702-711`) to validate override templates.
2. **Hardcoded-domain-state:** `wiring/documents_adapters.go:113` fabricated `status := "published"` for
   the profile-default template version instead of reading the owning module. This **masked a real bug**
   — an *obsolete* (or any non-published) profile-default version wrongly passed `resolveDefaultTemplate`.

Pre-execution investigation found templates **already owned** a cross-module port for version state:
`TemplateVersionPort.IsPublished(ctx, versionID) (bool, docTypeCode, error)` (Wave Z Z-7, consumed by
taxonomy). Its `(bool, …)` shape is insufficient for CD, whose `Resolve` distinguishes the raw status
string `published` / `obsolete` / draft (`controlleddocuments/domain/resolution.go:42,55,58`).

## Decision

Template-version state for cross-module consumers flows through the **existing templates-owned port** —
**extended, not duplicated.** `TemplateVersionPort` gains the raw-state primitive
`GetTemplateVersionState(ctx, tenantID, versionID) (status *string, docTypeCode string, err error)`;
`IsPublished` is retained unchanged as a *derived predicate* for taxonomy. The single owning adapter
(`TemplateVersionReader`) owns all `templates_*` SQL + tenant-scoping and reuses the one existing
`templateVersionQuery`.

Introducing a *parallel* `TemplateVersionStateReader` over the same tables was rejected: a second reader
is the duplication anti-pattern this milestone exists to kill (DDD single-owning-adapter;
consumer-driven contracts — raw state is the primitive, `IsPublished` the derived view). The operator
delegated this engineering call ("study it, reach the best solution by industry standards", 2026-06-15).

Consequently: CD's `PostgresTemplateVersionChecker` is **deleted**; CD wires the templates reader as
`tplCheck`, which satisfies CD's own `application.TemplateVersionChecker` consumer interface directly
(shape `(*string, string, error)` preserved — override-validation call sites behavior-identical). The
`documents_adapters.go` profile-default resolver reads the **real** status through the port.

Reads stay **live** (no snapshot/denormalization — design D4/Approach-3) and **off** the CD-create
lock-holding tx. The auto-path read (`service.go:308`) runs inside the create tx but on the reader's own
pool connection and is a plain non-authz `SELECT` — identical connection topology to the deleted
checker — so **H-PRE-1 is not in play and not regressed**; call sites were not moved.

## Consequences

- **0 `templates_*` SQL under `controlleddocuments/`** and **0 `status := "published"` in `wiring/`**
  (grep-verified) — both H-G instances of this class closed at the class level.
- One adapter owns template-version reads; taxonomy (`IsPublished`) and CD (`GetTemplateVersionState`)
  are consumers of one owning module, no duplicated `templates_*` SQL.
- The hardcode bug is fixed: an obsolete profile-default template version now correctly fails to resolve
  as publishable (intended correctness change, not a contract break).
- Status is always current (live read); no snapshot, no migration, no second source.
- **Bounded defer:** migrating taxonomy's `IsPublished` to explicit-tenant for signature consistency —
  trigger: next structural touch of the templates port (recorded in F4.2 non-goals).

## References
- Feature F4.2 — `docs/superpowers/milestones/grade-a-architecture-remediation/milestone-4-systemic-ports/f4.2-template-version-state-reader/spec.md`
- HS-6 reconciliation (extend-vs-introduce) — that milestone's `milestone.md`
- Governing spec — `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` §5.2 (H-G class), §M4
- Sibling port ADR [`0029-user-display-name-reader-port.md`](0029-user-display-name-reader-port.md)
- ADR [`0013-template-revision-labels.md`](0013-template-revision-labels.md) (templates version schema)
