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
    expect(result).toBe(`${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}-${CODE_PLACEHOLDER}`);
  });

  it('returns profileCode-areaCode-… when ready and loading', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: true,
      code: null,
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result).toBe('POP-GENERAL-…');
  });

  it('returns the code verbatim when ready, not loading, and code is present', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: false,
      code: 'POP-GENERAL-042',
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result).toBe('POP-GENERAL-042');
  });

  it('returns profileCode-areaCode-??? when ready, not loading, and code is null', () => {
    const result = formatCodePreview({
      ready: true,
      isLoading: false,
      code: null,
      profileCode: 'POP',
      areaCode: 'GENERAL',
    });
    expect(result).toBe(`POP-GENERAL-${CODE_PLACEHOLDER}`);
  });
});
