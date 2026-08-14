# Module: taxonomy — LEGACY BOUNDARY REFERENCE

> **Status:** LEGACY / current taxonomy boundary being dismantled
> **Marked:** 2026-08-14

The current taxonomy module combines concepts that the target no longer accepts as one bounded context.

Approved direction:

- `Area` → **Organization**;
- `DocumentProfile` → **DocumentType** direction;
- `DocumentFamily` → retain only if independent classification/navigation value survives;
- `GovernanceClass` → retain only if it drives independent business policy beyond duplicating ApprovalPolicy.

The current `editable_by_role`, route-shape derivation and other cross-domain configuration are not target authority.

Read:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

Use current code/schema only to understand migration impact. Use Git history for the former detailed taxonomy architecture if needed.
