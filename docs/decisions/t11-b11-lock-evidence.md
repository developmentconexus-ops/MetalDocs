---
id: t11-b11-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for the exact operator-LOCKED B11 Access Administration frontend Evidence preserved outside the merge candidate.
---

# T11 B11 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR.  
> **Scope:** B11 Access Administration operator-LOCKED P8 + P9/P10 proof Evidence.  
> **T11:** remains OPEN.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Preserved Git identity

```text
repository     developmentconexus-ops/MetalDocs
evidence ref   evidence/t11-b11-locks-20260825
exact commit   469a753904041e7800400dc1074510456aa50df8
exact tree     c4f04b75c3676dcde00caa07279824b3c653c7f3
```

The Evidence ref must remain reachable while T11/P11/P13/P14 still depends on the B11 LOCK artifact. The exact commit is the authoritative locator; the ref name is a human convenience.

## 2. Canonical B11 P8 LOCK artifact

```text
path on exact Evidence commit
docs/work/current/t11-b11-access-administration-p8-r5.html

Git blob
96094773435a88c357e308779639415d9853b327
```

This blob is the canonical B11 functional low-fidelity LOCK Evidence. Later P11 assembly must consume this exact artifact identity, not reconstruct B11 from prose or an earlier R1–R4 candidate.

## 3. Post-LOCK proof Evidence

The same exact Evidence commit preserves:

```text
operator LOCK
docs/work/current/t11-b11-operator-lock.md

P9 Screen Contract
docs/work/current/t11-b11-screen-contract.md

P10 Pattern Consolidation
docs/work/current/t11-b11-pattern-consolidation.md
```

It also preserves the prior R1–R4 learning/finding artifacts and the P6/P7/P8 planning path. Those are historical Evidence; they do not supersede the R5 LOCK.

## 4. Durable semantic authority

B11 durable meaning remains owned by current Product/architecture authority, especially:

```text
docs/architecture/authorization-and-audit.md
docs/decisions/access-assignment-read.md
docs/architecture/wire-contract.md
docs/architecture/frontend.md
docs/decisions/api-operation-census.md
```

This locator is Evidence provenance, not a second Authorization/Product authority.

## 5. Locked protected structure summary

The exact R5 artifact protects:

```text
/admin/access
→ Por Área / Grupos / Funções

Area lens
  Area-specific grants separate from Company-wide grants

Group lens
  one Group may hold different Roles across Company and multiple Areas
  Group access footprint visible before membership mutation
  no Group.area_id

Role lens
  fixed RoleView meaning is read-only

grant
  contextual Area/Group preselection where applicable
  explicit Subject × Role × Scope review
  same-key ambiguous retry

membership
  real existing-User selection
  exact User + Group consequence review for add/remove

boundary
  no browser effective-access engine
  no global matrix/search invention
  no custom Role/Permission editor
```

## 6. Retrieval law

When FP2/P11 opens:

```text
read current Product/architecture authority
→ read current roadmap
→ use this locator only for exact retained B11 LOCK identity
→ fetch R5 from exact Evidence commit/blob
→ assemble disposable P11 Evidence
→ preserve B11 protected semantics
→ reopen only on material integration Evidence
```

Do not edit the LOCKED blob in place. Any material correction requires the normal smallest-owner reopen and a newly operator-LOCKED artifact identity.

## 7. Retirement trigger

Keep this locator and Evidence ref while any accepted frontend conformance/reconstruction proof depends on B11 R5. Repository cleanup preference alone is not a retirement trigger.
