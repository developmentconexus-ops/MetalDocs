# T11 — B11 Access Administration — P7 R2 Rebaseline

> **Status:** CANDIDATE / DIRECTION OPERATOR-APPROVED / WRITTEN SPEC AWAITING OPERATOR REVIEW.  
> **Block:** B11 — Access Administration.  
> **Route:** `/admin/access`.  
> **Upstream correction:** `../../decisions/access-assignment-read.md` — OPERATOR-RATIFIED.  
> **P8 R1 disposition:** REVISE / UPSTREAM FINDING.  
> **Implementation:** BLOCKED.  
> **Artifact class:** temporary frontend-planning Evidence under `docs/work/**`; not durable Product/architecture authority.

## 1. Rebaseline trigger

P8 R1 proved that `Memberships / Role grants` made mutation possible but did not make the accepted access model understandable enough for administration.

The operator needs to navigate these questions directly:

```text
For this Area, who has which Role?
For this Group, where does it have access and with which Role?
For this fixed Role, what does it mean and where is it assigned?
```

B11-F1 therefore refines operation 31 so those questions can be answered through canonical filtered RoleAssignments without client-side fake completeness.

## 2. Protected semantic model

The UX must preserve:

```text
RoleAssignment
= Subject (User | Group)
× Role (fixed product Role)
× Scope (Company | Area)
```

A Group may have different Roles in different Areas and may also have Company-scoped grants.

Therefore:

```text
NO Group.area_id
NO assumption that Group belongs to one Area
NO editable Role/Permission matrix
NO per-User effective-access calculation in frontend
```

Company-scoped grants apply across Areas but remain Company-scoped grants. The UI must never relabel them as Area grants.

## 3. P7 R2 credible alternatives

### A — Preserve R1: Memberships / Role grants

```text
Access
├── Memberships
└── Role grants
```

**Strength:** maps directly to mutation operation families.

**Failure:** operator Evidence proved the navigation is mechanism-first. It makes the administrator reconstruct `Group → scopes → Roles` and `Area → subjects → Roles` mentally from a generic assignment ledger.

**Disposition:** REJECTED by P8 R1 Evidence.

### B — One global access matrix

```text
Rows    Users + Groups
Columns Areas + Company
Cells   Roles
```

**Strength:** appears to offer whole-company visibility at once.

**Failure:** the accepted model is many-to-many and potentially unbounded. One subject may hold multiple additive Roles in one scope. Building a complete matrix would require exhaustive cross-product traversal, client aggregation across paginated collections, or a new matrix API not proven necessary.

**Disposition:** REJECTED. It would manufacture a new query product and accidental complexity.

### C — Navigable access lenses

```text
/admin/access

Acessos
├── Por Área
├── Grupos
└── Funções
```

Each lens asks the server for a bounded canonical slice of the same RoleAssignment truth. No lens owns new Authorization state.

**Disposition:** LEADING CANDIDATE / OPERATOR-APPROVED DIRECTION.

## 4. Stable route and local IA

Stable Product route remains exactly:

```text
/admin/access
```

No new stable route family is required at P7. Local tabs/URL state may later be made addressable without creating new Product route semantics.

Local navigation:

```text
Acessos
├── Por Área
├── Grupos
└── Funções
```

The accepted global shell/chrome remains inherited from B01/B01N.

## 5. Lens 1 — Por Área

### 5.1 Area selection

Use accepted Organization supporting reads:

```text
listAreas
```

The Area list preserves lifecycle visibility. A retired Area remains historical/current identity where disclosed; B11 does not silently erase its grants.

The UI also offers a presentation-only lens:

```text
Toda a empresa
```

This means canonical `COMPANY` scope. It is not a synthetic Area and has no `area_id`.

### 5.2 Selected Area A

Two separately labeled canonical regions are required.

#### Acessos específicos desta Área

```text
listRoleAssignments
  ?scope_kind=area
  &area_id=A
```

Rows show:

```text
Subject type   User | Group
Subject        human-recognizable User/Group
Role           fixed Role
Scope          Área A
Action         revoke exact assignment when authorized
```

#### Acessos de toda a empresa

```text
listRoleAssignments
  ?scope_kind=company
```

Copy must explain:

> Estes grants têm escopo de toda a empresa e também se aplicam nesta Área. Eles continuam sendo grants Company-wide.

