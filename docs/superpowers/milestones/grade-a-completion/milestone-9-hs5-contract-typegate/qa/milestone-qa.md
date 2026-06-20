# Milestone 9 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-20  ·  **Verdict:** see C7 — **PASS**.
> Validator judged and wrote this file only; it edited no source, fixed no findings, flipped no status.

**Inputs loaded (all present, all readable):** `milestone.md`; F9.1–F9.4 `spec.md`/`plan.md`/`evidence.md`
(12 artifacts); program `README.md`; governing `mission.md` (§8 + §8 scope amendment); aggregate diff
`git diff 58dea742..HEAD` (2 commits: `2e3c8a8b` implementation, `9248d715` docs). No missing input.

## C1 — Spec & plan conformance (per feature)

All four `spec.md` carry `Status: approved (operator Option-A proceed, 2026-06-20) — code may begin`
(approval present before code). Each `spec.md` opens with a "Consumer contract (read from the consumer,
before the producer)" section citing the consumer site — F9.1/F9.2 cite the generated OpenAPI model +
`api.gen.go` line; F9.3 cites the FE consumer `DocumentEditorPage.tsx:95` `apiFetch<{url?:string}>`
(consumer-contract-first, the direction that picked 200+{url} over the stale 302); F9.4's consumer is the
mission §8 H-D class itself. No empty interview record on a contract-touching feature. Each `plan.md` is
execution-shaped (task list, files, TDD strategy, ordering) — not a re-spec. Each `evidence.md` acceptance
table maps row-for-row to its `spec.md` Validation Gate.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F9.1 duplicate-typed | ✅ producer emits generated `DocumentCreateResult` (the OpenAPI 201 model) | ✅ | ✅ no service-sig/status/spec change | `handler.go:686`; `TestDocumentCreateResult_WireContract` keys `document_id,initial_revision_id,session_id` |
| F9.2 comments-typed | ✅ 3 handlers emit generated `DocumentCommentResponse`; local `commentResponse` removed | ✅ | ✅ no domain-model/spec/decode change | `handler.go:1154,1249`; `decodeCommentContent` []-not-null; `TestDocumentCommentResponse_WireContract` |
| F9.3 revision-url-contract | ✅ direction chosen by FE consumer (200+{url}); spec aligned to runtime not reverse | ✅ | ✅ no redirect-flow/url-gen change | OpenAPI op now `200`+`RevisionUrlResponse` (yaml:2901-2906, 302 gone); `handler.go:1137`; FE+BE codegen carry the type |
| F9.4 noresponsemap-widen | ✅ gate (the consumer) now enforces any `map[string]<T>` | ✅ | ✅ scope/exemptions unchanged; widened existing analyzer (no new one) | `noresponsemap.go isMapStringLiteral` (value-type check removed); §5b widened; 4th site `finalizeDocument` closed by typing |

All three artifacts exist for every feature in `milestone.md`'s Features table. **C1 PASS.**

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (`-count=1` where re-run), not trusted from the transcript.

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| F9.1 | `go test -count=1 .../delivery/http/ -run DocumentCreateResult\|Duplicate -v` | `--- PASS: TestDocumentCreateResult_WireContract`; `--- PASS: TestDuplicateDocument_InternalError_DoesNotLeakDetail` | ✅ |
| F9.2 | `go test -count=1 .../delivery/http/ -run DocumentCommentResponse -v` | `--- PASS: TestDocumentCommentResponse_WireContract` (unresolved 8 keys; resolved adds parent_library_id,resolved_at) | ✅ |
| F9.3 | `go test -count=1 .../delivery/http/ -run RevisionUrl -v` + FE `npx tsc --noEmit -p tsconfig.build.json` | `--- PASS: TestRevisionUrlResponse_WireContract` (`{"url":"..."}`); `TSC_EXIT=0` | ✅ |
| F9.4 | `go test -count=1 ./tools/cilint/internal/analyzers/ -run NoResponseMap -v` | 10/10 PASS incl. `TestNoResponseMap_Positive_MapStringString` (flags `map[string]string`→`WriteJSON`) and `TestNoResponseMap_Negative_NonResponseMapStringString` (non-writer map passes) | ✅ |
| all | `GOFLAGS=-mod=mod go build ./...` | `BUILD_EXIT=0` | ✅ |
| all | `GOFLAGS=-mod=mod go run ./tools/cilint ./...` | `CILINT_EXIT=0` | ✅ |

Finalize wire lock + handler path also re-run green (`TestDocumentFinalizeResult_WireContract`,
`TestFinalizeDocument_*`, `TestFinalizeDocument_ReplayReturnsCreatedAndHeader`). **C2 PASS.**

## C3 — Senior review of the aggregate milestone diff

Reviewed `git diff 58dea742..HEAD` as one unit (26 files; src scoped to documents handler, openapi.yaml,
BE+FE codegen, cilint analyzer+test, documents tests, api-contract.md/documents.md wiki, milestone docs).

- **Handler:** the 4 named sites are typed — `duplicateDocument`→`DocumentCreateResult` (686),
  `toCommentResponse`→`DocumentCommentResponse` (1249, list at 1154), `signedRevisionURL`→`RevisionUrlResponse`
  (1137), `finalizeDocument`→`DocumentFinalizeResult` (630/642, **and** the same typed struct marshaled for
  the idempotency replay at 632 — no second source of truth for the replay body).
