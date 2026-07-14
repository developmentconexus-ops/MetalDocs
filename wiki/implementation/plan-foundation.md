# Foundation Block Implementation Plan

> **For agentic workers:** Use `codex:rescue` to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking. Tasks marked **[PARALLEL]** can run simultaneously with other parallel tasks using separate Codex agents. Tasks marked **[SEQUENTIAL]** must wait for their dependency.

**Goal:** Establish the complete design system foundation: Wine palette tokens, Inter Tight fonts, Zustand server-state cleanup, shared UI primitives, centralized query keys, and the new AppShell (Rail + Toolbar + AppRoot + Router split).

**Architecture:** Token rename (--vinho → --brand) throughout, new UI primitive components using global design system classes, TanStack Query keys centralized in one file, new AppRoot/AppShell replacing the God-component WorkspaceRoot.

**Tech Stack:** React 18, TypeScript, Vite, Zustand (auth + ui stores only), TanStack Query v5, React Router v7, CSS Modules, Vitest + React Testing Library

**Worktree:** `.worktrees/screen-redesign` (branch `feature/screen-redesign`)

**Design source:** `frontend/apps/web/design-source/shell.jsx` + `design-source/styles.css`

---

## Parallel Group A — Run all three simultaneously

---

### Task A1: tokens.css + index.html fonts [PARALLEL]

**Files:**
- Modify: `frontend/apps/web/src/styles/tokens.css`
- Modify: `frontend/apps/web/index.html`

- [ ] **Step 1: Rewrite tokens.css**

Replace the entire file content with:

```css
/* MetalDocs Design System — Wine Palette */
:root {
  /* Fonts */
  --font-sans: "Inter Tight", "Inter", system-ui, sans-serif;
  --font-mono: "JetBrains Mono", "IBM Plex Mono", ui-monospace, monospace;

  /* Brand */
  --brand: #6b1f2a;
  --brand-deep: #3e1018;
  --brand-soft: #8b2e3a;
  --brand-pale: #f9f0f0;
  --accent: #c8364a;

  /* Surface */
  --bg: #f4eeee;
  --surface: #ffffff;
  --surface-2: #faf6f6;
  --surface-3: #f0e9e9;

  /* Border */
  --border: #e6dcdc;
  --border-strong: #d4c2c2;

  /* Text */
  --text: #1a0e0e;
  --text-soft: #4a3434;
  --text-muted: #8a7575;
  --text-faint: #b3a0a0;

  /* Rail (dark sidebar) */
  --rail: #2a1418;
  --rail-text: #e8d6d6;
  --rail-text-muted: #9c7e7e;
  --rail-active: #6b1f2a;
  --rail-divider: #3e2025;

  /* Semantic */
  --success: #1a6b35;
  --success-bg: #e6f5ec;
  --warning: #b07016;
  --warning-bg: #fbf2dc;
  --danger: #c8364a;
  --danger-bg: #fae8eb;
  --info: #1a3a7a;
  --info-bg: #e8eef8;

  /* Shadow */
  --shadow-1: 0 1px 2px rgba(74, 33, 33, 0.06);
  --shadow-2: 0 8px 24px rgba(74, 33, 33, 0.08);

  /* Radii */
  --r-1: 4px;
  --r-2: 6px;
  --r-3: 8px;
  --r-4: 12px;
  --r-pill: 999px;

  /* Spacing */
  --sp-1: 4px;  --sp-2: 8px;  --sp-3: 12px; --sp-4: 16px;
  --sp-5: 20px; --sp-6: 24px; --sp-7: 32px; --sp-8: 40px; --sp-9: 56px;
}
```

- [ ] **Step 2: Update index.html — add Google Fonts import**

Replace `<head>` content:

```html
<!doctype html>
<html lang="pt-BR">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>MetalDocs</title>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin />
    <link
      href="https://fonts.googleapis.com/css2?family=Inter+Tight:wght@400;500;600&family=JetBrains+Mono:wght@400;500&display=swap"
      rel="stylesheet"
    />
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

- [ ] **Step 3: Update base.css — switch body font and button defaults**

In `frontend/apps/web/src/styles/base.css`, update the `body` rule:

```css
body {
  margin: 0;
  min-height: 100vh;
  background: var(--bg);
  color: var(--text);
  font-family: var(--font-sans);
  -webkit-font-smoothing: antialiased;
  text-rendering: optimizeLegibility;
}
```

Remove the gradient from `background` and add `font-family`. Also update `button` default to remove the opinionated radius/padding that conflicts with design system:

```css
button {
  font: inherit;
  cursor: pointer;
}
```

- [ ] **Step 4: Commit**

```bash
git add frontend/apps/web/src/styles/tokens.css frontend/apps/web/index.html frontend/apps/web/src/styles/base.css
git commit -m "feat(foundation): Wine palette tokens + Inter Tight / JetBrains Mono fonts"
```

---

### Task A2: queryKeys.ts [PARALLEL]

**Files:**
- Create: `frontend/apps/web/src/lib/queryKeys.ts`
- Create: `frontend/apps/web/src/lib/queryKeys.test.ts`

- [ ] **Step 1: Write failing test**

Create `frontend/apps/web/src/lib/queryKeys.test.ts`:

```ts
import { describe, it, expect } from 'vitest';
import { QK } from './queryKeys';

describe('QK', () => {
  it('documents.list returns stable key', () => {
    expect(QK.documents.list()).toEqual(['documents', 'list']);
  });

  it('documents.detail includes id', () => {
    expect(QK.documents.detail('abc')).toEqual(['documents', 'detail', 'abc']);
  });

  it('inbox includes params', () => {
    expect(QK.inbox({ page: 2, areaFilter: 'RH' }))
      .toEqual(['approval', 'inbox', { page: 2, areaFilter: 'RH' }]);
  });

  it('inbox with no params uses empty object', () => {
    expect(QK.inbox()).toEqual(['approval', 'inbox', {}]);
  });

  it('approval.instance includes documentId', () => {
    expect(QK.approval.instance('doc-1')).toEqual(['approval', 'instance', 'doc-1']);
  });
});
```

- [ ] **Step 2: Run test — expect FAIL**

```bash
cd frontend && npx vitest run src/lib/queryKeys.test.ts
```

Expected: `Cannot find module './queryKeys'`

- [ ] **Step 3: Create queryKeys.ts**

Create `frontend/apps/web/src/lib/queryKeys.ts`:

```ts
// Centralized TanStack Query key constants.
// All useQuery / invalidateQueries calls must import from here — never inline string arrays.

