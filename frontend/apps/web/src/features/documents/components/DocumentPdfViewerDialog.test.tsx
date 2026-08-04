import { fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../../lib/api';
import { DocumentPdfViewerDialog } from './DocumentPdfViewerDialog';

type QueryShape = ReturnType<typeof buildQuery>;

function buildQuery(overrides: Record<string, unknown> = {}) {
  return {
    data: undefined as Blob | undefined,
    error: null as unknown,
    isPending: false,
    isFetching: false,
    isSuccess: false,
    isError: false,
    refetch: vi.fn(),
    ...overrides,
  };
}

const hookState: { value: QueryShape } = { value: buildQuery({ isPending: true }) };

vi.mock('../queries/useDocumentPdfExportQuery', () => ({
  useDocumentPdfExportQuery: () => hookState.value,
}));

const createObjectURL = vi.fn(() => 'blob:metaldocs/pdf-1');
const revokeObjectURL = vi.fn();

function renderDialog(onClose = vi.fn()) {
  return {
    onClose,
    ...render(
      <DocumentPdfViewerDialog documentId='doc-1' documentLabel='POP-QUA-0148' onClose={onClose} />,
    ),
  };
}

describe('DocumentPdfViewerDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    hookState.value = buildQuery({ isPending: true });
    Object.defineProperty(URL, 'createObjectURL', { value: createObjectURL, configurable: true });
    Object.defineProperty(URL, 'revokeObjectURL', { value: revokeObjectURL, configurable: true });
  });

  it('loading: announces the fetch and renders no embed', () => {
    renderDialog();
    expect(screen.getByRole('status')).toHaveTextContent('Carregando o PDF do documento');
    expect(screen.queryByTitle('PDF do documento POP-QUA-0148')).not.toBeInTheDocument();
  });

  it('ready: embeds the PDF from an object URL (not a download link)', () => {
    const blob = new Blob(['%PDF-1.7'], { type: 'application/pdf' });
    hookState.value = buildQuery({ data: blob, isSuccess: true });

    renderDialog();

    expect(createObjectURL).toHaveBeenCalledWith(blob);
    const frame = screen.getByTitle('PDF do documento POP-QUA-0148');
    expect(frame.tagName).toBe('IFRAME');
    expect(frame).toHaveAttribute('src', 'blob:metaldocs/pdf-1');
    expect(document.querySelector('a[download]')).toBeNull();
  });

  it('revokes the object URL on unmount (no leaked blob)', () => {
    hookState.value = buildQuery({
      data: new Blob(['%PDF-1.7'], { type: 'application/pdf' }),
      isSuccess: true,
    });

    const { unmount } = renderDialog();
    expect(revokeObjectURL).not.toHaveBeenCalled();

    unmount();
    expect(revokeObjectURL).toHaveBeenCalledWith('blob:metaldocs/pdf-1');
  });

  it('error: surfaces the mapped problem message and retries', () => {
    const refetch = vi.fn();
    hookState.value = buildQuery({
      isError: true,
      error: new ApiError('ratelimit.exceeded', 429, 'rate limited'),
      refetch,
    });

    renderDialog();

    expect(screen.getByRole('alert')).toHaveTextContent('Muitas requisições em sequência');
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' }));
    expect(refetch).toHaveBeenCalledTimes(1);
  });

  it('settled with no bytes: fails closed instead of spinning forever', () => {
    hookState.value = buildQuery({ isSuccess: true, data: undefined });
    renderDialog();
    expect(screen.getByRole('alert')).toHaveTextContent('veio vazio');
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });

  it('closes via the close button and via Escape', () => {
    const { onClose } = renderDialog();

    fireEvent.click(screen.getByRole('button', { name: 'Fechar visualização do PDF' }));
    expect(onClose).toHaveBeenCalledTimes(1);

    fireEvent.keyDown(window, { key: 'Escape' });
    expect(onClose).toHaveBeenCalledTimes(2);
  });
});
