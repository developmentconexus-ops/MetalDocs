// The framework-free (`.`) door: the ONLY server-side eigenpal touchpoint.
// No React, no DOM — safe to inline into the docx-renderer Node bundle.
import { processTemplateDetailed } from '@eigenpal/docx-editor-core/headless';
import JSZip from 'jszip';

export type RenderErrorKind =
  | 'template_parse'      // malformed template / unreadable DOCX
  | 'template_render'     // engine failed while substituting
  | 'undefined_variable'  // a referenced variable was not provided
  | 'unknown';            // anything we cannot classify

/** MetalDocs-owned render failure. Replaces eigenpal's thrown TemplateError shape. */
export class RenderError extends Error {
  readonly kind: RenderErrorKind;
  readonly variable?: string;
  override readonly cause?: unknown;
  constructor(kind: RenderErrorKind, message: string, opts?: { variable?: string; cause?: unknown }) {
    super(message);
    this.name = 'RenderError';
    this.kind = kind;
    this.variable = opts?.variable;
    this.cause = opts?.cause;
  }
}

export interface RenderWarning {
  kind: RenderErrorKind;
  message: string;
  variable?: string;
}

export interface RenderResult {
  buffer: Uint8Array;
  replacedVariables: string[];
  unreplacedVariables: string[];
  warnings: RenderWarning[];
}

export interface ProcessTemplateOptions {
  /** Forwarded to eigenpal; today the renderer uses 'empty'. */
  nullGetter?: 'empty';
  /** W3C traceparent for trace-ready logging (span emission deferred to RF-1). */
  traceparent?: string;
}

export interface TemplateProcessor {
  processTemplate(
    docx: ArrayBuffer,
    variables: Record<string, string>,
    opts?: ProcessTemplateOptions,
  ): Promise<RenderResult>;
}

// --- vendor seam (kept narrow + injectable so tests need no real DOCX) ---

interface EigenpalTemplateError {
  message: string;
  variable?: string;
  type: 'parse' | 'render' | 'undefined' | 'unknown';
  originalError?: unknown;
}

interface EigenpalRawResult {
  buffer: ArrayBuffer;
  replacedVariables?: string[];
  unreplacedVariables?: string[];
  warnings?: EigenpalTemplateError[];
}

export interface EigenpalEngine {
  processTemplateDetailed(
    docx: ArrayBuffer,
    variables: Record<string, string>,
    opts: { nullGetter: 'empty' },
  ): EigenpalRawResult;
}

const TYPE_TO_KIND: Record<EigenpalTemplateError['type'], RenderErrorKind> = {
  parse: 'template_parse',
  render: 'template_render',
  undefined: 'undefined_variable',
  unknown: 'unknown',
};

function isTemplateError(e: unknown): e is EigenpalTemplateError {
  return (
    typeof e === 'object' &&
    e !== null &&
    'type' in e &&
    typeof (e as { type: unknown }).type === 'string' &&
    (e as { type: string }).type in TYPE_TO_KIND
  );
}

function translateThrow(e: unknown): RenderError {
  if (isTemplateError(e)) {
    return new RenderError(TYPE_TO_KIND[e.type], e.message, { variable: e.variable, cause: e.originalError ?? e });
  }
  const message = e instanceof Error ? e.message : String(e);
  return new RenderError('unknown', message, { cause: e });
}

function toWarning(w: EigenpalTemplateError): RenderWarning {
  return { kind: TYPE_TO_KIND[w.type] ?? 'unknown', message: w.message, variable: w.variable };
}

// --- ZIP timestamp normalization ---
//
// docxtemplater's syncZip() rewrites each modified XML part via PizZip's
// `zip.file(name, content, {createFolders: true})` with no `date` option
// (docxtemplater.js syncZip). PizZip's prepareFileAttrs then defaults an
// absent date to `new Date()` — the wall-clock instant of the write — and
// encodes it into the ZIP DOS header, truncated to 2-second resolution
// (object.js: `dosTime |= date.getSeconds() / 2`). Two renders of
// byte-identical input land in the same output only when they happen to
// fall in the same 2-second bucket; otherwise the produced .docx differs by
// exactly the stored timestamp on the parts docxtemplater touched. That
// silently breaks the forensic reconstruction endpoint
// (internal/modules/render/fanout/reconstruction.go), whose entire purpose
// is proving byte-for-byte reproducibility.
//
// docxtemplater exposes no `date` passthrough and PizZip sets dates
// per-entry at `.file()` time with no global override, so there is no
// upstream escape hatch. This module is the one server-side eigenpal
// touchpoint, so we normalize here: re-open the produced buffer with
// jszip, stamp every entry (not just the hash) to a fixed constant date,
// and re-serialize with the same DEFLATE/level-6 settings the engine
// itself uses. The result is byte-identical by construction for identical
// input, not merely hash-stable.
//
// DOS ZIP timestamps cannot represent anything earlier than 1980-01-01, so
// that floor doubles as an unambiguous "this is a synthetic sentinel, not a
// real save time" marker.
const NORMALIZED_ZIP_DATE = new Date('1980-01-01T00:00:00.000Z');

