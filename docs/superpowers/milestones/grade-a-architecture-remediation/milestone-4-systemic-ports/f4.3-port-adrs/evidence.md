# Feature F4.3 — Evidence

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Feature:** `f4.3-port-adrs`  ·  **Closed:** 2026-06-15
> **Contract:** [`spec.md`](spec.md) (the milestone.md F4.3 acceptance). Documentation-only — no code.

## What was implemented

- **ADR 0029** [`wiki/decisions/0029-user-display-name-reader-port.md`](../../../../../../wiki/decisions/0029-user-display-name-reader-port.md)
  — iam-owned `UserDisplayNameReader` boundary; context (H-G `iam_users` reach), decision (owning-module
  port, single+batch), consequences (reads live, no snapshot, off-tx/H-PRE-1), alternatives rejected
  (Approach 2 freeze-name; security tenant-scope JOIN bounded-deferred, out of scope).
- **ADR 0030** [`wiki/decisions/0030-template-version-state-port.md`](../../../../../../wiki/decisions/0030-template-version-state-port.md)
  — templates-owned `TemplateVersionPort` **extended** (not duplicated) with raw `GetTemplateVersionState`;
  context (CD `templates_*` reach + `status := "published"` hardcode), decision (extend owned port,
  `IsPublished` kept for taxonomy), consequences (reads live, no snapshot, off-tx), alternatives rejected
  (Approach 2; parallel duplicate reader).
- Registered both in `wiki/decisions/index.md`; cross-linked from F4.1 + F4.2 `spec.md`; module-doc
  cross-links + stamp bumps added by `wiki-curator` to iam.md, documents.md, approval.md, templates.md,
  controlled-documents.md.

Committed in this feature's commit (subject: `docs(milestone-4): F4.3 port ADRs 0029/0030 — UserDisplayNameReader + TemplateVersionPort boundaries`).

## Verification

| Check | Command / action | Result (evidence) | Real vs fixture |
|-------|------------------|-------------------|-----------------|
| Two ADRs exist, canonical headers | `ls wiki/decisions/0029-*.md wiki/decisions/0030-*.md` | both present; each opens with `> **Status:** Accepted 2026-06-15` + `> **Last verified:** 2026-06-15` | real |
| Registered in index | `grep -cE '\[0029\]\|\[0030\]' wiki/decisions/index.md` | `2` | real |
| Cross-linked from F4.1 spec | `grep -c 0029 f4.1-*/spec.md` | `1` | real |
| Cross-linked from F4.2 spec | `grep -c 0030 f4.2-*/spec.md` | `1` | real |
| Referenced by touched module docs | `grep -rlE '0029-user-display-name\|0030-template-version-state' wiki/modules/` | approval.md, controlled-documents.md, documents.md, iam.md, templates.md (5/5) | real |
| Reads-live/no-snapshot + alternatives recorded | manual read | both ADRs carry the D4/Approach-3 reads-live constraint + a rejected-alternatives clause | real |
| No code changed by F4.3 | this feature's diff is `wiki/` + `docs/superpowers/` only | no `.go` files in F4.3 commit | real |

## Acceptance vs spec Validation Gate

| Acceptance criterion (from spec.md) | Met? | Evidence |
|-------------------------------------|------|----------|
| Two ADRs exist with canonical headers | yes | row 1 |
| Both registered in `index.md` | yes | row 2 |
| Cross-linked from F4.1 + F4.2 specs | yes | rows 3–4 |
| Referenced by touched module wiki docs | yes | row 5 |
| Each ADR records reads-live/no-snapshot + alternatives rejected | yes | row 6 |
| No code changed by F4.3 | yes | row 7 |

## Review disposition

- Spec-compliance review: **PASS** — both ADRs match the F4.3 row deliverable (owner framing for the
  owning module, consumer framing for consumers; D4/Approach-3 constraint + rejected alternatives present).
- Code-quality review: N/A (no code). Doc-quality: ADRs follow the canonical `wiki/decisions/` template
  (Status/Last verified/Scope/Key files/Context/Decision/Consequences/References).

## Bounded defers

| Defer | Why bounded | Written trigger / owner |
|-------|-------------|-------------------------|
| ADR for an iam tenant-scope/membership port (security JOIN) | Not a decision made yet — only made when the defer trips | Next structural touch of `security/infrastructure/postgres/repository.go` or M5 re-audit flag; owner backend (recorded in ADR 0029 + F4.1 spec) |
