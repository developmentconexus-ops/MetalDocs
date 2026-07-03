import { describe, it, expect, beforeAll, afterAll } from 'vitest';
import { trace, context as otelContext } from '@opentelemetry/api';
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { InMemorySpanExporter, SimpleSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { W3CTraceContextPropagator } from '@opentelemetry/core';
import { extractTraceContext, withRenderSpan } from '../tracing.js';

// Real SDK + in-memory exporter (same pattern as the Go side's tracetest
// SpanRecorder) so the propagation seam (traceparent header -> parented
// span) is proven end-to-end without a network exporter.
const exporter = new InMemorySpanExporter();
let provider: NodeTracerProvider;

beforeAll(() => {
  provider = new NodeTracerProvider({ spanProcessors: [new SimpleSpanProcessor(exporter)] });
  provider.register({ propagator: new W3CTraceContextPropagator() });
});

afterAll(async () => {
  await provider.shutdown();
});

describe('extractTraceContext', () => {
  it('parents a span under an inbound traceparent header', async () => {
    // Produce a real traceparent by starting+ending an upstream span.
    const upstreamTracer = trace.getTracer('upstream-test');
    const upstreamSpan = upstreamTracer.startSpan('upstream-call');
    const upstreamTraceId = upstreamSpan.spanContext().traceId;
    const carrier: Record<string, string> = {};
    new W3CTraceContextPropagator().inject(trace.setSpan(otelContext.active(), upstreamSpan), carrier, {
      set: (c, k, v) => { c[k] = v; },
    });
    upstreamSpan.end();
    exporter.reset();

    const parentCtx = extractTraceContext({ traceparent: carrier['traceparent'] });
    await withRenderSpan(parentCtx, {}, async () => 'ok');

    const spans = exporter.getFinishedSpans();
    expect(spans).toHaveLength(1);
    expect(spans[0]!.name).toBe('docx_renderer.render_fanout');
    expect(spans[0]!.spanContext().traceId).toBe(upstreamTraceId);
  });

  it('produces a root span (does not throw) when no traceparent header is present', async () => {
    exporter.reset();
    const parentCtx = extractTraceContext({});
    await withRenderSpan(parentCtx, {}, async () => 'ok');

    const spans = exporter.getFinishedSpans();
    expect(spans).toHaveLength(1);
    // Root span: traceId is freshly generated, non-zero.
    expect(spans[0]!.spanContext().traceId).toMatch(/^[0-9a-f]{32}$/);
  });

  it('ignores a malformed traceparent header instead of throwing', () => {
    expect(() => extractTraceContext({ traceparent: 'not-a-real-traceparent' })).not.toThrow();
  });
});

describe('withRenderSpan', () => {
  it('sets attributes.tenant_id only when provided', async () => {
    exporter.reset();
    const parentCtx = extractTraceContext({});
    await withRenderSpan(parentCtx, { tenant_id: 'tenant-42' }, async () => 'ok');

    const spans = exporter.getFinishedSpans();
    expect(spans[0]!.attributes.tenant_id).toBe('tenant-42');
  });

  it('records the exception and rethrows on failure', async () => {
    exporter.reset();
    const parentCtx = extractTraceContext({});
    const boom = new Error('render exploded');

    await expect(
      withRenderSpan(parentCtx, {}, async () => {
        throw boom;
      }),
    ).rejects.toThrow('render exploded');

    const spans = exporter.getFinishedSpans();
    expect(spans).toHaveLength(1);
    expect(spans[0]!.status.code).toBe(2); // SpanStatusCode.ERROR
    expect(spans[0]!.events.some((e) => e.name === 'exception')).toBe(true);
  });
});
