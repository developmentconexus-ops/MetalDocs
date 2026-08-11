// ESLint flat config — Eigenpal Anti-Corruption Layer boundary guard (ADR 0046).
//
// This config was originally narrow — ONE invariant, the vendor `@eigenpal/*`
// may be imported only inside the two ACL walls (packages/eigenpal-adapter =
// server `.` door, packages/editor-ui = browser React wall) — plus the F1.4
// feature-boundary guard below. It is still not a full lint regime.
//
// The @typescript-eslint and react-hooks plugins were registered with every
// rule OFF solely so that pre-existing inline `eslint-disable` directives
// that name those rules resolve (otherwise ESLint errors "rule definition
// not found"). #91/A2.1 starts turning individual rules on, one at a time,
// each ratcheted via ESLint's own built-in suppressions mechanism (see
// `RATCHETED_RULES` below and `eslint-suppressions.json` /
// `eslint-suppressions.expiry.json` at the repo root) so the build goes red
// on *new* violations of a newly-enabled rule while pre-existing debt is
// baselined with an expiry date, not silently grandfathered forever.
import tseslint from 'typescript-eslint';
import reactHooks from 'eslint-plugin-react-hooks';

const restrictEigenpal = {
  'no-restricted-imports': [
    'error',
    {
      patterns: [
        {
          group: ['@eigenpal', '@eigenpal/*', '@eigenpal/**'],
          message:
            'ADR 0046: import @eigenpal only inside the ACL walls ' +
            '(packages/eigenpal-adapter, packages/editor-ui). Elsewhere use ' +
            '@metaldocs/eigenpal-adapter (server) or @metaldocs/editor-ui (browser).',
        },
      ],
    },
  ],
};

// ---------------------------------------------------------------------------
// F1.4 — FE feature-boundary guard (Milestone 1, GMR program).
//
// Invariant: a file under frontend/apps/web/src/features/<A>/** may not import
// from frontend/apps/web/src/features/<B>/** when A !== B. Shared/cross-cutting
// code (src/shared, src/lib, src/store, src/queries, src/components, etc.) is
// not a "feature" and stays unrestricted. Same-feature imports (A -> A) stay
// unrestricted.
//
// Zero-dep by design (see spec f1.4-fe-boundaries): built on the stock
// `no-restricted-imports` rule already in use above for the eigenpal ACL, no
// new eslint plugin. `no-restricted-imports` can't natively express "any
// sibling except my own feature" in one shot, so this generates one config
// block PER feature directory: each block is scoped (`files`) to that one
// feature's tree and forbids importing any OTHER feature directory, both via
// relative paths (`../../<other>/...`) and via the `@/features/<other>/...`
// alias form (the repo's vite/tsconfig currently define no `@` alias for
// `src`, and all existing feature imports are relative — the alias patterns
// are included defensively so a future alias doesn't silently reopen the gap).
//
// Grandfathering: the existing cross-feature edges (enumerated by hand at
// authoring time, see f1.4-fe-boundaries/evidence.md) are allowed via an
// explicit, finite, shrink-only ALLOWLIST below. Any cross-feature edge NOT
// in the allowlist is a new error. Entries are only ever removed (as edges
// get refactored to shared modules), never added.
const FEATURES_DIR = 'frontend/apps/web/src/features';

// Every feature directory under frontend/apps/web/src/features today.
// (Loose files directly in features/, e.g. featureFlags.ts, are not a
// feature dir and are not scoped by these blocks.)
const FEATURE_NAMES = [
  'approval',
  'auth',
  'controlled-documents',
  'dashboard',
  'documents',
  'feature-flags',
  'iam',
  'notifications',
  'password-change',
  'shell',
  'taxonomy',
  'templates',
  'tokens',
];

