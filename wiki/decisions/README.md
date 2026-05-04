# Decisions (ADRs)

> **Last verified:** 2026-05-03
> **Scope:** Architecture Decision Records. Numbered chronologically. Each captures context + decision + trade-offs.

- [0001-eigenpal-adoption.md](0001-eigenpal-adoption.md) — adopt eigenpal as the editor
- [0002-zone-purge.md](0002-zone-purge.md) — remove editable zones (executed 2026-04-25)
- [0003-token-syntax-migration.md](0003-token-syntax-migration.md) — migrate `{{uuid}}` → `{name}` (proposed)
- [0007-two-tier-authz.md](0007-two-tier-authz.md) — accept two distinct authz tiers (CapabilityService vs authz.Require); no schema migration needed (accepted 2026-05-03)
- [0008-placeholder-fixed-catalog.md](0008-placeholder-fixed-catalog.md) — replace user-fill placeholders with fixed 7-token computed catalog (2026-04-26)
- [0008-soft-archive-via-timestamp.md](0008-soft-archive-via-timestamp.md) — archive via `archived_at` timestamp; status field unchanged; `finalized_at` dropped in favour of `v_document_finalized` view (accepted 2026-05-03)
