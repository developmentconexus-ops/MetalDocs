# T8-E Fable independent review

> **Evidence only — non-authoritative.**
> Candidate authority remains `arch/t8e-wire-contract`; this review branch must never merge.

## Lead handoff

Repository: `developmentconexus-ops/MetalDocs`

Candidate branch: `arch/t8e-wire-contract`

Exact candidate HEAD under review: `ef329534fc9d5df3254d59c3787197fefa8435e6`

Review branch: `review/t8e-fable`

Gate: **T8-E — Executable Wire Contract**

Canonical Method: `developmentconexus-ops/conexus-methodology/METHOD.md` v1.0.0

Repository Standard: `developmentconexus-ops/conexus-methodology/REPOSITORY-STANDARD.md` v1.0.0

### Fresh-actor route

Reconstruct authority independently:

```text
AGENTS.md
→ docs/index.md
→ docs/roadmap.md
→ only the smallest owning authority pack needed for a finding
```

Do not use this handoff as architecture authority. Repository current authority wins.

### Candidate target

Adversarially challenge whether the candidate is the **smallest sustainable executable wire** for the accepted 78-operation `/api/v1` Product/T6 census, with no Writer-visible semantic choices left that belong in T8-E.

Lead evidence already includes executable probes for:

```text
78-row census / unique operationId / method+path
10 durable Idempotency-Key creations
13 ETag read/mutation domains
4 exact-byte resources
OpenAPI 3.0.3 -> oapi-codegen v2.8.0
OpenAPI 3.0.3 -> openapi-typescript 7.13.0
kin-openapi strict-request split
S3 create-only presign + exact Content-Length reference profile
document admission ceilings / adversarial DOCX fixtures
whole-candidate Global Coherence PASS
```

Recent bounded corrections were operator-approved and are now in their owning authorities:

```text
T3      remove unreachable ProviderSubjectBinding-disabled Audit event
T8-D    persist Governance Step label + immutable attempt label_snapshot
T8-D    Audit/transaction + 24h idempotency precision
T4/T5/T8-C/T8-D
        already-PDF RequireOfficialRendition(PDF) reuses exact admitted bytes;
        no renderer/copy/River job unless transformation is required
T8-C/T8-D
        server-side per-session CSRF synchronizer secret is reconstructible for GET /session
```

### Review focus

Try to **falsify**, not confirm, the candidate. In particular attack:

1. **Authority leakage / duplication** — any wire rule that re-owns T1→T8-D semantics, or any upstream semantic requirement the wire fails to encode.
2. **Hidden Writer decisions** — required/optional/nullability, unions, cross-field presence, ordering, status/header/problem mapping, ETag/idempotency/replay behavior, upload/byte semantics, Audit projection.
3. **YAGNI / overengineering** — fields, headers, problems, normalization, limits, jobs, security mechanisms, projection data or generic abstractions with no current consumer.
4. **Security correctness** — session/CSRF, disclosure precedence, idempotency replay authorization/expiry, exact-byte integrity, direct upload admission, Problem information leakage.
5. **Persistence/internal-contract executability** — especially parity after the bounded corrections; no wire property should require state/contracts that accepted owners cannot realize.
6. **Generation/tooling feasibility** — identify any OpenAPI 3.0.3 shape the stated Go/TypeScript boundaries cannot represent without a material semantic compromise.
7. **Structural Inversion** — look for anything inherited from familiar API/platform patterns rather than current MetalDocs Product requirements.
8. **Global Maximum** — propose a materially smaller/stronger alternative only if it preserves all accepted properties; do not add platform capability by preference.

### Output contract

Write your independent review **only in this file** below `## Fable response`.

For each material finding, provide:

```text
ID
severity: MATERIAL | MINOR
claim
owning authority implicated
concrete counterexample/failure
smallest correction
whether it reopens an accepted authority
```

Separate non-material observations from findings. If no material finding survives, say so explicitly and list the strongest attacks attempted.

Do not edit the candidate branch or any other file. Reviewer output is Evidence; the Lead adjudicates every finding.

## Fable response

_Pending independent Fable review._
