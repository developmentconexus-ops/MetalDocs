# ADR-0003: Token Syntax Migration

> **Last verified:** 2026-05-01
> **Status:** Stub. Expand with current migration state + remaining work when prioritized.
> **Date:** TBD
> **Decision status:** In progress — see [decisions/0008-placeholder-fixed-catalog.md](0008-placeholder-fixed-catalog.md) which superseded most of the original motivation.

## Context

Original placeholder format: `{{uuid}}` — generated UUIDs embedded in templates, mapped to user-fillable values via a separate schema. This had several problems:

- Author cognitive load: UUIDs are unreadable; authors couldn't edit templates without consulting the schema.
- Authoring tool coupling: eigenpal-native format used readable `{name}` placeholders; we had to translate.
- Audit traceability: UUIDs in frozen DOCX gave auditors no signal about what was substituted.

## Decision

Migrate from `{{uuid}}` to `{name}` syntax — the readable form eigenpal uses natively.

ADR-0008 then took this further by replacing the entire user-fill mechanism with the fixed 7-token catalog: `{doc_code}`, `{doc_title}`, `{revision_number}`, `{author}`, `{effective_date}`, `{approvers}`, `{controlled_by_area}`.

## Consequences

- Templates are now readable by humans without consulting a separate schema.
- `concepts/token-syntax.md` documents the `{name}` vs `{{uuid}}` distinction for legacy data.
- Old `{{uuid}}` placeholders in legacy templates need migration (status TBD).

## See also

- [concepts/token-syntax.md](../concepts/token-syntax.md)
- [concepts/placeholders.md](../concepts/placeholders.md)
- [decisions/0008-placeholder-fixed-catalog.md](0008-placeholder-fixed-catalog.md)
