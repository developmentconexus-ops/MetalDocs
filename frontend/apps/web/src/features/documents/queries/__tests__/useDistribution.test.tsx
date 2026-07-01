/**
 * T2 — RED test for F2.3 distribution wiring.
 * Tests the page + hooks against fixtured responses typed to generated schemas.
 * Run BEFORE T3–T6 to capture the RED baseline, then GREEN after.
 */
import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

// ── Page under test ──────────────────────────────────────────────────────────
import { DocumentDistributionPage } from '../../pages/DocumentDistributionPage';

// ── Route mock ───────────────────────────────────────────────────────────────
vi.mock('react-router-dom', async (importOriginal) => {
  const actual = await importOriginal<typeof import('react-router-dom')>();
  return {
    ...actual,
    useParams: () => ({ documentId: 'doc-dist-1' }),
  };
});

// ── doc-detail mock ──────────────────────────────────────────────────────────
vi.mock('../../queries/useDocumentDetailQuery', () => ({
  useDocumentDetailQuery: vi.fn(),
}));
import { useDocumentDetailQuery } from '../../queries/useDocumentDetailQuery';

// ── distribution API mocks ───────────────────────────────────────────────────
vi.mock('../../api/distribution', () => ({
  getDistributionSummary: vi.fn(),
  listDistributionRecipients: vi.fn(),
  getDistributionCoverage: vi.fn(),
}));
import {
  getDistributionSummary,
  listDistributionRecipients,
  getDistributionCoverage,
} from '../../api/distribution';
import type {
  DistributionSummaryResponse,
  DistributionRecipientsResponse,
  DistributionAreaCoverage,
} from '../../api/distribution';

// ── Fixtures typed to generated schemas ─────────────────────────────────────
const baseDoc = {
  id: 'doc-dist-1',
  tenant_id: 'tenant-1',
  template_version_id: 'template-1',
  name: 'Procedimento Distribuído',
  status: 'published',
  form_data_json: {},
  current_revision_id: 'rev-1',
  revision_version: 2,
  active_session_id: '',
  values_frozen_at: null,
  archived_at: null,
  created_at: '2026-05-19T00:00:00.000Z',
  updated_at: '2026-05-19T00:00:00.000Z',
  created_by: 'admin-user',
  revision_number: 3,
  controlled_document_id: 'cd-1',
  revision_title: 'Procedimento Distribuído',
  profile_code_snapshot: 'pop',
  process_area_code_snapshot: 'general',
  code: 'POP-GENERAL-014',
  current_revision_file_size_bytes: 1024,
  current_revision_page_count: 1,
  current_revision_page_count_source: 'server_renderer',
};

const summaryFixture: DistributionSummaryResponse = { total_targets: 42 };

const page1Fixture: DistributionRecipientsResponse = {
  items: [
    { user_id: 'u1', name: 'Ana Paula Mendes', area_code: 'rh', area_name: 'Recursos Humanos', source: 'area_grant' },
    { user_id: 'u2', name: 'Roberto Carvalho', area_code: null, area_name: null, source: 'user_grant' },
    { user_id: 'u3', name: 'Helena Tavares', area_code: 'ssma', area_name: 'SSMA', source: 'company_scope' },
  ],
  page: { next_cursor: 'cursor-page2', has_more: true },
};

const page2Fixture: DistributionRecipientsResponse = {
  items: [
    { user_id: 'u4', name: 'João Vitor Lima', area_code: 'prod', area_name: 'Produção', source: 'area_grant' },
  ],
  page: { next_cursor: null, has_more: false },
};

const coverageFixture: DistributionAreaCoverage[] = [
  { area_code: 'rh', area_name: 'Recursos Humanos', total: 20 },
  { area_code: 'ssma', area_name: 'SSMA', total: 22 },
];

const emptyRecipientsFixture: DistributionRecipientsResponse = {
  items: [],
  page: { next_cursor: null, has_more: false },
};

