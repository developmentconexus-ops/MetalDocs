import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { trace } from '@opentelemetry/api';
import { setupOTel } from '../otel.js';

// Mirrors internal/platform/observability/otel_test.go's coverage shape:
// inert-by-default when unconfigured, installed when an exporter is set.
describe('setupOTel', () => {
  const savedEnv: Record<string, string | undefined> = {};

  beforeEach(() => {
    savedEnv.OTEL_TRACES_EXPORTER = process.env.OTEL_TRACES_EXPORTER;
    savedEnv.OTEL_EXPORTER_OTLP_ENDPOINT = process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
    savedEnv.OTEL_SERVICE_NAME = process.env.OTEL_SERVICE_NAME;
    delete process.env.OTEL_TRACES_EXPORTER;
    delete process.env.OTEL_EXPORTER_OTLP_ENDPOINT;
    delete process.env.OTEL_SERVICE_NAME;
  });

  afterEach(async () => {
    for (const [k, v] of Object.entries(savedEnv)) {
      if (v === undefined) delete process.env[k];
      else process.env[k] = v;
    }
  });

  it('is inert when no exporter env is set (no crash, no SDK install)', async () => {
    const handle = setupOTel('test-docx-renderer');
    expect(handle.enabled).toBe(false);
    await expect(handle.shutdown()).resolves.toBeUndefined();
  });

  it('is inert when OTEL_TRACES_EXPORTER=none, even with an endpoint set', async () => {
    process.env.OTEL_TRACES_EXPORTER = 'none';
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://localhost:4318';
    const handle = setupOTel('test-docx-renderer');
    expect(handle.enabled).toBe(false);
    await expect(handle.shutdown()).resolves.toBeUndefined();
  });

  it('installs a real provider when OTEL_TRACES_EXPORTER=console', async () => {
    process.env.OTEL_TRACES_EXPORTER = 'console';
    const handle = setupOTel('test-docx-renderer');
    expect(handle.enabled).toBe(true);
    await expect(handle.shutdown()).resolves.toBeUndefined();
  });

  it('installs a real provider when only OTEL_EXPORTER_OTLP_ENDPOINT is set', async () => {
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://localhost:4318';
    const handle = setupOTel('test-docx-renderer');
    expect(handle.enabled).toBe(true);
    await expect(handle.shutdown()).resolves.toBeUndefined();
  });

  it('never throws for an unreachable OTLP endpoint at setup time (export failures are async/logged, not thrown)', () => {
    process.env.OTEL_EXPORTER_OTLP_ENDPOINT = 'http://127.0.0.1:1'; // reserved, nothing listens here
    expect(() => setupOTel('test-docx-renderer')).not.toThrow();
  });
});

describe('getTracer via @opentelemetry/api', () => {
  it('returns a usable tracer even with no provider installed (global no-op)', () => {
    const tracer = trace.getTracer('docx-renderer');
    expect(() => tracer.startSpan('noop-span').end()).not.toThrow();
  });
});
