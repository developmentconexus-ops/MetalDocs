// Mirrors internal/platform/observability/otel.go's SetupOTel contract so the
// 4th binary (docx-renderer, Node/TS) follows the SAME env conventions as the
// 3 Go binaries: unconfigured (no OTEL_TRACES_EXPORTER and no
// OTEL_EXPORTER_OTLP_ENDPOINT) installs nothing and the global tracer stays
// the OTel no-op default — zero overhead, no log spam, no crash. Configured
// (OTEL_TRACES_EXPORTER=otlp|console, or OTEL_EXPORTER_OTLP_ENDPOINT set)
// installs a real NodeTracerProvider exporting via OTLP/HTTP. This closes
// T-010 (wiki/modules/editor-ui-eigenpal-tech-debt.md) — REQ-OBS-3's last
// unpropagated hop (worker → docx-renderer).
import { trace, type Tracer } from '@opentelemetry/api';
import { W3CTraceContextPropagator } from '@opentelemetry/core';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { resourceFromAttributes } from '@opentelemetry/resources';
import { ConsoleSpanExporter, SimpleSpanProcessor, BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';
import { NodeTracerProvider } from '@opentelemetry/sdk-trace-node';
import { ATTR_SERVICE_NAME } from '@opentelemetry/semantic-conventions';

const DEFAULT_SERVICE_NAME = 'docx-renderer';

export interface OTelHandle {
  readonly enabled: boolean;
  readonly propagator: W3CTraceContextPropagator;
  shutdown(): Promise<void>;
}

const noopPropagator = new W3CTraceContextPropagator();

function noopHandle(): OTelHandle {
  return {
    enabled: false,
    propagator: noopPropagator,
    shutdown: async () => {},
  };
}

/**
 * Installs a NodeTracerProvider ONLY when the operator configured an exporter
 * (OTEL_TRACES_EXPORTER or OTEL_EXPORTER_OTLP_ENDPOINT set) — same gate as the
 * Go SetupOTel. Unconfigured => returns a disabled no-op handle; the global
 * @opentelemetry/api tracer stays the built-in no-op, so span creation calls
 * anywhere in the renderer cost nothing and never throw.
 *
 * serviceName defaults to "docx-renderer"; OTEL_SERVICE_NAME overrides it,
 * matching the Go convention (SetupOTel's serviceName + OTEL_SERVICE_NAME).
 */
export function setupOTel(serviceName: string = DEFAULT_SERVICE_NAME): OTelHandle {
  const exporterEnv = (process.env.OTEL_TRACES_EXPORTER ?? '').trim().toLowerCase();
  const endpoint = (process.env.OTEL_EXPORTER_OTLP_ENDPOINT ?? '').trim();

  // OTEL_TRACES_EXPORTER=none is the OTel-spec "explicitly off" value; matches
  // Go's short-circuit even when an endpoint is also set.
  if (exporterEnv === 'none') {
    return noopHandle();
  }
  if (exporterEnv === '' && endpoint === '') {
    return noopHandle();
  }

  const resolvedServiceName = (process.env.OTEL_SERVICE_NAME ?? '').trim() || serviceName;

  const exporter =
    exporterEnv === 'console'
      ? new ConsoleSpanExporter()
      : new OTLPTraceExporter(endpoint ? { url: `${endpoint.replace(/\/$/, '')}/v1/traces` } : {});

  // Console exporter should flush synchronously (SimpleSpanProcessor) so a
  // dev/test run sees spans immediately, mirroring the Go autoexport
  // console path's directness; OTLP uses BatchSpanProcessor for real export.
  const processor =
    exporterEnv === 'console' ? new SimpleSpanProcessor(exporter) : new BatchSpanProcessor(exporter);

  const provider = new NodeTracerProvider({
    resource: resourceFromAttributes({ [ATTR_SERVICE_NAME]: resolvedServiceName }),
    spanProcessors: [processor],
  });
  provider.register({ propagator: new W3CTraceContextPropagator() });

  return {
    enabled: true,
    propagator: new W3CTraceContextPropagator(),
    shutdown: () => provider.shutdown(),
  };
}

/** The renderer's tracer — a thin wrapper so callers don't import @opentelemetry/api directly. */
export function getTracer(): Tracer {
  return trace.getTracer('docx-renderer');
}
