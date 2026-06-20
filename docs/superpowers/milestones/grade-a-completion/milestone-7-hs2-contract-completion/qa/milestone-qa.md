# Milestone 7 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-20  ·  **Verdict:** see C7 (**PASS**).
> **Base before M7:** `45a03fa6` · **HEAD validated:** `dadb8275`.
> The validator judges and writes this file only; it edits no source, fixes no findings, flips no status.

## Inputs loaded

- Milestone spec `../milestone.md` (validation definition §2, hard-stops, HS-6 F7.4 extension).
- All 5 features' `spec.md` + `plan.md` + `evidence.md` (F7.1–F7.5) — all present and complete.
- Program `../README.md` (status table, hard-stops, terminal acceptance) + governing `../mission.md` chain.
- Aggregate M7 diff `git diff 45a03fa6..HEAD` (32 files, +1693/-168) and `git log 45a03fa6..HEAD`
  (5 commits: `12ff8fa1` F7.1, `142f6d07` F7.2, `16d2a9f5` F7.3, `89651dfb` F7.4, `dadb8275` F7.5).

No input missing or unreadable. (Untracked `f2.x-smoke*.log` files in the worktree are unrelated to M7
and outside the diff range — noted, not a blocker.)

## C1 — Spec & plan conformance (per feature)

Every feature carries `spec.md` (with a filled pre-commit approval line + a populated interview record),
an execution-shaped `plan.md`, and an `evidence.md` whose acceptance rows map row-for-row to the spec
Validation Gate. Consumer contract honored = producer matches the OpenAPI/consumer source of truth, verified
by reading the code (not trusting evidence).

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F7.1 audit | ✅ export-status emits `auditapi.AuditExportStatusResponse` (handler.go:261–274), list/export emit generated types; wire keys = OpenAPI projection | ✅ all 6 gates re-run green (C2) | ✅ hand-rolled mux kept, no rewire, no OpenAPI/codegen change | `f7.1-audit-typed/evidence.md` |
| F7.2 auth | ✅ `authLoginResponse`/`changePasswordResponse` JSON tags = OpenAPI `AuthLoginResponse`/`ChangePasswordResponse`; `expires_at` keeps `Format(time.RFC3339)` string → byte-identical | ✅ 5 gates green | ✅ pre-codegen ADR-0012 posture; 4 `recordAudit` maps kept (non-response) | `f7.2-auth-typed/evidence.md` |
| F7.3 search | ✅ `searchDocumentsResponse{Items}` = OpenAPI `SearchDocumentsResponse`; same non-nil `out` slice | ✅ 5 gates green incl. empty→`[]` | ✅ pre-codegen ADR-0012; item shape untouched | `f7.3-search-typed/evidence.md` |
| F7.4 documents | ✅ 4 generated models match 4 new OpenAPI 200 schemas; HS-6 `pdfCompleteResponse` matches prior `{document_id, final_pdf_s3_key}` literal | ✅ 6 gates green; 5 wire-parity tests verbose-PASS | ✅ opaque-items envelope (no over-modeling); no rewire; pdf_webhook off-spec preserved | `f7.4-…/evidence.md` |
| F7.5 gate+proof | ✅ honest two-part gate documented in `api-contract.md` §5b (Part A + Part B + non-response allowlist + anti-evasion); FE codegen scoped to F7.4's 4 schemas | ✅ 8 gates green | ✅ docs + FE regen + proof only; no allowlisted map converted | `f7.5-…/evidence.md` |

One **cosmetic discrepancy noted, not failing**: F7.1 spec Gate #1 says baseline grep was "5 hits → 1";
F7.1 evidence says "6 → 1". The post-state (1 surviving non-response decode buffer at audit handler.go:404)
is what I independently verified — the baseline-count phrasing is an artifact of whether the removed
`EventResponse.Payload` field is counted. Does not affect the gate outcome.

C1 = **PASS**.

## C2 — Gates re-run, isolated (validator-run, not trusted from evidence)

The binding two-part H-D gate + build + full suite + per-feature named tests, all re-run from clean state:

