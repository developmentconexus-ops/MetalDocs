# Module: Frontend UI Primitives (`components/ui/`)

> **Last verified:** 2026-07-02 (DOC-07b - full components/ui export-surface inventory added; prior: 2026-06-01 P2 consolidation: added Failure modes section; prior: 2026-05-13)
> **Scope:** Generic, domain-agnostic UI primitives living in `frontend/apps/web/src/components/ui/`. `SelectableCard` and `useRovingRadioGroup` are documented in full below (props, behavior, consumers). The full inventory table below covers the rest of the directory at a summary level; deep prop/behavior detail for those remains inline in the feature modules that introduced them.

> **Maturity:** L2
> **Key files:**
> - `frontend/apps/web/src/components/ui/SelectableCard.tsx:17` Ã¢â‚¬â€ `forwardRef` card button; `role="radio"` + `aria-checked`; accepts external `tabIndex` + `onKeyDown` for roving-focus integration
> - `frontend/apps/web/src/components/ui/SelectableCard.module.css:1` Ã¢â‚¬â€ `.idle`, `.selected`, `.disabled` state classes; brand border + pale fill on selected
> - `frontend/apps/web/src/components/ui/useRovingRadioGroup.ts:26` Ã¢â‚¬â€ roving-tabIndex hook for ARIA radiogroup pattern; returns `groupProps` + `getItemProps(index)`

---

## Full `components/ui/` inventory (DOC-07b)

Directory listing as of 2026-07-02: 15 component/hook source files (excluding `.module.css` and `primitives.test.tsx`). Barrel `index.ts` re-exports only 8 of them — the rest must be imported by direct path.

| Primitive | File | Barrel-exported? | External consumers | Notes |
|---|---|---|---|---|
| `Icon` | `Icon.tsx` | yes | 19 files | `IconName` union, 36 named icons (`Icon.tsx:3-9`) |
| `Avatar` | `Avatar.tsx` | yes | 15 files | size + color variants |
| `CodeChip` | `CodeChip.tsx` | yes | 6 files | inline code/tag chip |
| `StatusPill` | `StatusPill.tsx` | yes | 11 files | `DocumentStatus` union drives color |
| `Stepper` | `Stepper.tsx` | yes | 2 files | `StepperStep[]` + current index, wizard step nav |
| `SelectableCard` | `SelectableCard.tsx` | yes | multiple (see full section below) | documented in full below |
| `TabBar` | `TabBar.tsx` | yes | 1 file | bespoke roving-focus logic (T-002 open — not migrated to `useRovingRadioGroup`) |
| `WorkspaceHeroHeader` | `WorkspaceHeroHeader.tsx` | yes | 1 file | page hero/search header |
| `Dialog` | `Dialog.tsx` | yes | 10 files | modal shell |
| `SearchBar` | `SearchBar.tsx` | no | 2 files | not in barrel; imported by direct path |
| `SelectMenu` | `SelectMenu.tsx` | no | 1 file (`features/documents/components/wizard/steps/StepAreaCodeVisibility/index.tsx`) + internal use by `FilterDropdown` | not in barrel |
| `DrawerShell` | `DrawerShell.tsx` | no | 1 file (`features/iam/tabs/PeopleDetailDrawer.tsx`) | not in barrel |
| `useRovingRadioGroup` | `useRovingRadioGroup.ts` | no | multiple (see full section below) | documented in full below |
| `FilterDropdown` | `FilterDropdown.tsx` | no | 0 external — only used internally by `FormFieldBox.tsx` | dead if `FormFieldBox` stays unconsumed (see next row) |
| `TextFieldBox` / `DropdownFieldBox` | `FormFieldBox.tsx` | no | 0 — no file in `src/` imports `FormFieldBox` | dead code candidate; not wired to any screen |
| `TopbarDropdown` | `TopbarDropdown.tsx` | no | 0 — no file in `src/` imports it | dead code candidate |
| `Logo` | `Logo.tsx` | yes (`index.ts` exports it) | 0 — barrel exports `Logo` but no consumer imports it | dead code candidate despite barrel export |

Four files (`FormFieldBox.tsx`, `FilterDropdown.tsx`, `TopbarDropdown.tsx`, `Logo.tsx`) have zero consumers found in `frontend/apps/web/src/` as of this pass. Reported as an observation only — no deletion performed here (deletion is a code change, out of scope for this wiki pass; needs an owner decision on whether these are pre-built-ahead-of-use or genuinely obsolete).

---

## `SelectableCard`

A generic selectable card button. Renders a `<button role="radio" aria-checked>` with design-token styles for idle/selected/disabled states.

**Added:** initial version (exact commit not recorded). **`forwardRef` added:** commit 13595cdb Ã¢â‚¬â€ required so `useRovingRadioGroup.getItemProps(index).ref` can attach to the underlying `<button>` for programmatic focus on keyboard navigation.

### Props

```typescript
// frontend/apps/web/src/components/ui/SelectableCard.tsx:5
export type SelectableCardProps = {
  selected: boolean;
  onSelect: () => void;
  children: ReactNode;
  disabled?: boolean;
  className?: string;
  title?: string;
  ariaLabel?: string;
  tabIndex?: number;
  onKeyDown?: KeyboardEventHandler<HTMLButtonElement>;
};
```

`tabIndex` and `onKeyDown` are the integration surface for roving-focus: pass `getItemProps(i).tabIndex` and `getItemProps(i).onKeyDown` from `useRovingRadioGroup` to wire keyboard nav.

