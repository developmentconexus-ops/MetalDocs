import type { TokenKind } from './types';

export const IDENT_RE = /^[A-Za-z_][A-Za-z0-9_]*$/;

export const RESERVED_IDENTS = new Set<string>([
  '__proto__',
  'constructor',
  'prototype',
  'toString',
  'valueOf',
  'hasOwnProperty',
  'tenant_id',
  'document_id',
  'template_version_id',
  'revision_id',
  'session_id',
]);

export function isValidIdent(s: string): boolean {
  if (!IDENT_RE.test(s)) return false;
  return true;
}

export function isReservedIdent(s: string): boolean {
  return RESERVED_IDENTS.has(s);
}

export const MAX_SECTION_DEPTH = 1;

export interface RawTag {
  raw: string;
  kind: TokenKind;
  inner: string;
  start: number;
  end: number;
}

// The single scan grammar. `[^{}]+` body matches the freeze engine's default
// docxtemplater tag body; control prefixes are classified, everything else is `var`.
const SCAN_RE = /\{([#^/>])?([^{}]+)\}/g;

function prefixKind(prefix: string | undefined): TokenKind {
  switch (prefix) {
    case '#': return 'section';
    case '^': return 'inverted';
    case '/': return 'closing';
    case '>': return 'partial';
    default: return 'var';
  }
}

/** Every `{...}` tag in a flat string, control-prefix classified, inner trimmed. */
export function scanText(text: string): RawTag[] {
  const out: RawTag[] = [];
  SCAN_RE.lastIndex = 0;
  let m: RegExpExecArray | null;
  while ((m = SCAN_RE.exec(text)) !== null) {
    const [raw, prefix, body] = m;
    const inner = body.trim();
    if (inner === '') continue;
    out.push({ raw, kind: prefixKind(prefix), inner, start: m.index, end: m.index + raw.length });
  }
  return out;
}

export interface DetectedToken {
  name: string;
  kind: TokenKind;
  valid: boolean;
  reserved: boolean;
}

/**
 * Browser-side detection: every variable tag in the text, first-seen-deduped.
 * Detect broad (any `{ident}`), validate strict (snake_case + not reserved).
 * Control tags (#/^/>) are not variables and are excluded.
 */
export function detectTokens(text: string): DetectedToken[] {
  const out: DetectedToken[] = [];
  const seen = new Set<string>();
  for (const tag of scanText(text)) {
    if (tag.kind !== 'var') continue;
    if (seen.has(tag.inner)) continue;
    seen.add(tag.inner);
    out.push({
      name: tag.inner,
      kind: tag.kind,
      valid: isValidIdent(tag.inner) && !isReservedIdent(tag.inner),
      reserved: isReservedIdent(tag.inner),
    });
  }
  return out;
}