| Check | Command (validator-run) | Real output | Pass? |
|-------|--------------------------|-------------|-------|
| Part A | `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| grep -v _test.go` | exit 1, **0 hits** | ✅ |
| Part B | `grep -rEn 'map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| grep -v _test.go` | **11 hits, all non-response** (allowlist re-derived from first principles — see C5) | ✅ |
| H-G Grep B | `grep -rEn '"published"' internal/modules/ … \| grep -v _test.go \| grep -v '/domain/' \| grep -v 'api.gen.go'` | 7 hits, **all doc-comments** (no status/value literal) — consistent with post-M6 audit §6 | ✅ |
| H-G Grep C | `grep -rEn 'FROM …iam_user_roles' … \| grep -v 'internal/modules/iam/'` | exit 1, **0 hits** | ✅ |
| BE codegen fresh | `GOFLAGS=-mod=mod go generate ./internal/modules/documents/api/...` then `git diff --exit-code …/api.gen.go` | generate exit 0; **CLEAN, no uncommitted diff** | ✅ |
| Build | `GOFLAGS=-mod=mod go build ./...` | exit 0 | ✅ |
| Full suite | `GOFLAGS=-mod=mod go test -count=1 ./...` | exit 0 · **85 ok · 0 FAIL · 0 panic** | ✅ |
| F7.1 named tests | `go test -run 'TestAuditHandler_ExportStatusTypedShape\|…ExportPOSTTypedShape\|…ListEventsTypedShape\|TestHandleExport_405_Allow\|TestHandleExportSubresource_405_Allow' ./…/audit/delivery/http/` | `ok` | ✅ |
| F7.2 named tests | `go test -run 'TestAuthLoginResponse_WireContract\|TestChangePasswordResponse_WireContract' ./…/auth/delivery/http/` | `ok` | ✅ |
| F7.3 named tests | `go test -run 'TestSearchDocumentsResponse_WireContract\|…_EmptyIsArrayNotNull' ./…/search/delivery/http/` | `ok` | ✅ |
| F7.4 named tests | `go test -v -run 'TestFillInSchemaResponse_EnvelopeAndEmptyParity\|TestPutPlaceholderValueResponse_SecondPrecision\|TestViewDocumentResponse_ConditionalKeys\|TestPlaceholderOptionsResponse_BothBranchesParity\|TestPDFCompleteResponse_WireContract' ./…/documents/delivery/http/` | **5/5 RUN + PASS** | ✅ |

The F7.4 wire-parity tests were read and confirmed non-vacuous: they assert exact byte strings —
`"updated_at":"2026-06-20T12:00:00Z"` (second precision from `Truncate(time.Second)`),
`{"data":{"placeholder_schema":[]}}` (empty→`[]` not null), conditional view key sets
`pdf_status` vs `pdf_status,pdf_url,signed_url`, and boxing wire-neutrality for both option branches.

C2 = **PASS**.

## C3 — Senior review of the aggregate milestone diff

Reviewed `git diff 45a03fa6..HEAD` as one unit:

