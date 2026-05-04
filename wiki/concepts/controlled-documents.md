# Concept: Controlled Documents

> **Last verified:** 2026-05-01
> **Status:** Stub. Verify exact code-format string + sequence reset rules against domain code.
> **Scope:** What a Controlled Document (CD) is, code generation, profile + area binding, sequence counters.
> **Key files:**
> - `internal/modules/registry/` — CD module (verify path)
> - `frontend/apps/web/src/features/registry/RegistryListPage.tsx` — CD list
> - `frontend/apps/web/src/features/registry/RegistryCreateDialog.tsx` — CD create dialog

## What it is

A **Controlled Document (CD)** is a unique catalog slot — a code-numbered identity in the controlled-document registry. It binds:

- A **profile** (Tipo Documental).
- An **area**.
- A **sequence number** scoped to that (profile, area) pair.

The CD itself is a slot. The actual editable content lives in **document versions** that hang off the CD.

## Code format

`{profile-code}-{area-code}-{sequence-padded}`

Examples:

- `DC-RH-001` — first Descrição de Cargo in RH.
- `DC-RH-002` — second Descrição de Cargo in RH.
- `DC-QUA-001` — first Descrição de Cargo in Qualidade.
- `POP-PROD-014` — fourteenth Procedimento Operacional in Produção.

Sequence pads with leading zeros (verify width — usually 3 digits).

## Sequence rules

- One counter per `(profile, area)` pair.
- Monotonic — never reused even if a CD is archived.
- Resets: never (verify).

## Lifecycle of a CD

1. Created via "Novo Documento Controlado" — code generated, no version yet.
2. First version generated via "Gerar Documento" — clones from the profile's bound template.
3. Future revisions create additional versions on the same CD.

The CD itself doesn't have an approval state — its versions do. The CD just owns the code and the version history.

## See also

- [workflows/user-onboarding.md](../workflows/user-onboarding.md) — Step 5
- [modules/taxonomy.md](../modules/taxonomy.md)
- [modules/documents.md](../modules/documents.md)
