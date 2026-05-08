# Phase 3a — Structure Mirror

**Screen:** Templates List (`/templates-v2`)
**Date:** 2026-05-07

## DOM tree

```
TemplatesListPage
└── div.page
    └── div.content
        ├── WorkspaceHeroHeader (kicker, title, subtitle, action)
        │   └── action: button.newBtn
        │       ├── Icon name="plus"
        │       └── "Novo template"
        ├── TabBar (Todos · Publicados · Rascunhos · Arquivados)
        └── div.cardGrid
            └── TemplateCard (×6)
                └── div.card
                    ├── div.previewArea
                    │   ├── MiniDocPreview
                    │   │   └── div.miniDoc
                    │   │       ├── div.miniDocBrand
                    │   │       └── div.miniDocLine (×8)
                    │   └── div.badges
                    │       ├── StatusPill
                    │       └── CodeChip (version)
                    └── div.body
                        ├── div.title
                        ├── div.divider
                        └── div.footer
                            ├── Avatar (sm)
                            ├── span.author
                            └── span.time
```

## Class-name mapping

| Design source (screens-2.jsx) | Module class |
|---|---|
| outer wrapper (inline `flex:1, overflow:auto, padding 24/28`) | `TemplatesListPage.module.css → .page` |
| inner content stack (implicit) | `.content` |
| `<button className="btn btn-primary">` | `.newBtn` |
| `<div style={{ display:'grid', gridTemplateColumns:'repeat(3,1fr)' }}>` | `.cardGrid` |
| `<div className="card">` (template card root) | `TemplateCard.module.css → .card` |
| mini-doc-preview wrapper (height 110, surface-2 bg) | `.previewArea` |
| top-right badges column (status + version) | `.badges` |
| body padding 14 wrapper | `.body` |
| `<div className="h3">` title | `.title` |
| `<div className="divider">` | `.divider` |
| author row (avatar + name + time) | `.footer` |
| `<span className="caption">` first name | `.author` |
| `<span className="tiny mono">` updated | `.time` |
| inner "paper" white card (80×100) | `MiniDocPreview.module.css → .miniDoc` |
| brand top bar (height 4, brand bg) | `.miniDocBrand` |
| 8 horizontal text lines | `.miniDocLine` |

## Phase 0 cuts applied

- Removed: card description paragraph (`tpl.d`)
- Removed: bound profile pills row + `+N` overflow chip + "Não vinculado" fallback
- Removed: "Em revisão" tab (replaced with "Arquivados")
- Mock 5th card (`s: 'review'`) dropped; new mock `FORM — Formulário Antigo` (status=archived) added so each tab has data.

## Files created/modified

**Modified:**
- `frontend/apps/web/src/features/templates/TemplatesListPage.tsx` (full rewrite, mirrored DOM, mock data)
- `frontend/apps/web/src/features/templates/TemplatesListPage.module.css` (skeleton, empty rules)

**Created:**
- `frontend/apps/web/src/features/templates/components/TemplateCard.tsx`
- `frontend/apps/web/src/features/templates/components/TemplateCard.module.css`
- `frontend/apps/web/src/features/templates/components/MiniDocPreview.tsx`
- `frontend/apps/web/src/features/templates/components/MiniDocPreview.module.css`

## Anything that did not map cleanly

None. All design DOM nodes (post-cut) mapped 1:1 to module classes or primitive components. `TemplatesListPageProps` shape preserved (`onOpenTemplate`, `onCreate`) so `TemplatesListRoutePage.tsx` compiles unchanged; handlers will be wired to cards in Phase 3c.

## Typecheck

`pnpm tsc --noEmit -p tsconfig.build.json` — no new errors in Templates files. Pre-existing errors only in unrelated features (auth, documents, shell).
