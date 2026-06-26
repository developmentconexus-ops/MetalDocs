import type { FastifyInstance } from 'fastify';
import { z } from 'zod';
import type { Client } from 'minio';
import type { Env } from '../env.js';
import { getObjectBuffer, putObjectBuffer } from '../s3.js';
import { fanout } from '../render/fanout.js';
import { RenderError, type RenderErrorKind } from '@metaldocs/eigenpal-adapter';

const BodySchema = z.object({
  body_docx_s3_key: z.string().min(1),
  tenant_id: z.string().min(1),
  revision_id: z.string().min(1),
  placeholder_values: z.record(z.string()),
  composition_config: z.object({
    header_sub_blocks: z.array(z.string()).default([]),
    footer_sub_blocks: z.array(z.string()).default([]),
    sub_block_params: z.record(z.record(z.unknown())).default({}),
  }).default({ header_sub_blocks: [], footer_sub_blocks: [], sub_block_params: {} }),
  resolved_values: z.record(z.unknown()),
});

const DOCX_MIME =
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document';

function frozenDocxKey(tenantId: string, revisionId: string): string {
  return `tenants/${tenantId}/revisions/${revisionId}/frozen.docx`;
}

const KIND_TO_STATUS: Record<RenderErrorKind, number> = {
  template_parse: 400,
  undefined_variable: 400,
  template_render: 422,
  unknown: 500,
};

export function renderErrorToHttp(e: RenderError): {
  status: number;
  body: { error: 'render_failed'; kind: RenderErrorKind; message: string; variable?: string };
} {
  return {
    status: KIND_TO_STATUS[e.kind],
    body: {
      error: 'render_failed',
      kind: e.kind,
      message: e.message,
      ...(e.variable ? { variable: e.variable } : {}),
    },
  };
}

export function registerFanoutRoute(
  app: FastifyInstance,
  env: Env,
  s3Factory: () => Client,
): void {
  app.post('/render/fanout', async (req, reply) => {
    const parsed = BodySchema.safeParse(req.body);
    if (!parsed.success) {
      return reply
        .code(400)
        .send({ error: 'bad_request', details: parsed.error.format() });
    }
    const {
      body_docx_s3_key,
      tenant_id,
      revision_id,
      placeholder_values,
      composition_config,
      resolved_values,
    } = parsed.data;

    const output_key = frozenDocxKey(tenant_id, revision_id);

    const client = s3Factory();
    let bodyBuf: Buffer;
    try {
      bodyBuf = await getObjectBuffer(
        client,
        env.DOCX_RENDERER_S3_BUCKET,
        body_docx_s3_key,
      );
    } catch (err) {
      // A missing/unreadable body key is a caller error (bad input), not a server
      // fault — classify it instead of leaking a raw minio 500.
      const code = (err as { code?: string })?.code;
      if (code === 'NoSuchKey' || code === 'NotFound') {
        req.log.warn({ body_docx_s3_key }, 'body docx not found');
        return reply
          .code(404)
          .send({ error: 'body_docx_not_found', key: body_docx_s3_key });
      }
      throw err;
    }

    let result;
    try {
      result = await fanout({
        bodyDocx: new Uint8Array(bodyBuf),
        placeholderValues: placeholder_values,
        compositionConfig: composition_config,
        resolvedValues: resolved_values,
      });
    } catch (err) {
      if (err instanceof RenderError) {
        const { status, body } = renderErrorToHttp(err);
        req.log.warn({ kind: body.kind, variable: body.variable }, 'render failed');
        return reply.code(status).send(body);
      }
      throw err;
    }

    await putObjectBuffer(
      client,
      env.DOCX_RENDERER_S3_BUCKET,
      output_key,
      Buffer.from(result.buffer),
      DOCX_MIME,
    );

    return reply.code(200).send({
      content_hash: result.contentHash,
      final_docx_s3_key: output_key,
      unreplaced_vars: result.unreplacedVars,
      size_bytes: result.buffer.byteLength,
    });
  });
}
