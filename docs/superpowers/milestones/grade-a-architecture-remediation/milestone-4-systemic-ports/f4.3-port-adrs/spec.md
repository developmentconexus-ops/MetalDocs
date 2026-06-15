# Feature F4.3 — Spec

> **Milestone:** 4 — Systemic Ports (H-G class)  ·  **Folder:** `f4.3-port-adrs`
> **Status:** Approved (pre-code) 2026-06-15 / operator (documentation feature implementing the milestone's
> declared F4.3 acceptance — no contract ambiguity; no interview required).

> This feature is **documentation-only**: it records the two durable port decisions made in F4.1 and
> F4.2 as canonical ADRs in the now-clean `wiki/decisions/` ledger. No code changes.

## Interview record (fail-closed gate)

Not required — F4.3's deliverable, content, and acceptance are fully specified by the `milestone.md`
F4.3 row; the decisions themselves were already settled (and operator-delegated/approved) in F4.1/F4.2.

## Consumer contract (the milestone.md F4.3 acceptance — read, not invented)

The "consumers" of an ADR are the engineers and the milestone-validator; the contract is the F4.3 row:

> Two ADRs exist under `wiki/decisions/` with canonical `Status:` headers, registered in the decisions
> `index.md`, cross-linked from F4.1/F4.2 `spec.md`, and referenced by the touched module wiki docs.
> Each ADR records the design D4/Approach-3 constraint (reads live, no migration) and alternatives
> rejected (incl. Approach 2).

## What this feature implements

1. **ADR 0029 — `UserDisplayNameReader` port** (`wiki/decisions/0029-user-display-name-reader-port.md`):
   context = H-G `iam_users` display-name reach; decision = iam-owned port (single + batch); consequences
   = reads live / no snapshot / off-tx (H-PRE-1); alternatives rejected (Approach 2 freeze-name; the
   security tenant-scope JOIN recorded as a bounded defer, out of this port's scope).
2. **ADR 0030 — template-version state port** (`wiki/decisions/0030-template-version-state-port.md`):
   context = H-G `templates_*` reach + `status := "published"` hardcode; decision = **extend** the
   existing templates-owned `TemplateVersionPort` (not introduce a parallel one) with raw
   `GetTemplateVersionState`, `IsPublished` kept for taxonomy; consequences = reads live / no snapshot /
   off-tx; alternatives rejected (Approach 2; parallel duplicate reader).
3. Register both in `wiki/decisions/index.md`; cross-link from F4.1 + F4.2 `spec.md`; reference from the
   touched module wiki docs (iam, documents, approval, templates, controlled-documents) via wiki-curator.

## Non-goals (mandatory)

- **No** code change of any kind (F4.1/F4.2 already landed the code).
- **No** new decisions — ADRs only *record* decisions already made; no scope beyond the two ports.
- **No** broad wiki sweep — module-doc edits are limited to the ADR cross-link + stamp bump.
- **No** ADR for the deferred security tenant-scope port (that is a future feature when the defer trips).

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Proof command | Real vs fixture |
|----------------------|---------------|-----------------|
| Two ADRs exist with canonical `Status:`/`Last verified:` headers | `ls wiki/decisions/0029-*.md wiki/decisions/0030-*.md`; headers present | real |
| Both registered in `index.md` | `grep -nE '0029|0030' wiki/decisions/index.md` → two rows | real |
| Cross-linked from F4.1 + F4.2 specs | `grep -n '0029' f4.1-*/spec.md`; `grep -n '0030' f4.2-*/spec.md` | real |
| Referenced by touched module wiki docs | `grep -rlE '0029|0030' wiki/modules/` → iam, documents, approval, templates, controlled-documents | real |
| Each ADR records reads-live/no-snapshot + alternatives rejected | manual read of both ADRs | real |
| No code changed by F4.3 | `git diff --name-only <f4.2 commit>..HEAD` for this feature touches only `wiki/` + `docs/superpowers/` | real |

## ADR needed?

- [x] N/A — this feature **is** the ADR authoring. The two ADRs are its deliverable.
