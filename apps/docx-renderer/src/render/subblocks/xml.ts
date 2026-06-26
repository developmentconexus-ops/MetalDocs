// Sub-block OOXML is injected into the document via docxtemplater's raw token
// (`{@__header_composition__}` / `{@__footer_composition__}`), which inserts the
// value VERBATIM — it does NOT escape, unlike the text token `{name}`. Any tenant
// value baked into a sub-block scaffold therefore MUST be XML-escaped here, at the
// `<w:t>` sink, or a value containing `<`/`&`/`>` injects arbitrary OOXML into the
// frozen, hash-audited document. This is the single escaping point for that path.
const XML_ESCAPES: Record<string, string> = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
  "'": "&apos;",
};

export function escapeXml(s: string): string {
  return s.replace(/[&<>"']/g, (ch) => XML_ESCAPES[ch]);
}

/** Stringify an unknown value for safe interpolation into OOXML text. */
export function xmlText(v: unknown): string {
  if (v === null || v === undefined) return "";
  return escapeXml(String(v));
}
