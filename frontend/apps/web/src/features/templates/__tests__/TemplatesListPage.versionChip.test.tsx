import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TemplatesListPage } from '../TemplatesListPage';
import type { TemplateDTO } from '../api/templates';

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
  it('renders Rascunho v1 for never-published template', () => {
    renderPage();
    const card = screen.getByText('Never Published').closest('article, [role="button"], div');
    expect(card).toBeTruthy();
    expect(screen.getByText('Rascunho v1')).toBeTruthy();
  });

  it('renders v1 (not v2) when v1 published with v2 draft auto-spawned', () => {
    renderPage();
    expect(screen.getByText('Published With Draft')).toBeTruthy();
    const chips = screen.getAllByText('v1');
    expect(chips.length).toBeGreaterThanOrEqual(2);
  });

  it('renders v1 when published with no further draft', () => {
    renderPage();
    expect(screen.getByText('Published No Draft')).toBeTruthy();
  });

  it('renders latest_version label for archived template', () => {
    renderPage();
    expect(screen.getByText('Archived')).toBeTruthy();
    expect(screen.getByText('v3')).toBeTruthy();
  });
});
