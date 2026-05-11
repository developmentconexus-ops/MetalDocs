# Modules

> **Last verified:** 2026-05-11
> **Scope:** Per-module deep dives. One file per backend module / frontend feature area.

- [editor-ui-eigenpal.md](editor-ui-eigenpal.md) — eigenpal integration layer (Last verified: 2026-05-10)
- [editor-chrome.md](editor-chrome.md) — shared toolbar overlay + eigenpal CSS overrides for eigenpal-based pages; slot API, `VersionBadge`, `AutosaveStatus`; consumed by templates + documents (Last verified: 2026-05-10)
- **[templates_v2.md](templates_v2.md)** — **Arc42 + C4 living doc** — templates_v2 backend: authoring lifecycle, 20 HTTP routes, placeholder catalog enforcement, SoD probing, MinIO presigned upload/download, downstream snapshot contract for `documents` (Last verified: 2026-05-10)
- [templates_v2-tech-debt.md](templates_v2-tech-debt.md) — templates_v2 tech-debt register (14 items: 4 Critical / 6 Major / 4 Minor) (Last verified: 2026-05-10)
- [templates-v2.md](templates-v2.md) — (predecessor — retire pending R-100) frontend-heavy doc: List screen, creation wizard Steps 1–5, `TemplateEditorPage`, `EditorChrome` wiring (Last verified: 2026-05-10)
- [frontend-primitives.md](frontend-primitives.md) — generic `components/ui/` primitives: `SelectableCard` (forwardRef) + `useRovingRadioGroup` hook (Last verified: 2026-05-10)
- [documents.md](documents.md) — **Arc42 + C4** — document instance lifecycle; §5 Container view; §6 finalize trace; §8.1 two-tier authz + tripwire; `CreateDocumentTx` port; `internal/modules/documents/`, table `public.documents` (Last verified: 2026-05-10)
- [documents-tech-debt.md](documents-tech-debt.md) — documents tech-debt register (T-001..T-010) (Last verified: 2026-05-10)
- [registry.md](registry.md) — controlled-document catalog, code generation, active-document FULL OUTER JOIN (E10), registry detail page, PublishedDownloadCell (Last verified: 2026-05-04)
- **[taxonomy.md](taxonomy.md)** — **Arc42 + C4 living doc** — taxonomy: document families (global), profiles, areas; 16 HTTP routes `/api/v2/taxonomy/*`; per-tenant scoping; deactivation guards; 5C/5M/6m debt; companion tech-debt register + refactor backlog (Last verified: 2026-05-11)
- [taxonomy-tech-debt.md](taxonomy-tech-debt.md) — taxonomy tech-debt register (16 items: T-001..T-005 Critical; T-006..T-010 Major; T-011..T-016 Minor) (Last verified: 2026-05-11)
- **[approval.md](approval.md)** — Arc42 + C4 living doc — 16-route sign-off chain; SoD, J1 eligibility, quorum, transactional outbox, 4-layer defense-in-depth authz; §6 Submit/Signoff/Inbox sequence diagrams (Last verified: 2026-05-10)
- [approval-tech-debt.md](approval-tech-debt.md) — approval tech-debt register (12 items: T-001 RFC 9457 Critical; T-002 OpenAPI gap Critical; T-003..T-006 Major; T-007..T-012 Minor) (Last verified: 2026-05-10)
- render-fanout.md — TBD (Last verified: 2026-05-04)
- [iam.md](iam.md) — IAM module (Arc42 + C4): capabilities, roles, area memberships, tier-1 `CapabilityService.CanDo`, tier-2 `authz.Require`, Postgres tripwire, system_admin bypass, group grants (Last verified: 2026-05-10)
- [iam-tech-debt.md](iam-tech-debt.md) — IAM tech-debt register (T-001..T-012) (Last verified: 2026-05-10)
- [auth.md](auth.md) — auth module (Arc42 + C4): session cookie authn, bcrypt, per-account lockout, HMAC-signed opaque tokens, ManagedUser admin ops; 2 Critical / 3 Major / 7 Minor tech-debt items (Last verified: 2026-05-10)
- [auth-tech-debt.md](auth-tech-debt.md) — auth tech-debt register (T-001..T-012) (Last verified: 2026-05-10)