### CSS states

| Class | When |
|---|---|
| `.idle` | `!selected && !disabled` Ã¢â‚¬â€ hover lifts border + background |
| `.selected` | `selected === true` Ã¢â‚¬â€ brand 2px border + pale fill + shadow-1 |
| `.disabled` | `disabled === true` Ã¢â‚¬â€ muted, cursor: not-allowed, opacity 0.7 |

Focus ring: `2px solid var(--brand)` via `:focus-visible` on `.idle` and `.selected`.

### Usage pattern

```tsx
// Paired with useRovingRadioGroup Ã¢â‚¬â€ see StepScope.tsx or StepPermissions.tsx
const { groupProps, getItemProps } = useRovingRadioGroup({ ... });

<div {...groupProps}>
  {items.map((item, i) => (
    <SelectableCard
      key={item.id}
      ref={getItemProps(i).ref}
      tabIndex={getItemProps(i).tabIndex}
      onKeyDown={getItemProps(i).onKeyDown}
      selected={selectedId === item.id}
      onSelect={() => setSelectedId(item.id)}
    >
      {item.label}
    </SelectableCard>
  ))}
</div>
```

---

## `useRovingRadioGroup`

**Added:** commit 13595cdb. Extracted from the ad-hoc roving-tabIndex pattern used in `TabBar` and `StepPermissions` into a reusable hook for ARIA radiogroup keyboard navigation.

### API

```typescript
// frontend/apps/web/src/components/ui/useRovingRadioGroup.ts:4
export type RovingRadioGroupConfig = {
  count: number;
  selectedIndex: number;
  onSelect: (newIndex: number) => void;
  orientation?: 'horizontal' | 'vertical' | 'both'; // default: 'both'
  ariaLabel: string;
};

export type RovingRadioGroupResult = {
  groupProps: { role: 'radiogroup'; 'aria-label': string };
  getItemProps: (index: number) => {
    ref: RefCallback<HTMLElement>;
    tabIndex: number;
    onKeyDown: KeyboardEventHandler<HTMLElement>;
  };
};
```

### Behavior

- `orientation: 'horizontal'` Ã¢â‚¬â€ ArrowLeft/ArrowRight + Home/End only.
- `orientation: 'vertical'` Ã¢â‚¬â€ ArrowUp/ArrowDown + Home/End only.
- `orientation: 'both'` (default) Ã¢â‚¬â€ all four arrow keys + Home/End.
- Navigation wraps around (modulo `count`).
- `tabIndex`: active index gets `0`; all others get `-1`. If `selectedIndex === -1`, first item gets `0`.
- On key press: calls `onSelect(nextIndex)` and `refs.current[nextIndex]?.focus()` (programmatic focus requires `SelectableCard` to be a `forwardRef` component).
- `event.preventDefault()` is called on handled keys to suppress scroll.

### Consumers

- `StepScope.tsx` (template wizard Step 1) Ã¢â‚¬â€ profile cards radiogroup.
- `StepPermissions.tsx` (template wizard Step 4) Ã¢â‚¬â€ mode segmented control.

`TabBar.tsx` predates this hook and manages its own roving logic inline Ã¢â‚¬â€ migration is deferred until a third consumer appears.

---

## Failure modes

| Failure | Symptom | Detection | Response |
|---|---|---|---|
| Consumer forgets `forwardRef` on a custom replacement for `SelectableCard` | `useRovingRadioGroup.getItemProps(i).ref` cannot attach; keyboard focus stays on first card | Manual: arrow keys do not move focus; React `ref` warning in console | Use `SelectableCard` as-is (already `forwardRef`); if a new primitive is needed, wrap with `forwardRef` |
| `useRovingRadioGroup({ count: 0 })` | Calls to `getItemProps(i)` past `count - 1` would attempt focus on `undefined` ref | Manual: empty radiogroup with arrow-key navigation | Caller must guard render with `if (items.length === 0) return …` |
| `orientation: 'horizontal'` paired with vertically stacked cards | Up/Down keys do nothing; users see a dead control | UX QA — keyboard sweep | Pass `orientation: 'vertical'` (or default `'both'`) to match visual layout |
| `selectedIndex === -1` with non-empty group | First item gets `tabIndex=0`; user assumes nothing is selected | Visual: no `.selected` style on any card | Expected — caller sets `selectedIndex` once user picks; consumer story documented in `StepScope.tsx` |

## Cross-refs

- [architecture/frontend-structure.md](../architecture/frontend-structure.md) Ã¢â‚¬â€ canonical rule: `components/ui/` is design-system only; domain-agnostic; no imports from `features/`
- [modules/templates.md](templates.md) Ã¢â‚¬â€ primary consumer of both primitives
- [backlog/novo-documento.md](../backlog/novo-documento.md) Ã¢â‚¬â€ deferred `SelectableCardGroup` wrapper (held until 2nd consumer of group pattern appears)

---

## Governance Links

- [frontend-primitives-tech-debt.md](frontend-primitives-tech-debt.md)
- [backlog/frontend-primitives-refactor.md](../backlog/frontend-primitives-refactor.md)


## 11. Risks & Technical Debt

- Critical: 0
- Major: 1
- Minor: 2

Refactor backlog: [../backlog/frontend-primitives-refactor.md](../backlog/frontend-primitives-refactor.md)
