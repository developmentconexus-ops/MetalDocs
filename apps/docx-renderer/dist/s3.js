import { Client } from 'minio';
export function makeS3Client(env) {
    const url = new URL(env.DOCGEN_V2_S3_ENDPOINT);
    return new Client({
        endPoint: url.hostname,
        port: Number(url.port || (env.DOCGEN_V2_S3_USE_SSL ? 443 : 80)),
        useSSL: env.DOCGEN_V2_S3_USE_SSL,
        accessKey: env.DOCGEN_V2_S3_ACCESS_KEY,
        secretKey: env.DOCGEN_V2_S3_SECRET_KEY,
    });
}
export async function getObjectBuffer(client, bucket, key) {
    const stream = await client.getObject(bucket, key);
    const chunks = [];
    for await (const c of stream)
        chunks.push(Buffer.isBuffer(c) ? c : Buffer.from(c));
    return Buffer.concat(chunks);
}
export async function putObjectBuffer(client, bucket, key, data, contentType) {
    await client.putObject(bucket, key, data, data.byteLength, { 'Content-Type': contentType });
}