They render separately from Area-scoped assignments. The frontend does not merge, clone or rewrite them as Area records.

### 5.3 Toda a empresa

When the Company lens is selected:

```text
listRoleAssignments?scope_kind=company
```

Only Company-scoped assignments are shown.

### 5.4 Grant from Area lens

`Conceder acesso` from selected Area preselects the real Area scope but still composes the existing command:

```text
Subject  User | Group
Role     one Role allowed for Area scope
Scope    selected Area
Review
Grant
```

The server remains final scope/role authority.

For `Toda a empresa`, Company scope is preselected and only Roles whose `allowed_scope_kinds` includes Company are selectable.

## 6. Lens 2 — Grupos

### 6.1 Group selection

Use accepted Organization supporting read:

```text
listGroups
```

Selected Group identity remains Organization-owned.

### 6.2 Group detail structure

```text
Grupo: <name>

Resumo
  identity only

Acessos do grupo
  canonical RoleAssignments across Company and Areas

Membros
  current GroupMemberships
```

The UX may place `Acessos` before `Membros` because the security consequence of membership depends on what the Group is granted.

### 6.3 Acessos do grupo

```text
listRoleAssignments?group_id=G
```

Each row exposes:

```text
Scope
  Toda a empresa
  OR Area code + name

Role
  fixed Role label/code

Role meaning
  server-returned RoleView permissions

Action
  revoke exact assignment_id
```

Example truthful footprint:

```text
Aprovadores Financeiro

Financeiro        Aprovador
Comercial         Visualizador
Toda a empresa    Visualizador
```

This display is a projection of three distinct canonical RoleAssignments. It does not mean the Group belongs to Financeiro, Comercial or one organizational Area.

### 6.4 Grant from Group lens

`Conceder acesso` preselects the selected Group as Subject, then asks for:

```text
Role
Scope = Toda a empresa | exact Area
Review
Grant
```

A Group may receive another Role in a different Area or another additive Role in a scope if accepted server truth permits it. The frontend does not enforce uniqueness assumptions absent authority.

### 6.5 Members

```text
listGroupMembers
addGroupMember
removeGroupMember
```

Before membership mutation, the Group access footprint remains visible in the same detail context.

Confirmation copy must be bounded and truthful:

```text
Adicionar <User> a <Group>?

A pessoa passará a participar deste grupo e poderá receber acesso por meio
dos RoleAssignments atuais e futuros do grupo. Revise os acessos do grupo
nesta tela antes de confirmar.

Cancelar | Adicionar ao grupo
```

Removal:

```text
Remover <User> de <Group>?

Os acessos derivados deste grupo deixarão de se aplicar após a remoção.
A pessoa ainda pode manter acesso por grants diretos ou outros grupos.

Cancelar | Remover do grupo
```

The UI never claims a complete resulting per-User permission set.

## 7. Lens 3 — Funções

This lens explains the fixed Product roles. It is read-only.

Primary read:

```text
listRoles
```

Each Role card/detail shows:

```text
friendly Role label
canonical RoleCode
where it may be assigned
server-returned PermissionCode bundle
human presentation of those permissions
```

Example:

```text
Aprovador
Escopos: Empresa ou Área

Pode:
- ver documentos vigentes
- atuar em governança
```

Presentation labels do not change the server-returned permission bundle.

There is no custom Role/Permission editor.

### Optional assignments for selected Role

The Role detail may show canonical assignments using:

```text
listRoleAssignments?role=<RoleCode>
```

This is a filtered view of existing assignments, not a Role-owned membership list.

## 8. Grant composer — shared interaction

Grant creation is reachable contextually from Area and Group lenses and may also be exposed as a general action.

The semantic command remains:

```text
Subject × Role × Scope
```

Review repeats all three dimensions before mutation:

```text
Quem      <User or Group>
Função    <Role>
Onde      Toda a empresa | <Area>

Esta concessão é aditiva.
Grants existentes não são alterados.
```

Role meaning is inspectable before confirmation.

There is still no `Edit grant`. Changing a grant remains:

```text
revoke exact old RoleAssignment
+
new deliberate grant
```

as two explicit security decisions.

## 9. Revoke interaction

Revoke always targets one exact `assignment_id`.

Confirmation:

```text
Revogar este acesso?

Quem      <User or Group>
Função    <Role>
Onde      <actual canonical scope>

Somente este RoleAssignment será removido.
Outros grants diretos ou via grupos podem continuar válidos.
```