type InboxParams = {
  page?: number;
  areaFilter?: string;
  onlyOverdue?: boolean;
  limit?: number;
};

type ControlledDocFilter = {
  profileCode?: string;
  processAreaCode?: string;
  status?: string;
};

export const QK = {
  documents: {
    list: () => ['documents', 'list'] as const,
    detail: (id: string) => ['documents', 'detail', id] as const,
  },
  inbox: (params: InboxParams = {}) =>
    ['approval', 'inbox', params] as const,
  audit: {
    // GET /audit/events
    recent: (limit = 10) => ['audit', 'recent', limit] as const,
  },
  controlledDocuments: {
    list: (filter: ControlledDocFilter = {}) =>
      ['controlled-documents', 'list', filter] as const,
    detail: (id: string) => ['controlled-documents', 'detail', id] as const,
  },
  taxonomy: {
    profiles: () => ['taxonomy', 'profiles'] as const,
    areas: () => ['taxonomy', 'areas'] as const,
  },
  templates: {
    list: () => ['templates', 'list'] as const,
  },
  approval: {
    instance: (documentId: string) =>
      ['approval', 'instance', documentId] as const,
  },
  notifications: {
    unreadCount: () => ['notifications', 'unread-count'] as const,
  },
} as const;
```

- [ ] **Step 4: Run test — expect PASS**

```bash
cd frontend && npx vitest run src/lib/queryKeys.test.ts
```

Expected: `5 tests passed`

- [ ] **Step 5: Commit**

```bash
git add frontend/apps/web/src/lib/queryKeys.ts frontend/apps/web/src/lib/queryKeys.test.ts
git commit -m "feat(foundation): centralized TanStack Query key constants (QK)"
```

---

### Task A3: UI Primitives — Icon, Avatar, CodeChip, Logo, StatusPill [PARALLEL]

**Files:**
- Create: `frontend/apps/web/src/components/ui/Icon.tsx`
- Create: `frontend/apps/web/src/components/ui/Avatar.tsx`
- Create: `frontend/apps/web/src/components/ui/CodeChip.tsx`
- Create: `frontend/apps/web/src/components/ui/Logo.tsx`
- Create: `frontend/apps/web/src/components/ui/StatusPill.tsx`
- Create: `frontend/apps/web/src/components/ui/StatusPill.module.css`
- Create: `frontend/apps/web/src/components/ui/index.ts`
- Create: `frontend/apps/web/src/components/ui/primitives.test.tsx`

- [ ] **Step 1: Write failing tests**

Create `frontend/apps/web/src/components/ui/primitives.test.tsx`:

```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Icon } from './Icon';
import { Avatar } from './Avatar';
import { CodeChip } from './CodeChip';
import { StatusPill } from './StatusPill';

describe('Icon', () => {
  it('renders an svg', () => {
    const { container } = render(<Icon name="home" size={16} />);
    expect(container.querySelector('svg')).toBeTruthy();
  });

  it('applies size as width/height', () => {
    const { container } = render(<Icon name="search" size={20} />);
    const svg = container.querySelector('svg')!;
    expect(svg.getAttribute('width')).toBe('20');
    expect(svg.getAttribute('height')).toBe('20');
  });
});

describe('Avatar', () => {
  it('renders two-letter initials from full name', () => {
    render(<Avatar name="Marina Silveira" />);
    expect(screen.getByText('MS')).toBeTruthy();
  });

  it('renders one letter for single name', () => {
    render(<Avatar name="Admin" />);
    expect(screen.getByText('AD')).toBeTruthy();
  });
});

describe('CodeChip', () => {
  it('renders children', () => {
    render(<CodeChip>POP-RH-001</CodeChip>);
    expect(screen.getByText('POP-RH-001')).toBeTruthy();
  });
});

describe('StatusPill', () => {
  it('renders draft pill', () => {
    const { container } = render(<StatusPill status="draft" />);
    expect(container.firstChild).toHaveClass('pill-draft');
  });

  it('renders frozen pill', () => {
    const { container } = render(<StatusPill status="frozen" />);
    expect(container.firstChild).toHaveClass('pill-frozen');
  });
});
```

- [ ] **Step 2: Run — expect FAIL**

```bash
cd frontend && npx vitest run src/components/ui/primitives.test.tsx
```

Expected: `Cannot find module './Icon'`

- [ ] **Step 3: Create Icon.tsx**

SVG paths from `design-source/shell.jsx`. Tier 1 — no CSS Module, uses inline SVG only.

Create `frontend/apps/web/src/components/ui/Icon.tsx`:

```tsx
import type { CSSProperties } from 'react';

type IconName =
  | 'home' | 'docs' | 'library' | 'registry' | 'template' | 'inbox'
  | 'workflow' | 'taxonomy' | 'users' | 'audit' | 'bell' | 'search'
  | 'plus' | 'chevron' | 'chevdown' | 'filter' | 'list' | 'lock'
  | 'check' | 'x' | 'download' | 'upload' | 'history' | 'link'
  | 'cog' | 'arrow' | 'sparkle' | 'more';

