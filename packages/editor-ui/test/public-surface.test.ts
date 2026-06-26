import { describe, it, expect } from 'vitest';
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';

const here = dirname(fileURLToPath(import.meta.url));

/**
 * ACL guard: the public surface of @metaldocs/editor-ui must never name the
 * vendor. If a vendor type leaks into index.ts or types.ts, a vendor swap stops
 * being a one-package change. Mapping files (comment-mapping.ts, MetalDocsEditor.tsx)
 * are INSIDE the wall and may import @eigenpal — they are not part of this check.
 */
describe('editor-ui public surface', () => {
  for (const file of ['../src/index.ts', '../src/types.ts']) {
    it(`${file} contains no @eigenpal reference`, () => {
      const contents = readFileSync(resolve(here, file), 'utf8');
      expect(contents).not.toMatch(/@eigenpal/);
    });
  }
});
