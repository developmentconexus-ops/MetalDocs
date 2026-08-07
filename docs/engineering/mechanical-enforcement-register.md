# Mechanical-Enforcement Register

> **Opened:** 2026-08-06 · **Owner:** whoever is running the current standardization program
> **Status:** standing register — never "closed"; entries close individually when their firing mechanism lands.

## What this register is for

MetalDocs is mid-campaign on one idea: **replace hand-maintained agreement with mechanical
derivation.** The errors module was standardized. The HTTP/API surface became generated from the
spec (ADR 0090/0091). Approval routes were standardized and made config-first. AuthZ is next.

The meta-defect behind all of them was named in the 2026-07-03 final architecture review:
**hand-synced enumerations**. Two or more places that must agree, with nothing forcing them to.
Appendix B cause 2 of the defect-class catalog is the same finding from the defect side.

This register is where that observation gets *durable* instead of being re-derived each program.
Anything spotted in passing — during design, review, debugging, anything — that is currently kept
correct by discipline rather than by a mechanism goes here, whether or not the current task touches
it. Recording is cheap; rediscovering is not.

**The operator's standing rule this register serves** (2026-08-06, restated from CLAUDE.md's
Global Maximum section): *never prefer a solution because it is what is already implemented.* If the
implemented thing is genuinely best, it stays and the entry says so with the reason. Otherwise the
global-maximum structure gets named here, even when it is out of the current boundary.

## How to read an entry

Every entry must carry a **firing mechanism** — the concrete thing that makes drift a red build, a
boot fatal, or an unrepresentable state. An entry without one is a wish, not a plan.

Preference order for firing mechanisms, strongest first. This ordering is doctrine
([[no-fallback-principle]], "prefer unrepresentable over guarded"), not taste:

1. **Unrepresentable** — the drift cannot be expressed. One declaration, everything else derived
   from it (FK instead of repeated CHECK; one function with a parameter instead of two functions).
2. **Boot fatal** — the process refuses to start. `assertSurface` (ADR 0091) is the template.
3. **Red build** — a lint or generator-diff test fails in CI.
4. **Runtime assertion** — fails closed at the point of use.
5. *(not acceptable alone)* A doc, a comment, or a checklist.

| Field | Meaning |
|---|---|
| **Surfaces** | The places that must currently agree, with `file:line`. |
| **Kept correct by** | What holds it together *today* — usually "discipline". |
| **Firing mechanism** | What should hold it, per the order above. |
| **Verified** | Whether the claim was read first-hand, and when. |

---

## Open entries

### ME-01 — Role vocabulary declared on six surfaces

**Surfaces**
| Surface | Roles |
|---|---|
| `internal/platform/iamtypes/role.go:39` `validRoles` | 8 |
| `internal/platform/iamtypes/role.go:75` `areaRoles` | 7 |
| `api/openapi/v1/openapi.yaml:4836` `UserRole` | 8 |
| `api/openapi/v1/openapi.yaml:4845` `AreaRole` | 7 |
| `db/baseline/0001_current_schema.sql:1662` `user_process_areas` CHECK | 7 |
| `db/baseline/0001_current_schema.sql:1276` `iam_user_roles` CHECK | **5** |
| `db/baseline/0001_current_schema.sql:1245` `iam_group_roles.role` | **no constraint** |

**Kept correct by** discipline, and it already failed. Four surfaces agree in two consistent pairs.
The `iam_user_roles` CHECK of 5 is mirrored by no Go set and no OpenAPI enum, and the three roles it
omits — `area_admin`, `qms_admin`, `signer` — are the three the rest of the system treats as
first-class. That unmirrored table is the one tier-1 reads.

**Firing mechanism** — level 1. A `role_catalog` table referenced by FK; the Go sets and OpenAPI
enums generated from it. Adding a role becomes an INSERT, not a seven-place edit.

**Verified** 2026-08-06, first-hand. **Owner:** the authz grant-unification program (in design).

---

### ME-02 — REQ-AUTHZ-5's CI binding covers capabilities, not roles

REQ-AUTHZ-5 (`wiki/architecture/backend-target-architecture.md:162-169`) requires the declaration
surfaces be CI-bound so drift is a red build. That guard exists and works — **for capabilities.**
No guard compares the seven role surfaces in ME-01, which is why ME-01 survived a requirement
written to catch exactly its shape.

**This is the most instructive entry in the register.** A firing mechanism that covers the wrong
noun reads, in every audit, exactly like one that works. When adding a guard, write down what it
does *not* cover, in the guard.

**Firing mechanism** — level 3, extending the existing surface-coherence lint to the role catalog.
Subsumed by ME-01's level-1 fix if that lands first; keep this entry open until one of them does.

**Verified** 2026-08-06. **Owner:** authz grant-unification program.

