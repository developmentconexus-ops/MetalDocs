# Feature F0.1 — versionref-cutover — Spec

> **Milestone:** 0 — VersionRef contract refactor  ·  **Folder:** `f0.1-versionref-cutover`
> **Status:** Approved (pre-code)
> **Approved before code:** 2026-07-03 / Leandro (via mission /goal directive + committed Yellow gate + committed plan) — *no implementation begins until this line is filled.* ✅

> This is the feature's **contract**, approved **before any code**. What it must do and how it is
> proven — not how it is built (that is `plan.md` = the committed
> `docs/superpowers/plans/2026-07-03-versionref-template-contract.md`).

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Was a consumer-contract interview needed? | **None needed** — the consumer contract is already explicit and discovered, not guessed: (a) the FE consumers' required shape is enumerated in the Yellow gate artifact §8 and plan Task 9; (b) the exact wire shape is locked in gate artifact §7 + §10 (8 constraints) and restated field-by-field in this milestone's `validation-contract.md` §1; (c) the pin-test contract is in `validation-contract.md` §2. The brainstorming/consumer-discovery step was performed during the `developing-new-work` gate that produced the analysis artifact. This row is the C1 evidence. |

## Consumer contract (FIRST — before any producer)

- **Consumers:**
  - FE (templates list + wizard): `templates.ts`, `StepTemplate.tsx`, `StepConfirm.tsx`,
    `TemplatesListPage.tsx`, `useTemplateArtifact.ts`; taxonomy `usePublishedTemplatesQuery.ts`,
    `ProfileEditDialog.tsx`.
  - Backend contract guards: `template_dto_nullable_fields_test.go` (marshal pins), repository
    projection tests.
- **Contract:** exactly `validation-contract.md` §1 (wire shapes) + §2 (pin behaviors) + §3 (FE
  gating). Key points the consumer relies on: `latest_version` is a required nested
  `TemplateVersionRef`; `published_version` is required-and-nullable (present-and-null, never
  absent); consumers gate clonability on `published_version == null` (the whole object), never on
  an inner field; the four flat scalars are gone.
- **Source of truth for the contract:** `api/openapi/v1/openapi.yaml` (`TemplateVersionRef` +
  reshaped `TemplateDTO`) → generated `internal/modules/templates/api/api.gen.go` +
  `frontend/apps/web/src/lib/api-types/index.d.ts`. ADR 0065 is the governing decision.

## What this feature implements

The full atomic pre-v1 cutover of templates' version-pointer wire contract from five coupled flat
scalars to two nested `TemplateVersionRef` value objects — spec + regenerated BE/FE types + domain
`VersionRef` value object + `TemplateRead` read model (aggregate stops carrying join-projection
scalars) + repository ref projection + delivery mapper emitting nested objects + FE consumers gating
on the single nullable object. Producer satisfies the consumer contract above. Plan tasks 2–12.

## Non-goals (mandatory)

- **Documents module** (`DocumentSummary`/`DocumentDetailResponse`, `current_revision_*`) — Plan 2,
  deferred. Any documents wire change here is scope drift (HS-6).
- **DB schema/migration** — none; SQL SELECTs gain projection columns only.
- **getTemplate detail envelope** — unchanged (`latest_version: VersionDTO` stays).
- **Write-path / command params** — response-shape reshaping only; no authz, tenancy, or outbox change.
- **New shared Go DTO package / cross-context schema** — `TemplateVersionRef` lives in the templates
  module only (locked constraint 2).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| P1 present-and-null | `TestTemplateDTO_PublishedVersionRef_PresentAndNull` | real (marshal) |
| P2 nested ref field-set exactly `{id,number,revision_number,status}` | same file, field-set assertion | real |
| P3 four removed keys absent | same file | real |
| P4 published → full ref object | `TestTemplateDTO_PublishedVersionRef_PresentWhenSet` | real |
| Backend builds + tests green | `go build ./...`; `go test ./internal/modules/templates/... -count=1` | real |
| Spec valid | `& .\scripts\openapi-lint-local.ps1` → "valid" | real |
| FE types + consumers | `pnpm exec tsc --noEmit`; `pnpm exec vitest run src/features/{documents,templates,taxonomy}` | real |
| Live wire | `GET /api/v1/templates` shows nested refs + `published_version:null` key-present, zero removed keys | real (live API) |
| Live wizard | `/documents/new` Step 3 — unpublished not selectable, status-precise badge | real (preview) |

> TDD: pins P1–P4 written/rewritten as failing-first against the new shape, then implement to green.

## ADR needed?

- [x] Durable decision made → ADR 0065 (authored in sibling feature F0.2):
  `wiki/decisions/0065-version-references-are-nested-value-objects.md`.
