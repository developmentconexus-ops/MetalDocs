/**
 * Regression guard for SLICE 8b (unit 3.2): the submit-time approval-route
 * preview must gate the "Submeter para aprovação" button. Race guarded by
 * TemplateEditorPage.tsx's `previewNotReady = isDraft && !approvalPreviewQuery.isSuccess`
 * (see the button's `disabled` prop and label around TemplateEditorPage.tsx:364-371).
 *
 * Harness mirrors template-author-page-convergence.test.tsx (same hook mocks
 * for useTemplateDraft/useTemplateAutosave/useTemplateSchemas/editor-ui), plus:
 *  - a mock of useTemplateVersionApprovalPreviewQuery so isSuccess/data are
 *    controlled directly instead of depending on a real network round-trip.
 *  - a mock of useAuthStore granting the `template.submit` capability, so
 *    canSubmit() does not itself disable the button and the preview gate is
 *    isolated as the only variable under test.
 */
import React from 'react';
import { render, screen } from '@testing-library/react';
import { QueryClientProvider, QueryClient } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';

// Stable references: TemplateEditorPage has effects keyed on `draft.version`
// and `schemaState.schemas`, so the hook mocks MUST return the SAME object
// identity across renders. A fresh literal per render retriggers
// setLiveVersion / setLocalSchemas every render and spins an infinite loop
// (OOM). Mirrors template-author-page-convergence.test.tsx, the sibling test
// that mounts this page successfully.
const draftTemplate = { template_id: 'tpl-1', name: 'Test Template' };
const draftVersion = {
  template_id: 'tpl-1',
  version_number: 1,
  revision_number: 0,
  status: 'draft',
  docx_storage_key: null,
};
const baseSchemas = {
  placeholders: [],
  composition: { headerSubBlocks: [], footerSubBlocks: [], subBlockParams: {} },
};
const refetch = vi.fn();
const queueDocx = vi.fn();
const flush = vi.fn().mockResolvedValue(true);
const saveSchemas = vi.fn();

// Mutable preview-query state so each test controls isSuccess/data directly —
// mirrors the mockCapabilities pattern in DocumentWorkspacePage.test.tsx.
let mockPreviewState: { data: unknown; isSuccess: boolean; isLoading: boolean } = {
  data: { route_id: null, route_name: null, stages: [] },
  isSuccess: true,
  isLoading: false,
};

vi.mock('prosemirror-view/style/prosemirror.css', () => ({}));
vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: React.forwardRef((_props: Record<string, unknown>, ref) => {
    React.useImperativeHandle(ref, () => ({
      getDetectedTokens: () => [],
      insertToken: () => {},
      getUsedTokens: () => [],
      getPageCount: () => null,
      focus: () => {},
    }));
    return <div data-testid="editor" />;
  }),
}));

vi.mock('../hooks/useTemplateDraft', () => ({
  useTemplateDraft: () => ({
    loading: false,
    error: null,
    template: draftTemplate,
    version: draftVersion,
    docxBytes: null,
    refetch,
  }),
}));
vi.mock('../hooks/useTemplateAutosave', () => ({
  useTemplateAutosave: () => ({
    queueDocx,
    flush,
    status: 'idle',
    hasPending: () => false,
  }),
}));
vi.mock('../hooks/useTemplateSchemas', () => ({
  useTemplateSchemas: () => ({
    schemas: baseSchemas,
    loading: false,
    error: null,
    save: saveSchemas,
    saving: false,
  }),
}));
vi.mock('../queries/useTemplateVersionApprovalPreviewQuery', () => ({
  useTemplateVersionApprovalPreviewQuery: () => mockPreviewState,
}));
// Stable state object created ONCE — real zustand returns a constant store
// reference, so `useAuthStore((s) => s.user)` reads a stable `s.user` across
// renders. Rebuilding `{ user: {...} }` per call would hand back a fresh
// reference every render and spin an infinite re-render loop (OOM).
vi.mock('../../../store/auth.store', () => {
  const state = { user: { capabilities: ['template.submit'] } };
  return {
    useAuthStore: (selector: (s: typeof state) => unknown) => selector(state),
  };
});

import { TemplateEditorPage } from '../pages/TemplateEditorPage';

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <TemplateEditorPage templateId="tpl-1" versionNum={1} />
    </QueryClientProvider>,
  );
}

describe('TemplateEditorPage — submit gated on approval-route preview (SLICE 8b)', () => {
  beforeEach(() => {
    mockPreviewState = { data: { route_id: null, route_name: null, stages: [] }, isSuccess: true, isLoading: false };
  });

  it('blocks submit while the preview has not resolved (race case)', async () => {
    mockPreviewState = { data: undefined, isSuccess: false, isLoading: true };
    renderPage();

    const submitBtn = await screen.findByRole('button', { name: 'Carregando rota de aprovação…' });
    expect(submitBtn).toBeDisabled();
  });

  it('does not block submit once the preview has resolved (control case)', async () => {
    mockPreviewState = { data: { route_id: 'route-1', route_name: 'Rota padrão', stages: [] }, isSuccess: true, isLoading: false };
    renderPage();

    const submitBtn = await screen.findByRole('button', { name: 'Submeter para aprovação' });
    expect(submitBtn).not.toBeDisabled();
  });
});
