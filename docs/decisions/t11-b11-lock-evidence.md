---
id: t11-b11-lock-evidence
kind: evidence-locator
owner: architecture
summary: Durable locator for the exact operator-LOCKED B11 Access Administration frontend Evidence preserved outside the merge candidate.
---

# T11 B11 LOCK Evidence Locator

> **Status:** ACTIVE DURABLE EVIDENCE LOCATOR / R6 BASE + R7 AMENDMENT PRESERVED / TWO BOUNDED REOPENS OPEN.  
> **Scope:** B11 Access Administration operator-LOCKED frontend Evidence and later bounded falsifications.  
> **T11:** remains OPEN.  
> **Implementation:** BLOCKED by `../roadmap.md`.

## 1. Preserved Evidence package

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

These exact refs remain valid Evidence. They are **not currently sufficient as a complete reconstruction package** because later PR #173 review exposed two additional R6 behavior contradictions that have not yet been operator-ratified after correction.

## 2. R7 amendment remains valid

R7 correctly supersedes only the R6 grant User picker page-fidelity behavior:

```text
raw op6 UserPage boundary
→ every returned User remains in the page
→ ENABLED selectable
→ DISABLED visible but unavailable
→ opaque cursor traversal
→ no pre-pagination state filter
→ no hidden all-page crawl
```

R7 proof remains:

```text
static + Chromium     12 / 12 PASS
JavaScript parse      PASS
operator re-LOCK      APPROVED
P9 R7 delta           READY / PASS
```

The newly opened findings do not falsify R7.

## 3. Open bounded finding A — add-member membership knowledge

Exact R6 Evidence currently implements the add-member User picker with complete fixture membership knowledge:

```text
current = Set(all members of selected Group)
→ every already-member User disabled as "já membro"
```

Current accepted read authority exposes only paginated op27 `GroupMemberPage`; there is no accepted per-User GroupMembership lookup and no complete membership projection.

Therefore the exact R6 behavior is not currently realizable without one of the forbidden shortcuts:

```text
hidden all-page op27 crawl
OR
invented membership lookup/search authority
```

Smallest correction candidate, pending operator authorization/ratification:

```text
op6 UserPage remains the picker source
→ do not claim complete GroupMembership knowledge
→ DISABLED User may still be unavailable from User state truth
→ already-member relation may be unknown before PUT
→ idempotent op28 reconciles first-add 201 vs already-exists 204
→ optionally use only membership rows already legitimately loaded as local guidance
```

No backend reopen is yet proven necessary.

## 4. Open bounded finding B — repeated grant confirmation

Exact R6 Evidence keeps the grant confirm action active after a successful create and each subsequent click appends another fixture assignment while `state.grantKey` remains the same logical key.

That contradicts the accepted idempotency law:

```text
same Idempotency-Key
+ same normalized command fingerprint
→ replay stored success
→ zero second semantic mutation
```

Smallest correction candidate, pending operator authorization/ratification:

```text
successful grant
→ command becomes terminal in the current dialog
→ close or disable repeated confirmation
OR
explicitly model same-key replay with the same success identity
→ never append a second assignment
```

No backend reopen is justified by this finding.

## 5. Preserved R6 structure outside the open findings

The later findings do not reopen:

```text
/admin/access
Por Área / Grupos / Funções
Area-specific vs Company-wide separation
Group multi-scope footprint
Group member visible pagination itself
add-member User pagination mechanics themselves
grant Group pagination
grant Area pagination
fixed Role meaning
membership consequence copy
contextual Area/Group grant entry
Subject × Role × Scope review
exact revoke
ambiguous transport same-key retry law
B11-F1 op31 precision
Authorization authority
P10 consolidation
89-operation census
```

## 6. Prior R5 Evidence remains historical

```text
evidence ref   evidence/t11-b11-locks-20260825
exact commit   469a753904041e7800400dc1074510456aa50df8
exact tree     c4f04b75c3676dcde00caa07279824b3c653c7f3
R5 Git blob    96094773435a88c357e308779639415d9853b327
```

R5 remains historical Evidence for the learning path and unaffected early LOCK semantics.

## 7. Durable semantic authority

B11 durable meaning remains owned by current Product/architecture authority, especially:

```text
docs/architecture/authorization-and-audit.md
docs/decisions/access-assignment-read.md
docs/architecture/wire-contract.md
docs/architecture/frontend.md
docs/decisions/api-operation-census.md
```

This locator records Evidence provenance and falsification state; it is not a second Authorization/Product authority.

## 8. Retrieval law while reopen is unresolved

Until the two open findings are corrected and operator re-LOCKed:

```text
R6 + R7 may be used only as preserved Evidence
→ do not treat them as a fully realizable P11 reconstruction package
→ preserve all unaffected semantics
→ exclude the falsified complete-membership picker knowledge
→ exclude the falsified repeated-success duplicate-grant behavior
```

After lawful bounded corrections, this locator must be updated with the new exact amendment Evidence identities before P11 reconstruction.

## 9. Retirement trigger

Keep R5, R6, R7 Evidence refs and this locator while accepted frontend conformance/reconstruction depends on B11. Repository cleanup preference alone is not a retirement trigger.
