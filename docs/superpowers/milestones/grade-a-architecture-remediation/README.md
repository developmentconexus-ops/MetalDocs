# Program: Grade-A Architecture Remediation

> **Governing spec:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`
> **Status:** In progress (M0/M1/M2/M3 passed; M4 features F4.1–F4.6 built but **close gate blocked**; M4b F4b.1/F4b.2 done — legacy `metaldocs.documents` cluster dropped (`071931c9`); the remaining seed-fix was found to be a **test-architecture defect**, so milestone **4c** (unified test-fixture framework) is opened to close the M4 blocker at root cause, then re-dispatch the M4 validator)
> **Owner / operator:** leandrotca.work (operator) + backend agent (Opus 4.8)

Take the backend's three formerly-C audit dimensions (module-boundaries/DDD, contract/API,
composition/observability) to **Grade A−/A**, and fully close the **H-D class** (handler/contract
field drift; tri-source route drift) and the **H-G class** (cross-module reach-without-a-port +
hardcoded domain state) — not just the instances. Every fix carries evidence; symptom-patching is a
hard-stop. **Terminal acceptance:** the M5 independent multi-agent re-audit passes the §6 pass bar
(3 dimensions ≥ A−, 0 new Critical/Major, H-D and H-G classes at 0) and the operator signs off Grade A.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-docs-destaling` | One unambiguous progression surface; stale docs stop polluting agent context | passed | [PASS](milestone-0-docs-destaling/qa/milestone-qa.md) |
| 1 | `milestone-1-reach-a-blockers` | Close all 4 Grade-A blockers + the error-contract (bare-405) tail | passed | [PASS](milestone-1-reach-a-blockers/qa/milestone-qa.md) |
| 2 | `milestone-2-contract-tail` | Eliminate handler-emits-undeclared-field drift (H-D), one FE regen | passed | [PASS](milestone-2-contract-tail/qa/milestone-qa.md) |
| 3 | `milestone-3-mechanical-quality` | Harden code-quality + persistence; dead-surface deletes, tx-hazard hoist | passed | [PASS](milestone-3-mechanical-quality/qa/milestone-qa.md) |
| 4 | `milestone-4-systemic-ports` | Close H-G class via shared ports (UserDisplayNameReader, TemplateVersionStateReader) | blocked | FAIL→4b ([qa](milestone-4-systemic-ports/qa/milestone-qa.md)) |
| 4b | `milestone-4b-legacy-schema-teardown` | Drop dead `metaldocs.documents` duplicate cluster + `template_audit_log` (root cause of F4.1a Gate #5); fix Family-B tripwire seeds | partial — superseded by 4c | F4b.1+F4b.2 done (drop migration `071931c9`); F4b.3/F4b.4 → 4c |
| 4c | `milestone-4c-test-fixture-framework` | One unified test-fixture framework (factories on `testdb`) + migrate all stateful tests + CI grep-guards; closes the M4 blocker at root cause | in-progress | — |
| 5 | `milestone-5-independent-re-audit` | Prove Grade A by independent fresh multi-agent re-audit (authoritative) | planned | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open).

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| 2026-06-14 | HS-1 | M0 close gate — validator PASS presented to operator | **Approved** by operator 2026-06-14; M0 → passed, M1 opened |
| 2026-06-14 | HS-1 | M1 close gate — validator PASS (C1–C7) presented to operator | **Approved** by operator 2026-06-14 (option 2); M1 → passed. Condition: run a bounded **test-infra-rebaseline** micro-task before M2 to discharge the F1.3 AC5 full-HTTP-E2E defer. **Condition met 2026-06-14** — full HTTP `seed→finalize→signoff` E2E green, snapshot read-back `matches=t`; evidence `milestone-1-reach-a-blockers/test-infra-rebaseline/evidence.md`. M2 awaits operator go. |
| 2026-06-14 | HS-1 | M2 open gate — operator approved opening M2 ("Open M2 — spec it") | **Approved** by operator 2026-06-14; M2 → in-progress, `milestone.md` authored up front before any feature |
| (carry-forward) | HS-2 | F0.1 watch — FE eigenpal `file:` path defer | **Did not trip in M2.** The single `gen:api` ran via the present `openapi-typescript` binary only (no FE `pnpm install`), so the eigenpal `file:` path was never exercised. Defer still open for any future FE `pnpm install`. |
| 2026-06-14 | HS-1 | M2 close gate — `milestone-validator` C1–C7 **PASS** presented to operator | **Approved** by operator 2026-06-14 ("Open M3 — spec it"); M2 → passed. Verdict: `milestone-2-contract-tail/qa/milestone-qa.md` |
| 2026-06-14 | HS-1 | M3 open gate — operator approved opening M3 | **Approved** by operator 2026-06-14; M3 → in-progress, `milestone.md` authored up front (F3.1–F3.6, no execution detail) before any feature |
| 2026-06-14 | HS-6 | M3 spec stale — pre-F3.1 investigation found Wave 2.11 (`63f74368`) + 2.12 (H-6a) already did F3.1 (7 orphan deletes) and F3.2 (H-PRE-1 deadlock hoist) *before* the 06-13 audit that re-flagged them; F3.6 ("dead camelCase MarshalJSON") doesn't exist | **Replanned** by operator 2026-06-14 ("amend milestone.md, then execute"). `milestone.md` amended: F3.1/F3.2 → verify-already-done evidence rows; F3.6 struck (security-unsafe stale finding); real remainder = F3.3, F3.4, F3.5. Doc-drift (stale wiki/GitNexus refs to deleted symbols) handed to wiki-curator. |
| 2026-06-14 | HS-6 | F3.4 spec one-liner wrong — governing-spec `COUNT(*) OVER()` prescription assumed OFFSET, but the code is keyset/cursor pagination (a window on the cursor-filtered query counts only the post-cursor tail) | **Replanned** — operator delegated the engineering call ("what do you recommend as a Google senior engineer"); chose **Approach B (CTE single-query)**: count over the base-filtered set in a CTE *before* the cursor predicate. Recorded in `f3.4-…/spec.md` HS-6 reconciliation. |
| 2026-06-15 | HS-6 | F3.5 site set stale — milestone named 3 `DeleteObject` sites (`:537/:740/:331`); current tree has **1** (file is 710 lines; `:331` not a DeleteObject) | **Resolved under the row's own "or documented" clause** (not a full stop) — single real site `service.go:534` fixed; two stale-named sites documented non-existent. `f3.5-…/spec.md` + `evidence.md`. |
| 2026-06-15 | HS-1 | M3 close gate — `milestone-validator` C1–C7 **PASS** presented to operator | **Approved** by operator 2026-06-15 (`/milestone M4` = open M4). M3 → passed; verdict `milestone-3-mechanical-quality/qa/milestone-qa.md`. |
| 2026-06-15 | HS-1 | M4 open gate — operator approved opening M4 | **Approved** by operator 2026-06-15; M4 → in-progress, `milestone.md` authored up front (F4.1–F4.3, no execution detail) with an HS-6 scope-reconciliation note. Scope decision (later **SUPERSEDED**, see below): operator chose defer security on the premise its `iam_users` reads were tenant-scope JOINs. F4.1 = the 3 display-name reads. |
| 2026-06-15 | HS-4 | M4 close gate — `milestone-validator` returned **FAIL**: census omitted `auth/sessions_admin.go:32` cross-module `iam_users.display_name` read; opened fix feature `f4.4-auth-session-display-name-port` | **In progress.** Verifying the finding surfaced a deeper defect (HS-6 below). |
| 2026-06-15 | HS-6 | M4 census undercount — verifying the validator's 1 site found **3 more**: `security` `ListLockouts:89`/`CountRecentFailedLoginsByUser:137`/`ListNewDeviceLogins:191` DO read `iam_users.display_name` (original census mislabeled them "no display-name read"). Authoritative count = **4** display-name reaches outside `iam/` (1 auth + 3 security); the earlier "defer security" decision rested on a false premise. | **Replanned — operator chose Option 2 (FULL CLOSE) 2026-06-15.** Close all 4 in M4 incl. building the previously-deferred **iam tenant-scope/membership port**. `milestone.md` census **corrected**; features **F4.4 (auth) + F4.5 (iam membership port) + F4.6 (security)** added. Genuine remaining defer narrowed to security's non-display-name aggregate JOINs (MfaCoverage/CountRecentLockouts) + security's `auth_identities` reads. |
| 2026-06-15 | HS-4 | M4 close gate (post F4.4–F4.6) — `milestone-validator` returned **FAIL** on one isolated gate: **F4.1a Gate #5** (`TestCreateDocumentTx_PopulatesAllSnapshotColumns`) is environment-coupled (passes only when the operator DSN omits `search_path`). H-G class-zero bar itself PASSED. Opened named fix feature `f4.1b-testdb-search-path-robustness`. | **Superseded by the HS-6 below.** `/systematic-debugging` of F4.1b found the failure is a *symptom*: bare unqualified `documents` in 40+ runtime SQL sites resolves by `search_path`, and a dead legacy `metaldocs.documents` duplicate shadows the real `public.documents` under any metaldocs-first `search_path`. F4.1b's harness-adapt was rejected as symptom-patching. |
| 2026-06-15 | HS-6 | Root cause is a schema defect, not a harness bug — **two** duplicate tables exist in both schemas (`documents`, `template_audit_log`); `metaldocs.documents` anchors a 7-table dead FK satellite cluster (`document_attachments`, `document_collaboration_presence`, `document_edit_locks`, `document_template_assignments`, `document_versions`, `document_versions_mddm`, `workflow_approvals`). All **zero runtime Go refs** (dead, not live). Fixing it is a `metaldocs-database` migration outside F4.1b's named gate. | **Replanned — operator chose "new milestone: drop legacy cluster" + "fix Family-B test seeds now" 2026-06-15.** Opened **milestone 4b** (`milestone-4b-legacy-schema-teardown`) to verify-dead + `DROP` the cluster + both duplicates and mirror into the curated baseline; fix Family-B tripwire seeds (set `metaldocs.asserted_caps`). F4.1b → **superseded** (no harness change; `db.go` stays at HEAD — proof of fix-not-adapt). On 4b PASS + HS-1, **re-dispatch the M4 validator** (F4.1a Gate #5 re-greens with the operator DSN unchanged). Stash `stash@{0}` (rejected harness edits) retained until 4b green, then dropped. |
| 2026-06-15 | HS-1 | M4b open gate — operator authorized opening 4b (scope answer: "new milestone: drop legacy cluster") | **Approved** by operator 2026-06-15; 4b → in-progress, `milestone.md` authored up front (F4b.1–F4b.4, no execution detail) before any feature. |
| 2026-06-15 | HS-6 | M4b F4b.3 seed-fix is a **test-architecture defect, not a seed bug** — two integration harnesses coexist with no governance rule (`testdb` template-DB-per-test vs `pgtest` shared-DB); ~35 hardcoded-tenant + ~28 inline `set_config` (one `is_local=false` leaks caps) + ~11 copy-paste-taxonomy + 5 local seed helpers drifted across files; the per-tenant-code seed patch collides (`document_profiles_pkey` 23505) on the shared DB. | **Replanned — operator chose "Full framework now" 2026-06-15.** Opened **milestone 4c** (`milestone-4c-test-fixture-framework`): build factories on `testdb` + migrate ALL stateful test files + CI grep-guards, as its own milestone, before closing the program. **Supersedes M4b's F4b.3/F4b.4** (F4b.1 census + F4b.2 drop migration `071931c9` stand). Abandoned per-tenant-code WIP in the tree (`postgres_approval_repository_test.go`, `scheduled_publish_job_test.go`, +minor `repository_revision_history_integration_test.go`/`authz_bypass_test.go`) is **replaced** by the factory migration (F4c.2), not built upon. |
| 2026-06-15 | HS-1 | M4c open gate — `milestone.md` authored up front (F4c.1–F4c.5, no execution detail); presented to operator for Phase-2 agreement | **Approved** by operator 2026-06-15 ("Approve — start F4c.1"); M4c → in-progress. Executing F4c.1 only, stop at its evidence before F4c.2. |

## Program close-out / reconciliation

Fill in only when the last milestone has passed:

- [ ] Every planned feature (M0–M4, M4b) has a complete evidence row.
- [ ] Zero unplanned scope merged; anything added is recorded with rationale.
- [ ] Every bounded defer has a written trigger and an owner.
- [ ] M5 re-audit passed the §6 pass bar — link the evidence.
- [ ] Forward roadmap (F0.3) reflects the executed program and any deferred triggers.
- [ ] Operator sign-off: <date / name>
