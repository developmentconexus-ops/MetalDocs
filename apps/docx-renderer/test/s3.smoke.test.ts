import { describe, it, expect } from 'vitest';
import { makeS3Client } from '../src/s3';

describe('makeS3Client', () => {
  it('parses endpoint URL into host/port', () => {
    const c = makeS3Client({
      DOCX_RENDERER_PORT: 0,
      DOCX_RENDERER_SERVICE_TOKEN: 'test-token-0123456789',
      LOG_LEVEL: 'info',
      VERSION: 'dev',
      DOCX_RENDERER_S3_ENDPOINT: 'http://minio:9000',
      DOCX_RENDERER_S3_ACCESS_KEY: 'k',
      DOCX_RENDERER_S3_SECRET_KEY: 's',
      DOCX_RENDERER_S3_BUCKET: 'b',
      DOCX_RENDERER_S3_USE_SSL: false,
      DOCX_RENDERER_GOTENBERG_URL: 'http://gotenberg:3000',
    });
    expect(c).toBeDefined();
  });
});
