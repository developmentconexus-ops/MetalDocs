import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { ApprovalInstance, StageInstance } from '../../../api/approvalTypes';
import type { ArtifactAction, ArtifactDecisionModel } from '../../../../shared/controlled-artifact/types';
import { DecisionFooter } from '../DecisionFooter';

const mutateAsyncMock = vi.fn();

vi.mock('../../../queries/useReviewVerdictMutation', () => ({
  useReviewVerdictMutation: () => ({ mutateAsync: mutateAsyncMock }),
}));

function makeInstance(overrides: Partial<ApprovalInstance> = {}): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'route-1',
    tenant_id: 'tenant-1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-04-15T10:00:00.000Z',
    completed_at: null,
    stages: [],
    etag: '"v3"',
    frozen_content_hash: null,
    ...overrides,
  } as ApprovalInstance;
}

function makeStage(overrides: Partial<StageInstance> = {}): StageInstance {
  return {
    id: 'stage-1',
    stage_index: 0,
    label: 'Revisão técnica',
    status: 'active',
    signoffs: [],
    actors: [],
    stage_kind: 'review',
    due_at: null,
    ...overrides,
  } as StageInstance;
}

function makeDecision(overrides: Partial<ArtifactDecisionModel> = {}): ArtifactDecisionModel {
  return {
    kicker: 'Assinatura requerida',
    heading: 'Registrar decisão',
    description: 'Sua decisão é registrada com efeito de assinatura eletrônica.',
    options: [
      {
        key: 'approve',
        label: 'Assinar e aprovar',
        description: 'Aprova o documento.',
        tone: 'approve',
        submitLabel: 'Assinar e aprovar',
        requiresReason: false,
      },
      {
        key: 'reject',
        label: 'Assinar e devolver',
        description: 'Devolve o documento.',
        tone: 'reject',
        submitLabel: 'Assinar e devolver',
        requiresReason: true,
      },
    ],
    reasonLabel: 'Justificativa',
    reasonPlaceholder: 'Comentário…',
    password: { label: 'Senha' },
    legal: { text: 'Confirmo que revisei o conteúdo integralmente e que esta decisão tem efeito de assinatura eletrônica conforme a MP 2.200-2/2001.' },
    signer: { name: 'Ana Revisora', detail: 'ana@example.com', note: 'Assinatura digital gerada no ato da confirmação.' },
    defaultOptionKey: null,
    submit: vi.fn().mockResolvedValue(undefined),
    ...overrides,
  };
}