// Shrink-only allowlist of pre-existing cross-feature edges (owner: Leandro;
// trigger: incremental de-coupling — entries are removed as edges are
// refactored to shared modules, never added to). Enumerated 2026-07-03 via a
// full scan of `frontend/apps/web/src/features/**/*.{ts,tsx}` relative
// imports resolving into a different feature directory. 19 distinct
// (from -> to) pairs / 112 individual import statements.
const ALLOWLIST = [
  { from: 'approval', to: 'controlled-documents' },
  { from: 'approval', to: 'documents' },
  { from: 'approval', to: 'taxonomy' },
  { from: 'controlled-documents', to: 'documents' },
  { from: 'dashboard', to: 'approval' },
  { from: 'dashboard', to: 'documents' },
  { from: 'documents', to: 'approval' },
  { from: 'documents', to: 'controlled-documents' },
  { from: 'documents', to: 'iam' },
  { from: 'documents', to: 'taxonomy' },
  { from: 'documents', to: 'templates' },
  { from: 'shell', to: 'auth' },
  { from: 'shell', to: 'notifications' },
  { from: 'taxonomy', to: 'templates' },
  { from: 'templates', to: 'iam' },
  { from: 'templates', to: 'taxonomy' },
  { from: 'templates', to: 'tokens' },
  { from: 'tokens', to: 'iam' },
  { from: 'tokens', to: 'templates' },
];

function allowedTargets(featureName) {
  return ALLOWLIST.filter((e) => e.from === featureName).map((e) => e.to);
}

// glob `group` patterns match the raw import-specifier string, and relative
// imports (`../tokens/...`, `../../tokens/...`) vary in `../` depth per file,
// which a fixed-depth glob can't generalize over (verified: minimatch's `**`
// does not cross a leading `../` boundary). ESLint >=9.3's `no-restricted-imports`
// supports `patterns[].regex` (this repo is on ESLint 10.5.0), which is the
// zero-dep-correct primitive here: anchor on "N `../` or `./` segments then
// the other feature's name then a `/` or end-of-string" so touching e.g.
// "documents" never false-matches "controlled-documents" or vice versa.
function otherFeatureRegex(other) {
  // Escape regex metachars in the feature name (none of the current names
  // need it, but keep this correct if a future feature dir has one).
  const escaped = other.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  return `^((\\.\\./)+|\\./)features/${escaped}(/|$)|^((\\.\\./)+|\\./)${escaped}(/|$)|^@/features/${escaped}(/|$)`;
}

// One no-restricted-imports config block per feature, forbidding every OTHER
// feature except the ones this feature is explicitly allowlisted to import.
const featureBoundaryConfigs = FEATURE_NAMES.map((featureName) => {
  const forbiddenFeatures = FEATURE_NAMES.filter(
    (other) => other !== featureName && !allowedTargets(featureName).includes(other),
  );

  return {
    files: [`${FEATURES_DIR}/${featureName}/**/*.ts`, `${FEATURES_DIR}/${featureName}/**/*.tsx`],
    rules: {
      // Flat config: a later block's `rules['no-restricted-imports']` entry
      // REPLACES (does not merge with) an earlier one for the same matched
      // file. Re-include the eigenpal ACL patterns here so per-feature
      // blocks don't silently drop that guard for files under features/**.
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            ...restrictEigenpal['no-restricted-imports'][1].patterns,
            ...forbiddenFeatures.map((other) => ({
              regex: otherFeatureRegex(other),
              message:
                `F1.4: cross-feature import into "${featureName}" from sibling feature "${other}" ` +
                `is not allowed (frontend/apps/web hard-rule #8). Use src/shared, src/lib, ` +
                `src/store, or src/queries for cross-cutting code, or add an explicit, reviewed ` +
                `entry to the shrink-only ALLOWLIST in eslint.config.mjs if this edge is a ` +
                `deliberate, pre-existing exception.`,
            })),
          ],
        },
      ],
    },
  };
});

