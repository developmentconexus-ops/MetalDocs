# T11 — B11 Access Administration — Operator LOCK

> **Status:** OPERATOR-LOCKED / P8 COMPLETE.  
> **Locked artifact:** `docs/work/current/t11-b11-access-administration-p8-r5.html`  
> **Canonical Git blob:** `96094773435a88c357e308779639415d9853b327`  
> **Method:** pinned DevelopmentConexus `METHOD.md + FRONTEND-METHOD.md`.  
> **Implementation:** BLOCKED by `docs/roadmap.md`.

## 1. Operator disposition

```text
OPERATED / REVIEWED
operator disposition received   APPROVED
final P8 disposition             LOCK
material unresolved P8 issues   0
assistant/reviewer LOCK authority none
```

The operator first approved the inherited MetalDocs functional-low-fidelity frame in R4. A bounded content challenge then found four local interaction defects. R5 corrected only those defects while preserving the approved frame. The operator subsequently approved R5, which is the canonical LOCK artifact.

## 2. Protected structure

B11 LOCK protects the following interaction structure, not final visual design:

```text
stable route /admin/access
inherited B01/B01N shell + review grammar
local IA: Por Área / Grupos / Funções

Por Área
  Area selector + presentation-only Toda a empresa lens
  Area-scoped RoleAssignments shown separately
  Company-scoped RoleAssignments shown separately
  Company grants never relabeled as Area grants
  contextual grant entry preselects the real selected scope

Grupos
  Group identity remains Organization-owned
  access footprint shows one Group across Company + multiple Areas
  Group never acquires Group.area_id
  footprint remains visible before membership mutation
  member list is canonical GroupMembership truth
  add membership uses a real paginated existing-User picker
  exact User + Group consequence review before add
  exact User + Group consequence review before remove
  contextual grant entry preselects the exact Group subject

Funções
  six fixed Product Roles
  RoleView permissions + allowed scopes are explanatory/read-only
  no Role/Permission editor
  canonical RoleAssignment slice may be inspected by Role

grant
  Subject × Role × Scope remains explicit
  final review is mandatory before create
  selected Role constrains admissible scope kinds
  create is additive; existing grants are not silently edited
  ambiguous transport outcome retries the same logical command / Idempotency-Key

revoke
  one exact assignment_id
  consequence copy never promises removal of all effective access

collections
  real server-owned filtered pagination
  no page-local search presented as global
  no hidden crawl of all pages to manufacture completeness

boundary
  configuration inspection != complete per-User effective-access explanation
  frontend never computes Authorization
  no custom roles, permission matrix, generic IAM framework or Group single-Area ownership
```

## 3. Lock-time assumption disposition

| Assumption | Operator disposition | Reopen trigger |
|---|---|---|
| `B11-R2-A` Group footprint comprehension | VALIDATED FOR CURRENT LAUNCH P8 | operator/users cannot safely understand a Group's multiple scope-specific Roles |
| `B11-R2-B` Area configuration comprehension | VALIDATED FOR CURRENT LAUNCH P8 | Area-specific vs Company-wide configuration proves materially confusing |
| `B11-R2-C` Membership consequence | VALIDATED FOR CURRENT LAUNCH P8 | safe membership decisions require additional canonical consequence truth |
| `B11-R2-D` No per-User effective-access troubleshooter required | SUFFICIENT / VALIDATED FOR CURRENT LAUNCH P8 | a real administrator/security/compliance consumer must answer “why can User X do Y?” or equivalent effective-access investigation |
| `B11-R2-E` Filtered pagination sufficiency | VALIDATED FOR CURRENT LAUNCH P8 | real Launch scale/findability makes filtered seek pagination materially impractical |

These dispositions do not assert that richer IAM capability is never useful. They assert only that current Launch administration is sufficiently operable without importing unproven capability.

## 4. Upstream authority retained

B11-F1 remains current and sufficient:

`docs/decisions/access-assignment-read.md`

It refines existing operation 31 only. Current application-operation census remains 89. No operation 90+, effective-access engine, global matrix/search capability, custom Role/Permission editor or new semantic owner is introduced by this LOCK.

Authorization meaning remains owned by `docs/architecture/authorization-and-audit.md`.

## 5. Prior revisions

```text
R1  REVISE / UPSTREAM FINDING
R2  REVISE / VISUAL TOPOLOGY REGRESSION
R3  REVISE / LOCKED-WIREFRAME FIDELITY FAILURE
R4  FRAME OPERATOR-APPROVED; CONTENT REVISE / LOCAL FINDINGS
R5  OPERATOR-LOCKED
```

Prior artifacts remain Evidence of the learning path; they are not the LOCK baseline.

## 6. Post-LOCK gate

```text
P8   COMPLETE / OPERATOR-LOCKED
P9   NEXT — bidirectional Screen Contract trace
P10  AFTER P9 — bounded pattern consolidation
B12  NOT OPEN
FP2  NOT OPEN
T12  NOT OPEN
implementation BLOCKED
```

A P9 contradiction reopens only the smallest affected B11/frontend or upstream authority scope. P10 may consolidate only semantics proven across repeated LOCKED blocks; visual similarity is insufficient.
