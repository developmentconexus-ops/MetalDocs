# F5 — Idempotency map completion (contract-first)

> **Milestone:** M1 · **Findings:** 16, 17 · **Status:** spec approved
> **Approved:** 2026-07-06 — rule pre-agreed in `milestone.md` F5 acceptance: "map a mutating route
> into the platform idempotent set **iff** the OpenAPI spec declares an `Idempotency-Key` header for
> it; else spec+regen first, or a bounded defer with a written trigger." No operator interview: the
> disposition rule is already in the milestone spec.

## Consumer contract

The platform idempotency middleware (`idempotency.Require`, wired via each module's
`idempotentRoutes` map) requires an `Idempotency-Key` **only** for routes whose OpenAPI operation
**declares** that header. The invariant to hold: **the Go idempotent-route set == the set of spec
operations that declare `Idempotency-Key`.** Adding a route to the Go map that the spec does *not*
declare would make the middleware reject a header-less request the contract says is valid — a
contract-first violation, and it would break the existing header-less FE callers.

Disposition of the two findings:

| Finding | Op | Spec declares `Idempotency-Key`? | Disposition |
|---|---|---|---|
| 16 | `POST /documents/{id}/review` (mark-reviewed) | **yes** | Already in `router.go` `idempotentRoutes` (line 36). Verify parity; no change. |
| 17 | `POST /templates/{id}/archive`, `PUT /templates/{id}/approval-config` | **no** | **Bounded defer** (see below). Not added to the Go map. |

**Source of truth:** `api/openapi/v1/openapi.yaml` (which ops carry the `Idempotency-Key` header
parameter); `router.go` + `templates/delivery/http/handler.go` `idempotentRoutes` maps.

## Deviation flagged for HS-1

The governing spec §3 framed finding 17 as an "M1 fix" ("3-line addition to the map"). Executing it
verbatim would violate contract-first (the two ops declare no `Idempotency-Key`) **and** break the
existing header-less approval-config FE caller. Per the pre-agreed F5 rule, #17 is instead a
**bounded contract-first defer**:

- **Why safe now:** `PUT approval-config` is HTTP-idempotent by definition (full-replace PUT); `POST
  archive` is OCC/status-guarded (double-archive is a no-op conflict, not a duplicate side effect).
- **Written trigger:** if the OpenAPI spec later adds an `Idempotency-Key` header to either op, add
  it to `templates` `idempotentRoutes` **and** add a spec↔map parity test in the same change.
- **Owner:** templates module.

This defer is surfaced explicitly at the HS-1 operator gate as a deviation from the §3 "M1 fix" framing.

## Non-goals
- No consolidation of the bespoke in-handler idempotency stores (signoff×2, route-admin×3) — that is
  finding 22, deferred program-wide (README defer register).
- No new `Idempotency-Key` header added to the OpenAPI spec in M1 (that would be the trigger action,
  not M1 scope).

## Validation Gate
- **Full parity sweep:** enumerate every mutating op in the spec that declares `Idempotency-Key`;
  confirm each is in the corresponding Go `idempotentRoutes` map, and each Go map entry has a
  spec-declared header. Result: 25/25 covered, zero orphans.
- Finding 16: `POST /documents/{id}/review` present in `router.go` `idempotentRoutes` — confirmed.
- Finding 17: archive + approval-config **absent** from both spec header decls and Go maps —
  confirmed consistent (defer, not gap).
- `go build ./...` + `go test ./...` green.
