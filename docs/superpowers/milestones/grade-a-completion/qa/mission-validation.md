# Mission Terminal Acceptance — Verdict (2026-06-19 re-run)

> Written by: the mission-validator subagent (separation of powers). Validates against: ../mission.md §8.
> Run: 2026-06-19 · HEAD: ad8e6fc8 · Verdict: see bottom.
> Artifact judged: `wiki/backend/_artifacts/architecture-re-audit-2026-06-19.md` (10-dim fan-out, skeptic-per-Critical/Major, class grep counters, synthesis).
>
> Supersedes the 2026-06-16 verdict at HEAD `9a2a2f8d` (also FAIL) — kept in git history; this is the post-M5 terminal re-run.

## §8 Pass-bar (quoted)
> (1) module-boundaries, contract-api, and composition all ≥ A−; (2) 0 skeptic-confirmed new
> Critical/Major; (3) H-D = 0; (4) H-G = 0. All four must hold simultaneously.

I did not trust the artifact's transcripts. I independently re-ran the §6 grep commands, spot-checked
two cited Major sites with `Read`, verified the §5 refuted Major against its file header + ADR cite,
and re-ran the whole-repo test suite.

## Per-criterion results

| # | §8 criterion | Method run (command/agent) | Real evidence | Pass? |
|---|--------------|----------------------------|---------------|-------|
| 1 | 3 formerly-C dims (module-boundaries, contract-api, composition) ≥ A− | Read §2 scorecard + §3 check 1 verdict in re-audit-2026-06-19.md | module-boundaries **A−** (was B+, +1 — F5.x ports landed), composition **A−** (held), **contract/API B−** (regressed from B+; broader audit found 24 surviving `writeJSON(...map[string]any{...})` sites on public delivery routes + tri-source drift on templates lifecycle 200 schema) | FAIL |
| 2 | 0 skeptic-confirmed Critical/Major | Read §4 confirmed-findings table; independently opened 2 cited sites | **5** skeptic-confirmed Majors persist, all in Contract/API: (a) `templates/.../routes_lifecycle.go:46,100,164,196,239` — spot-checked line 46, emits `writeJSON(w, http.StatusOK, map[string]any{"data": map[string]any{"version": dto}})`; (b) `templates/.../routes_query.go:73,104,145,211,260`; (c) `iam/.../admin_handler.go:341,378` — spot-checked line 341, emits `writeJSON(w, http.StatusOK, map[string]any{"user_id":..., "role":..., "display_name":...})`; (d) `iam/.../sessions_handler.go:132,138,158`; (e) `iam/.../observability_handler.go:81,109`. | FAIL |
| 3 | H-D = 0 | Validator independently re-ran: `grep -rEn 'writeJSON.*map\[string\]any' internal/modules/*/delivery/http/ --include='*.go' \| wc -l` | **24** (matches §6 Grep A line-for-line: 2 sites in iam/admin_handler, 2 in iam/routes_memberships, 1 in iam/sessions_handler, 5 in security/handler, 2 in taxonomy/routes_{areas,families}, 1 in templates/routes_catalog, 5 in templates/routes_lifecycle, 5 in templates/routes_query, 1 in templates/routes_schema) | FAIL |
| 4 | H-G = 0 | Validator independently re-ran §6 Grep A/B/C | iam_users cross-module reads: **0** (no output). iam_user_roles cross-module reads: **0** (no output). `"published"` literal: 7 hits, all in `documents/approval/application/{obsolete,publish,supersede}_service.go` — independently inspected; every hit is a Go doc-comment describing intrinsic workflow transitions, none is a SQL literal or a status comparison. H-G class **= 0**. F5.1 fix at `templates/infrastructure/template_version_reader.go:45` (typed `templatesdomain.VersionStatusPublished`) and F5.2 fix at `auth/infrastructure/postgres/repository.go:117` (`UserTenantReader` IAM port) both verified by the disappearance from the greps. | PASS |
| Aux-A | All M0–M5 milestones closed (validator PASS + HS-1) | Read README.md status table | M0 (HS-1 2026-06-15), M1 (HS-1 2026-06-15), M2 (HS-1 2026-06-16), M3 (HS-1 2026-06-16), M4 (HS-1 2026-06-16), M5 (HS-1 2026-06-19) — all `passed`. Mission procedurally ready for terminal validation. | PASS |
| Aux-B | `go test -count=1 ./...` green from clean state | `go test -count=1 ./...` (validator-run) | All packages `ok`; exit code **0**. No FAIL/panic/build-error lines. (Green tests do not rescue the verdict — §8 fails on checks 1–3 regardless, per spec.) | PASS |
| Aux-C | §5 refuted Major properly grounded | Read `iam/.../routes_memberships.go:1-15` header + cross-referenced ADR 0012 cite | Header lines 1-6 explicitly state: *"Hand-rolled rather than codegen-served — IAM is still pre-codegen on the BE side per ADR 0012 partial rollout"*. Refutation is documented-model, not hand-wave; affected sites (lines 168, 235) correctly demoted to §7 Minor #30 with a written ADR-0012 trigger. | PASS |
| Aux-D | All 5 confirmed Majors are in-scope (not on §2 Non-Goals / §5 refuted list) | Cross-check §4 sites vs mission.md §2 + the 8 prior refuted findings (WriteTimeout, cross-tenant revoke, etc.) | All 5 are typed-response defects squarely in the M1 contract-api / H-D class. None on Non-Goals; none on prior refuted list. In scope → the bounded HS-5 micro-milestone can close them. | PASS |