describe('DecisionFooter', () => {
  beforeEach(() => {
    mutateAsyncMock.mockReset();
    mutateAsyncMock.mockResolvedValue({ verdict_id: 'v1', was_replay: false, outcome: 'ok' });
  });

  describe('review mode', () => {
    it('renders "Pronto para aprovação" and "Solicitar mudanças" with no verdict submitted yet', () => {
      render(
        <DecisionFooter
          decision={null}
          actions={[]}
          instance={makeInstance()}
          activeStage={makeStage()}
          onRefetchInstance={vi.fn()}
        />,
      );
      expect(screen.getByRole('button', { name: 'Pronto para aprovação' })).toBeTruthy();
      expect(screen.getByRole('button', { name: 'Solicitar mudanças' })).toBeTruthy();
    });

    it('"Solicitar mudanças" opens a dialog whose submit is disabled until comment is non-empty', () => {
      render(
        <DecisionFooter
          decision={null}
          actions={[]}
          instance={makeInstance()}
          activeStage={makeStage()}
          onRefetchInstance={vi.fn()}
        />,
      );
      fireEvent.click(screen.getByRole('button', { name: 'Solicitar mudanças' }));
      const submitBtn = screen.getByRole('button', { name: 'Enviar solicitação' });
      expect(submitBtn).toBeDisabled();

      const textarea = screen.getByRole('textbox');
      fireEvent.change(textarea, { target: { value: 'Corrigir seção 3' } });
      expect(submitBtn).not.toBeDisabled();
    });

    it('submitting "Pronto para aprovação" calls useReviewVerdictMutation with the right args', async () => {
      const onRefetchInstance = vi.fn();
      render(
        <DecisionFooter
          decision={null}
          actions={[]}
          instance={makeInstance({ etag: '"v9"' })}
          activeStage={makeStage({ id: 'stage-9' })}
          onRefetchInstance={onRefetchInstance}
        />,
      );
      fireEvent.click(screen.getByRole('button', { name: 'Pronto para aprovação' }));
      await waitFor(() => {
        expect(mutateAsyncMock).toHaveBeenCalledWith({
          instanceId: 'inst-1',
          stageId: 'stage-9',
          etag: '"v9"',
          body: { verdict: 'ready' },
        });
      });
    });

    it('submitting "Solicitar mudanças" with a comment calls the mutation with request_changes + comment', async () => {
      const onRefetchInstance = vi.fn();
      render(
        <DecisionFooter
          decision={null}
          actions={[]}
          instance={makeInstance()}
          activeStage={makeStage()}
          onRefetchInstance={onRefetchInstance}
        />,
      );
      fireEvent.click(screen.getByRole('button', { name: 'Solicitar mudanças' }));
      fireEvent.change(screen.getByRole('textbox'), { target: { value: 'Corrigir seção 3' } });
      fireEvent.click(screen.getByRole('button', { name: 'Enviar solicitação' }));
      await waitFor(() => {
        expect(mutateAsyncMock).toHaveBeenCalledWith({
          instanceId: 'inst-1',
          stageId: 'stage-1',
          etag: '"v3"',
          body: { verdict: 'request_changes', comment: 'Corrigir seção 3' },
        });
      });
      await waitFor(() => expect(onRefetchInstance).toHaveBeenCalled());
    });

    it('does not render a password field or signoff CTA', () => {
      render(
        <DecisionFooter
          decision={null}
          actions={[]}
          instance={makeInstance()}
          activeStage={makeStage()}
          onRefetchInstance={vi.fn()}
        />,
      );
      expect(screen.queryByLabelText(/senha/i)).toBeNull();
      expect(screen.queryByRole('button', { name: /Aprovar e assinar/i })).toBeNull();
    });
  });

  describe('approval mode', () => {
    it('renders the decision panel with password field, no verdict CTAs', () => {
      render(
        <DecisionFooter
          decision={makeDecision()}
          actions={[]}
          instance={makeInstance()}
          activeStage={makeStage({ stage_kind: 'approval' })}
          onRefetchInstance={vi.fn()}
        />,
      );
      expect(screen.getByLabelText('Senha')).toBeTruthy();
      expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).toBeNull();
      expect(screen.queryByRole('button', { name: 'Solicitar mudanças' })).toBeNull();
    });

    it('shows meaning-of-signature line "declara aprovação" for the approve option', () => {
      render(
        <DecisionFooter
          decision={makeDecision({ defaultOptionKey: 'approve' })}
          actions={[]}
          instance={makeInstance()}
          activeStage={makeStage({ stage_kind: 'approval' })}
          onRefetchInstance={vi.fn()}
        />,
      );
      fireEvent.click(screen.getByRole('radio', { name: /Assinar e aprovar/ }));
      expect(screen.getByText(/declara aprovação/)).toBeTruthy();
    });

    it('shows meaning-of-signature line "declara rejeição" for the reject option', () => {
      render(
        <DecisionFooter
          decision={makeDecision()}
          actions={[]}
          instance={makeInstance()}
          activeStage={makeStage({ stage_kind: 'approval' })}
          onRefetchInstance={vi.fn()}
        />,
      );
      fireEvent.click(screen.getByRole('radio', { name: /Assinar e devolver/ }));
      expect(screen.getByText(/declara rejeição/)).toBeTruthy();
    });

    it('renders "Outras ações" secondary group when actions are present', () => {
      const action: ArtifactAction = {
        key: 'publish',
        label: 'Publicar / Agendar',
        variant: 'primary',
        available: true,
        run: vi.fn(),
      };
      render(
        <DecisionFooter
          decision={makeDecision()}
          actions={[action]}
          instance={makeInstance()}
          activeStage={makeStage({ stage_kind: 'approval' })}
          onRefetchInstance={vi.fn()}
        />,
      );
      expect(screen.getByText('Outras ações')).toBeTruthy();
      expect(screen.getByRole('button', { name: 'Publicar / Agendar' })).toBeTruthy();
    });
  });
});
