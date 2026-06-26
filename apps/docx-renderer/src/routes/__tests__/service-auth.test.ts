import { describe, expect, test, beforeAll, afterAll } from 'vitest';
import type { FastifyInstance } from 'fastify';

// The X-Service-Token wall is the sidecar's only authentication boundary. These
// tests guard the reject paths and the /health exemption.
const TOKEN = 'test-token-service-auth';

let app: FastifyInstance;

beforeAll(async () => {
  process.env.DOCX_RENDERER_SERVICE_TOKEN = TOKEN;
  process.env.DOCX_RENDERER_S3_ACCESS_KEY = 'key';
  process.env.DOCX_RENDERER_S3_SECRET_KEY = 'sec';
  const { buildApp } = await import('../../index.js');
  app = await buildApp();
}, 30000);

afterAll(async () => {
  if (app) await app.close();
});

describe('service-auth wall', () => {
  test('rejects a protected route with no token', async () => {
    const res = await app.inject({ method: 'POST', url: '/render/fanout', payload: {} });
    expect(res.statusCode).toBe(401);
    expect(res.json().error).toBe('unauthorized');
  });

  test('rejects a protected route with a wrong token', async () => {
    const res = await app.inject({
      method: 'POST',
      url: '/render/fanout',
      headers: { 'x-service-token': 'wrong-token' },
      payload: {},
    });
    expect(res.statusCode).toBe(401);
  });

  test('exempts /health from auth', async () => {
    const res = await app.inject({ method: 'GET', url: '/health' });
    expect(res.statusCode).toBe(200);
    expect(res.json().status).toBe('ok');
  });
});
