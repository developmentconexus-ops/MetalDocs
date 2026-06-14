# Milestone 0 — Validation Verdict (C1–C7)

> **Written by:** the `milestone-validator` subagent — *not* the main session (separation of powers).
> **Validates against:** `../milestone.md` (the up-front spec) + each feature's `plan.md` / `evidence.md`.
> **Binding procedure:** `.claude/skills/milestone/references/milestone-end-validation.md`.
> **Run:** 2026-06-14  ·  **Verdict:** see C7 — **PASS**.
> The validator judges and writes this file; the **main session flips status only on a PASS**. The
> validator never edits code, fixes findings, or flips status.

## Inputs loaded

| Input | Loaded | Note |
|-------|--------|------|
| Milestone spec | yes | `../milestone.md` |
| Program README | yes | `../../README.md` (status table; M0 = in-progress; M0 is the FIRST milestone) |
| Governing spec (M0 §6) | yes | `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md` |
| F0.1–F0.5 `plan.md` + `evidence.md` | yes | F0.1 also has `drift-ledger.md` |
| Aggregate diff | yes | `git show 5cb254dd` (23 files, F0.2–F0.5) + F0.1 tracked files (`wiki/decisions/index.md`, `0001`, `0027`, ledger). `a49f833c` (skill-upgrade) correctly excluded — not M0 content. |

No input missing or unreadable. Working tree clean; HEAD = `5cb254dd`.

**Milestone class:** docs-governance (docs-only). Per `milestone.md` §2.2 there is **no canonical
*code* checklist** and **no producer/consumer *code* contract** — so C1's "consumer contract honored"
column is **N/A for this milestone** (recorded as such, with justification, below). Features were
closed under the older `plan.md`+`evidence.md` lifecycle that predates the `spec.md`
consumer-contract-first template; judged on declared acceptance per `milestone.md` rows, not on a
missing `spec.md` (anachronism avoided).

---

## C1 — Spec & plan conformance (per feature)

Consumer-contract column = **N/A (docs-only milestone — no code contract; justified above).** Each
feature judged against its `milestone.md` acceptance row + `evidence.md`.

| Feature | Consumer contract honored? | Acceptance met? | Non-goals respected? | Evidence |
|---------|----------------------------|-----------------|----------------------|----------|
| F0.1 ADR audit + ledger | N/A (docs) | ✅ | ✅ (no decision body rewritten; 0 DECISION-DRIFT → no HS-6) | All 25 ADRs carry a vocabulary-valid `> **Status:**`; `drift-ledger.md` = 23 MATCH + 2 STATUS-DRIFT (0001, 0027) corrected + 0 DECISION-DRIFT; `index.md` reconciled |
| F0.2 Stale-ref repair | N/A (docs) | ✅ | ✅ (`.agents/skills/` class left untouched per operator ruling; no decision text altered) | Gate A grep = **0** md-links to deleted `docs/` non-superpowers paths; every residual `docs/` token in M0-touched files is a past-tense "removed at re-baseline" note or a surviving `docs/superpowers/` path |
| F0.3 Roadmap consolidation | N/A (docs) | ✅ | ✅ (carried-forward items referenced, not re-adjudicated — HS-6 held) | Exactly **1** forward roadmap (`wiki/roadmap.md`); `backend/roadmap.md` + `backlog/roadmap.md` both bannered HISTORICAL → `../roadmap.md` (resolves) |
| F0.4 Backlog hygiene | N/A (docs) | ✅ | ✅ (no active-defer file archived; no content rewritten) | Census applied operator "fully-closed-only" rule; one closed file (`api-contract-hardening.md`) marked CLOSED; `backlog/index.md` splits active vs closed/superseded |
| F0.5 Archive convention | N/A (docs) | ✅ | ✅ (roadmaps retained-in-place = documented scope decision, recorded in governance map) | `wiki/_archive/` + README exist; 2 docs relocated as **git renames** (R098, R094 — history preserved); all 7 outbound + 6 inbound links resolve; governance map carries the 4 rows |

All five acceptance rows met. **C1 PASS.**

## C2 — Gates re-run, isolated

Re-run by the validator from clean state (HEAD `5cb254dd`, clean tree) — not trusted from the evidence
transcripts. ERE/PCRE gotcha (`\>`/`\|`) avoided; zero/all results verified two ways.

