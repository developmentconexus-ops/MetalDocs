# Module: Controlled Documents — LEGACY TARGET CONCEPT

> **Status:** LEGACY / separate target context being retired
> **Marked:** 2026-08-14

The current `internal/modules/controlleddocuments` code still participates in runtime behavior, but the target no longer accepts `ControlledDocument` as a separate public/domain object beside `Document` and `DocumentRevision`.

Its legitimate responsibilities — stable governed identity, business code, owning Area/DocumentType binding, numbering/sequence and effective-revision identity — are being re-homed inside the Controlled Information design and explicit supporting policies such as NumberSeries.

Target authority:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

Do not create new seams, ports, permissions or workflow logic to preserve this module as a target bounded context. The detailed former living doc is available in Git history for later migration/deletion mapping.
