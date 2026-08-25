# T11 — B11 Access Administration — P9 R6 pagination delta

> **Status:** COMPLETE / OPERATOR PARTIAL RE-LOCK / PASS.  
> **P8 delta:** `docs/work/current/t11-b11-access-administration-p8-r6.html`  
> **P8 R6 blob:** `26e8905c5c5012aba59280b1001f62529ed4dfd0`  
> **Finding:** `docs/work/current/t11-b11-p8-r6-review-finding.md`  
> **Scope:** only the paginated collection/selection surfaces falsified by PR #173 review.

## 1. Preserved contract

The following R5/P9 decisions remain protected and were not redesigned by R6:

```text
/admin/access route
Por Área / Grupos / Funções IA
Group multi-scope access footprint
Area-specific vs Company-wide separation
fixed Role meaning from RoleView
membership consequence / add / remove semantics
Subject × Role × Scope final grant review
exact RoleAssignment revoke
ambiguous create retry with same Idempotency-Key
no effective-access engine
no Group.area_id
no custom Role / Permission editor
no operation 90+
```

## 2. Re-locked material surfaces

| ID | Surface | Accepted read truth | R6 interaction proof | Client state | Failure rule | Forbidden shortcut | Status |
|---|---|---|---|---|---|---|---|
| R6-01 | Selected Group members | op27 `listGroupMembers` / `GroupMemberPage` | visible Previous/Next traversal; later-page member can be inspected and removed | current page + server cursor/history presentation state | continuation failure preserves loaded page and does not call it complete | hidden all-page crawl; first page labeled full membership | READY |
| R6-02 | Add-member User picker | op6 `listUsers` / `UserPage` | visible Previous/Next traversal; later-page enabled User can be selected and reviewed | current page + selected `user_id` | continuation failure preserves loaded candidates/draft | preload all Users; frontend-owned User directory | READY |
| R6-03 | Grant User subject picker | op6 `listUsers` / `UserPage` | paged list replaces complete `<select>`; later-page User can reach final grant review | current page + selected `user_id` | continuation failure preserves page and existing grant draft | preload/crawl all Users; invent search | READY |
| R6-04 | Grant Group subject picker | op22 `listGroups` / `GroupPage` | paged list; later-page Group can reach final grant review | current page + selected `group_id` | continuation failure preserves page/draft | preload/crawl all Groups; Authorization owning Group identity | READY |
| R6-05 | Grant Area scope picker | op16 `listAreas` / `AreaPage` | paged list; later-page Area can reach final grant review | current page + selected `area_id` | continuation failure preserves page/draft | preload/crawl all Areas; synthetic global Area search | READY |

## 3. Cursor / continuation law

R6 models the existing paginated read law rather than inventing offset/total semantics.

Presentation labels are intentionally bounded:

```text
Página N · há mais
Página N · fim da lista
```

The frontend does not require or infer a global total count.

Implementation realization must keep server continuation tokens opaque. A browser may retain enough navigation history to offer a Previous affordance, but it does not reconstruct, mutate or interpret cursor meaning.

For every affected read:

```text
successful continuation
  → replace the visible collection window with the returned page
  → preserve exact selected identity only when still meaningful

failed continuation
  → keep the already loaded page authoritative
  → expose retry/recovery state
  → do not clear the page
  → do not relabel loaded rows as the complete collection
```

## 4. Contextual grant proof

R6 preserves the R5 contextual entry law:

```text
Group lens
  → preselect exact current group_id
  → open the Group picker at the page containing that Group
  → Role + Scope remain deliberate

Area lens
  → preselect exact current area_id (or Company scope)
  → open the Area picker at the page containing that Area
  → Subject + Role remain deliberate
```

Preselection is navigation/form draft only. It never becomes Authorization authority.

## 5. Bidirectional trace delta

### Product/backend → frontend

```text
op27 GroupMemberPage
  → selected Group member pager

op6 UserPage
  → add-member pager
  → grant User-subject pager

op22 GroupPage
  → grant Group-subject pager

op16 AreaPage
  → grant Area-scope pager
```

### Frontend → Product/backend

```text
member Previous/Next
  → op27 cursor continuation

add-member User Previous/Next
  → op6 cursor continuation

grant User Previous/Next
  → op6 cursor continuation

grant Group Previous/Next
  → op22 cursor continuation

grant Area Previous/Next
  → op16 cursor continuation
```

No new operation or semantic owner is introduced.

## 6. Proof and operator disposition

Exact R6 Git blob:

```text
26e8905c5c5012aba59280b1001f62529ed4dfd0
```

Targeted proof on the exact byte-identical candidate:

```text
structural R6 verifier    12 / 12 PASS
Chromium behavior         23 / 23 PASS
JavaScript parse          PASS
```

The browser probes include:

```text
Group member page 2 contains Sofia Barros
op27 continuation failure preserves the current page
Grant User page 3 selects/reviews Mariana Costa
Grant Group page 2 selects/reviews Segurança Operacional
Grant Area page 3 selects/reviews Compras
op16 continuation failure preserves the current page
contextual Group grant still preselects Aprovadores Financeiro
contextual Area grant still preselects Comercial
add-member User picker reaches/reviews Mariana on a later page
mobile structure remains operable without horizontal body overflow
```

GitHub remote R6 is byte-identical to that verified candidate via blob `26e8905c5c5012aba59280b1001f62529ed4dfd0`.

**Operator disposition:** APPROVED / PARTIAL RE-LOCK.

The operator explicitly approved the R6 bounded pagination delta after being asked to judge only those reopened surfaces. Therefore R6-01..R6-05 are READY and this P9 delta is PASS.

## 7. Closure

```text
reopened P8 surfaces                 5 / 5 OPERATOR RE-LOCKED
P9 pagination-delta controls         5 / 5 READY
supporting reads                     op6 / op16 / op22 / op27 BOUND WITH VISIBLE CONTINUATION
new application operations           0
operation 90+ consumed               0
hidden collection crawls             0
unresolved P9 R6 findings            0
```

**P9 R6 verdict:** PASS.

The original R5/P9 proof remains valid for all unaffected semantics. R6 supersedes R5 only for the five pagination/identity-picker surfaces listed above. Preserve exact amended Evidence before removing temporary `docs/work/**` from the merge candidate.
