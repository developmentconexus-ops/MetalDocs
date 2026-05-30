import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TemplatesListPage } from '../TemplatesListPage';
import type { TemplateDTO } from '../api/templates';

// ADR 0013 — Template Revision Labels.
// Fixtures cover the 4 chip states:
//   1. Never published                                       → no REV chip
//   2. Published v1 + auto-spawned v2 draft (current_revision_number=0) → REV00
//   3. Published v1, no further draft (current_revision_number=0)       → REV00
//   4. Archived with current_revision_number=1                          → REV01

vi.mock('../queries/useTemplatesQuery', () => ({
  useTemplatesQuery: () => ({
    data: { templates: fixtures, meta: { limit: 50, offset: 0 } },
    isLoading: false,
    isError: false,
    error: null,
  }),
}));

const fixtures: TemplateDTO[] = [
  {
    id: '11111111-1111-1111-1111-111111111111',
    tenant_id: 't1',
    key: 'tpl-never-published',
    name: 'Never Published',
    doc_type_code: null,
    description: null,
    latest_version: 1,
    published_version_id: null,
    published_version_number: null,
    current_revision_number: null,
    created_by: 'u1',
    created_at: new Date().toISOString(),
    archived_at: null,
  } as TemplateDTO,
  {
    id: '22222222-2222-2222-2222-222222222222',
    tenant_id: 't1',
    key: 'tpl-published-with-draft',
    name: 'Published With Draft',
    doc_type_code: null,
    description: null,
    latest_version: 2,
    published_version_id: 'ver-1',
    published_version_number: 1,
    current_revision_number: 0,
    created_by: 'u1',
    created_at: new Date().toISOString(),
    archived_at: null,
  } as TemplateDTO,
  {
    id: '33333333-3333-3333-3333-333333333333',
    tenant_id: 't1',
    key: 'tpl-published-no-draft',
    name: 'Published No Draft',
    doc_type_code: null,
    description: null,
    latest_version: 1,
    published_version_id: 'ver-x',
    published_version_number: 1,
    current_revision_number: 0,
    created_by: 'u1',
    created_at: new Date().toISOString(),
    archived_at: null,
  } as TemplateDTO,
  {
    id: '44444444-4444-4444-4444-444444444444',
    tenant_id: 't1',
    key: 'tpl-archived',
    name: 'Archived',
    doc_type_code: null,
    description: null,
    latest_version: 3,
    published_version_id: 'ver-y',
    published_version_number: 2,
    current_revision_number: 1,
    created_by: 'u1',
    created_at: new Date().toISOString(),
    archived_at: new Date().toISOString(),
  } as TemplateDTO,
];

function renderPage() {
  const qc = new QueryClient();
  return render(
    <QueryClientProvider client={qc}>
      <TemplatesListPage onOpenTemplate={() => {}} onCreate={() => {}} />
    </QueryClientProvider>,
  );
}

describe('TemplatesListPage version chip', () => {
  it('renders no REV chip for never-published template (draft honesty rule)', () => {
    renderPage();
    const card = screen.getByLabelText('Abrir template Never Published');
    expect(card).toBeTruthy();
    // No REV?? chip rendered for templates without a published revision.
    expect(within(card).queryByText(/^REV\d{2}$/)).toBeNull();
  });

  it('renders REV00 for v1-published-with-v2-draft (chip ignores draft v2)', () => {
    renderPage();
    expect(screen.getByText('Published With Draft')).toBeTruthy();
    // REV00 appears for both v1-published rows below.
    expect(screen.getAllByText('REV00').length).toBeGreaterThanOrEqual(2);
  });

  it('renders REV00 when v1 published with no further draft', () => {
    renderPage();
    expect(screen.getByText('Published No Draft')).toBeTruthy();
  });

  it('renders REV01 for archived template whose published version was v2', () => {
    renderPage();
    expect(screen.getByText('Archived')).toBeTruthy();
    expect(screen.getByText('REV01')).toBeTruthy();
  });
});
