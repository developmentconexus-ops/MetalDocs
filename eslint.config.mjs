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
);