- **Scope is exactly the cited surface** — source changes only in audit/auth/search/documents
  `delivery/http/` + `documents/api/api.gen.go` (regenerated) + `openapi.yaml` (+4 schemas) + FE
  `index.d.ts` (regenerated) + `wiki/architecture/api-contract.md` (§5b gate). Templates and IAM
  (M6's typed responses) are untouched (`git diff --name-only … internal/modules/templates/ internal/modules/iam/` non-test = none).
- **Every response writer passes a typed value** — all `writeJSON`/`writeFillInJSON`/`WriteJSON` call
  sites in the swapped files pass a generated `documentsapi.*`/`auditapi.*` or hand-rolled
  `authLoginResponse`/`changePasswordResponse`/`searchDocumentsResponse`/`pdfCompleteResponse`/`user` —
  no map literals, no `any(map[...])` laundering (grep for `any\(map\[string\]` = 0).
- **No dead code / no split-brain** — `EventResponse` type genuinely removed (no type/literal refs
  remain; `buildEventResponses` retargeted to `[]auditapi.AuditEventItem`; a GitNexus index hint to the
  contrary is stale and was disproven by direct grep). The `toAnySlice` helper is a single generic boxer,
  no duplication. Source of truth for each shape is single (OpenAPI for codegen modules; the hand-rolled
  struct for pre-codegen/off-spec).
- **HS-2 boundary held** — no `NewStrictHandler`/`ServerInterface`/`HandlerFromMux`/`RegisterHandlers`
  wiring added; documents still routes via hand-rolled `mux.HandleFunc`; auth/search stay pre-codegen.
- **Contract-first order intact** — `openapi.yaml` changed alongside `api.gen.go`; FE `index.d.ts`
  reflects only the 4 documents schemas (`unknown[]` opaque arrays, `?` optional signed_url/pdf_url,
  string `updated_at`). No codegen committed without a matching source change.

Staff-engineer bar met? **✅**

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (backend-api / contract) | pass | Typed bodies on every public 200; OpenAPI declares the bodies for codegen ops; build + full suite green; codegen fresh BE + FE. |
| Regression vs M0–M6 | all still pass | Full `go test -count=1 ./...` = 0 FAIL (M0–M6 sentinels included). H-G Grep B/C = 0 (no F5.1/F5.2/F5.7 regression). Templates/IAM M6 typed responses untouched. Diff confined to cited modules. |

C4 = **PASS**.

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before (post-M6, HEAD `5650b328`) | After (HEAD `dadb8275`) | Root-cause-fixed evidence |
|-------------|-----------------------------------|--------------------------|---------------------------|
| H-D response-literal count | 10 surviving (via `writeFillInJSON` alias + multi-line maps Grep A is blind to) | **0** | Part A = 0 **and** Part B = 11 hits all independently re-classified as non-response: audit:404 decode buffer (feeds `AuditEventItem.Payload`); auth:98/109/127 + :204 `recordAudit` params (marshalled to audit emit, never written to response — verified at recordAudit body); iam people_handler:321/:454 same; controlleddocuments routes:89/:224 `formData` command input from request; documents handler:615 `ContentFormData` command input to a service request struct; security:54 `signalItem.Evidence` struct field decl. **Zero response literals.** Root cause (untyped response bodies) fixed by emitting typed structs, not masked. |
| Confirmed Major (audit export-status) | 1 (generated `AuditExportStatusResponse` existed but unused) | **0** | handler.go:261–274 now builds + emits `auditapi.AuditExportStatusResponse` with conditional `ExpiresAt`/`Error` pointers; `TestAuditHandler_ExportStatusTypedShape` strict-decodes it. |
| Gate blindness (the M6 defect) | Grep A alone (one-liner-blind) | **Fixed** | `api-contract.md` §5b documents the honest two-part gate + non-response allowlist + an explicit anti-evasion clause forbidding response-shaped allowlist entries and helper-hiding. Future milestones inherit the honest measurement. |

The root cause is fixed, not symptom-patched: the gate was made *harder* (Part B completeness) and still
reads 0, and the typed bodies are proven byte-parity, not grep-evasion.

**Could it be built better?** The opaque `items: {}` → `[]interface{}` envelopes for documents fill-in /
placeholder-options are a deliberate, recorded proportionality call (spec Q8) — they kill the H-D defect
(the untyped *envelope*) without re-modeling a 17-field nested domain type. A future mission that wants
full FE type-safety on those polymorphic items could faithfully model `Placeholder`/`UserOptionView` in
OpenAPI — that is a successor enhancement, not an M7 defect. Recorded as next-mission input; **does not
FAIL** this milestone (the current construction is sound and within the operator's typed-body-parity scope).

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean: per-feature named tests re-run and mapped in C2.*
- [ ] Fixture/mock passed off as real-provider proof — *clean: gates are real greps + real `go build`/`go test`; wire-parity tests assert exact bytes.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean: contracts read from OpenAPI / generated projection; producer verified against them.*
- [ ] Split-brain (one fact, two sources of truth) — *clean: single source per shape; `EventResponse` retired so no twin.*
- [ ] Self-judged close / validator edited or fixed code — *clean: validator wrote only this verdict; no source edited, no status flipped.*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean: diff confined to cited surface; the only extension (HS-6 pdf_webhook) is operator-approved and recorded in milestone.md F7.4 + F7.4 evidence.*
- [ ] Symptom-patch (bar "moved" by masking, root cause intact) — *clean: gate made stricter (Part B) and still 0; no `any(map[...])` laundering (grep = 0).*

All unchecked = clean.

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass independently: **code-wise** (typed bodies, no dead code, no split-brain, HS-2
  boundary held, contract-first order intact) and **function-wise** (build + 85-package full suite green
  from clean state; H-D = 0 under the honest two-part gate; the one confirmed Major closed; per-feature
  wire-parity tests assert real byte-level invariants).
- Handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only the main session, on this PASS
