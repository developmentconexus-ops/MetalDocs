# ADR 0066 — Optimistic-concurrency transport: documents use If-Match, templates use body lock_version (intentional split, If-Match is the target)

> **Status:** Accepted
> **Date:** 2026-07-04
> **Program:** global-maximum-remediation · **Milestone:** M4 (versioning kernel correctness), feature F4.3

## Context

MetalDocs has **two** optimistic-concurrency (OCC) mechanisms, both compare-and-swap over a
single monotonic integer version column. They differ only in **transport**:

- **documents / approval:** the OCC precondition rides the HTTP `If-Match: "vN"` request header
  (RFC 7232 conditional request), parsed by `parseIfMatch`
  (`internal/modules/documents/approval/http/handler.go:145-164`) to `documents.revision_version`.
- **templates:** the OCC precondition rides a required JSON body field `expected_lock_version`
  (`internal/modules/templates/delivery/http/routes_schema.go`), mapped to
  `templates_template_version.lock_version`. This is templates' **only** OCC write endpoint
  (`UpdateSchemas`), and it uses the body field self-consistently with its own `lock_version`
  field and `stale_lock_version` error.

M4 F4.3 was chartered to "unify the concurrency idiom." The M4 validation contract §3.4 originally
decided **"unify on If-Match; migrate templates (the minority) to it."** That decision rested on a
premise that verification (F4.3, 2026-07-04) proved **false**:

- **Claim:** templates already largely uses If-Match and `UpdateSchemas` is the lone straggler.
  **Truth:** templates has **zero** If-Match usage anywhere (backend + frontend). Every If-Match
  endpoint in the OpenAPI spec is tagged `[documents]` / `[approval]`.
- **Claim:** an existing system-wide If-Match decision ("DEC-01") makes templates non-conformant.
  **Truth:** the referenced decision is **CON-01** in `wiki/modules/documents.md` — a
  **documents-module-internal** decision (submit canonical over finalize), **not** a system-wide
  OCC transport ADR. No cross-module If-Match standard exists that templates violates.

So migrating templates' `expected_lock_version` → `If-Match` would **not** be "finishing a
convergence to one idiom." It would **create a new cross-module OCC standard** and **import a
documents-local decision into templates** — a genuine architectural decision on its own, riding
inside a milestone whose objective is **versioning kernel correctness**, touching a module (templates)
that M4 has no other reason to change. That is scope creep plus needless regression risk on a
correct, self-consistent module (CLAUDE.md: "stop on architecture contradictions instead of patching
around them"; module-boundary rule).

## Decision

**The two OCC transports are an intentional, documented split, not a defect.**

1. **documents / approval** transport OCC via the `If-Match` header. Unchanged.
2. **templates** transport OCC via the body `expected_lock_version` field. Unchanged.

Both remain **internally self-consistent** within their module; each is a correct CAS over one
integer version column. The DB stays the last-line enforcer of monotonicity in both.

**`If-Match` is the stated long-term target transport** for the codebase. Industry convention for
HTTP OCC is `ETag` + `If-Match` (RFC 7232; Google AIP-154 resource freshness; Zalando REST
guidelines; Stripe/Microsoft patterns): it is HTTP-native, cache/proxy-visible, and keeps the
precondition out of every request schema. documents already sit on this idiomatic side.

**Full unification onto `If-Match` is deferred to its own deliberate change** — a cross-module wire
contract migration (templates OpenAPI + regen + handler + FE consumers + tests), decided and executed
as a first-class piece of work (candidate: M9 governance-hygiene, or a standalone milestone), **not**
smuggled into a correctness milestone. When it happens it supersedes this ADR's split with uniform
`If-Match`.

## Consequences

- **M4 does not touch templates.** F4.3's outcome is this ADR, not a templates code change. Lower
  risk; module boundary respected; M4 stays focused on the versioning kernel.
- The two-idiom state is now **recorded and intentional**, with a named target direction — a client
  reading the contract sees documents=header, templates=body, and this ADR explains why and where it
  is going. The review's dimension-6 "two ways to do OCC" DEBT is **acknowledged and tracked**, not
  silently tolerated.
- A future unifier has a clear charter: migrate templates' `UpdateSchemas` to `If-Match: "vN"` (reuse
  the documents `parseIfMatch` shape or a shared helper), contract-first (openapi + regen, zero
  hand-edits), keep the `lock_version` column and its CAS; then retire this ADR.
- No runtime behavior changes as a result of this ADR (documentation/decision only).

## References

- M4 validation contract: `docs/superpowers/milestones/global-maximum-remediation/milestone-4-versioning-kernel/validation-contract.md` §3 (with the 2026-07-04 erratum recording the false-premise correction)
- F4.3 feature home: `docs/superpowers/milestones/global-maximum-remediation/milestone-4-versioning-kernel/f4.3-concurrency-idiom/`
- documents OCC: `internal/modules/documents/approval/http/handler.go:145-164` (`parseIfMatch`)
- templates OCC: `internal/modules/templates/delivery/http/routes_schema.go` (`expected_lock_version`)
- documents-internal decision cited in the original (false) premise: `wiki/modules/documents.md` CON-01
- Standards basis for the target: RFC 7232 (If-Match / ETag), Google AIP-154, Zalando REST guidelines
