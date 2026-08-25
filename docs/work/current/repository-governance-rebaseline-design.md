# MetalDocs Repository Governance Rebaseline — Design

> **Status:** OPERATOR-APPROVED DESIGN / TEMPORARY WORK
> **Branch:** `chore/repository-governance-rebaseline`
> **Scope:** repository governance and context routing only. B11 remains frozen on PR #173.

## Goal

Restore the fast pre-global operating model without reintroducing the failed external methodology-router dependency: three local stable methods, `AGENTS.md` as bootstrap/router, `docs/index.md` as task/intention router, a compact snapshot `docs/roadmap.md`, and current decision/provenance documents that describe only controls that actually exist.

## Target operating model

```text
fresh session
→ revalidate repository / branch / main / PR / required
→ AGENTS.md
→ docs/roadmap.md
→ select local method(s)
→ docs/index.md
→ smallest relevant authority pack
→ search/open only the relevant section/operation first
→ expand only when another owner can materially change the conclusion
```

Local methods:

```text
docs/development/engineering-method.md
  reasoning / Global Maximum / materiality / proof / reopen

docs/development/repository-method.md
  repository continuity / selective context / documentation / Git / PR / Evidence

docs/development/frontend-product-experience-planning-method.md
  P0-P14 / P8 / operator LOCK / P9 / P10 / bounded upstream reopen
```

`engineering-method.md` and `frontend-product-experience-planning-method.md` remain byte-identical to their currently accepted versions.

## Authority boundaries

| Surface | Owns | Must not own |
|---|---|---|
| `AGENTS.md` | bootstrap, method routing, hard stops, verification entrypoint | method prose, roadmap history, Product semantics |
| `docs/roadmap.md` | current stage/block, blockers, implementation gate, exact next action | review chronology, full finding narratives, Evidence history |
| `docs/index.md` | task/intention → smallest starting authority pack + conditional expansion + negative guidance | Product/architecture semantics, status |
| `docs/decisions/index.md` | compact current decision/disposition registry | review chronology, duplicate decision bodies |
| `docs/development/repository-method.md` | stable repository operating method | Product status/semantics, MetalDocs-specific mechanics |
| `docs/development/engineering-rules.md` | MetalDocs-specific Git/CI/provenance/implementation specializations | reusable method text |
| `docs/work/**` | temporary branch-only working state | durable authority or merge-candidate content |
| Evidence refs/locators | exact byte recovery/provenance | current Product/status authority |

## Selective context law

Default context is intentionally small but never a correctness cap:

```text
AGENTS + roadmap + selected method(s) + 1-2 task owners
```

`docs/index.md` should route by task/intention with:

```text
START WITH
ADD WHEN
DO NOT READ BY DEFAULT
```

When an owner file is large, search the relevant symbol/section/operation first. Read the whole file only when whole-owner coherence can materially change the conclusion. Expand beyond the starting pack whenever Evidence names another owner capable of falsifying the result.

No hard token/file budget may override correctness.

## Roadmap law

The roadmap is a **snapshot, not a journal**. It contains only:

- current program/stage/block;
- integrated checkpoints relevant to progression;
- active blocking findings/prerequisites at compact level;
- implementation allowed/blocked;
- exact next action/hard stops.

Detailed Bxx iteration/review history belongs to temporary work, Git/PR history, or exact Evidence locators when byte recovery is required.

## Acceptance-increment law

```text
main
→ one branch + Draft PR for one coherent increment
→ work/iteration
→ operator decision gates
→ one full invariant/adversarial sweep before Ready
→ preserve one final exact Evidence checkpoint when required
→ remove docs/work/**
→ Ready once
→ required
→ explicit merge authorization
→ squash merge
→ delete ordinary head branch
```

If frontend Evidence proves an upstream Product/backend/wire authority insufficient, split/reopen the smallest owner rather than growing one PR into a permanent workspace.

## Evidence law

Preserve exact Evidence only when a current consumer needs byte-level reconstruction. Prefer one final canonical Evidence checkpoint per accepted block/increment. Earlier iterations remain in branch/Git history unless independently required.

Evidence locator stays short: canonical ref/commit/blob, protected semantics, retrieval law, reopen/retirement trigger. It is not a second roadmap.

## Migration corrections

This rebaseline must remove active claims that no longer match repository operation:

- no external `conexus-methodology` router/pin in the active route;
- no active `REPO-STD-V1` decision row as current authority;
- no statement that Evidence refs are currently SHA-checked by CI when current `required` does not do that;
- no reference to non-existent documentation roots as current required structure;
- no duplicate documentation-governance authority when repository-method + engineering-rules already own the meaning.

## Files in this increment

Create:
- `docs/development/repository-method.md`

Modify:
- `AGENTS.md`
- `docs/index.md`
- `docs/roadmap.md` only enough to record the governance rebaseline candidate without importing B11 branch status
- `docs/development/engineering-rules.md`
- `docs/decisions/index.md`
- `docs/decisions/repository-reset.md`
- `.github/workflows/ci.yml` only to require the new repository-method file and protect objective current properties

Delete if no unique current consumer remains:
- `docs/development/documentation.md`

Temporary design/plan under `docs/work/current/` must be removed before Ready.

## Deliberately out of scope

- B11/R8 correction or PR #173 content;
- deletion of any existing branch/ref;
- rewriting Engineering Method or Frontend Method;
- Product/architecture semantic redesign;
- architecture document splitting;
- repository-history rewrite;
- large CI framework or context-budget enforcement;
- changing implementation authorization.

## Success tests

A fresh actor can answer from the live tree:

1. where are we? → roadmap;
2. which method applies? → AGENTS;
3. which Product/architecture owners should I start with? → index;
4. which decision refined that owner? → decisions index;
5. what is Evidence versus authority? → repository method / locator;
6. can frontend reopen old planning when real Evidence proves it insufficient? → Engineering + Frontend methods;
7. can this be done without chat archaeology or recursive whole-repo reading? → yes.

Structural checks:

- Engineering Method hash unchanged;
- Frontend Method hash unchanged;
- no active external methodology-router claims in bootstrap/governance files;
- `AGENTS.md`, `docs/index.md`, `docs/roadmap.md` remain compact and non-duplicative;
- `required` still objective and single-entry;
- merge candidate contains no `docs/work/**`;
- B11 remains frozen/not integrated by this increment.
