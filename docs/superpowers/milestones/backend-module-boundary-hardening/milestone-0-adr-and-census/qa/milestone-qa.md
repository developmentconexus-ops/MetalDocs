# Milestone 0 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `spec.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-20  ·  **Verdict:** see C7 — **PASS**.
> The validator judged and wrote only this file; it edited no source, fixed no finding, flipped no status.

## Inputs loaded (none missing)

- Milestone spec `../milestone.md`; mission `../mission.md` (§3 D1–D6, §5, §8, §9); program `../README.md`; discovery brief referenced.
- All three features' `spec.md` / `plan.md` / `evidence.md` (F0.1/F0.2/F0.3); F0.1 evidence Addendum (HS-4 amendment).
- F0.2 `census.md` + `hs-6-scope-decision.md`.
- Governing artifact `wiki/decisions/0039-cross-module-base-table-read-boundary.md` + `wiki/decisions/index.md`.
- Aggregate diff: ADR-0039 (new), decisions index (2 rows), the cilint analyzer `hgcrossmodule.go` + `_test.go` (new), `analyzers.go` (one line wired). No production SQL touched — confirmed by the guard's still-full pending ledger and clean `go build`.

## C1 — Spec & plan conformance (per feature)

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F0.1 adr-0039 | ✅ — ADR D1/D3 rule is the exact mechanical definition F0.2 census + F0.3 guard consume; producer matches consumer | ✅ — file exists, Status `Accepted 2026-06-20 (amended)`, worked table classifies all §5 rows 3–15 with 0 unclassified, D4 exemptions named, `effective_to IS NULL` cited (ADR 0037), H-PRE-1 note, links to 0022/0037/0038 | ✅ — one wiki md file only; no Go/SQL/view/port; ADR 0037 not reinterpreted | ADR §D1–D6, worked table; index rows 42–43 |
| F0.2 binding-census | ✅ — census applies ADR-0039 as-is (invents no verdict); produces table→owner map + in-scope list F0.3 guard consumes | ✅ — reproduces the ~20 brief sites at cited lines, widened to full owned-base-table set (9-module owner map), `Unclassified: 0`, brief-delta + coverage statement with named residual; N1+X1–X8 surfaced | ✅ — measured only, zero SQL edited; static read, no false runtime green | `census.md`; spot-checks below |
| F0.3 cilint-guard | ✅ — analyzer encodes census owner map + in-scope (pending) + exempt (X1–X8) sets + ADR-0039 allowlist; wired into `RunAll` | ✅ — 8 `TestHGCrossModule_*` pass (bite + 7 green); full-tree `go run ./tools/cilint ./...` exit 0; `go build`/`go test ./tools/cilint/...` green; baseline load-bearing (drop-B1 → exit 1, re-verified by validator) | ✅ — detection only, no SQL edited; pending ledger left FULL (emptying = false green); literal-token residual recorded | `hgcrossmodule.go`, `_test.go`; C2 below |

All three `spec.md` files carry a filled `Approved before code: 2026-06-20 / leandrotca` line; interview records populated (locked-decision citations, fail-closed gate satisfied); each `plan.md` is execution-shaped (task list, files touched, test strategy); each `evidence.md` acceptance table maps row-for-row to its spec Validation Gate. **C1 PASS.**

## C2 — Gates re-run, isolated (validator ran these, did not trust transcripts)

| Gate | Command re-run | Real output | Pass? |
|------|----------------|-------------|-------|
| Build | `go build ./...` | `BUILD_EXIT=0` (clean) | ✅ |
| HG unit suite | `go test ./tools/cilint/internal/analyzers/ -run TestHGCrossModule -v` | 8/8 PASS (`_Positive_CrossModuleRead`, `_Negative_OwnTable`, `_SubpackageSameModule`, `_CommentMention`, `_PendingBaseline`, `_Exempt`, `_AllowDirective`, `_OutOfScope_NonModuleFile`); `ok metaldocs/tools/cilint/internal/analyzers` | ✅ |
| Full cilint suite | `go test ./tools/cilint/...` | `ok` (no test files in root pkg) `TEST_EXIT=0` | ✅ |
| Full-tree guard | `go run ./tools/cilint ./...` | `CILINT_EXIT=0` — green; all live cross-module reads ∈ pending ∪ exempt | ✅ |
| **Skeptic: green not vacuous** | dropped B1 (`{documents/repository/repository.go, controlled_documents}`) from `hgPendingRemediation`, re-ran guard | `CILINT_EXIT=1` — `internal\modules\documents\repository\repository.go:1701: [hgcrossmodule] module "documents" reads "controlleddocuments"'s base table "controlled_documents" with raw SQL (H-G, ADR-0039 D1)…` fired at the exact real site; file restored, guard back to exit 0 | ✅ |

