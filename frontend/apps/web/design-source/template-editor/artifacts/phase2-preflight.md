# Phase 2 — Pre-flight · template-editor

Date: 2026-05-10
Tier: Heavy

Inline (no subagent boot) — primitive audit cache hit + zero new shared atoms + zero codegen.

## Primitive audit cache (14-day rule, skill v1.3)

All consumed primitives audited within last 14 days — re-audit skipped.

| Primitive | Last verified | Source artifact |
|---|---|---|
| `EditorChrome` + slots + `VersionBadge` + `AutosaveStatus` | 2026-05-06 | `wiki/modules/editor-chrome.md` |
| `StatusPill` | 2026-05-08 | `wiki/modules/documents.md` (LibraryPage extension) |
| `PlaceholderCatalogPanel` | exists, in-tree | `features/templates/PlaceholderCatalogPanel.tsx` |
| `useTemplateDraft` / `useTemplateAutosave` / `useTemplateSchemas` | in-tree, no API change | `features/templates/hooks/` |
| Inner rail `<aside class="rail">` pattern | 2026-05-06 | `DocumentEditorPage.module.css` (mirror these tokens) |

No tsc in this phase (skill v1.3 — runs in Phase 4).

## Codegen

None. Zero new endpoints. Existing `templatesV2.ts` client + `submitForReview` cover all calls.

## Eigenpal outline spike (Phase 1.8 #3 follow-up)

Verified surface in `node_modules/@eigenpal/docx-js-editor/dist/DocumentAgent-BYeiUfe-.d.ts`:

```ts
agent.getAgentContext(outlineMaxChars?: number): AgentContext;
agent.getStyles(): StyleInfo[];
agent.getDocument(): Document;
// AgentContext.outline: ParagraphOutline[]   ← per-paragraph outline data
// Paragraph.formatting.outlineLevel: 0..9    ← Word heading level
```

**No direct `getHeadings()` method.** Headings derivable two ways:

1. **Walk `agent.getDocument()` body** → filter paragraphs where `formatting.outlineLevel` is set OR `style` matches Word heading slot (`Heading1`..`Heading9`). Map to `{level, text}`. Cheapest, most accurate.
2. Use `agent.getAgentContext({outlineMaxChars: 0}).outline` → `ParagraphOutline[]` — already filtered to outline-relevant paragraphs by eigenpal. Less work but couples to whatever heuristic eigenpal uses internally.

**Decision (Phase 3c):** Option 1. Walk document body, derive `Heading[]` in a small util `features/templates/lib/readHeadings.ts`. Keeps dependency surface explicit; `outlineLevel` + `style` are stable Word OOXML constructs.

```ts
// Phase 3c skeleton (not yet written)
export function readHeadings(agent: DocumentAgent): Heading[] {
  const doc = agent.getDocument();
  const out: Heading[] = [];
  walk(doc.body, (para, idx) => {
    const lvl = para.formatting?.outlineLevel;
    if (lvl != null && lvl >= 0 && lvl <= 5) {
      out.push({ id: `h-${idx}`, level: (lvl + 1) as 1|2|3|4|5|6, text: extractText(para) });
    }
  });
  return out;
}
```

Outline icon ships in Phase 3c. No backlog defer.

## Route stub

Existing route `/templates/:templateId/versions/:versionNum` (already wired) stays. Page rename is Vite-friendly — `routes.tsx` import path updated in same commit as file rename. No URL change.

## Phase 3 entry conditions

- [x] Layout locked (Phase 0)
- [x] Component tree locked (Phase 1)
- [x] No new primitives, no codegen, no shared atoms (Phase 2)
- [x] Eigenpal outline surface verified (Phase 2)
- [ ] Phase 3a: structure mirror — rename file + restructure JSX (inline, main agent, Heavy tier)
- [ ] Phase 3b: style port — CSS Module restyle to mirror `DocumentEditorPage.module.css` rail/canvas tokens (subagent)
- [ ] Phase 3c: state wiring — outline panel + `readHeadings` util + `ApiError`/`resolveErrorMessage` upgrade for submit/import (subagent)

Ready for Phase 3a (inline DOM mirror).
