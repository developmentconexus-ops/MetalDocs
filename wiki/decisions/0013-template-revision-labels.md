# ADR 0013 — Template Revision Labels (REV-style, Backend-Canonical)

> **Last verified:** 2026-05-29
> **Scope:** Adopt `REV{nn}` chip labels for template versions, mirroring documents' existing revision-code convention. Make `revision_number` a **first-class persisted column** on `templates_template_version`, exposed via OpenAPI as `current_revision_number` on `TemplateDTO`. Frontend renders the field directly through a shared formatter — no off-by-one math in the UI.
> **Out of scope:** Renaming `version_number` column (kept — it remains the lifecycle counter). Documents-side changes (already uses REV).
> **Key files (touched in implementation PR):**
> - `migrations/01XX_templates_revision_number.sql` — new migration adding column + backfill + unique index
> - `internal/modules/templates/domain/template.go` — `TemplateVersion.RevisionNumber int`, `Template.CurrentRevisionNumber *int`
> - `internal/modules/templates/repository/postgres.go` — INSERT allocation, SELECT projection (LEFT JOIN already in place from PR #33)
> - `internal/modules/templates/repository/mappers.go` — scanner update
> - `api/openapi/v1/openapi.yaml` and `api/openapi/v1/partials/templates.yaml` — `current_revision_number` on `TemplateDTO`
> - `internal/modules/templates/api/api.gen.go` — regenerated
> - `frontend/apps/web/src/lib/api-types/index.d.ts` — regenerated
> - `frontend/apps/web/src/lib/labels/revisionCode.ts` — lifted shared helper (currently at `features/documents/lib/documentDetailMeta.ts:78`)
> - `frontend/apps/web/src/features/templates/TemplatesListPage.tsx` — chip composition
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepTemplate.tsx:102` — wizard template pill
> - `frontend/apps/web/src/features/documents/components/wizard/steps/StepConfirm.tsx:68` — confirm-step template pill
> - `frontend/apps/web/src/features/templates/__tests__/TemplatesListPage.versionChip.test.tsx` — fixtures updated to REV codes
>
> **Predecessor:** ADR 0012 (`contract-first-api`) — `published_version_number` was added via that contract path on PR #33 (`bug/templates-version-chip`, commit `57db4e53`). This ADR builds on the same contract-first pattern.

---

## Status

Accepted — 2026-05-29 (implemented on `feat/templates-rev-labels`)

Awaiting ratification. Implementation lives on `feat/templates-rev-labels` (not yet opened).

---

## Context

### Convention drift between modules

Documents (regulated, instance-side) labels revisions as `REV00 / REV01 / …`. The persistence model:

- Column: `documents.revision_number INT NOT NULL` — added by `migrations/0131_documents_v2_state_columns.sql:19`.
- Allocation: at INSERT, `COALESCE((SELECT MAX(d2.revision_number) + 1 FROM documents d2 WHERE d2.controlled_document_id = $6), 0)` — `internal/modules/documents/repository/repository.go:115`. **0-based at runtime** (default `1` in schema is never observed; INSERT always supplies the value).
- Uniqueness: `ux_documents_v2_cd_revision` on `(controlled_document_id, revision_number) WHERE controlled_document_id IS NOT NULL`.
- API: `RevisionNumber: integer/int64` on the document DTO (`api/openapi/v1/openapi.yaml:3574`), required.
- FE: `formatRevisionCode(n)` at `frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts:78` renders `REV${String(Math.max(n, 0)).padStart(2, '0')}`.

Templates currently label the same concept as `v1 / v2 / …`. After PR #33 the chip is honest (reads `published_version_number` instead of `latest_version`) but the *label scheme* still diverges from documents.

### What PR #33 left on the table

PR #33 fixed correctness only:

- Backend: `published_version_number *int` added to `Template`, `LEFT JOIN templates_template_version pv` in repo SELECTs, set at both publish paths.
- FE: chip reads `published_version_number` for published, `latest_version` for drafts.

The *label redesign* — switching from `v{n}` to REV-style — was deferred because:

1. It changes user-visible strings across multiple surfaces.
2. The mapping decision (1-based version vs 0-based revision) has cross-module implications.
3. It requires structural alignment with documents' regulated-doc vocabulary — an ADR-grade decision.

### Why FE off-by-one rendering was rejected

An earlier draft of this ADR proposed Alternative B below — render `REV{published_version_number - 1:02d}` in the FE. **Rejected.** Reasons:

1. **Hidden contract knowledge in the UI.** The off-by-one becomes a tribal-knowledge invariant the FE must know about the backend's 1-based version counter. Future maintainers see `v - 1` and can't tell whether it's intentional or a bug.
2. **Asymmetry with documents.** Documents emits a clean 0-based `revision_number` on its DTO. The FE renders it directly with no math. Templates introducing an FE-side translation breaks the cross-module mental model the moment a reader compares the two renderers.
3. **Audit and tooling consumers.** Any non-FE consumer of the API (audit pipeline, ETL, future reports) would need to know to subtract `1` to get the regulated revision code. The translation belongs in one place — the canonical contract — not in every consumer.
4. **Wiki/ADR architecture rules.** `wiki/architecture/backend-api-structure.md` requires "OpenAPI first" and forbids reconstructing contract-derived values outside the contract boundary. An FE-side `-1` violates the spirit of contract-first.

### Numbering semantics: 0-based vs 1-based

| Concept | Schema column | Counter base | First value | Example |
|---|---|---|---|---|
| Document revision | `documents.revision_number` | **0-based** (runtime) | `0` | `REV00` |
| Template version (lifecycle) | `templates_template_version.version_number` | **1-based** | `1` | (internal: `v1`) |
| Template revision (new, this ADR) | `templates_template_version.revision_number` | **0-based** | `0` | `REV00` |

Templates keep `version_number` because it remains the lifecycle counter — every draft, including unpublished tails, advances it. `revision_number` is the regulated-revision counter — what the chip and the audit log display to humans.

By construction (no version_number gaps allowed in the templates lifecycle today), `revision_number == version_number - 1` for every row. The column is therefore **redundant in raw arithmetic** but **non-redundant in contract intent**: it is the stable, contract-named field that all consumers read.

---

## Decision

### D-1. Persist `revision_number` on `templates_template_version`

New column:

```sql
ALTER TABLE templates_template_version
  ADD COLUMN IF NOT EXISTS revision_number INT NOT NULL DEFAULT 0;
```

Backfill in same migration (one-shot):

```sql
UPDATE templates_template_version SET revision_number = version_number - 1;
```

Unique index per template:

```sql
CREATE UNIQUE INDEX IF NOT EXISTS ux_templates_version_revision
  ON templates_template_version (template_id, revision_number);
```

### D-2. Allocate at INSERT (repository)

Mirror documents' atomic allocation pattern. In `repository/postgres.go` version-creation paths:

```go
COALESCE((SELECT MAX(rn) + 1 FROM templates_template_version
          WHERE template_id = $X), 0)
```

The DEFAULT `0` in the schema is a safety floor; the INSERT always supplies the value via the subquery.

### D-3. Domain types

```go
// internal/modules/templates/domain/template.go

type TemplateVersion struct {
    ...
    VersionNumber  int  // 1-based, lifecycle counter (unchanged)
    RevisionNumber int  // 0-based, regulated-revision counter (NEW)
    ...
}

type Template struct {
    ...
    PublishedVersionNumber  *int  // from PR #33 — kept for tooling/audit
    CurrentRevisionNumber   *int  // NEW — revision_number of the published version
    ...
}
```

`CurrentRevisionNumber` is `nil` when the template has never been published. When non-nil, it equals the `revision_number` of the row referenced by `published_version_id`.

### D-4. Repository projection

`GetTemplate` / `GetTemplateByKey` / `ListTemplates` already `LEFT JOIN templates_template_version pv ON pv.id = t.published_version_id` (added in PR #33). Add `pv.revision_number AS current_revision_number` to the projection. Scanner updated to `sql.NullInt32 → *int`.

### D-5. Application setter

In `application/lifecycle.go`, both publish paths (`PublishTemplateVersion`, `Approve` Accept branch) already set `PublishedVersionID` and `PublishedVersionNumber`. Extend the same blocks to set `CurrentRevisionNumber = &version.RevisionNumber`.

### D-6. Contract

`TemplateDTO` gains:

```yaml
current_revision_number:
  type: integer
  format: int32
  nullable: true
  minimum: 0
  description: >
    Regulated revision number of the currently published template version.
    0-based (REV00, REV01, ...). Null when the template has never been published.
```

Optional: also add `revision_number` to the `TemplateVersionDTO` schema if/when that DTO is consumed by the FE. (Out of scope for this ADR unless a screen needs it.)

### D-7. Codegen

`go generate ./internal/modules/templates/api/...` regenerates `api.gen.go`. `pnpm gen:api` regenerates `frontend/apps/web/src/lib/api-types/index.d.ts`.

### D-8. Frontend rendering

1. Lift `formatRevisionCode` from `features/documents/lib/documentDetailMeta.ts:78` to a shared location — proposed `frontend/apps/web/src/lib/labels/revisionCode.ts`. Re-export from the documents path for backward compatibility (no documents-side import changes).
2. Update three sites to render directly from `current_revision_number`:

```tsx
// TemplatesListPage chip
const revisionLabel =
  dto.current_revision_number != null
    ? formatRevisionCode(dto.current_revision_number)
    : null; // draft / never-published: no REV code (see D-8.4)
```

3. **Strict draft policy (resolves Open Question #1).** When `current_revision_number` is `null` (template has never been published), the chip renders **no REV code**. The "RASCUNHO" status pill already communicates draft state — fabricating a `REV??` label from `latest_version` would imply a regulated-revision identity the template does not yet have. Honesty rule from PR #33 extends to label naming: no value, no label. The FE renders `null`/empty string in the version-code slot for drafts and lets layout collapse, or shows only the status pill if no other content exists.

4. Wizard pills (`StepTemplate.tsx:102`, `StepConfirm.tsx:68`) read `current_revision_number` and render via the helper. If `null`, the pill is omitted entirely (the template selector should not show un-published templates in this context anyway — to be confirmed during implementation).

### D-9. Tests

- **Migration smoke**: backfill correctness — `revision_number = version_number - 1` for every existing row.
- **Repo allocation test**: inserting versions 1, 2, 3 yields `revision_number` 0, 1, 2.
- **Repo SELECT test**: `current_revision_number` populated when `published_version_id` set, null otherwise.
- **Application lifecycle test**: both publish paths set `CurrentRevisionNumber`.
- **Handler test**: `current_revision_number` present in JSON response.
- **FE vitest**: `TemplatesListPage.versionChip.test.tsx` fixtures updated:

| Fixture state | New expectation |
|---|---|
| Never-published v1 (`current_revision_number` null) | **no REV chip** (only "RASCUNHO" status pill) |
| Published v1 + auto-draft v2 (`current_revision_number=0`) | `REV00` |
| Published v1 only (`current_revision_number=0`) | `REV00` |
| Archived latest_version=3, published_version_number=2 (`current_revision_number=1`) | `REV01` |

- **Shared helper unit tests** for `formatRevisionCode` (already exist in documents; lifted into shared module unchanged).

### D-10. Audit log

- New audit message rendering may use REV codes.
- Stored audit payload keeps `version_number` (numeric, contract-stable).
- Audit list UI update is **bounded defer** — separate sweep, separate PR.

### D-11. Compatibility

- `published_version_number` (PR #33) stays on `TemplateDTO`. It remains the contract-stable answer to "which version is live." Removing it would break the chip-honesty contract PR #33 just established.
- `version_number` is unchanged. Drop only after a full deprecation cycle, never inside this ADR.

---

## Consequences

### Positive

- Cross-module vocabulary unified. `REV02` on a template list and `REV02` on a document list mean the same generation of regulated artifact.
- Contract-canonical: FE renders the backend field directly via a shared helper. No FE arithmetic.
- Audit / ETL / future consumers all read `current_revision_number` from one canonical source.
- Architectural alignment with `wiki/architecture/backend-api-structure.md` and ADR 0012 (contract-first).

### Negative

- Schema migration on a populated table (small — templates is low-volume).
- One redundant column per version row (4 bytes). Negligible.
- User-visible string change. Mitigation: changelog entry + release note.

### Neutral

- `version_number` and `revision_number` remain coupled today (`rn = vn - 1`). The column gives us **future room** to break that coupling (branched versions, non-monotonic flow) without contract churn.

---

## Alternatives Considered

### A. Status quo (`v{n}`) after PR #33

Pro: zero risk. Con: cross-module inconsistency persists; auditor confusion persists. **Rejected.**

### B. FE off-by-one rendering (`REV{published_version_number - 1:02d}`)

Pro: pure FE change, no migration. Con: hides a contract translation in the UI; asymmetric with documents; every non-FE consumer must learn the off-by-one; violates contract-first. **Rejected.** This is the alternative the user explicitly rejected when reviewing an earlier draft of this ADR.

### C. Backend-canonical `revision_number` column + contract field (**chosen**)

Pro: contract-first, matches documents, no FE arithmetic. Con: schema migration. **Accepted.**

### D. Generated column `revision_number GENERATED ALWAYS AS (version_number - 1) STORED`

Pro: no INSERT-side allocation logic, automatic. Con: Postgres GENERATED has trigger-interaction edge cases and limits row updates; documents doesn't use it; introducing a different pattern for a parallel concept is asymmetric. **Rejected.**

### E. Drop `version_number`, keep only `revision_number`

Pro: one counter, no off-by-one anywhere. Con: massive blast radius (FK refs, audit history, every consumer). **Rejected.**

---

## Implementation Plan

Single PR on `feat/templates-rev-labels` off `main`, after PR #33 merges.

1. Migration `01XX_templates_revision_number.sql`: column + backfill + unique index.
2. Repo INSERT: allocate `revision_number` via `MAX+1` subquery on every version-creation path (locate via `grep -nE "INSERT INTO templates_template_version" internal/modules/templates`).
3. Repo SELECT: add `pv.revision_number AS current_revision_number` to the three projections already touched in PR #33.
4. Repo scanner: extend `scanTemplate` to map `sql.NullInt32 → *int` for the new field.
5. Domain: `TemplateVersion.RevisionNumber int`, `Template.CurrentRevisionNumber *int`.
6. Application: set `CurrentRevisionNumber` at both publish paths (same blocks as `PublishedVersionNumber` in PR #33).
7. OpenAPI: add `current_revision_number` to `TemplateDTO` in `api/openapi/v1/openapi.yaml` **and** `api/openapi/v1/partials/templates.yaml`.
8. Codegen: `go generate ./internal/modules/templates/api/...` + `pnpm gen:api`.
9. FE helper lift: move `formatRevisionCode` to `frontend/apps/web/src/lib/labels/revisionCode.ts`; re-export from old path.
10. FE render: `TemplatesListPage.tsx`, `StepTemplate.tsx:102`, `StepConfirm.tsx:68`.
11. Tests: migration, repo, application, handler, vitest.
12. Wiki: update `wiki/modules/templates.md` "Version chip source-of-truth" section to reference this ADR and the new column.
13. Release note entry flagging the user-visible label change.

---

## Verification

Backend gates (per `metaldocs-backend-api` skill):

```powershell
npx @redocly/cli lint api/openapi/v1/openapi.yaml
$env:GOFLAGS = "-mod=mod"
go generate ./internal/modules/templates/api/...
go build ./...
go test ./internal/modules/templates/... -count=1
.\scripts\start-api.ps1 -Build
```

Frontend gates:

```powershell
cd frontend/apps/web
pnpm gen:api
npx tsc --noEmit
pnpm vitest run features/templates
```

Live preview drive of all four chip states; capture snapshots under `.qa-reports/templates-rev-labels.md`.

---

## Open Questions

1. ~~**Draft revision rendering.**~~ **Resolved 2026-05-29 (user):** drafts render **no REV code**. Honesty rule — no published revision exists, so no `REV??` label is fabricated. Encoded in D-8.3.
2. **Audit-log row renderer scope.** Update in this PR or separate sweep?
3. **Feature flag.** Roll out behind `templates.rev_labels = on` for safe rollback?

---

## References

- ADR 0012 — Contract-first API (`wiki/decisions/0012-contract-first-api.md`)
- PR #33 — Honest version chip + TemplateDTO `$ref` refactor (precondition)
- `wiki/architecture/backend-api-structure.md` — Canonical module pattern + verification gates
- `wiki/architecture/api-contract.md` — Contract-first overview
- `wiki/concepts/iso-segregation.md` — Audit-language conventions
- `wiki/modules/documents.md` — Existing REV implementation
- `migrations/0131_documents_v2_state_columns.sql:19` — Documents' `revision_number` migration (pattern source)
- `internal/modules/documents/repository/repository.go:115` — Documents' atomic allocation pattern
- `frontend/apps/web/src/features/documents/lib/documentDetailMeta.ts:78` — `formatRevisionCode` helper to lift
