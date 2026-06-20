# HS-6 — Scope-shape decision (M0/F0.2 census)

> **Raised:** 2026-06-20 · **Status:** RESOLVED 2026-06-20 (operator) · **Unblocks:** F0.3 guard allowlist, M0 gate
> **Trigger:** mission §9 HS-6 ("M0 census surfaces in-scope sites that change a milestone's shape → stop,
> surface, replan"). Evidence: `./census.md` Part 2.

## What the widen found

The brief's "~20" reproduce cleanly and route to M1–M4 unchanged. Widening to the **full owned-base-table
set** (per F0.2 spec Q2) surfaced two things the mission's milestones do not currently cover:

1. **N1 — document-domain (in-shape):** `documents/application/fillin_service.go:225` raw-reads
   `templates_template_version` (templates) for `placeholder_schema`. Same class as Category B; fits M2 by
   adding one templates read-port (ADR 0030 `TemplateVersionPort` precedent).

2. **X1–X8 — auth/audit/platform (contested):** raw cross-module base-table reads the mission §2 Non-Goal
   excluded as "already ported", but which are in fact **still raw SQL**:
   - `security` → `auth_identities` (×3), `auth_sessions` (×2), `audit_events` (×1)
   - `iam/observability` → `audit_events` (×4), `auth_identities` (×2)
   - `iam/presence` → `auth_identities` (×1)
   - `templates.ListAudit` → `audit_events` (×1)
   - `jobs/stuck_instance_watchdog` → `approval_instances`, `approval_stage_instances`

   Under **ADR-0039 D1 broad reading** these are violations. The mission **terminal §8 bar requires H-G=0
   under the broad reading** → they would be flagged by the terminal re-audit. The Non-Goal and the terminal
   bar contradict each other. That contradiction is the operator's to resolve (it is not mine to narrow or
   expand unilaterally — per CLAUDE.md "stop on architecture contradictions").

## The decisions required

### Decision 1 — N1 (fillin → templates_template_version)
- **Option 1a (recommended):** fold N1 into **M2** as one added templates `placeholder_schema` read-port.
  Small, in-shape, precedented (ADR 0030). Updates M2's feature count by one.
- **Option 1b:** record N1 out-of-scope with reason (defer to a successor mission).

### Decision 2 — X1–X8 (auth / audit / platform reads)
- **Option 2a (recommended): principled ADR-0039 exemption + guard allowlist; mission scope unchanged.**
  Encode in ADR-0039 D3 two exemption *categories*, then allowlist these sites in the F0.3 guard:
  - **audit_events = platform append-sink** — `audit` owns the table but it is a cross-cutting telemetry
    sink every module writes via `AppendAudit[Tx]`; cross-module *read* projections of a platform sink are a
    different class than reading a domain module's private table. (Covers X3, X4, X7.)
  - **auth reads already ADR-dispositioned** — `auth_identities`/`auth_sessions` cross-module reads in
    security/iam were audited and accepted by the **parent grade-a M4** under ADR 0029/0031 (0031 explicitly
    sanctions `= ANY(ids)` scoping because `auth_identities` has no `tenant_id`). They are dispositioned, not
    unported; re-opening them re-litigates a closed parent decision. (Covers X1, X2, X5, X6.)
  - **jobs (X8):** decide whether the `jobs` worker layer is subject to the module-boundary rule at all
    (it is infrastructure operating *on* the approval domain, not a peer domain module). Recommend: exempt as
    worker-layer, note for a future jobs-boundary pass.
  This keeps the mission's shape (M1–M4 unchanged) and makes "H-G=0 under both readings" *coherent* by
  recording the exemptions as principled, not by pretending the reads don't exist.
  → Requires an **ADR-0039 amendment** (add the two exemption categories + jobs note) = HS-4 back to F0.1.
- **Option 2b:** expand the mission with a new milestone (M5) that ports X1–X8 to views/ports. Largest scope
  increase; re-does parent-M4 work; likely out of this mission's appetite.
- **Option 2c:** accept as-is, no ADR change; let the terminal re-audit flag-and-exempt them ad hoc.
  Weakest — leaves the Non-Goal/terminal-bar contradiction unresolved and the guard scope ambiguous.

## Recommendation

**1a + 2a.** Fold N1 into M2; resolve X1–X8 via a principled ADR-0039 exemption (platform sink +
parent-ADR-dispositioned auth + worker-layer jobs) recorded in the ADR and the F0.3 allowlist. This holds the
mission's appetite, keeps M1–M4 shape, and makes the terminal "both readings" bar honest. Cost: one ADR-0039
amendment (HS-4 → F0.1) before F0.3 finalizes its expected-red/allowlist sets.

## Resolution (operator) — 2026-06-20

**Ruling:** Decision 1 → **Option 1a** (fold N1 into M2). Decision 2 → **Option 2a** (principled ADR-0039
exemption; mission shape unchanged).

**Actions taken:**
1. **ADR-0039 amended** (HS-4 → F0.1): added exemptions **D3(d)** platform append-sink, **D3(e)**
   parent-ADR-dispositioned auth, **D3(f)** worker-layer; added the N1 + X1–X8 classification table and an
   honest "0 violations *outside the recorded allowlist*" scope note for the terminal bar. Header marked
   *amended 2026-06-20*. (`wiki/decisions/0039-*.md`)
2. **N1 folded into M2:** recorded in the mission work inventory + M2 (mission §5/§7 replan note,
   HS-6-authorized). M2 gains one templates `placeholder_schema` read-port.
3. **F0.3 guard allowlist** will enumerate X1–X8 with per-site D3(d)/(e)/(f) justification; expected-red set =
   N1 + Categories A/B/C/C4 (the to-be-ported sites), expected-green = X1–X8 + already-compliant.

**Result:** census remains **0 unclassified**; mission appetite held (M1–M4 shape unchanged, M2 +1 port);
terminal "both readings" bar redefined honestly as 0-outside-allowlist. F0.3 unblocked.
