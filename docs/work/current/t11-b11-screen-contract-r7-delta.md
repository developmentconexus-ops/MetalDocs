# T11 — B11 Access Administration — P9 R7 User Picker Delta

> **Status:** CANDIDATE / AWAITING OPERATOR RE-LOCK.  
> **P8 delta:** `docs/work/current/t11-b11-grant-user-picker-p8-r7.html`  
> **P8 R7 blob:** `3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df`  
> **Finding:** `docs/work/current/t11-b11-p8-r7-review-finding.md`.

## Preserved contract

All R6/P9 rows remain protected except the grant User picker traversal previously named R6-03.

## Reopened row

| Surface | Accepted truth | Correct interaction | Failure | Forbidden shortcut | Status |
|---|---|---|---|---|---|
| Grant User subject picker | op6 `listUsers` / raw `UserPage`, `user_id ASC`, cursor pagination | page raw server rows; render ENABLED selectable and DISABLED visible/unavailable; later page can reach final Subject × Role × Scope review | failed continuation preserves loaded page/draft | pre-pagination client state filter; hidden all-page crawl; invented state filter/search | CANDIDATE |

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

The frontend does not reshape cursor pages after receipt.

## Exact R7 probe

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

On exact local bytes matching blob `3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df`:

```text
static + Chromium 12 / 12 PASS
JavaScript parse PASS
```

## Exit gate

```text
operator operates exact R7 delta
→ explicit re-LOCK of grant User picker only
→ this row READY / PASS
→ preserve amended Evidence
→ update durable B11 locator as R6 + R7 amendment
→ clean docs/work/**
→ close PR review thread with exact proof
```

Until then, B11 has one bounded material finding open and PR #173 remains Draft.
