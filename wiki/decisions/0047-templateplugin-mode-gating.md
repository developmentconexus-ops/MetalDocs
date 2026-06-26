# 0047 — templatePlugin mode-gating

- **Status:** Proposed
- **Last verified:** 2026-06-23
- **Scope:** When eigenpal's `templatePlugin` (placeholder authoring) is active in the MetalDocs editor.

## Context

The editor supports multiple modes (`template-draft`, `document-edit`, `readonly`,
`review`). Placeholder authoring (eigenpal `templatePlugin`) is only meaningful while
drafting a template, not while editing or reviewing a filled document. The gating rule
lived implicitly in `MetalDocsEditor.tsx` with no decision record (tech-debt T-007).
Under the ACL (ADR 0046) this becomes an adapter responsibility behind the
`placeholders` capability, so the rule must be recorded.

## Decision

`templatePlugin` is assembled only for the template-authoring mode. The MetalDocs
`EditorMode` → capability mapping is the single source of truth: the `placeholders`
capability is enabled for `template-draft` and disabled otherwise. Under ADR 0046 the
adapter owns the eigenpal-specific realization (assembling `templatePlugin`); callers
only toggle the capability.

## Consequences

- Placeholder authoring cannot leak into document-edit/readonly/review modes.
- The rule is testable at the capability layer (assert the assembled plugin set per mode).
- Closes tech-debt T-007.

## Alternatives considered

- **Always register `templatePlugin`, hide its UI per mode.** Rejected: leaves authoring affordances reachable and couples UI visibility to plugin presence.

## Related

- ADR 0046 (Eigenpal Anti-Corruption Layer).
- `packages/editor-ui/src/MetalDocsEditor.tsx` (current mode→plugin assembly).
