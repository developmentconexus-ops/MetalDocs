# Tech Debt Register — <Module>

> Companion to `wiki/modules/<module>.md`. Lists known gaps, smells, and missing-ADR items. **Debt only — no fix prescriptions.** Fixes belong in `wiki/backlog/<module>-refactor.md`.

**Last verified:** YYYY-MM-DD

## Severity scale

The category names are useful only when paired with concrete triggers — abstract words ("important", "impactful") drift in the composer's head and produce inconsistent ratings across modules. Use the trigger list. When in doubt and the bug is on a regulated path, escalate one level.

### Critical — at least one trigger fires
- Authn/authz bypass: a code path lets a request mutate or read without the capability the spec requires.
- Regulated audit-trail gap: a mutation on an ISO 9001 / QMS / regulated path is not written to the audit sink.
- Multi-tenant data leak: a query path can return rows from a different tenant.
- Data-loss path: a code path can drop / overwrite / silently truncate user data.
- Contract violation that downstream consumers rely on (e.g. error shape, idempotency replay window).
- Schema/version drift the boot check is supposed to catch but does not.

### Major — at least one trigger fires
- Defense-in-depth gap: only one layer protects a mutation that the spec calls for multiple layers on.
- Governance / observability sink wired to `nil` on a regulated path.
- Duplicated write surfaces with different semantics for the same use case (only one is live, but both compile).
- Documented contract not followed by this module yet (e.g. RFC 9457 envelope on a v1 route) — measurable user impact via tooling that depends on the contract.
- Cross-module dependency that blocks another module's clean refactor.

### Minor — code-smell / latent / docs
- Symbol naming collision across packages (consumers must qualify imports).
- Missing Go doc comments on exported symbols.
- Latent debt: the surface for the bug exists in code but no caller hits it today.
- Bidirectional dependency that is non-circular today but would be hard to detangle.
- Missing standalone ADR for a rule that is already enforced by code + tests.

### How to rate when triggers overlap
Pick the highest matching tier. If T-005 is both "regulated audit-trail gap" (Critical trigger) and "measurable user impact" (Major trigger), it is Critical. Justify the call in the row's `Observation` field so a future reviewer can audit the rating without re-deriving it.

## Items

### T-001 · <one-line title>
- **Severity:** critical | major | minor
- **Surface:** `internal/modules/<m>/<file>.go:LL` (file:line anchor)
- **Observation:** what is wrong, as a fact. No "should".
- **Evidence:** link to data-flow artifact, surface-scan row, or persistence row that exposed it
- **Linked backlog row:** `backlog/<module>-refactor.md#R-NNN` or `none yet`
- **Linked ADR:** `wiki/decisions/<adr>.md` or `missing-ADR`

### T-002 · ...

---

## Coverage stats (computed at compose time)

- Public symbols undocumented: N / M
- Operations missing C4 placement: N / M
- Cross-deps missing in §5/§8: N / M
- State transitions missing in §6: N / M
- Decisions without ADR link: N / M