| Check (feature) | Command re-run | Real output | Pass? |
|-----------------|----------------|-------------|-------|
| F0.1 status presence | `for f in wiki/decisions/00*.md; do grep -qE '^> \*\*Status:\*\*' "$f" \|\| echo NO:$f; done` | empty (25/25 carry status); printed each — all vocabulary-valid | ✅ |
| F0.1 ADR count | `ls -1 wiki/decisions/00*.md \| wc -l` | `25` | ✅ |
| F0.1 0027 reconcile | grep `27 remaining` / `29 total` in ADR + `index.md:31` | ADR = "27 remaining (29 total incl. 2 from 0234)"; index = "all 29" — consistent, no contradiction | ✅ |
| F0.2 Gate A | `grep -rnoP '\]\(\.{0,2}/?(?:\.\./)*docs/(?!superpowers)[^)]+\)' wiki` | no match (exit 1) — **0** stale `docs/` md-links | ✅ |
| F0.2 `docs/adr` in decisions | `grep -rn 'docs/adr' wiki/decisions/` | no match (exit 1) | ✅ |
| F0.2 Gate B (M0 files) | residual-`docs/`-token sweep + manual read of `quality/*`, `documentation-governance.md` | every M0-owned residual token is explicit `(removed)`/past-tense or a surviving `docs/superpowers/` path — none presents a deleted path as live | ✅ |
| F0.3 single roadmap | `grep -rln '^# .*[Rr]oadmap' wiki` + `grep -rln HISTORICAL …` | 3 roadmap files − 2 bannered HISTORICAL = **1** forward (`wiki/roadmap.md`); all roadmap links resolve | ✅ |
| F0.5 archive + renames | `ls wiki/_archive/backlog` + `git diff --find-renames --name-status a49f833c 5cb254dd` | both docs present; `R098` / `R094` (renames, history preserved) | ✅ |
| F0.5 link depth | resolve all 7 outbound (`../../…`) + 6 inbound (`../_archive/backlog/…`) targets with `test -f` | all 13 resolve (the self-caught QA-5 depth fix holds) | ✅ |
| F0.4 no deleted-without-trace | `ls -1 wiki/backlog/*.md \| wc -l` + CLOSED-banner read | 26 remain + 2 archived = 28 preserved; archived file carries CLOSED banner; index splits active/closed | ✅ |
| Docs-only guard | `git show --name-only 5cb254dd \| grep -E '\.(go\|ts\|sql\|yaml…)'` | no match — **0** code files in the M0 commit | ✅ |

Every gate re-ran green from clean state. **C2 PASS.**

## C3 — Senior review of the aggregate milestone diff

Whole-milestone diff (`5cb254dd` + F0.1 files) reviewed as one unit.

- **Split-brain (the headline risk): resolved.** Two competing forward roadmaps collapsed to one
  (`wiki/roadmap.md`); the other two carry top-of-file HISTORICAL banners and are recorded once in the
  governance migration map (single source of truth for where docs went). The 0027 RLS table count is
  now stated consistently across ADR body + `index.md:31` (no two-numbers contradiction).
- **Dead refs:** the two `docs/audits/…` links F0.2 owned in `modules/frontend/iam.md` were correctly
  de-staled to past-tense notes (no longer links). No M0 edit re-introduced a link to a deleted path.
- **One feature breaking another:** none. F0.3 banners + F0.4 closure judgment + F0.5 physical move +
  governance-map rows compose cleanly; the F0.4→F0.5 hand-off (relocation) is fully discharged.
- **Decision-body integrity:** every edit touched only status lines / path refs / index rows / prose
  notes; no ADR *decision* section rewritten (HS-6 guard held; 0 DECISION-DRIFT in the ledger).
- **Findings:** none rising to a staff-engineer block.
- **Staff-engineer bar met?** ✅

## C4 — Workflow-class QA + regression

| Check | Outcome | Notes |
|-------|---------|-------|
| Canonical checklist (doc-QA, per `milestone.md` §2.2) | **pass** | ADR status-gate green (25/25 vocabulary-valid); 0-stale-`docs/` grep clean; broken-link sweep classified (below) |
| Broken-link sweep — M0-owned vs pre-existing | **pass** | Independent full-wiki census = **132** broken relative links: 23 `.agents/skills/` (operator-ruled OUT), 19 code-file/artifact refs (pre-existing OUT), 90 "other" — **every one of the 90 is in a file M0 did not touch** (`standards/golang/*` historical-review refs; `controlled-documents.md`). The 4 broken links in M0-touched `iam.md` are 3 pre-existing code-refs (present verbatim at `a49f833c`, before M0) + 1 `.agents/skills/` ref — **none M0-introduced.** New M0 files (`wiki/roadmap.md`, `_archive/README.md`) have **0** broken links. **M0 introduced zero new breakage and owns zero pre-existing breakage.** |
| Regression vs prior milestones | **N/A — M0 is the first milestone** | Nothing prior to regress. Confirmed M0 did not break the existing wiki index/link graph: all M0-touched/created link targets resolve; pre-existing broken-link count is unchanged in class and not increased by M0. |

