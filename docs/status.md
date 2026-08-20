---
id: repository-status
kind: authority
owner: architecture
summary: Sole current MetalDocs stage, gate, and next-action authority.
---

# Current status

```text
REPOSITORY MODE                              CLEAN-SLATE / ARCHITECTURE-FIRST
REPOSITORY RESET                             ACTIVE REVIEW GATE / PR #134
PRODUCT CONTRACT                             OPERATOR-APPROVED
WHOLE-PRODUCT ALIGNMENT / OWNERSHIP          OPERATOR-APPROVED
T1 → T8-D                                    CLOSED / OPERATOR-RATIFIED
T8-E EXECUTABLE WIRE CONTRACT                PAUSED AT ACCEPTED CHECKPOINT
T8-F → T12                                   NOT OPEN — SEE decisions/stage-program.md
IMPLEMENTATION                               BLOCKED
LEGACY IMPLEMENTATION                        ABSENT FROM RESET TREE
```

## Current gate

Review and ratify the clean-slate repository reset. Temporary review material is under `docs/work/current/` only while this PR remains Draft.

The accepted T8-E work is preserved durably at `docs/reference/t8e-checkpoint.md`; the current 78-operation census is owned by `docs/decisions/api-operation-census.md`.

## Provenance safety

PR #131 and PR #132 are closed provenance, but their source branches remain reachable and MUST NOT be deleted until equivalent immutable archival tags/refs have been created and recorded in `docs/decisions/repository-reset.md`.

Current protected provenance refs:

```text
docs/a8-authz-approval-redesign-ledger @ d8b1c6d31e704e9552a14faa7764c634a29b081d
docs/repository-information-architecture @ b0ebe54cb010e9837a25f7b778f3d9814d283cb8
```

## After reset ratification

Before merge:

```text
Lead adjudication complete
→ operator ratification explicit
→ delete docs/work/current/**
→ switch this status to T8-E ACTIVE
→ required CI green on the final tree
→ squash merge
```

Then T8-E resumes from the durable checkpoint. T8-F may open only after T8-E ratification. The remaining stage ownership and final implementation gate are defined in `docs/decisions/stage-program.md`.

Implementation remains blocked throughout.