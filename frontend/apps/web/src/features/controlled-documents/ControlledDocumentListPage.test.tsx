import { render, screen, waitFor } from '@testing-library/react';
import { MemoryRouter, Route, Routes } from 'react-router-dom';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ControlledDocumentListPage } from './ControlledDocumentListPage';

vi.mock('./api/controlledDocuments', async (importOriginal) => {
  const orig = await importOriginal<typeof import('./api/controlledDocuments')>();
  return {
    ...orig,
    fetchControlledDocuments: vi.fn(),
  };
});

vi.mock('./ControlledDocumentDetailPage', () => ({
  ControlledDocumentDetailPage: ({ id }: { id: string }) => <div data-testid="registry-detail-page">{id}</div>,
}));

import * as api from './api/controlledDocuments';

describe('ControlledDocumentListPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('opens detail mode from the modern /controlled-documents base path', async () => {
    vi.mocked(api.fetchControlledDocuments).mockResolvedValue([]);

    render(
      <MemoryRouter initialEntries={['/controlled-documents/750afeba-6e35-4dd4-8a74-2b51b9f9090c']}>
        <Routes>
          <Route path="/controlled-documents/*" element={<ControlledDocumentListPage basePath="/controlled-documents" />} />
        </Routes>
      </MemoryRouter>,
    );

    await waitFor(() => {
      expect(screen.getByTestId('registry-detail-page')).toHaveTextContent('750afeba-6e35-4dd4-8a74-2b51b9f9090c');
    });
  });
});
