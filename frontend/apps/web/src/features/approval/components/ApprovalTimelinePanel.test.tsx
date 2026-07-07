import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { ApprovalInstance } from '../api/approvalTypes';
import { ApprovalTimelinePanel } from './ApprovalTimelinePanel';

function makeInstance(): ApprovalInstance {
  return {
    id: 'inst-1',
    document_id: 'doc-1',
    route_id: 'route-1',
    tenant_id: 'tenant-1',
    status: 'in_progress',
    submitted_by: 'maria',
    submitted_at: '2026-04-15T10:00:00.000Z',
    completed_at: null,
    stages: [
      {
        id: 'stage-1',
        stage_index: 0,
        label: 'Revisão Técnica',
        status: 'active',
        signoffs: [
          {
            id: 'sign-1',
            actor_user_id: 'joao',
            decision: 'approve',
            signature_method: 'password_reauth',
            signed_at: '2026-04-15T11:00:00.000Z',
            reason: 'Tudo certo',
            signature_meaning: 'approval',
          },
        ],
        actors: [{ user_id: 'joao', display_name: 'joao', status: 'approved', decision: 'approve' }],
        due_at: '2026-04-16T10:00:00.000Z',
      },
    ],
    etag: 'v1',
    frozen_content_hash: 'a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90',
  };
}

describe('ApprovalTimelinePanel', () => {
  it('loading state shows skeleton', () => {
    render(<ApprovalTimelinePanel instance={null} loading error={null} />);
    expect(screen.getByText('Carregando timeline...')).toBeTruthy();
  });

  it('empty instance shows empty state', () => {
    render(<ApprovalTimelinePanel instance={null} loading={false} />);
    expect(screen.getByText('Nenhum evento de aprovação registrado.')).toBeTruthy();
  });

  it('renders submitted by and actor name', () => {
    render(<ApprovalTimelinePanel instance={makeInstance()} loading={false} />);
    expect(screen.getByText(/maria/)).toBeTruthy();
    expect(screen.getByText(/joao/)).toBeTruthy();
  });

  it('renders stage with signoff decision', () => {
    render(<ApprovalTimelinePanel instance={makeInstance()} loading={false} />);
    expect(screen.getByText('Revisão Técnica')).toBeTruthy();
    expect(screen.getByText(/Aprovou/)).toBeTruthy();
  });

  it('error state renders with message', () => {
    const onRetry = vi.fn();
    render(
      <ApprovalTimelinePanel
        instance={null}
        loading={false}
        error="Falha ao carregar timeline."
        onRetry={onRetry}
      />,
    );

    expect(screen.getByText('Falha ao carregar timeline.')).toBeTruthy();
    fireEvent.click(screen.getByRole('button', { name: 'Tentar novamente' }));
    expect(onRetry).toHaveBeenCalledOnce();
  });

  // CON-10 regression: the runtime instance status is
  // `in_progress | approved | rejected | cancelled` — it never emits
  // `completed`. Before the fix, this branch checked for the nonexistent
  // `completed` value, so the "Status final" node silently never rendered
  // for approved/rejected instances.
  it.each([
    ['approved' as const, 'Aprovado'],
    ['rejected' as const, 'Rejeitado'],
    ['cancelled' as const, 'Cancelado'],
  ])('renders "Status final" for a terminal %s instance', (status, label) => {
    const instance: ApprovalInstance = {
      ...makeInstance(),
      status,
      completed_at: '2026-04-16T09:00:00.000Z',
    };
    render(<ApprovalTimelinePanel instance={instance} loading={false} />);
    expect(screen.getByText('Status final')).toBeTruthy();
    expect(screen.getByText(new RegExp(label))).toBeTruthy();
  });

  it('does not render "Status final" while an instance is in progress', () => {
    render(<ApprovalTimelinePanel instance={makeInstance()} loading={false} />);
    expect(screen.queryByText('Status final')).toBeNull();
  });
});
