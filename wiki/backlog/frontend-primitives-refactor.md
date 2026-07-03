# Refactor Backlog - frontend-primitives

**Last verified:** 2026-07-02 (DOC-07b — R-001 closed)

| id | title | debt_id | effort | impact | blocked_by | owner | status | pr |
|---|---|---|---|---|---|---|---|---|
| R-001 | Expand frontend-primitives module coverage to full exported UI surface | T-001 | M | major | - | - | closed 2026-07-02 | - |
| R-002 | Normalize roving keyboard behavior by migrating TabBar to shared hook | T-002 | S | minor | - | - | open | - |
| R-003 | Add governance rule for `components/ui` domain-agnostic boundary | T-003 | S | minor | - | - | open | - |

## R-001 closure evidence (DOC-07b)

- `wiki/modules/frontend-primitives.md` now has a full inventory table for all 15 `components/ui` primitive/hook files, with barrel-export status and consumer counts verified against `frontend/apps/web/src/`.
- Surfaced (not fixed — code change, different owner): `FormFieldBox.tsx` (`TextFieldBox`/`DropdownFieldBox`), `FilterDropdown.tsx`, `TopbarDropdown.tsx`, and `Logo.tsx` have zero consumers in the current tree.
- See `wiki/modules/frontend-primitives-tech-debt.md` T-001 (closed) for full citation.
