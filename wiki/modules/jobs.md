# Module: jobs — LEGACY MODULE CLASSIFICATION

> **Status:** CURRENT runtime orchestration / NOT a target bounded context
> **Marked:** 2026-08-14

The current `internal/modules/jobs` directory hosts River periodic maintenance/orchestration over other domains. The architecture audit already showed it has no coherent business vocabulary of its own.

Target ruling:

> Periodic/background orchestration is composition/infrastructure, not a business bounded context merely because it lives under `internal/modules/` today.

The jobs themselves remain real operational responsibilities and must survive where needed; the **module classification** does not.

Read target authority:

- `wiki/architecture/cohesive-platform-redesign.md`
- `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`

Do not invent business domain ownership inside `jobs`. Re-home each job later with its owning domain or explicit orchestration/composition boundary. Detailed previous living documentation remains available through Git history.
