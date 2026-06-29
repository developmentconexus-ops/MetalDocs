import { describe, expect, it } from 'vitest';
import { tokenRoutes } from './routes';

describe('tokenRoutes', () => {
  it('gates templates/tokens on token.view', () => {
    const route = tokenRoutes.find((r) => r.path === 'templates/tokens');
    expect(route).toBeDefined();
    expect((route?.handle as { requiresCapability?: string }).requiresCapability).toBe('token.view');
    expect((route?.handle as { workspaceView?: string }).workspaceView).toBe('templates');
  });
});
