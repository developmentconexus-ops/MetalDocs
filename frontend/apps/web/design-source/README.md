# Design Source

Reference HTML/PNG exports from claude.design and other design tools. Committed to the repo so the wiki and the `metaldocs-frontend` skill can point at concrete artifacts. Excluded from the production bundle by Vite — purely a developer reference.

## Layout

One folder per screen, kebab-case slug:

```
design-source/
├── README.md                    ← this file
├── <screen-slug>/
│   ├── <slug>.html              ← original export, read-only reference
│   ├── <slug>.png               ← screenshot (preferred over reading HTML)
│   └── NOTES.md                 ← implementer notes (required)
└── ...
```

## NOTES.md template

Every screen folder must have a `NOTES.md` with at minimum:

```markdown
# <Screen name>

**Owning feature:** `features/<domain>` (e.g., `features/documents`)
**Target route:** `/documents/:id/edit` (or whatever)
**Page file:** `features/<domain>/pages/<PageName>.tsx`

## Reused primitives

- `components/ui/Button`
- `components/ui/Card`
- ...

## New primitives needed (if any)

- (if a composition repeats across 2+ screens, list it here so we extract it)

## Data sources

- `GET /api/v1/...` via `useXxxQuery`
- ...

## Open questions

- ...
```

## Workflow

1. Drop the export into a new slug folder.
2. Write the `NOTES.md` (5 minutes — this is the contract for the implementer).
3. Reference the slug in the implementation task.
4. Implementer reads `NOTES.md` first, then builds per `wiki/architecture/frontend-structure.md`.
5. Once shipped, link the slug from the relevant `wiki/modules/<feature>.md` doc.

## Why commit these files?

- **Versioned design history** — diffs show how a screen evolved.
- **Single source of truth** — the wiki and skill point at concrete files, not external links that rot.
- **Onboarding** — new contributors can browse the gallery without external tool access.

The HTML files are excluded from the production bundle by Vite's default ignore for non-`src/` assets, so there is zero shipping cost.
