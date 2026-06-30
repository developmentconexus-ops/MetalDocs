/**
 * F-FE5: VersionActionPanel — type='button' on all action buttons + pt-BR labels/toasts.
 */
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { VersionActionPanel } from '../VersionActionPanel';
import type { VersionDTO } from '../api/templates';
import { useAuthStore } from '../../../store/auth.store';
import type { CurrentUser, UserRole } from '../../../lib/types';

vi.mock('../api/templates', async () => {
  const actual = await vi.importActual<typeof import('../api/templates')>('../api/templates');
  return {
    ...actual,
    approveVersion: vi.fn(),
    reviewVersion: vi.fn(),
  };
});

import { approveVersion, reviewVersion } from '../api/templates';

const approveMock = approveVersion as unknown as ReturnType<typeof vi.fn>;
const reviewMock = reviewVersion as unknown as ReturnType<typeof vi.fn>;

function seedActor(roles: string[], capabilities: string[]): void {
  const user: CurrentUser = {
    userId: 'u-1',
    tenantId: 't-1',
    tenantName: 'Test Tenant',
    username: 'tester',
    displayName: 'Tester',
    mustChangePassword: false,
    roles: roles as UserRole[],
    capabilities,
  };
  useAuthStore.setState({ user });
}

function makeVersion(overrides: Partial<VersionDTO> = {}): VersionDTO {
  return {
    id: 'ver-1',
    template_id: 'tpl-1',
    version_number: 1,
    revision_number: 0,
    status: 'in_review',
    docx_storage_key: null,
    content_hash: null,
    metadata_schema: null,
    placeholder_schema: null,
    author_id: 'author-1',
    pending_reviewer_role: null,
    pending_approver_role: 'approver',
    reviewer_id: null,
    approver_id: null,
    submitted_at: null,
    reviewed_at: null,
    approved_at: null,
    published_at: null,
    obsoleted_at: null,
    created_at: new Date().toISOString(),
    ...overrides,
  };
}

beforeEach(() => {
  approveMock.mockReset();
  reviewMock.mockReset();
  useAuthStore.setState({ user: null });
});

describe('VersionActionPanel — type=button (F-FE5)', () => {
  it('in_review no-reviewer: Publicar and Rejeitar buttons have type=button', () => {
    seedActor(['approver'], ['template.approve']);
    render(
      <VersionActionPanel version={makeVersion()} onVersionUpdate={() => {}} />,
    );
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) {
      expect(btn).toHaveAttribute('type', 'button');
    }
  });

  it('in_review with-reviewer: Aprovar Revisão and Rejeitar buttons have type=button', () => {
    seedActor(['reviewer'], ['template.review']);
    render(
      <VersionActionPanel
        version={makeVersion({ pending_reviewer_role: 'reviewer' })}
        onVersionUpdate={() => {}}
      />,
    );
    const buttons = screen.getAllByRole('button');
    expect(buttons.length).toBeGreaterThanOrEqual(2);
    for (const btn of buttons) {
      expect(btn).toHaveAttribute('type', 'button');
    }
  });

  it('approved: Publicar and Rejeitar buttons have type=button', () => {
    seedActor(['approver'], ['template.approve', 'template.publish']);
    render(
      <VersionActionPanel
        version={makeVersion({ status: 'approved' })}
        onVersionUpdate={() => {}}
      />,
    );
    const buttons = screen.getAllByRole('button');
    for (const btn of buttons) {
      expect(btn).toHaveAttribute('type', 'button');
    }
  });
});

describe('VersionActionPanel — pt-BR labels (F-FE5)', () => {
  it('in_review no-reviewer: shows pt-BR button labels', () => {
    seedActor(['approver'], ['template.approve']);
    render(
      <VersionActionPanel version={makeVersion()} onVersionUpdate={() => {}} />,
    );
    expect(screen.getByRole('button', { name: 'Publicar' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Rejeitar' })).toBeInTheDocument();
    expect(screen.getByText('Ações do aprovador')).toBeInTheDocument();
  });

  it('in_review with-reviewer: shows pt-BR button labels and heading', () => {
    seedActor(['reviewer'], ['template.review']);
    render(
      <VersionActionPanel
        version={makeVersion({ pending_reviewer_role: 'reviewer' })}
        onVersionUpdate={() => {}}
      />,
    );
    expect(screen.getByRole('button', { name: 'Aprovar Revisão' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Rejeitar' })).toBeInTheDocument();
    expect(screen.getByText('Ações do revisor')).toBeInTheDocument();
  });

  it('approved: shows pt-BR button labels', () => {
    seedActor(['approver'], ['template.approve', 'template.publish']);
    render(
      <VersionActionPanel
        version={makeVersion({ status: 'approved' })}
        onVersionUpdate={() => {}}
      />,
    );
    expect(screen.getByRole('button', { name: 'Publicar' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Rejeitar' })).toBeInTheDocument();
    expect(screen.getByText('Ações do aprovador')).toBeInTheDocument();
  });

  it('shows pt-BR success toast on publish (no-reviewer path)', async () => {
    seedActor(['approver'], ['template.approve']);
    const v = makeVersion();
    const published = { ...v, status: 'published' as const };
    approveMock.mockResolvedValueOnce(published);

    render(<VersionActionPanel version={v} onVersionUpdate={() => {}} />);
    fireEvent.click(screen.getByRole('button', { name: 'Publicar' }));

    await waitFor(() =>
      expect(screen.getByText('Publicado')).toBeInTheDocument(),
    );
  });

  it('shows pt-BR success toast on reject (no-reviewer path)', async () => {
    seedActor(['approver'], ['template.approve']);
    const v = makeVersion();
    const rejected = { ...v, status: 'draft' as const };
    approveMock.mockResolvedValueOnce(rejected);

    render(<VersionActionPanel version={v} onVersionUpdate={() => {}} />);
    fireEvent.click(screen.getByRole('button', { name: 'Rejeitar' }));

    await waitFor(() =>
      expect(screen.getByText('Rejeitado — volta para rascunho')).toBeInTheDocument(),
    );
  });

  it('shows pt-BR success toast on review approve', async () => {
    seedActor(['reviewer'], ['template.review']);
    const v = makeVersion({ pending_reviewer_role: 'reviewer' });
    const reviewed = { ...v, status: 'approved' as const };
    reviewMock.mockResolvedValueOnce(reviewed);

    render(<VersionActionPanel version={v} onVersionUpdate={() => {}} />);
    fireEvent.click(screen.getByRole('button', { name: 'Aprovar Revisão' }));

    await waitFor(() =>
      expect(screen.getByText('Revisão aprovada')).toBeInTheDocument(),
    );
  });
});
