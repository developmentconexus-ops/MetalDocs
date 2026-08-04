import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';

import { describe, expect, it } from 'vitest';

import generated from '../error-codes.generated.json';

/**
 * Guards the frontend's use of the backend Problem-code vocabulary.
 *
 * The existing coverage test (errorMessages.coverage.test.ts) checks the MESSAGE
 * MAP against the generated snapshot. It cannot see the other, more dangerous
 * half: code strings hard-coded into control flow. `if (error.code === 'X')`
 * does not break when the backend renames X — the branch simply stops firing,
 * forever, silently. The ADR 0089 sweep found ten such branches, including
 * `client.ts`'s session-expiry check, which had been comparing against a code the
 * backend no longer sent.
 *
 * So this test reads the same generated snapshot and asserts that every string
 * literal appearing in a POSITION THAT MEANS "a backend problem code" is one the
 * backend actually registers. It has no allow-list to maintain: the positions are
 * matched syntactically, and the vocabulary comes from the generator.
 *
 * When this fails, the fix is never to add an exception — it is to look up the
 * code's new name in the backend catalog (wiki/references/problem-codes.md, also
 * generated) and use that.
 */

const SRC = join(__dirname, '..', '..', '..');

/**
 * Positions that unambiguously hold a Problem code. Each pattern captures the
 * literal in group 1.
 *
 * Deliberately NOT "any dotted lowercase string": the frontend is full of dotted
 * strings that are capabilities (`audit.read`), audit event names
 * (`audit.export.requested`), and template tokens (`a.b`). A broad shape match
 * would drown the real signal in those, and a drowned signal gets an allow-list,
 * and an allow-list is how drift gets legitimized in the first place.
 */
const POSITIONS: Array<{ label: string; re: RegExp }> = [
  { label: '.code === …', re: /\.code\s*[=!]==?\s*['"]([A-Za-z][A-Za-z0-9_.]*)['"]/g },
  { label: 'new ApprovalError(…)', re: /new ApprovalError\(\s*['"]([A-Za-z][A-Za-z0-9_.]*)['"]/g },
  { label: 'ApiError.fromLegacy(…)', re: /ApiError\.fromLegacy\(\s*['"]([A-Za-z][A-Za-z0-9_.]*)['"]/g },
  { label: 'resolveErrorMessage(…)', re: /resolveErrorMessage\(\s*['"]([A-Za-z][A-Za-z0-9_.]*)['"]/g },
];

/**
 * A Problem code always contains a dot or is UPPER_SNAKE (the pre-ADR-0089 shape,
 * which must still be caught so a missed rename is reported rather than ignored).
 * Requiring one of those shapes is what lets the `.code === …` position skip
 * `typeof candidate.code === 'string'` type guards without a lookbehind hack.
 */
const CODE_SHAPED = /\.|^[A-Z][A-Z0-9_]*$/;

/**
 * The second rule, and the one that survives new call sites nobody added to
 * POSITIONS: ANY string literal whose prefix is one of the ten closed semantic
 * families is claiming to be a Problem code, wherever it appears. Since ADR 0089
 * the families are closed and every code is `family.rest`, so this is exact — and
 * it does not fire on the frontend's other dotted strings (capabilities like
 * `audit.read`, audit event names, template tokens), whose prefixes are not
 * families.
 *
 * This is why the bare `code: '…'` position is absent above: `code` is an
 * overloaded field name in this codebase (profile codes, area codes, role codes),
 * so matching it flagged 'JUR', 'POP', 'approver'. A check that cries wolf gets an
 * allow-list, and an allow-list is how drift becomes legitimate.
 */
const FAMILIES = [
  'request',
  'validation',
  'auth',
  'permission',
  'notfound',
  'state',
  'precondition',
  'conflict',
  'ratelimit',
  'internal',
];
const FAMILY_SHAPED = new RegExp(`['"]((?:${FAMILIES.join('|')})\\.[a-z0-9_.]+)['"]`, 'g');

/**
 * Codes that are legitimately not in the backend snapshot.
 *
 * `authn.expired` is synthesized by lib/api/client.ts when the session lapses —
 * it is dispatched, never received. `unmatched.code` is the single sentinel every
 * fixture uses when it deliberately exercises an unknown-code fallback path;
 * giving those one shared name (rather than a handful of ad-hoc invented strings)
 * is what keeps this list at two entries instead of growing one per test.
 */
const NOT_BACKEND_CODES = new Set(['authn.expired', 'unmatched.code']);

/**
 * Files that own a DIFFERENT dotted vocabulary which happens to collide with a
 * family prefix. Kept to whole files, listed individually, each with its reason —
 * never a directory glob, because a glob would silently swallow future error
 * handling added under the same path.
 */
const EXEMPT_FILES = new Set([
  // The PT-BR message map: its keys ARE the vocabulary, and parity is already
  // enforced in both directions by errorMessages.coverage.test.ts.
  'lib/api/errorMessages.ts',
  // This file quotes example codes in its own patterns and prose.
  'lib/api/__tests__/problemCodeVocabulary.test.ts',
  // Audit EVENT names (`auth.login`, `auth.session.revoked`, ...) — a separate
  // taxonomy that predates the Problem families and collides with `auth.` by
  // coincidence. They are event types, never Problem.code values.
  'features/iam/presenters/audit-event-presenter.ts',
]);

function walk(dir: string, out: string[] = []): string[] {
  for (const entry of readdirSync(dir)) {
    if (entry === 'node_modules') continue;
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) walk(full, out);
    else if (/\.tsx?$/.test(entry)) out.push(full);
  }
  return out;
}

describe('problem code vocabulary', () => {
  it('never references a code the backend does not register', () => {
    const known = new Set((generated as { codes: string[] }).codes);
    const offenders: string[] = [];

    for (const file of walk(SRC)) {
      const rel = relative(SRC, file).replace(/\\/g, '/');
      if (EXEMPT_FILES.has(rel)) continue;
      const text = readFileSync(file, 'utf8');
      const lines = text.split('\n');

      const checks = [...POSITIONS, { label: 'family-prefixed literal', re: FAMILY_SHAPED }];
      for (const { label, re } of checks) {
        for (const [i, line] of lines.entries()) {
          re.lastIndex = 0;
          let m: RegExpExecArray | null;
          while ((m = re.exec(line)) !== null) {
            const code = m[1];
            if (!CODE_SHAPED.test(code)) continue;
            if (known.has(code) || NOT_BACKEND_CODES.has(code)) continue;
            offenders.push(`${rel}:${i + 1}  ${label}  '${code}'`);
          }
        }
      }
    }

    expect(
      offenders,
      `Frontend references Problem code(s) the backend does not emit. A renamed code\n` +
        `leaves the branch dead rather than broken, so this is caught here and nowhere\n` +
        `else. Look the condition up in wiki/references/problem-codes.md and use the\n` +
        `current code — do not add an exception.\n\n  ${offenders.join('\n  ')}\n`,
    ).toEqual([]);
  });
});
