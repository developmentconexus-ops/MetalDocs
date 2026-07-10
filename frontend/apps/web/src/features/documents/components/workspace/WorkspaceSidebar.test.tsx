import type { ComponentProps } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { describe, expect, it, vi } from 'vitest';
import { WorkspaceSidebar } from './WorkspaceSidebar';
import type { DocumentDetail } from '../../api/documents';
import type { ApprovalInstance, StageInstance } from '../../../approval/api/approvalTypes';

function makeDoc(overrides: Partial<DocumentDetail> = {}): DocumentDetail {
  return {
    id: 'doc-1',
    code: 'POP-QUA-0148',
    name: 'POP Limpeza de Linha',
    status: 'under_review',
    revision_version: 3,
    revision_number: 3,
    controlled_document_id: 'cd-1',
    current_revision_id: 'rev-1',
    created_by: 'user-author-1',
    created_at: '2026-05-01T10:00:00.000Z',
    ...overrides,
  } as DocumentDetail;
}

function makeStage(overrides: Partial<StageInstance> = {}): StageInstance {
  return {
    id: 'stage-1',
    stage_index: 0,
    label: 'Revisão técnica',
    status: 'active',
    stage_kind: 'review',
    signoffs: [],
    actors: [],
    due_at: null,
    ...overrides,
  } as StageInstance;
}

function makeInstance(overrides: Partial<ApprovalInstance> = {}): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'r1',
    tenant_id: 't1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-05-01T10:00:00.000Z',
    completed_at: null,
    stages: [makeStage()],
    etag: '"v1"',
    frozen_content_hash: null,
    viewer: {
      is_author: false,
      eligible_for_active_stage: true,
      has_signed_active_stage: false,
      via_delegation_from: null,
    },
    verdicts: [],
    ...overrides,
  } as ApprovalInstance;
}

type Props = ComponentProps<typeof WorkspaceSidebar>;

function renderSidebar(overrides: Partial<Props> = {}) {
  const doc = overrides.doc ?? makeDoc();
  const instance = 'instance' in overrides ? (overrides.instance ?? null) : makeInstance();
  const props: Props = {
    documentId: 'doc-1',
    doc,
    instance,
    instanceLoading: false,
    instanceError: null,
    onRetryInstance: vi.fn(),
    activeStage: instance?.stages.find((s) => s.status === 'active'),
    onRefetchInstance: vi.fn(),
    ...overrides,
  };
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter>
        <WorkspaceSidebar {...props} />
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe('WorkspaceSidebar', () => {
  it('renders the embedded meta panel, the timeline, and a link to the record view', () => {
    renderSidebar();
    expect(screen.getByText('Identificação')).toBeInTheDocument();
    expect(screen.getByLabelText('Timeline de aprovação')).toBeInTheDocument();
    const link = screen.getByRole('link', { name: /ficha completa/i });
    expect(link).toHaveAttribute('href', '/documents/doc-1/details');
  });

  it('renders verdict CTAs when the viewer is eligible on an active review stage', () => {
    renderSidebar();
    expect(screen.getByRole('button', { name: 'Pronto para aprovação' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Solicitar mudanças' })).toBeInTheDocument();
  });

  it('renders no verdict CTAs and no footer for an ineligible (observing) viewer', () => {
    const instance = makeInstance({
      viewer: {
        is_author: false,
        eligible_for_active_stage: false,
        has_signed_active_stage: false,
        via_delegation_from: null,
      },
    });
    renderSidebar({ instance, activeStage: instance.stages[0] });
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).not.toBeInTheDocument();
    expect(screen.queryByTestId('approval-sidebar-footer')).not.toBeInTheDocument();
  });

  it('renders no footer at all when there is no approval instance, but keeps the meta panel', () => {
    renderSidebar({ instance: null, activeStage: undefined });
    expect(screen.queryByTestId('approval-sidebar-footer')).not.toBeInTheDocument();
    expect(screen.getByText('Identificação')).toBeInTheDocument();
  });

  it('renders a passed lifecycle action in the reviewing footer and fires run on click', () => {
    const run = vi.fn();
    renderSidebar({
      lifecycleActions: [
        { key: 'cancel', label: 'Cancelar instância', variant: 'secondary', available: true, run },
      ],
    });
    // Verdict CTAs (reviewing) AND the lifecycle action both render.
    expect(screen.getByRole('button', { name: 'Pronto para aprovação' })).toBeInTheDocument();
    const cancelBtn = screen.getByRole('button', { name: 'Cancelar instância' });
    expect(cancelBtn).toBeInTheDocument();
    cancelBtn.click();
    expect(run).toHaveBeenCalledTimes(1);
  });

  it('renders a lifecycle action in the observing footer even with no verdict CTAs', () => {
    const instance = makeInstance({
      viewer: {
        is_author: false,
        eligible_for_active_stage: false,
        has_signed_active_stage: false,
        via_delegation_from: null,
      },
    });
    renderSidebar({
      instance,
      activeStage: instance.stages[0],
      lifecycleActions: [
        { key: 'cancel', label: 'Cancelar instância', variant: 'secondary', available: true, run: vi.fn() },
      ],
    });
    expect(screen.queryByRole('button', { name: 'Pronto para aprovação' })).not.toBeInTheDocument();
    expect(screen.getByTestId('approval-sidebar-footer')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Cancelar instância' })).toBeInTheDocument();
  });

  it('renders no lifecycle footer when the action list is empty and the viewer is observing', () => {
    const instance = makeInstance({
      viewer: {
        is_author: false,
        eligible_for_active_stage: false,
        has_signed_active_stage: false,
        via_delegation_from: null,
      },
    });
    renderSidebar({ instance, activeStage: instance.stages[0], lifecycleActions: [] });
    expect(screen.queryByTestId('approval-sidebar-footer')).not.toBeInTheDocument();
  });

  it('surfaces an instance error in the timeline panel and wires the retry callback', () => {
    const onRetryInstance = vi.fn();
    renderSidebar({
      instance: null,
      activeStage: undefined,
      instanceError: 'Erro ao carregar dados de aprovação.',
      onRetryInstance,
    });
    expect(screen.getByRole('alert')).toHaveTextContent('Erro ao carregar dados de aprovação.');
    screen.getByRole('button', { name: 'Tentar novamente' }).click();
    expect(onRetryInstance).toHaveBeenCalledTimes(1);
  });
});
