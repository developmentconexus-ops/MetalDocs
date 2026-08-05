import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { ApiError } from '../../../lib/api';
import { ExtendSlaDialog } from './ExtendSlaDialog';

describe('ExtendSlaDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('disables confirm until both a new date and a reason are provided', () => {
    render(
      <ExtendSlaDialog
        open
        stageLabel="Revisão técnica"
        currentDueAt="2026-08-01T00:00:00.000Z"
        isSubmitting={false}
        onConfirm={vi.fn()}
        onClose={vi.fn()}
      />,
    );

    expect(screen.getByRole('button', { name: 'Confirmar prorrogação' })).toBeDisabled();

    fireEvent.change(screen.getByLabelText('Novo prazo'), {
      target: { value: '2026-09-01T10:00' },
    });
    expect(screen.getByRole('button', { name: 'Confirmar prorrogação' })).toBeDisabled();

    fireEvent.change(screen.getByLabelText('Motivo da prorrogação'), {
      target: { value: 'Aprovador em licença médica' },
    });
    expect(screen.getByRole('button', { name: 'Confirmar prorrogação' })).not.toBeDisabled();
  });

  it('calls onConfirm with an ISO due_at and the trimmed reason, then onClose is left to the caller', async () => {
    const onConfirm = vi.fn().mockResolvedValue(undefined);
    const onClose = vi.fn();

    render(
      <ExtendSlaDialog
        open
        stageLabel="Revisão técnica"
        currentDueAt="2026-08-01T00:00:00.000Z"
        isSubmitting={false}
        onConfirm={onConfirm}
        onClose={onClose}
      />,
    );

    fireEvent.change(screen.getByLabelText('Novo prazo'), {
      target: { value: '2026-09-01T10:00' },
    });
    fireEvent.change(screen.getByLabelText('Motivo da prorrogação'), {
      target: { value: '  motivo válido  ' },
    });

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar prorrogação' }));
    });

    await waitFor(() => expect(onConfirm).toHaveBeenCalledTimes(1));
    const [dueAtArg, reasonArg] = onConfirm.mock.calls[0] as [string, string];
    expect(reasonArg).toBe('motivo válido');
    expect(new Date(dueAtArg).toISOString()).toBe(dueAtArg);
  });

  it('surfaces the resolved backend error and does not clear the form when onConfirm rejects', async () => {
    const onConfirm = vi.fn().mockRejectedValue(
      new ApiError('validation.sla_extension_not_forward', 422, 'not forward'),
    );

    render(
      <ExtendSlaDialog
        open
        stageLabel="Revisão técnica"
        currentDueAt="2026-08-01T00:00:00.000Z"
        isSubmitting={false}
        onConfirm={onConfirm}
        onClose={vi.fn()}
      />,
    );

    fireEvent.change(screen.getByLabelText('Novo prazo'), {
      target: { value: '2026-01-01T10:00' },
    });
    fireEvent.change(screen.getByLabelText('Motivo da prorrogação'), {
      target: { value: 'motivo' },
    });

    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'Confirmar prorrogação' }));
    });

    await waitFor(() => {
      expect(screen.getByRole('alert')).toBeTruthy();
      expect(
        screen.getByText('A nova data limite deve ser posterior ao prazo atual.'),
      ).toBeTruthy();
    });
  });

  it('calls onClose and does not call onConfirm when Voltar is clicked', () => {
    const onClose = vi.fn();
    const onConfirm = vi.fn();

    render(
      <ExtendSlaDialog
        open
        stageLabel="Revisão técnica"
        currentDueAt="2026-08-01T00:00:00.000Z"
        isSubmitting={false}
        onConfirm={onConfirm}
        onClose={onClose}
      />,
    );

    fireEvent.click(screen.getByRole('button', { name: 'Voltar' }));

    expect(onClose).toHaveBeenCalledTimes(1);
    expect(onConfirm).not.toHaveBeenCalled();
  });
});
