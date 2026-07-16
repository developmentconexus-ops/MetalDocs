import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import type { DocumentPdfStatus } from '../../hooks/editor/useDocumentPdfStatus';
import { PdfCanvas } from './PdfCanvas';

const hookState: { value: DocumentPdfStatus } = {
  value: { status: 'pending', stalled: false, retry: vi.fn() },
};

vi.mock('../../hooks/editor/useDocumentPdfStatus', () => ({
  useDocumentPdfStatus: () => hookState.value,
}));

describe('PdfCanvas', () => {
  beforeEach(() => {
    hookState.value = { status: 'pending', stalled: false, retry: vi.fn() };
  });

  it('ready: renders the official PDF embed with the presigned URL', () => {
    hookState.value = { status: 'ready', url: 'https://s3/final.pdf', stalled: false, retry: vi.fn() };
    render(<PdfCanvas documentId="doc-1" />);
    const frame = screen.getByTitle('Documento oficial (PDF)');
    expect(frame).toHaveAttribute('src', 'https://s3/final.pdf');
  });

  it('pending: renders the generating state, no embed', () => {
    render(<PdfCanvas documentId="doc-1" />);
    expect(screen.getByRole('status')).toHaveTextContent('Gerando o PDF oficial');
    expect(screen.queryByTitle('Documento oficial (PDF)')).not.toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'Tentar novamente' })).not.toBeInTheDocument();
  });

  it('pending + stalled: keeps the generating copy and offers retry', () => {
    const retry = vi.fn();
    hookState.value = { status: 'pending', stalled: true, retry };
    render(<PdfCanvas documentId="doc-1" />);
    expect(screen.getByRole('status')).toHaveTextContent('Gerando o PDF oficial');
    expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' }));
    expect(retry).toHaveBeenCalledTimes(1);
  });

  it('failed: renders the error alert and retry re-polls', () => {
    const retry = vi.fn();
    hookState.value = { status: 'failed', stalled: false, retry };
    render(<PdfCanvas documentId="doc-1" />);
    expect(screen.getByRole('alert')).toHaveTextContent('Não foi possível gerar o PDF');
    fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' }));
    expect(retry).toHaveBeenCalledTimes(1);
  });

  it('ready but missing url: surfaces the error alert (fail closed, no infinite pending)', () => {
    hookState.value = { status: 'ready', stalled: false, retry: vi.fn() };
    render(<PdfCanvas documentId="doc-1" />);
    expect(screen.getByRole('alert')).toHaveTextContent('Não foi possível gerar o PDF');
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });
});
