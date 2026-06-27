import { describe, expect, it } from 'vitest';
import { scanText } from '../src/grammar';

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
