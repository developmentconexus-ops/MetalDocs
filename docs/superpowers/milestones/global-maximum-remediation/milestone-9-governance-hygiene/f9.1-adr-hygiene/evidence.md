# F9.1 — adr-hygiene: evidence

> Feature folder: `f9.1-adr-hygiene/`. Contract: `../validation-contract.md` §1. Executed 2026-07-06.

## 1. RED sweep (pre-edit, captured before any Task-2/3 edit)

Command (documented verbatim in `wiki/standards/documentation-governance.md` "ADR status field"
section):

```bash
cd wiki/decisions && for f in [0-9]*.md; do awk -v fn="$f" '/^> \*\*Status:\*\*/{inb=1; total=0; lines=0} inb { if (!/^>/) {inb=0} else if (lines>0 && /^> \*\*[A-Za-z ]+:\*\*/) {inb=0} else {total+=length($0); lines++} } END{ if (total>400 || lines>3) print fn": "lines" lines, "total" chars" }' "$f"; done; cd ../..
```

Output (RED — 4 violators, matches the spec's pre-verified inventory exactly):

```
0015-async-freeze-pin-materialize.md: 2 lines, 410 chars
0022-authz-capability-coherence.md: 1 lines, 2757 chars
0027-rls-adoption-sequencing.md: 1 lines, 664 chars
0070-tenant-lifecycle-onboarding-export-crypto-shred-erasure.md: 2 lines, 467 chars
```

## 2. GREEN sweep (post-edit, same command, same tree)

```
(no output — 0 violations)
```

Ran again after all Task 1–4 edits landed. 0 violations confirmed across all 70 ADR files (71 with
the new companion doc excluded — it is not `0022-execution-history.md`-numbered per the sweep's
`[0-9]*.md` glob... actually it *is* numbered and matched by the glob but carries no `> **Status:**`
line, so the sweep correctly skips it as a non-ADR-status file).

## 3. Split ADRs + companion-doc / in-file-section choice

| ADR | Violation | Choice | Rationale |
|---|---|---|---|
| **0022** | 1 line, 2757 chars (13-phase mega-status) | **Companion doc** — `wiki/decisions/0022-execution-history.md` | History was long (13 phases + Wave Z + a nested Last-verified changelog) — clearly over the plan's "½ page" threshold for in-file. |
| **0027** | 1 line, 664 chars | **In-file `## Status history` section** | Status text (~3 sentences: original decision, Wave Z collapse, current-reality note) fits comfortably under half a page; two existing `## Amendment` sections already live in-file, so a matching in-file history section keeps the file's own convention. |
| **0070** | 2 lines, 467 chars — root cause was the `> **Module(s):**` field-marker containing parens, which the sweep's field-marker regex (`^> \*\*[A-Za-z ]+:\*\*`) does not recognize as a new field, so it kept absorbing into the "status block" | **Rename field** `Module(s):` → `Modules:` (no content change, no relocation needed) | The Status line itself (`Accepted 2026-07-05`) was already ≤3 lines/≤400 chars in isolation; the violation was a sweep-regex/field-naming interaction, not an actual mega-status. Renaming the field to match the regex's recognized shape is the minimal, information-lossless fix — no history existed to relocate. |
| **0015** | 2 lines, 410 chars | **In-file `## Status history` section** | The second line was a short "Current reality (2026-06)" note (4 sentences) — well under a companion-doc threshold. |

## 4. 0013 supersession research trail + final stamp

**Research performed (plan Task 4):**
- Read ADR 0013 D-2 (allocate `revision_number` at INSERT via `COALESCE(MAX+1)`) and D-5 (publish
  paths set `CurrentRevisionNumber`).
- Read ADR 0052 (manual versioning) in full — its Decision items 1–3 remove the *implicit auto-spawn*
  of a next draft on `Approve`/`PublishTemplateVersion`, making `CreateNextVersion` the sole trigger.
  0052 explicitly says (line ~21): "The machinery to create a next version already existed and is
  correct (`nextVersionNumber` + `spawnNextDraft`); the defect was *who triggers it*." — i.e. 0052
  never touches the counter allocation mechanics 0013 defined.
- Read runtime: `internal/modules/templates/repository/postgres.go:261-297` (`CreateVersion`) and
  `:304-335` (its Tx twin) both still allocate `revision_number` via
  `COALESCE((SELECT MAX(rn.revision_number)+1 …), 0)`, with an explicit code comment "ADR 0013:
  revision_number allocated atomically per template_id, mirroring documents.revision_number." All four
  `SELECT` projections (lines 92-93, 117-118, 173-174) still read `lv.revision_number`/`pv.revision_number`
  and expose them via `TemplateDTO`. `application/lifecycle.go` comments (lines 211, 232, 378, 398)
  confirm: "no longer spawns the next revision (M1·T2); use CreateNextVersion" — i.e. what changed is
  *when/who* spawns a version, not the REV-label/counter model.
- Cross-checked the review artifact cited in the READ FIRST prompt
  (`docs/superpowers/analysis/2026-07-03-final-architecture-global-maximum-review.md:105`), which claims
  0052/0053 "moved past" 0013. **This claim is not supported by the code or by 0052's own text** — 0052
  is scoped to the version-*creation trigger* only and says nothing about retiring `revision_number`,
  the REV label format, or the persisted column.

**Finding:** 0013 is **not wholly superseded**. Its persisted-column + REV-label decision (D-1 through
D-11, D-8's rendering rule) is live, unmodified runtime truth. Only the *version-creation trigger*
context 0013 implicitly assumed (that a template naturally spawns a new revision at approve/publish
time, mirroring the old feature-branch behaviour) was later removed by 0052. That is an amendment to a
narrow adjacent behavior, not a supersession of 0013's actual decision.

**Final stamp (0013):**
```
Status: Accepted (amended by 0052 — version-creation trigger; REV labels + persisted
revision_number unchanged)
Last verified: 2026-07-06 (F9.1 research confirmed the counter mechanics are still live)
```

**Reciprocal note added to 0052** (new bullet under its header block):
> **Amends:** ADR 0013 — this decision amends 0013's *version-creation trigger* (removes the implicit
> auto-spawn on approve/publish; `CreateNextVersion` is the sole trigger going forward). It does **not**
> touch 0013's `revision_number` persisted-column mechanics or `REV{nn}` labeling — those stay exactly
> as 0013 specified (the counter allocation in `internal/modules/templates/repository/postgres.go` is
> unchanged by this ADR).

**index.md rows updated:** 0013's `Status`/`Superseded by` cells now read "Accepted (amended by 0052 —
version-creation trigger only)" / "0052 (partial — trigger only)"; 0052's `Current relevance` cell
gained a leading "Amends ADR 0013 (version-creation trigger only, not the counter/labeling mechanics)"
clause.

