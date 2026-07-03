# Tokens Module Tech Debt

> **Last verified:** 2026-06-28 (SP-2 update: TD-1 closed)
> **Scope:** Bounded defers and active minor issues for `internal/modules/tokens`.
> **Severity rubric:** Critical = blocks correctness guarantee; Major = architectural risk or user-visible data hazard; Minor = cleanup / developer experience.

---

## TD-1: Computed-token / dictionary collision (CLOSED — SP-2)

- **Severity:** ~~major~~ closed
**Status:** Resolved by ADR 0049 (SP-2 creation-time prevention + pinning)
**Opened:** 2026-06-28 (SP-1 design, spec §8)
**Closed:** 2026-06-28 (SP-2 delivery)

### Problem (historical)

The `tokens` module introduced a second class of tokens alongside the computed-token catalog owned by `templates`. The concern was that at SP-2 render time both catalogs might be merged into a single substitution map, causing collision if a dictionary entry and a computed placeholder shared the same `name`.

### Resolution

SP-2 chose creation-time pinning (Arch-A, ADR 0049) — the render-time catalog merge never happens:

1. **Reserved-name guard (D4):** `tokens.application.Service.Create` rejects any dictionary entry name that equals a native/computed resolver key (422 `RESERVED_NAME`). A dictionary entry `REVISION` cannot be created if `REVISION` is a computed-resolver key.
2. **Schema-save defense-in-depth (D5):** `templates.ValidatePlaceholders` rejects a `PHDictionary` reference whose `Name` equals a native/computed key (`ErrPlaceholderReservedName`).
3. **Creation-time pinning (D1):** Dictionary values are resolved off-tx at document creation and pinned as `source='dictionary'` rows. Render receives pre-resolved `value_text` values from `document_placeholder_values` — it does not merge catalogs.

There is no render-time merge map; there is therefore no collision to resolve. The precedence question (SP-5 row in ADR 0048) is also moot.

### Files

- `internal/modules/tokens/application/service.go` — `ReservedNames` guard (D4)
- `internal/modules/templates/application/schema.go` — `PHDictionary` defense-in-depth (D5)
- `wiki/decisions/0049-tenant-dictionary-token-substitution.md` — governing ADR

---

## TD-2: strictjson promotion shim in documents/approval module (CLOSED — ARC-03)

- **Severity:** ~~minor~~ closed
**Status:** Resolved — both the tokens-handler half and the documents/approval shim removal are done
**Opened:** 2026-06-28 (SP-1 Task 1 observation)
**Updated:** 2026-06-28 (SP-1 final review + SP-2 cleanup)
**Closed:** 2026-07-02 (ARC-03, grade-A simplification program)

### Problem (historical)

Task 1 of SP-1 promoted `internal/platform/strictjson` as the canonical strict-JSON decoder for new handlers. As part of that task, a shim was noted in the documents and/or approval module that re-exports or inline-wraps the strict decoder before the platform package existed. The shim can be removed once the owning module migrates its own JSON decode calls to import `internal/platform/strictjson` directly.

**Tokens-handler half — RESOLVED.** SP-1's final review switched both decode sites in `internal/modules/tokens/delivery/http/handler.go` to `strictjson.Decode` (`CreateToken` line 99, `UpdateToken` line 145) and removed the `encoding/json` import. The tokens handler now rejects unknown fields (strict decode), and a guard test asserts this (`test(tokens): assert strict decode rejects unknown fields`). This sub-item is closed.

**Documents/approval shim — RESOLVED (ARC-03).** The shim lived at `internal/modules/documents/approval/http/contracts/strictjson.go` (a `package contracts` file that re-exported `strictjson.Decode`/`ErrContentType`/`ErrBodyTooLarge`/`ErrEmptyBody`). All 9 call sites across the approval HTTP layer (`cancel_handler.go`, `obsolete_handler.go`, `publish_handler.go`, `signoff_handler.go`, `supersede_handler.go`, `submit_handler.go`, `doc_approval_handler.go` ×2, `route_admin_handler.go` ×3) now call `strictjson.Decode` directly, and `errors.go`'s `MapErrorToResponse` now matches `strictjson.ErrContentType`/`ErrBodyTooLarge`/`ErrEmptyBody` directly — decode strictness and problem+json status/code mapping are unchanged (409/415/413/400 mappings preserved verbatim). The shim file was deleted once zero references remained. `contracts_test.go`'s duplicate `TestDecode` suite (identical coverage to `internal/platform/strictjson/strictjson_test.go`) was removed along with its now-unused imports; `errors_test.go` was updated to reference `strictjson.Err*` and its now-unused `contracts` import was dropped.

### Resolution

- ~~Tokens handler: adopt `strictjson.Decode` (reject unknown fields).~~ **Done (SP-1 review).** Both sites use `strictjson.Decode`; `encoding/json` removed; covered by a strict-decode guard test.
- ~~Documents/approval module: remove the shim once that module adopts `internal/platform/strictjson`.~~ **Done (ARC-03, 2026-07-02).** All call sites migrated to the platform package; shim file deleted; zero `contracts.Decode`/`contracts.Err*` references remain (grep-proof).

### Files

- `internal/platform/strictjson/` — canonical platform package (SP-1 Task 1)
- `internal/modules/tokens/delivery/http/handler.go` — uses `strictjson.Decode` at both sites (resolved)
- `internal/modules/documents/approval/http/*.go` — migrated to `strictjson.Decode` / `strictjson.Err*` directly (ARC-03)
- `internal/modules/documents/approval/http/contracts/strictjson.go` — **deleted** (ARC-03); `contracts` package retains only its request-DTO types