// ── Helper ───────────────────────────────────────────────────────────────────
function wrap() {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  // MemoryRouter needed because ArtifactHero now uses react-router Link
  const { MemoryRouter } = require('react-router-dom') as typeof import('react-router-dom');
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(
      MemoryRouter,
      null,
      React.createElement(QueryClientProvider, { client: qc }, children),
    );
}

describe('DocumentDistributionPage — F2.3 distribution wiring', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default: doc-detail resolves immediately
    vi.mocked(useDocumentDetailQuery).mockReturnValue({
      isLoading: false,
      isError: false,
      data: baseDoc,
      refetch: vi.fn(),
    } as never);

    // Default: all distribution fetchers resolve to happy-path fixtures
    vi.mocked(getDistributionSummary).mockResolvedValue(summaryFixture);
    vi.mocked(listDistributionRecipients).mockResolvedValue(page1Fixture);
    vi.mocked(getDistributionCoverage).mockResolvedValue(coverageFixture);
  });

  // ── V1: illustrative literals gone ─────────────────────────────────────────
  it('renders NO illustrative watermark or Em-breve banner', async () => {
    render(<DocumentDistributionPage />, { wrapper: wrap() });
    await waitFor(() => expect(screen.queryByText(/42/)).toBeInTheDocument());

    expect(screen.queryByText(/Dados ilustrativos · Em breve/)).toBeNull();
    // The "em-breve" banner with "ilustrativos" text must be gone
    expect(screen.queryByText(/todos os números.*são.*ilustrativos/i)).toBeNull();
    // No "Em breve" badge in the hero badges area
    const badges = document.querySelectorAll('[class*="soonBadge"]');
    expect(badges.length).toBe(0);
  });

  // ── V2a: live total_targets renders ────────────────────────────────────────
  it('renders live total_targets from summary endpoint', async () => {
    render(<DocumentDistributionPage />, { wrapper: wrap() });
    await waitFor(() => expect(screen.getByText('42')).toBeInTheDocument());
  });

  // ── V2b: recipient list renders name + origin chip ─────────────────────────
  it('renders recipient names and origin chips from live data', async () => {
    render(<DocumentDistributionPage />, { wrapper: wrap() });

    await waitFor(() => expect(screen.getByText('Ana Paula Mendes')).toBeInTheDocument());
    expect(screen.getByText('Roberto Carvalho')).toBeInTheDocument();
    expect(screen.getByText('Helena Tavares')).toBeInTheDocument();

    // Origin chips: area_grant → area_name; user_grant → "Concessão direta"; company_scope → "Toda a empresa"
    // "Recursos Humanos" appears in both the RecipientsCard origin chip AND CoverageByArea row — use getAllByText
    expect(screen.getAllByText('Recursos Humanos').length).toBeGreaterThan(0);
    expect(screen.getByText('Concessão direta')).toBeInTheDocument();
    expect(screen.getByText('Toda a empresa')).toBeInTheDocument();
  });

  // ── V2c: by-area coverage rows render ─────────────────────────────────────
  it('renders coverage by area rows with area_name and total', async () => {
    render(<DocumentDistributionPage />, { wrapper: wrap() });

    // "Recursos Humanos" appears in both CoverageByArea and RecipientsCard origin chip
    await waitFor(() => expect(screen.getAllByText('Recursos Humanos').length).toBeGreaterThan(0));
    // SSMA appears in CoverageByArea rows
    expect(screen.getAllByText('SSMA').length).toBeGreaterThan(0);
    // totals appear somewhere (20 and 22)
    expect(screen.getByText('20')).toBeInTheDocument();
    expect(screen.getByText('22')).toBeInTheDocument();
  });

  // ── V3: numerator is tracking-pending — never a fabricated number ──────────
  it('renders the tracking-pending note for numerator surfaces', async () => {
    render(<DocumentDistributionPage />, { wrapper: wrap() });

    await waitFor(() => expect(screen.getByText('42')).toBeInTheDocument());

    // The honest disclosure must appear (at least one instance for numerator surfaces)
    const pending = screen.getAllByText(/acompanhamento.*ainda não disponível/i);
    expect(pending.length).toBeGreaterThan(0);

    // No fabricated zeros: no "0%" or "0 de N" should appear
    expect(screen.queryByText(/0%/)).toBeNull();
    expect(screen.queryByText(/0 de \d/)).toBeNull();
  });

  // ── V4: CTAs remain aria-disabled with deferred-with-trigger note ──────────
  it('hero CTAs remain aria-disabled pointing at the parked mission', async () => {
    render(<DocumentDistributionPage />, { wrapper: wrap() });

    const ctaLabels = [
      'Lembrete em massa',
      'Exportar relatório',
      'Adicionar destinatários',
      'Política de fanout',
    ];
    for (const label of ctaLabels) {
      const btn = screen.getByRole('button', { name: new RegExp(label) });
      expect(btn.getAttribute('aria-disabled')).toBe('true');
    }
  });

  // ── V5: "Carregar mais" appends page 2 via next_cursor ────────────────────
  it('"Carregar mais" fetches next page and appends recipients', async () => {
    // First call → page1, second call (with cursor) → page2
    vi.mocked(listDistributionRecipients)
      .mockResolvedValueOnce(page1Fixture)
      .mockResolvedValueOnce(page2Fixture);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    await waitFor(() => expect(screen.getByText('Ana Paula Mendes')).toBeInTheDocument());

    const loadMoreBtn = screen.getByRole('button', { name: /carregar mais/i });
    expect(loadMoreBtn).toBeInTheDocument();

    fireEvent.click(loadMoreBtn);

    await waitFor(() => expect(screen.getByText('João Vitor Lima')).toBeInTheDocument());
    // Page 1 recipients still present (append, not replace)
    expect(screen.getByText('Ana Paula Mendes')).toBeInTheDocument();

    // After page 2 (has_more: false) the button should be gone
    await waitFor(() =>
      expect(screen.queryByRole('button', { name: /carregar mais/i })).toBeNull(),
    );
  });

  // ── V5b: no "Carregar mais" when list exhausted ────────────────────────────
  it('does not render "Carregar mais" when has_more is false from the start', async () => {
    vi.mocked(listDistributionRecipients).mockResolvedValue(page2Fixture);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    await waitFor(() => expect(screen.getByText('João Vitor Lima')).toBeInTheDocument());
    expect(screen.queryByRole('button', { name: /carregar mais/i })).toBeNull();
  });

  // ── V8a: honest empty state ────────────────────────────────────────────────
  it('renders honest empty state when total_targets=0 and items=[]', async () => {
    vi.mocked(getDistributionSummary).mockResolvedValue({ total_targets: 0 });
    vi.mocked(listDistributionRecipients).mockResolvedValue(emptyRecipientsFixture);
    vi.mocked(getDistributionCoverage).mockResolvedValue([]);

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    // total=0 should render "0" (real zero is meaningful)
    await waitFor(() => expect(screen.getByText('0')).toBeInTheDocument());
    // empty recipients message
    expect(screen.getByText(/nenhum destinatário/i)).toBeInTheDocument();
  });

  // ── V8b: per-section error — page stays up, each section shows its own alert ─
  it('renders per-section error alerts (not a full-page crash) when all distribution fetches fail', async () => {
    vi.mocked(getDistributionSummary).mockRejectedValue(new Error('Network error'));
    vi.mocked(listDistributionRecipients).mockRejectedValue(new Error('Network error'));
    vi.mocked(getDistributionCoverage).mockRejectedValue(new Error('Network error'));

    render(<DocumentDistributionPage />, { wrapper: wrap() });

    // Page shell (heading) stays rendered — no full-page takeover
    expect(screen.getByRole('heading', { name: /Distribuição & cobertura de leitura/ })).toBeInTheDocument();

    // Per-section error messages appear
    await waitFor(() =>
      expect(screen.getByText(/não foi possível carregar os destinatários/i)).toBeInTheDocument(),
    );
    expect(screen.getByText(/não foi possível carregar a cobertura por área/i)).toBeInTheDocument();

    // Summary error: total_targets shows "—" (em-dash), not 0 or a number
    const emDashes = screen.getAllByText('—');
    expect(emDashes.length).toBeGreaterThan(0);
  });
});
