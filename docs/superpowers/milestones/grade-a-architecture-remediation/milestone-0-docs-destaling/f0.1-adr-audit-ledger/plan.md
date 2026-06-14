# F0.1 — ADR Audit + Ledger Refresh — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
> **Milestone:** M0 Docs De-Staling · **Feature:** F0.1 · **Spec:** `../milestone.md` · **Governing:** `docs/superpowers/specs/2026-06-14-grade-a-architecture-remediation-design.md`

**Goal:** Prove every `wiki/decisions/` ADR's status header is present, drawn from the canonical vocabulary, **and that its recorded decision still matches the code** — then refresh `decisions/index.md` so the ledger has zero drift.

**Architecture:** Docs-only feature. No source code changes. "Tests" are deterministic doc-QA gates: the status-gate PowerShell script (must emit nothing) and targeted greps/file-reads that confirm each ADR's core claim against the live tree. Drift is corrected by editing the ADR status line and/or the index row — never by rewriting a *decision* (that would be HS-6, stop-and-surface).

**Tech Stack:** Markdown, PowerShell status-gate script (`wiki/decisions/README.md`), Grep/Read over `wiki/` + backend Go + `db/migrations/`.

---

## Pre-flight facts (already established — do not re-derive)

- Status-gate script (`wiki/decisions/README.md:28-34`) currently emits **nothing** → every ADR already has a `> **Status:**` header. The status *presence* gate is green; this feature proves status *accuracy* + decision-vs-code match.
- **25** ADR files exist: 0001, 0002, 0003, 0007, 0008–0028 (gaps 0004–0006 permanent — they lived only in the now-deleted `docs/adr/` tree).
- `decisions/index.md` was last rebuilt 2026-06-13 (Z-27). It is the ledger to refresh.
- **Two drift candidates already spotted** (must be resolved in this feature):
  1. `decisions/index.md:34` — "Legacy ADR material in `docs/adr/` remains historical/reference content" — but `docs/adr/` was **deleted** in this branch (see `git status`). Stale reference.
  2. Table-count mismatch: ADR `0027` status text says RLS on "**all 27** remaining tenant-scoped tables"; `index.md:31` relevance says "RLS live on **all 29** tenant-scoped tables". One of the two numbers is wrong — reconcile against migration `db/migrations/0237_rls_all_tenant_tables.sql`.

---

## File Structure

- Modify (only if drift found): individual `wiki/decisions/00NN-*.md` status lines.
- Modify: `wiki/decisions/index.md` (ledger refresh + `Last verified:` stamp + line-34 stale ref + 27/29 reconcile).
- Create (working artifact, kept): `docs/superpowers/milestones/grade-a-architecture-remediation/milestone-0-docs-destaling/f0.1-adr-audit-ledger/drift-ledger.md` — the per-ADR verification matrix that is the evidence for this feature.

---

### Task 1: Build the per-ADR drift ledger (verification matrix)

**Files:**
- Create: `.../f0.1-adr-audit-ledger/drift-ledger.md`

- [ ] **Step 1: Create the ledger skeleton**

Create `drift-ledger.md` with this exact table header and one row per ADR (27 rows):

```markdown
# F0.1 — ADR decision-vs-code drift ledger

> One row per `wiki/decisions/` ADR. `Match?` = does the recorded decision still hold in code?
> Verdict vocabulary: `MATCH` (decision holds, status accurate) · `STATUS-DRIFT` (decision holds but status line wrong/stale) · `LEDGER-DRIFT` (index.md row wrong) · `DECISION-DRIFT` (the decision itself no longer holds → STOP, HS-6).

| ADR | Core claim (what the decision asserts) | Check run (cmd / file:line) | Observed | Verdict | Action |
|-----|----------------------------------------|-----------------------------|----------|---------|--------|
| 0001 | | | | | |
```

- [ ] **Step 2: Fill the ledger — verify each ADR's core claim against code**

For each ADR, read its decision, then run the matching check. Use the `Current relevance` column of `index.md` as the claim hint, but verify against the **live tree**, not the index. Concrete checks per ADR:

