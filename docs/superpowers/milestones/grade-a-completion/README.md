# Program: Grade-A Completion (close the F5.1 re-audit gap)

> **Governing spec:** `./mission.md`
> **Evidence base:** `./discovery-brief.md`
> **Status:** Planning (mission specced + scaffolded; M0 not yet started)
> **Owner / operator:** leandrotca.work
> **Parent program:** `../grade-a-architecture-remediation/` — this mission **is** that program's F5.2
> remediation under M5/HS-5; its terminal acceptance closes the parent's M5 and is the basis for Grade-A sign-off.

Close the gap surfaced by the program's F5.1 independent re-audit
(`wiki/backend/_artifacts/architecture-re-audit-2026-06-15.md`, **VERDICT: MICRO-WAVE NEEDED**): 3 formerly-C
dimensions below A− (module-boundaries B+, contract-api C+, composition B+), 21 skeptic-confirmed
Critical/Major findings, H-D=4, H-G=1 (+1 new boundary site). Five milestones close the findings by
root-cause family in the locked order **authz → contract → observability → quality → ports-last**. **Terminal
acceptance:** the main session re-runs the F5.1 10-dimension re-audit and an independent `mission-validator`
judges it against §6 — 3 dims ≥ A−, 0 new Critical/Major, H-D=0, H-G=0.

## Milestones

