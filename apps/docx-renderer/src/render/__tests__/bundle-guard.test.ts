import { describe, it, expect } from 'vitest';
import { readFileSync, existsSync } from 'node:fs';
import { resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const META = resolve(__dirname, '../../../dist/meta.json');

describe('server bundle is framework-free', () => {
  it('inlines @metaldocs/eigenpal-adapter but never React or eigenpal-react', () => {
    if (!existsSync(META)) {
      throw new Error('dist/meta.json missing — run `npm run build -w @metaldocs/docx-renderer` before this test');
    }
    const meta = JSON.parse(readFileSync(META, 'utf8')) as { inputs: Record<string, unknown> };
    const inputs = Object.keys(meta.inputs);
    // The adapter source IS inlined (it is a @metaldocs/* workspace pkg).
    expect(inputs.some((p) => p.includes('eigenpal-adapter/src/index'))).toBe(true);
    // React / eigenpal-react must NOT be inlined into the Node bundle.
    const reactLeak = inputs.filter(
      (p) => /node_modules[\\/](react|react-dom)[\\/]/.test(p) || p.includes('docx-editor-react'),
    );
    expect(reactLeak).toEqual([]);
  });
});