**C2 PASS.** The drop-B1 experiment independently reproduces F0.3's load-bearing claim: the guard detects the real in-scope reads, so exit 0 means the F0.2 baseline is **complete**, not that the regex matched nothing.

### Census spot-checks (validator re-grepped the live tree)

- B1 `documents/repository/repository.go:1701` → `SELECT profile_code FROM controlled_documents` (CD-owned) — confirmed.
- B2 `controlleddocuments/infrastructure/repository.go:532` → `FROM document_revisions` (documents-owned) — confirmed.
- B8 `iam/.../area_catalog_reader.go:28` → `EXISTS(… FROM metaldocs.document_process_areas)` (taxonomy-owned) — confirmed.
- C3 `documents/approval/repository/postgres_approval_repository.go:1136` → `FROM metaldocs.user_process_areas` in-tx — confirmed.
- C4 `search/.../v2documents/reader.go:69/70/102` → `public.documents` / `controlled_documents` / `user_process_areas` — confirmed.
- N1 `documents/application/fillin_service.go:225` → `FROM templates_template_version … JOIN documents` (templates-owned) — confirmed.
- X1 `security/.../postgres/repository.go:121,185,236` → `FROM metaldocs.auth_identities`; X8 `jobs/stuck_instance_watchdog/job.go:147` → `FROM approval_instances` — confirmed genuine cross-module reads being carved out (not non-violations).

### Ownership-correction verification

- `document_process_areas` / `document_profiles` non-test writes live **only** in `taxonomy/infrastructure/repository.go` (INSERT :491,:527) — confirms census correction "taxonomy, not iam."
- `auth_failure_counters` writes (INSERT :64, DELETE :86) live in `documents/approval/.../postgres_auth_failure_rate_limiter.go` — same module that reads it → census false-positive drop is correct; owner map maps it to `documents`.
- `audit_events` non-test writes live in `audit/infrastructure/postgres/writer.go` — confirms D3(d) platform-sink ownership.

### Independent completeness check (could the guard hide a live read?)

Grepped every owner-map table that appears on **neither** allowlist (`iam_users`, `iam_user_roles`, `document_families`, `document_comments`, `document_profiles`, `templates_template`). Every non-test `FROM`/`JOIN` hit is a **same-module** read (iam←iam, taxonomy←taxonomy, documents←documents, templates←templates) — correctly not flagged (owner == reader). No orphan cross-module read exists outside the recorded sets. The exit-0 green is earned.

## C3 — Senior review of the aggregate milestone diff

- **No split-brain.** The owner map exists once (census prose ↔ `hgOwnerByTable`); the in-scope set exists once (census Part 1 ↔ `hgPendingRemediation`); the exemptions exist once (ADR D3(d)–(f) ↔ census X1–X8 ↔ `hgExempt`). All three derive from the single ADR-0039 rule. The "table→owner" fact is duplicated between census markdown and the Go map, but that is a definition-and-its-mechanization (intended), not two competing sources of truth.
- **No dead code / no superseded approach.** The analyzer mirrors the established `noResponseMap` sibling (AST-BasicLit scan, slash-normalized path match, inline-allow directive, shared `parseFile`/`readSource`/`getLine`). Clean reuse.
- **No feature broke another.** Other cilint analyzers' findings unchanged (full-tree exit 0; suite green).
- **Minor (non-blocking) finding.** `hgListed` matches `(fileSuffix, table)` but does not re-assert the reader module equals the site's expected reader. Today every pending/exempt file is unique to one module so no false-suppression occurs; this is a coarseness defer F0.3 already records (file+table allowlist). Recorded as input to M2–M4, not a gate failure.
- Staff-engineer bar met? ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Contract/architecture-truth checklist (definition+gate milestone) | pass | ADR follows docs governance (status from canonical vocab, next-free number 0039, index synced, `Last verified` stamped, supersession/links). No runtime claim made for a static result (correctly noted). |
| Code-quality checklist (F0.3 analyzer) | pass | `go build` clean, `go test ./tools/cilint/...` ok, `gofmt`-shaped, sibling-pattern consistent. |
| Regression vs prior milestones | all pass | M0 is first in this program. Cross-program: existing analyzers (`noresponsemap`, `nosqltxindomain`, `txownership`, …) still green — adding `hgcrossmodule` did not change their findings (full-tree exit 0). `go build ./...` not regressed. |

