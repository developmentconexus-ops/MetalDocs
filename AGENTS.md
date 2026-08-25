# MetalDocs Agent Bootstrap

## Start

Before relying on chat, handoff, or remembered state:

1. revalidate repository identity, current branch/HEAD, remote `main`, and the relevant PR state;
2. read [`docs/roadmap.md`](docs/roadmap.md) to understand the current Product/planning state and implementation gate;
3. for material engineering decisions, follow [`docs/development/engineering-method.md`](docs/development/engineering-method.md);
4. for frontend Product Experience planning, follow [`docs/development/frontend-product-experience-planning-method.md`](docs/development/frontend-product-experience-planning-method.md);
5. use [`docs/index.md`](docs/index.md) as a navigation aid when useful, never as a reading boundary.

There is **no fixed file count, owner count, or context budget**. Investigate any repository area, Git history, Evidence, code, contracts, runtime behavior, research, or external source that can materially change, challenge, or falsify the conclusion. Whole-repository review is explicitly allowed when the task calls for it. Context efficiency is an optimization, not a correctness boundary.

## Methods and authority

The repository adopts these local method files:

- [`engineering-method.md`](docs/development/engineering-method.md) — DevelopmentConexus Engineering Method v1.0.0;
- [`frontend-product-experience-planning-method.md`](docs/development/frontend-product-experience-planning-method.md) — Frontend Product Experience Planning Method v2.3.

Their contents are intentionally shared unchanged across the DevelopmentConexus product repositories. Do not locally rewrite or reinterpret them by convenience.

- Current accepted repository authority outranks chat, handoff, memory, historical snapshots, and reviewer preference.
- `docs/roadmap.md` owns current stage/status/allowed work/next action.
- Accepted Product, architecture, contract, and decision artifacts own their stated semantics.
- Evidence, code, tests, research, Git history, and reviewer output support decisions; they do not become Product authority merely by existing.
- Seek the **Global Maximum**, not merely the best answer inside the current structure.
- If downstream Evidence falsifies an upstream decision, stop only the affected scope and reopen the smallest owning authority according to the adopted methods. Do not silently patch around the contradiction.

## MetalDocs safety rails

Unless explicitly reopened by accepted authority:

- Product implementation remains blocked while `docs/roadmap.md` says `IMPLEMENTATION BLOCKED`.
- Do not restore removed legacy implementation merely because it existed or was tested.
- Historical mechanism reuse requires a current named consumer and the proof-backed gate in `docs/architecture/technical-baseline.md`.
- Embedded pre-reset `wiki/...` paths are provenance only, not current routing.
- B01–B10 LOCKs remain accepted unless material Evidence falsifies them.
- B11/B12, FP2/P11, T12, and later implementation work open only when `docs/roadmap.md` explicitly permits them.
- Assistant/reviewer/tool output cannot set operator-only LOCK state.

## Git and verification

Preserve unowned state. Never force-push or rewrite shared history. Never merge without explicit operator authorization.

Required CI is intentionally one objective aggregate check named `required`:

```text
GitHub Actions → required
```

Run additional targeted proof when the current claim requires it. A red required CI check should mean an objective repository property is broken, not that a planning preference, context budget, router convention, methodology pin, Evidence-branch SHA, or documentation ceremony was violated.

`CLAUDE.md` has no independent Product, architecture, roadmap, status, or methodology authority.
