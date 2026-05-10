import type { DocxEditorRef } from '@eigenpal/docx-js-editor/react';

export type Heading = {
  id: string;
  level: 1 | 2 | 3 | 4 | 5 | 6;
  text: string;
};

/**
 * Derive headings from the eigenpal document agent.
 *
 * Uses `agent.getAgentContext().outline` — eigenpal-filtered ParagraphOutline[]
 * with `isHeading` + `headingLevel` already resolved against Word's heading slots
 * (`Heading1..Heading9`) and OOXML `outlineLevel`. Levels above 6 collapse to 6
 * to keep the panel readable.
 */
export function readHeadings(editorRef: React.RefObject<DocxEditorRef> | null | undefined): Heading[] {
  const agent = editorRef?.current?.getAgent?.();
  if (!agent || typeof agent.getAgentContext !== 'function') return [];

  let context: ReturnType<NonNullable<typeof agent.getAgentContext>>;
  try {
    context = agent.getAgentContext();
  } catch {
    return [];
  }

  const outline = context?.outline ?? [];
  const out: Heading[] = [];
  for (const para of outline) {
    if (!para.isHeading) continue;
    const lvl = para.headingLevel ?? 1;
    const clamped = (lvl < 1 ? 1 : lvl > 6 ? 6 : lvl) as Heading['level'];
    const text = (para.preview ?? '').trim();
    if (!text) continue;
    out.push({ id: `h-${para.index}`, level: clamped, text });
  }
  return out;
}