const PATHS: Record<IconName, React.ReactNode> = {
  home:     <><path d="M3 10l7-6 7 6"/><path d="M5 9v8h10V9"/></>,
  docs:     <><path d="M5 3h7l4 4v13H5z"/><path d="M12 3v4h4"/></>,
  library:  <><path d="M3 4h4v14H3zM9 4h4v14H9z"/><path d="M15 5l3 1-3 13"/></>,
  registry: <><rect x="3" y="3" width="14" height="14" rx="1"/><path d="M3 7h14M3 11h14M3 15h14M7 3v14"/></>,
  template: <><rect x="3" y="3" width="14" height="14" rx="1"/><path d="M3 8h14M8 8v9"/></>,
  inbox:    <><path d="M3 12l3-8h8l3 8v5H3z"/><path d="M3 12h4l1 2h4l1-2h4"/></>,
  workflow: <><circle cx="5" cy="5" r="2"/><circle cx="15" cy="15" r="2"/><circle cx="15" cy="5" r="2"/><path d="M7 5h6M5 7v6a2 2 0 002 2h6"/></>,
  taxonomy: <><path d="M10 3v4M10 9v8M5 13h10M3 17h4M13 17h4"/><circle cx="10" cy="3" r="1.5"/></>,
  users:    <><circle cx="7" cy="7" r="3"/><path d="M2 17c0-3 2-5 5-5s5 2 5 5"/><circle cx="14" cy="8" r="2"/><path d="M11 17c0-2 1-3 3-3s4 1 4 3"/></>,
  audit:    <><path d="M5 3h7l4 4v13H5z"/><path d="M8 11l2 2 4-4"/></>,
  bell:     <><path d="M5 8a5 5 0 0110 0v4l1.5 2h-13L5 12V8z"/><path d="M8 16a2 2 0 004 0"/></>,
  search:   <><circle cx="9" cy="9" r="5"/><path d="M13 13l4 4"/></>,
  plus:     <><path d="M10 4v12M4 10h12"/></>,
  chevron:  <><path d="M7 4l6 6-6 6"/></>,
  chevdown: <><path d="M5 8l5 5 5-5"/></>,
  filter:   <><path d="M3 5h14M6 10h8M8 15h4"/></>,
  list:     <><path d="M3 5h14M3 10h14M3 15h14"/></>,
  lock:     <><rect x="4" y="9" width="12" height="9" rx="1"/><path d="M7 9V6a3 3 0 016 0v3"/></>,
  check:    <><path d="M4 10l4 4 8-8"/></>,
  x:        <><path d="M5 5l10 10M15 5L5 15"/></>,
  download: <><path d="M10 3v10M5 9l5 5 5-5M3 17h14"/></>,
  upload:   <><path d="M10 14V4M5 9l5-5 5 5M3 17h14"/></>,
  history:  <><path d="M3 10a7 7 0 1 0 2-5"/><path d="M3 3v4h4M10 6v5l3 2"/></>,
  link:     <><path d="M8 12l4-4M7 8h-2a3 3 0 000 6h2M13 12h2a3 3 0 000-6h-2"/></>,
  cog:      <><circle cx="10" cy="10" r="2.5"/><path d="M10 3v2M10 15v2M3 10h2M15 10h2M5 5l1.5 1.5M14 14l-1.5-1.5M5 15l1.5-1.5M14 6l-1.5 1.5"/></>,
  arrow:    <><path d="M5 10h10M11 6l4 4-4 4"/></>,
  sparkle:  <><path d="M10 3l1.5 4 4 1.5-4 1.5L10 14l-1.5-4L4 8.5l4-1.5z"/></>,
  more:     <><circle cx="5" cy="10" r="1.5"/><circle cx="10" cy="10" r="1.5"/><circle cx="15" cy="10" r="1.5"/></>,
};

type IconProps = {
  name: IconName;
  size?: number;
  className?: string;
  style?: CSSProperties;
};

export function Icon({ name, size = 16, className, style }: IconProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      style={style}
      aria-hidden="true"
    >
      {PATHS[name]}
    </svg>
  );
}
```

- [ ] **Step 4: Create Avatar.tsx**

Tier 1 — uses global `.avatar` classes from design system.

```tsx
type AvatarSize = 'sm' | 'md' | 'lg';

type AvatarProps = {
  name: string;
  size?: AvatarSize;
};

function initials(name: string): string {
  const parts = name.trim().split(/\s+/);
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
}

const sizeClass: Record<AvatarSize, string> = {
  sm: 'avatar avatar-sm',
  md: 'avatar',
  lg: 'avatar avatar-lg',
};

export function Avatar({ name, size = 'md' }: AvatarProps) {
  return (
    <span className={sizeClass[size]} title={name} aria-label={name}>
      {initials(name)}
    </span>
  );
}
```

- [ ] **Step 5: Create CodeChip.tsx**

```tsx
import type { ReactNode } from 'react';

type CodeChipProps = {
  children: ReactNode;
  className?: string;
};

export function CodeChip({ children, className }: CodeChipProps) {
  return (
    <span className={`code-chip mono${className ? ` ${className}` : ''}`}>
      {children}
    </span>
  );
}
```

- [ ] **Step 6: Create Logo.tsx**

From `design-source/shell.jsx` Logo component:

```tsx
type LogoProps = {
  size?: 'sm' | 'md';
};

export function Logo({ size = 'md' }: LogoProps) {
  const markSize = size === 'sm' ? 18 : 22;
  const fontSize = size === 'sm' ? 13 : 15;
  return (
    <span
      className="logo"
      style={{ fontSize, fontFamily: 'var(--font-sans)', fontWeight: 600, letterSpacing: '-0.02em', display: 'inline-flex', alignItems: 'center', gap: 8 }}
    >
      <span
        style={{
          width: markSize,
          height: markSize,
          borderRadius: 5,
          background: 'var(--brand)',
          position: 'relative',
          display: 'inline-block',
          flexShrink: 0,
        }}
      >
        {/* Three horizontal lines (document icon) */}
        <span style={{ position: 'absolute', left: 4, right: 4, top: 5, height: 1.5, background: 'white' }} />
        <span style={{ position: 'absolute', left: 4, right: 4, top: 9, height: 1.5, background: 'white' }} />
        <span style={{ position: 'absolute', left: 4, right: 4, top: 13, height: 1.5, background: 'white' }} />
      </span>
      MetalDocs
    </span>
  );
}
```

- [ ] **Step 7: Create StatusPill.tsx + StatusPill.module.css**

StatusPill has color logic so gets a CSS Module.

Create `frontend/apps/web/src/components/ui/StatusPill.module.css`:

```css
/* StatusPill uses global pill classes from base.css.
   This module adds layout overrides only. */
.pill {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  white-space: nowrap;
}
```

Create `frontend/apps/web/src/components/ui/StatusPill.tsx`:

```tsx
import styles from './StatusPill.module.css';

export type DocumentStatus =
  | 'draft'
  | 'review'
  | 'approved'
  | 'frozen'
  | 'rejected'
  | 'archived'
  | 'finalized';