// ---------------------------------------------------------------------------
// A2.1 — frontend lint rule activation with ratchets (issue #91).
//
// Mechanism: ESLint 10's native "Suppressions" feature (`eslint --help` →
// "Suppressing Violations"), NOT a bespoke parallel baseline system. A
// suppressions file records a per-file, per-rule VIOLATION COUNT captured at
// baseline time; on every later run, a file/rule pair is allowed up to its
// recorded count, and only the delta above that count is reported as a real
// error. Fixing a baselined violation does not fail the build (the repo's
// `lint` script passes `--pass-on-unpruned-suppressions` for exactly this
// reason — see root package.json); adding a NEW one for an already-baselined
// file/rule pair does. A brand-new file/rule pair with zero baseline entries
// has zero tolerance from the moment the rule is enabled.
//
// A suppressions file has no expiry of its own, so `eslint-suppression-expiry`
// (tools/verify/registry.go, scripts/check-eslint-suppression-expiry.sh) is a
// second, repo-authored gate: every rule that appears anywhere in
// eslint-suppressions.json must have a live (non-expired) entry in
// eslint-suppressions.expiry.json, or the build goes red. That is the "not
// silently" half of the acceptance criterion — a baseline can persist past
// its date only via a reviewed edit that renews it.
//
// Rules enabled this slice (measured baseline at 32f88247, see the A2.1 PR
// description for the full per-rule count table):
//   - react-hooks/rules-of-hooks       0 findings — enabled clean, no baseline
//   - @typescript-eslint/no-unused-vars   38 findings — baselined, expiry recorded
//   - @typescript-eslint/no-explicit-any  13 findings — baselined, expiry recorded
//   - react-hooks/exhaustive-deps         10 findings — baselined, expiry recorded
// react-hooks/rules-of-hooks and react-hooks/exhaustive-deps are scoped to
// actual React application code (frontend/apps/web/src, packages/editor-ui/src)
// rather than every *.ts/*.tsx ESLint otherwise lints: Playwright's fixture
// callback signature (`async ({ page }, use, testInfo) => ...`) has a
// parameter literally named `use`, which rules-of-hooks' name heuristic
// mistakes for the React `use()` hook outside src/ (frontend/apps/web/e2e/).
// That is a real false positive, not debt to baseline — scoping the rule
// away from non-React code is the correct fix, not a suppression.
//
// Not enabled this slice, and why: the rest of `typescript-eslint`'s and
// `react-hooks`' recommended sets were measured (0 additional
// @typescript-eslint findings beyond the four above; react-hooks' newer
// React-Compiler-oriented diagnostics — set-state-in-effect 23,
// refs 5, preserve-manual-memoization 4, purity 3 findings — are a distinct,
// larger judgment call about React Compiler readiness this repo hasn't made
// yet). Enabling those is deliberately deferred to a future A2.1 follow-up,
// not silently dropped.
// A2.1 review round 2 (R2): two real, permanent limits of this ratchet, named
// here rather than left for the next reader to discover by surprise. R1
// (tools/verify check "eslint-suppression-baseline-growth") closes the
// direction the baseline can move — it can never grow relative to the merge
// base with origin/main — but it does not, and cannot, make the mechanism
// below identity-preserving or self-pruning. Both limits are inherent to
// ESLint 10's Suppressions feature as used here, not a gap this repo's own
// code introduced.
//
// 1. Suppression is COUNT-based per (file, rule), not FINDING-based. ESLint's
//    suppressions file records "file X has N violations of rule Y", not
//    which N. Proven live in this PR's review: removing the one baselined
//    `no-unused-vars` violation in
//    frontend/apps/web/src/lib/api/problem.ts while introducing a DIFFERENT
//    unused var in the same file leaves the count at 1, and `pnpm run lint`
//    passes clean — the baseline silently swapped which finding it pins. A
//    baselined file is a pinned COUNT, not a pinned set of findings; do not
//    assume the suppressions file records what was originally baselined,
//    only how much of it there was.
//
// 2. `--pass-on-unpruned-suppressions` (package.json's "lint" script)
//    disables ESLint's prune-forcing behaviour. Confirmed against ESLint
//    10.5.0 source (lib/cli.js:497-509): without the flag, fixing a
//    baselined violation exits 2 with "There are suppressions left that do
//    not occur anymore," forcing an immediate prune. With it, a stale count
//    (the violation is long gone but the suppressions entry still claims it)
//    sits indefinitely — the baseline can only shrink via a deliberate
//    `--prune-suppressions` run (see `pnpm run lint:prune`). This is a
//    DELIBERATE trade, not an oversight: without the flag, fixing one
//    baselined finding as a drive-by in an unrelated PR would fail that PR's
//    build over a count mismatch it did not cause. The compensating forcing
//    function is eslint-suppressions.expiry.json's expiry date (checked by
//    tools/verify's "eslint-suppression-expiry") — a rule's baseline cannot
//    sit stale past its expiry regardless of pruning.
//
// A2.1 review round 3 (Finding 2): a THIRD limit existed here, unlabelled,
// until this round — CLAUDE.md's "Global Maximum, Not Local Maximum" treats
// an unlabelled local maximum as its own defect, separate from the underlying
// bug. tools/verify's "eslint" check used to run `pnpm run lint` — the flags
// actually executed lived in package.json's "lint" script body, which
// nothing in the check pinned or content-checked. A one-line package.json
// diff appending "--suppress-all" made that script silently absorb any new
// violation into eslint-suppressions.json and exit 0, defeating every rule
// activated below while still reporting PASS (reproduced live in this
// round's cold review — see tools/verify/registry.go's "eslint" check and
// scripts/check-eslint-suppression-baseline-growth.sh's header for the full
// writeup). This is now closed, not merely documented: the "eslint" check
// runs a pinned Argv directly (no package.json script indirection to edit),
// and the growth check compares committed content at HEAD instead of the
// working tree, so an in-run mutation is structurally invisible to it rather
// than just harder to time. Residual gap, named rather than left implicit:
// package.json's devDependencies (plus pnpm-lock.yaml) still govern which
// eslint binary the pinned Argv resolves via `pnpm exec` — a supply-chain
// substitution there (e.g. a pnpm override redirecting the "eslint" package)
// is not something either check closes. That is the same pinned-third-party
// trust boundary already named in the "eslint" check's FixtureWaiver, not
// new debt introduced by this fix; closing it fully would mean verifying the
// resolved eslint binary's integrity (e.g. a checksum/provenance check on
// node_modules/.bin/eslint), which is out of scope for this slice.
const RATCHETED_TS_RULES = {
  '@typescript-eslint/no-unused-vars': 'error',
  '@typescript-eslint/no-explicit-any': 'error',
};

