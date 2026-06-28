import React from 'react';
import { act, fireEvent, render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { DetectedToken } from '@metaldocs/shared-tokens';
import type { TemplateSchemas } from '../api/templates';

// Regression guard for SP-0: the editor's body `onChange` must NEVER write back
// into the template schema. The old "schema-corruption path"
// (getAgent().getVariables() -> localSchemas.placeholders -> schemaState.save)
// was removed because a body edit could overwrite the author's placeholder
// schema. This test reintroduces a body change and asserts the schema save is
// not invoked. If a future refactor reattaches a schema write to the editor
// change handler, this test goes red.
//
// Raised as Major-3 in the SP-0 Grade-A review; closes
// `wiki/backlog/template-editor.md#convergence-test-rewrite`.

const saveSchemas = vi.fn();
const queueDocx = vi.fn();
const flush = vi.fn();

// What the editor reports as detected tokens on each change. The page feeds
// these into the read-only "tokens in use" panel state — NOT into the schema.
let detectedTokens: DetectedToken[] = [];

const baseSchemas: TemplateSchemas = {
  placeholders: [],
  composition: { headerSubBlocks: [], footerSubBlocks: [], subBlockParams: {} },
};

// Stable references: the page has effects keyed on `draft.version` /
// `schemaState.schemas`, so the hook mocks MUST return the same object identity
// across renders. Returning fresh literals would retrigger setLiveVersion every
// render and spin an infinite loop.
const draftTemplate = { template_id: 'tpl-1', name: 'Test Template' };
const draftVersion = {
  template_id: 'tpl-1',
  version_number: 1,
  revision_number: 0,
  status: 'draft',
  docx_storage_key: null,
};
const refetch = vi.fn();

vi.mock('prosemirror-view/style/prosemirror.css', () => ({}));
vi.mock('@metaldocs/editor-ui', () => ({
  MetalDocsEditor: React.forwardRef(({ onChange }: { onChange?: () => void }, ref) => {
    React.useImperativeHandle(ref, () => ({
      getDetectedTokens: () => detectedTokens,
      insertToken: () => {},
      getUsedTokens: () => [],
      getPageCount: () => null,
      focus: () => {},
    }));
    return (
      <button data-testid="mock-editor-change" onClick={() => onChange?.()}>
        change
      </button>
    );
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
  useTemplateAutosave: () => ({ queueDocx, flush, status: 'idle', hasPending: () => false }),
}));
vi.mock('../hooks/useTemplateSchemas', () => ({
  useTemplateSchemas: () => ({ schemas: baseSchemas, loading: false, error: null, save: saveSchemas, saving: false }),
}));
vi.mock('../api/catalog', () => ({
  fetchPlaceholderCatalog: () =>
    Promise.resolve([
      { key: 'doc_code', label: 'Código do documento', description: '' },
      { key: 'author', label: 'Autor', description: '' },
    ]),
}));

import { TemplateEditorPage } from '../pages/TemplateEditorPage';

function renderPage() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <TemplateEditorPage templateId="tpl-1" versionNum={1} />
    </QueryClientProvider>,
  );
}

// The editor change handler debounces token classification by 400ms.
const DEBOUNCE_MS = 400;

describe('TemplateEditorPage — body edit never writes the schema', () => {
  beforeEach(() => {
    saveSchemas.mockReset();
    queueDocx.mockReset();
    flush.mockReset();
    detectedTokens = [];
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('does not call schemaState.save when the editor body changes', async () => {
    renderPage();
    // Catalog (and therefore the authoring panel) is loaded.
    await screen.findByTestId('token-doc_code');

    // Author edits the body; the editor reports detected tokens.
    detectedTokens = [
      { name: 'doc_code', kind: 'var', valid: true },
      { name: 'author', kind: 'var', valid: true },
    ] as DetectedToken[];

    fireEvent.click(screen.getByTestId('mock-editor-change'));

    // Let the debounce fire and any resulting effects settle.
    await act(async () => {
      await new Promise((r) => setTimeout(r, DEBOUNCE_MS + 50));
    });

    // The detected tokens flow into read-only panel state only...
    expect(screen.getByTestId('token-doc_code').getAttribute('data-used')).toBe('true');
    // ...and the schema save is never invoked by a body edit.
    expect(saveSchemas).not.toHaveBeenCalled();
  });
});
