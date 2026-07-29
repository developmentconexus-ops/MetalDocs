import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { Component as DashboardPage } from './DashboardPage';
import { useAuthStore } from '../../../store/auth.store';
import type { InboxItem } from '../../approval/api/approvalTypes';

vi.mock('../queries/useDashboardInboxQuery', () => ({
  useDashboardInboxQuery: vi.fn(),
}));

vi.mock('../queries/useDashboardActivityQuery', () => ({
  useDashboardActivityQuery: vi.fn(),
}));

vi.mock('../../documents/queries/useLibraryStatsQuery', () => ({
  useLibraryStatsQuery: vi.fn(),
}));

import { useDashboardInboxQuery } from '../queries/useDashboardInboxQuery';
import { useDashboardActivityQuery } from '../queries/useDashboardActivityQuery';
import { useLibraryStatsQuery } from '../../documents/queries/useLibraryStatsQuery';

// A real controlled-document uuid: the truthful fallback the row may render only
// when the human code is absent — never alongside/instead of it.
const CONTROLLED_DOCUMENT_UUID = '3f6a1c2e-9b41-4d3a-8f27-2c5e7a1b9d04';

function makeItem(overrides: Partial<InboxItem> = {}): InboxItem {
  return {
    instance_id: 'inst-1',
    subject_kind: 'document',
    subject_key: 'doc-1',
    subject_title: 'POP Limpeza de Linha',
    subject_ref: 'doc-1',
    controlled_document_id: CONTROLLED_DOCUMENT_UUID,
    controlled_document_code: 'POP-QUA-0555',
    area_code: 'QUA',
    submitted_by: 'maria',
    submitted_at: '2026-04-14T10:00:00.000Z',
    stage_label: 'Revisão L2',
    quorum_progress: '1/2',
    stage_kind: 'review',
    due_at: '2026-04-21T10:00:00.000Z',
    ...overrides,
  };
}

function mockQueries(items: InboxItem[]) {
  vi.mocked(useDashboardInboxQuery).mockReturnValue({
    data: { items, total: items.length },
    isLoading: false,
    isError: false,
  } as unknown as ReturnType<typeof useDashboardInboxQuery>);

  vi.mocked(useLibraryStatsQuery).mockReturnValue({
    data: { by_status: {}, by_area: {} },
    isLoading: false,
    isError: false,
  } as unknown as ReturnType<typeof useLibraryStatsQuery>);

  vi.mocked(useDashboardActivityQuery).mockReturnValue({
    data: { items: [] },
    isLoading: false,
    isError: false,
  } as unknown as ReturnType<typeof useDashboardActivityQuery>);
}

function renderPage() {
  return render(
    <MemoryRouter initialEntries={['/']}>
      <DashboardPage />
    </MemoryRouter>,
  );
}

describe('DashboardPage', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    useAuthStore.setState({ authState: 'ready', user: null });
  });

  it('renders a pending row from the inbox query', () => {
    mockQueries([makeItem()]);
    renderPage();

    expect(screen.getByText('POP Limpeza de Linha')).toBeInTheDocument();
    expect(screen.getByText('Revisão L2')).toBeInTheDocument();
  });

  // F-QA4-8: the row identifies the document by its canonical human code
  // (controlled_documents.code) — the value the reader recognizes and the only
  // one getKindFromCode can parse. The controlled-document uuid is a fallback
  // for wire drift, never what a coded row displays.
  it('pins the pending row to controlled_document_code, never the controlled-document uuid', () => {
    mockQueries([makeItem()]);
    renderPage();

    expect(screen.getByText('POP-QUA-0555')).toBeInTheDocument();
    expect(screen.queryByText(CONTROLLED_DOCUMENT_UUID)).toBeNull();
    // Kind chip is parsed off the code, not off the uuid.
    expect(screen.getByText('POP')).toBeInTheDocument();
  });
});
