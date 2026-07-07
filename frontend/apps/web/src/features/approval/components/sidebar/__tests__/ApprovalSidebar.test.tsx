import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import type { TrackedChange } from '@metaldocs/editor-ui';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import type { ApprovalInstance, StageInstance } from '../../../api/approvalTypes';
import type { ArtifactDecisionModel } from '../../../../shared/controlled-artifact/types';
import { ApprovalSidebar } from '../ApprovalSidebar';

const mutateAsyncMock = vi.fn();

vi.mock('../../../queries/useReviewVerdictMutation', () => ({
  useReviewVerdictMutation: () => ({ mutateAsync: mutateAsyncMock }),
}));

vi.mock('../../../../documents/queries/useDocumentCommentsQuery', () => ({
  useDocumentCommentsQuery: () => ({ data: [], isLoading: false, isError: false }),
}));

function makeStage(overrides: Partial<StageInstance> = {}): StageInstance {
  return {
    id: 'stage-1',
    stage_index: 0,
    label: 'Revisão técnica',
    status: 'active',
    signoffs: [],
    actors: [
      { user_id: 'u1', display_name: 'Ana', status: 'waiting', decision: null },
      { user_id: 'u2', display_name: 'Bruno', status: 'waiting', decision: null },
    ],
    stage_kind: 'review',
    due_at: null,
    ...overrides,
  } as StageInstance;
}

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
    stages: [makeStage()],
    etag: '"v3"',
    frozen_content_hash: 'a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90',
    ...overrides,
  } as ApprovalInstance;
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

function renderSidebar(props: Partial<React.ComponentProps<typeof ApprovalSidebar>> = {}) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const instance = props.instance ?? makeInstance();
  const activeStage = props.activeStage ?? instance.stages[0];
  return render(
    <QueryClientProvider client={qc}>
      <ApprovalSidebar
        editorMode="review"
        isApprovalStage={false}
        instance={instance}
        activeStage={activeStage}
        documentId="doc-1"
        contentHash="hash-abc"
        frozenContentHash={instance.frozen_content_hash}
        decision={null}
        actions={[]}
        trackedChanges={[]}
        onRefetchInstance={vi.fn()}
        {...props}
      />
    </QueryClientProvider>,
  );
}

describe('ApprovalSidebar', () => {
  beforeEach(() => {
    mutateAsyncMock.mockReset();
    mutateAsyncMock.mockResolvedValue({ verdict_id: 'v1', was_replay: false, outcome: 'ok' });
  });

  it('review mode: renders SuggestionList + two verdict CTAs, no signoff button, no password input', () => {
    const trackedChanges: TrackedChange[] = [
      { revisionId: 'tc-1', author: 'Ana', type: 'insertion', excerpt: 'trecho inserido' },
    ];
    renderSidebar({ isApprovalStage: false, decision: null, trackedChanges });

    expect(screen.getByRole('button', { name: 'Pronto para aprovação' })).toBeTruthy();
    expect(screen.getByRole('button', { name: 'Solicitar mudanças' })).toBeTruthy();
    expect(screen.queryByRole('button', { name: /Aprovar e assinar/i })).toBeNull();
    expect(screen.queryByLabelText(/senha/i)).toBeNull();
  });

  it('approval mode: renders ArtifactDecisionPanel (password present) + "Outras ações", no verdict CTAs/suggestions', () => {
    const action = {
      key: 'publish',
      label: 'Publicar / Agendar',
      variant: 'primary' as const,
      available: true,
      run: vi.fn(),
    };
    renderSidebar({
      isApprovalStage: true,
      decision: makeDecision(),
      actions: [action],
      activeStage: makeStage({ stage_kind: 'approval' }),
    });

    expect(screen.getByLabelText('Senha')).toBeTruthy();
    expect(screen.getByText('Outras ações')).toBeTruthy();
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).toBeNull();
    expect(screen.queryByRole('button', { name: 'Solicitar mudanças' })).toBeNull();
  });

  it('renders exactly ONE timeline in the DOM', () => {
    renderSidebar();
    const timelines = screen.getAllByRole('region', { name: /timeline de aprovação/i });
    expect(timelines).toHaveLength(1);
  });

  it('hides the frozen content hash while the integrity disclosure is collapsed, shows it after expanding', async () => {
    renderSidebar();
    const hash = 'a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90';
    expect(screen.queryByText(hash)).toBeNull();

    const summary = screen.getByText(/Conteúdo verificado/);
    fireEvent.click(summary);

    await waitFor(() => {
      expect(screen.getByText(hash)).toBeTruthy();
    });
  });

  it('scroll container has overflow-y auto and the footer is sticky', () => {
    renderSidebar();
    const scrollRegion = screen.getByTestId('approval-sidebar-scroll');
    expect(scrollRegion.style.overflowY).toBe('auto');
    const footer = screen.getByTestId('approval-sidebar-footer');
    expect(footer.style.position).toBe('sticky');
  });
});