| # | Milestone | Objective (one line) | Status | Gate result |
|---|-----------|----------------------|--------|-------------|
| 0 | `milestone-0-authz-session` | Close the 4 auth/authz/session defects at their shared root (incl. B1 effective_from security bug) | passed (HS-1 approved 2026-06-15) | [PASS](milestone-0-authz-session/qa/milestone-qa.md) |
| 1 | `milestone-1-contract-integrity` | Drive contract-api ≥ A− and **H-D → 0** — typed responses, spec-conformant status/body | passed (HS-1 approved 2026-06-15) | [PASS](milestone-1-contract-integrity/qa/milestone-qa.md) |
| 2 | `milestone-2-observability` | Drive composition ≥ A− — slog, scrapeable job metrics, app-level OTel spans | passed (HS-1 approved 2026-06-16) | [PASS](milestone-2-observability/qa/milestone-qa.md) |
| 3 | `milestone-3-quality-tail` | Lift code-quality/legacy — wire IAMUserOptions (functional), type-safety, dead-code, minor sweep | passed (HS-1 approved 2026-06-16) | [PASS](milestone-3-quality-tail/qa/milestone-qa.md) |
| 4 | `milestone-4-module-ports` | *(LAST)* Drive module-boundaries ≥ A− and **H-G → 0** — published constant, IAM role port, MfaCoverage port | passed (HS-1 approved 2026-06-16) | [PASS](milestone-4-module-ports/qa/milestone-qa.md) |
| 5 | `milestone-5-hs5-remediation` | HS-5 micro-milestone — close 4 re-audit gaps: H-G→0, H-D→0, 4 Majors, contract-api+module-boundaries ≥ A− | passed (HS-1 approved 2026-06-19) | [PASS](milestone-5-hs5-remediation/qa/milestone-qa.md) |
| 6 | `milestone-6-hs5-contract-sweep` | HS-5 micro-milestone — drive contract-api ≥ A− and H-D Grep A → 0: typed `*JSONResponse` on 5 confirmed Major hot-sites + class sweep + OpenAPI 200 alignment on templates lifecycle | passed (HS-1 approved 2026-06-19) | [PASS](milestone-6-hs5-contract-sweep/qa/milestone-qa.md) |
| 7 | `milestone-7-hs2-contract-completion` | HS-2 decision (operator-scoped typed-body parity, not full StrictServerInterface rewire) — close the surviving contract gap: audit/auth/search/documents typed responses + OpenAPI 200 declarations + honest H-D gate redefinition | **validator PASS 2026-06-20** (F7.1–F7.5 closed; HS-1 operator gate pending) | [PASS](milestone-7-hs2-contract-completion/qa/milestone-qa.md) |
| 8 | `milestone-8-grade-a-contract-completion` | HS-5 (4th-miss) closure — typed-everywhere + gate-scope honesty: presence/metrics typed bodies, search→taxonomy port, deactivation session enforcement, problem+json 405, widen §5b/§8 gate to the whole public surface | **PASS** (milestone-validator VERDICT PASS 2026-06-20; F8.1–F8.6 closed; HS-1 gate + post-M8 re-audit pending) | `qa/milestone-qa.md` |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open).
The **Gate result** column links each milestone-validator verdict (`milestone-<n>/qa/milestone-qa.md`);
`passed` requires a validator **PASS** *and* operator HS-1 approval. Each milestone's `milestone.md` is
authored later by the `milestone` skill (its Phase 2), in a fresh session — not by `/mission`.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| 2026-06-16 | HS-5 | Terminal re-audit FAIL — 4 §8 checks unmet (H-G=2, H-D=2, 4 Majors, 2 dims below A−) | M5 bounded micro-milestone opened; operator approved 2026-06-16 |
| 2026-06-19 | HS-5 | Terminal re-audit FAIL — 3 §8 checks unmet (contract-api B−, 5 confirmed Majors, H-D=24); H-G=0 PASS. Contract/API missed twice (HS-2 signal noted). | M6 `milestone-6-hs5-contract-sweep` opened; operator chose bounded sweep over HS-2 redesign (2026-06-19) |
| 2026-06-20 | HS-2 | Post-M6 terminal re-audit FAIL — 3 §8 checks unmet (contract-api **B**, 1 confirmed Major, H-D=10 via `writeFillInJSON`/multiline-map sites Grep A is blind to); H-G=0 PASS. **Contract/API missed a third consecutive time (B+ → B− → B)** — the HS-2 redesign-boundary signal. M6's own HS-5 rule said "do not open a bounded M7 by default." | Operator weighed the HS-2 redesign boundary against discovery evidence: (a) the A− bar was reached **twice without** StrictServerInterface (templates std-server wrappers + typed bodies; IAM hand-rolled typed structs per ADR 0012), and (b) auth + search have **no codegen pipeline at all**, making a full codegen-first rewire disproportionate to what §8 requires (contract ≥ A−, 0 Majors, H-D=0 — not a specific framework). **Operator decision 2026-06-20: open M7 as a bounded typed-body-parity sweep**, not the full rewire. M7 `milestone-7-hs2-contract-completion` opened. |
| 2026-06-20 | HS-5 | **Post-M7 terminal re-audit FAIL — all 4 §8 checks unmet** (`architecture-re-audit-2026-06-20-post-m7.md`, `qa/mission-validation.md`, mission-validator corroborated): contract-api **B+**, composition **B+** (both below A−), **5 confirmed Majors**, **honest H-D=2** (presence + metrics response literals the §5b path-scoped gate misses), **honest H-G=1** (search→taxonomy `document_profiles`). Build+tests green. **Contract/API's 4th consecutive miss (B+ → B− → B → B+).** Root cause: §5b/§8 gates scoped narrower than intent — sites outside `internal/modules/*/delivery/http/` + the two IAM tables survived every bounded sweep. | Operator decision 2026-06-20: option A — **typed-everywhere + gate-scope honesty**. Opened **M8** `milestone-8-grade-a-contract-completion` (F8.1 presence typed body, F8.2 metrics typed envelope, F8.3 search→taxonomy port, F8.4 deactivation session enforcement, F8.5 problem+json 405, F8.6 widen §5b/§8 gate + CI guard). Spec + feature tree scaffolded; execution in a fresh session. |

## Program close-out / reconciliation

M0–M6 passed validator + HS-1. Terminal acceptance has run twice — both FAIL on the Contract/API
dimension. M6 closed all 5 prior contract Majors and zeroed Grep A, but the **post-M6 re-audit
(2026-06-20, HEAD `5650b328`) FAILed** the §8 bar: contract-api **B** (not A−), 1 confirmed Major
(audit export status), and **H-D = 10** sites that Grep A's one-liner pattern structurally cannot
see (the `writeFillInJSON` alias + multi-line map construction). This was Contract/API's **third
consecutive miss** (B+ → B− → B) — the HS-2 redesign-boundary signal. Operator opened **M7**
(bounded typed-body parity, not full StrictServerInterface rewire — see hard-stops 2026-06-20).
Mission stays open.

