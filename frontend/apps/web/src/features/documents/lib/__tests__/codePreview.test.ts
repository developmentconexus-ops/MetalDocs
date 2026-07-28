import { describe, expect, it } from 'vitest';

import { CODE_PLACEHOLDER, formatCodePreview } from '../codePreview';

describe('formatCodePreview', () => {
  it('returns ???-???-??? when profile and area are not selected (not ready)', () => {
    const result = formatCodePreview({
      ready: false,
      isLoading: false,
      code: null,
      profileCode: null,
      areaCode: null,
    });
    expect(result).toEqual({
      text: `${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}`,
      state: 'unselected',
    });
  });

  it('returns profileCode-areaCode-… when ready and loading', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: true,
      code: null,
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result).toEqual({ text: 'POP-GENERAL-…', state: 'loading' });
  });

  it('returns the code verbatim when ready, not loading, and code is present', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: false,
      code: 'POP-GENERAL-042',
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result).toEqual({ text: 'POP-GENERAL-042', state: 'ready' });
  });

  it('returns profileCode-areaCode-??? when ready, not loading, and code is null', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: false,
      code: null,
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result).toEqual({
      text: `POP-GENERAL-${CODE_PLACEHOLDER}`,
      state: 'unavailable',
    });
  });

  // F-QA4-1: a failed preview must be distinguishable from "nothing selected"
  // and from "still loading" — all three used to collapse into the ??? text.
  it('reports the error state when the preview-code query failed', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: false,
      isError: true,
      code: null,
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result).toEqual({
      text: `POP-GENERAL-${CODE_PLACEHOLDER}`,
      state: 'error',
    });
  });

  it('keeps loading distinct from error while the query is still in flight', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: true,
      isError: true,
      code: null,
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result.state).toBe('loading');
  });

  it('prefers the error state over a stale resolved code', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: false,
      isError: true,
      code: 'POP-GENERAL-042',
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result.state).toBe('error');
  });
});
