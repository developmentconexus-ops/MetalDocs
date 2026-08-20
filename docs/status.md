---
id: repository-status
kind: authority
owner: architecture
summary: Sole current MetalDocs stage, gate, and next-action authority.
---

# Current status

```text
REPOSITORY MODE                              CLEAN-SLATE / ARCHITECTURE-FIRST
REPOSITORY RESET                             OPERATOR-RATIFIED / PR #134 MERGE PENDING
PRODUCT CONTRACT                             OPERATOR-APPROVED
WHOLE-PRODUCT ALIGNMENT / OWNERSHIP          OPERATOR-APPROVED
T1 → T8-D                                    CLOSED / OPERATOR-RATIFIED
T8-E EXECUTABLE WIRE CONTRACT                ACTIVE / ACCEPTED CHECKPOINT
T8-F → T12                                   NOT OPEN — SEE decisions/stage-program.md
IMPLEMENTATION                               BLOCKED
LEGACY IMPLEMENTATION                        ABSENT FROM RESET TREE
```

## Current gate

Finalize the ratified clean-slate repository reset, then continue T8-E in a fresh small Draft PR from updated `main`.

The accepted T8-E checkpoint is durable at `docs/reference/t8e-checkpoint.md`; the current 78-operation application census is owned by `docs/decisions/api-operation-census.md`.

## T8-E exact next action

```text
create fresh T8-E Draft PR from updated main
→ materialize the accepted checkpoint as the active proposal
→ freeze the remaining 78-operation executable ledger
→ close exact schemas / enums / nullability / success headers
→ close operation-specific RFC 9457 problem-code matrix
→ close filters / deterministic ordering / allowed_actions
→ close request-body limits from evidence
→ prove Go + TypeScript generation/conformance feasibility
→ final subtractive/global-coherence pass
→ one final independent Fable challenge
→ explicit operator ratification
```

T8-F may open only after T8-E ratification.

## Provenance safety

PR #131 and PR #132 are closed provenance, but their source branches remain reachable and MUST NOT be deleted until equivalent immutable archival tags/refs have been created and recorded in `docs/decisions/repository-reset.md`.

Current protected provenance refs:

```text
docs/a8-authz-approval-redesign-ledger @ d8b1c6d31e704e9552a14faa7764c634a29b081d
docs/repository-information-architecture @ b0ebe54cb010e9837a25f7b778f3d9814d283cb8
```

The remaining stage ownership and final implementation gate are defined in `docs/decisions/stage-program.md`.

Implementation remains blocked.