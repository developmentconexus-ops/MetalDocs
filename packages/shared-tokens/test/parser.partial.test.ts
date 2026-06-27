import { describe, expect, it } from 'vitest';
import { parseDocxTokens } from '../src/parser';
import { makeDocx } from './fixtures';

const doc = (inner: string) => `<?xml version="1.0"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body><w:p><w:r><w:t>${inner}</w:t></w:r></w:p></w:body>
</w:document>`;

describe('parseDocxTokens partial/control handling', () => {
  it('does NOT emit a token for a partial tag, and flags it as an error', async () => {
    const res = await parseDocxTokens(await makeDocx(doc('Ref {>frag} here')));
    expect(res.tokens.find((t) => t.ident === 'frag')).toBeUndefined();
    expect(res.errors.some((e) => e.type === 'unsupported_construct')).toBe(true);
  });

  it('still emits valid var tokens unchanged', async () => {
    const res = await parseDocxTokens(await makeDocx(doc('Hi {client_name}')));
    expect(res.tokens.map((t) => [t.kind, t.ident])).toEqual([['var', 'client_name']]);
  });
});