const RATCHETED_REACT_HOOKS_RULES = {
  'react-hooks/rules-of-hooks': 'error',
  'react-hooks/exhaustive-deps': 'error',
};

export default tseslint.config(
  {
    ignores: [
      '**/dist/**',
      '**/node_modules/**',
      '**/coverage/**',
      '**/build/**',
      '**/.vite/**',
      '**/.claude/**',
      // The two test-runner configs that name the vendor in `server.deps.inline`
      // regexes (not imports) for the test env. Scoped to the exact files instead
      // of a `**/*.config.*` wildcard, which would let any config bypass the guard.
      'apps/docx-renderer/vitest.config.ts',
      'packages/editor-ui/vitest.config.ts',
    ],
  },
  {
    files: ['**/*.ts', '**/*.tsx'],
    // Pre-existing inline disable directives reference these rules; register the
    // plugins so the names resolve. The rules themselves stay off.
    plugins: {
      '@typescript-eslint': tseslint.plugin,
      'react-hooks': reactHooks,
    },
    languageOptions: { parser: tseslint.parser },
    linterOptions: { reportUnusedDisableDirectives: 'off' },
    rules: restrictEigenpal,
  },
  {
    // The two sanctioned ACL walls import the vendor internally by design.
    files: ['packages/eigenpal-adapter/**', 'packages/editor-ui/**'],
    rules: { 'no-restricted-imports': 'off' },
  },
  ...featureBoundaryConfigs,
  {
    // A2.1: general TS hygiene, repo-wide (same scope the base config already
    // lints) — no false-positive risk from non-React code, unlike the
    // react-hooks rules below.
    files: ['**/*.ts', '**/*.tsx'],
    rules: RATCHETED_TS_RULES,
  },
  {
    // A2.1: react-hooks rules scoped to actual React application source —
    // see the comment above RATCHETED_REACT_HOOKS_RULES for why this is
    // narrower than the base config's **/*.ts,**/*.tsx.
    files: [
      'frontend/apps/web/src/**/*.ts',
      'frontend/apps/web/src/**/*.tsx',
      'packages/editor-ui/src/**/*.ts',
      'packages/editor-ui/src/**/*.tsx',
    ],
    rules: RATCHETED_REACT_HOOKS_RULES,
  },
);
