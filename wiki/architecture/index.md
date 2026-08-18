# Architecture

> **Last verified:** 2026-08-18
> **Scope:** Durable system architecture truth and active target-design routing.

## Active target design — read first

- **[launch-v1-product-contract.md](launch-v1-product-contract.md)** — **ACCEPTED Launch V1 product authority.** Required capabilities, journeys, invariants and scope tiers.
- **[whole-product-alignment-review.md](whole-product-alignment-review.md)** — **OPERATOR-ADJUDICATED Whole-Product GCR authority.** A1–A10 accepted; records the structural rebaseline.
- **[launch-v1-ownership-topology.md](launch-v1-ownership-topology.md)** — **OPERATOR-APPROVED semantic ownership authority.** Launch = Authentication + Organization + Authorization + Controlled Documents + Audit, with future-evolution law.
- **[r10-technical-architecture.md](r10-technical-architecture.md)** — **ACTIVE REBASELINED TECHNICAL-STAGE AUTHORITY.** T1→T7 decomposition approved; T1 is currently open.
- [../references/current-agent-handoff.md](../references/current-agent-handoff.md) — exact fresh-session checkpoint / next gate.
- [../standards/root-cause-global-maximum-method.md](../standards/root-cause-global-maximum-method.md) — binding DevelopmentConexus Engineering Method v1.0.0 mirror.

## Active staging

- `../../docs/superpowers/analysis/2026-08-18-r10-t1-semantic-state-invariants-candidate.md` — **NON-AUTHORITATIVE T1 candidate / operator adjudication packet.**

Staging does not become target authority until operator adjudication and promotion.

## Active Launch ownership

```text
BUSINESS
Authentication
Organization
Authorization
Controlled Documents

SUPPORTING SEMANTIC
Audit
```

Not semantic owners in Launch:

```text
storage/staging/integrity → mechanism
render/view/editor         → mechanism
Search                     → rebuildable projection
async/outbox/jobs          → mechanism
Historical Migration      → cutover capability
backup/restore             → operations/readiness
```

Known Launch+/Future capabilities remain an explicit architectural horizon in `launch-v1-ownership-topology.md`. They receive stable attachment seams but no dormant implementation.

## Approved technical descent

```text
T1 Semantic State & Invariants                         ACTIVE
T2 Governance, Effectivity & Lifecycle Transactions   NOT OPEN
T3 Authorization & Audit Enforcement                  NOT OPEN
T4 Exact Content, Storage Integrity & Restore         NOT OPEN
T5 Durable Async, Search & External Effects           NOT OPEN
T6 Canonical API / Frontend Journeys                  NOT OPEN
T7 Historical Migration & Cutover                     NOT OPEN
```

The old `R10-A → B1→B6 → C→D→E→F` stage order is superseded as active routing.

## Prior redesign / technical evidence

- [cohesive-platform-redesign.md](cohesive-platform-redesign.md) — **SUPERSEDED FOR ACTIVE ROUTING / prior-design compatibility page.** Full former narrative lives in Git history.
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md` — frozen prior R3–R9.5 historical/product-domain evidence, subordinate to the Product Contract where different.
- prior R10 B1–B6 candidates/reviews — evidence only where their invariants survive re-derivation.
- old R10-C Artifact physical-integrity candidate — paused historical evidence; do not repair/promote.
- [launch-v1-scope-rebaseline.md](launch-v1-scope-rebaseline.md) — narrow operator-approved Records-Governance defer overlay; subordinate to Product Contract/GCR/topology.

## Stable cross-cutting references

These remain valid only where they do not conflict with current Product Contract, GCR, ownership or rebaselined technical authority:

- [backend-api-structure.md](backend-api-structure.md) — backend route/OpenAPI structure.
- [api-contract.md](api-contract.md) — contract-first policy and generated-surface rules.
- [api-design-system.md](api-design-system.md) — API behavior standards.
- [frontend-structure.md](frontend-structure.md) — frontend structure and boundary rules.
- [tenant-context.md](tenant-context.md) — current/runtime trust-boundary evidence; target tenancy mechanics are re-derived under the single-company authority and known pooled-tenancy horizon.
- [trusted-proxy.md](trusted-proxy.md) — network trust boundary.
- [rate-limiting.md](rate-limiting.md) — request throttling architecture.
- [tech-stack.md](tech-stack.md), [deployment.md](deployment.md) — supporting runtime/operations references.

## Legacy/current-state architecture references

The following describe prior targets or current implementation and **must not be used as current Launch product/domain authority**:

- [backend-target-architecture.md](backend-target-architecture.md)
- [backend-blueprint.md](backend-blueprint.md)
- [system-overview.md](system-overview.md)
- [data-model.md](data-model.md)
- [../backend/index.md](../backend/index.md)

Runtime/schema/OpenAPI win only for “what runs today,” not “what target are we designing.”

When pages disagree on Launch capability, semantic ownership or stage routing, use:

```text
Product Contract
→ adjudicated Whole-Product GCR
→ Launch ownership topology
→ rebaselined R10 technical authority
→ current handoff
```
