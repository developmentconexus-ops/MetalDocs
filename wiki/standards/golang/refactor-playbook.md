---
extends:
  - project-process
evidence:
  - wiki/reviews/2026-05-21-go-backend-review.md
enforced_by:
  - manual-review
---
# Refactor Playbook

## Overview

Use this playbook to bring any `internal/modules/<name>` or `internal/platform/<name>` package up to the Go quality bar without mixing unrelated work.

## Phase 0: Scope

Identify the package path and run `ecc:go-review` with `wiki/standards/golang/README.md` as the rubric. Save findings to `wiki/reviews/<date>-go-backend-review/<module-slug>.md`.

## Phase 1: Critical Fixes

Land all Critical findings first. Each fix gets its own commit:

```text
fix(<module>): <finding-ID> <one-line description>
```

Update the findings doc with the commit SHA.

## Phase 2: High Fixes

Land High findings next. Security-boundary Highs take priority over API-shape Highs. Use the same commit and evidence discipline.

## Phase 3: Lint Baseline

Run:

```bash
golangci-lint run ./path/to/module/...
```

Document surviving issues in `wiki/reviews/<date>-go-backend-review/<module-slug>-lint-baseline.md`. Gate future PRs with `--new-from-rev` if a full cleanup is too large.

## Phase 4: Medium and Low

Medium findings get follow-up PRs. Low findings are fixed opportunistically when touching the same file.

## Phase 5: Bar Update

If a fix introduces a new reusable pattern, update the closest doc in `wiki/standards/golang/` in the same PR.

## Split-and-Conquer Rule

If a module exceeds 500 LoC of changes, split it like platform #2a was split from the broader platform review.

## LoC Bucket Guidance

- `<200` LoC: one PR
- `200-500` LoC: two or three PRs
- `>500` LoC: split by sub-module or concern

## Tracker Update

After each phase, update `wiki/reviews/2026-05-21-go-backend-review.md` and the cursor in project memory.
