// The framework-free (`.`) door: the ONLY server-side eigenpal touchpoint.
// No React, no DOM — safe to inline into the docx-renderer Node bundle.
import { processTemplateDetailed } from '@eigenpal/docx-editor-core/headless';

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
  ): RenderResult;
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

/** Build a TemplateProcessor over an injected engine (production passes the real eigenpal engine). */
export function makeTemplateProcessor(engine: EigenpalEngine): TemplateProcessor {
  return {
    processTemplate(docx, variables, opts) {
      let raw: EigenpalRawResult;
      try {
        raw = engine.processTemplateDetailed(docx, variables, { nullGetter: opts?.nullGetter ?? 'empty' });
      } catch (e) {
        throw translateThrow(e);
      }
      return {
        buffer: new Uint8Array(raw.buffer),
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
