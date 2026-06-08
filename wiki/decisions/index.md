# Decisions

> **Last verified:** 2026-06-03
> **Scope:** Durable ADRs and consequential technical decisions.

- [0001-eigenpal-adoption.md](0001-eigenpal-adoption.md)
- [0002-zone-purge.md](0002-zone-purge.md)
- [0003-token-syntax-migration.md](0003-token-syntax-migration.md)
- [0007-two-tier-authz.md](0007-two-tier-authz.md)
- [0008-placeholder-fixed-catalog.md](0008-placeholder-fixed-catalog.md)
- [0009-pdf-dispatch-outbox.md](0009-pdf-dispatch-outbox.md)
- [0010-soft-archive-via-timestamp.md](0010-soft-archive-via-timestamp.md)
- [0011-cd-atomic-create.md](0011-cd-atomic-create.md)
- [0012-contract-first-api.md](0012-contract-first-api.md)
- [0013-template-revision-labels.md](0013-template-revision-labels.md) — **Proposed** (awaiting ratification)
- [0017-signoff-idempotency-fingerprint.md](0017-signoff-idempotency-fingerprint.md) — signoff idempotency fingerprint = client-stable inputs only (F-002 correction)
- [0018-approval-route-lifecycle.md](0018-approval-route-lifecycle.md) — approval route lifecycle: terminal-on-deactivate state machine, version OCC, in-use guard, capability pin, reason audit
- [0019-cap-audit-read-and-session-manage.md](0019-cap-audit-read-and-session-manage.md) — tier-1 caps for audit read + session manage (PR-2)
- [0020-admin-center-six-tab-ia.md](0020-admin-center-six-tab-ia.md) — **Accepted** — Admin Center 6-tab IA, shipped at PR-12
- [0021-tenant-vs-platform-admin-separation.md](0021-tenant-vs-platform-admin-separation.md) — **Accepted** — tenant admin vs. platform admin scope split, shipped at PR-12
- [0023-authz-area-markers.md](0023-authz-area-markers.md) — **Accepted** — honest positive authz-area markers (`source: tx` derived_from / `x-authz-area-none`) replace the negative `x-authz-skip-area`; dormant `authz-call-present` lint deleted (Phase F · FD-1)
- [0024-openapi-single-base-path.md](0024-openapi-single-base-path.md) — **Accepted** — AD-1: one `servers.url: /api/v1` + relative path keys; PATH-BASE-PREFIX gate kills the double-prefix bug class
- [0025-error-envelope-rfc9457.md](0025-error-envelope-rfc9457.md) — **Accepted** — AD-2: RFC 9457 Problem is the only error shape; ApiErrorEnvelope retired; ENVELOPE-DRIFT blocking, zero exemptions
- [0026-unified-authz-enforcement.md](0026-unified-authz-enforcement.md) — **Accepted** — AD-3: unified capability+area+grants is the only per-resource authz model; dead ABAC AccessPolicy path removed (extends ADR 0022)
- [2026-06-03-audit-events-cursor-shape.md](2026-06-03-audit-events-cursor-shape.md) — **Closed 2026-06-08** — `/audit/events` runtime reconciled to the nested `page.{next_cursor,has_more}` CursorPage shape (Phase F re-audit); FE dual-shape adapter removed

Legacy ADR material in `docs/adr/` remains historical/reference content until reconciled deliberately.