| ADR | Core claim | Verify by |
|-----|-----------|-----------|
| 0001 | eigenpal is the editor; `@eigenpal/docx-js-editor` from `vendor/eigenpal/` | Glob `vendor/eigenpal/**`; Grep `@eigenpal/docx-js-editor` in `frontend/` package.json |
| 0002 | editable zones purged; placeholders sole variability | Grep `editable[_-]?zone` (case-insensitive) in backend — expect 0 live uses |
| 0003 | Historical stub, superseded by 0008 | Confirm status already `Historical (... superseded by ADR 0008 ...)`; 0008 exists |
| 0007 | two-tier authz: tier-1 `CanDo` middleware + tier-2 `authz.Require` | Grep `func.*CanDo` and `authz.Require` in backend |
| 0008 | 7 computed tokens, fixed catalog, no fill-in form | Grep the placeholder catalog (7 tokens) per `wiki/concepts/placeholders.md` |
| 0009 | PDF outbox `pdf_dispatch_outbox` + `PDFOutboxWorker` | Grep `pdf_dispatch_outbox`, `PDFOutboxWorker` |
| 0010 | soft-archive via `archived_at` timestamp | Grep `archived_at` in documents domain |
| 0011 | atomic CD create; `cd_sequence_counters` keyed (tenant,profile,area) | Grep `cd_sequence_counters`; confirm key columns in migration |
| 0012 | contract-first; hand-written HTTP structs prohibited for migrated modules | Confirm `api.gen.go` exists + oapi-codegen config present |
| 0013 | `revision_number` on `templates_template_version`; chip `REV{nn}` | Grep `revision_number` + `templates_template_version` |
| 0014 | service renamed to `apps/docx-renderer/`; `DOCX_RENDERER_*` env | Glob `apps/docx-renderer/**`; Grep `DOCX_RENDERER_` |
| 0015 | `Pin` in-tx + `Materialize` async; tx mandatory (amended Z-5) | Grep `func.*Pin`, `func.*Materialize`; confirm no optional-tx-enlistment path remains |
| 0016 | 4 View-grade caps `metrics.view`/`membership.view`/`user.view`/`taxonomy.view` | Grep each cap string in the capability registry |
| 0017 | replay fingerprint = `sha256(docID,decision,reason,contentHash)`; `ErrConflict`→409 | Grep the fingerprint construction in `signoff`/approval |
| 0018 | route lifecycle: 2-state, OCC via `If-Match`, `governance_events` audit | Grep `governance_events`; route lifecycle handler |
| 0019 | `CapAuditRead`, `CapSessionManage` exist in registry | Grep `CapAuditRead`, `CapSessionManage` |
| 0020 | 6-tab IA (7 in practice w/ memberships on this branch); one route+cap per tab | Grep admin-center tab config; confirm memberships tab present |
| 0021 | `/admin/*` tenant surface; platform under `/platform/*`; FE pivots on caps | Grep route prefixes; confirm no role-name FE gates per ADR |
| 0022 | single typed capability registry; area-grade CI-enforced; 13 phases done | Grep `api-lint` rules `no-inline-capability`/`authz-area-scope-binding`; registry file |
| 0023 | positive `x-authz-area` markers; `authz-call-present` lint **deleted** | Grep `x-authz-area` in openapi; confirm `authz-call-present` lint gone |
| 0024 | one base path `servers.url: /api/v1`; `PATH-BASE-PREFIX` gate | Grep `servers:` block in `openapi.yaml`; `PATH-BASE-PREFIX` in api-lint |
| 0025 | RFC 9457 `Problem` only; `ApiErrorEnvelope` retired; `ENVELOPE-DRIFT` blocking | Grep `Problem` type + confirm `ApiErrorEnvelope` absent; `ENVELOPE-DRIFT` rule |
| 0026 | ABAC path dead; `document_access_policies` dropped (migration 0232) | Confirm migration 0232 drops the table; Grep `AccessPolicy` is dead |
| 0027 | `auth_identities` tenant-global; RLS on all tenant-scoped tables (migration 0237) | Read `db/migrations/0237_rls_all_tenant_tables.sql`; **count the tables → resolves the 27/29 mismatch** |
| 0028 | `/audit/events` nested `page.{next_cursor,has_more}`; FE dual-shape adapter removed | Grep audit events cursor shape in handler |

Record the actual command and observed result in each row. If any ADR returns **DECISION-DRIFT** (the decision itself no longer holds in code), **STOP** — that is HS-6 (off-plan for a docs milestone); surface it to the operator, do not edit the decision.

- [ ] **Step 3: Commit the ledger**

