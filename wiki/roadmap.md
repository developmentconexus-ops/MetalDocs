# MetalDocs Roadmap

> **Last verified:** 2026-08-14
> **Status:** ACTIVE — Cohesive Platform Redesign
> **Scope:** Single forward progression surface for MetalDocs.

## Active program

MetalDocs is in a **design-first whole-platform architecture reset**.

Canonical program authority:

- [architecture/cohesive-platform-redesign.md](architecture/cohesive-platform-redesign.md)
- staging ledger: `docs/superpowers/analysis/2026-08-14-cohesive-platform-redesign-ledger.md`
- recovery pointer: [references/current-agent-handoff.md](references/current-agent-handoff.md)

**No product implementation is authorized while the integrated-design gate is closed.**

## Progression

The forward order is semantic, not module-by-module:

| Phase | Outcome | Status |
|---|---|---|
| R0 — AuthN / Organization / AuthZ | identity boundary, Areas/Groups, five roles, scoped grants | **design approved** |
| R1 — Approval V1 | versioned sequential human approval model, participants, decisions, evidence | **design approved** |
| R2 — Controlled Information north star | Document + Revision, template-as-revision-role, Area moved to Organization, release boundary | **direction approved** |
| R3 — Controlled Information configuration | DocumentType, Family, GovernanceClass, TemplateDesignation/default semantics | **NEXT** |
| R4 — Document / Revision lifecycle | draft/submission/effectivity, reason-for-change, obsolete/supersede, immutable evidence | pending |
| R5 — Numbering + template payload | NumberSeries, template provenance/spec, derived-document semantics | pending |
| R6 — Periodic Review + Renditions + Release | review cadence, rendering provenance, reconstruction, effectivity | pending |
| R7 — Distribution + Tokens + Audit + Notifications + Search | supporting product semantics and event/read-model boundaries | pending |
| R8 — Tenant lifecycle / Security | owner/platform authority, deletion lifecycle, external IdP trigger | pending |
| R9 — Final Authorization matrix | final Permission Catalog, role bundles, workflow/domain checks | pending |
| R10 — Technical target | bounded contexts, data model, table/tx ownership, events, APIs, frontend journeys, build-vs-buy | pending |
| R11 — Migration specification | explicit delete/move/rename/rewrite/retain map from current code | pending |
| R12 — Final ADR/spec review | promote durable truth, adversarial review, operator approval | pending |
| R13 — Implementation plan | code-ready plan with no architecture ambiguity | blocked by R12 |
| R14 — Implementation | execute the accepted plan | blocked |

## Historical plans

All earlier roadmap/milestone/spec execution sequences are **historical evidence only**. The old `docs/superpowers` planning tree was removed from the live repository on 2026-08-14 and remains available through Git history.

Do not carry an old open item forward unless the active redesign independently confirms that the underlying requirement still exists and gives it an owner in the new target.

## Next action

Continue R3 in the active redesign ledger:

```text
DocumentType
+ DocumentFamily
+ GovernanceClass
+ TemplateDesignation/default policy
```

No code.
