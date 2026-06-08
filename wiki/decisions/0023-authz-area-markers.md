# ADR 0023 — Honest authz-area markers (retire the negative escape hatch)

> **Status:** Accepted 2026-06-08 (api-contract-hardening Phase F · FD-1 = Option A). Lint + spec + docs only — **no runtime / enforcement change**.
> **Last verified:** 2026-06-08
> **Scope:** How an OpenAPI operation declares its tier-2 area-authorization posture in the spec, and what the `AUTHZ-DRIFT` lint enforces. Extends ADR [`0007`](0007-two-tier-authz.md) (two-tier model + codegen-rejection) and ADR [`0022`](0022-authz-capability-coherence.md) (capability coherence; F6 dead-scaffolding finding).
> **Out of scope:** Tier-1 capability routing (`permissions.go`); the runtime enforcement mechanism (tx-layer `authz.Require` + Postgres tripwire) — unchanged here; any OpenAPI **shape** change.
> **Key files:**
> - `api/openapi/v1/openapi.yaml` — the 16 area ops now carry positive markers
> - `scripts/api-lint/spec_rules.go` — `checkAuthz` / `checkAuthzAreaShape` (`AUTHZ-DRIFT`)
> - `scripts/api-lint/code_rules.go` — `tripwire-pairing` (the `authz-call-present` rule was deleted here)
> - `internal/modules/iam/authz/authz.go:51` — tier-2 `Require(ctx, tx, cap, areaCode)`
> - `internal/modules/documents/application/document_area.go` — `LoadDocumentAreaCode` (the DB-derived area source)
> - `db/migrations/0142b_role_capabilities_v2_enforce.sql` — the Postgres tripwire (real enforcer)

## Context

ADR 0007's codegen-rejection amendment shipped an annotation-driven lint, `authz-call-present`: every op carrying `x-authz-area` must have a matching `authz.Require(req.Body.AreaCode | req.<Op>Params.X)` call in the operationId-named handler. The rule was **dormant from day one and stayed at 0 hits across all of ADR 0022's phases** — flagged explicitly as "dead scaffolding" in the ADR 0022 post-migration audit (finding F6).

The reason is architectural, not a bug: MetalDocs does **not** read the enforced area from a request field. It loads the area from the **DB row inside the transaction** (`LoadDocumentAreaCode`, the controlled-document `FOR UPDATE` row, etc.) — un-spoofable, per ADR 0007's tx-coupling — and proves the capability was checked by setting the `metaldocs.asserted_caps` GUC that the Postgres tripwire reads on every mutation. The `req.Body.AreaCode` handler shape the rule scanned for never existed in any module.

To keep the dormant rule green, all 16 area ops carried a **negative `x-authz-skip-area: true` + `x-authz-skip-reason`** marker. This was actively misleading: `skip-area` reads as "this op has no area check," when in fact every one of them is area-enforced — more strictly than a handler-body check would be. The spec lied about its own security posture to satisfy a lint that modelled the wrong architecture.

## Decision

Delete the dead rule and make the markers tell the truth. (FD-1 = Option A.)

1. **Delete `authz-call-present`** (`checkAuthzCallPresent` + its exclusive helpers `inspectAuthzCalls` / `expectedAuthzCall` / `pascalCase*` / `indexModuleFuncs` / `indexedFunc`, and the `handlers_*` testdata). The standing static guarantees are unchanged and sufficient:
   - `tripwire-pairing` — every mutating repository SQL pairs with an `authz.Require` call.
   - `authz-area-scope-binding` (ADR 0022 Phase 7 AST guard) — bans `authz.Require(<areaGradeCap>, "tenant")`, binding each typed area-grade cap to a real area at its call site.
   - the Postgres tripwire trigger — rejects any mutation whose required cap is not asserted in the tx.

2. **Replace the 16 negative markers with positive provenance markers** that say *where the enforced area comes from*:

   ```yaml
   # DB-derived (the common case): area loaded from the row inside the tx
   x-authz-area:
     source: tx
     derived_from: "documents.process_area_code_snapshot (fallback controlled_documents.process_area_code)"
     note: "..."   # optional: where/how the tier-2 check runs

   # Request-target: the area is the legitimate action target (e.g. "grant in area X")
   x-authz-area:
     source: body          # or: path
     field: area_code

   # Genuinely area-less (tenant-global)
   x-authz-area-none: "Templates are tenant-global; tier-1 capability-gated, no per-area scope."
   ```

   `source: tx` is a first-class value alongside the pre-existing `body` / `path`. The 16 ops map to: 12 `source: tx` (10 document lifecycle/edit + 2 controlled-document changeStatus), 3 request-target (`grantAreaMembership` body, `revokeAreaMembership` path, `atomicCreateControlledDocument` body), 1 `x-authz-area-none` (`archiveTemplate`).

3. **`AUTHZ-DRIFT` validates marker shape** (`checkAuthzAreaShape`): `source: tx` requires a non-empty `derived_from`; `source: body|path` requires a non-empty `field`; any other `source` is drift. Every state-transition POST must carry exactly one of `x-authz-area` / `x-authz-area-none` / `x-authz-custom`. The rule stays **blocking**.

## Consequences

**Wins.**
- The spec stops lying: every area op now declares its real provenance instead of a false "skip."
- One fewer dead lint to mislead future readers into thinking area ops are statically call-graph-verified (they are DB+tripwire-verified, which is stronger).
- `AUTHZ-DRIFT` now does positive shape validation, so a marker that omits its provenance is a red build.

**Costs / non-changes.**
- No runtime, handler, enforcement, or DB change — purely lint + spec + docs. Regenerating the 6 oapi-codegen packages re-rolls the embedded-spec blob (the markers are `x-` extensions) but changes no generated types or routes; the 4 `BaseURL: "/api/v1"` generated-router mounts stay intact.
- `x-authz-custom: true` is retained as the escape hatch for a future op that computes area by a bespoke runtime path.

## Alternatives considered

| Option | Verdict | Reason |
|---|---|---|
| **B — build an interprocedural call-graph lint** to actually verify the tx-layer `authz.Require` per op | Rejected | Re-proves what the Postgres tripwire already enforces at the data layer — the exact "duplicate the DB guarantee in static code" duplication ADR 0007's codegen-rejection turned down. High complexity, zero added safety. |
| Keep `x-authz-skip-area` but document it better | Rejected | The marker is structurally dishonest (negative name for a positively-enforced op); no amount of `reason` prose fixes "skip" meaning "enforced." |
| Move authz into handlers so `authz-call-present` could fire | Rejected | ADR 0007 already rejected this — pre-tx handlers cannot supply the tx, and request-supplied area is spoofable. |

## References
- ADR [`0007-two-tier-authz.md`](0007-two-tier-authz.md) — two-tier model + codegen-rejection amendment (now notes this deletion)
- ADR [`0022-authz-capability-coherence.md`](0022-authz-capability-coherence.md) — F6 dead-scaffolding finding + the FD-1 closing amendment
- `wiki/architecture/api-design-system.md` §6 — two-tier authz convention (marker model)
- `wiki/concepts/authz-tiers.md` — tier-1/tier-2 reference
- `wiki/backlog/api-contract-hardening.md` — Phase F · FD-1
