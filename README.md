# MetalDocs

MetalDocs is being rebuilt from its ratified product and architecture authorities.

The previous implementation was intentionally removed from the live tree because it represented a superseded product and technical model. Git provenance remains reachable outside the live documentation tree.

## Start here

- [Documentation index](docs/index.md)
- [Current status](docs/status.md)
- [Repository reset decision](docs/decisions/repository-reset.md)
- [Paused T8-E checkpoint](docs/reference/t8e-checkpoint.md)

## Current posture

```text
Product / semantic architecture through T8-D   RATIFIED
Repository clean-slate reset                    ACTIVE REVIEW GATE
T8-E executable API contract                    PAUSED AT ACCEPTED CHECKPOINT
Implementation                                 BLOCKED
Legacy implementation                          ABSENT FROM RESET TREE
```

After the reset gate is ratified, temporary `docs/work/**` material is deleted and T8-E resumes from the durable checkpoint.

Do not reconstruct the previous implementation from Git history unless a current ratified authority explicitly proves a reusable mechanism.