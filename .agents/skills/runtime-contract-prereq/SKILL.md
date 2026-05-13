---
name: runtime-contract-prereq
description: Use when a MetalDocs task discovers startup drift, migration drift, auth/session failure, route mismatch, or runtime/OpenAPI/generated/frontend-wrapper mismatch that must be repaired before feature work continues.
---

# Runtime + Contract Prerequisite Audit

Read and follow `.claude/skills/runtime-contract-prereq/SKILL.md`.

## Required sources

- `.claude/skills/runtime-contract-prereq/SKILL.md`
- `wiki/architecture/backend-api-structure.md`
- `wiki/architecture/api-contract.md`
- `wiki/architecture/frontend-structure.md` when frontend wrappers are involved

This bridge exists so Codex sessions can discover the canonical prerequisite workflow and source of truth.

Stop if feature work is trying to continue through a failing prerequisite boundary.
Stop if canonical guidance is missing or conflicts with the required sources.