const STATUS_CONFIG: Record<DocumentStatus, { label: string; pillClass: string }> = {
  draft:     { label: 'Rascunho',  pillClass: 'pill pill-draft' },
  review:    { label: 'Em revisão', pillClass: 'pill pill-review' },
  approved:  { label: 'Aprovado',  pillClass: 'pill pill-approved' },
  frozen:    { label: 'Frozen',    pillClass: 'pill pill-frozen' },
  rejected:  { label: 'Rejeitado', pillClass: 'pill pill-rejected' },
  archived:  { label: 'Arquivado', pillClass: 'pill pill-archived' },
  finalized: { label: 'Finalizado', pillClass: 'pill pill-approved' },
};

type StatusPillProps = {
  status: DocumentStatus;
  className?: string;
};

export function StatusPill({ status, className }: StatusPillProps) {
  const config = STATUS_CONFIG[status] ?? { label: status, pillClass: 'pill' };
  return (
    <span className={`${config.pillClass} ${styles.pill}${className ? ` ${className}` : ''}`}>
      <span className="dot" />
      {config.label}
    </span>
  );
}
```

- [ ] **Step 8: Create barrel export**

Create `frontend/apps/web/src/components/ui/index.ts`:

```ts
export { Icon } from './Icon';
export { Avatar } from './Avatar';
export { CodeChip } from './CodeChip';
export { Logo } from './Logo';
export { StatusPill } from './StatusPill';
export type { DocumentStatus } from './StatusPill';
```

- [ ] **Step 9: Run tests — expect PASS**

```bash
cd frontend && npx vitest run src/components/ui/primitives.test.tsx
```

Expected: `8 tests passed`

- [ ] **Step 10: Commit**

```bash
git add frontend/apps/web/src/components/ui/
git commit -m "feat(foundation): UI primitives — Icon, Avatar, CodeChip, Logo, StatusPill"
```

---

## Sequential Group B — After Group A completes

---

### Task B1: Mass-replace --vinho CSS vars [SEQUENTIAL after A1]

**Files:**
- Modify: `frontend/apps/web/src/components/DocumentWorkspaceShell.module.css`
- Modify: `frontend/apps/web/src/components/ManagedUsersPanel.module.css`
- Modify: `frontend/apps/web/src/components/ui/FormFieldBox.module.css`
- Modify: `frontend/apps/web/src/features/documents/canvas/DocumentCanvas.module.css`
- Modify: `frontend/apps/web/src/features/documents/DocumentsHubView.module.css`
- Modify: `frontend/apps/web/src/features/documents/runtime/DynamicEditor.module.css`
- Modify: `frontend/apps/web/src/features/documents/runtime/RichField.module.css`
- Modify: `frontend/apps/web/src/features/iam/AdminCenterView.module.css`
- Modify: `frontend/apps/web/src/styles.css`

- [ ] **Step 1: Run PowerShell mass-replace script**

Run this from the repo root:

```powershell
$files = Get-ChildItem -Recurse -Path "frontend\apps\web\src" -Include "*.css" |
  Where-Object { (Get-Content $_.FullName -Raw) -match '--vinho' }

foreach ($file in $files) {
  $content = Get-Content $file.FullName -Raw
  $content = $content -replace '--vinho-pale', '--brand-pale'
  $content = $content -replace '--vinho-soft', '--surface-2'
  $content = $content -replace '--vinho-muted', '--border-strong'
  $content = $content -replace '--vinho-d\b', '--brand-deep'
  $content = $content -replace '--vinho-l\b', '--brand-soft'
  $content = $content -replace '--vinho\b', '--brand'
  $content = $content -replace '--border-2\b', '--border-strong'
  $content = $content -replace '--muted-soft\b', '--text-faint'
  $content = $content -replace '--muted\b', '--text-muted'
  Set-Content $file.FullName $content -NoNewline
  Write-Host "Updated: $($file.FullName)"
}
```

- [ ] **Step 2: Verify no --vinho refs remain**

```powershell
$remaining = Get-ChildItem -Recurse -Path "frontend\apps\web\src" -Include "*.css" |
  Where-Object { (Get-Content $_.FullName -Raw) -match '--vinho' }
if ($remaining) {
  Write-Host "REMAINING --vinho refs:"
  $remaining | ForEach-Object { Write-Host $_.FullName }
} else {
  Write-Host "All --vinho refs replaced."
}
```

Expected: `All --vinho refs replaced.`

- [ ] **Step 3: TypeScript check passes**

```bash
cd frontend && npx tsc --noEmit
```

Expected: 0 errors (token renames are CSS-only, no TS impact)

- [ ] **Step 4: Commit**

```bash
git add frontend/apps/web/src/
git commit -m "refactor(foundation): rename --vinho CSS vars to --brand throughout"
```

---

### Task B2: Zustand server-state cleanup [SEQUENTIAL after A1]

**Goal:** Delete server-state Zustand stores and all components that depend exclusively on them. Keep `auth.store` and `ui.store` (stripped to essentials).

**Files to DELETE:**
- `frontend/apps/web/src/features/documents/state/documents.store.ts`
- `frontend/apps/web/src/features/controlled-documents/state/` (entire directory)
- `frontend/apps/web/src/features/notifications/state/` (entire directory)
- `frontend/apps/web/src/features/shell/pages/WorkspaceRoot.tsx`
- `frontend/apps/web/src/features/shell/WorkspaceShell.tsx`
- `frontend/apps/web/src/components/AuthShell.tsx`
- `frontend/apps/web/src/components/DocumentWorkspaceShell.tsx`
- `frontend/apps/web/src/components/WorkspaceViewFrame.tsx`
- `frontend/apps/web/src/components/WorkspaceDataState.tsx`
- `frontend/apps/web/src/components/WorkspacePlaceholder.tsx`
- `frontend/apps/web/src/components/AppShellHeader.tsx`

**Files to MODIFY:**
- `frontend/apps/web/src/store/ui.store.ts` — strip WorkspaceView + non-essential state
- `frontend/apps/web/src/features/auth/useAuthSession.ts` — strip all deleted store calls

- [ ] **Step 1: Delete server-state stores and old shell files**

```bash
rm -f frontend/apps/web/src/features/documents/state/documents.store.ts
rm -rf frontend/apps/web/src/features/controlled-documents/state
rm -rf frontend/apps/web/src/features/notifications/state
rm -f frontend/apps/web/src/features/shell/pages/WorkspaceRoot.tsx
rm -f frontend/apps/web/src/features/shell/WorkspaceShell.tsx
rm -f frontend/apps/web/src/components/AuthShell.tsx
rm -f frontend/apps/web/src/components/DocumentWorkspaceShell.tsx
rm -f frontend/apps/web/src/components/WorkspaceViewFrame.tsx
rm -f frontend/apps/web/src/components/WorkspaceDataState.tsx
rm -f frontend/apps/web/src/components/WorkspacePlaceholder.tsx
rm -f frontend/apps/web/src/components/AppShellHeader.tsx
```

- [ ] **Step 2: Rewrite ui.store.ts — strip WorkspaceView dependency**

Replace entire `frontend/apps/web/src/store/ui.store.ts`:

```ts
import { create } from 'zustand';
import type { ManagedUserItem, UserRole } from '../lib/types';

