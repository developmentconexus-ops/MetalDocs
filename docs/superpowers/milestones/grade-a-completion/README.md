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
| 6 | `milestone-6-hs5-contract-sweep` | HS-5 micro-milestone — drive contract-api ≥ A− and H-D Grep A → 0: typed `*JSONResponse` on 5 confirmed Major hot-sites + class sweep + OpenAPI 200 alignment on templates lifecycle | in-progress (spec approved 2026-06-19) | — |

Status vocabulary: `planned` → `in-progress` → `passed` (operator-approved) / `blocked` (hard-stop open).
The **Gate result** column links each milestone-validator verdict (`milestone-<n>/qa/milestone-qa.md`);
`passed` requires a validator **PASS** *and* operator HS-1 approval. Each milestone's `milestone.md` is
authored later by the `milestone` skill (its Phase 2), in a fresh session — not by `/mission`.

## Hard-stops raised

| When | HS id | What | Resolution |
|------|-------|------|------------|
| 2026-06-16 | HS-5 | Terminal re-audit FAIL — 4 §8 checks unmet (H-G=2, H-D=2, 4 Majors, 2 dims below A−) | M5 bounded micro-milestone opened; operator approved 2026-06-16 |
| 2026-06-19 | HS-5 | Terminal re-audit FAIL — 3 §8 checks unmet (contract-api B−, 5 confirmed Majors, H-D=24); H-G=0 PASS. Contract/API missed twice (HS-2 signal noted). | M6 `milestone-6-hs5-contract-sweep` opened; operator chose bounded sweep over HS-2 redesign (2026-06-19) |

## Program close-out / reconciliation

M0–M5 passed validator + HS-1. Terminal acceptance **ran 2026-06-19 (HEAD `ad8e6fc8`) — VERDICT: FAIL**.
Operator opened **HS-5 M6** (bounded contract-sweep) over an HS-2 redesign. Mission stays open.

- [x] Every planned feature (M0..M5) has a complete evidence row. (M5: F5.1–F5.8 each have evidence.md.)
- [x] Zero unplanned scope — the only added item (F5.8 + the 2 bundled otel tests) is recorded with
      rationale and operator-ratified; semconv vendor payload reclassified as prerequisite repair.
- [x] Every bounded defer has a written trigger and an owner. (M5 features close with no open defers;
      Model-B time-bounded memberships gated behind a successor ADR per ADR 0037 D4.)
- [x] Terminal acceptance ran (2026-06-19, fresh session). Artifacts:
      `wiki/backend/_artifacts/architecture-re-audit-2026-06-19.md`,
      `qa/mission-validation.md`. Verdict: **FAIL** (3 of 4 §8 checks unmet — contract/API B−,
      5 confirmed Majors, H-D=24; H-G=0 PASS). Commit `cecb559d` (local).
- [ ] **HS-5 M6 — `milestone-6-hs5-contract-sweep`** (operator-approved 2026-06-19): drive contract-API
      ≥ A− and H-D Grep A → 0. Scope in `qa/mission-validation.md` Verdict §. Execute in a **fresh
      session** via the `milestone` skill (Phase 2 milestone spec + features). On M6 close, re-run
      F5.1 fan-out and re-dispatch `mission-validator`. Contract/API has missed twice — if M6's
      re-audit misses a third time, treat as HS-2 (codegen-first StrictServerInterface adoption).
- [ ] Terminal acceptance passed — link the new `qa/mission-validation.md` + the new re-audit report.
- [ ] Parent program M5 closed; `../grade-a-architecture-remediation/README.md` updated.
- [ ] Operator Grade-A sign-off: <date / name>
