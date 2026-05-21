import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { ControlledDocumentDetailPage } from './ControlledDocumentDetailPage';

vi.mock('./api/controlledDocuments', async (importOriginal) => {
  const orig = await importOriginal<typeof import('./api/controlledDocuments')>();
  return {
    ...orig,
    fetchControlledDocument: vi.fn(),
    fetchActiveDocumentInstance: vi.fn(),
    obsoleteControlledDocument: vi.fn(),
    createRevision: vi.fn(),
  };
});

vi.mock('../approval/components/ControlledDocumentDetailPanel', () => ({
  ControlledDocumentDetailPanel: ({ documentId }: { documentId: string }) => (
    <div data-testid="registry-detail-panel">{documentId}</div>
  ),
}));

vi.mock('./PublishedDownloadCell', () => ({
  PublishedDownloadCell: ({ documentId }: { documentId: string }) => (
    <div data-testid="published-download-cell">{documentId}</div>
  ),
}));

import * as api from './api/controlledDocuments';

function makeControlledDoc(overrides: Partial<{
  id: string; code: string; title: string; status: 'active' | 'obsolete' | 'superseded';
  profileCode: string; processAreaCode: string; ownerUserId: string; createdAt: string; updatedAt: string;
}> = {}) {
  return {
    id: 'cd-1',
    code: 'SOP-001',
    title: 'Test Doc',
    status: 'active' as const,
    profileCode: 'sop',
    processAreaCode: 'ops',
    departmentCode: null,
    ownerUserId: 'user-1',
    sequenceNum: 1,
    overrideTemplateVersionId: null,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

beforeEach(() => {
  vi.clearAllMocks();
});

afterEach(() => {
  vi.restoreAllMocks();
});

describe('ControlledDocumentDetailPage E10', () => {
  it('shows published banner when only published revision exists', async () => {
    vi.mocked(api.fetchControlledDocument).mockResolvedValue(makeControlledDoc() as never);
    vi.mocked(api.fetchActiveDocumentInstance).mockResolvedValue({
      publishedDocumentId: 'pub-1',
    });

    render(<ControlledDocumentDetailPage id="cd-1" onBack={() => {}} />);

    await waitFor(() => {
      expect(screen.getByText(/Revisão publicada/i)).toBeTruthy();
    });
    expect(screen.getByTestId('published-download-cell')).toBeTruthy();
    // No active draft panel
    expect(screen.queryByTestId('registry-detail-panel')).toBeNull();
    // New Revision button shown since doc is active but no active draft
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Nova Revisão/i })).toBeTruthy();
    });
  });

  it('shows both active panel and published banner when both exist', async () => {
    vi.mocked(api.fetchControlledDocument).mockResolvedValue(makeControlledDoc() as never);
    vi.mocked(api.fetchActiveDocumentInstance).mockResolvedValue({
      documentId: 'active-doc-1',
      approvalState: 'draft',
      contentHash: 'abc',
      revisionVersion: 2,
      publishedDocumentId: 'pub-1',
    });

    render(<ControlledDocumentDetailPage id="cd-1" onBack={() => {}} />);

    await waitFor(() => {
      expect(screen.getByTestId('registry-detail-panel')).toBeTruthy();
    });
    expect(screen.getByText(/Revisão publicada/i)).toBeTruthy();
    expect(screen.getByTestId('published-download-cell')).toBeTruthy();
    // No "Nova Revisão" when active doc exists
    expect(screen.queryByRole('button', { name: /Nova Revisão/i })).toBeNull();
  });

  it('shows Nova Revisão form when no instance at all', async () => {
    vi.mocked(api.fetchControlledDocument).mockResolvedValue(makeControlledDoc() as never);
    vi.mocked(api.fetchActiveDocumentInstance).mockResolvedValue(null);

    render(<ControlledDocumentDetailPage id="cd-1" onBack={() => {}} />);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /Nova Revisão/i })).toBeTruthy();
    });
    expect(screen.queryByTestId('registry-detail-panel')).toBeNull();
    expect(screen.queryByText(/Revisão publicada/i)).toBeNull();
  });
});