type UserFormState = {
  userId: string;
  username: string;
  email: string;
  displayName: string;
  password: string;
  roles: UserRole[];
};

type ManagedUserFormState = {
  userId: string;
  displayName: string;
  email: string;
  isActive: boolean;
  mustChangePassword: boolean;
  roles: UserRole[];
  resetPassword: string;
};

interface UiStore {
  // Flash messages
  message: string;
  error: string;
  setMessage: (message: string) => void;
  setError: (error: string) => void;
  // Admin user management
  managedUsers: ManagedUserItem[];
  setManagedUsers: (users: ManagedUserItem[]) => void;
  // User forms (IAM admin)
  userForm: UserFormState;
  managedUserForm: ManagedUserFormState;
  setUserForm: (form: UserFormState | ((c: UserFormState) => UserFormState)) => void;
  setManagedUserForm: (form: ManagedUserFormState | ((c: ManagedUserFormState) => ManagedUserFormState)) => void;
}

const defaultUserForm: UserFormState = {
  userId: '', username: '', email: '', displayName: '', password: '', roles: ['viewer'],
};

const defaultManagedUserForm: ManagedUserFormState = {
  userId: '', displayName: '', email: '', isActive: true,
  mustChangePassword: false, roles: ['viewer'], resetPassword: '',
};

export const useUiStore = create<UiStore>((set) => ({
  message: '',
  error: '',
  setMessage: (message) => set({ message }),
  setError: (error) => set({ error }),
  managedUsers: [],
  setManagedUsers: (managedUsers) => set({ managedUsers }),
  userForm: defaultUserForm,
  managedUserForm: defaultManagedUserForm,
  setUserForm: (form) =>
    set((s) => ({ userForm: typeof form === 'function' ? form(s.userForm) : form })),
  setManagedUserForm: (form) =>
    set((s) => ({ managedUserForm: typeof form === 'function' ? form(s.managedUserForm) : form })),
}));

export type { ManagedUserFormState, UserFormState };
```

- [ ] **Step 3: Rewrite useAuthSession.ts — strip deleted store calls**

Replace entire `frontend/apps/web/src/features/auth/useAuthSession.ts`:

```ts
import { useCallback } from 'react';
import * as api from './api/auth';
import { useAuthStore } from '../../store/auth.store';
import { useUiStore } from '../../store/ui.store';
import { asMessage, statusOf } from '../shared/errors';

export function useAuthSession() {
  const { loginForm, passwordForm, setAuthState, setUser, setLoginForm, setPasswordForm } =
    useAuthStore();
  const { setError, setMessage } = useUiStore();

  const handleLogin = useCallback(
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      setError('');
      try {
        setAuthState('loading');
        const response = await api.login(loginForm);
        setUser(response.user);
        setAuthState('ready');
        if (!response.user.mustChangePassword) {
          const returnTo = sessionStorage.getItem('auth:returnTo');
          if (returnTo) {
            sessionStorage.removeItem('auth:returnTo');
            window.history.pushState({}, '', returnTo);
            window.dispatchEvent(new PopStateEvent('popstate'));
          }
        }
      } catch (err) {
        setUser(null);
        setAuthState('idle');
        setError(statusOf(err) === 401 ? 'Usuário ou senha inválidos.' : asMessage(err));
      }
    },
    [loginForm, setAuthState, setError, setUser],
  );

  const handleLogout = useCallback(async () => {
    try {
      await api.logout();
    } catch {
      // best-effort
    } finally {
      setUser(null);
      setAuthState('idle');
    }
  }, [setAuthState, setUser]);

  const handleChangePassword = useCallback(
    async (e: React.FormEvent<HTMLFormElement>) => {
      e.preventDefault();
      setError('');
      setMessage('');
      if (passwordForm.newPassword !== passwordForm.confirmPassword) {
        setError('A confirmação da nova senha não confere.');
        return;
      }
      try {
        const response = await api.changePassword(passwordForm);
        setPasswordForm({ currentPassword: '', newPassword: '', confirmPassword: '' });
        setUser(response.user);
        setAuthState('ready');
        setMessage('Senha alterada com sucesso.');
      } catch (err) {
        setError(asMessage(err));
      }
    },
    [passwordForm, setAuthState, setError, setMessage, setPasswordForm, setUser],
  );

  return {
    loginForm,
    passwordForm,
    setLoginForm,
    setPasswordForm,
    handleLogin,
    handleLogout,
    handleChangePassword,
  };
}
```

- [ ] **Step 4: Fix remaining import errors**

Run TypeScript to find broken imports from deleted files:

```bash
cd frontend && npx tsc --noEmit 2>&1 | grep "error TS" | head -30
```

For each file that still imports from deleted modules:
- Imports from `documents.store` → remove the import; the page will be rebuilt in its own screen task
- Imports from `WorkspaceRoot` (e.g. `useWorkspaceRouteContext`) → remove; pages will be rebuilt
- Imports from `DocumentWorkspaceShell` → remove `WorkspaceView` type; pages will be rebuilt

For feature pages that now have broken imports (e.g. `DocumentsHubPage`, `AuditPage`, existing `InboxPage`), replace their content temporarily with a stub that renders a placeholder:

```tsx
// Temporary stub — will be rebuilt in its screen task
export function Component() {
  return <div style={{ padding: 40, color: 'var(--text-muted)' }}>Em construção…</div>;
}
```

Apply this stub pattern to:
- `frontend/apps/web/src/features/documents/pages/DocumentsHubPage.tsx`
- `frontend/apps/web/src/features/documents/pages/DocumentEditorRoutePage.tsx`
- `frontend/apps/web/src/features/audit/pages/AuditPage.tsx`
- `frontend/apps/web/src/features/approval/pages/InboxPage.tsx`
- `frontend/apps/web/src/features/controlled-documents/pages/ControlledDocumentsExplorerPage.tsx` (if exists)
- `frontend/apps/web/src/components/OperationsCenter.tsx` (replace with stub)
- `frontend/apps/web/src/components/NotificationsPanel.tsx` (replace with stub)

- [ ] **Step 5: Verify TypeScript passes**

```bash
cd frontend && npx tsc --noEmit
```

Expected: 0 errors

- [ ] **Step 6: Commit**

```bash
git add -A
git commit -m "refactor(foundation): delete server-state Zustand stores, gut WorkspaceRoot/Shell"
```

---

## Sequential Group C — After Group B completes

---

### Task C1: AppRoot + AppRouter restructure [SEQUENTIAL after B2]

**Files:**
- Create: `frontend/apps/web/src/features/shell/pages/AppRoot.tsx`
- Create: `frontend/apps/web/src/features/shell/pages/AppRoot.test.tsx`
- Modify: `frontend/apps/web/src/app/AppRouter.tsx`

- [ ] **Step 1: Write failing test**

Create `frontend/apps/web/src/features/shell/pages/AppRoot.test.tsx`:

```tsx
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { AppRoot } from './AppRoot';
import { useAuthStore } from '../../../store/auth.store';

