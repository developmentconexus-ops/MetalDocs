# Modules

> **Last verified:** 2026-05-10
> **Scope:** Per-module deep dives. One file per backend module / frontend feature area.

- [editor-ui-eigenpal.md](editor-ui-eigenpal.md) — eigenpal integration layer (Last verified: 2026-05-06)
- [editor-chrome.md](editor-chrome.md) — shared toolbar overlay + eigenpal CSS overrides for eigenpal-based pages; slot API, `VersionBadge`, `AutosaveStatus`; consumed by templates + documents (Last verified: 2026-05-06)
- [templates-v2.md](templates-v2.md) — template authoring, versioning, approval; List screen + creation wizard Steps 1–5; reducer SSOT (`selectMaxReachableStep` + auto-clamp); mocked DOCX flow + mocked permissions roles/areas/counts; Step 5 visual-only submit (Last verified: 2026-05-10)
- [frontend-primitives.md](frontend-primitives.md) — generic `components/ui/` primitives: `SelectableCard` (forwardRef) + `useRovingRadioGroup` hook (Last verified: 2026-05-10)
- [documents.md](documents.md) — **Arc42 + C4** — document instance lifecycle; §5 Container view; §6 finalize trace; §8.1 two-tier authz + tripwire; `CreateDocumentTx` port; `internal/modules/documents/`, table `public.documents` (Last verified: 2026-05-10)
- [documents-tech-debt.md](documents-tech-debt.md) — documents tech-debt register (T-001..T-010) (Last verified: 2026-05-10)
- [registry.md](registry.md) — controlled-document catalog, code generation, active-document FULL OUTER JOIN (E10), registry detail page, PublishedDownloadCell (Last verified: 2026-05-04)
- [taxonomy.md](taxonomy.md) — families (global), profiles, areas; CRUD routes, scoping distinction, deactivation guards (Last verified: 2026-05-02)
- [approval.md](approval.md) — routes, signoffs, ISO segregation; known gaps D4/E4/outbox (Last verified: 2026-05-03)
- render-fanout.md — TBD (Last verified: 2026-05-04)
- [iam.md](iam.md) — IAM module (Arc42 + C4): capabilities, roles, area memberships, tier-1 `CapabilityService.CanDo`, tier-2 `authz.Require`, Postgres tripwire, system_admin bypass, group grants (Last verified: 2026-05-10)
- [iam-tech-debt.md](iam-tech-debt.md) — IAM tech-debt register (T-001..T-012) (Last verified: 2026-05-10)
- [auth.md](auth.md) — auth module (Arc42 + C4): session cookie authn, bcrypt, per-account lockout, HMAC-signed opaque tokens, ManagedUser admin ops; 2 Critical / 3 Major / 7 Minor tech-debt items (Last verified: 2026-05-10)
- [auth-tech-debt.md](auth-tech-debt.md) — auth tech-debt register (T-001..T-012) (Last verified: 2026-05-10)
