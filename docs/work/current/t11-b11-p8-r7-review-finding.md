# T11 — B11 P8 R7 — Grant User Picker Finding

> **Status:** MATERIAL / BOUNDED / R6-03 REOPEN ONLY.  
> **Trigger:** PR #173 review after R6 re-LOCK.  
> **Implementation:** BLOCKED.

## Finding

R6 correctly added visible pagination for the grant User picker, but it formed its fixture collection as:

```text
all Users
→ client filter state=ENABLED
→ paginate
```

Current op6 authority is instead:

```text
GET /api/v1/users
listUsers
UserPage
PAGED
user_id ASC
no state filter
```

Therefore R6 changed page boundaries before rendering. A production frontend could reproduce those pages only by crawling/post-filtering multiple server pages, which B11 explicitly forbids.

## Smallest-owner disposition

```text
CURRENT STRUCTURE CONFIRMED
+ P8/P9 local correction only
```

No Product/backend/wire reopen is justified.

Reopened only:

```text
Grant User subject picker
P9 R6-03 op6 traversal proof
```

Preserved:

```text
all other R6 pagination surfaces
R6 member pagination
R6 add-member User pagination
R6 grant Group pagination
R6 grant Area pagination
B11 IA/frame/non-pagination semantics
B11-F1
P10
89-operation census
```

## Target invariant

```text
server UserPage order/boundaries remain authoritative
→ render every returned User in that page
→ ENABLED User selectable
→ DISABLED User visible but unavailable
→ no local pre-pagination filter
→ no hidden all-page crawl
```

## P8 R7 proof target

Exact candidate:

```text
docs/work/current/t11-b11-grant-user-picker-p8-r7.html
Git blob 3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
```

Required operator checks:

```text
page 1: João / Beatriz / Rafael / Ana
page 2: Bruno / Carla / Paulo DISABLED / Sofia
Paulo cannot be selected
continuation failure preserves page 2
page 3: Luciana / Diego / Mariana / Felipe
Mariana can be selected and reaches exact grant review
```

Targeted pre-operator proof on exact local bytes:

```text
static + Chromium 12 / 12 PASS
JavaScript parse PASS
```

Only the operator may re-LOCK this delta.