---

### ME-03 — Document list visibility is a hand-written binary, not a permission query

`internal/modules/documents/delivery/http/handler.go:435-439` is the entire visibility model for
listing documents:

```go
if !isAdmin && callerUserID != "" {
    opts.CreatedBy = callerUserID
}
```

Either the caller is `system_admin` and sees everything, or sees only what they created. An area
`viewer` holding a grant in `user_process_areas` does **not** see that area's documents.

This is the disjoint-grant-tables defect surfacing as a product bug rather than a permission bug:
the list can only ask the question tier-1 can answer, and tier-1 cannot see the area model. The port
that answers it correctly already exists and is unused here — `AreaCapabilityReader`
(`internal/modules/iam/domain/area_capability_reader_port.go:19`) returns
`(tenant_wide bool, areas []string)` for a capability.

**Firing mechanism** — level 1: delete `IsSystemAdmin` entirely so "sees everything" is not
expressible as a boolean, only as "holds `document.read` at tenant scope". Ratified 2026-08-06.

**Verified** 2026-08-06, first-hand. **Owner:** authz grant-unification program.

---

### ME-04 — Request params → domain filter fields have no completeness guard

The search `Query` struct carried four fields (`Subject`, `BusinessUnit`, `Classification`, `Tag`)
that the handler populated and no reader ever consumed. A caller passing `classification=CONFIDENTIAL`
got unfiltered results and **no error** — a silent no-op on a field whose name implies access
control. Deleted 2026-08-06 (`6425b21a`), but the class is open: nothing prevents the next one.

Two independent gaps, and they need different mechanisms:
- a declared parameter the query layer never reads (silent no-op);
- a domain filter field nothing populates (dead weight that reads as capability).

**Firing mechanism** — level 3. A lint pairing declared query parameters against the fields the SQL
builder actually consumes. The generated-params direction (level 1) is the better answer if the
codegen can be pushed that far — worth a spike before settling for the lint.

**Verified** 2026-08-06, first-hand. **Owner:** unassigned.

---

### ME-05 — Confirm the tripwire CASE arms are generated, not hand-written

`public.enforce_capability_asserted()` (`db/baseline/0001_current_schema.sql:304-515`) is a large
per-table `CASE` mapping tables to required capabilities. GMR M2 recorded that tripwire arms became
generated from the Go capability registry with two blocking drift/parity lints — so this is very
likely already mechanized and the baseline is generator *output*.

**Not verified.** Recorded as a question, not a finding. If it is generated, close this entry and
note it in ME-02 as a positive example of a correctly-scoped guard. If any arm is hand-maintained,
it becomes a level-1/level-3 entry in its own right.

**Owner:** unassigned. Cheap to settle — find the generator or fail to.

---

### ME-06 — Watch for per-document sharing; it is the trigger that retires RBAC-with-scope

Not a defect. A recorded **decision trigger**, so the choice gets revisited on evidence instead of
inertia.

The authz model (capability bundles + scope on the binding) is NIST-RBAC-shaped and was chosen over
a policy engine (Cedar, OPA) and over Zanzibar/OpenFGA/SpiceDB on merit: MetalDocs' policy has no
attribute conditions and no relationship traversal, so those engines buy unused expressiveness while
charging a second runtime, a policy language, and — decisively — a **synchronization problem between
the grants in Postgres and the engine's own data**, which is the Class 35 defect re-created at a new
boundary.

**The trigger that flips this:** a requirement for per-document or per-object sharing ("share this
SOP with these three people"). RBAC-with-scope cannot express it, and the workaround — a role per
document — is the direct road back to the defect this program removes. On that day, a relation-tuple
engine becomes the global maximum and this entry becomes a program.

**Verified** as a decision, 2026-08-06. **Owner:** whoever gets that requirement first.

---

## Positive template — what a landed firing mechanism looks like

Recorded because "what good looks like" is easier to copy than to describe, and because the register
should not read as a list of failures only.

**`assertSurface`** (`apps/api/cmd/metaldocs-api/surface.go:25-30`, ADR 0091) — four boot checks
proving the mounted HTTP surface is exactly the declared one, aggregated into one error so a failure
is fixable in one restart, ending in `os.Exit(1)`. Level 2. Two properties worth copying:

1. **Ownership is evaluated per-publisher, not against the global union.** A global-set check would
   let every publisher claim a distinct tag while silently mounting each other's routes, and all the
   other checks would still pass. The mechanism was designed against the way it could be fooled, not
   only against the way it should work.
2. **Its ADR states what it does not claim** — that the surface is *exactly the declared one*, not
   that any handler is correct. A guard whose scope is written down cannot be mistaken for a
   stronger one later. ME-02 is what happens when that is left implicit.
