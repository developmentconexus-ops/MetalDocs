import { fileURLToPath } from 'node:url';
import { resolve } from 'node:path';
import Fastify, { type FastifyInstance } from 'fastify';
import { loadEnv } from './env.js';
import { registerServiceAuth } from './service-auth.js';
import { registerRoutes } from './routes/index.js';
import { makeS3Client } from './s3.js';
import { setupOTel } from './observability/otel.js';

export async function buildApp(): Promise<FastifyInstance> {
  const env = loadEnv();
  const app = Fastify({ logger: { level: env.LOG_LEVEL } });

  registerServiceAuth(app, env.DOCX_RENDERER_SERVICE_TOKEN);

  app.get('/health', async () => ({ status: 'ok', version: env.VERSION }));

  let cachedClient: ReturnType<typeof makeS3Client> | null = null;
  const s3Factory = () => (cachedClient ??= makeS3Client(env));

  registerRoutes(app, env, s3Factory);
  return app;
}

if (resolve(fileURLToPath(import.meta.url)) === resolve(process.argv[1])) {
  // Installed before buildApp() so the fanout route's tracer is ready before
  // the first request; inert (no-op, no log spam) when OTEL_TRACES_EXPORTER /
  // OTEL_EXPORTER_OTLP_ENDPOINT are unset, matching the 3 Go binaries'
  // SetupOTel contract (REQ-OBS-3).
  const otel = setupOTel();
  const env = loadEnv();
  buildApp().then((app) => {
    const shutdown = async (): Promise<void> => {
      await app.close().catch(() => {});
      await otel.shutdown().catch(() => {});
      process.exit(0);
    };
    process.once('SIGTERM', () => { void shutdown(); });
    process.once('SIGINT', () => { void shutdown(); });

    app.listen({ port: env.DOCX_RENDERER_PORT, host: '0.0.0.0' })
       .catch((err) => { app.log.fatal(err); process.exit(1); });
  });
}
