import { describe, expect, it } from 'vitest';
import * as Tokens from '../src/index';

describe('shared-tokens barrel', () => {
  it('exports the browser-safe grammar surface', () => {
    expect(typeof Tokens.scanText).toBe('function');
    expect(typeof Tokens.detectTokens).toBe('function');
    expect(typeof Tokens.isValidIdent).toBe('function');
  });

  it('does NOT re-export the JSZip-backed docx parser (browser-bundle purity)', () => {
    expect((Tokens as Record<string, unknown>).parseDocxTokens).toBeUndefined();
  });
});
