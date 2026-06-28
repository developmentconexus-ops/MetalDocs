# System-impact analysis — <work title>

**Date:** YYYY-MM-DD
**Intent (one line):** <what the operator asked for>
**Work type:** module | feature
**Author:** developing-new-work skill
**Verdict:** 🟢 Green | 🟡 Yellow | 🔴 Red  *(fill in §10)*

> Same ten sections for module and feature work. For a feature, mark module-only rows **N/A** with a
> one-line reason — do not delete them. Every row is a question the system forced you to answer.

---

## 1. Classify & own
*(CLAUDE.md Orientation rule)*

- **Work type:** module | feature
- **Owning module(s):** `<module>` — why it owns this.
- **Explicitly NOT owning:** `<module>` — why not (prevents the wrong-module trap).
- **Cross-module edges (with direction):** `A → B` means A depends on B. List each edge and confirm
  it goes through B's published Go interface, never B's repo/SQL/domain.
- **Ambiguity?** If the owning module is unclear → record **AS-3**, resolve before continuing.

## 2. Foundation verdict
*(Global-Maximum rule)*

- **Base you'd build on:** <what exists today>.
- **Sound, or legacy/patch/workaround?** <judgement, with evidence>.
- **If patchy:** name the global-maximum structure (e.g. a kernel/framework boundary, not a one-off
  tweak) + the trade-off. If the work would optimize *inside* a patch → record **AS-2**, operator
  decides.

## 3. Invariant alignment
*(the 6 non-negotiables — see `references/invariant-checklist.md`)*

| Invariant | Touched? | How satisfied | Helper to reuse |
|-----------|----------|---------------|-----------------|
| AuthZ = capabilities, never roles | | | |
| Contract-first (OpenAPI + oapi-codegen) | | | |
| Multi-tenant pooled (`tenant_id` / tx-local GUC / 404 cross-tenant) | | | |
| Async = transactional outbox | | | |
| DB enforces invariants (triggers/constraints) | | | |
| Cross-module via published interface only | | | |

Any violation → record **AS-1**.

## 4. Capability wiring
*(if the work adds/changes a capability — see `references/capability-wiring.md`; else **N/A**: <reason>)*

Walk the 10 ordered touchpoints; record the concrete cap name(s) and the per-touchpoint decision:
1. const + `validCapabilities` · 2. scope classify · 3. tier-1 route→cap · 4. tier-2 `authz.Require`
in-tx · 5. seed grants · 6. DB tripwire · 7. guard tests green · 8. bump `TestCapabilityRegistrySize`
· 9. CI capability-coherence (REQ-AUTHZ-5) · 10. H-PRE-1 off-tx authz read.

## 5. Module wiring
*(if the work births a module — see `references/module-wiring.md`; else **N/A**: <reason>)*

Record the ordered birth steps for the new module: folders → domain + `port.go` → application +
`ports.go` → infra repo (own tables only) → delivery `Handler` + `RegisterRoutes` → api
`cfg.yaml` + `gen.go` → OpenAPI tag + tagged routes → optional `module.go` → composition-root wiring
→ migration → wiki docs.

## 6. Frameworks to reuse, not reinvent
*(see `references/frameworks-catalog.md`)*

List each platform primitive the work will use and confirm reuse (not a hand-rolled equivalent):
`TxRunner`, `tenant.FromContext`, `authz.SeedTxIdentity`/`Require`, `problem.New`/`Write`,
`httpresponse`, `audit.NewEvent`/`Record`, outbox repo, `contracts.Decode`, `testdb` factory.

## 7. Contract & data

- **OpenAPI-first:** routes added/changed in `api/openapi/v1/openapi.yaml` → module `cfg.yaml` + `gen.go`
  → regenerate. New tags? New DTOs?
- **Migration:** `db/migrations/0NNN_*.sql` plan; `tenant_id` on every new tenant table; DB
  constraints/triggers for invariants.
- **Destructive change?** expand/contract plan (never break the live contract in one step).

## 8. Test & QA plan
*(see `references/test-qa-gates.md`)*

- **Canonical framework:** `testdb` integration factory; `//go:build integration`; R1–R4 discipline.
- **Which of the 6 QA gates apply** (module vs feature).
- **Evidence shape:** commands (`go build ./...`, `go test ./...`, `.\scripts\check-system-runnable.ps1`)
  + outcomes + review/QA disposition + bounded defers.

## 9. Docs / ADR
*(see `references/docs-adr-governance.md`)*

- **Wiki:** `wiki/modules/<name>.md` (12-section) + `<name>-tech-debt.md` + `wiki/modules/index.md`
  (module work); affected doc + `Last verified` refresh (feature work).
- **REQ IDs cited:** <IDs from `wiki/architecture/backend-target-architecture.md`>.
- **ADR required?** yes/no — a MUST-deviation or a policy change ⇒ yes; which ADR is created/superseded.

## 10. Verdict & locked constraints

- **Verdict:** 🟢 Green (proceed) | 🟡 Yellow (proceed; ADR/risk flagged) | 🔴 Red (STOP; redesign gate first).
- **Open hard-stops:** AS-1 / AS-2 / AS-3 — none, or list with resolution status.
- **Locked constraints handed to brainstorming:** the non-negotiables the design must honor
  (e.g. "new caps `token.view` + `token_dictionary.manage` ⇒ bump registry size", "supersede ADR 0008",
  "render must not depend on templates").
