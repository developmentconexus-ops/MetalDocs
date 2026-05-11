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
 *
 * The ref parameter is typed broadly (`{ getAgent?: () => unknown }`) so both the
 * raw `DocxEditorRef` and the wrapper `MetalDocsEditorRef` (which does not expose
 * `getAgent` directly) can be passed. When `getAgent` is missing the function returns [].
 */
export function readHeadings(editorRef: React.RefObject<{ getAgent?(): unknown }> | null | undefined): Heading[] {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const agent = editorRef?.current?.getAgent?.() as any;
  if (!agent || typeof agent.getAgentContext !== 'function') return [];

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let context: any;
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
