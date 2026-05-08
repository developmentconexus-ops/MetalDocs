# Decisions (ADRs)

> **Last verified:** 2026-05-08
> **Scope:** Architecture Decision Records. Numbered chronologically. Each captures context + decision + trade-offs.

- [0001-eigenpal-adoption.md](0001-eigenpal-adoption.md) — adopt eigenpal as the editor
- [0002-zone-purge.md](0002-zone-purge.md) — remove editable zones (executed 2026-04-25)
- [0003-token-syntax-migration.md](0003-token-syntax-migration.md) — migrate `{{uuid}}` → `{name}` (proposed)
- [0007-two-tier-authz.md](0007-two-tier-authz.md) — accept two distinct authz tiers (CapabilityService vs authz.Require); no schema migration needed (accepted 2026-05-03)
- [0008-placeholder-fixed-catalog.md](0008-placeholder-fixed-catalog.md) — replace user-fill placeholders with fixed 7-token computed catalog (2026-04-26)
- [0009-pdf-dispatch-outbox.md](0009-pdf-dispatch-outbox.md) — transactional outbox for PDF dispatch via `pdf_dispatch_outbox` (accepted 2026-05-03)
- [0010-soft-archive-via-timestamp.md](0010-soft-archive-via-timestamp.md) — archive via `archived_at` timestamp; status field unchanged; `finalized_at` dropped in favour of `v_document_finalized` view (accepted 2026-05-03; renumbered from 0008 due to collision)
- [0011-cd-atomic-create.md](0011-cd-atomic-create.md) — atomic CD create + per-area 3-segment numbering + idempotency-key adoption (2026-05-07)
- [0012-contract-first-api.md](0012-contract-first-api.md) — adopt spec-as-source-of-truth via oapi-codegen; root cause of `documents.name` empty-name bug; migration scope and residual risks (2026-05-08)