- [x] Every planned feature (M0..M5) has a complete evidence row. (M5: F5.1–F5.8 each have evidence.md.)
- [x] Zero unplanned scope — the only added item (F5.8 + the 2 bundled otel tests) is recorded with
      rationale and operator-ratified; semconv vendor payload reclassified as prerequisite repair.
- [x] Every bounded defer has a written trigger and an owner. (M5 features close with no open defers;
      Model-B time-bounded memberships gated behind a successor ADR per ADR 0037 D4.)
- [x] Terminal acceptance ran (2026-06-19, fresh session). Artifacts:
      `wiki/backend/_artifacts/architecture-re-audit-2026-06-19.md`,
      `qa/mission-validation.md`. Verdict: **FAIL** (3 of 4 §8 checks unmet — contract/API B−,
      5 confirmed Majors, H-D=24; H-G=0 PASS). Commit `cecb559d` (local).
- [x] **HS-5 M6 — `milestone-6-hs5-contract-sweep`** (operator-approved 2026-06-19): closed all 5
      prior contract Majors + zeroed Grep A. Validator PASS + HS-1 approved. Post-M6 terminal re-audit
      ran 2026-06-20 (HEAD `5650b328`) — **FAIL** (`architecture-re-audit-2026-06-19-post-m6.md`,
      `qa/mission-validation.md`): contract-api B, 1 Major, H-D=10. Third consecutive Contract/API miss.
- [ ] **HS-2 M7 — `milestone-7-hs2-contract-completion`** (operator-scoped 2026-06-20): drive
      contract-API ≥ A−, close the 1 confirmed Major + all 10 surviving H-D sites across audit / auth /
      search / documents, and **redefine the H-D acceptance gate** so it is no longer blind to the
      `writeFillInJSON` alias + multi-line map construction (the flaw that let M6 report Grep A = 0 while
      10 sites survived). Operator chose **typed-body parity** (audit/documents via existing codegen;
      auth/search via hand-rolled typed structs per ADR 0012) over a full codegen-first
      StrictServerInterface rewire — the lighter pattern that already produced A− twice and satisfies §8.
      On M7 close, re-run F5.1 fan-out and re-dispatch `mission-validator`.
- [x] **Post-M7 terminal acceptance ran (2026-06-20, HEAD `dadb8275`) — FAIL** (all 4 §8 checks):
      `wiki/backend/_artifacts/architecture-re-audit-2026-06-20-post-m7.md` + `qa/mission-validation.md`
      (mission-validator corroborated). Contract-api B+, composition B+, 5 confirmed Majors, honest
      H-D=2, honest H-G=1. M7's in-scope work was sound (audit-export Major closed, 4 docs schemas
      aligned, in-path Part B allowlisted) but 2 response literals + 1 cross-schema reach sit outside
      the gate scope. Commit `9ed83235` (local). **Contract/API 4th consecutive miss → HS-5.**
- [ ] **HS-5 M8 — `milestone-8-grade-a-contract-completion`** (operator option A, 2026-06-20):
      typed-everywhere + gate-scope honesty. F8.1–F8.6 close the 5 Majors and widen §5b/§8 to the
      whole public surface. Spec + feature tree scaffolded this session; execution in a fresh session.
      On M8 close, re-run the post-M8 re-audit + `mission-validator`. **5th miss ⇒ no M9 by default
      (HS-5): full codegen-first rewire or §8 re-scope is the operator's call.**
- [ ] Terminal acceptance passed — link the new `qa/mission-validation.md` + the new re-audit report.
- [ ] Parent program M5 closed; `../grade-a-architecture-remediation/README.md` updated.
- [ ] Operator Grade-A sign-off: <date / name>