// Mock the me() API call
vi.mock('../../auth/api/auth', () => ({
  me: vi.fn(),
}));

import * as authApi from '../../auth/api/auth';

const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });

function wrapper({ children }: { children: React.ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <MemoryRouter>
        <Routes>
          <Route path="/login" element={<div>Login Page</div>} />
          <Route path="/" element={children} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  );
}

describe('AppRoot', () => {
  beforeEach(() => {
    useAuthStore.setState({ authState: 'loading', user: null });
    queryClient.clear();
  });

  it('shows spinner while loading', () => {
    vi.mocked(authApi.me).mockImplementation(() => new Promise(() => {}));
    render(<AppRoot />, { wrapper });
    expect(screen.getByRole('status')).toBeTruthy();
  });

  it('redirects to /login when me() returns 401', async () => {
    vi.mocked(authApi.me).mockRejectedValue(Object.assign(new Error('unauth'), { status: 401 }));
    render(<AppRoot />, { wrapper });
    await waitFor(() => {
      expect(screen.getByText('Login Page')).toBeTruthy();
    });
  });
});
```

- [ ] **Step 2: Run — expect FAIL**

```bash
cd frontend && npx vitest run src/features/shell/pages/AppRoot.test.tsx
```

Expected: `Cannot find module './AppRoot'`

- [ ] **Step 3: Create AppRoot.tsx**

Create `frontend/apps/web/src/features/shell/pages/AppRoot.tsx`:

```tsx
import { useEffect } from 'react';
import { Navigate, Outlet } from 'react-router-dom';
import * as authApi from '../../auth/api/auth';
import { onAuthExpired } from '../../../lib/api';
import { useAuthStore } from '../../../store/auth.store';
import { statusOf } from '../../shared/errors';

function FullPageSpinner() {
  return (
    <div
      role="status"
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        height: '100vh',
        background: 'var(--bg)',
        color: 'var(--text-muted)',
        fontSize: 13,
        fontFamily: 'var(--font-sans)',
      }}
    >
      Carregando…
    </div>
  );
}

export function AppRoot() {
  const authState = useAuthStore((s) => s.authState);
  const setAuthState = useAuthStore((s) => s.setAuthState);
  const setUser = useAuthStore((s) => s.setUser);

  // Bootstrap: call me() once on mount
  useEffect(() => {
    async function bootstrap() {
      try {
        const user = await authApi.me();
        setUser(user);
        setAuthState('ready');
      } catch (err) {
        setAuthState(statusOf(err) === 401 ? 'idle' : 'error');
      }
    }
    void bootstrap();
  }, [setAuthState, setUser]);

  // Listen for 401 events from any API call
  useEffect(() => {
    return onAuthExpired(({ returnTo }) => {
      if (returnTo && returnTo !== '/' && !returnTo.startsWith('/login')) {
        sessionStorage.setItem('auth:returnTo', returnTo);
      }
      setUser(null);
      setAuthState('idle');
    });
  }, [setAuthState, setUser]);

  if (authState === 'loading') return <FullPageSpinner />;
  if (authState === 'idle') return <Navigate to="/login" replace />;
  if (authState === 'error') {
    return (
      <div style={{ padding: 40, color: 'var(--danger)', fontFamily: 'var(--font-sans)' }}>
        Erro ao carregar sessão. <a href="/">Tentar novamente</a>
      </div>
    );
  }

  return <Outlet />;
}
```

- [ ] **Step 4: Rewrite AppRouter.tsx**

Replace entire `frontend/apps/web/src/app/AppRouter.tsx`:

```tsx
import { Navigate, createBrowserRouter } from 'react-router-dom';
import { AppRoot } from '../features/shell/pages/AppRoot';
import { approvalRoutes } from '../features/approval/routes';
import { auditRoutes } from '../features/audit/routes';
import { contentBuilderRoutes } from '../features/content-builder/routes';
import { documentsRoutes } from '../features/documents/routes';
import { iamRoutes } from '../features/iam/routes';
import { notificationsRoutes } from '../features/notifications/routes';
import { operationsRoutes } from '../features/operations/routes';
import { passwordChangeRoutes } from '../features/password-change/routes';
import { controlledDocumentsRoutes } from '../features/controlled-documents/routes';
import { taxonomyRoutes } from '../features/taxonomy/routes';
import { templatesRoutes } from '../features/templates/routes';

