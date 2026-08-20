# MetalDocs

MetalDocs is being rebuilt from its ratified product and architecture authorities.

The previous implementation was intentionally removed from the live tree because it represented a superseded product and technical model. Git provenance remains reachable outside the live documentation tree.

## Start here

- [Documentation index](docs/index.md)
- [Current status](docs/status.md)
- [Repository reset decision](docs/decisions/repository-reset.md)
- [T8-E checkpoint](docs/reference/t8e-checkpoint.md)

## Current posture

```text
Product / semantic architecture through T8-D   RATIFIED
Repository clean-slate reset                    OPERATOR-RATIFIED / MERGE PENDING
T8-E executable API contract                    ACTIVE / ACCEPTED CHECKPOINT
Implementation                                 BLOCKED
Legacy implementation                          ABSENT FROM RESET TREE
```

After PR #134 merges, continue T8-E from the durable checkpoint in a fresh small Draft PR. Do not reconstruct the previous implementation from Git history unless a current ratified authority explicitly proves a reusable mechanism.