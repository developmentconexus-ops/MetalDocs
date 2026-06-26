import { createHash, timingSafeEqual } from 'node:crypto';
import type { FastifyInstance } from 'fastify';

// Hash both sides to a fixed 32-byte width before comparing so the compare is
// constant-time regardless of input length — a raw length check would leak the
// token length via timing.
function safeCompare(a: string, b: string): boolean {
  const ha = createHash('sha256').update(a).digest();
  const hb = createHash('sha256').update(b).digest();
  return timingSafeEqual(ha, hb);
}

export function registerServiceAuth(app: FastifyInstance, token: string): void {
  app.addHook('onRequest', async (req, reply) => {
    if (req.url === '/health') return;
    const header = req.headers['x-service-token'];
    if (typeof header !== 'string' || !safeCompare(header, token)) {
      return reply.code(401).send({ error: 'unauthorized' });
    }
  });
}