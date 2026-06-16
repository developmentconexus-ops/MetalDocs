# Milestone 3 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-16  ·  **Verdict:** see C7 — **PASS**.
> Run only after every feature is closed (each has a complete `evidence.md`). The validator judges and
> writes this file; the **main session flips status only on a PASS**. The validator never edits code,
> fixes findings, or flips status.

## Inputs loaded (fail-fast check)

| Input | Loaded | Note |
|-------|--------|------|
| Milestone spec `../milestone.md` | yes | 5 features F3.1–F3.5, objective, bar, rabbit holes, HS table |
| F3.1 spec/plan/evidence | yes | all present |
| F3.2 spec/plan/evidence | yes | all present |
| F3.3 spec/plan/evidence | yes | all present |
| F3.4 spec/plan/evidence | yes | all present |
| F3.5 spec/plan/evidence | yes | all present |
| Program `../README.md` (status table, terminal acceptance) | yes | M0/M1/M2 passed; M3 in-progress |
| Governing spec `../mission.md` (linked) | referenced via README + milestone.md §5 E1–E6 | |
| Aggregate milestone diff (`22a80208..HEAD`) | yes | 16 commits; 30 files; +1706/−47 |
| Prior-milestone QA verdicts (M0/M1/M2) | yes | read for regression-sentinel definitions |

No input missing or unreadable. Proceeding to judge.

## C1 — Spec & plan conformance (per feature)

All five features carry an approved-before-code `spec.md` (Approved line filled: `2026-06-16 / leandrotca.work` on F3.1; `Approved — 2026-06-16` on F3.2–F3.5), a populated interview / pre-spec investigation record, an execution-shaped `plan.md`, and an `evidence.md` whose acceptance table maps row-for-row to the spec Validation Gate. Consumer contracts were honored producer-side (verified by reading the consumer sites, not trusting the prose).

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F3.1 | yes — `IAMUserOptionsReader.ListUserOptions(ctx, tenantID) ([]UserOption, error)` at `internal/modules/documents/application/iam_user_options.go:11-12` matches the adapter `wiring.DocumentsIAMUserOptions` exactly (filter IsActive, sort lower(DisplayName)/UserID, non-nil empty, error-propagate) | yes — gates 1–9 met (5 unit subtests, runtime real proof labeled `real`, wire grep, regression, N/A authz) | yes — no port shape change, no IAM API change, nil-safe branch at module.go kept | `f3.1/evidence.md` |
| F3.2 | yes — both `NewFreezeService` callers (`apps/api/.../main.go:791`, `apps/worker/.../main.go:105`) already passed `*fanout.Client` which satisfies the now-typed `FanoutClient`; all `documentComments` interface impls + test fakes updated to 3-param `ListDocumentComments` | yes — gates 1–6 met; authz-scope check recorded (userID never forwarded to repo; delivery gate `authorizeDocumentScope` unchanged) | yes — repo method, authz logic, other FreezeService surface untouched | `f3.2/evidence.md` |
| F3.3 | yes — `Pin` now calls narrow `ReadFreezeAt`; `Freeze`/`Materialize` keep `ReadSnapshotWithFreezeAt` and still consume `snap`; `SnapshotReader` (documents-owned app interface) extended by one method | yes — gates 1–7 met | yes — freeze pipeline otherwise untouched; ReadSnapshotWithFreezeAt unchanged | `f3.3/evidence.md` |
| F3.4 | yes — objectstore key catalog: dead exports removed; live `TemplatesPresigner` (different format, active callers) untouched | yes — gates 1–3 met; 0 Go refs | yes — only template_keys.go + its test deleted | `f3.4/evidence.md` |
| F3.5 | yes — §7 boxed universe (4 items) each closed-with-cite or deferred-with-trigger+owner; nothing silently skipped | yes — gates 1–4 met | yes — no i18n infra, no cross-module extraction, no New-caller change, no sig_ format change | `f3.5/evidence.md` |

**C1 result: PASS.** Every feature in `milestone.md`'s Features table has all three artifacts with the required structure; consumer contracts read from the consumer side; non-goals respected.

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (`-count=1` on touched packages; not trusted from transcripts). Working dir `C:\Users\leandro.theodoro\Documents\MetalDocs`.