The wording never promises removal of all effective access.

## 10. Pagination / completeness behavior

Every filtered RoleAssignment lens remains truly paginated.

Forbidden:

```text
client loads one page then labels it as complete
client crawls all pages to build a hidden global matrix
page-local search presented as global search
```

When `has_more=true`, the UI explicitly indicates more canonical grants exist and provides continuation.

Group access footprint and Area views may therefore span pages. P8 R2 must test whether this remains operationally sufficient. If not, another read precision must be proven rather than fabricated.

## 11. What P7 R2 deliberately does not solve

Still absent/unproven:

```text
per-User effective access explanation
"why can User X do Y?" troubleshooter
all-access global matrix
search for User / Group / Area
bulk grants or memberships
custom Roles / Permissions
nested or dynamic Groups
Group organizational owner Area
access certification / periodic reviews
```

B11-F1 solves canonical assignment inspection, not every enterprise IAM problem.

## 12. Failure and concurrency behavior carried forward

Existing P7 R1 behavior remains where unaffected:

```text
403 permission.denied
404 selected Organization identity absent/non-disclosable through its owning read
409 state.conflict on createRoleAssignment
422 validation / Idempotency-Key reuse
ambiguous createRoleAssignment transport outcome -> same logical command / same key retry
membership mutation vs offboarding -> server serialization; browser never predicts winner
```

Filtered op31 invalid query combinations return `400 request.invalid` according to B11-F1.

## 13. Accessibility / responsive structure

P8 R2 must prove:

```text
local Acessos navigation is keyboard-operable
Area/Group/Role selection has visible focus and semantic labels
assignment rows retain Subject / Role / Scope meaning without color dependence
Company-wide vs Area-specific regions remain distinguishable in reading order
confirmations have deterministic focus entry/return
status/error changes are announced
```

Narrow viewport:

```text
Por Área / Grupos / Funções local nav reflows without semantic loss
Area and Group detail regions stack vertically
RoleAssignment rows become labeled key/value blocks rather than unreadable wide tables
Company-wide section remains visibly separate from Area-specific section
Grant review remains before the mutation action
```

## 14. P8 R2 falsification targets

### B11-R2-A — Group footprint comprehension

```text
A Group with different Roles in multiple Areas plus a Company Role is immediately understandable without implying Group.area_id.
```

### B11-R2-B — Area configuration comprehension

```text
An Area administrator can understand Area-scoped grants and Company-wide grants that also apply, without those grants being conflated.
```

### B11-R2-C — Membership consequence

```text
Seeing the Group's canonical access footprint in the same context plus bounded consequence copy is sufficient to make add/remove membership decisions safely without per-User effective-access calculation.
```

### B11-R2-D — Remaining effective-access gap

```text
Administrators can perform Launch access configuration safely without a separate per-User effective-access troubleshooter.
```

If D fails, it is a new material upstream Finding. B11-F1 must not be stretched silently into an effective-access engine.

### B11-R2-E — Filtered pagination sufficiency

```text
Server-filtered paginated slices are operationally sufficient at Launch scale without global matrix/search capability.
```

## 15. Global Maximum decision

Current accepted Evidence supports:

```text
CURRENT STRUCTURE CONFIRMED AFTER BOUNDED REBASELINE

/admin/access

Acessos
├── Por Área
├── Grupos
└── Funções
```

Why this is the current Global Maximum:

- it mirrors the human questions proven by operator Evidence instead of endpoint families;
- it preserves canonical Subject × Role × Scope truth;
- it allows one Group to hold different Roles across multiple Areas without inventing organizational ownership;
- it makes Company-wide grants visible where they matter while preserving their real scope;
- it uses one refined existing read operation rather than a matrix/search/effective-access platform;
- it keeps richer IAM capabilities falsifiable and deferred instead of importing enterprise complexity prematurely.

## 16. P7 R2 exit gate

```text
B11-F1 upstream authority         OPERATOR-RATIFIED / DURABLE
P8 R1                             REVISE / UPSTREAM FINDING
P7 R2 leading candidate           Por Área / Grupos / Funções
direction disposition             OPERATOR-APPROVED
written P7 R2 disposition         AWAITING OPERATOR REVIEW
P8 R2                             NOT STARTED
LOCK                              NONE
```

P8 R2 may begin only after explicit operator approval of this exact written P7 R2.
