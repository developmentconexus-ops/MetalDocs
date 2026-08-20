---
id: engineering-rules
kind: authority
owner: engineering
summary: Defines repository-local engineering, Git, review, decision-consumption, and proof rules for the clean-slate rebuild.
---

# Engineering rules

## Method

Material decisions use the canonical DevelopmentConexus Engineering Method in `developmentconexus-ops/conexus-methodology/METHOD.md`.

Required posture:

```text
root cause before patch
smallest sustainable solution
one semantic authority per meaning
mechanism != authority
proof before implementation
unknown remains unknown
prepare the seam, not dormant implementation
```

## Implementation gate

`docs/status.md` decides whether implementation is allowed. While implementation is blocked:

- do not create application code or schemas;
- do not create deployment/runtime infrastructure;
- do not resurrect old implementation for convenience;
- do not prebuild future capabilities.

## Decision consumption

Before a remaining T-stage, read its current owning authorities plus `docs/decisions/forward-obligations.md`.

```text
PRESERVE → baseline evidence unless materially disproved
REOPEN   → owning stage must decide deliberately
DEFERRED → preserve seam/counterexample; create no dormant implementation
```

Current semantic authorities outrank older forward-obligation wording. When a stage closes a forward obligation, update the durable obligation page rather than creating an amendment chain.

## Evidence and reuse

Historical code may be inspected only for a concrete technical question. A removed unit is reusable only when all are true:

1. a current ratified consumer exists;
2. its public contract carries no superseded semantic authority;
3. its dependency direction fits the target;
4. its proof asserts the target property rather than the old shape;
5. reuse is simpler than rewrite after transition cost.

## Pull requests

Use one coherent ratifiable gate per branch/PR. Architecture/governance PRs remain Draft while temporary work exists.

Final architecture review uses one temporary `docs/work/current/ai-dialog.md`. Fable challenges; Lead adjudicates; operator ratifies. Reviewer output never becomes authority by itself.

Repository rules currently require:

```text
status context: required
ruleset id: 20560142
all review conversations resolved before merge
```

Renaming/removing the `required` job without updating the repository ruleset would silently un-gate `main` and is prohibited.

## Git and provenance

- no direct commits to `main`;
- no force-push/rewrite of shared history;
- prefer squash merge for architecture/governance gates;
- do not create live-tree archive/tombstone directories;
- destructive cleanup is allowed only when replacement truth or deliberate absence is explicit;
- an unmerged authority branch that is the only reachable provenance ref MUST remain reachable until an equivalent immutable archival ref/tag exists.

Current protected provenance refs for the reset are recorded in `docs/status.md` and `docs/decisions/repository-reset.md`.

## Verification

The repository is intentionally architecture-first while implementation is blocked. The current CI must prove the **allowed tree shape**, not enumerate known legacy names.

A failing legacy check is not a reason to restore legacy machinery. First ask whether the check protects a current target property. If not, retire it; if yes, replace it with the smallest check that proves that property.

## Closure

Do not claim `done`, `green`, or `merged` without revalidating current remote HEAD, required status, unresolved review threads, and the final changed-file/tree shape.