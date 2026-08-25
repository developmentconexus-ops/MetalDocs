# MetalDocs Agent Bootstrap

## Start

Before relying on chat, handoff, or memory:

1. revalidate repository identity, branch/HEAD, remote `main`, relevant PR and `required` when applicable;
2. read [`docs/roadmap.md`](docs/roadmap.md) for current stage/block, blockers, implementation gate and next action;
3. select only the local method(s) required by the task;
4. use [`docs/index.md`](docs/index.md) to reach the smallest relevant Product/architecture/decision authority pack;
5. start with the relevant section/operation inside large owners and expand only when another source can materially change or falsify the conclusion.

Default task context is intentionally small: normally the applicable method(s) + **1–2 task owners**. This is a routing default, never a correctness cap. Do not recursively read `docs/`, history, old PRs, Evidence refs, research or removed implementation without a named material reason.

## Methods

| Work | Method |
|---|---|
| Material engineering / Global Maximum / proof / reopen | [`engineering-method.md`](docs/development/engineering-method.md) |
| Repository / Git / context / documentation / PR continuity | [`repository-method.md`](docs/development/repository-method.md); add Engineering Method when the governance decision itself is material |
| Frontend Product Experience | Engineering Method + [`frontend-product-experience-planning-method.md`](docs/development/frontend-product-experience-planning-method.md) + Product/architecture owners from `docs/index.md` |

Engineering Method v1.0.0 and Frontend Method v2.3 are accepted shared texts and must not be locally rewritten by convenience. Repository operation is local; there is no external methodology router or moving methodology dependency in the active path.

## Authority

- Current accepted repository authority outranks chat, handoff, historical snapshots, PR descriptions and reviewer preference.
- `docs/roadmap.md` alone owns mutable stage/status/allowed work/next action.
- Product, architecture, contract and bounded decision files own their stated semantics.
- `docs/index.md` and `docs/decisions/index.md` route to owners; Evidence/Git/research/reviewer output support or falsify decisions but do not silently become Product authority.
- Seek the **Global Maximum**. If downstream/frontend Evidence proves current authority insufficient, reopen only the smallest owner through the adopted methods; do not silently patch around it or degrade a proven user need to fit the current backend.

## MetalDocs safety rails

Unless explicitly reopened by accepted authority:

- implementation stays blocked while the roadmap says `IMPLEMENTATION BLOCKED`;
- do not restore superseded implementation by convenience;
- historical mechanism reuse must pass `docs/architecture/technical-baseline.md`;
- embedded pre-reset `wiki/...` paths are provenance only;
- B01–B10 LOCKs remain accepted unless material Evidence falsifies them;
- B11/B12, FP2/P11, T12 and implementation open only when the roadmap permits them;
- assistant/reviewer/tool output cannot set operator-only LOCK state.

## Git and verification

Preserve unowned state. No force-push/shared-history rewrite. Never merge without explicit operator authorization.

Use one coherent acceptance increment per PR. Draft is the workspace; Ready means the complete candidate is believed integration-ready. Temporary `docs/work/**` never enters a merge candidate/main.

Required CI has one objective aggregate check:

```text
required
```

Run targeted proof for claims that need it. CI does not decide Global Maximum, UX quality, context budgets, historical Evidence freshness or reviewer preference.

`CLAUDE.md` has no independent Product, architecture, roadmap, status or methodology authority.
