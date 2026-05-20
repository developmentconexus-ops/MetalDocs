import { act, fireEvent, render, screen } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import * as approvalApi from '../api/approvalApi';
import { ApprovalError } from '../api/mutationClient';
import { SupersedePublishDialog } from './SupersedePublishDialog';

vi.mock('../api/approvalApi');

describe('SupersedePublishDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks();
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-04-22T09:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('publish now happy path', async () => {
    const onClose = vi.fn();
    const onSuccess = vi.fn();
    vi.mocked(approvalApi.publish).mockResolvedValue({ document_id: 'doc-1' });

    render(
      <SupersedePublishDialog
        documentId="doc-1"
        contentHash="hash-1"
        revisionVersion={3}
        onClose={onClose}
        onSuccess={onSuccess}
      />,
    );

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar publicação' }));
      await vi.runAllTimersAsync();
    });
    expect(vi.mocked(approvalApi.publish)).toHaveBeenCalledWith(
      'doc-1',
      { content_hash: 'hash-1' },
      { ifMatch: '"v3"' },
    );
  });

  it('schedule happy path - datetime converted to UTC', async () => {
    const onClose = vi.fn();
    const onSuccess = vi.fn();
    vi.mocked(approvalApi.schedulePublish).mockResolvedValue({
      document_id: 'doc-1',
      scheduled_at: '2026-04-22T13:10:00.000Z',
    });

    render(
      <SupersedePublishDialog
        documentId="doc-1"
        contentHash="hash-1"
        revisionVersion={3}
        onClose={onClose}
        onSuccess={onSuccess}
      />,
    );

    fireEvent.click(screen.getByLabelText('Agendar publicação'));
    fireEvent.change(screen.getByLabelText('Data e hora da publicação'), {
      target: { value: '2026-04-22T10:10' },
    });

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar publicação' }));
      await vi.runAllTimersAsync();
    });
    expect(vi.mocked(approvalApi.schedulePublish)).toHaveBeenCalledWith(
      'doc-1',
      {
        effective_from: '2026-04-22T13:10:00.000Z',
      },
      { ifMatch: '"v3"' },
    );
  });

  it('past date shows validation error and blocks submit', async () => {
    vi.mocked(approvalApi.schedulePublish).mockResolvedValue({
      document_id: 'doc-1',
      scheduled_at: '2026-04-22T13:10:00.000Z',
    });

    render(
      <SupersedePublishDialog
        documentId="doc-1"
        contentHash="hash-1"
        revisionVersion={3}
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByLabelText('Agendar publicação'));
    fireEvent.change(screen.getByLabelText('Data e hora da publicação'), {
      target: { value: '2026-04-20T09:00' },
    });

    fireEvent.click(screen.getByRole('button', { name: 'Confirmar publicação' }));

    expect(screen.getByText('A data deve ser pelo menos 5 minutos no futuro.')).toBeTruthy();
    expect(vi.mocked(approvalApi.schedulePublish)).not.toHaveBeenCalled();
  });

  it('supersede is mandatory when publishedDocumentId exists', async () => {
    vi.mocked(approvalApi.supersede).mockResolvedValue({ document_id: 'doc-2' });

    render(
      <SupersedePublishDialog
        documentId="doc-1"
        contentHash="hash-2"
        revisionVersion={5}
        publishedDocumentId="doc-published"
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    expect(screen.queryByLabelText('Substituir versão publicada atual')).toBeNull();
    expect(screen.getByText('Esta publicação substituirá automaticamente a versão publicada atual.')).toBeTruthy();
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar publicação' }));
      await vi.runAllTimersAsync();
    });
    expect(vi.mocked(approvalApi.supersede)).toHaveBeenCalledWith(
      'doc-1',
      { superseded_document_id: 'doc-published' },
      { ifMatch: '"v5"' },
    );
    expect(vi.mocked(approvalApi.publish)).not.toHaveBeenCalled();
  });

  it('schedule mode communicates replacement when publishedDocumentId exists', async () => {
    vi.mocked(approvalApi.schedulePublish).mockResolvedValue({
      document_id: 'doc-1',
      scheduled_at: '2026-04-22T13:10:00.000Z',
    });

    render(
      <SupersedePublishDialog
        documentId="doc-1"
        contentHash="hash-2"
        revisionVersion={5}
        publishedDocumentId="doc-published"
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByLabelText('Agendar publicação'));

    expect(
      screen.getByText('A publicação agendada substituirá automaticamente a versão publicada atual.'),
    ).toBeTruthy();
  });

  it('sends superseded_document_id when scheduling against an existing published head', async () => {
    vi.mocked(approvalApi.schedulePublish).mockResolvedValue({
      document_id: 'doc-1',
      scheduled_at: '2026-04-22T13:10:00.000Z',
    });

    render(
      <SupersedePublishDialog
        documentId="doc-approved-1"
        contentHash="hash-1"
        revisionVersion={2}
        publishedDocumentId="doc-published-1"
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByLabelText('Agendar publicação'));
    fireEvent.change(screen.getByLabelText('Data e hora da publicação'), {
      target: { value: '2026-04-22T10:10' },
    });
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar publicação' }));
      await vi.runAllTimersAsync();
    });
    expect(vi.mocked(approvalApi.schedulePublish)).toHaveBeenCalledWith(
      'doc-approved-1',
      {
        effective_from: '2026-04-22T13:10:00.000Z',
        superseded_document_id: 'doc-published-1',
      },
      { ifMatch: '"v2"' },
    );
  });

  it('capability gate render', async () => {
    vi.mocked(approvalApi.publish).mockResolvedValue({ document_id: 'doc-1' });

    render(
      <SupersedePublishDialog
        documentId="doc-1"
        contentHash="hash-1"
        revisionVersion={3}
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    expect(screen.getByRole('dialog', { name: 'Publicação' })).toBeTruthy();
    expect(screen.getByText('Publicar agora')).toBeTruthy();
  });
  it('shows stale guidance for precondition conflicts instead of a generic error', async () => {
    vi.mocked(approvalApi.schedulePublish).mockRejectedValue(
      new ApprovalError('conflict.stale_revision', 412, 'stale revision'),
    );

    render(
      <SupersedePublishDialog
        documentId="doc-approved-1"
        contentHash="hash-1"
        revisionVersion={2}
        publishedDocumentId="doc-published-1"
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByLabelText('Agendar publicação'));
    fireEvent.change(screen.getByLabelText('Data e hora da publicação'), {
      target: { value: '2026-04-22T10:10' },
    });

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar publicação' }));
      await vi.runAllTimersAsync();
    });

    expect(
      screen.getByText('O documento foi alterado. Atualize a página antes de tentar novamente.'),
    ).toBeTruthy();
    expect(screen.queryByText('Não foi possível concluir a publicação. Tente novamente.')).toBeNull();
  });

  it('shows invalid supersede target guidance inline', async () => {
    vi.mocked(approvalApi.schedulePublish).mockRejectedValue(
      new ApprovalError('publish.invalid_supersede_target', 409, 'invalid target'),
    );

    render(
      <SupersedePublishDialog
        documentId="doc-approved-1"
        contentHash="hash-1"
        revisionVersion={2}
        publishedDocumentId="doc-published-1"
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByLabelText('Agendar publicação'));
    fireEvent.change(screen.getByLabelText('Data e hora da publicação'), {
      target: { value: '2026-04-22T10:10' },
    });

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar publicação' }));
      await vi.runAllTimersAsync();
    });

    expect(
      screen.getByText(
        'A publicação agendada não pode continuar porque a versão publicada vigente mudou. Atualize a página e revise o contexto antes de tentar novamente.',
      ),
    ).toBeTruthy();
  });
});