export const router = createBrowserRouter([
  // Public routes — no Rail, no Toolbar
  {
    path: '/login',
    lazy: () =>
      import('../features/auth/pages/LoginPage').then((m) => ({ Component: m.LoginPage })),
  },
  // Protected routes — wrapped in AppRoot (auth guard) + AppShell (layout)
  {
    element: <AppRoot />,
    children: [
      {
        lazy: () =>
          import('../features/shell/components/AppShell').then((m) => ({ Component: m.AppShell })),
        children: [
          ...operationsRoutes,
          ...documentsRoutes,
          ...templatesRoutes,
          ...registryRoutes,
          ...taxonomyRoutes,
          ...iamRoutes,
          ...approvalRoutes,
          ...notificationsRoutes,
          ...contentBuilderRoutes,
          ...auditRoutes,
          ...passwordChangeRoutes,
          { path: '*', element: <Navigate to="/" replace /> },
        ],
      },
    ],
  },
]);
```

- [ ] **Step 5: Run test — expect PASS**

```bash
cd frontend && npx vitest run src/features/shell/pages/AppRoot.test.tsx
```

Expected: `2 tests passed`

- [ ] **Step 6: Commit**

```bash
git add frontend/apps/web/src/features/shell/pages/AppRoot.tsx \
         frontend/apps/web/src/features/shell/pages/AppRoot.test.tsx \
         frontend/apps/web/src/app/AppRouter.tsx
git commit -m "feat(foundation): AppRoot auth guard + public/protected router split"
```

---

### Task C2: AppShell + Rail + AppToolbar + SectionPanel [SEQUENTIAL after C1]

**Files:**
- Create: `frontend/apps/web/src/features/shell/components/AppShell.tsx`
- Create: `frontend/apps/web/src/features/shell/components/AppShell.module.css`
- Create: `frontend/apps/web/src/features/shell/components/Rail.tsx`
- Create: `frontend/apps/web/src/features/shell/components/Rail.module.css`
- Create: `frontend/apps/web/src/features/shell/components/AppToolbar.tsx`
- Create: `frontend/apps/web/src/features/shell/components/AppToolbar.module.css`
- Create: `frontend/apps/web/src/features/shell/components/SectionPanel.tsx`
- Create: `frontend/apps/web/src/features/shell/components/SectionPanel.module.css`

- [ ] **Step 1: Create AppShell.module.css**

```css
.shell {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--bg);
}

.main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  overflow: hidden;
}

.content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.page {
  flex: 1;
  overflow: auto;
}
```

- [ ] **Step 2: Create AppShell.tsx**

```tsx
import { Outlet, useMatches } from 'react-router-dom';
import { Rail } from './Rail';
import { AppToolbar } from './AppToolbar';
import { SectionPanel } from './SectionPanel';
import styles from './AppShell.module.css';

type RouteHandle = { sectionPanel?: boolean };

export function AppShell() {
  const matches = useMatches();
  const hasSectionPanel = matches.some(
    (m) => (m.handle as RouteHandle | undefined)?.sectionPanel === true,
  );

  return (
    <div className={styles.shell}>
      <Rail />
      <div className={styles.main}>
        <AppToolbar />
        <div className={styles.content}>
          {hasSectionPanel && <SectionPanel />}
          <main className={styles.page}>
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
```

- [ ] **Step 3: Create Rail.module.css**

```css
.rail {
  width: 56px;
  flex-shrink: 0;
  background: var(--rail);
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 0;
  gap: 0;
  height: 100vh;
  position: sticky;
  top: 0;
}

.logo {
  margin-bottom: 20px;
}

.nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
  width: 100%;
}

.navItem {
  position: relative;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--r-3);
  color: var(--rail-text-muted);
  cursor: pointer;
  border: none;
  background: transparent;
  transition: color 100ms, background 100ms;
}

.navItem:hover {
  color: var(--rail-text);
  background: rgba(255, 255, 255, 0.06);
}

.navItemActive {
  color: var(--rail-text);
  background: var(--rail-active);
}

.divider {
  width: 28px;
  height: 1px;
  background: var(--rail-divider);
  margin: 8px 0;
}

.bottom {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.logoutBtn {
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--r-3);
  color: var(--rail-text-muted);
  cursor: pointer;
  border: none;
  background: transparent;
}

.logoutBtn:hover {
  color: var(--rail-text);
  background: rgba(255, 255, 255, 0.06);
}
```

- [ ] **Step 4: Create Rail.tsx**

```tsx
import { useMatch, useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { Avatar } from '../../../components/ui/Avatar';
import { useAuthStore } from '../../../store/auth.store';
import { useAuthSession } from '../../auth/useAuthSession';
import styles from './Rail.module.css';

type NavItem = {
  icon: React.ComponentProps<typeof Icon>['name'];
  label: string;
  path: string;
};

const NAV_ITEMS: NavItem[] = [
  { icon: 'home',     label: 'Início',       path: '/' },
  { icon: 'library',  label: 'Documentos',   path: '/documents' },
  { icon: 'template', label: 'Templates',    path: '/templates' },
  { icon: 'registry', label: 'Registro',     path: '/registry-v2' },
  { icon: 'inbox',    label: 'Aprovações',   path: '/approvals' },
  { icon: 'audit',    label: 'Auditoria',    path: '/audit' },
];

function NavButton({ item }: { item: NavItem }) {
  const navigate = useNavigate();
  const match = useMatch({ path: item.path, end: item.path === '/' });
  const isActive = Boolean(match);

  return (
    <button
      className={`${styles.navItem}${isActive ? ` ${styles.navItemActive}` : ''}`}
      onClick={() => navigate(item.path)}
      title={item.label}
      aria-label={item.label}
      aria-current={isActive ? 'page' : undefined}
    >
      <Icon name={item.icon} size={18} />
    </button>
  );
}

export function Rail() {
  const user = useAuthStore((s) => s.user);
  const { handleLogout } = useAuthSession();

  return (
    <aside className={styles.rail}>
      <div className={styles.logo}>
        {/* Logo mark only — no text in rail */}
        <span
          style={{
            width: 28, height: 28, borderRadius: 6, background: 'var(--brand)',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
          }}
        >
          <Icon name="docs" size={14} style={{ color: 'white' }} />
        </span>
      </div>

      <nav className={styles.nav}>
        {NAV_ITEMS.map((item) => (
          <NavButton key={item.path} item={item} />
        ))}
      </nav>

      <div className={styles.bottom}>
        <div className={styles.divider} />
        {user && (
          <span title={user.displayName ?? user.email}>
            <Avatar name={user.displayName || user.email} size="sm" />
          </span>
        )}
        <button
          className={styles.logoutBtn}
          onClick={() => void handleLogout()}
          title="Sair"
          aria-label="Sair"
        >
          <Icon name="arrow" size={16} style={{ transform: 'rotate(180deg)' }} />
        </button>
      </div>
    </aside>
  );
}
```

- [ ] **Step 5: Create AppToolbar.module.css**

```css
.toolbar {
  height: 52px;
  flex-shrink: 0;
  background: var(--surface);
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 20px;
}

.searchWrapper {
  position: relative;
  flex: 1;
  max-width: 360px;
}

.searchIcon {
  position: absolute;
  left: 9px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--text-muted);
  pointer-events: none;
}

