---
id: t11-b11-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for the exact operator-LOCKED B11 Access Administration frontend Evidence preserved outside the merge candidate.
---

# T11 B11 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR / R6 CURRENT.  
> **Scope:** B11 Access Administration operator-LOCKED P8 + P9/P10 proof Evidence.  
> **T11:** remains OPEN.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Current preserved Git identity

```text
repository     developmentconexus-ops/MetalDocs
evidence ref   evidence/t11-b11-r6-locks-20260825
exact commit   6dbcec41a43dc2a74629351e22b748188e5c6dc4
exact tree     c5054688c68068457a6c46add198c1797cddec0a
```

The Evidence ref must remain reachable while T11/P11/P13/P14 still depends on B11 LOCK reconstruction. The exact commit is the canonical locator; the ref name is a human convenience.

## 2. Canonical B11 P8 LOCK artifact

```text
path on exact Evidence commit
docs/work/current/t11-b11-access-administration-p8-r6.html

Git blob
26e8905c5c5012aba59280b1001f62529ed4dfd0
```

R6 is now the canonical complete B11 functional low-fidelity LOCK Evidence because it contains the previously accepted R5 structure plus the operator re-locked pagination corrections.

Later P11 assembly must consume this exact R6 artifact identity, not reconstruct B11 from prose, R5, or earlier R1–R4 candidates.

## 3. R6 re-LOCK / P9 proof Evidence

The same exact R6 Evidence commit preserves:

```text
operator partial re-LOCK
docs/work/current/t11-b11-p8-r6-operator-relock.md

review finding / bounded reopen basis
docs/work/current/t11-b11-p8-r6-review-finding.md

P9 pagination delta
docs/work/current/t11-b11-screen-contract-r6-delta.md
```

The P9 R6 delta closes the five reopened collection-selection surfaces:

```text
R6-01 selected Group member pagination          op27
R6-02 add-member User pagination                op6
R6-03 grant User pagination                     op6
R6-04 grant Group pagination                    op22
R6-05 grant Area pagination                     op16
```

All five were operator re-LOCKED. The delta P9 verdict is PASS.

## 4. Prior R5 Evidence remains preserved

The earlier Evidence checkpoint remains reachable:

```text
evidence ref   evidence/t11-b11-locks-20260825
exact commit   469a753904041e7800400dc1074510456aa50df8
exact tree     c4f04b75c3676dcde00caa07279824b3c653c7f3
R5 Git blob    96094773435a88c357e308779639415d9853b327
```

R5 remains valid historical Evidence for the B11 learning path and for semantics not falsified by the PR #173 pagination finding. It is **not** the current complete reconstruction artifact because its Group-member and grant identity-picker pagination proof was insufficient.

The original R5 Evidence commit also preserves the original operator LOCK, full P9 Screen Contract, P10 Pattern Consolidation, and the prior R1–R4 learning/finding artifacts.

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

The exact R6 artifact protects:

```text
/admin/access
→ Por Área / Grupos / Funções

Area lens
  Area-specific grants separate from Company-wide grants

Group lens
  one Group may hold different Roles across Company and multiple Areas
  Group access footprint visible before membership mutation
  Group members traverse op27 pagination visibly
  no Group.area_id

Role lens
  fixed RoleView meaning is read-only

grant
  contextual Area/Group preselection where applicable
  paginated User / Group / Area identity selection
  explicit Subject × Role × Scope review
  same-key ambiguous retry

membership
  paginated existing-User selection
  exact User + Group consequence review for add/remove

continuation
  supporting read cursors remain opaque
  visible page survives failed continuation
  loaded page is never presented as complete when more may exist

boundary
  no browser effective-access engine
  no hidden all-page crawl
  no global matrix/search invention
  no custom Role/Permission editor
```

## 7. Proof summary

R6 targeted verification on the exact canonical blob:

```text
structural verifier          12 / 12 PASS
Chromium behavior            23 / 23 PASS
JavaScript parse             PASS
operator partial re-LOCK     APPROVED
P9 R6 pagination delta       PASS
operation 90+ consumed       0
```

The original R5 proof and P10 remain valid for unaffected behavior.

## 8. Retrieval law

When FP2/P11 opens:

```text
read current Product/architecture authority
→ read current roadmap
→ use this locator only for exact retained B11 LOCK identity
→ fetch R6 from exact R6 Evidence commit/blob
→ use prior R5 Evidence only when historical comparison is materially useful
→ assemble disposable P11 Evidence
→ preserve B11 protected semantics
→ reopen only on material integration Evidence
```

Do not edit the LOCKED blob in place. Any material correction requires the normal smallest-owner reopen and a newly operator-LOCKED artifact identity.

## 9. Retirement trigger

Keep this locator and the current R6 Evidence ref while any accepted frontend conformance/reconstruction proof depends on B11. Repository cleanup preference alone is not a retirement trigger.