interface OrderedZipEntry {
  path: string;
  entry: JSZip.JSZipObject;
}

/**
 * Re-serializes a docx (ZIP) buffer with every entry's stored date pinned to
 * NORMALIZED_ZIP_DATE, so identical decompressed content always produces
 * identical bytes on disk.
 *
 * Entry order is preserved exactly as jszip enumerates it from the source
 * buffer (insertion order of the ZIP's own central directory) — this pass
 * only overwrites the date field per entry, it never reorders or drops one,
 * so a second normalization of the same source buffer enumerates the same
 * order every time.
 *
 * createFolders is explicitly disabled on every write: with jszip's default
 * (createFolders: true), adding a file whose parent directory has no
 * existing entry auto-creates one via the same `date || new Date()`
 * fallback this function exists to eliminate. Every entry — files and any
 * literal directory records the source zip already contains — is
 * re-added explicitly here, so no implicit, un-normalized entry can slip
 * in via that path.
 */
async function normalizeZipTimestamps(buffer: ArrayBuffer): Promise<Uint8Array> {
  const source = await JSZip.loadAsync(buffer);

  // jszip sets the archive-level comment on the instance at load time
  // (lib/load.js: `zip.comment = zipEntries.zipComment`) and reads it back at
  // generate time (lib/object.js: `opts.comment || this.comment || ""`), but
  // its bundled index.d.ts never declares `comment` on the JSZip interface —
  // only on JSZipObject and on the option bags. The narrow read below exists
  // for that type gap alone; without it the archive comment is silently
  // dropped on re-serialization.
  const archiveComment = (source as unknown as { comment?: string | null }).comment ?? undefined;

  const ordered: OrderedZipEntry[] = [];
  source.forEach((relativePath, entry) => {
    ordered.push({ path: relativePath, entry });
  });

  const normalized = new JSZip();
  for (const { path, entry } of ordered) {
    if (entry.dir) {
      normalized.file(path, null, {
        dir: true,
        createFolders: false,
        date: NORMALIZED_ZIP_DATE,
        comment: entry.comment ?? undefined,
        unixPermissions: entry.unixPermissions ?? undefined,
        dosPermissions: entry.dosPermissions ?? undefined,
      });
      continue;
    }
    const content = await entry.async('uint8array');
    normalized.file(path, content, {
      createFolders: false,
      date: NORMALIZED_ZIP_DATE,
      comment: entry.comment ?? undefined,
      compression: 'DEFLATE',
      compressionOptions: { level: 6 },
      unixPermissions: entry.unixPermissions ?? undefined,
      dosPermissions: entry.dosPermissions ?? undefined,
    });
  }

  return normalized.generateAsync({
    type: 'uint8array',
    compression: 'DEFLATE',
    compressionOptions: { level: 6 },
    comment: archiveComment,
    platform: 'DOS',
  });
}

/** Build a TemplateProcessor over an injected engine (production passes the real eigenpal engine). */
export function makeTemplateProcessor(engine: EigenpalEngine): TemplateProcessor {
  return {
    async processTemplate(docx, variables, opts) {
      let raw: EigenpalRawResult;
      try {
        raw = engine.processTemplateDetailed(docx, variables, { nullGetter: opts?.nullGetter ?? 'empty' });
      } catch (e) {
        throw translateThrow(e);
      }
      let buffer: Uint8Array;
      try {
        buffer = await normalizeZipTimestamps(raw.buffer);
      } catch (e) {
        throw translateThrow(e);
      }
      return {
        buffer,
        replacedVariables: raw.replacedVariables ?? [],
        unreplacedVariables: raw.unreplacedVariables ?? [],
        warnings: (raw.warnings ?? []).map(toWarning),
      };
    },
  };
}

/** Thin bridge from the real eigenpal API (warnings: string[]) to EigenpalEngine. */
const realEngine: EigenpalEngine = {
  processTemplateDetailed(docx, variables, opts) {
    const result = processTemplateDetailed(docx, variables, opts);
    return {
      buffer: result.buffer,
      replacedVariables: result.replacedVariables,
      unreplacedVariables: result.unreplacedVariables,
      // The real eigenpal emits string[] warnings; we surface them as unknown-kind
      // structured warnings so the rest of the pipeline stays uniform.
      warnings: (result.warnings ?? []).map((msg) => ({
        type: 'unknown' as const,
        message: msg,
      })),
    };
  },
};

/** Production singleton backed by the real eigenpal headless engine. */
export const eigenpalTemplateProcessor: TemplateProcessor = makeTemplateProcessor(realEngine);
