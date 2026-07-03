# Milestone 0 — Validation Contract (D4, authored BEFORE implementation)

> **Program:** global-maximum-remediation · **Milestone:** 0 — VersionRef contract refactor
> **Authored:** 2026-07-03 — **before any implementation task began** (mission.md D4).
> **Binding:** This document is the acceptance contract. Implementation is compared against it;
> drift is **HS-7** (fix the code to the contract, or re-open the contract with operator approval —
> never silently rationalize the contract to match the code). The `milestone-validator` reads this
> file to judge C1/C2 compliance.
> **Derived from:** the 8 locked constraints of
> `docs/superpowers/analysis/2026-07-03-versionref-contract-refactor-system-impact.md` §10 and the
> per-task expected outputs of `docs/superpowers/plans/2026-07-03-versionref-template-contract.md`.

---

## 1. Expected wire shapes (the contract truth)

### 1.1 `TemplateVersionRef` (new component schema)

Compact reference to a template version. **All four fields required, none nullable.**

```json
{
  "id": "<uuid>",              // string, format uuid — the version row id
  "number": 1,                 // integer ≥ 1 — internal version counter
  "revision_number": 0,        // integer int32 ≥ 0 — regulated REV{nn} (ADR 0013)
  "status": "under_review"     // enum: draft | under_review | approved | published | obsolete
}
```

- Generated Go: struct `TemplateVersionRef{ Id openapi_types.UUID; Number int; RevisionNumber int32; Status TemplateVersionRefStatus }`.
- Generated FE: `components["schemas"]["TemplateVersionRef"]` with all four keys non-optional.

### 1.2 `TemplateDTO` (reshaped — list projection)

- **Added / reshaped:**
  - `latest_version: TemplateVersionRef` — **required, non-null** (a template always has ≥1 version).
  - `published_version: TemplateVersionRef | null` — **required-and-nullable**: the KEY is always
    present; its value is a full ref object when the template has a published version, or literal
    JSON `null` when it has never been published. **Never absent.**
- **Removed (four flat scalars — MUST NOT appear anywhere on the wire):**
  - `latest_revision_number`
  - `published_version_id`
  - `published_version_number`
  - `current_revision_number`
- **Byte-identical (unchanged):** `id, tenant_id, doc_type_code, key, name, description, created_by,
  created_by_display_name, created_at, updated_at, archived_at`.
- **Required list (final):** `[id, tenant_id, doc_type_code, key, name, latest_version,
  published_version, created_by, created_at, archived_at]`.

Generated Go non-negotiable: `TemplateDTO.LatestVersion TemplateVersionRef` (value, `json:"latest_version"`);
`TemplateDTO.PublishedVersion *TemplateVersionRef` (`json:"published_version"` — **NO `omitempty`**, so
a nil pointer marshals as `"published_version": null`, not an absent key). If oapi-codegen emits an
inline anonymous struct instead of a named `*TemplateVersionRef`, that is HS-3: restructure the spec
(intermediate named schema) until a named pointer type is generated. The pin test enforces wire truth
regardless of the Go rendering.

### 1.3 `getTemplate` envelope (detail view — UNCHANGED)

`latest_version: VersionDTO` (full object) stays as-is. Same field name carrying a compact ref in the
list and the full object in detail is **AIP view semantics** (ADR 0065 §4) — the previous
int-vs-object collision on `latest_version` dissolves. No change to the envelope in this milestone.

### 1.4 Out of scope this milestone (documents)

`DocumentSummary` / `DocumentDetailResponse` (`current_revision_id` + `revision_version` +
`revision_number`) are **NOT touched** in M0. Any change to them is scope drift (HS-6). The
`DocumentRevisionRef` migration is Plan 2, deferred to pre-v1.

---

## 2. Expected pin-test behaviors (backend contract guards)

File: `internal/modules/templates/delivery/http/template_dto_nullable_fields_test.go` (rewritten).
These are the ADR 0065 contract pins — repaired, not deleted (contract/invariant guard class).

| Pin | Setup | Assertion |
|-----|-------|-----------|
| **P1 — present-and-null** | `TemplateRead` with `Published == nil` (never published), `Latest` set | Marshaled JSON has key `published_version` present with value **exactly `null`** (string compare `string(raw["published_version"]) == "null"`). The 9f86828b guarantee. |
| **P2 — nested ref field-set** | same read, `Latest.Status = under_review` | `latest_version` decodes to an object whose key-set is **exactly** `{id, number, revision_number, status}` (no more, no fewer) and `status == "under_review"`. |
| **P3 — removed keys absent** | same read | Marshaled top-level object does **NOT** contain any of: `latest_revision_number`, `published_version_id`, `published_version_number`, `current_revision_number`. |
| **P4 — present-when-set** | `TemplateRead` with `Published` = a full `VersionRef{status: published}` | `published_version` decodes to a full ref object with all four keys and `status == "published"`. |

All four are non-integration (pure marshal) — MUST pass in `go test ./internal/modules/templates/...`
without the `integration` build tag.

Repository projection guards (`postgres_test.go`, `list_templates_pagination_test.go`, integration
fixtures): fixture SELECT column lists extended to the new projection set (`lv.id`, `lv.status`,
`pv.status` added). Guards of ADR 0013 revision semantics → retarget assertions to
`TemplateRead.Latest/Published`. One-off scaffolding tests that break → delete, deletion noted in
evidence.

---

## 3. Expected FE consumer gating (single-object rule — locked constraint 8)

Every consumer gates on the **whole nullable object** (`published_version == null`), **never on an
inner field**. Expected post-change behavior:

