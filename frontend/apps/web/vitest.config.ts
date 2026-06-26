import { defineConfig } from 'vitest/config';
import { resolve } from 'node:path';

export default defineConfig({
  resolve: {
    alias: {},
  },
  test: {
    environment: 'jsdom',
    environmentOptions: {
      jsdom: {
        url: 'http://localhost/',
      },
    },
    setupFiles: ['./vitest.setup.ts'],
    testTimeout: 15000,
    globals: true,
    exclude: [
      '**/node_modules/**',
      '**/dist/**',
      // Playwright specs; vitest cannot run @playwright/test.
      '**/*.spec.ts',
    ],
  },
});
