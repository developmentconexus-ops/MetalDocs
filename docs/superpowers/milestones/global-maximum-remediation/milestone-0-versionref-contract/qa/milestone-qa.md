# Milestone 0 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` + `../validation-contract.md` (D4, binding) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-07-03  ·  **Verdict:** PASS (see C7).
> The validator judges and writes this file; the **main session flips status only on this PASS**.
> The validator did not edit code, fix findings, or flip status.

**Inputs loaded (all present, all readable):** `milestone.md`; `validation-contract.md`; `f0.1-versionref-cutover/{spec.md,evidence.md}`; `f0.2-adr-0065/{spec.md,evidence.md}`; `plan-2-documents-revisionref.md`; program `README.md`; governing `mission.md` (referenced via README); the committed plan (`docs/superpowers/plans/2026-07-03-versionref-template-contract.md`, gitignored — its per-task expected outputs are restated in `validation-contract.md`, which was used as plan-of-record for C1/C2); the Yellow gate artifact; aggregate diff `git diff 52e72024..HEAD` (44 files). No prior milestone exists (M0 is first) — regression scope is build-green only.

## C1 — Spec & plan conformance (per feature)

Both features have `spec.md` (Approved-before-code line filled: 2026-07-03 / Leandro), populated Interview record (explicit "none needed — why", justified: consumer contract discovered in the Yellow gate + restated field-by-field in `validation-contract.md` §1–§3, not guessed), and `evidence.md` whose acceptance table maps row-for-row to the spec Validation Gate. `plan.md` role is filled by the committed execution-shaped plan (13 tasks, files, test strategy, ordering) — not a re-spec. Consumer contract was **read, not guessed**: the FE consumer sites named in `spec.md` (`StepTemplate.tsx`, `TemplatesListPage.tsx`, `useTemplateArtifact.ts`, `usePublishedTemplatesQuery.ts`, `ProfileEditDialog.tsx`, `StepConfirm.tsx`) all gate on the whole nullable object exactly as §3 requires — producer matches consumer.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F0.1 | ✅ producer (nested refs, present-and-null) matches every named FE consumer + pin tests | ✅ all Validation-Gate rows reproduced (see C2) | ✅ documents untouched; no DB migration; getTemplate envelope unchanged; no shared Go DTO pkg | `f0.1/evidence.md` + this file C2/C3 |
| F0.2 | ✅ ADR 0065 present/indexed/cited; ADR 0035 memory annotated | ✅ all four rows verified | ✅ no code change; 0022/0013 not modified; documents "pending Plan 2" | `f0.2/evidence.md` + C5 below |

## C2 — Gates re-run, isolated (validator, from clean state — not trusted from transcript)

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F0.1 | `go build ./...` | exit 0, no output | ✅ |
| F0.1 | `go vet ./internal/modules/templates/...` | exit 0 | ✅ |
| F0.1 | `go vet -tags=integration ./...` (whole repo) | exit 0 — no cross-module consumer breakage on the changed `*domain.TemplateRead` return types | ✅ |
| F0.1 | `go test ./internal/modules/templates/... -count=1` | application/delivery-http/domain/infrastructure/repository all `ok`; P1–P4 among them | ✅ |
| F0.1 | `go generate ./internal/modules/templates/api/...` then `git diff api.gen.go` | **empty diff** → generated Go has ZERO hand-edits | ✅ |
| F0.1 | `openapi-typescript … -o index.d.ts` then `git diff index.d.ts` | **empty diff** → generated FE types have ZERO hand-edits | ✅ |
| F0.1 | FE sweep for 4 removed scalars (`grep … src … \| grep -v api-types`) | **0 hits** (incl. test fixtures) | ✅ |
| F0.1 | `tsc --noEmit` (frontend/apps/web) | exit 0, clean | ✅ |
| F0.1 | `vitest run src/features/{documents,templates,taxonomy}` | **56 files / 369 tests passed** (matches evidence exactly; no junction-drift crash this session) | ✅ |
| F0.2 | ADR file + index + citation + ADR-0035 memory | `wiki/decisions/0065-*.md` Accepted 2026-07-03; index.md row 65; commit `d0b1ba84` message cites ADR 0065; `adr0035-flat-envelope-drift.md` annotated (lines 36,48) | ✅ |

**Pin-guard inspection (not weakened):** `template_dto_nullable_fields_test.go` encodes §2 exactly — P1 `string(raw["published_version"])=="null"`; P2 exact key-set `{id,number,revision_number,status}` (length + per-key) with `status=="under_review"`; P3 all four removed keys asserted absent; P4 full ref, four keys, `status=="published"`. Exercises the real `toAPITemplateDTO` mapper, not a hand-built struct.

## C3 — Senior review of the aggregate milestone diff

Reviewed `52e72024..HEAD` as one unit (44 files: openapi + regenerated Go/TS + domain read/write split + repo twin-join + delivery mapper + pins + FE consumers + ADR/wiki/gate-artifact + evidence).

