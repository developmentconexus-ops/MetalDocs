# T11 — B11 Access Administration — P8 R2 Functional Evidence Plan

> **Status:** EXECUTION PLAN / OPERATOR-AUTHORIZED.
> **Block:** B11 — Access Administration.
> **Authority:** `docs/decisions/access-assignment-read.md` + operator-approved `docs/work/current/t11-b11-access-administration-p7-r2.md`.
> **Artifact class:** temporary frontend-planning Evidence under `docs/work/**`; not Product implementation and must be absent from the eventual merge candidate/main.

## Goal

Build one browser-operable low-fidelity P8 R2 that proves or falsifies the rebaselined B11 IA:

```text
/admin/access

Acessos
├── Por Área
├── Grupos
└── Funções
```

The artifact must make canonical RoleAssignment configuration understandable without introducing `Group.area_id`, a client-side effective-access engine, a global matrix, custom Roles/Permissions, or operation 90+.

## Falsification targets

```text
B11-R2-A  Group footprint comprehension
B11-R2-B  Area configuration comprehension
B11-R2-C  Membership consequence
B11-R2-D  Remaining effective-access gap
B11-R2-E  Filtered pagination sufficiency
```

## Deterministic fixtures

The fixture set must contain at least:

```text
Group "Aprovadores Financeiro"
  approver @ Financeiro
  viewer   @ Comercial
  viewer   @ Company

Group "Equipe Comercial"
  author @ Comercial

Group "Diretoria"
  viewer @ Company

User "Mariana Costa"
  available for membership mutation

Areas
  Comercial
  Financeiro
  Jurídico
  plus enough additional Areas to force pagination

RoleAssignments
  enough per filtered slice to exercise continuation
```

Company-wide assignments must always remain visibly Company-scoped even when displayed while an Area is selected.

## Required interactions

### Por Área

- select `Toda a empresa` or one exact Area;
- show Area-scoped assignments and Company-scoped assignments in separate regions;
- paginate each canonical filtered slice independently where `has_more=true`;
- grant access with the selected scope prefilled;
- revoke one exact assignment.

### Grupos

- select a Group;
- show its canonical access footprint across Company and multiple Areas before members;
- show Role meaning from fixed RoleView data;
- paginate the filtered `group_id` slice;
- grant access with Group preselected;
- add/remove members while the access footprint remains visible;
- consequence copy must tell the administrator to review the visible group footprint without claiming complete per-User effective access.

### Funções

- inspect all six fixed Roles;
- show canonical code, allowed scope kinds, and server-returned PermissionCode bundle in human language;
- optionally show canonical filtered assignments for the selected Role;
- no edit controls for Role or Permission.

### Shared grant/revoke

- `Subject × Role × Scope → Review → Grant`;
- scope options constrained by the selected Role;
- ambiguous create outcome retries the same logical command/Idempotency-Key;
- revoke targets one exact assignment and never promises removal of all effective access.

### Failure evidence

Operable deterministic controls for at least:

```text
400 request.invalid for impossible op31 filter combination
403 permission.denied
404 notfound.resource
409 state.conflict
422 validation.failed
```

## TDD / verification sequence

1. Write a verifier that expects the P8 R2 artifact and its protected interactions.
2. Run it before the artifact exists and require RED caused by the missing artifact.
3. Build the minimal HTML/CSS/vanilla-JS artifact.
4. Run the verifier against the exact bytes until GREEN.
5. Inspect desktop and narrow-mobile rendering in a real browser DOM.
6. Correct only defects exposed by verification/inspection; rerun the full verifier after the final byte change.
7. Publish the exact verified bytes to `docs/work/current/t11-b11-access-administration-p8-r2.html` and prove remote Git blob identity.
8. Update roadmap to `P8 R2 CANDIDATE / OPERATOR REVIEW`; do not LOCK.

## Hard stops

```text
no Group.area_id
no client-computed effective permissions
no global User/Group/Area search
no hidden crawl of all pages
no global access matrix
no custom Role/Permission editor
no operation 90+
no P9/P10 before explicit operator LOCK
no Product/runtime implementation
```
