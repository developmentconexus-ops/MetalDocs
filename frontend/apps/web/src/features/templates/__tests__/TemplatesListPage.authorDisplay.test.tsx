import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen, within } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { TemplatesListPage } from '../TemplatesListPage';
import type { TemplateDTO } from '../api/templates';

// FE-08 (P3, wiki/backlog/templates.md rows 8/9): TemplateDTO now carries
// created_by_display_name (resolved server-side via the iam
// UserDisplayNameReader port) and updated_at. TemplatesListPage must prefer
// both over the raw created_by id / created_at fallback, while still
// tolerating rows where the API omits them (older API version / no
// resolvable display name).

let fixtures: TemplateDTO[] = [];

vi.mock('../queries/useTemplatesQuery', () => ({
  useTemplatesQuery: () => ({
    data: { templates: fixtures, meta: { limit: 50, offset: 0 } },
    isLoading: false,
    isError: false,
    error: null,
  }),
}));

function baseFixture(overrides: Partial<TemplateDTO>): TemplateDTO {
  return {
    id: '11111111-1111-1111-1111-111111111111',
    tenant_id: 't1',
    key: 'tpl-1',
    name: 'Template One',
    doc_type_code: null,
    description: null,
    latest_version: { id: 'v1', number: 1, revision_number: 0, status: 'draft' },
    published_version: null,
    created_by: 'u1',
    created_at: new Date('2026-01-01T00:00:00Z').toISOString(),
    archived_at: null,
    ...overrides,
  } as TemplateDTO;
}

function renderPage() {
  const qc = new QueryClient();
  return render(
    <QueryClientProvider client={qc}>
      <TemplatesListPage onOpenTemplate={() => {}} onCreate={() => {}} />
    </QueryClientProvider>,
  );
}

describe('TemplatesListPage author + updated display (FE-08)', () => {
  it('renders created_by_display_name instead of the raw created_by uuid when present', () => {
    fixtures = [
      baseFixture({
        id: '11111111-1111-1111-1111-111111111111',
        name: 'Has Display Name',
        created_by: 'a1b2c3d4-0000-0000-0000-000000000000',
        created_by_display_name: 'Maria Silva',
      }),
    ];
    renderPage();
    const card = screen.getByLabelText('Abrir template Has Display Name');
    expect(within(card).getByText('Maria')).toBeTruthy();
    expect(within(card).queryByText('a1b2c3d4-0000-0000-0000-000000000000')).toBeNull();
  });

  it('falls back to the raw created_by uuid when created_by_display_name is absent', () => {
    fixtures = [
      baseFixture({
        id: '22222222-2222-2222-2222-222222222222',
        name: 'No Display Name',
        created_by: 'raw-user-id',
      }),
    ];
    renderPage();
    const card = screen.getByLabelText('Abrir template No Display Name');
    expect(within(card).getByText('raw-user-id')).toBeTruthy();
  });

  it('prefers updated_at over created_at for the relative timestamp when both are present', () => {
    fixtures = [
      baseFixture({
        id: '33333333-3333-3333-3333-333333333333',
        name: 'Recently Updated',
        created_at: new Date(Date.now() - 400 * 24 * 60 * 60 * 1000).toISOString(), // ~1 year ago
        updated_at: new Date().toISOString(), // today
      }),
    ];
    renderPage();
    const card = screen.getByLabelText('Abrir template Recently Updated');
    expect(within(card).getByText('hoje')).toBeTruthy();
  });

  it('falls back to created_at for the relative timestamp when updated_at is absent', () => {
    fixtures = [
      baseFixture({
        id: '44444444-4444-4444-4444-444444444444',
        name: 'Never Updated',
        created_at: new Date().toISOString(), // today
      }),
    ];
    renderPage();
    const card = screen.getByLabelText('Abrir template Never Updated');
    expect(within(card).getByText('hoje')).toBeTruthy();
  });
});