## C5 — Quality-bar re-measure + retrospective

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| H-G done-bar: definition locked + CI-enforced | implicit/disputed ("14" vs "~20", 2 ownership errors) | ADR-0039 Accepted + machine-checkable | Root cause (no machine-checkable definition) fixed by the analyzer, not symptom-patched: the count is not hand-asserted — `go run ./tools/cilint ./...` deterministically decides it; drop-B1 proves it bites at the real line; completeness check proves no orphan read. |

- Could it be built better? The terminal "H-G=0 under both readings" was honestly re-defined as "0 violations outside the recorded allowlist" once D3(d)–(f) carve-outs were added — this is the correct, non-deceptive reconciliation (carve-outs enumerated + justified, not pretended absent). One refinement for later: make the allowlist key reader-module-aware (see C3 minor finding) so a future file rename can't silently widen suppression. Defer, not a blocker.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean (per-feature acceptance mapped in C1/C2; the green is per-site-baselined, not a bare suite-green).* 
- [ ] Fixture/mock passed off as real-provider proof — *clean (unit tests labeled fixture/bite; the full-tree run + drop-B1 + spot-checks are real live-tree, labeled real; M0 has no runtime/Docker step and none is claimed).* 
- [ ] Consumer contract guessed — *clean (contracts read from locked mission decisions D1–D6 + ADR + census, cited).* 
- [ ] Split-brain — *clean (one fact per source; definition↔mechanization is intended).* 
- [ ] Self-judged close / validator edited code — *clean (validator restored the one experimental edit; flipped no status).* 
- [ ] Scope drift — *clean (N1 + X1–X8 surfaced via HS-6, operator-ruled 2026-06-20, ADR amended, mission §5 row 16 + §2 replanned; nothing added silently). Non-empty pending ledger is correct-for-M0, not drift.* 
- [ ] Symptom-patch — *clean (bar moved by a machine check, root cause fixed).* 

All unchecked = clean.

### Non-blocking hygiene note (not a gate failure)

The program `README.md` "Hard-stops raised" table is still empty though HS-6 was raised and resolved this milestone. HS-6 is fully recorded in `census.md`, `hs-6-scope-decision.md` (operator ruling 1a+2a), the ADR amendment, and the mission replan — so the record is complete; only the README index row is unbackfilled. Recommend the main session add the HS-6 row when it flips M0 status. Does not affect the verdict.

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass. **Code-wise:** the analyzer is senior-level, contract-clean, no split-brain, no dead code; ADR + census + guard are one coherent definition→mechanization chain. **Function-wise:** the guard does end-to-end what M0 promised — it green-flags a complete, accurate baseline and bites at the exact real site when an entry is dropped (validator-reproduced), the census is 0-unclassified with every spot-checked site confirmed against the live tree, and ADR-0039 classifies every §5 site mechanically. M0's done-bar (definition + census + CI guard, **no porting**) is met; the full pending ledger is correct for M0.
- Handed back to the main session to flip status and present the HS-1 operator gate (which also carries the ADR-0039 operator ratification, deferred by design per F0.1).

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): pending — includes ADR-0039 ratification (F0.1 bounded defer) and the HS-6 ruling on record.
> - Status flip in `README.md`: pending PASS+HS-1 (and backfill the HS-6 row — C6 hygiene note).
