import Ajv from 'ajv';
import { z } from 'zod';
import { getObjectBuffer, putObjectBuffer } from '../s3.js';
import { processDocx } from '../render/processDocx.js';
const BodySchema = z.object({
    template_docx_key: z.string().min(1),
    schema_key: z.string().min(1),
    form_data: z.record(z.unknown()),
    output_key: z.string().min(1),
});
const DOCX_MIME = 'application/vnd.openxmlformats-officedocument.wordprocessingml.document';
export function registerRenderRoutes(app, env, s3Factory) {
    app.post('/render/docx', async (req, reply) => {
        const parsed = BodySchema.safeParse(req.body);
        if (!parsed.success) {
            return reply.code(400).send({ error: 'bad_request', details: parsed.error.format() });
        }
        const { template_docx_key, schema_key, form_data, output_key } = parsed.data;
        const client = s3Factory();
        const [templateBuf, schemaBuf] = await Promise.all([
            getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, template_docx_key),
            getObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, schema_key),
        ]);
        let schema;
        try {
            schema = JSON.parse(schemaBuf.toString('utf8'));
        }
        catch {
            return reply.code(422).send({ error: 'schema_invalid', message: 'schema is not valid JSON' });
        }
        const ajv = new Ajv({ strict: false });
        const validate = ajv.compile(schema);
        if (!validate(form_data)) {
            return reply.code(422).send({ error: 'form_data_invalid', errors: validate.errors ?? [] });
        }
        const { buffer, contentHash, unreplacedVars } = await processDocx(new Uint8Array(templateBuf), form_data);
        await putObjectBuffer(client, env.DOCGEN_V2_S3_BUCKET, output_key, Buffer.from(buffer), DOCX_MIME);
        return reply.code(200).send({
            output_key,
            content_hash: contentHash,
            size_bytes: buffer.byteLength,
            warnings: [],
            unreplaced_vars: unreplacedVars,
        });
    });
}
