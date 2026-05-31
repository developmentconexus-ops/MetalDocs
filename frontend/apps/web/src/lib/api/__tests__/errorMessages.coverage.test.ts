import { describe, expect, it } from 'vitest';

import { errorMessages } from '../errorMessages';
import generated from '../error-codes.generated.json';

interface Snapshot {
  codes: string[];
}

const snapshot = generated as Snapshot;

describe('errorMessages coverage', () => {
  it('has a Portuguese mapping for every backend Problem code', () => {
    const missing = snapshot.codes.filter((code) => !(code in errorMessages));

    expect(missing, `Missing PT-BR mapping for backend code(s). Run \`go run ./scripts/dump-error-codes.go\` if codes were just added, then add entries to frontend/apps/web/src/lib/api/errorMessages.ts. Missing: ${missing.join(', ')}`).toEqual([]);
  });

  it('does not declare stale mappings for codes no longer emitted by the backend', () => {
    // FE-synthetic codes that never appear in backend Problem responses but
    // are dispatched by the API client itself (see lib/api/client.ts).
    const frontendOnly = new Set<string>(['authn.expired']);

    const known = new Set(snapshot.codes);
    const stale = Object.keys(errorMessages).filter(
      (code) => !known.has(code) && !frontendOnly.has(code),
    );

    expect(stale, `errorMessages.ts maps codes the backend no longer emits. Remove them or regenerate the snapshot. Stale: ${stale.join(', ')}`).toEqual([]);
  });
});
