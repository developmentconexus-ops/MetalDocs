export { diffTokensVsSchema, type SchemaDiff } from './diff';
export { WHITELIST, BLACKLIST, classifyBlacklist, isElementAllowed } from './ooxml';
export {
  IDENT_RE, RESERVED_IDENTS, MAX_SECTION_DEPTH, isValidIdent, isReservedIdent,
  scanText, detectTokens,
  type RawTag, type DetectedToken,
} from './grammar';
export type { Token, TokenKind, ParseError, ParseResult } from './types';
// NOTE: parseDocxTokens (JSZip/fast-xml-parser) is intentionally NOT re-exported here.
// It is server-side only — import it from './parser' directly. Keeping it off the
// barrel keeps @metaldocs/shared-tokens dep-free for the browser bundle (spec §4.1).
