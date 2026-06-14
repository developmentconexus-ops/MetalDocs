# Merge Reconciliation — `main` vs `qa/iam-area-membership`

> **Date:** 2026-06-14
> **Trigger:** Operator request "merge qa to main".
> **Verdict:** **Do not `git merge`. `qa` supersedes `main` entirely — reset `main` to `qa`.** A literal merge produces 2220 conflicts (the history was re-hashed, so git treats identical content as unrelated `add/add`). The correct operation is to make `main` point at `qa`'s tip.

## What the branches actually are

`qa` is **`main`'s history rewritten** (a history scrub re-hashed every post-split commit — the F-18 secret cleanup), **plus 299 additional June v1 commits**. It is not a parallel reconstruction and not a normal feature branch.

| Metric | Value |
|---|---|
| Common ancestor | `905ddfd3` — 2026-04-10 |
| `main` tip | `c63be611` — 2026-06-03 (IAM Admin Center PR-1..12 + 12b) |
| `qa` tip | `c7f06f2e` — 2026-06-14 (v1 re-baseline) |
| Commit **messages** on `main` but NOT on `qa` | **10** — all trivial (`docs(review): mark Hx fixed …` tracking notes, May 21–22) + 1 superseded `restore eigenpal tarball` |
| Commit **messages** on `qa` but NOT on `main` | **299** (qa's extra v1 work) |
| `git merge-tree main qa` | rc=1, **2220 conflicts** (re-hash artifact, not real divergence) |

## "main-unique files" were superseded, not lost

The ~22 files present at `main`'s tip but absent from `qa` are **prototypes `qa` later replaced or relocated**, confirmed by qa history:

- `frontend/.../iam/AreaMembershipAdminPage.tsx` (old single page) → `qa` rebuilt it as a full **7th Admin Center tab**: `tabs/MembershipsTab.tsx`, `components/{GrantMembershipDialog,RevokeMembershipDialog,MembershipsDirectory,MembershipKpiStrip,MembershipsFilterBar,UserMembershipsTable}`, `queries/*`, `mutations/*`, plus backend `delivery/http/routes_memberships.go`, `infrastructure/postgres/{user_area_repository,area_catalog_reader}.go`, `application/membership_governance_logger.go`, `tests/integration/iam/membership_area_scope_test.go`. qa commits: *"PR-2 frontend rebuild + 7th Admin Center tab"*, *"close 1 HIGH + 1 HIGH + 1 MEDIUM on Area Membership Admin"*.
- `internal/modules/documents/http/*` → `qa` relocated to `documents/delivery/http/*`.
- Remaining `documents/`, `render/fanout/`, `taxonomy/` files: superseded by qa's 299 later commits (every substantive `main` commit message appears on `qa`).

Conclusion: **`qa ⊇ main`** in content. `main` holds nothing worth porting.

## Action taken

```
git tag archive/main-pre-qa-20260614 main        # old main recoverable forever
git checkout main
git reset --hard qa/iam-area-membership           # main := qa
git push origin main --force-with-lease           # origin/main was 1946 behind
```

`main` is now the canonical source line. Old `main` preserved at tag `archive/main-pre-qa-20260614`.

## Method (reproducible)

```
BASE=905ddfd375a4040d5796b1cf1ca829988006ae73
git log --format='%s' $BASE..main  | sort > main_msgs   # vs qa: only 10 differ (all trivial)
git log --format='%s' $BASE..HEAD  | sort > qa_msgs      # 299 qa-only
git merge-tree --write-tree main HEAD                    # rc=1, 2220 conflicts (re-hash artifact)
```