- **No split-brain:** OpenAPI is the single contract source; BE `api.gen.go` and FE `index.d.ts` both carry
  `RevisionUrlResponse` + `DocumentFinalizeResult` and the revision-url op is `GetDocumentRevisionUrl200JSONResponse`
  (200, not 302) — generated, not hand-edited. The wiki notes (api-contract §5b, documents.md:258) match runtime.
- **No dead code:** local `commentResponse` struct removed (not orphaned); helper `parseCreateResultUUIDs` /
  `decodeCommentContent` each single-purpose.
- **No feature broke another; no scope drift:** the only "extra" site (`finalizeDocument`) is explicitly the
  F9.4-row class-closure the widening exists to force, recorded with rationale in F9.4 evidence/plan — not
  unplanned. Test fakes switched to real UUIDs (`inst_1`→`2222…`, `inst-test`→`3333…`) because the typed body
  now `uuid.Parse`s InstanceID; correct accommodation of runtime truth, not a patch.
- **Borderline reviewed-clean:** `placeholder_options_handler.go:87` `map[string]string` is boxed via
  `selectOptions`→`toAnySlice` into the **typed** `PlaceholderOptionsResponse{Options []any}` envelope (the M7
  F7.4 polymorphic pattern); the body passed to `writeFillInJSON` is a typed struct, so it is not an H-D
  response literal — correctly not flagged, wire parity proven by `TestPlaceholderOptionsResponse_BothBranchesParity`.

Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api: typed-body, error-envelope, contract-first regen) | pass | F9.3 followed OpenAPI→BE codegen→FE codegen order; typed bodies on all 4 sites; problem+json error paths preserved |
| §5b Part A grep (full ROUTE_PATHS) | pass | only `health.go:24/33` returned — the recorded file-level exemption (`noResponseMapExemptFiles`); 0 non-exempt response literals |
| §5b Part B grep (every survivor on non-response allowlist) | pass | all survivors are allowlist categories: `recordAudit` params, `formData`/`ContentFormData`/`formMap` command-inputs, declared-dynamic metrics + health, presence/iam internal lookup maps, security `Evidence` domain-mirror, templates mapping helpers — none reach a 2xx writer untyped (cilint exit 0 corroborates) |
| Regression vs prior milestones (M0–M8) | all still pass | `go test -count=1 ./...`: 115 ok/no-test packages, **0 FAIL**, 0 panic; FE `tsc` exit 0 |

**C4 PASS.**

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before (post-M8) | After (post-M9) | Root-cause-fixed evidence |
|-------------|------------------|-----------------|---------------------------|
| 3 post-M8 contract Majors (untyped `map[string]string` on public routes) | 3 confirmed Majors, Contract/API B+ | typed bodies on all 3 sites (duplicate/comment/revision-url) | `handler.go:686/1249/1137` emit generated types; wire-contract locks pin the OpenAPI key sets; not symptom-patched (no string-shape masking) |
| H-D class by §8 intent | 3 (analyzer `map[string]any`-only let `map[string]string` through) | 0 by intent — analyzer flags any `map[string]<T>` | `noresponsemap.go isMapStringLiteral` value-type check removed; `TestNoResponseMap_Positive_MapStringString` proves it catches the exact evasion; the 4th hidden site (`finalizeDocument`) was surfaced by the widened gate (evidence row: pre-fix `go run ./tools/cilint` exited 1 at handler.go:636) and **closed by typing, not suppressed** |

Root cause (each gate names one shape, next read finds the adjacent shape) is closed at the **class** level:
the gate is now type-agnostic for `map[string]<T>` reaching a 2xx writer, laundering-resistant for
built-then-written locals and writer aliases, and full-surface scoped. Could it be built better? The full
codegen-first StrictServerInterface rewire would make typed bodies structurally unavoidable rather than
analyzer-enforced — but that was the operator's explicitly-declined option (HS-5/HS-2, milestone.md
Appetite); the §8 bar is grade/Major/class, not a framework, so the current construction is sound. Note for
the terminal re-audit (not a milestone defect): a stale legacy operationId string `GetApiV2…RevisionsRidUrl`
remains in the documents.md:258 row text — cosmetic wiki artifact, pre-existing, generated op is correct.

**C5 PASS.**

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean: each feature's named test re-run and mapped in C2.*
- [ ] Fixture/mock passed off as real-provider proof — *clean: wire-contract tests are real serialization of generated types; handler-path tests exercise the live mux; no fixture claimed as provider.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean: F9.3 read `DocumentEditorPage.tsx:95`; F9.1/F9.2 read the generated OpenAPI models.*
- [ ] Split-brain (one fact, two sources of truth) — *clean: OpenAPI single source; BE+FE codegen generated; replay body uses the same typed struct.*
- [ ] Self-judged close / validator edited or fixed code — *clean: validator wrote only this verdict; status unflipped.*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean: the 4th `finalizeDocument` site is the F9.4-row class-closure, recorded with rationale.*
- [ ] Symptom-patch (bar moved by masking, root cause intact) — *clean: Majors fixed by typing; gate widened at the class.*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass. Code-wise: senior-level, contract-clean, no split-brain, no dead code, no guessed
  contract. Function-wise: end-to-end the 4 sites emit their declared typed bodies at the spec status codes;
  the widened analyzer demonstrably catches the `map[string]string` evasion class; full repo build + tests +
  cilint + FE typecheck green; M0–M8 regression-clean.
- On **PASS** — handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only the main session, only on this PASS
> - Terminal acceptance (post-M9 re-audit + `mission-validator` against §8) is separate from this milestone
>   gate; HS-5 6th-miss rule applies if it still misses.
