import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import * as approvalApi from '../api/approvalApi';
import { CancelInstanceDialog } from './CancelInstanceDialog';

vi.mock('../api/approvalApi');

describe('CancelInstanceDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('renders the dialog and disables the confirm button when reason is empty', () => {
    render(
      <CancelInstanceDialog
        documentId='doc-1'
        onClose={vi.fn()}
        onSuccess={vi.fn()}
      />,
    );

    expect(screen.getByRole('dialog', { name: 'Cancelar instância de aprovação' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Confirmar cancelamento' })).toBeDisabled();
  });

  it('enables the confirm button when a reason is typed, calls cancel + onSuccess + onClose on submit', async () => {
    const onClose = vi.fn();
    const onSuccess = vi.fn();
    vi.mocked(approvalApi.cancel).mockResolvedValue({ document_id: 'doc-1' } as Awaited<ReturnType<typeof approvalApi.cancel>>);

    render(
      <CancelInstanceDialog
        documentId='doc-1'
        onClose={onClose}
        onSuccess={onSuccess}
      />,
    );

    const textarea = screen.getByLabelText('Motivo do cancelamento');
    fireEvent.change(textarea, { target: { value: '  motivo  ' } });

    const confirmBtn = screen.getByRole('button', { name: 'Confirmar cancelamento' });
    expect(confirmBtn).not.toBeDisabled();

    await act(async () => {
      fireEvent.click(confirmBtn);
    });

    await waitFor(() => {
      expect(vi.mocked(approvalApi.cancel)).toHaveBeenCalledTimes(1);
      expect(vi.mocked(approvalApi.cancel)).toHaveBeenCalledWith('doc-1', { reason: 'motivo' });
      expect(onSuccess).toHaveBeenCalledTimes(1);
      expect(onClose).toHaveBeenCalledTimes(1);
    });
  });

  it('shows an error alert and does NOT call onClose when cancel rejects', async () => {
    const onClose = vi.fn();
    const onSuccess = vi.fn();
    vi.mocked(approvalApi.cancel).mockRejectedValue(new Error('server error'));

    render(
      <CancelInstanceDialog
        documentId='doc-1'
        onClose={onClose}
        onSuccess={onSuccess}
      />,
    );

    fireEvent.change(screen.getByLabelText('Motivo do cancelamento'), {
      target: { value: 'reason' },
    });

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar cancelamento' }));
    });

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeTruthy();
      expect(screen.getByText('Erro ao cancelar instância.')).toBeTruthy();
    });

    expect(onClose).not.toHaveBeenCalled();
  });

  it('calls onClose and does NOT call cancel when Voltar is clicked', () => {
    const onClose = vi.fn();

    render(
      <CancelInstanceDialog
        documentId='doc-1'
        onClose={onClose}
        onSuccess={vi.fn()}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Voltar' }));

    expect(onClose).toHaveBeenCalledTimes(1);
    expect(vi.mocked(approvalApi.cancel)).not.toHaveBeenCalled();
  });
});
