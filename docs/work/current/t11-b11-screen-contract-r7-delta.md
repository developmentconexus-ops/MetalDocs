# T11 — B11 Access Administration — P9 R7 User Picker Delta

> **Status:** COMPLETE / OPERATOR RE-LOCKED / PASS.  
> **P8 delta:** `docs/work/current/t11-b11-grant-user-picker-p8-r7.html`  
> **P8 R7 blob:** `3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df`  
> **Finding:** `docs/work/current/t11-b11-p8-r7-review-finding.md`.

## Preserved contract

All R6/P9 rows remain protected except the grant User picker traversal previously named R6-03. R7 supersedes only that row.

## Closed row

| Surface | Accepted truth | Correct interaction | Failure | Forbidden shortcut | Status |
|---|---|---|---|---|---|
| Grant User subject picker | op6 `listUsers` / raw `UserPage`, `user_id ASC`, cursor pagination | page raw server rows; render ENABLED selectable and DISABLED visible/unavailable; later page can reach final Subject × Role × Scope review | failed continuation preserves loaded page/draft | pre-pagination client state filter; hidden all-page crawl; invented state filter/search | READY / PASS |

## Bidirectional trace

```text
op6 UserPage
→ raw grant User page
→ per-row availability guidance
→ selected enabled user_id
→ existing createRoleAssignment review/command
```

```text
grant User Previous/Next
→ op6 opaque cursor continuation

DISABLED row
→ visible recognition only
→ no selectable user_id
```

The frontend does not reshape cursor pages after receipt and does not require any new op6 filter.

## Exact R7 proof

```text
page 1 raw boundary
  João Almeida
  Beatriz Silva
  Rafael Siqueira
  Ana Torres

page 2 raw boundary
  Bruno Vieira
  Carla Nunes
  Paulo Mendes — DISABLED / unavailable
  Sofia Barros

page 3 raw boundary
  Luciana Prado
  Diego Ramos
  Mariana Costa
  Felipe Moraes
```

Exact R7 bytes:

```text
Git blob              3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
static + Chromium     12 / 12 PASS
JavaScript parse      PASS
operator disposition  APPROVED / RE-LOCK
```

## Verdict

```text
R6-03 grant User picker      SUPERSEDED BY R7
R7 grant User picker         READY / PASS
new backend capability       0
new application operation    0
hidden all-page crawl        0
pre-pagination state filter  0
unresolved P9 R7 findings    0
```

R6 remains the accepted complete B11 base artifact for all unaffected structure. R7 is the exact operator-LOCKED amendment for grant User picker page fidelity.
