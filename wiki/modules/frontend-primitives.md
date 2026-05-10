# Module: Frontend UI Primitives (`components/ui/`)

> **Last verified:** 2026-05-10
> **Scope:** Generic, domain-agnostic UI primitives living in `frontend/apps/web/src/components/ui/`. Covers `SelectableCard` and `useRovingRadioGroup`. Other primitives (WorkspaceHeroHeader, TabBar, StatusPill, Stepper) are documented inline in the feature modules that introduced them.
> **Out of scope:** Feature-specific components (`features/<x>/components/`), shared wizard primitives (`features/shared/components/wizard/`), design-system tokens (see `architecture/frontend-structure.md`).
> **Key files:**
> - `frontend/apps/web/src/components/ui/SelectableCard.tsx:17` — `forwardRef` card button; `role="radio"` + `aria-checked`; accepts external `tabIndex` + `onKeyDown` for roving-focus integration
> - `frontend/apps/web/src/components/ui/SelectableCard.module.css:1` — `.idle`, `.selected`, `.disabled` state classes; brand border + pale fill on selected
> - `frontend/apps/web/src/components/ui/useRovingRadioGroup.ts:26` — roving-tabIndex hook for ARIA radiogroup pattern; returns `groupProps` + `getItemProps(index)`

---

## `SelectableCard`

A generic selectable card button. Renders a `<button role="radio" aria-checked>` with design-token styles for idle/selected/disabled states.

**Added:** initial version (exact commit not recorded). **`forwardRef` added:** commit 13595cdb — required so `useRovingRadioGroup.getItemProps(index).ref` can attach to the underlying `<button>` for programmatic focus on keyboard navigation.

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
| `.idle` | `!selected && !disabled` — hover lifts border + background |
| `.selected` | `selected === true` — brand 2px border + pale fill + shadow-1 |
| `.disabled` | `disabled === true` — muted, cursor: not-allowed, opacity 0.7 |

Focus ring: `2px solid var(--brand)` via `:focus-visible` on `.idle` and `.selected`.

### Usage pattern

```tsx
// Paired with useRovingRadioGroup — see StepScope.tsx or StepPermissions.tsx
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

- `orientation: 'horizontal'` — ArrowLeft/ArrowRight + Home/End only.
- `orientation: 'vertical'` — ArrowUp/ArrowDown + Home/End only.
- `orientation: 'both'` (default) — all four arrow keys + Home/End.
- Navigation wraps around (modulo `count`).
- `tabIndex`: active index gets `0`; all others get `-1`. If `selectedIndex === -1`, first item gets `0`.
- On key press: calls `onSelect(nextIndex)` and `refs.current[nextIndex]?.focus()` (programmatic focus requires `SelectableCard` to be a `forwardRef` component).
- `event.preventDefault()` is called on handled keys to suppress scroll.

### Consumers

- `StepScope.tsx` (template wizard Step 1) — profile cards radiogroup.
- `StepPermissions.tsx` (template wizard Step 4) — mode segmented control.

`TabBar.tsx` predates this hook and manages its own roving logic inline — migration is deferred until a third consumer appears.

---

## Cross-refs

- [architecture/frontend-structure.md](../architecture/frontend-structure.md) — canonical rule: `components/ui/` is design-system only; domain-agnostic; no imports from `features/`
- [modules/templates-v2.md](templates-v2.md) — primary consumer of both primitives
- [backlog/novo-documento.md](../backlog/novo-documento.md) — deferred `SelectableCardGroup` wrapper (held until 2nd consumer of group pattern appears)
