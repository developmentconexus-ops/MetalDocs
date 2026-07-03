# F0.2 — ADR 0065 + ADR 0035 class annotation · Evidence

Feature: record the durable decision behind F0.1 and close the flat/envelope
drift class it resolves. Committed in `d0b1ba84` alongside the backend cutover.

## Artifacts
- **ADR 0065** authored: `wiki/decisions/0065-version-references-are-nested-value-objects.md` — "version references are nested value objects" (read/write model split `domain.TemplateRead` vs `domain.Template`; `latest_version` required ref; `published_version` required-and-nullable ref; four coupled flat scalars removed). Status Accepted.
- **Decisions index** updated: `wiki/decisions/index.md` — row added after 0064.
- **ADR 0035 class annotation**: the flat/envelope drift class ([[adr0035-flat-envelope-drift]]) is annotated as resolved for the template version-pointer surface — the nested-ref shape removes the parallel-scalar coupling that let hand-written `body.data.X` consumers drift.

## Verification
- Files present and internally linked (index row resolves to the ADR file).
- No code in this feature; correctness is F0.1's runtime + gate evidence, which cite ADR 0065 as the governing decision.

## Bounded defers
- Memory note `adr0035-flat-envelope-drift.md` to be annotated (Task 12, wiki/memory sync) — narrows the class to the still-open documents surface (Plan 2).