| Consumer | Expected gate / behavior |
|----------|--------------------------|
| `features/templates/api/templates.ts` | Override shrinks (generated `latest_version`/`published_version` correct as-is); exports `VersionRef` type. |
| `wizard/steps/StepTemplate.tsx` | `selectable = template.published_version != null`; selection passes `published.id`; unselectable badge text keyed off `latest_version.status` (`under_review`→"em revisão", `approved`→"aguardando publicação", else "sem versão publicada"). Resolves the `:80` TODO. |
| `wizard/steps/StepConfirm.tsx` | Label uses `template.published_version != null ? …revision_number… : name`. |
| `templates/TemplatesListPage.tsx` | status = archived → `archived`; `published_version != null` → `published`; else `draft`. Version chip = `published_version?.revision_number ?? latest_version.revision_number`. |
| `templates/adapters/useTemplateArtifact.ts` | `published_version != null ? formatRevisionCode(published_version.revision_number) : EM_DASH`; hint "vigente" / "não publicada". |
| `taxonomy/queries/usePublishedTemplatesQuery.ts` | filters `t.published_version != null` with type guard → `PublishedTemplate`. |
| `taxonomy/components/ProfileEditDialog.tsx` | `value={t.published_version.id}` — no `!` assertion (narrowed by the query type guard). |

Sweep expectation: `grep -rn "published_version_id\|current_revision_number\|latest_revision_number\|published_version_number" src --include=*.ts --include=*.tsx | grep -v api-types` returns **zero non-test hits** (test fixtures updated to the new shape).

`tsc --noEmit` clean. `vitest run src/features/{documents,templates,taxonomy}` green (StepTemplate
regression fixtures cover absent-KEY drift case (cast), `published_version:null` case, full-ref
published case, and a `latest_version.status:under_review` badge case).

---

## 4. Expected live-drive outputs (runtime QA — mission D4, runtime-visible milestone)

Start via `.\scripts\start-api.ps1 -Build` (PowerShell only; never bash/source .env). API on :8081.

**Drive 1 — `GET /api/v1/templates`** (cookie-session login, seeded local dev creds):
- Each item's `latest_version` is a nested object `{id, number, revision_number, status}`.
- Never-published items: `"published_version": null` (**key present**, literal null).
- Published items: `published_version` is a full ref object with `"status": "published"`.
- **None** of the four removed keys appear on any item.

**Drive 2 — `GET /api/v1/documents`**: responds 200; `DocumentSummary` shape **unchanged** from
pre-M0 (proves no scope drift into documents). (Read-only confirmation; documents untouched.)

**Drive 3 — `/documents/new` wizard Step 3** (preview browser): a never-published template card is
**disabled / not selectable** with the status-precise unselectable badge; a published template card
is selectable and shows "publicada". Screenshot captured as evidence. This is the exact regression
the 9f86828b bug represented — the acceptance proof that the shape fix closed it.

---

## 5. Expected gate/command outputs (evidence set)

| Command | Expected result |
|---------|-----------------|
| `& .\scripts\openapi-lint-local.ps1` | `Woohoo! Your API description is valid.` |
| `go build ./...` | exit 0, no output |
| `go vet ./internal/modules/templates/...` | exit 0 |
| `go test ./internal/modules/templates/... -count=1` | all non-integration PASS (P1–P4 among them) |
| `go vet -tags=integration ./internal/modules/templates/...` | compiles (integration tests build-check only; running them is a bounded defer — box constraint) |
| `pnpm exec tsc --noEmit` | clean |
| `pnpm exec vitest run src/features/documents src/features/templates src/features/taxonomy` | all green (node_modules junction drift is a known env risk — if vitest dies on module resolution, report verbatim, keep tsc as the gate, record bounded defer) |

**Bounded defers pre-authorized by this contract:** (a) full integration suite not run (20+ min box);
(b) integration-tagged templates tests build-checked only, not executed; (c) vitest blocked by
node_modules junction drift → tsc is the type gate, defer with the complete-pnpm-install trigger.
Each must be recorded in `evidence.md` with its trigger. Any defer beyond these three is new scope —
surface it.

---

## 6. Expected ADR / docs end-state

- `wiki/decisions/0065-version-references-are-nested-value-objects.md` exists (Status: Accepted
  2026-07-03), row added to `wiki/decisions/index.md`, cited by the cutover commit message.
- ADR 0035 memory (`adr0035-flat-envelope-drift.md`) + doc annotated: optional-vs-null subclass
  **structurally closed** for templates; documents pending Plan 2.
- `wiki/modules/templates.md`, `wiki/architecture/api-contract.md` (+ `api-design-system.md` if it
  enumerates DTO conventions) updated with the nested-ref rule; `Last verified` re-stamped 2026-07-03.
- Gate artifact §10 constraints 5 and 6 amended per plan Task 1 Step 2 (read-model split wording;
  SQL projection-extension wording).

---

## 7. Definition of done for M0 (all must hold)

1. §1 wire shapes shipped exactly (spec + regenerated code, zero hand-edits).
2. §2 pins P1–P4 pass; projection guards repaired; deletions noted.
3. §3 FE gating in place; sweep zero non-test hits; tsc clean; vitest green (or bounded defer).
4. §4 live drives produce the stated outputs; wizard screenshot captured.
5. §5 commands produce the stated results; only the three pre-authorized defers, each recorded.
6. §6 ADR 0065 + ADR 0035 annotation + wiki + gate-artifact amendment landed.
7. All work committed locally (standing auth); **never pushed**; `docs/release/` never committed.
8. `milestone-validator` verdict PASS written to `qa/milestone-qa.md`.
