import { describe, expect, it } from 'vitest';
import { documentsRoutes } from './routes';

describe('documentsRoutes', () => {
  it('declares the editor route before generic document detail routes', () => {
    const editorRouteIndex = documentsRoutes.findIndex((route) => route.path === 'documents/:documentId/edit');
    const detailRouteIndex = documentsRoutes.findIndex((route) => route.path === 'documents/:documentId');

    expect(editorRouteIndex).toBeGreaterThanOrEqual(0);
    expect(detailRouteIndex).toBeGreaterThanOrEqual(0);
    expect(editorRouteIndex).toBeLessThan(detailRouteIndex);
  });
});