| Feature | Command re-run | Real output | Pass? |
|---------|----------------|-------------|-------|
| all | `go build ./...` | clean exit, BUILD_EXIT=0 | yes |
| F3.1 | `go test -count=1 -run TestDocumentsIAMUserOptions -v ./apps/api/internal/wiring/` | `--- PASS: TestDocumentsIAMUserOptions` + all 5 subtests PASS; `ok metaldocs/apps/api/internal/wiring 1.737s` | yes |
| F3.1 | `grep IAMUserOptions apps/api/cmd/metaldocs-api/` | `main.go:429: IAMUserOptions: wiring.NewDocumentsIAMUserOptions(authService),` | yes |
| F3.2 | `grep 'fanoutClient any\|fc, _' freeze_service.go` | No matches | yes |
| F3.2 | source diff | `ListDocumentComments(ctx, tenantID, documentID string)` 3-param in service.go:434, handler.go:70 interface, handler.go:1116 callsite | yes |
| F3.3 | `grep '_ = snap' freeze_service.go` | No matches | yes |
| F3.3 | source diff | Pin → `ReadFreezeAt`; repo adds narrow `SELECT values_frozen_at` (snapshot_repository.go); same QueryRow/Scan + not-found semantics as old path | yes |
| F3.4 | `grep -RIn 'TemplateDocxKey\|TemplateSchemaKey' --include=*.go` + `ls template_keys.go` | No matches; file absent | yes |
| F3.5 | `grep -n Deprecated service.go` | `127: // Deprecated: use NewService.` | yes |
| F3.5 | `grep -n 'sha1\|sha256' security/application/service.go` | `14: "crypto/sha256"`, `180: sha256.New()` (no sha1) | yes |
| all | `go test -count=1 ./internal/modules/documents/... ./internal/modules/security/... ./apps/api/internal/wiring/...` | every package `ok`, 0 fail | yes |

No gate failed on isolated re-run; nothing flaky or environment-coupled surfaced.

> **Runtime-real note (F3.1 gate 6):** the evidence labels the placeholder-options body `real` (live dev API + seeded admin tenant), explicitly disclosing the smoke doc was SQL-inserted because no seed doc had a user-type placeholder — identical handler code path, honestly labeled. The validator did not re-run the live API smoke (requires a running dev stack), but the disclosure is honest and the 5 fixture unit subtests + the wired composition-root line independently prove the adapter contract and its injection. This satisfies milestone §1's "real or explicitly-labeled in-process fake against the real adapter contract" bar — it is not a pure mock bypassing the adapter.

## C3 — Senior review of the aggregate milestone diff

Whole-milestone diff (`22a80208..HEAD`) reviewed as one unit. Source files: freeze_service.go (E2+E4), service.go (E3+E6-deprecate), handler.go (E3), snapshot_repository.go (E4), security/service.go (E6 sha256), template_keys.go (+test) deleted (E5), main.go (E1 wire), plus test-fake updates.