**C4 PASS.**

## C5 — Quality-bar re-measure + retrospective

The bar M0 moved: **docs de-staled → one unambiguous progression surface (single source of truth).**

| Bar / class | Before | After | Root-cause-fixed evidence |
|-------------|--------|-------|---------------------------|
| Forward roadmap | 2 competing surfaces (`backend/`, `backlog/`), no top-level | **1** forward (`wiki/roadmap.md`); 2 predecessors bannered historical | Root cause (two live plans) eliminated, not masked — predecessor bodies retained as record, recorded once in governance map |
| ADR ledger vs code | `index.md:34` referenced deleted `docs/adr/`; 0001 vendor path stale; 0027 27-vs-29 ambiguous | ledger reconciled; 0001/0027 status corrected; stale `docs/adr` ref removed | drift-ledger: 23 MATCH + 2 corrected STATUS-DRIFT + 0 DECISION-DRIFT; status-gate green |
| Stale `docs/` graph | 11 wiki refs to deleted `docs/` trees | **0** md-links to deleted `docs/` paths | Gate A grep empty; residual tokens are external URLs / past-tense removal notes |
| Archive convention | none | `wiki/_archive/` + README; superseded docs moved as git renames (not destroyed); governance map = index of record | history preserved (R098/R094); all moved-doc links resolve |

Root cause fixed, not symptom-patched: ambiguity/staleness eliminated at source (one roadmap, ledger
matches code, archive-not-delete) rather than papered over.

- **Could it be built better?** Two non-blocking retrospective notes (next-milestone input, not FAILs):
  (1) A pre-existing broken-link landscape (~132 links: `standards/golang/*` review refs, code-file
  refs) remains out of M0 scope and would benefit from a future dedicated de-staling pass if those
  surfaces are revived. (2) F0.1's HS-2 defer (FE eigenpal `file:` path in `frontend/apps/web/package.json`
  resolves through a non-existent repo-root `vendor/eigenpal/`) is a real pre-existing **code** defect
  with a written trigger (fix before any FE `pnpm install`, incl. M2) — correctly held out of a docs
  milestone; the main session/operator should ensure it is scheduled before M2.

## C6 — Forbidden-list (any hit = FAIL)

- [ ] Suite-green reported as a pass without per-feature acceptance mapped to evidence — *clean (each feature's acceptance mapped to a re-run command).* 
- [ ] Fixture/mock passed off as real-provider proof — *clean (docs-only; gates are real greps/file reads I re-ran).* 
- [ ] Consumer contract guessed rather than read — *clean (N/A — no code contract; acceptance read from `milestone.md`).* 
- [ ] Split-brain (one fact, two sources of truth) — *clean (two-roadmap split-brain resolved; 0027 count single-stated; governance map single source).* 
- [ ] Self-judged close / validator edited code — *clean (main session dispatched this independent validator; validator wrote only this verdict file, edited/fixed nothing).* 
- [ ] Scope drift (work beyond spec, no rationale) — *clean (docs-only; 0 code files; F0.5 roadmap retain-in-place recorded with rationale).* 
- [ ] Symptom-patch (bar moved by masking) — *clean (root causes eliminated at source).* 

All unchecked = clean. **C6 PASS.**

## C7 — Verdict

- **VERDICT: PASS**
- Both dimensions pass: **code-wise** (contract-clean, no split-brain, no dead refs, decision bodies
  intact, docs-only with 0 code touched) and **function-wise/QA** (every feature's declared acceptance
  independently re-verified from clean state; M0 introduced zero new broken links and owns zero of the
  pre-existing breakage landscape; one forward roadmap; ledger reconciled; archive convention live with
  history preserved).
- No failed check; no fix feature required. Milestone may advance.

> **Main-session actions (post-verdict, NOT the validator's):**
> - Operator gate (HS-1): **pending** — not yet run; main session presents it on this PASS.
> - Status flipped in `README.md` (M0 `in-progress` → `passed`): **pending** — main session only, on this PASS.
> - Carry-forward watch: F0.1 HS-2 defer (FE eigenpal `file:` path) — ensure scheduled before any FE `pnpm install` / M2.
