---
id: t11-b11-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for the exact operator-LOCKED B11 Access Administration frontend Evidence preserved outside the merge candidate.
---

# T11 B11 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR / R6 BASE + R7 AMENDMENT CURRENT.  
> **Scope:** B11 Access Administration operator-LOCKED P8 + P9/P10 proof Evidence.  
> **T11:** remains OPEN.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Current reconstruction package

B11 current frontend reconstruction is intentionally a two-part exact package:

```text
R6 complete base Evidence
  ref      evidence/t11-b11-r6-locks-20260825
  commit   6dbcec41a43dc2a74629351e22b748188e5c6dc4
  tree     c5054688c68068457a6c46add198c1797cddec0a
  full P8  docs/work/current/t11-b11-access-administration-p8-r6.html
  blob     26e8905c5c5012aba59280b1001f62529ed4dfd0

R7 grant-User-picker amendment Evidence
  ref      evidence/t11-b11-r7-amendment-20260825
  commit   5c3b407c1bc0e789da823570a27c33e5f8f777c3
  tree     077b25ffb9e5460f563ed84f7eedd4ed3a01d52f
  delta    docs/work/current/t11-b11-grant-user-picker-p8-r7.html
  blob     3e9130fd7b9e5b6b414b5c8e96faf6c6644cb4df
```

R6 remains the complete B11 low-fidelity base. R7 supersedes only the grant User picker page-fidelity behavior. Neither artifact may be silently reconstructed from prose.

## 2. R7 amendment law

R7 exists because PR #173 review proved one material defect in R6:

```text
R6 grant User picker
  filtered ENABLED Users before pagination

op6 authority
  raw UserPage
  PAGED
  user_id ASC
  no state filter
```

Current protected behavior is therefore:

```text
raw op6 UserPage boundary
→ render every returned User
→ ENABLED selectable
→ DISABLED visible but unavailable
→ opaque cursor traversal
→ failed continuation preserves loaded page/draft
→ no pre-pagination client state filter
→ no hidden all-page crawl
→ no invented op6 filter/search
```

R7 exact proof:

```text
static + Chromium     12 / 12 PASS
JavaScript parse      PASS
operator re-LOCK      APPROVED
P9 R7 delta           READY / PASS
```

R7 supersedes only the grant User picker row previously named R6-03. All other R6 re-LOCKed pagination surfaces remain current.

## 3. R6 base proof remains protected

The R6 Evidence commit preserves:

```text
operator partial re-LOCK
docs/work/current/t11-b11-p8-r6-operator-relock.md

review finding / bounded reopen basis
docs/work/current/t11-b11-p8-r6-review-finding.md

P9 pagination delta
docs/work/current/t11-b11-screen-contract-r6-delta.md
```

R6 continues to own the accepted full B11 structure, including:

```text
/admin/access
Por Área / Grupos / Funções
Area-specific vs Company-wide separation
Group multi-scope footprint
Group member pagination op27
add-member User pagination op6
grant Group pagination op22
grant Area pagination op16
fixed Role meaning
membership consequence
contextual Area/Group grant entry
Subject × Role × Scope review
exact revoke
same-key ambiguous retry
```

R7 changes none of those semantics.

## 4. Prior R5 Evidence remains historical

```text
evidence ref   evidence/t11-b11-locks-20260825
exact commit   469a753904041e7800400dc1074510456aa50df8
exact tree     c4f04b75c3676dcde00caa07279824b3c653c7f3
R5 Git blob    96094773435a88c357e308779639415d9853b327
```

R5 remains useful historical Evidence for the learning path and unaffected early LOCK semantics. It is not the current complete reconstruction package.

## 5. Durable semantic authority

B11 durable meaning remains owned by current Product/architecture authority, especially:

```text
docs/architecture/authorization-and-audit.md
docs/decisions/access-assignment-read.md
docs/architecture/wire-contract.md
docs/architecture/frontend.md
docs/decisions/api-operation-census.md
```

This locator is Evidence provenance, not a second Authorization/Product authority.

## 6. Current protected structure summary

```text
/admin/access
→ Por Área / Grupos / Funções

Area lens
  Area-specific grants separate from Company-wide grants

Group lens
  Group may hold different Roles across Company and multiple Areas
  footprint visible before membership mutation
  member continuation is visible
  no Group.area_id

Role lens
  fixed RoleView meaning is read-only

grant
  contextual Area/Group preselection where applicable
  Group / Area pickers preserve R6 cursor law
  User picker uses raw op6 page boundaries per R7
  DISABLED User remains visible but unavailable
  explicit Subject × Role × Scope review
  same-key ambiguous retry

membership
  paginated existing-User selection
  exact User + Group consequence review

boundary
  no browser effective-access engine
  no hidden all-page crawl
  no global matrix/search invention
  no custom Role/Permission editor
  no application operation 90+
```

## 7. Retrieval law

When FP2/P11 opens:

```text
read current Product/architecture authority
→ read current roadmap
→ fetch exact R6 full base from its Evidence ref/blob
→ apply exact R7 grant-User-picker amendment from its Evidence ref/blob
→ do not reuse the superseded R6 User-picker pre-pagination filtering behavior
→ assemble disposable P11 Evidence
→ preserve all other R6 protected semantics
→ reopen only on material integration Evidence
```

Do not edit either LOCKED Evidence artifact in place. Any further material correction requires the normal smallest-owner reopen and a newly operator-LOCKED Evidence identity.

## 8. Retirement trigger

Keep the R6 base ref, R7 amendment ref, and this locator while any accepted frontend conformance/reconstruction proof depends on B11. Repository cleanup preference alone is not a retirement trigger.