- **Duplication / split-brain:** none. E1's adapter is the single production source for the consumer port; the nil-safe branch in module.go is documented defense-in-depth, not a competing source of truth. The generated `api.gen.go`/`routes_generated.go` `ListDocumentComments` HTTP wrapper is the unchanged transport contract and is independent of the internal 3-param service method — no two-sources-of-truth.
- **Dead code:** removed, not added. E2 deletes the `fc, _ := ...(FanoutClient)` runtime cast and replaces it with a compile-time typed param + package-scope guard `var _ FanoutClient = (*fanout.Client)(nil)`. E3 removes the dead `userID`. E4 removes the fetch-then-discard. E5 deletes the dead exports.
- **One feature breaking another:** none. F3.2's typed param and F3.4's deletion both compile clean (`go build ./...` = 0); F3.3's `ReadFreezeAt` preserves the old not-found semantics (same `QueryRow().Scan()` returning `sql.ErrNoRows`), so Pin idempotency is unchanged.
- **Guessed contracts:** none. F3.1 read the consumer port; F3.2 read the method body + both production callers before retyping/removing; F3.3 read all three call sites; F3.4 verified zero callers (direct + string-form).
- **Right-seam fixes (milestone quality goal #1):** E1 at composition root (main.go:429), E2 at constructor signature, E3 at method signature, E4 at the read site (new narrow query), E5 by deletion (not a deprecation comment) — exactly as milestone §4 demands.

- Findings: none blocking. **Staff-engineer bar met? yes.**

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (`wiki/quality/backend-api-qa-checklist.md`, code-quality lens) | pass | typed dependency at constructor seam (E2), no `any` smuggling at the public application boundary, no dead parameter (E3), no dead export (E5), no fetch-then-discard (E4); test-framework gate honored (F3.1 unit test is table-driven + function-typed in-memory fake + UUID fixtures, no sloppy strings; touched fakes drive-by migrated) |
| Whole-repo regression `go test ./...` | all pass | 85 packages `ok`, 0 `FAIL` (force-fresh `-count=1` on touched modules; remainder cached/clean) |
| Regression vs M0 (authz/session) | holds | F3.2 R1 security check: `authorizeDocumentScope(w, r, docID)` still runs before the service call (`tenantID, _, ok :=` then early return on `!ok`); removed `userID` was never an authz input. No authz path weakened. |
| Regression vs M1 (contract-integrity, H-D) | holds | M3 diff adds **zero** new `map[string]any` in any `*/delivery/http/*.go` (grep of `^+` added lines = none). M1's named DRIFT count (4→0) is untouched; F3.1 evidence's "40 pre-existing raw-grep hits" are the M1-scoped-out strict-literal set, correctly attributed (F3.1 added 0). |
| Regression vs M2 (observability) | holds | M2 sentinel `grep -RIn NewTextHandler internal/modules/jobs/` = 0 (slog composition unchanged); composition-root injection unchanged except the single additive E1 wire line. |

## C5 — Quality-bar re-measure + retrospective

Milestone bar (milestone.md §Bar): the six F5.1 §6 code-quality / dead-code findings fixed **at root**, not symptom-patched, with runtime proof for E1 and build/test proof for E2–E5, and a complete §7 row table for E6.

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| E1 IAMUserOptions never wired (functional) | placeholder picker returns empty list | wired at composition root; adapter returns active, sorted, tenant-scoped users | `main.go:429` wire line + adapter + runtime body (labeled real) `[{user_id,display_name}×4]` sorted ASC |
| E2 `fanoutClient any` | runtime cast `fc, _ := ...(FanoutClient)` | typed param + compile-time guard | grep `fanoutClient any\|fc, _` = 0; `var _ FanoutClient = (*fanout.Client)(nil)` |
| E3 dead `userID` param | lying 4-param signature | 3-param honest signature; authz gate proven independent | service.go:434 + handler.go:70/1116; authz-scope check recorded |
| E4 `_ = snap` discard | 4-column SELECT, 3 discarded | narrow `SELECT values_frozen_at` on Pin path | grep `_ = snap` = 0; new `ReadFreezeAt`; Freeze/Materialize untouched |
| E5 dead exported keys | 2 unused exports | file deleted | grep `TemplateDocxKey\|TemplateSchemaKey` = 0; `template_keys.go` absent |
| E6 §7 Minor tail | 4 open minors | 2 closed (Deprecated New, sha1→sha256) + 2 deferred (PT i18n, tenantIDFromRequest→M4) with trigger+owner | close-out table; grep Deprecated + sha256 confirm; defers verified live (PT const present, 3 tenantIDFromRequest sites) |

Root cause fixed at the correct seam in every case; no symptom-patch (no deleted handler, no silenced grep, no muted test, no swallowed error). E5 is a real deletion, not the forbidden "deprecated comment" shortcut.

- Could it be built better? Minor, non-blocking: (1) the three `tenantIDFromRequest` copies remain duplicated — correctly deferred to M4 with a named owner, the right call given M3's surgical scope. (2) F3.1 runtime smoke relied on a SQL-inserted doc due to a seed-data gap; a seeded fixture doc with a user-type placeholder would make future re-validation reproducible without manual SQL (recorded as F3.1's own bounded defer). Neither indicates an unsound current construction. No new defer required from the validator.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean: every feature has a row-for-row acceptance table; suite-green is in addition to, not in place of, per-feature mapping.*
- [ ] Fixture/mock passed off as real-provider proof — *clean: F3.1 explicitly distinguishes the 5 fixture unit subtests from the `real`-labeled runtime body, and discloses the SQL-inserted smoke doc.*
- [ ] Consumer contract guessed rather than read from the consumer — *clean: each feature read the consumer/caller before changing the producer.*
- [ ] Split-brain (one fact, two sources of truth) — *clean.*
- [ ] Self-judged close / validator edited or fixed code — *clean: validator wrote only this verdict file; no source touched; status not flipped.*
- [ ] Scope drift (work beyond the spec, no rationale) — *clean: the M3 commit range touches exactly the 30 planned files. Untracked working-tree files (otel test stubs, smoke logs, M4 milestone artifacts, vendor dirs) are NOT in the M3 commit set — they are pre-existing clutter from other sessions, outside this milestone's diff.*
- [ ] Symptom-patch (bar moved by masking, root cause intact) — *clean: see C5; E5 is a real deletion, not a deprecation comment.*

(All unchecked = clean.)

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (senior-grade, contract-clean, no split-brain, no dead code, no guessed contracts; every fix at the right seam) and **function-wise/QA** (every per-feature Validation Gate GREEN on isolated re-run; E1 functional bar proven; whole-repo regression 85/0; M0/M1/M2 sentinels all hold).
- On **PASS** — handed back to the main session to flip status and present the HS-1 operator gate.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending
> - Status flipped in `README.md`: no — only on PASS + operator HS-1 approval
