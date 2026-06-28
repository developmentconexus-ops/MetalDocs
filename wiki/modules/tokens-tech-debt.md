# Tokens Module Tech Debt

> **Last verified:** 2026-06-28
> **Scope:** Bounded defers and active minor issues for `internal/modules/tokens`.
> **Severity rubric:** Critical = blocks correctness guarantee; Major = architectural risk or user-visible data hazard; Minor = cleanup / developer experience.

---

## TD-1: Computed-token / dictionary collision (deferred to SP-2)

**Severity:** Major
**Status:** Open — deferred to SP-2
**Opened:** 2026-06-28 (SP-1 design, spec §8)

### Problem

The `tokens` module introduces a second class of tokens alongside the computed-token catalog owned by `templates`. At SP-1 these two catalogs are independent: dictionary entries live in `token_dictionary_entries`; computed tokens are defined in the template's `placeholder_schema`. Neither consumes the other.

At SP-2 the render-fanout module will merge both catalogs into a single substitution map before rendering a document. If a dictionary entry and a computed placeholder share the same `name` (e.g., both define `REVISION`), the render merge has two options — precedence (one silently shadows the other) or rejection (render fails). Neither has been specified.

Concrete scenario: author defines `{REVISION}` as a computed placeholder in template v3; a tenant admin later creates a dictionary entry `REVISION = "R1"`. At SP-2 render time both resolve to the token `{REVISION}` — the substitution map has a collision.

### Impact

Incorrect or non-deterministic document content if the render module's merge strategy is not explicit. This cannot be detected at dictionary-write time (the write path has no visibility into template placeholder schemas).

### Deferred resolution (SP-2 contract)

Before the render-fanout module consumes `DictionaryReader`, the SP-2 design must specify:

1. **Merge precedence rule** — e.g. computed tokens always win over dictionary tokens, or vice versa.
2. **Conflict detection** — whether the render path should detect and reject collisions at render time (422 or 409 to the caller) vs. silently apply precedence.
3. **Dictionary-write guard (optional)** — whether POST/PUT to `/api/v1/tokens` should check for name collision against the tenant's active placeholder schemas and warn (not block).

The `domain.DictionaryReader` interface is already published and stable; no tokens-module change is needed to implement TD-1 at SP-2 — the fix lives entirely in the render-fanout module.

### Files

- `internal/modules/tokens/domain/port.go` — `DictionaryReader` (the published port SP-2 will consume)
- `internal/modules/render-fanout/` — SP-2 merge logic (not yet implemented)
- `wiki/concepts/placeholders.md` — full placeholder concept including computed token catalog

---

## TD-2: strictjson promotion shim in documents/approval module (active, minor)

**Severity:** Minor
**Status:** Active (cleanup opportunity; not blocking)
**Opened:** 2026-06-28 (SP-1 Task 1 observation)

### Problem

Task 1 of SP-1 promoted `internal/platform/strictjson` as the canonical strict-JSON decoder for new handlers. As part of that task, a shim was noted in the documents and/or approval module that re-exports or inline-wraps the strict decoder before the platform package existed. The shim can be removed once the owning module migrates its own JSON decode calls to import `internal/platform/strictjson` directly.

The tokens handler (`delivery/http/handler.go`) does not use `strictjson` — it uses `json.NewDecoder(r.Body).Decode`. This is consistent with the pattern for non-strict decode where unknown fields are acceptable. If the spec requires strict decode for tokens (reject unknown fields), the handler should be updated to use `strictjson.Decode`.

### Resolution

- Documents/approval module: remove the shim once that module adopts `internal/platform/strictjson`.
- Tokens handler: evaluate whether strict decode (reject unknown fields) is required for POST/PUT `/tokens`; if so, switch to `strictjson.Decode` in a follow-up.

### Files

- `internal/platform/strictjson/` — canonical platform package (SP-1 Task 1)
- `internal/modules/tokens/delivery/http/handler.go` — uses `json.NewDecoder` (non-strict)
- `internal/modules/documents/` or `internal/modules/approval/` — shim to be removed
