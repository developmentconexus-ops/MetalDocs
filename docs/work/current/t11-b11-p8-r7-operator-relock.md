# T11 — B11 Access Administration — P8 R7 Operator Re-LOCK

> **Status:** OPERATOR-APPROVED / RE-LOCK.  
> **Artifact:** `docs/work/current/t11-b11-grant-user-picker-p8-r7.html`  
> **Git blob:** `3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df`  
> **Scope:** grant User picker page fidelity only.

## Trigger

PR #173 review correctly identified that R6 filtered enabled Users before paginating the grant User picker even though op6 `listUsers` exposes an unfiltered cursor-paginated `UserPage` ordered by `user_id ASC`.

That changed page boundaries and would have required the hidden crawl/post-filter explicitly forbidden by B11.

## Preserved scope

The finding did not reopen:

```text
/admin/access route
Por Área / Grupos / Funções IA
R4/R5 low-fi frame
R6 Group member pagination
R6 add-member User pagination
R6 grant Group pagination
R6 grant Area pagination
membership semantics
Role semantics
contextual Area/Group grant entry
Subject × Role × Scope final review
exact revoke
same-key ambiguous retry
B11-F1 op31 precision
Authorization authority
P10
89-operation census
```

## R7 correction

R7 proves the exact smallest correction:

```text
op6 raw UserPage
→ preserve server page boundary
→ ENABLED row selectable
→ DISABLED row visible but unavailable
→ no pre-pagination client filter
→ no hidden page crawl
→ no new op6 query/filter/search
```

Exact proof on the operator-operated artifact:

```text
Git blob              3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
static + Chromium     12 / 12 PASS
JavaScript parse      PASS
```

Representative raw page boundary:

```text
page 2
  Bruno Vieira
  Carla Nunes
  Paulo Mendes — DISABLED / unavailable
  Sofia Barros
```

Paulo remains visible because he belongs to the raw op6 page, but cannot become the selected grant User. Mariana Costa remains reachable on a later raw page and reaches the normal Subject × Role × Scope review.

## Operator disposition

**APPROVED.**

This approval is interpreted according to the explicit gate presented to the operator: re-LOCK of the grant User picker page-fidelity delta only.

Therefore:

```text
R6-03 grant User picker      superseded only where page fidelity conflicted
R7 grant User picker         LOCKED
P9 R7 row                    READY / PASS
all unaffected R6 structure PRESERVED
```

## Reopen triggers

Reopen only on material Evidence such as a proven need for server-side User filtering/search, inability to keep opaque cursor traversal usable at real scale, or P11 integration exposing a contradiction. Preference for hiding disabled Users or preloading/crawling all Users is not a reopen trigger.
