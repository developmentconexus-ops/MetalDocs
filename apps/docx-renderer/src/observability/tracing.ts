// The render/fanout route's tracing seam: extract an inbound W3C traceparent
// (propagated by the worker→renderer HTTP hop per REQ-OBS-3) and run render
// work inside a child span parented to it. Kept separate from otel.ts (SDK
// bootstrap) so this file is pure propagation/span-wrapping logic, easy to
// unit-test without spinning up a real exporter.
import {
  context as otelContext,
  propagation,
  SpanStatusCode,
  type Context,
} from '@opentelemetry/api';
import { getTracer } from './otel.js';

/**
 * Extracts a W3C traceparent (+ tracestate, if present) from an inbound
 * header carrier into a new otel Context. When no valid traceparent header is
 * present, or OTel is not installed (no-op API), this returns a context that
 * yields a root span — never throws, matching SetupOTel's "unconfigured is
 * inert" contract.
 */
export function extractTraceContext(headers: Record<string, string | string[] | undefined>): Context {
  const carrier: Record<string, string> = {};
  const traceparent = headers['traceparent'];
  if (typeof traceparent === 'string') carrier['traceparent'] = traceparent;
  const tracestate = headers['tracestate'];
  if (typeof tracestate === 'string') carrier['tracestate'] = tracestate;
  return propagation.extract(otelContext.active(), carrier);
}

export interface RenderSpanAttributes {
  /** Present only when already available in the caller's structured logs — no new sensitive data. */
  tenant_id?: string;
}

/**
 * Runs fn inside a span named "docx_renderer.render_fanout", parented to the
 * extracted trace context (root span when none was present). Records the
 * error and sets Error status on throw, then rethrows unchanged — callers'
 * existing error handling (RenderError -> HTTP classification) is untouched.
 */
export async function withRenderSpan<T>(
  parentCtx: Context,
  attrs: RenderSpanAttributes,
  fn: () => Promise<T>,
): Promise<T> {
  const tracer = getTracer();
  return tracer.startActiveSpan(
    'docx_renderer.render_fanout',
    { attributes: attrs.tenant_id ? { tenant_id: attrs.tenant_id } : undefined },
    parentCtx,
    async (span) => {
      try {
        const result = await fn();
        span.setStatus({ code: SpanStatusCode.OK });
        return result;
      } catch (err) {
        span.recordException(err instanceof Error ? err : String(err));
        span.setStatus({ code: SpanStatusCode.ERROR, message: err instanceof Error ? err.message : String(err) });
        throw err;
      } finally {
        span.end();
      }
    },
  );
}
