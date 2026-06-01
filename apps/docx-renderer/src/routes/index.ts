import type { FastifyInstance } from 'fastify';
import type { Env } from '../env.js';
import type { Client } from 'minio';
import { registerFanoutRoute } from './fanout.js';

export function registerRoutes(app: FastifyInstance, env: Env, s3Factory: () => Client): void {
  registerFanoutRoute(app, env, s3Factory);
}