## Pass bar

- Met? **NO.** Only **1 of 4** binding checks pass (Check 4: H-G = 0). The contract/API dimension regressed from B+ to B− under the broader fresh audit; H-D class count = 24 (M5 closed the 2 prior-cited spot sites but did not sweep the class); 5 confirmed Majors persist — all in the contract/API root-cause family.
- Deciding evidence: the validator-run §6 Grep A returned 24 lines exactly as cited in the artifact; spot-checks at `routes_lifecycle.go:46` and `admin_handler.go:341` confirmed real `writeJSON(..., map[string]any{...})` emits on public routes; H-G greps returned no output (PASS); test suite green (does not rescue).

## Forbidden-list (any hit = FAIL)

- [ ] Fixture/mock passed off as real-provider proof — N/A (greps are real-code state at HEAD; tests are the real suite)
- [ ] A criterion marked pass without a command actually run — none; every PASS/FAIL row carries an independently re-run command + real output
- [ ] Split-brain / guessed contract surfaced in the aggregate diff — yes, the audit surfaced templates-lifecycle handler↔OpenAPI tri-source drift; recorded honestly as a Major, not glossed
- [ ] Self-judged / validator edited or fixed code — no; validator only re-grepped, Read'd, ran tests, and wrote this verdict file

## Verdict

- **VERDICT: FAIL**
- **Failed criteria:** §8 checks **#1** (contract/API B− < A−), **#2** (5 confirmed Majors, target 0), **#3** (H-D = 24, target 0). Check **#4** (H-G = 0) PASS. M0–M5 closure and the test suite are green.

- **Bounded HS-5 remediation micro-milestone needed (Grade-A Completion M6):**
  1. **Convert the 5 cited Major hot-sites to generated typed `*JSONResponse` structs:**
     - `internal/modules/templates/delivery/http/routes_lifecycle.go:46,100,164,196,239` — submit / review / archive / upsertApprovalConfig / approveTemplateVersion. Each must serialize its declared OpenAPI 200 type.
     - `internal/modules/templates/delivery/http/routes_query.go:73,104,145,211,260` — list / get / getVersion typed responses.
     - `internal/modules/iam/delivery/http/admin_handler.go:341,378` — UpsertUserAndAssignRole / handleReplaceUserRoles typed envelopes.
     - `internal/modules/iam/delivery/http/sessions_handler.go:132,138,158` — sessions list typed.
     - `internal/modules/iam/delivery/http/observability_handler.go:81,109` — KPI / usage typed.
  2. **Complete OpenAPI alignment for the templates lifecycle endpoints** (declare the 200 body schemas; align `approveTemplateVersion` to its declared `ApproveTemplateVersionResponse`). Re-run `oapi-codegen`; regen FE codegen in contract-first order; update wiki Last-verified stamps.
  3. **Class sweep — drive H-D Grep A to 0:** also close the remaining writeJSON-map[string]any sites flagged in §6 Grep A: `security/delivery/http/handler.go:67,94,107,130,173`; `taxonomy/.../routes_{areas,families}.go`; `templates/.../routes_{catalog,schema}.go`. The 2 sites in `iam/.../routes_memberships.go:168,235` may either be closed in-scope or carried as an operator-approved bounded defer with a written ADR-0012 trigger (per §5 refute + §7 Minor #30).
  4. **Exit bar:** §6 Grep A returns **0** hits (or only ADR-0012-deferred lines with operator-approved triggers); contract/API graded **≥ A−** on re-audit; **0** skeptic-confirmed Majors. Whole-repo `go test ./...` green; no prior-milestone regression.
  5. **Then:** main session re-runs the F5.1 10-dimension fan-out (`Workflow` — 10 sonnet auditors + skeptic-per-Critical/Major + class counters + synthesis) and re-dispatches the mission-validator. Operator decides continue vs replan at the next HS-1.

- **The mission stays open.** The main session does **not** flip program status, does **not** declare Grade A, and does **not** execute §12 close-out. M5's own milestone-validator PASS + HS-1 stand; the FAIL here is the wider, fresh independent audit catching that the H-D class was broader than the 2026-06-16 spot-count of 2 captured — exactly the gap the terminal gate exists to surface, not a regression of M5.

- Per §8: any single dimension missing twice = HS-2 design-boundary signal. Contract/API has now missed twice (B+ on 2026-06-16, B− on 2026-06-19). The operator should consider whether the next loop is a bounded sweep or an HS-2 redesign boundary (e.g. codegen-first StrictServerInterface adoption across the remaining modules).
