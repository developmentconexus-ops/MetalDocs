# ADR 0065 — Version references are nested value objects in wire contracts

Status: Accepted (2026-07-03)

## Context

TemplateDTO carried five parallel scalars for two version pointers:
`latest_version` (int) + `latest_revision_number`, and `published_version_id` +
`published_version_number` + `current_revision_number` (all three nullable,
coupled: null together or not at all — one LEFT JOIN produces them). The
coupling was not expressible in OpenAPI; each consumer had to know it. The
2026-07-03 HIGH governance bug (unpublished template selectable in the
controlled-document wizard, fixed 9f86828b) was this shape's first casualty:
consumers gated on one of three independently drift-able fields. Additionally,
`latest_version` was `integer` in TemplateDTO but a full VersionDTO object in
the getTemplate envelope — one key, two types, one API.

DocumentSummary/DocumentDetailResponse carry the same class of coupled triple
(`current_revision_id`, `revision_version`, `revision_number`) — required and
non-nullable, so no correctness bug, but the same maintainability defect.

## Decision

1. A wire-contract field set that references another resource version/revision
   (id + counters + labels + status) is always ONE nested required object —
   a version-reference value object — never parallel scalars.
2. When the pointer may not exist, the OBJECT is nullable as a whole
   (required, present-and-null, never absent), and consumers gate on the
   single object (`x.published_version == null`), never on inner fields.
3. Refs are per-bounded-context schemas (`TemplateVersionRef`,
   `DocumentRevisionRef`) — the pattern is unified, the schema is not.
   Forcing one shared schema across contexts is false unification.
4. In list views the ref is compact; in detail views the same field name may
   carry the full version object (AIP view semantics).
5. Pre-v1 exception: this lands as an atomic cutover (no expand/contract) —
   no external API consumers exist and v1 ships as a clean re-baseline.
   Post-v1, the same change class requires versioned deprecation.

## Consequences

- TemplateDTO: `latest_version: TemplateVersionRef` (required),
  `published_version: TemplateVersionRef | null` (required-nullable). Removed:
  `latest_revision_number`, `published_version_id`, `published_version_number`,
  `current_revision_number`.
- `TemplateVersionRef` includes `status`, unblocking "why not selectable" UX.
- Documents migrate `current_revision_id`/`revision_version`/`revision_number`
  → `current_revision: DocumentRevisionRef` in a follow-up plan (its
  `revision_version` participates in If-Match concurrency flows).
- Complements ADR 0035 (flat typed bodies): 0035 governs envelope shape, 0065
  governs field grouping. Closes 0035's optional-vs-null drift subclass
  structurally.
- Domain aggregates stop carrying join-projection scalars; read models own
  them (`domain.TemplateRead`).
