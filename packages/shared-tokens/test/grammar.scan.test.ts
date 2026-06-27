import { describe, expect, it } from 'vitest';
import { scanText, detectTokens } from '../src/grammar';

describe('scanText', () => {
  it('extracts a plain variable tag', () => {
    expect(scanText('Hello {name} world')).toEqual([
      { raw: '{name}', kind: 'var', inner: 'name', start: 6, end: 12 },
    ]);
  });

  it('classifies control prefixes # ^ / >', () => {
    const kinds = scanText('{#s}{^i}{/s}{>p}').map((t) => [t.kind, t.inner]);
    expect(kinds).toEqual([
      ['section', 's'],
      ['inverted', 'i'],
      ['closing', 's'],
      ['partial', 'p'],
    ]);
  });

  it('trims inner whitespace', () => {
    expect(scanText('{ name }')[0]).toMatchObject({ kind: 'var', inner: 'name' });
  });

  it('ignores empty braces and unclosed braces', () => {
    expect(scanText('{} {oops')).toEqual([]);
  });

  it('captures multiple tags with correct offsets', () => {
    const t = scanText('{a}{b}');
    expect(t.map((x) => [x.inner, x.start, x.end])).toEqual([
      ['a', 0, 3],
      ['b', 3, 6],
    ]);
  });
});

describe('detectTokens', () => {
  it('returns valid var tokens', () => {
    expect(detectTokens('{doc_code} and {author}')).toEqual([
      { name: 'doc_code', kind: 'var', valid: true, reserved: false },
      { name: 'author', kind: 'var', valid: true, reserved: false },
    ]);
  });

  it('SURFACES invalid-ident var tokens (no silent drop)', () => {
    // {a.b} substitutes/acts at freeze but is not snake_case — must be visible.
    expect(detectTokens('{a.b}')).toEqual([
      { name: 'a.b', kind: 'var', valid: false, reserved: false },
    ]);
  });

  it('flags reserved idents', () => {
    expect(detectTokens('{tenant_id}')).toEqual([
      { name: 'tenant_id', kind: 'var', valid: false, reserved: true },
    ]);
  });

  it('excludes control tags (#/^/>) from variable detection', () => {
    expect(detectTokens('{#loop}{name}{/loop}{>partial}')).toEqual([
      { name: 'name', kind: 'var', valid: true, reserved: false },
    ]);
  });

  it('dedups first-seen by name', () => {
    expect(detectTokens('{x} {x}').map((t) => t.name)).toEqual(['x']);
  });
});
