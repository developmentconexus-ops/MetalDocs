# Phase 0 — Audit · novo-template-identidade

User confirmed 2026-05-09.

| Element | Maps to | Decision | Reason |
|---|---|---|---|
| Profile recap card | `state.profileCode` (Step 1), only `scopeType==='profile'` | Keep | Confirms scope. `generic` → render "Genérico" badge instead. |
| "Trocar" button | `dispatch GO_TO_STEP step:1` | Keep | Back-nav. |
| Versão inicial v1.0 (read-only) | static literal | Keep static | Backend assigns; UI just shows. |
| Nome do template (input, required) | POST body `name` | Keep | Primary identifier. |
| Descrição (textarea) | POST body `description` | Keep | Optional. |
| Code preview "TPL-POP-009" | mock client-side | **Defer** (mocked) | No `GET next-code/<profile>` endpoint. Backlog: `wiki/backlog/novo-template-wizard.md`. Mock format: `TPL-{PROFILE_CODE}-XXX`, `TPL-GEN-XXX` for generic. |
| Tags chip input | — | **Cut** | Backend has no tags field; user OK to drop. |
| Backend `key` field | NOT in design | **Defer** (auto-stub) | Auto-generate slug from name on submit (Step 5). Backlog row for proper UX. |

Cut block: Tags input.
Defer blocks: code preview (mocked with TODO), key generation (Step 5 concern).