```bash
git add docs/superpowers/milestones/grade-a-architecture-remediation/milestone-0-docs-destaling/f0.1-adr-audit-ledger/drift-ledger.md
git commit -m "docs(m0): F0.1 ADR decision-vs-code drift ledger"
```

---

### Task 2: Apply STATUS-DRIFT corrections (only ADRs the ledger flagged)

**Files:**
- Modify: each `wiki/decisions/00NN-*.md` whose ledger verdict is `STATUS-DRIFT`.

- [ ] **Step 1: For each STATUS-DRIFT ADR, correct only the `> **Status:**` line**

Use a vocabulary value from `wiki/decisions/README.md:13-22`. Do not touch the decision body. If the ledger found zero STATUS-DRIFT rows, skip this task and note "no status drift" in evidence.

- [ ] **Step 2: Re-run the status-presence gate (must stay green)**

Run:
```powershell
foreach ($f in Get-ChildItem wiki/decisions/00*.md) { if (-not (Select-String -Path $f -Pattern '^> \*\*Status:\*\*' -Quiet)) { "NO STATUS: $($f.Name)" } }
```
Expected: **no output**.

- [ ] **Step 3: Commit (only if changes were made)**

```bash
git add wiki/decisions/00*.md
git commit -m "docs(m0): F0.1 correct drifted ADR status lines"
```

---

### Task 3: Refresh the `index.md` ledger

**Files:**
- Modify: `wiki/decisions/index.md`

- [ ] **Step 1: Fix the two known drifts**

1. Line 34 stale `docs/adr/` reference: `docs/adr/` is deleted. Replace the sentence with the post-deletion reality (legacy ADR material was removed; the governance migration map — F0.5 — records where it went). Phrase it so it does not re-introduce a link to a deleted path.
2. 27-vs-29 table count: set the `index.md:31` relevance text and ADR `0027`'s status to the **same** number — the one proven by counting tables in `db/migrations/0237_rls_all_tenant_tables.sql` (Task 1, ADR 0027 row). Both surfaces must agree.

- [ ] **Step 2: Reconcile every index row against its ADR status**

For each of the 27 rows, confirm the index `Status` column matches the ADR file's `> **Status:**` line (post-Task-2) and the `Current relevance` text matches the Task 1 observed state. Update any mismatched row.

- [ ] **Step 3: Bump the `Last verified:` stamp**

Set `index.md:3` to: `> **Last verified:** 2026-06-14 (M0/F0.1 — decision-vs-code re-audit; docs/adr stale ref removed; RLS table count reconciled)`.

- [ ] **Step 4: Commit**

```bash
git add wiki/decisions/index.md
git commit -m "docs(m0): F0.1 refresh decisions ledger (stale ref + count reconcile + stamp)"
```

---

### Task 4: Final acceptance gate (F0.1 done-when proof)

**Files:** none (verification only).

- [ ] **Step 1: Status-presence gate green**

Run the script from Task 2 Step 2. Expected: **no output**.

- [ ] **Step 2: Every ADR has a vocabulary-valid status**

Run:
```powershell
$vocab = 'Accepted|Historical|Superseded by ADR|Deprecated|Proposed'
foreach ($f in Get-ChildItem wiki/decisions/00*.md) {
  $s = (Select-String -Path $f -Pattern '^> \*\*Status:\*\*' | Select-Object -First 1).Line
  if ($s -notmatch $vocab) { "BAD STATUS: $($f.Name) => $s" }
}
```
Expected: **no output**.

- [ ] **Step 3: No surviving `docs/adr/` reference in the decisions ledger**

Run:
```powershell
Select-String -Path wiki/decisions/*.md -Pattern 'docs/adr'
```
Expected: **no output** (any hit is stale-ref drift → fix in Task 3).

- [ ] **Step 4: Drift ledger shows zero unresolved DECISION-DRIFT**

Confirm `drift-ledger.md` has 25 rows, every verdict is `MATCH` (or a corrected STATUS/LEDGER-DRIFT), and **no** `DECISION-DRIFT` left open. Any open DECISION-DRIFT = HS-6, stop.

- [ ] **Step 5: Write the feature evidence row**

Copy `.claude/skills/milestone/templates/feature-evidence.md` → `.../f0.1-adr-audit-ledger/evidence.md`; fill commands + observed output from Steps 1–4, the ledger link, review disposition, and any bounded defers (e.g. index relevance text that depends on F0.2/F0.5 — note the trigger).
