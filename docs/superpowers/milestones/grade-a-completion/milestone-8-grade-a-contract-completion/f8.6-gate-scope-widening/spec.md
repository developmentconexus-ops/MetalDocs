# Feature F8.6 — Spec (SEED — approval pending)

> **Milestone:** 8 — Grade-A Contract & Boundary Completion  ·  **Folder:** `f8.6-gate-scope-widening`
> **Status:** Drafting (seed — root-cause fix for the 4th miss; **interview + approval pending**)
> **Approved before code:** PENDING — *no implementation begins until this line is filled (Phase 3, fresh session).*
> **Sequenced LAST** — the widened gate must measure F8.1–F8.5 after they land.

## Interview record (fail-closed gate)

| # | Question | Answer |
|---|----------|--------|
| 1 | Why widen the gate at all? | The §5b H-D greps are path-scoped to `internal/modules/*/delivery/http/` and H-G greps only the two IAM tables; presence/metrics/search evaded every prior bounded sweep. Widening the gate to the true public surface is the root-cause fix that makes the class non-regressable. |
| 2 | Mechanical guard or doc-only? | **Both** — amend §5b + mission §8 (doc contract) AND add a CI guard (grep gate or `tools/cilint` analyzer) failing on response-literal `map[string]any` on any registered route. |
| 3 | Are presence/metrics/health in-scope or exempt? | In-scope (they are public spec-declared routes). Health may be declared an explicit exemption if operator confirms; record it rather than leave it implicit. |

## Consumer contract (FIRST)

- **Consumer(s):** the terminal re-audit + `mission-validator` (they execute §8); future contributors (the CI guard).
- **Contract:** the H-D/H-G class definitions and their grep commands cover **every package that registers a public route** — including `internal/modules/iam/presence/`, `internal/platform/observability/`, and health — not just `internal/modules/*/delivery/http/`. A response-literal `map[string]any` anywhere on a registered route fails the gate.
- **Source of truth:** `wiki/architecture/api-contract.md` §5b; mission `mission.md` §8.

## What this feature implements

1. `api-contract.md` §5b — widen Part A/B grep paths to the full public-route surface; widen the H-G definition to "any cross-module/cross-schema owned-table read."
2. `mission.md` §8 — restate the H-D/H-G bar against the true surface (and record any health exemption).
3. A CI guard (`tools/cilint` analyzer or a committed grep script) that fails the build on a response-literal `map[string]any` on a registered route.

## Non-goals (mandatory)

- No source-behavior change (this feature is gate/doc/CI only).
- No relaxation of the bar — widening only.

## Validation Gate (concrete — approved before code)

| Acceptance criterion | Named test / proof command | Real vs fixture |
|----------------------|----------------------------|-----------------|
| §5b + §8 cover the full public surface | read the amended sections | real |
| Widened H-D grep at HEAD = 0 (post F8.1–F8.5) | the new §5b commands | real |
| Widened H-G grep at HEAD = 0 (post F8.3) | the new §8 commands | real |
| CI guard fails on a planted violation, passes clean | run the guard against a temp violation, then clean | real |

## ADR needed?

- [x] **Durable decision** — amends a governing contract (§5b/§8 scope) → record the scope change as an ADR or an explicit §8 amendment note; link here: _TBD in execution_.
