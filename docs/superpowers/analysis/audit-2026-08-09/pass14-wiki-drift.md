# Pass 14 — Wiki / Memory Drift Audit

**Date:** 2026-08-09
**Baseline:** main @ 418070bf
**Status:** reproduced-current (every drift item below verified by direct file read against this baseline in the `arch-audit` worktree; DO NOT edit wiki — record only, per audit brief)
**Note on remediation ownership:** every drift item below is scoped to the audit program itself (F-AUD-05 — wiki-sync follow-up), not filed as a fresh, standalone issue. None of these are new defects discovered independently of this audit pass.

---

## Classification key

- **current** — matches code truth as of 418070bf, no action needed.
- **stale** — was true, code moved on, doc didn't; low risk if surrounding context is otherwise sound.
- **historical** — correctly framed as a past-state record (changelog entry, "Prior:" note); not drift.
- **target-only** — describes an intended future state, correctly labelled as such; not drift.
- **contradictory** — two truths on file (doc vs. doc, or doc vs. code) with no cross-link reconciling them; higher risk than plain staleness.
- **dangerous** — could actively mislead an implementer into wrong action (wrong enforcement claim, wrong file path used as ground truth, wrong module count used to scope a boundary check).
- **cosmetic** — wrong but inert (a stale count in prose that doesn't feed any decision).

---

## DANGEROUS drift (could mislead an implementer)

### D1. `wiki/architecture/backend-blueprint.md:328` — false "MET + CI-locked" claim
**Claim:** REQ-TOP-2 (platform packages must not import modules) is marked **"MET + CI-locked"**, citing a `platformboundary` analyzer that exits 0.
**Code truth:** `tools/cilint/internal/analyzers/platformboundary.go` contains an explicit `platformBoundaryAllowed` map exempting `internal/platform/bootstrap`, `internal/platform/authn`, `internal/platform/docgenv2` — the analyzer passes **because it excludes the known violators**, not because the violation doesn't exist. Confirmed live violations: `internal/platform/bootstrap/api.go` (imports audit/auth/iam), `internal/platform/authn/config.go`, `internal/platform/authn/context.go`, `internal/platform/docgenv2/templates_snapshot_reader.go`.
**Classification:** contradictory (doc claim vs. the guard's own source code) — **dangerous**: a reader trusting "MET" would conclude the invariant is closed when the tooling itself documents it as open, allowlisted debt (see pass13 §3/§9 blind-spot matrix row 4). Cross-referenced with pass13's finding on the same analyzer.

### D2/D3. Dead `lease-reaper` described as live, in two files
**D2 — `wiki/architecture/backend-target-architecture.md`** (line ~16, "Last verified: 2026-06-11"): states `metaldocs-api` hosts 4 leader-elected janitors "via an in-process scheduler with DB-backed lease/heartbeat single-runner semantics," explicitly naming **lease-reaper**.
**D3 — `wiki/backend/repo-topology.md`** (scheduler section, "Last verified: 2026-08-03"): independently describes "lease-reaper on ticker intervals."
**Code truth:** zero grep hits for `lease-reaper`/`lease_reaper`/`LeaseReaper` anywhere in the tree. The real mechanism (post ADR 0067, Accepted 2026-07-04 — *after* backend-target-architecture.md's 2026-06-11 verification date) is River leader-election (`riverjobs`/`riverBundle` in `apps/api/cmd/metaldocs-api/main.go`). **CLAUDE.md itself correctly states** "the old Postgres-lease scheduler and its `lease-reaper` are retired (M5)."
**Classification:** stale, describing a fully deleted subsystem as current — **dangerous**, and worse for D2 because `backend-target-architecture.md` is the file **CLAUDE.md names as the single governing target spec** ("source of truth when this list drifts"). Two independent docs corroborating each other's staleness compounds the risk that a reader treats it as confirmed rather than checking code.

### D4. `wiki/backend/repo-topology.md:178-190` — module table missing 4 of 15 modules
**Claim:** module table lists 11 modules.
**Code truth:** `internal/modules/` has exactly 15: approval, audit, auth, controlleddocuments, distribution, documents, iam, jobs, notifications, render, search, security, taxonomy, templates, tokens. The table at lines 178-190 **omits approval, distribution, notifications, and tokens entirely** — not merely a wrong count in prose, but a structurally incomplete enumeration.
**Classification:** stale — **dangerous**, because a reader using this table to scope a cross-module boundary check (exactly the kind of task this audit itself is performing) would silently miss 4 real modules, including `approval` (the module every recent ADR — 0082, 0083, 0093 — centers on).

### D5. `wiki/modules/index.md` — canonical module index missing 2 of 15, one module undocumented entirely
**Claim:** "Core product modules" section lists 9 modules (approval, audit, auth, controlled-documents, documents, iam, taxonomy, templates, tokens).
**Code truth:** checked against the filesystem — **`distribution` has no wiki module doc at all** (no `wiki/modules/distribution.md`, not referenced in any section of the index), and **`wiki/modules/security.md` / `security-tech-debt.md` / `security-signals.md` exist on disk but are linked nowhere in `index.md`** (checked Core, Frontend-focused, and Supporting sections — none mention `security`). A reader following the index as the module map gets 13 of 15, silently missing 2, with `distribution` having zero documentation surface anywhere.
**Classification:** stale/incomplete — **dangerous**, since `index.md`'s own stated scope is "durable per-module knowledge... first-stop," and its "Last verified" date (2026-08-05) postdates both modules' existence.

---

## CONTRADICTORY drift (two truths, no cross-link)

### C1. ADR 0007 vs. the later authz-grant-unification analysis
**`wiki/decisions/0007-two-tier-authz.md`** — Status "Accepted (amended 2026-05-11)," frames the tier-1/tier-2 disjoint-grant-table split as settled, intentional design.
**Counter-truth:** `docs/superpowers/analysis/2026-08-06-authz-grant-unification-decisions.md` explicitly states this is "a documented unfinished migration, not a design," calling it "an incomplete migration ratified retroactively as architecture — an unlabelled local maximum under CLAUDE.md's Global Maximum rule." Independently confirmed: tier-1 reads `iam_user_roles ∪ iam_group_members⋈iam_group_roles`; tier-2 reads `role_capabilities⋈user_process_areas` — different role coverage (tier-2's CHECK includes `area_admin`/`qms_admin`/`signer`, tier-1's doesn't).
**Gap:** ADR 0007 has **no "Status history" section** pointing at this later finding — unlike ADR 0072 (see the CURRENT section below), which does exactly this correctly.
**Classification:** contradictory — **dangerous** for anyone reading ADR 0007 in isolation and concluding the two-tier split is closed, ratified design, when it is tracked debt awaiting a not-yet-filed ADR 0092. This matches this MEMORY file's own note (`authz-grant-model-dual-source`).

### C2. ADR 0093 references pending ADR 0092, which does not exist as a file
**`wiki/decisions/0093-controlled-information-context-template-as-role.md`** (Status: "Accepted as a design ruling 2026-08-07. Not implemented — no code or schema change is authorized by this ADR alone.") forward-references a pending ADR 0092 (authz-grant-unification) as its dependency.
**Code truth:** confirmed via targeted directory listing and find — **no `wiki/decisions/0092-*.md` file exists**. `wiki/decisions/index.md` and ADR 0093 both reference it prospectively; C1's authz-grant-unification analysis doc is presumably the eventual basis for it but is not itself a filed, numbered ADR.
**Classification:** target-only reference correctly labelled as "not implemented" within ADR 0093 itself — **not dangerous on its own** (ADR 0093 is honest about its own non-implemented status), but the dangling forward-reference is worth closing before ADR 0093 is cited as authoritative in any implementation review, since a reader following the citation chain hits a dead end.

---

## STALE drift (correct historically, superseded, low/no operational trap)

### S1. Module-count prose staleness (3 files, same root cause as D4/D5 but purely cosmetic where the count doesn't feed a table/decision)
- `wiki/architecture/backend-blueprint.md` — "12 business modules" (lines 10, 50 mermaid, 175); real count 15. **Cosmetic** — surrounding prose elsewhere in the same doc correctly names newer modules (including `approval`) by name, so the numeric label is stale but not structurally misleading the way D4's incomplete table is.
- `wiki/architecture/backend-target-architecture.md:33` — mermaid diagram labeled "12 bounded-context modules." Same class, same file already flagged dangerous for D2 (the lease-reaper claim is the higher-severity issue in this file).
- `wiki/backend/repo-topology.md:56,176` — "11 business modules" (mermaid + heading). On its own this line is cosmetic, but it directly feeds the incomplete table at D4 — **classified dangerous there, cosmetic here** (the count label itself vs. the table it introduces are two separate defects in the same file).

### S2. `wiki/architecture/backend-target-architecture.md` — 2 months stale overall
"Last verified: 2026-06-11 (Wave 1)" — this is CLAUDE.md's designated governing target spec, yet predates ADR 0082 (2026-07-12, approval promoted to 15th top-level module) and ADR 0093 (2026-08-07, documents/controlleddocuments/templates peer-context ruling) entirely. REQ-TOP-1's current wording ("Cross-module access goes through a module's application service or published Go interface — never another module's repository, SQL, or domain internals. MUST") is now in tension with ADR 0093's later finding that these three modules are *not proven* to be three peer bounded contexts, and is separately, currently violated in production code by the 23 baselined `hgcrossmodule` findings (pass13 §3, §7).
**Classification:** stale, compounding into **contradictory** against ADR 0093 — flagged dangerous already under D2 for the lease-reaper claim in the same file; this is an additional, independent staleness vector in the same document.

### S3. `wiki/backend/repo-topology.md:459` — middleware chain description incomplete
Omits `otel` and `method_not_allowed` from the described chain. Real chain (`apps/api/cmd/metaldocs-api/chain.go:25`, verified against CLAUDE.md's own claim — no drift in CLAUDE.md itself here): `panic_recovery → otel → http_obs → cors → origin_protection → pre_auth_login_rate_limit → authn → iam_authz → presence_bump → rate_limit → method_not_allowed` (11 links).
**Classification:** stale — borderline cosmetic/dangerous; a new-route author relying on this shorter chain could miss that OTel spans and the 405 handler are already-standard links, though the actually-executing code (`chain.go`) is what governs runtime behavior regardless of what this doc says, which caps the real-world blast radius. Scored **cosmetic-leaning**.

### S4. `wiki/modules/iam.md:396` — stale pre-ADR-0082 file path
Cites `internal/modules/documents/approval/application/cancel_service.go:12` as a tier-2+BypassSystem example.
**Code truth:** `internal/modules/approval/application/cancel_service.go` exists; the nested `documents/approval/...` path does not (ADR 0082 moved it, Accepted 2026-07-12). iam.md's own "Last verified" is 2026-07-14 — **two days after** the move — meaning the stale anchor was missed during that very verification pass, not merely predates it.
**Classification:** stale — cosmetic (the cited behavior/invariant is still true; only the path is wrong, so a `grep` on the wrong path returns nothing rather than misleading about behavior).

### S5. `wiki/modules/controlled-documents.md` — most stale of the 5 module docs spot-checked
"Last verified: 2026-06-15" — predates both ADR 0082 (2026-07-12) and ADR 0093 (2026-08-07), zero references to either. Not itself a false claim (no peer-context assertion found in the file to contradict ADR 0093), but silent on the pending merge ruling that directly targets this module.
**Classification:** stale/incomplete, not contradictory — **cosmetic**, but flagged since it is one of ADR 0093's three named merge-candidate modules and hasn't been touched in the 53 days leading up to that ADR.

---

## CURRENT / no drift found (checked and clean — recorded so this audit pass isn't re-run against these)

- **CLAUDE.md** system facts — 15-module list, 4-binary description, and the exact middleware-chain wording all verified directly against `internal/modules/` (`ls` = exactly the 15 named), the 4 `cmd/` binaries, and `apps/api/cmd/metaldocs-api/chain.go:25`. No drift found in CLAUDE.md itself.
- **ADR 0072** (`wiki/decisions/0072-approval-nested-exception-and-boundary-model.md`) — exemplary self-maintenance: explicit "## Status history" section correctly records ruling (a) superseded 2026-07-12 by ADR 0082, rulings (b)/(c) still in force. **This is the positive counter-example the audit should hold up** against C1/D2's failure to do the same.
- **ADR 0082** — Accepted 2026-07-12, correctly scoped ("Supersedes: ADR 0072 ruling (a) only"), documents its own later-executed follow-up inline. No drift.
- **ADR 0083, 0086, 0087, 0022** — all confirmed "Accepted" via direct header read; no contradicting later ADR found for any of the four.
- **`wiki/modules/approval.md`** — "Last verified: 2026-08-05," documents the approval-accountability-loop feature and its known gap (stage-2+ approvers get no notification) accurately and currently; matches this session's own MEMORY note. No stale nested-path references found.
- **`wiki/modules/templates.md`** — no stale nested `documents/approval` path references; contains its own explicit changelog entry self-correcting for the ADR 0082 path change on 2026-07-12 (i.e., unlike iam.md at S4, templates.md's own verification pass caught and fixed its anchor).
- **`internal/modules/jobs/`** — 8 real subdirectories (`approval_sla_surfacer`, `audit_integrity_validator`, `document_review_surfacer`, `idempotency_janitor`, `maintenance`, `outbox_retention`, `release_hold_reconciler`, `stuck_instance_watchdog`, `tenantdata`) match CLAUDE.md's job list exactly.
- **`wiki/decisions/0093`** — internally consistent; correctly labelled not-implemented, correctly forward-references the not-yet-filed ADR 0092 (flagged separately as C2, but the label discipline within 0093 itself is sound).

---

## Top 10 wiki drifts, ranked by risk (for the closing summary)

1. **D1** — `backend-blueprint.md` claims REQ-TOP-2 "MET + CI-locked" when the guard's own source documents 4 live exemptions. Highest severity: a false enforcement claim, not just a stale fact.
2. **D2** — `backend-target-architecture.md` (CLAUDE.md's designated governing target spec) describes the retired `lease-reaper` as live, 2 months past its own verification date and past ADR 0067.
3. **D3** — `repo-topology.md` independently repeats the same dead-`lease-reaper` claim, compounding D2.
4. **D5** — `wiki/modules/index.md`, the stated "first-stop" module reference, is missing `distribution` (undocumented anywhere) and `security` (documented but unlinked) — 2 of 15 modules invisible to a reader trusting the index.
5. **D4** — `repo-topology.md`'s module table omits approval, distribution, notifications, tokens (4 of 15) — structurally incomplete, not just a wrong count.
6. **C1** — ADR 0007 presents the dual-source authz-grant split as settled design with no status-history link to the later finding that it's an unfinished migration.
7. **S2** — `backend-target-architecture.md`'s REQ-TOP-1 wording is now in tension with ADR 0093's later peer-bounded-context ruling and with the 23 live `hgcrossmodule` baseline findings — the "governing spec" is out of sync with both a later ADR and current code.
8. **C2** — ADR 0093 forward-references ADR 0092, which does not exist as a filed document — a dead citation link for anyone following the chain, though ADR 0093 itself is honest about its non-implemented status.
9. **S1** — module-count staleness ("11"/"12" vs. 15) repeated across `backend-blueprint.md`, `backend-target-architecture.md`, and `repo-topology.md` — cosmetic individually, but the same wrong number recurring in 3 independent files raises the odds a reader treats it as corroborated.
10. **S4/S5** — `iam.md`'s stale pre-ADR-0082 file-path anchor (missed by its own 2026-07-14 verification pass, 2 days after the move it should have caught) and `controlled-documents.md`'s 53-day silence on the ADR 0093 merge ruling that names it directly.

**Positive counter-example worth citing:** ADR 0072's "Status history" section and `templates.md`'s self-correcting changelog entry are both examples of the wiki-maintenance discipline working correctly — the drift above is a partial, not systemic, failure of the update-on-change practice.

**Remediation ownership:** all items above are wiki-sync follow-up under the audit program's own tracking (F-AUD-05), not new standalone defects — consistent with the audit brief's instruction not to edit wiki content during this pass.