**Disposition — why not a plain "Superseded" stamp:** the spec's interview row 4 anticipated exactly
this branch point (`Contract §0.3, §1.3; HS-6 pre-authorized path`) — a plain `Superseded by 0052` would
misrepresent runtime truth (the superseded/superseding convention implies the successor fully replaces
the predecessor's decision surface; here 0052 replaces one narrow trigger-timing behavior and leaves
0013's actual subject — the persisted revision-number model and REV labeling — completely intact and in
force). Stamping a false full-supersession would itself be a governance-truth defect of the same class
this feature exists to fix. The forbidden-list item "Marking 0013 Superseded without a researched
superseding reference" (§7) is satisfied by researching first and stamping the researched true state,
not the review artifact's unverified claim.

## 5. Governance-doc rule anchor

`wiki/standards/documentation-governance.md`, new section: **"## ADR status field (F9.1 — permanent
rule)"** — inserted immediately before the existing "## Secrets in documentation (D-4a — permanent
rule)" section. Contains: the ≤3-line/≤400-char rule + block-boundary definition, the canonical
vocabulary, the history-outside-the-field rule (in-file section vs companion doc), and the sweep
command verbatim (same one-liner used for RED/GREEN above).

## 6. `git diff --stat` (scope proof)

```
 wiki/decisions/0013-template-revision-labels.md    |  5 ++--
 wiki/decisions/0015-async-freeze-pin-materialize.md | 14 +++++++--
 wiki/decisions/0022-authz-capability-coherence.md  |  7 +++--
 wiki/decisions/0027-rls-adoption-sequencing.md     | 35 +++++++++++++++++++---
 wiki/decisions/0052-template-manual-versioning.md  |  1 +
 wiki/decisions/0070-tenant-lifecycle-onboarding-export-crypto-shred-erasure.md | 2 +-
 wiki/decisions/index.md                            |  8 ++---
 wiki/standards/documentation-governance.md         | 29 ++++++++++++++++++
 8 files changed, 86 insertions(+), 15 deletions(-)
```

Plus one new untracked file: `wiki/decisions/0022-execution-history.md` (the companion doc; new files
don't appear in `git diff --stat` for tracked content, confirmed via `git status --short`).
Untracked/pre-existing outside this feature's scope and NOT touched by this session:
`docs/release/` (pre-existing untracked dir, unrelated), and the sibling
`f9.2-traceability-gate/` feature folder (scaffolded by milestone tooling, not authored here).

No `api/openapi/`, `migrations/`, `db/`, or capability/authz code was touched. No ADR
Context/Decision/Consequences content was altered — only status-field text, one field rename
(`Module(s):`→`Modules:`), and new `## Status history` sections / the one companion doc.

## 7. Validation-gate self-check (spec.md §Validation Gate)

1. **Sweep GREEN** — §2 above, 0 violations. ✅
2. **Split integrity** — 0022 companion doc spot-checked against `git show HEAD:...` pre-split text in
   §3 workflow (Phase 3, Phase 4, Phase 5, Wave Z Z-6, Phase 6 entries all verified byte-consistent
   with the original mega-status line, restructured not summarized). ✅
3. **0013 chain** — §4 above: researched true state stamped, 0052 back-references 0013, index.md rows
   agree. ✅
4. **Rule documented** — §5 above. ✅
5. **No collateral edits** — §6 above. ✅

## Deviations from spec/plan (flagged)

- **0070 and 0015 were NOT split into a companion doc or a substantial in-file history section** the
  way 0022/0027 were, because their sweep failures were not "mega-status changelog" cases in the sense
  the spec's Task 2/3 framing anticipated:
  - 0070's failure was a **sweep-regex/field-naming artifact** (`Module(s):` parens breaking the
    field-marker regex), not embedded history — fixed by a one-word field rename, zero content moved.
  - 0015's failure was a genuinely short second status line (a "Current reality" clause), which was
    still relocated (per the rule: "execution history lives OUTSIDE the status field") into a small
    in-file `## Status history` section rather than a companion doc, per the plan's explicit "prefer
    in-file section when history <½ page" guidance.
  Both are within the plan's proportional-treatment latitude (Task 3), not a deviation from binding
  acceptance — flagged here only so the validator does not expect two more `NNNN-execution-history.md`
  files that were never warranted.
- **No other deviations.** All five plan tasks executed as specified; the 0013 stamp used the spec's
  explicitly pre-authorized non-full-supersession outcome (interview row 4 / HS-6 path), which was the
  expected outcome, not a fallback.