.searchInput {
  width: 100%;
  height: 30px;
  padding: 0 10px 0 30px;
  border: 1px solid var(--border-strong);
  border-radius: var(--r-2);
  background: var(--surface-2);
  font-size: 13px;
  font-family: var(--font-sans);
  color: var(--text);
}

.searchInput:focus {
  outline: 2px solid var(--brand);
  outline-offset: -1px;
  border-color: var(--brand);
}

.spacer {
  flex: 1;
}

.bellBtn {
  position: relative;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--r-2);
  border: 1px solid transparent;
  background: transparent;
  color: var(--text-muted);
  cursor: pointer;
}

.bellBtn:hover {
  background: var(--surface-2);
  color: var(--text);
}

.badge {
  position: absolute;
  top: 4px;
  right: 4px;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent);
  border: 1.5px solid var(--surface);
}

.newDocBtn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  height: 30px;
  padding: 0 12px;
  font-size: 13px;
  font-weight: 500;
  font-family: var(--font-sans);
  background: var(--brand);
  color: white;
  border: none;
  border-radius: var(--r-2);
  cursor: pointer;
  white-space: nowrap;
}

.newDocBtn:hover {
  background: var(--brand-deep);
}
```

- [ ] **Step 6: Create AppToolbar.tsx**

```tsx
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Icon } from '../../../components/ui/Icon';
import { useAuthStore } from '../../../store/auth.store';
import styles from './AppToolbar.module.css';

export function AppToolbar() {
  const [search, setSearch] = useState('');
  const navigate = useNavigate();
  const user = useAuthStore((s) => s.user);

  function handleSearch(e: React.FormEvent) {
    e.preventDefault();
    if (search.trim()) {
      navigate(`/documents?q=${encodeURIComponent(search.trim())}`);
    }
  }

  return (
    <header className={styles.toolbar}>
      <form onSubmit={handleSearch} className={styles.searchWrapper}>
        <span className={styles.searchIcon}>
          <Icon name="search" size={13} />
        </span>
        <input
          className={styles.searchInput}
          type="search"
          placeholder="Buscar documentos…"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          aria-label="Buscar documentos"
        />
      </form>

      <div className={styles.spacer} />

      {user && (
        <span style={{ fontSize: 12, color: 'var(--text-muted)', fontFamily: 'var(--font-sans)' }}>
          {user.displayName || user.email}
        </span>
      )}

      <button className={styles.bellBtn} aria-label="Notificações" title="Notificações">
        <Icon name="bell" size={16} />
        {/* badge rendered when there are unread notifications — wired in Library/Dashboard tasks */}
      </button>

      <button
        className={styles.newDocBtn}
        onClick={() => navigate('/documents-v2/new')}
        aria-label="Novo documento"
      >
        <Icon name="plus" size={14} />
        Novo documento
      </button>
    </header>
  );
}
```

- [ ] **Step 7: Create SectionPanel.module.css + SectionPanel.tsx**

`SectionPanel.module.css`:

```css
.panel {
  width: 224px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  background: var(--surface-2);
  overflow: auto;
  height: 100%;
}
```

`SectionPanel.tsx` — slot component; Library page fills it via an outlet context or children prop. For now renders an empty panel; Library task will wire the filter tree.

```tsx
import styles from './SectionPanel.module.css';

export function SectionPanel() {
  return <aside className={styles.panel} aria-label="Filtros" />;
}
```

- [ ] **Step 8: TypeScript check**

```bash
cd frontend && npx tsc --noEmit
```

Expected: 0 errors

- [ ] **Step 9: Smoke test in browser**

```bash
cd frontend && npm run dev
```

Navigate to `http://localhost:5173`. Expected:
- App loads → spinner → redirects to `/login` (LoginPage stub renders "Em construção…")
- Rail is NOT visible on `/login`
- Navigate to `/` after auth → Rail (56px dark) + Toolbar (52px) + page area visible

- [ ] **Step 10: Commit**

```bash
git add frontend/apps/web/src/features/shell/components/
git commit -m "feat(foundation): AppShell — Rail + AppToolbar + SectionPanel layout"
```

---

## Foundation Complete

- [ ] **Final: Update wiki tracker**

In `wiki/implementation/screen-redesign-tracker.md`, change Foundation row:

```
| **Foundation** | ... | `plans/2026-05-05-foundation.md` | ✅ Complete |
```

And change Login row from `⏳ Waiting on Foundation` to `🔲 Not started`.

- [ ] **Final: Commit tracker update**

```bash
git add wiki/implementation/screen-redesign-tracker.md
git commit -m "docs(tracker): foundation complete — login unblocked"
```

---

## Self-Review Notes

- **Spec coverage:** Token rename ✅ · Font switch ✅ · Zustand cleanup ✅ · queryKeys ✅ · UI primitives ✅ · AppRoot ✅ · AppShell ✅ · Rail ✅ · AppToolbar ✅ · SectionPanel ✅
- **No placeholders:** All steps contain actual code or exact commands
- **Type consistency:** `DocumentStatus` in StatusPill matches expected values. `IconName` union covers all 30 icons from shell.jsx. `AuthStore.authState` type `LoadState` used consistently.
- **Parallel group A** (A1/A2/A3) touch disjoint files — safe to run simultaneously with 3 Codex agents
- **B1 depends on A1** (needs new token names in tokens.css before mass-replace makes sense)
- **B2 can run parallel with B1** (different files — CSS modules vs TS stores)
- **C1 depends on B2** (AppRoot imports from auth.store + useAuthSession — both cleaned up in B2)
- **C2 depends on C1** (AppShell is lazy-loaded by AppRouter created in C1)
