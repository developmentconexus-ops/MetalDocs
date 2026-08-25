# T11 — B11 Access Administration — P8 R4 Content Findings

> **Status:** P8 CONTENT FINDING / LOCAL REVISE.  
> **Artifact reviewed:** `docs/work/current/t11-b11-access-administration-p8-r4.html`  
> **Frame fidelity:** OPERATOR-APPROVED / UNAFFECTED.  
> **Upstream authority:** B11-F1 UNAFFECTED.  
> **LOCK:** NONE.

## Trigger

After the operator approved the R4 wireframe/frame fidelity, a bounded adversarial content walkthrough was run under the pinned `METHOD.md + FRONTEND-METHOD.md` and MetalDocs engineering specialization.

The review found four local P8 interaction defects. None proves missing Product/backend authority.

## Findings

### F1 — membership target copy was hard-coded

R4 `Adicionar membro` always named `Aprovadores Financeiro`, even after selecting another Group.

Protected property:

```text
confirmation target = exact currently selected User + exact currently selected Group
```

Disposition: **REVISE P8 locally**.

### F2 — membership User selection was not operable

R4 simulated adding Mariana directly instead of proving the accepted supporting-read flow:

```text
listUsers
→ choose eligible existing User
→ review consequence
→ add GroupMembership
```

Protected property: P8 must prove a material selection interaction, not only a preselected happy path.

Disposition: **REVISE P8 locally**.

### F3 — remove membership lacked deliberate confirmation

R4 removed a member immediately from the row action.

Protected property:

```text
security-bearing membership removal
→ exact User + Group confirmation
→ bounded consequence copy
→ deliberate remove
```

Disposition: **REVISE P8 locally**.

### F4 — grant lost lens context

R4 exposed only the generic top-level `Conceder acesso` flow. The operator-approved P7 R2 requires contextual grant entry from Area and Group lenses:

```text
Area lens  → preselect real scope
Group lens → preselect exact Group subject
```

The remaining dimensions stay deliberate and reviewable.

Protected property: contextual administration must reduce accidental wrong-target/wrong-scope recomposition without creating new Authorization authority.

Disposition: **REVISE P8 locally**.

## Root cause

R4 correctly restored the LOCKED wireframe grammar but over-compressed several material B11 interactions while doing so. The frame correction accidentally removed interaction detail already required by the accepted P7 R2.

## Global-Maximum decision

```text
CURRENT STRUCTURE CONFIRMED
+ local P8 interaction correction
```

No new API, Role, Permission, effective-access engine, search capability, or semantic owner is justified.

## R5 correction boundary

R5 may change only the affected B11 content interactions:

```text
real paginated User picker for add membership
dynamic User + Group review copy
remove-membership confirmation
contextual Area/Group grant entry with preselection
```

The operator-approved R4 frame remains unchanged.
