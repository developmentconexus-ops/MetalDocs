# Module: approval — LEGACY CURRENT-STATE REFERENCE

> **Status:** LEGACY / target boundary under Cohesive Platform Redesign
> **Marked:** 2026-08-14

The current `internal/modules/approval` implementation still runs until migration, but its historical living-architecture page is **not target authority**.

Target approval semantics are defined in:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

The approved V1 direction is a small, versioned, sequential governed-document Approval engine with ordered human steps, named/group/role-in-area participants, ANY/ALL completion, `accept` / `return_for_changes`, audited reassignment and optional reauthentication. Generic BPM, old StageKind split, drift policies, M-of-N and the current broad delegation model are not V1 requirements.

Do not repair or extend the old Approval architecture by inertia. Use Git history for detailed legacy implementation archaeology when needed for the later migration map.
