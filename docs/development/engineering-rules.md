---
id: engineering-rules
kind: authority
owner: engineering
summary: Defines repository-local engineering, Git, review, and proof rules that remain valid during the clean-slate rebuild.
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

## Git

- no direct commits to `main`;
- no force-push/rewrite of shared history;
- prefer squash merge for architecture/governance gates;
- do not create live-tree archive branches/directories as a substitute for Git history;
- destructive cleanup is allowed only when the replacement truth or deliberate absence is explicit.

## Closure

Do not claim `done`, `green`, or `merged` without revalidating the current remote HEAD and required checks.

A failing legacy check is not automatically a reason to restore legacy machinery. First ask whether the check protects a current target property. If not, retire it; if yes, replace it with the smallest check that proves that property.