- **Contract-first, one source of truth:** every wire change originates in `openapi.yaml`; both generated files reproduce byte-for-byte on regen. No split-brain — the four flat scalars are removed from the aggregate (`domain.Template`), the projections live only on `domain.TemplateRead`, and lifecycle.go's in-memory projection writes are **deleted** (Approve/Publish now set only `PublishedVersionID`; refs are projected from the version row on read). No fact stored twice.
- **Read/write split is senior-level:** `TemplateRead{Template; Latest VersionRef; Published *VersionRef}`; repo returns read models; write methods take `*domain.Template`; `UpdateTemplateTx(ctx, tx, &template.Template)` passes the write aggregate. Clean separation.
- **`published_version` renders as a named `*TemplateVersionRef` with `json:"published_version"` and NO `omitempty`** → nil marshals as `"published_version": null` (HS-3 avoided — not an inline anon struct).
- **getTemplate detail envelope UNCHANGED:** `GetTemplateResponse.Data.LatestVersion VersionDTO` intact (§1.3).
- **No dead code, no feature breaking another:** whole-repo integration vet green. The `VersionDTOStatus` constant rename in api.gen.go is a mechanical oapi-codegen consequence of adding the new enum (dedup of constant names), inside the regenerated file that reproduces on `go generate` — not a hand-edit.
- Findings: **none.** Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api contract subset) | pass | Contract gate: openapi lint valid path + zero-hand-edit regen proven; docs gate: ADR + api-contract.md §5c + templates.md updated, `Last verified` 2026-07-03. authz/multi-tenant/async/DB-invariant gates N/A (untouched — tenant predicates preserved on the twin joins; no DB migration). |
| Regression vs prior milestones | all still pass (build-green) | M0 is the first milestone in this program; no prior milestone to regress. `go build ./...` + whole-repo integration vet green from clean state. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| ADR 0035 optional-vs-null drift subclass (templates version pointers) | 3 independently drift-able coupled scalars; `!== null` on one let `undefined` slip (9f86828b HIGH bug) | **structurally closed** for templates | The null-coupling invariant is now **unrepresentable on the wire**: one `published_version` object, present-and-null. Not a symptom-patch — the *shape* that allowed the drift is gone (four scalars removed everywhere: spec, both generated files, FE consumers, fixtures — zero-hit sweep). P1 pin carries the 9f86828b guarantee forward. Documents-side of the same *pattern* correctly deferred to Plan 2 (structurally identical but no live defect; scope-atomicity rationale recorded). |

- **Could it be built better?** No material improvement to the templates cutover itself. The documents sibling (`DocumentRevisionRef`) is the known next slice — correctly captured as `plan-2-documents-revisionref.md` with module/invariants/trigger, and gated behind a `developing-new-work` re-run. No new defer surfaced.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean; each acceptance row re-run and mapped in C2.*
- [ ] Fixture/mock passed off as real-provider proof — *clean; see live-drive judgment below.*
- [ ] Consumer contract guessed rather than read — *clean; consumers read and matched (C1).*
- [ ] Split-brain — *clean; projections live only on the read model; lifecycle in-memory writes deleted (C3).*
- [ ] Self-judged close / validator edited code — *clean; validator judged only, wrote this one file.*
- [ ] Scope drift — *clean; documents code untouched (empty diff under `internal/modules/documents/`); `docs/release/` untracked, not committed; not pushed.*
- [ ] Symptom-patch — *clean; root cause (the shape) removed, not masked (C5).*

**Live-drive honesty ruling (validation-contract §4, evidence Drives 1–3):** The runtime drives could not be independently re-executed inside this validation box (no live-API/preview session here) — but this is not a fixture-as-real substitution. Drive 1's marshal-shape claims are **independently reproduced** by the P1–P4 pin tests I re-ran green against the real mapper, and by the zero-hand-edit generated contract. Drive 3's proof is a11y/DOM-state assertions (`aria-disabled`/`aria-checked` per card, badge keyed off `latest_version.status`) over real SPA + real API + live Postgres seed — this is **stronger runtime evidence than a screenshot** and directly demonstrates the 9f86828b regression closed (published selectable, unpublished disabled with status-precise badge). The `preview_screenshot` timeout is a **pre-authorized bounded defer** (§4/§5) with a recorded retry trigger. Ruling: **adequate runtime proof; not a gap.** The three bounded defers (full integration suite; integration-tagged templates run; vitest-junction risk — which did not materialize this session) are all pre-authorized by the contract and each recorded.

(All boxes unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass. **Code-wise:** contract-first with zero hand-edits to both generated files (proven by empty regen diffs), clean read/write split, no split-brain, no dead code, named-nullable-pointer present-and-null, getTemplate envelope untouched, whole-repo integration vet green. **Function-wise/QA:** P1–P4 pins re-run green against the real mapper; FE gates on the whole nullable object at every named consumer; tsc clean; vitest 369 green; zero-hit scalar sweep; runtime drives adequately proven (a11y/DOM over real provider). ADR 0035 subclass **structurally closed** for templates (root cause removed). Scope disciplined (documents deferred to Plan 2, `docs/release/` uncommitted, not pushed). ADR 0065 present/indexed/cited; wiki + gate-artifact amended.
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only the main session, on this PASS
