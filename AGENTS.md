# MetalDocs Agent Bootstrap

## Start

Before relying on chat, handoff, or remembered state:

1. revalidate repository identity, current branch/HEAD, remote `main`, relevant PR state, and `required` when applicable;
2. read [`docs/roadmap.md`](docs/roadmap.md) for the current stage/block, implementation gate, blockers, and exact next action;
3. select only the local method(s) required by the task;
4. use [`docs/index.md`](docs/index.md) to reach the smallest relevant Product/architecture/decision authority pack;
5. start with the relevant section/operation inside large owners and expand only when another source can materially change or falsify the conclusion.

Default task context is intentionally small, normally **1–2 task-owning repository authorities plus the applicable method(s)**. This is a routing default, never a correctness cap. Do not recursively read `docs/`, Git history, closed PRs, Evidence refs, research, or removed implementation without a named material reason.

## Local methods

MetalDocs operates with three local methods:

- [`engineering-method.md`](docs/development/engineering-method.md) — DevelopmentConexus Engineering Method v1.0.0: materiality, root cause, Global Maximum, proof, challenge, decision and reopen;
- [`repository-method.md`](docs/development/repository-method.md) — repository continuity: authority recovery, selective context, documentation, Evidence, Git and acceptance increments;
- [`frontend-product-experience-planning-method.md`](docs/development/frontend-product-experience-planning-method.md) — Frontend Product Experience Planning Method v2.3: P0–P14, functional P8, operator LOCK, P9/P10 and bounded upstream reopen.

Use them proportionally:

```text
material engineering decision
  → Engineering Method

repository / Git / context / documentation / PR continuity
  → Repository Operating Method
  → add Engineering Method when the governance decision itself is material

frontend Product Experience planning
  → Engineering Method + Frontend Method
  → then the Product/architecture owners routed by docs/index.md
```

`engineering-method.md` and `frontend-product-experience-planning-method.md` are shared accepted method texts and must not be locally rewritten by convenience. The Repository Operating Method owns the local operating model; there is no external methodology router or moving methodology dependency in the active path.

## Authority

- Current accepted repository authority outranks chat, handoff, memory, historical snapshots, PR descriptions, and reviewer preference.
- `docs/roadmap.md` is the sole mutable stage/status/allowed-work/next-action authority.
- Accepted Product, architecture, contract, and bounded decision artifacts own their stated semantics.
- `docs/index.md` and `docs/decisions/index.md` route to owners; they do not become parallel semantic authorities.
- Evidence, code, tests, research, Git history, and reviewer output support or falsify decisions; they do not become Product authority merely by existing.
- Seek the **Global Maximum**, not merely the best answer inside the current structure.
- If downstream/frontend Evidence proves current Product/backend/wire authority insufficient, stop only the affected scope and reopen the smallest owning authority through the adopted methods. Do not silently patch around the contradiction and do not degrade a proven user need merely to fit the current backend.

## MetalDocs safety rails

Unless explicitly reopened by accepted authority:

- Product implementation remains blocked while `docs/roadmap.md` says `IMPLEMENTATION BLOCKED`.
- Do not restore removed legacy implementation merely because it existed or was tested.
- Historical mechanism reuse requires a current named consumer and the proof-backed gate in `docs/architecture/technical-baseline.md`.
- Embedded pre-reset `wiki/...` paths are provenance only, not current routing.
- B01–B10 frontend LOCKs remain accepted unless material Evidence falsifies them.
- B11/B12, FP2/P11, T12, and later implementation work open only when `docs/roadmap.md` explicitly permits them.
- Assistant/reviewer/tool output cannot set operator-only LOCK state.

## Git and verification

Preserve unowned state. Never force-push or rewrite shared history. Never merge without explicit operator authorization.

Normal work uses one coherent acceptance increment per PR. Keep the PR Draft while it is a workspace; make it Ready only when it is believed integration-ready after its final proof/review sweep. Temporary `docs/work/**` never enters a merge candidate/main.

Required CI is one objective aggregate check named:

```text
required
```

Run targeted proof when the current claim requires it. A red `required` check should mean an objective repository property is broken, not that a planning preference, context budget, UX judgment, historical Evidence ref, or documentation ceremony was violated.

`CLAUDE.md` has no independent Product, architecture, roadmap, status, or methodology authority.
