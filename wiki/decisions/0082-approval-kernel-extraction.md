# ADR 0082 — Approval kernel extraction to a first-class module (supersedes ADR 0072)

> **Status:** Accepted 2026-07-12
> **Supersedes:** [ADR 0072](0072-approval-nested-exception-and-boundary-model.md) ruling (a) only.
> **Scope:** ROADMAP unit 3.1 / approval-remediation M3 — promote
> `internal/modules/documents/approval` to a top-level bounded context `internal/modules/approval`
> (the 15th module), realign the module-boundary guard to treat it as first-class.

## Context

ADR 0072 (Accepted 2026-07-06, M9 F9.5) **rejected** promoting `documents/approval` to its own
top-level module *at that time*, on the ground that it was one DDD aggregate with dense bidirectional
coupling to `documents`, and that splitting it mid-hygiene-milestone would be an interface redesign
with no functional gain. Crucially, ADR 0072 recorded a **named promotion trigger** verbatim:

> *"if approval ever needs an independent lifecycle — its own deploy cadence, its own owning team, or
> a **second bounded-context consumer that isn't `documents`** — the promotion plan starts from this
> ADR's coupling-edge inventory rather than re-deriving it."*

**The trigger has fired.** The ratified review/approval workflow model (spec
`docs/superpowers/specs/2026-07-08-approval-workflow-coherence-design.md` §5, Milestone C) rewires the
`templates` module's approval onto the same kernel — a **second bounded-context consumer that is not
`documents`**. Executing the promotion now is following ADR 0072's own recorded plan, not reversing
its reasoning.

### Live coupling census (2026-07-12, production `.go`, non-test)

ADR 0072's 2026-07-06 "dense, bidirectional, 100+ edges" estimate counted test edges and is stale.
The live production coupling is materially lighter:

| Edge | Count | Layer | Post-extraction disposition |
|---|---|---|---|
| `documents` → `approval` | 2 files | `application` (delivery/http handler), `domain` (infrastructure/active_instance_reader) | Allowed — published layers; import-path rename only |
| `approval` → `documents` | 24 | 17 `domain` + 7 `application` | Allowed — published layers (approval now depends on documents' published surface) |
| `jobs/approval_sla_surfacer` → `approval` | 1 | `domain` | Allowed |
| `jobs/stuck_instance_watchdog` → `approval` | 1 | `application` | Allowed |
| `templates` → `approval` (new, M3 P3) | — | `application`/`api` | Allowed — published surface only |
| `audit` → `approval/http/router` | **0** | — | **False positive** — the reference is a code *comment*, not a Go import (`sed -n '/^import (/,/^)/p'` confirms no approval import in the audit handler). No re-port required. |

**True cross-module violations after the pure relocate: 0.** `check-module-boundaries.ps1` →
`[module-boundaries] OK` on the relocated tree, because approval's published surface
(`domain`/`application`/`api`) is already covered by the guard's layer allow-list.

## Decision

### (a) `documents/approval` → `internal/modules/approval` — promoted (reverses ADR 0072 ruling (a))

`documents/approval` becomes the top-level bounded context `internal/modules/approval` (the 15th
module). The relocation is a **pure move** (`git mv`, 165 renamed files, byte-identical staged content;
import prefix `metaldocs/internal/modules/documents/approval` → `metaldocs/internal/modules/approval`
across 111 `.go` files + lint/staticcheck path-string consumers; one `//go:generate` relative-path
depth fix). ZERO behavior change — `go build`/`go vet`/unit suites green post-move.

The kernel generalizes from document-specific keying to `(subject_kind, subject_key)` (M3 phases 2–3;
`document+profile_code` existing rows backfilled, `template+doc_type` added). `documents` and
`templates` consume the kernel through its published application-service / api surface only.

### (b) Boundary-guard realigned — approval is first-class, nested-exception retired

`scripts/check-module-boundaries.ps1` no longer special-cases a `documents/approval` nested family.
`approval` is treated exactly like any other module: cross-module imports may target only its published
surface (`domain`/`application`/`api`). This is **stricter** than the old nested model, which allowed
edges between `documents` and `documents/approval` **at any layer**: `documents` may now no longer reach
`approval/http` or `approval/infrastructure`, only approval's published layers (and vice-versa). The
dead `$approvalPublishedExtra` variable and the `$bothInDocumentsFamily` bypass are removed.

**Proofs (P1.S4, mirroring ADR 0072's discipline):**
1. GREEN on the relocated tree with the realigned guard.
2. Negative plant: a blank import `_ "metaldocs/internal/modules/approval/infrastructure"` added to a
   genuinely-external module (`internal/modules/jobs/stuck_instance_watchdog/job.go`) → guard RED,
   naming exactly `internal/modules/jobs/stuck_instance_watchdog/job.go ->
   metaldocs/internal/modules/approval/infrastructure`.
3. Revert: `git diff --exit-code` on the planted file = clean; guard GREEN again.

### What ADR 0072 rulings survive

Rulings **(b)** (one `infrastructure/` persistence directory per module) and **(c)** (the boundary-guard
allow-model keyed to REQ-TOP-1: layer allow-list + `$publishedPackages` + empty `$debtAllowList`) remain
in force unchanged. Only ruling (a) — "approval stays nested" — is superseded.

## Consequences

- **Positive.** Approval is a clean one-module-one-directory bounded context; the boundary guard is now
  stricter around it (persistence/http no longer cross-reachable between documents and approval);
  templates and documents share one approval kernel instead of two parallel approval implementations.
- **Costs.** 111 import paths moved (mechanical, compiler-verified). `documents` and `approval` now have
  a legitimate cross-module dependency in both directions, each through published layers — this is
  normal inter-module coupling, guarded by the layer allow-list.
- **Superseded trigger.** ADR 0072's promotion trigger is now consumed; no further "promote approval"
  decision is pending.

## Alternatives considered

- **Keep the nested exception, add templates as a third intra-documents nest** — rejected: templates is
  not part of the documents aggregate; nesting it under documents to reuse approval would couple two
  unrelated bounded contexts and defeat REQ-TOP-1.
- **Two parallel approval implementations (status quo)** — rejected: the ratified model requires the
  same eQMS rigor (SoD, quorum, delegation, e-signature, instance state machine) for template versions
  as for documents; duplicating the kernel is the local maximum ADR 0072 warned against.
