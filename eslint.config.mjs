// ESLint flat config — Eigenpal Anti-Corruption Layer boundary guard (ADR 0046).
//
// This config is intentionally narrow: it enforces ONE invariant — the vendor
// `@eigenpal/*` may be imported only inside the two ACL walls
// (packages/eigenpal-adapter = server `.` door, packages/editor-ui = browser
// React wall). It is NOT a full lint regime.
//
// The @typescript-eslint and react-hooks plugins are registered with their
// rules OFF solely so that pre-existing inline `eslint-disable` directives that
// name those rules resolve (otherwise ESLint errors "rule definition not
// found"). Turning those rules on is a separate, deliberate future decision.
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
);
